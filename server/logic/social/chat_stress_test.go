package social

import (
	"fmt"
	"testing"
	"time"

	"github.com/kabuke/ChroniclesFormosa/server/database"
	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/session"
)

func TestChatStress_GlobalBroadcast(t *testing.T) {
	database.InitDB(":memory:")
	sm := session.GetManager()
	count := 1000

	// 1. 建立 1000 個 Session
	for i := 0; i < count; i++ {
		username := fmt.Sprintf("user%d", i)
		s := &session.UserSession{
			SessionID: fmt.Sprintf("stress_%d", i),
			Username:  username,
		}
		sm.AddSession(s)
	}

	// 2. 發送一條全域訊息
	s0 := sm.GetSessionByUsername("user0")
	req := &pb.ChatMessage{
		Channel: pb.ChatChannelType_CHANNEL_GLOBAL,
		Content: "Stress Test Message",
	}

	start := time.Now()
	HandleChatSend(s0, req)
	
	t.Logf("1000 個 Session 的全域訊息處理耗時: %v", time.Since(start))
}

func TestChatStress_FactionBroadcast(t *testing.T) {
	database.InitDB(":memory:")
	sm := session.GetManager()
	count := 1000

	// 1. 建立 1000 個同陣營 Session
	for i := 0; i < count; i++ {
		username := fmt.Sprintf("fuser%d", i)
		s := &session.UserSession{
			SessionID: fmt.Sprintf("fstress_%d", i),
			Username:  username,
			FactionID: 1,
		}
		sm.AddSession(s)
	}

	// 2. 進行陣營廣播 (立即廣播模式)
	s0 := sm.GetSessionByUsername("fuser0")
	req := &pb.ChatMessage{
		Channel: pb.ChatChannelType_CHANNEL_FACTION,
		Content: "Faction Stress Test",
	}

	start := time.Now()
	HandleChatSend(s0, req)
	elapsed := time.Since(start)

	t.Logf("1000 個同陣營 Session 的廣播耗時: %v", elapsed)
	
	if elapsed > 100*time.Millisecond {
		t.Errorf("廣播耗時過長: %v", elapsed)
	}
}
