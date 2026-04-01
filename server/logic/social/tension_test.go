package social

import (
	"testing"

	"github.com/kabuke/ChroniclesFormosa/server/database"
	"github.com/kabuke/ChroniclesFormosa/server/model"
	"github.com/kabuke/ChroniclesFormosa/server/session"
)

func TestProcessTensionTick(t *testing.T) {
	database.InitDB(":memory:")
	db := database.GetDB()
	db.AutoMigrate(&model.Village{})
	// 初始化 Session Manager 避免廣播時崩潰
	session.GetManager() 

	v1 := model.Village{Name: "多元庄", PopMinNan: 33, PopHakka: 33, PopIndigenous: 34, Food: 100}
	v2 := model.Village{Name: "單一庄", PopMinNan: 100, Food: 100}
	db.Create(&v1)
	db.Create(&v2)

	processTensionTick()

	var r1, r2 model.Village
	db.First(&r1, "name = ?", "多元庄")
	db.First(&r2, "name = ?", "單一庄")

	if r1.TensionValue <= r2.TensionValue {
		t.Errorf("多元庄成長 (%d) 應大於單一庄 (%d)", r1.TensionValue, r2.TensionValue)
	}
}

func TestRiotTrigger(t *testing.T) {
	database.InitDB(":memory:")
	db := database.GetDB()
	db.AutoMigrate(&model.Village{})
	session.GetManager()

	v := model.Village{
		Name: "爆發庄",
		PopMinNan: 100, PopHakka: 100, PopIndigenous: 100,
		TensionValue: 100,
		Food: 1000,
	}
	db.Create(&v)

	// 直接測試 triggerRiot
	triggerRiot(&v)

	if v.PopMinNan >= 100 { t.Error("人口未減少") }
	if v.Food >= 1000 { t.Error("糧食未減少") }
	
	t.Logf("械鬥結果 - 人口: %d, 糧食: %d", v.PopMinNan, v.Food)
}
