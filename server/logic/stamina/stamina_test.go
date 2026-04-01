package stamina

import (
	"testing"

	"github.com/kabuke/ChroniclesFormosa/server/database"
	"github.com/kabuke/ChroniclesFormosa/server/model"
)

func TestStaminaConsumption(t *testing.T) {
	p := &model.Player{Username: "tester", Stamina: 50}
	
	if !ConsumeStamina(p, 20) {
		t.Error("應可扣除足夠的精力")
	}
	if p.Stamina != 30 {
		t.Errorf("剩餘精力錯誤: %d", p.Stamina)
	}

	if ConsumeStamina(p, 40) {
		t.Error("精力不足時不應扣除成功")
	}
}

func TestRestoreAll(t *testing.T) {
	database.InitDB(":memory:")
	db := database.GetDB()
	db.AutoMigrate(&model.Player{})

	db.Create(&model.Player{Username: "p1", Stamina: 90})
	db.Create(&model.Player{Username: "p2", Stamina: 98})

	RestoreAll(5)

	var r1, r2 model.Player
	db.Where("username = ?", "p1").First(&r1)
	db.Where("username = ?", "p2").First(&r2)

	if r1.Stamina != 95 { t.Errorf("p1 恢復後應為 95, 實得: %d", r1.Stamina) }
	if r2.Stamina != 100 { t.Errorf("p2 恢復後應上限為 100, 實得: %d", r2.Stamina) }
}
