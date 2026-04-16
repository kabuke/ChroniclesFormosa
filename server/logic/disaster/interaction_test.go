package disaster

import (
	"testing"

	"github.com/kabuke/ChroniclesFormosa/server/database"
	"github.com/kabuke/ChroniclesFormosa/server/model"
)

func TestEarthquakeDuringBattle(t *testing.T) {
	db := database.GetDB()
	db.Exec("DELETE FROM villages")

	// 建立兩個庄頭
	v1 := &model.Village{ID: 1, Name: "A", Food: 1000, Wood: 1000, Loyalty: 80, Soldiers: 50, TensionValue: 100}
	v2 := &model.Village{ID: 2, Name: "B", Food: 1000, Wood: 1000, Loyalty: 80, Soldiers: 50, TensionValue: 100}
	
	db.Create(v1)
	db.Create(v2)

	// In the real code, "battle" might be handled by tension/combat logic.
	// But earthquakes occur agnostically of combat state.
	// We simply trigger earthquake and verify damage applies.

	TriggerEarthquake(true) // trigger harmful earthquake

	var dbV1, dbV2 model.Village
	db.First(&dbV1, 1)
	db.First(&dbV2, 2)

	if dbV1.Food == 1000 || dbV2.Food == 1000 {
		t.Errorf("Expected food to decrease after harmful earthquake")
	}
	if dbV1.Wood == 1000 || dbV2.Wood == 1000 {
		t.Errorf("Expected wood to decrease after harmful earthquake")
	}
	if dbV1.Loyalty >= 80 || dbV2.Loyalty >= 80 {
		t.Errorf("Expected loyalty to decrease after harmful earthquake")
	}
}
