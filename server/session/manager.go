package session

import (
	"log"
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

	// 可靠性相關
	LastAck          uint64 // 對端確認收到的序號
	MaxClientSeq     uint64 // 本服已收到的最大客戶端序號
	NextSeq          uint64 // 本服負責分配的下一個發送序號
	RemoteWindowSize uint32 // 對端的接收窗口

	// 業務相關
	Username  string
	FactionID int32
	VillageID int64

	// 回調
	TriggerFlush func()
	SendEnvelope func(*pb.Envelope)

	IsDirty bool // 標記是否被更新過尚未存檔
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
	if size > 0 {
		s.RemoteWindowSize = size
	}
}

func (s *UserSession) SetUsername(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Username = name
	s.IsDirty = true
}

// SetConn 設定或替換底層 UDP Session
func (s *UserSession) SetConn(conn *kcp.UDPSession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Conn = conn
}

// ClearConn 清除連線 (通常在斷線時呼叫，保留 Session 狀態)
func (s *UserSession) ClearConn() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Conn = nil
}

// QueueMessage 分配發送序號(Seq)、加入發送佇列與歷史記錄
func (s *UserSession) QueueMessage(originEnv *pb.Envelope) {
	env := proto.Clone(originEnv).(*pb.Envelope)

	s.mu.Lock()
	if env.Header == nil {
		env.Header = &pb.Header{}
	}

	env.Header.Seq = s.NextSeq
	env.Header.SessionId = s.SessionID
	s.NextSeq++
	s.IsDirty = true

	// 儲存於歷史紀錄（用於 Session Resume 回放），保留最近 150 筆
	s.History = append(s.History, env)
	if len(s.History) > 150 {
		s.History = s.History[1:]
	}

	s.Outbox = append(s.Outbox, env)
	s.mu.Unlock()
}

// Acknowledge 清除對方已確認的發送隊列訊息
func (s *UserSession) Acknowledge(ack uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ack > s.LastAck {
		s.LastAck = ack
		s.IsDirty = true
	}

	newOutbox := make([]*pb.Envelope, 0, len(s.Outbox))
	for _, env := range s.Outbox {
		if env.Header != nil && env.Header.Seq > ack {
			newOutbox = append(newOutbox, env)
		}
	}
	s.Outbox = newOutbox
}

// FlushOutbox 執行「應用層滑動窗口」發送
func (s *UserSession) FlushOutbox(winLimit uint32, sendFunc func(*pb.Envelope) error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	effWin := winLimit
	if s.RemoteWindowSize > 0 && s.RemoteWindowSize < effWin {
		effWin = s.RemoteWindowSize
	}
	if effWin < 1 {
		effWin = 1
	}

	for len(s.Outbox) > 0 {
		inFlight := (s.NextSeq - 1) - s.LastAck
		// 若正在飛行中的包超出窗口，暫停發送
		if inFlight >= uint64(effWin) {
			break
		}

		env := s.Outbox[0]
		s.Outbox = s.Outbox[1:]

		sendingEnv := proto.Clone(env).(*pb.Envelope)
		if sendingEnv.Header == nil {
			sendingEnv.Header = &pb.Header{}
		}
		// 每次發送都夾帶最新的 Ack (Server 確認客戶端目前的封包進度)
		sendingEnv.Header.Ack = s.MaxClientSeq

		if err := sendFunc(sendingEnv); err != nil {
			log.Printf("Failed to flush message to %s: %v", s.SessionID, err)
			// KCP 底層通常會自己處理緩衝，如果返回錯誤通常是嚴重斷連，這裡就跳出交由心跳處理
			break
		}
	}
}

func (s *UserSession) ClearOutbox() {
	s.mu.Lock()
	s.Outbox = make([]*pb.Envelope, 0)
	s.mu.Unlock()
}

// =========================================================================

type getReq struct {
	id   string
	resp chan *UserSession
}

type getByUserReq struct {
	username string
	resp     chan *UserSession
}

// OnSessionExpired 提供外部模塊註冊，當 Session 超時斷開時觸發
var OnSessionExpired func(username string)

type SessionManager struct {
	sessions map[string]*UserSession

	// Actor 模型的 Channels
	registerCh    chan *UserSession
	unregisterCh  chan string
	getCh         chan getReq
	getByUserCh   chan getByUserReq
	forwardCh     chan *pb.Envelope
	allSessionsCh chan chan []*UserSession
	saveInterval  time.Duration
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
			forwardCh:     make(chan *pb.Envelope, 5000), // 高吞吐量緩衝
			allSessionsCh: make(chan chan []*UserSession, 10),
		}
		globalManager.LoadSessions()
		go globalManager.runLoop()  // 啟動 Actor 核心迴圈
		go globalManager.saveLoop() // 定期觸發存檔
	})
	return globalManager
}

// CreateSession 握手後建立新的 Session 實體並給予臨時 Secret
func (m *SessionManager) CreateSession(id string, secret []byte) *UserSession {
	s := &UserSession{
		SessionID:        id,
		SharedSecret:     secret,
		NextSeq:          1,
		RemoteWindowSize: 128,
		LastActive:       time.Now(),
		Outbox:           make([]*pb.Envelope, 0),
		History:          make([]*pb.Envelope, 0),
	}
	m.registerCh <- s
	return s
}

func (m *SessionManager) GetSession(id string) *UserSession {
	resp := make(chan *UserSession, 1)
	m.getCh <- getReq{id: id, resp: resp}
	s := <-resp
	if s != nil {
		s.mu.Lock()
		s.LastActive = time.Now()
		s.mu.Unlock()
	}
	return s
}

func (m *SessionManager) GetSessionByUsername(username string) *UserSession {
	resp := make(chan *UserSession, 1)
	m.getByUserCh <- getByUserReq{username: username, resp: resp}
	return <-resp
}

// AddToForwardQueue 新增到轉發隊列，供廣播與聊天使用 (現在直接寫入 Channel)
func (m *SessionManager) AddToForwardQueue(env *pb.Envelope) {
	m.forwardCh <- env
}

// runLoop (Actor 核心迴圈) 接管所有 map 的讀寫，保證無鎖安全 (Lock-free)
func (m *SessionManager) runLoop() {
	gcTicker := time.NewTicker(10 * time.Minute)
	defer gcTicker.Stop()

	for {
		select {
		case s := <-m.registerCh:
			m.sessions[s.SessionID] = s

		case id := <-m.unregisterCh:
			delete(m.sessions, id)

		case req := <-m.getCh:
			req.resp <- m.sessions[req.id]

		case req := <-m.getByUserCh:
			var target *UserSession
			for _, s := range m.sessions {
				s.mu.RLock()
				name := s.Username
				lastActive := s.LastActive
				s.mu.RUnlock()

				if name == req.username {
					if target == nil || lastActive.After(target.LastActive) {
						target = s
					}
				}
			}
			req.resp <- target

		case req := <-m.allSessionsCh:
			// Snapshot copy for persistence
			sessionsCopy := make([]*UserSession, 0, len(m.sessions))
			for _, s := range m.sessions {
				sessionsCopy = append(sessionsCopy, s)
			}
			req <- sessionsCopy

		case env := <-m.forwardCh:
			m.handleForward(env)

		case <-gcTicker.C:
			// gc
			for id, sess := range m.sessions {
				sess.mu.RLock()
				lastActive := sess.LastActive
				sess.mu.RUnlock()
				if time.Since(lastActive) > 60*time.Minute {
					// 超過 60 分鐘沒傳送過封包，我們視為真實下線 (不是短暫斷線)
					log.Printf("[SessionGC] Session expired (idle > 60m): %s (User: %s)", id, sess.Username)
					if sess.Username != "" && OnSessionExpired != nil {
						OnSessionExpired(sess.Username)
					}
					delete(m.sessions, id)
				}
			}
		}
	}
}

// handleForward 處理轉發邏輯，因為在 runLoop 內執行，所以存取 m.sessions 是極度安全的
func (m *SessionManager) handleForward(env *pb.Envelope) {
	var targets []*UserSession

	// 判斷是否為 Chat
	if chatEnv, ok := env.Payload.(*pb.Envelope_Chat); ok && chatEnv.Chat.Receiver != "" {
		targetName := chatEnv.Chat.Receiver

		if targetName[0] == '#' {
			// TODO: Channel routing (#village_10) 先 Broadbast 頂替
			for _, s := range m.sessions {
				targets = append(targets, s)
			}
		} else {
			// 私聊
			var t *UserSession
			for _, s := range m.sessions {
				s.mu.RLock()
				name := s.Username
				lastActive := s.LastActive
				s.mu.RUnlock()

				if name == targetName {
					if t == nil || lastActive.After(t.LastActive) {
						t = s
					}
				}
			}
			if t != nil {
				targets = append(targets, t)
				// Echo to Sender
				if env.Header != nil && env.Header.SessionId != t.SessionID {
					if senderSess, exists := m.sessions[env.Header.SessionId]; exists {
						targets = append(targets, senderSess)
					}
				}
			}
		}
	} else {
		// 全廣播
		for _, s := range m.sessions {
			targets = append(targets, s)
		}
	}

	// 分發至各自的 UserSession Queue 中，這是 thread-safe 的
	for _, s := range targets {
		s.QueueMessage(env)
		if s.TriggerFlush != nil {
			go s.TriggerFlush()
		}
	}
}

// saveLoop 週期性觸發持久化
func (m *SessionManager) saveLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		m.SaveSessions()
	}
}

// SaveSessions 持久化 (僅取出有 Dirty 標記的 Session 更新至 SQLite WAL)
func (m *SessionManager) SaveSessions() {
	// 從 runLoop 索取一份 Thread-Safe 的 Snapshot
	req := make(chan []*UserSession, 1)
	m.allSessionsCh <- req
	sessions := <-req // sessions 得到全部活動清單

	var dirtyStates []*model.SessionState
	for _, s := range sessions {
		s.mu.Lock()
		if s.IsDirty {
			state := &model.SessionState{
				SessionID:    s.SessionID,
				Username:     s.Username,
				SharedSecret: s.SharedSecret,
				LastAck:      s.LastAck,
				NextSeq:      s.NextSeq,
				MaxClientSeq: s.MaxClientSeq,
				UpdatedAt:    time.Now(),
			}
			dirtyStates = append(dirtyStates, state)
			s.IsDirty = false
		}
		s.mu.Unlock()
	}

	if len(dirtyStates) > 0 {
		if err := repo.NewSessionRepo().UpsertBatch(dirtyStates); err != nil {
			log.Printf("[SessionDB] Failed to UPSERT session states: %v\n", err)
		} else {
			// 定期呼叫移除 24h 前的過期連線
			_ = repo.NewSessionRepo().DeleteExpired()
		}
	}
}

// LoadSessions 載入歷史
func (m *SessionManager) LoadSessions() {
	states, err := repo.NewSessionRepo().LoadActive()
	if err != nil {
		log.Printf("[SessionDB] LoadSessions warning: %v\n", err)
		return
	}

	for _, st := range states {
		s := &UserSession{
			SessionID:        st.SessionID,
			SharedSecret:     st.SharedSecret,
			LastAck:          st.LastAck,
			NextSeq:          st.NextSeq,
			MaxClientSeq:     st.MaxClientSeq,
			Username:         st.Username,
			RemoteWindowSize: 128,
			LastActive:       time.Now(),
			Outbox:           make([]*pb.Envelope, 0),
			History:          make([]*pb.Envelope, 0),
		}
		// 因為此時 runLoop 還沒跑，可以直接操作 maps
		m.sessions[st.SessionID] = s 
	}
	if len(states) > 0 {
		log.Printf("[SessionDB] 📥 Loaded %d active sessions for Resume-playback\n", len(states))
	}
}

