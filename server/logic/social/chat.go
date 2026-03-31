package social

import (
	"log"
	"time"

	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/session"
)

// HandleChatSend 處理聊天發送邏輯，決定廣播範圍
func HandleChatSend(s *session.UserSession, req *pb.ChatMessage) {
	// 1. 補完訊息發送者與時間
	req.Sender = s.Username
	req.Timestamp = time.Now().Unix()

	log.Printf("[Chat] %s (%v): %s", s.Username, req.Channel, req.Content)

	// 2. 封裝成 Envelope
	env := &pb.Envelope{
		Payload: &pb.Envelope_Chat{
			Chat: req,
		},
	}

	// 3. 根據頻道進行廣播路由
	sm := session.GetManager()

	switch req.Channel {
	case pb.ChatChannelType_CHANNEL_GLOBAL:
		// 全服廣播：轉發給所有在線 Session
		sm.AddToForwardQueue(env)

	case pb.ChatChannelType_CHANNEL_FACTION:
		// 陣營廣播：僅發送給相同 FactionID 的玩家
		broadcastToFilter(env, func(other *session.UserSession) bool {
			return other.FactionID == s.FactionID
		})

	case pb.ChatChannelType_CHANNEL_VILLAGE:
		// 庄頭廣播：僅發送給相同 VillageID 的玩家
		if s.VillageID == 0 {
			// 無庄頭玩家不能在庄頭頻道發言
			return
		}
		broadcastToFilter(env, func(other *session.UserSession) bool {
			return other.VillageID == s.VillageID
		})

	case pb.ChatChannelType_CHANNEL_REGION:
		// 區域廣播：這部分通常會結合 AOI，目前暫時以「同一個 Hive 節點」為範圍
		// TODO: 未來精細化為座標半徑範圍
		sm.AddToForwardQueue(env) // 暫時全服代替

	case pb.ChatChannelType_CHANNEL_PRIVATE:
		// 私聊：單對單推送
		if req.Receiver != "" {
			target := sm.GetSessionByUsername(req.Receiver)
			if target != nil {
				target.QueueMessage(env)
				if target.TriggerFlush != nil {
					target.TriggerFlush()
				}
				// 也要發回給發送者 (Echo)
				s.QueueMessage(env)
				if s.TriggerFlush != nil {
					s.TriggerFlush()
				}
			}
		}
	}
}

// broadcastToFilter 輔助函式：根據過濾條件進行廣播
// 注意：這在大量玩家時可能會有效能瓶頸，未來應優化為訂閱制頻道
func broadcastToFilter(env *pb.Envelope, filter func(*session.UserSession) bool) {
	sm := session.GetManager()
	// 目前 SessionManager.AddToForwardQueue 是全廣播
	// 我們暫時直接遍歷 (這是不夠 Actor 的做法，之後應優化進 SessionManager)
	// 但為了 Phase 2 快速跑通，先在此實作過濾邏輯
	
	// TODO: 優化 SessionManager 支援頻道訂閱制 (Topic-based)
	// 目前先透過 session.globalManager 的 Snapshot 功能 (雖然它是 private)
	// 我們在 SessionManager 補一個公開方法來獲取 Snapshot
	sessions := sm.GetAllSessions()
	for _, s := range sessions {
		if filter(s) {
			s.QueueMessage(env)
			if s.TriggerFlush != nil {
				go s.TriggerFlush()
			}
		}
	}
}
