package faction

import (
	"log"
	"sync"

	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/database"
	"github.com/kabuke/ChroniclesFormosa/server/model"
	"github.com/kabuke/ChroniclesFormosa/server/session"
	"gorm.io/gorm"
)

type FactionStats struct {
	TotalPop   int32
	BuffMultiplier float64
}

var (
	GlobalFactionStats = make(map[int32]*FactionStats)
	statsMu sync.RWMutex
)

// RebalanceFactions 重新計算全服陣營平衡狀態
func RebalanceFactions() {
	db := database.GetDB()
	
	var players []model.Player
	db.Find(&players)

	newStats := make(map[int32]*FactionStats)
	for i := int32(1); i <= 3; i++ {
		newStats[i] = &FactionStats{BuffMultiplier: 1.0}
	}

	for _, p := range players {
		if p.FactionID >= 1 && p.FactionID <= 3 {
			newStats[p.FactionID].TotalPop++
		}
	}

	totalActivePop := int32(0)
	for _, s := range newStats {
		totalActivePop += s.TotalPop
	}

	// NPC 補位與天命加成
	for fid, s := range newStats {
		if s.TotalPop == 0 {
			// 某陣營完全沒人，NPC 接管並給予強力 Buff
			s.BuffMultiplier = 1.5
			log.Printf("[Balance] 🤖 NPC Tide: Faction %d is now controlled by AI (1.5x Buff)", fid)
			broadcastSystemMsg(fid, "【天命】某勢力獲得海外援軍（NPC）支持，實力大幅提升！")
		} else if totalActivePop > 10 {
			// 人口最少但有人，給予 1.2x 加成
			isMin := true
			for otherID, otherS := range newStats {
				if otherID != fid && otherS.TotalPop < s.TotalPop {
					isMin = false; break
				}
			}
			if isMin {
				s.BuffMultiplier = 1.2
			}
		}
	}

	checkInternalStrife(db, newStats)

	statsMu.Lock()
	GlobalFactionStats = newStats
	statsMu.Unlock()

	broadcastBuffs(newStats)
}

func broadcastSystemMsg(factionID int32, content string) {
	env := &pb.Envelope{
		Payload: &pb.Envelope_Chat{
			Chat: &pb.ChatMessage{
				Channel: pb.ChatChannelType_CHANNEL_GLOBAL,
				Sender:  "廟口說書人",
				Content: content,
			},
		},
	}
	session.GetManager().AddToForwardQueue(env)
}

func broadcastBuffs(stats map[int32]*FactionStats) {
	sm := session.GetManager()
	sessions := sm.GetAllSessions()
	for _, s := range sessions {
		if s.FactionID == 0 { continue }
		multiplier := stats[s.FactionID].BuffMultiplier
		
		env := &pb.Envelope{
			Payload: &pb.Envelope_FactionBuff{
				FactionBuff: &pb.FactionBuffSync{
					Multiplier: multiplier,
				},
			},
		}
		s.QueueMessage(env)
		if s.TriggerFlush != nil { go s.TriggerFlush() }
	}
}

func checkInternalStrife(db *gorm.DB, stats map[int32]*FactionStats) {
	var villages []model.Village
	db.Find(&villages)

	for i := range villages {
		v := &villages[i]
		if v.FactionID == 0 { continue }
		factionPop := stats[v.FactionID].TotalPop
		if factionPop == 0 { continue }

		var vPop int64
		db.Model(&model.Player{}).Where("village_id = ?", v.ID).Count(&vPop)

		ratio := float64(vPop) / float64(factionPop)
		if ratio > 0.40 {
			v.TensionValue += 5
			if v.TensionValue > 100 { v.TensionValue = 100 }
			db.Save(v)
		}
	}
}

func GetBuffMultiplier(factionID int32) float64 {
	statsMu.RLock()
	defer statsMu.RUnlock()
	if s, ok := GlobalFactionStats[factionID]; ok { return s.BuffMultiplier }
	return 1.0
}
