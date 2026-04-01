package village

import (
	"log"
	"time"

	"github.com/kabuke/ChroniclesFormosa/server/logic/faction"
	"github.com/kabuke/ChroniclesFormosa/server/repo"
)

// StartEconomyEngine 啟動全服村莊資源產出的背景循環 (World Engine Core)
func StartEconomyEngine() {
	// Phase 1 測試用：每一分鐘結算一次
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Println("[EconomyEngine] ⚙️ World Economy Engine Started. Tick: 1 minute.")

	for range ticker.C {
		settleEconomy()
	}
}

func settleEconomy() {
	log.Println("[EconomyEngine] ⚖️ Calculating Faction Balance...")
	faction.RebalanceFactions()

	vRepo := repo.NewVillageRepo()
	villages, err := vRepo.FindAll()
	if err != nil {
		log.Printf("[EconomyEngine] Error fetching villages: %v", err)
		return
	}

	pRepo := repo.NewPlayerRepo()
	var totalYield int64

	for _, v := range villages {
		pop, _ := pRepo.CountByVillageID(v.ID)
		
		// 基礎經濟資源產出公式: 基礎保障(10) + (村莊等級 * 人口 * 10)
		production := int64(10 + (v.Level * int32(pop) * 10))
		
		// 應用陣營加成 (Phase 2: 天命加成)
		multiplier := faction.GetBuffMultiplier(v.FactionID)
		finalProd := int64(float64(production) * multiplier)

		v.Wood += finalProd
		v.Food += finalProd
		v.Iron += finalProd

		if err := vRepo.Update(v); err != nil {
			log.Printf("[EconomyEngine] Failed to update village %d: %v", v.ID, err)
		} else {
			if pop > 0 {
				log.Printf("[EconomyEngine] 🌾 Village '%s' produced +%d resources. (Pop: %d, Buff: %.1fx)", v.Name, finalProd, pop, multiplier)
			}
			totalYield += finalProd
		}
	}

	if totalYield > 0 {
		log.Printf("[EconomyEngine] Tick finalized. Total resources generated across Formosa: %d", totalYield)
	}
}
