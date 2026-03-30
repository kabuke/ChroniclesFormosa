package village

import (
	"log"
	"time"

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
		
		// 經濟資源產出公式: 基礎保障(10) + (村莊等級 * 人口 * 10)
		production := int64(10 + (v.Level * int32(pop) * 10))
		
		v.Wood += production
		v.Food += production
		v.Iron += production

		if err := vRepo.Update(v); err != nil {
			log.Printf("[EconomyEngine] Failed to update village %d: %v", v.ID, err)
		} else {
			// 只在有真實產出的時候印出 (避開沒人、沒升級的空莊頭瘋狂洗頻)
			if pop > 0 {
				log.Printf("[EconomyEngine] 🌾 Village '%s' produced +%d resources. (Pop: %d)", v.Name, production, pop)
			}
			totalYield += production
		}
	}

	if totalYield > 0 {
		log.Printf("[EconomyEngine] Tick finalized. Total resources generated across Formosa: %d", totalYield)
	}
}
