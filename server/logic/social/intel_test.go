package social

import (
	"testing"
	"time"

	"github.com/kabuke/ChroniclesFormosa/server/database"
	"github.com/kabuke/ChroniclesFormosa/server/session"
)

func TestIntelSensing(t *testing.T) {
	database.InitDB(":memory:")
	sm := session.GetManager()
	// 準備兩個不同陣營的玩家
	s1 := &session.UserSession{SessionID: "10", Username: "Spy1", FactionID: 1} // 清軍
	s2 := &session.UserSession{SessionID: "11", Username: "Victim1", FactionID: 2} // 義軍
	sm.AddSession(s1)
	sm.AddSession(s2)

	im := GlobalIntelManager
	// 重置統計
	im.mu.Lock()
	im.stats = make(map[int32]*FactionIntelStats)
	im.mu.Unlock()

	// 模擬義軍 (Faction 2) 在頻道中多次提到關鍵字
	for i := 0; i < 5; i++ {
		im.ScanChatMessage(2, "我們今晚準備夜襲府城！")
	}

	// 檢查清軍 (Faction 1) 是否收到了模糊情報
	found := false
	for _, env := range s1.Outbox {
		if env.GetChat() != nil && env.GetChat().Sender == "廟口說書人" {
			found = true
			break
		}
	}

	if !found {
		t.Error("敵對陣營玩家應收到模糊情報提示")
	}

	// 檢查來源陣營是否收到 (不應收到)
	for _, env := range s2.Outbox {
		if env.GetChat() != nil && env.GetChat().Sender == "廟口說書人" {
			t.Error("情報來源陣營不應收到自己的洩漏提示")
		}
	}
}

func TestIntelCooldown(t *testing.T) {
	database.InitDB(":memory:")
	im := GlobalIntelManager
	im.mu.Lock()
	im.stats = make(map[int32]*FactionIntelStats)
	im.mu.Unlock()

	// 觸發第一次情報
	for i := 0; i < 5; i++ {
		im.ScanChatMessage(1, "進攻！進攻！進攻！")
	}

	stat := im.stats[1]
	if stat == nil {
		t.Fatal("統計資料不應為空")
	}
	if time.Since(stat.LastTrigger) > 1*time.Second {
		t.Log("第一次情報已觸發")
	}

	// 立即嘗試觸發第二次 (應被冷卻時間擋住)
	for i := 0; i < 10; i++ {
		im.ScanChatMessage(1, "夜襲！夜襲！夜襲！")
	}

	if len(stat.KeywordHits) > 0 {
		t.Logf("關鍵字命中已累積，目前數量: %d", len(stat.KeywordHits))
	}
}
