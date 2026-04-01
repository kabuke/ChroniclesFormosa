package social

import (
	"testing"
	"time"

	"github.com/kabuke/ChroniclesFormosa/server/database"
	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/session"
)

func TestHandleChatSend_Global(t *testing.T) {
	database.InitDB(":memory:")
	sm := session.GetManager()
	s1 := &session.UserSession{SessionID: "1", Username: "Alice"}
	sm.AddSession(s1)
	time.Sleep(10 * time.Millisecond)

	req := &pb.ChatMessage{
		Channel: pb.ChatChannelType_CHANNEL_GLOBAL,
		Content: "Hello World",
	}

	HandleChatSend(s1, req)
}

func TestHandleChatSend_Faction(t *testing.T) {
	database.InitDB(":memory:")
	sm := session.GetManager()
	s1 := &session.UserSession{SessionID: "2", Username: "Alice", FactionID: 1}
	s2 := &session.UserSession{SessionID: "3", Username: "Bob", FactionID: 1}
	s3 := &session.UserSession{SessionID: "4", Username: "Charlie", FactionID: 2}

	sm.AddSession(s1)
	sm.AddSession(s2)
	sm.AddSession(s3)
	time.Sleep(10 * time.Millisecond)

	req := &pb.ChatMessage{
		Channel: pb.ChatChannelType_CHANNEL_FACTION,
		Content: "Faction Message",
	}

	HandleChatSend(s1, req)

	if len(s2.Outbox) == 0 { t.Error("同陣營玩家應收到訊息") }
	if len(s3.Outbox) != 0 { t.Error("不同陣營玩家不應收到訊息") }
}

func TestHandleChatSend_Village(t *testing.T) {
	database.InitDB(":memory:")
	sm := session.GetManager()
	s1 := &session.UserSession{SessionID: "5", Username: "V1", VillageID: 100}
	s2 := &session.UserSession{SessionID: "6", Username: "V2", VillageID: 100}
	s3 := &session.UserSession{SessionID: "7", Username: "V3", VillageID: 101}

	sm.AddSession(s1)
	sm.AddSession(s2)
	sm.AddSession(s3)
	time.Sleep(10 * time.Millisecond)

	req := &pb.ChatMessage{
		Channel: pb.ChatChannelType_CHANNEL_VILLAGE,
		Content: "Village Message",
	}

	HandleChatSend(s1, req)

	if len(s2.Outbox) == 0 { t.Error("同庄頭玩家應收到訊息") }
	if len(s3.Outbox) != 0 { t.Error("不同庄頭玩家不應收到訊息") }
}

func TestHandleChatSend_Private(t *testing.T) {
	database.InitDB(":memory:")
	sm := session.GetManager()
	s1 := &session.UserSession{SessionID: "8", Username: "Sender"}
	s2 := &session.UserSession{SessionID: "9", Username: "Receiver"}

	sm.AddSession(s1)
	sm.AddSession(s2)
	time.Sleep(10 * time.Millisecond)

	req := &pb.ChatMessage{
		Channel: pb.ChatChannelType_CHANNEL_PRIVATE,
		Receiver: "Receiver",
		Content: "Secret",
	}

	HandleChatSend(s1, req)

	if len(s2.Outbox) == 0 { t.Error("私訊接收者應收到訊息") }
}
