package social

import (
	"log"
	"math"
	"time"

	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/model"
	"github.com/kabuke/ChroniclesFormosa/server/repo"
	"github.com/kabuke/ChroniclesFormosa/server/session"
)

// StartTensionEngine 啟動全服族群緊張儀計算循環
func StartTensionEngine() {
	// 每 30 秒計算一次漂移
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	log.Println("[TensionEngine] 🧨 Ethnic Tension Engine Started. Tick: 30s.")

	for range ticker.C {
		tickTension()
	}
}

func tickTension() {
	vRepo := repo.NewVillageRepo()
	villages, err := vRepo.FindAll()
	if err != nil {
		return
	}

	for _, v := range villages {
		oldValue := v.TensionValue
		
		// 1. 計算增量
		delta := calculateTensionDelta(v)
		v.TensionValue += delta

		// 限制在 0-100
		if v.TensionValue < 0 { v.TensionValue = 0 }
		if v.TensionValue > 100 { v.TensionValue = 100 }

		// 2. 處理爆發 (分類械鬥)
		if v.TensionValue >= 100 {
			triggerRiot(v)
		}

		// 3. 存檔與廣播 (若有變化)
		if v.TensionValue != oldValue {
			_ = vRepo.Update(v)
			broadcastTension(v)
		}
	}
}

func calculateTensionDelta(v *model.Village) int32 {
	totalPop := float64(v.PopMinNan + v.PopHakka + v.PopIndigenous)
	if totalPop < 10 {
		return -1 // 人煙稀少，自然降溫
	}

	delta := 0.0

	// A. 族群競爭效應 (Entropy-based or Ratio-based)
	// 如果兩族群人數接近，緊張度上升
	r1 := float64(v.PopMinNan) / totalPop
	r2 := float64(v.PopHakka) / totalPop
	r3 := float64(v.PopIndigenous) / totalPop

	// 使用標準差的倒數來模擬「接近程度」
	avg := 0.333
	variance := ((r1-avg)*(r1-avg) + (r2-avg)*(r2-avg) + (r3-avg)*(r3-avg)) / 3
	stdDev := math.Sqrt(variance)

	// stdDev 越小代表族群越勢均力敵，容易發生衝突
	if stdDev < 0.15 {
		delta += 2.0
	}

	// B. 糧食壓力
	if v.Food < 100 {
		delta += 1.5
	}

	// C. 穩定度/治安抑制
	stabilityFactor := float64(v.Stability) / 100.0 // 0~1
	delta -= stabilityFactor * 1.5

	return int32(math.Round(delta))
}

func triggerRiot(v *model.Village) {
	log.Printf("[TensionEngine] 💥 RIOT TRIGGERED in '%s'!", v.Name)
	
	// 懲罰：隨機減少各族群人口 10%
	v.PopMinNan = int32(float64(v.PopMinNan) * 0.9)
	v.PopHakka = int32(float64(v.PopHakka) * 0.9)
	v.PopIndigenous = int32(float64(v.PopIndigenous) * 0.9)
	
	// 資源損毀
	v.Food /= 2
	v.Wood /= 2
	
	// 事件發生後降溫
	v.TensionValue = 40
	v.Stability = 30 // 治安大降
	
	// 發送全服/全庄頭廣播
	env := &pb.Envelope{
		Payload: &pb.Envelope_Chat{
			Chat: &pb.ChatMessage{
				Channel: pb.ChatChannelType_CHANNEL_GLOBAL,
				Sender:  "【系統公告】",
				Content: "「分類械鬥爆發！」" + v.Name + " 境內各族群爆發武裝衝突，傷亡慘重。",
			},
		},
	}
	session.GetManager().AddToForwardQueue(env)
}

func broadcastTension(v *model.Village) {
	level := "PEACE"
	if v.TensionValue > 80 {
		level = "RIOT"
	} else if v.TensionValue > 60 {
		level = "TENSE"
	} else if v.TensionValue > 30 {
		level = "UNEASY"
	}

	env := &pb.Envelope{
		Payload: &pb.Envelope_Tension{
			Tension: &pb.TensionUpdate{
				VillageId:    v.ID,
				TensionValue: v.TensionValue,
				VisualLevel:  level,
			},
		},
	}

	// 只廣播給該庄頭的玩家
	sm := session.GetManager()
	sessions := sm.GetAllSessions()
	for _, s := range sessions {
		if s.VillageID == v.ID {
			s.QueueMessage(env)
			if s.TriggerFlush != nil {
				go s.TriggerFlush()
			}
		}
	}
}
