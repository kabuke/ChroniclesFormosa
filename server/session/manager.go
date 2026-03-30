package session

import (
	"log"
	"sync"
	"time"

	pb "github.com/kabuke/ChroniclesFormosa/resource"
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

	mu sync.RWMutex
}

func (s *UserSession) UpdateMaxClientSeq(seq uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if seq > s.MaxClientSeq {
		s.MaxClientSeq = seq
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

// SessionManager 負責管理全伺服器的玩家連線 Session
type SessionManager struct {
	sessions     map[string]*UserSession
	ForwardQueue []*pb.Envelope
	mu           sync.RWMutex
}

var globalManager *SessionManager
var once sync.Once

func GetManager() *SessionManager {
	once.Do(func() {
		globalManager = &SessionManager{
			sessions:     make(map[string]*UserSession),
			ForwardQueue: make([]*pb.Envelope, 0),
		}
		globalManager.LoadSessions()
		go globalManager.gc()
		go globalManager.saveLoop()
		go globalManager.forwardLoop() // 20 TPS 的轉發迴圈
	})
	return globalManager
}

// CreateSession 握手後建立新的 Session 實體並給予臨時 Secret
func (m *SessionManager) CreateSession(id string, secret []byte) *UserSession {
	m.mu.Lock()
	defer m.mu.Unlock()

	s := &UserSession{
		SessionID:        id,
		SharedSecret:     secret,
		NextSeq:          1,
		RemoteWindowSize: 128,
		LastActive:       time.Now(),
		Outbox:           make([]*pb.Envelope, 0),
		History:          make([]*pb.Envelope, 0),
	}
	m.sessions[id] = s
	return s
}

func (m *SessionManager) GetSession(id string) *UserSession {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[id]
	if ok {
		s.mu.Lock()
		s.LastActive = time.Now()
		s.mu.Unlock()
	}
	return s
}

func (m *SessionManager) GetSessionByUsername(username string) *UserSession {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var target *UserSession
	for _, s := range m.sessions {
		s.mu.RLock()
		name := s.Username
		lastActive := s.LastActive
		s.mu.RUnlock()

		if name == username { // 支援多裝置登入時，取最新活躍的一個
			if target == nil || lastActive.After(target.LastActive) {
				target = s
			}
		}
	}
	return target
}

// AddToForwardQueue 新增到轉發隊列，供廣播與聊天使用
func (m *SessionManager) AddToForwardQueue(env *pb.Envelope) {
	m.mu.Lock()
	m.ForwardQueue = append(m.ForwardQueue, env)
	m.mu.Unlock()
}

// forwardLoop (20 TPS) 從隊列中批量消費業務訊息路由
func (m *SessionManager) forwardLoop() {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for range ticker.C {
		m.processForwardQueue()
	}
}

func (m *SessionManager) processForwardQueue() {
	m.mu.Lock()
	if len(m.ForwardQueue) == 0 {
		m.mu.Unlock()
		return
	}

	queue := m.ForwardQueue
	m.ForwardQueue = make([]*pb.Envelope, 0)
	m.mu.Unlock()

	// 複製當前的 sessions 進行遍歷路由
	m.mu.RLock()
	sessions := make([]*UserSession, 0, len(m.sessions))
	for _, s := range m.sessions {
		sessions = append(sessions, s)
	}
	m.mu.RUnlock()

	for _, env := range queue {
		var targets []*UserSession

		if chatEnv, ok := env.Payload.(*pb.Envelope_Chat); ok && chatEnv.Chat.Receiver != "" {
			target := chatEnv.Chat.Receiver

			if target[0] == '#' {
				// TODO: 頻道聊天 (如 #village_10) 等完整系統好再實作，先轉廣播
				targets = sessions
			} else {
				// 私聊，將 Target 導向該用戶
				if s := m.GetSessionByUsername(target); s != nil {
					targets = append(targets, s)
					// Echo 發送給發送者本人以達到客戶端顯示的同步感
					if env.Header != nil && env.Header.SessionId != s.SessionID {
						if senderSess := m.GetSession(env.Header.SessionId); senderSess != nil {
							targets = append(targets, senderSess)
						}
					}
				}
			}
		} else {
			// 全廣播模式
			targets = sessions
		}

		for _, s := range targets {
			s.QueueMessage(env)
			if s.TriggerFlush != nil {
				go s.TriggerFlush()
			}
		}
	}
}

// gc 回收超過 60 分鐘無活動之 Session
func (m *SessionManager) gc() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		m.mu.Lock()
		for id, s := range m.sessions {
			s.mu.RLock()
			lastActive := s.LastActive
			s.mu.RUnlock()
			if time.Since(lastActive) > 60*time.Minute {
				delete(m.sessions, id)
			}
		}
		m.mu.Unlock()
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

// SaveSessions 持久化
func (m *SessionManager) SaveSessions() {
	// TODO: Phase 1 任務 3.6 完成 database 層後，這裡實作非同步批次更新 SQLite WAL
	// 類似：database.GetDB().Exec("INSERT INTO ... ON CONFLICT DO UPDATE")
}

// LoadSessions 載入歷史
func (m *SessionManager) LoadSessions() {
	// TODO: Phase 1 任務 3.6 完成 database 層後，從 SQLite 取回未過期的 Session
}
