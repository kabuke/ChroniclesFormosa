package faction

import (
	"fmt"
	"testing"

	"github.com/kabuke/ChroniclesFormosa/server/database"
	"github.com/kabuke/ChroniclesFormosa/server/model"
)

func setupDB() {
	_ = database.InitDB(":memory:")
	db := database.GetDB()
	db.AutoMigrate(&model.Player{}, &model.Village{})
	// 清空表數據，確保測試獨立性
	db.Exec("DELETE FROM players")
	db.Exec("DELETE FROM villages")
}

func TestRebalanceFactions(t *testing.T) {
	setupDB()
	
	// 1. 模擬沒人的陣營 (NPC 接管)
	RebalanceFactions()
	
	m1 := GetBuffMultiplier(1)
	if m1 != 1.5 { t.Errorf("無人陣營應有 1.5x NPC Buff, 目前: %f", m1) }

	// 2. 模擬有人但人數最少
	database.GetDB().Create(&model.Player{Username: "f1_1", FactionID: 1}) // 1人
	database.GetDB().Create(&model.Player{Username: "f2_1", FactionID: 2}) // 2人
	database.GetDB().Create(&model.Player{Username: "f2_2", FactionID: 2})
	database.GetDB().Create(&model.Player{Username: "f3_1", FactionID: 3}) // 3人
	database.GetDB().Create(&model.Player{Username: "f3_2", FactionID: 3})
	database.GetDB().Create(&model.Player{Username: "f3_3", FactionID: 3})

	// 為了觸發 1.2x，總人口多加一些 filler
	for i := 0; i < 10; i++ {
		database.GetDB().Create(&model.Player{Username: fmt.Sprintf("filler_%d", i), FactionID: 3})
	}

	RebalanceFactions()
	
	m_min := GetBuffMultiplier(1)
	if m_min != 1.2 { t.Errorf("人數最少陣營應有 1.2x Buff, 目前: %f", m_min) }
}

func TestCheckInternalStrife(t *testing.T) {
	setupDB()
	db := database.GetDB()

	// 準備一個陣營，裡面只有一個大庄頭佔比過高 (> 40%)
	v := model.Village{Name: "權臣庄", FactionID: 1, TensionValue: 50}
	db.Create(&v)
	
	// Faction 1 總人口 2 人，此庄佔 1 人 = 50%
	db.Create(&model.Player{Username: "boss", FactionID: 1, VillageID: v.ID})
	db.Create(&model.Player{Username: "pleb", FactionID: 1, VillageID: 0})

	RebalanceFactions()

	var updatedV model.Village
	db.First(&updatedV, v.ID)
	if updatedV.TensionValue <= 50 {
		t.Errorf("佔比過高庄頭應增加緊張值 (ID:%d), 目前: %d", v.ID, updatedV.TensionValue)
	}
}
