package social

import (
	"log"
	"strings"
	"sync"
	"time"

	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/session"
)

// SensitiveKeywords 定義會觸發情報感測的關鍵字及其權重
var SensitiveKeywords = map[string]float32{
	"出兵": 2.0, "進攻": 2.0, "夜襲": 3.0, "偷襲": 2.5,
	"府城": 1.5, "打狗": 1.5, "諸羅": 1.5, "竹塹": 1.5,
	"北上": 1.0, "南下": 1.0, "集合": 1.0, "開戰": 2.0,
}

type FactionIntelStats struct {
	KeywordHits map[string]int
	LastTrigger time.Time
}

type IntelManager struct {
	// 記錄每個陣營最近 5 分鐘內的關鍵字命中情況
	// key: FactionID
	stats map[int32]*FactionIntelStats
	mu    sync.Mutex
}

var GlobalIntelManager = &IntelManager{
	stats: make(map[int32]*FactionIntelStats),
}

// ScanChatMessage 掃描聊天訊息並更新統計
func (m *IntelManager) ScanChatMessage(factionID int32, content string) {
	if factionID == 0 {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	stat, ok := m.stats[factionID]
	if !ok {
		stat = &FactionIntelStats{
			KeywordHits: make(map[string]int),
			LastTrigger: time.Now().Add(-10 * time.Minute),
		}
		m.stats[factionID] = stat
	}

	// 1. 關鍵字比對
	found := false
	for kw := range SensitiveKeywords {
		if strings.Contains(content, kw) {
			stat.KeywordHits[kw]++
			found = true
		}
	}

	if !found {
		return
	}

	// 2. 判斷是否觸發情報生成 (冷卻時間 2 分鐘，防止情報刷屏)
	if time.Since(stat.LastTrigger) > 2*time.Minute {
		totalHits := 0
		for _, hits := range stat.KeywordHits {
			totalHits += hits
		}

		if totalHits >= 5 {
			m.generateAndBroadcastIntel(factionID, stat)
			// 重置統計與觸發時間
			stat.KeywordHits = make(map[string]int)
			stat.LastTrigger = time.Now()
		}
	}
}

// generateAndBroadcastIntel 生成模糊情報並推播給敵對陣營
func (m *IntelManager) generateAndBroadcastIntel(sourceFaction int32, stat *FactionIntelStats) {
	// 1. 決定受眾 (非來源陣營的所有玩家)
	// 2. 構造模糊情報
	factionName := "某勢力"
	switch sourceFaction {
	case 1: factionName = "清軍"
	case 2: factionName = "義軍"
	case 3: factionName = "原民"
	}

	intelText := "【廟口說書人】傳聞：" + factionName + "內部近期頻繁提及軍事行動，似乎正圖謀大事..."
	
	log.Printf("[IntelEngine] 📢 Intelligence Generated for Faction %d: %s", sourceFaction, intelText)

	env := &pb.Envelope{
		Payload: &pb.Envelope_Chat{
			Chat: &pb.ChatMessage{
				Channel: pb.ChatChannelType_CHANNEL_GLOBAL,
				Sender:  "廟口說書人",
				Content: intelText,
			},
		},
	}

	// 只廣播給敵對陣營的玩家
	sm := session.GetManager()
	sessions := sm.GetAllSessions()
	for _, s := range sessions {
		// 不發送給情報來源陣營的人，增加「被偷聽」的感覺
		if s.FactionID != sourceFaction && s.FactionID != 0 {
			s.QueueMessage(env)
			if s.TriggerFlush != nil {
				go s.TriggerFlush()
			}
		}
	}
}
