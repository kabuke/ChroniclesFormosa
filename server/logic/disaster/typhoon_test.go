package disaster

import (
	"testing"
	"github.com/kabuke/ChroniclesFormosa/server/database"
	"github.com/kabuke/ChroniclesFormosa/server/model"
)

func TestTyphoonAffectedCountAndFoodDamage(t *testing.T) {
	db := database.GetDB()
	
	// Clear villages
	db.Exec("DELETE FROM villages")

	// Insert 10 villages
	for i := 0; i < 10; i++ {
		db.Create(&model.Village{Food: 1000})
	}

	// Trigger Typhoon
	TriggerTyphoon()

	var villages []model.Village
	db.Find(&villages)

	affectedCount := 0
	for _, v := range villages {
		if v.Food == 500 {
			affectedCount++
		} else if v.Food != 1000 {
			t.Errorf("Unexpected food amount: %d", v.Food)
		}
	}

	if affectedCount < 3 || affectedCount > 6 {
		t.Errorf("Affected count out of bounds: %d", affectedCount)
	}

	// Test case: limited villages
	db.Exec("DELETE FROM villages")
	db.Create(&model.Village{Food: 1000})
	db.Create(&model.Village{Food: 1000})

	TriggerTyphoon()
	db.Find(&villages)

	affectedCount2 := 0
	for _, v := range villages {
		if v.Food == 500 {
			affectedCount2++
		}
	}

	if affectedCount2 != 2 {
		t.Errorf("Expected exactly 2 villages affected when total is 2, got %d", affectedCount2)
	}
}
