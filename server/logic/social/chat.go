package social

import (
	"time"

	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/session"
)

// HandleChatSend 處理聊天發送邏輯
func HandleChatSend(s *session.UserSession, req *pb.ChatMessage) {
	req.Sender = s.Username
	req.Timestamp = time.Now().Unix()

	env := &pb.Envelope{
		Payload: &pb.Envelope_Chat{
			Chat: req,
		},
	}

	sm := session.GetManager()

	switch req.Channel {
	case pb.ChatChannelType_CHANNEL_GLOBAL:
		sm.AddToForwardQueue(env)

	case pb.ChatChannelType_CHANNEL_FACTION:
		broadcastToFilter(env, func(other *session.UserSession) bool {
			return other.FactionID == s.FactionID
		})

	case pb.ChatChannelType_CHANNEL_VILLAGE:
		if s.VillageID == 0 {
			sendSystemPrivate(s, "您尚未加入任何庄頭，無法在庄頭頻道發言。")
			return
		}
		broadcastToFilter(env, func(other *session.UserSession) bool {
			return other.VillageID == s.VillageID
		})

	case pb.ChatChannelType_CHANNEL_REGION:
		sm.AddToForwardQueue(env)

	case pb.ChatChannelType_CHANNEL_PRIVATE:
		if req.Receiver != "" {
			target := sm.GetSessionByUsername(req.Receiver)
			if target != nil {
				target.QueueMessage(env)
				if target.TriggerFlush != nil { go target.TriggerFlush() }
				
				// 僅在非發送給自己的情況下回顯給發送者
				if target.SessionID != s.SessionID {
					s.QueueMessage(env)
					if s.TriggerFlush != nil { go s.TriggerFlush() }
				}
			}
		}
	}
}

func broadcastToFilter(env *pb.Envelope, filter func(*session.UserSession) bool) {
	sm := session.GetManager()
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

func sendSystemPrivate(s *session.UserSession, content string) {
	env := &pb.Envelope{
		Payload: &pb.Envelope_Chat{
			Chat: &pb.ChatMessage{
				Channel: pb.ChatChannelType_CHANNEL_PRIVATE,
				Sender:  "【系統】",
				Content: content,
			},
		},
	}
	s.QueueMessage(env)
	if s.TriggerFlush != nil { go s.TriggerFlush() }
}
