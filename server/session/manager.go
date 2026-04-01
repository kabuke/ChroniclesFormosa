package session

import (
	"sync"
	"time"

	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/model"
	"github.com/kabuke/ChroniclesFormosa/server/repo"
	"github.com/xtaci/kcp-go/v5"
	"google.golang.org/protobuf/proto"
)

type UserSession struct {
	SessionID    string
	SharedSecret []byte
	Conn         *kcp.UDPSession
	LastActive   time.Time
	History      []*pb.Envelope
	Outbox       []*pb.Envelope

	LastAck          uint64 
	MaxClientSeq     uint64 
	NextSeq          uint64 
	RemoteWindowSize uint32 

	Username  string
	FactionID int32
	VillageID int64

	TriggerFlush func()
	SendEnvelope func(*pb.Envelope)

	IsDirty bool 
	mu      sync.RWMutex
}

func (s *UserSession) UpdateMaxClientSeq(seq uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if seq > s.MaxClientSeq {
		s.MaxClientSeq = seq
		s.IsDirty = true
	}
}

func (s *UserSession) GetMaxClientSeq() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.MaxClientSeq
}

func (s *UserSession) UpdateRemoteWindow(size uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if size > 0 { s.RemoteWindowSize = size }
}

func (s *UserSession) SetUsername(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Username = name
	s.IsDirty = true
}

func (s *UserSession) SetConn(conn *kcp.UDPSession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Conn = conn
}

func (s *UserSession) ClearConn() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Conn = nil
}

func (s *UserSession) QueueMessage(originEnv *pb.Envelope) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	env := proto.Clone(originEnv).(*pb.Envelope)
	if env.Header == nil { env.Header = &pb.Header{} }

	env.Header.Seq = s.NextSeq
	env.Header.SessionId = s.SessionID
	s.NextSeq++
	s.IsDirty = true

	s.History = append(s.History, env)
	if len(s.History) > 150 { s.History = s.History[1:] }
	s.Outbox = append(s.Outbox, env)
}

func (s *UserSession) Acknowledge(ack uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ack > s.LastAck {
		s.LastAck = ack
		s.IsDirty = true
	}
	newOutbox := make([]*pb.Envelope, 0, len(s.Outbox))
	for _, env := range s.Outbox {
		if env.Header != nil && env.Header.Seq > ack { newOutbox = append(newOutbox, env) }
	}
	s.Outbox = newOutbox
}

func (s *UserSession) FlushOutbox(winLimit uint32, sendFunc func(*pb.Envelope) error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	effWin := winLimit
	if s.RemoteWindowSize > 0 && s.RemoteWindowSize < effWin { effWin = s.RemoteWindowSize }
	if effWin < 1 { effWin = 1 }

	for len(s.Outbox) > 0 {
		inFlight := (s.NextSeq - 1) - s.LastAck
		if inFlight >= uint64(effWin) { break }

		env := s.Outbox[0]
		s.Outbox = s.Outbox[1:]

		sendingEnv := proto.Clone(env).(*pb.Envelope)
		if sendingEnv.Header == nil { sendingEnv.Header = &pb.Header{} }
		sendingEnv.Header.Ack = s.MaxClientSeq

		if err := sendFunc(sendingEnv); err != nil { break }
	}
}

// SessionManager 管理器
type getReq struct { id string; resp chan *UserSession }
type getByUserReq struct { username string; resp chan *UserSession }

type SessionManager struct {
	sessions map[string]*UserSession
	registerCh    chan *UserSession
	unregisterCh  chan string
	getCh         chan getReq
	getByUserCh   chan getByUserReq
	forwardCh     chan *pb.Envelope
	allSessionsCh chan chan []*UserSession
}

var globalManager *SessionManager
var once sync.Once

func GetManager() *SessionManager {
	once.Do(func() {
		globalManager = &SessionManager{
			sessions:      make(map[string]*UserSession),
			registerCh:    make(chan *UserSession, 100),
			unregisterCh:  make(chan string, 100),
			getCh:         make(chan getReq, 100),
			getByUserCh:   make(chan getByUserReq, 100),
			forwardCh:     make(chan *pb.Envelope, 5000),
			allSessionsCh: make(chan chan []*UserSession, 10),
		}
		globalManager.LoadSessions()
		go globalManager.runLoop()
		go globalManager.saveLoop()
	})
	return globalManager
}

func (m *SessionManager) CreateSession(id string, secret []byte) *UserSession {
	s := &UserSession{SessionID: id, SharedSecret: secret, NextSeq: 1, RemoteWindowSize: 128, LastActive: time.Now()}
	m.registerCh <- s
	return s
}

func (m *SessionManager) AddSession(s *UserSession) {
	m.registerCh <- s
}

func (m *SessionManager) UnregisterSession(id string) { m.unregisterCh <- id }

func (m *SessionManager) GetSession(id string) *UserSession {
	resp := make(chan *UserSession, 1)
	m.getCh <- getReq{id: id, resp: resp}
	return <-resp
}

func (m *SessionManager) GetSessionByUsername(username string) *UserSession {
	resp := make(chan *UserSession, 1)
	m.getByUserCh <- getByUserReq{username: username, resp: resp}
	return <-resp
}

func (m *SessionManager) GetAllSessions() []*UserSession {
	req := make(chan []*UserSession, 1)
	m.allSessionsCh <- req
	return <-req
}

func (m *SessionManager) AddToForwardQueue(env *pb.Envelope) { m.forwardCh <- env }

func (m *SessionManager) runLoop() {
	for {
		select {
		case s := <-m.registerCh: m.sessions[s.SessionID] = s
		case id := <-m.unregisterCh: delete(m.sessions, id)
		case req := <-m.getCh: req.resp <- m.sessions[req.id]
		case req := <-m.getByUserCh:
			var target *UserSession
			for _, s := range m.sessions {
				if s.Username == req.username { target = s; break }
			}
			req.resp <- target
		case req := <-m.allSessionsCh:
			list := make([]*UserSession, 0, len(m.sessions))
			for _, s := range m.sessions { list = append(list, s) }
			req <- list
		case env := <-m.forwardCh: m.handleForward(env)
		}
	}
}

func (m *SessionManager) handleForward(env *pb.Envelope) {
	// 修正：handleForward 僅處理「全服廣播」或「群組廣播」，私聊由 logic 層直接發送
	for _, s := range m.sessions {
		s.QueueMessage(env)
		if s.TriggerFlush != nil { go s.TriggerFlush() }
	}
}

func (m *SessionManager) saveLoop() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C { m.SaveSessions() }
}

func (m *SessionManager) SaveSessions() {
	sessions := m.GetAllSessions()
	var dirty []*model.SessionState
	for _, s := range sessions {
		s.mu.RLock()
		if s.IsDirty {
			dirty = append(dirty, &model.SessionState{
				SessionID: s.SessionID, Username: s.Username, SharedSecret: s.SharedSecret,
				LastAck: s.LastAck, NextSeq: s.NextSeq, MaxClientSeq: s.MaxClientSeq,
			})
			s.IsDirty = false
		}
		s.mu.RUnlock()
	}
	if len(dirty) > 0 { repo.NewSessionRepo().UpsertBatch(dirty) }
}

func (m *SessionManager) LoadSessions() {
	states, _ := repo.NewSessionRepo().LoadActive()
	for _, st := range states {
		m.sessions[st.SessionID] = &UserSession{
			SessionID: st.SessionID, SharedSecret: st.SharedSecret, LastAck: st.LastAck,
			NextSeq: st.NextSeq, MaxClientSeq: st.MaxClientSeq, Username: st.Username,
			RemoteWindowSize: 128, LastActive: time.Now(),
		}
	}
}
