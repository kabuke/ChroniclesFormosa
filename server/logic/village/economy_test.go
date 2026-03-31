package village

import (
	"testing"

	"github.com/kabuke/ChroniclesFormosa/server/database"
	"github.com/kabuke/ChroniclesFormosa/server/model"
)

func TestSettleEconomy(t *testing.T) {
	db := database.GetDB()

	// 清空資料
	db.Exec("DELETE FROM villages")
	db.Exec("DELETE FROM players")

	// 1. 準備資料：莊頭 A (等級 1, 人口 2)
	vA := model.Village{Name: "Village A", Level: 1, Wood: 0, Food: 0, Iron: 0}
	db.Create(&vA)
	db.Create(&model.Player{Username: "user1", VillageID: vA.ID})
	db.Create(&model.Player{Username: "user2", VillageID: vA.ID})

	// 2. 執行結算
	settleEconomy()

	// 3. 驗證產出
	// 公式: 10 + (Level * Pop * 10) = 10 + (1 * 2 * 10) = 30
	var updatedVA model.Village
	db.First(&updatedVA, vA.ID)
	
	expected := int64(30)
	if updatedVA.Wood != expected {
		t.Errorf("Expected Wood %d, got %d", expected, updatedVA.Wood)
	}
	if updatedVA.Food != expected {
		t.Errorf("Expected Food %d, got %d", expected, updatedVA.Food)
	}
	if updatedVA.Iron != expected {
		t.Errorf("Expected Iron %d, got %d", expected, updatedVA.Iron)
	}
}
