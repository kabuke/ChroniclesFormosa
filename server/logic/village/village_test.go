package village

import (
	"os"
	"testing"

	"github.com/kabuke/ChroniclesFormosa/server/database"
	"github.com/kabuke/ChroniclesFormosa/server/model"
)

func TestMain(m *testing.M) {
	// 初始化記憶體資料庫供測試使用
	_ = database.InitDB(":memory:")
	database.GetDB().AutoMigrate(&model.Player{}, &model.Village{})
	
	code := m.Run()
	os.Exit(code)
}

func TestJoinVillage(t *testing.T) {
	db := database.GetDB()

	// 1. 準備測試資料
	v := model.Village{Name: "Test Village", Level: 1}
	db.Create(&v)

	p := model.Player{Username: "testplayer", PasswordHash: "hash", VillageID: 0}
	db.Create(&p)

	// 2. 測試正常加入
	err := JoinVillage("testplayer", v.ID)
	if err != nil {
		t.Fatalf("Expected nil error, got %v", err)
	}

	// 驗證 DB 更新
	var updatedP model.Player
	db.Where("username = ?", "testplayer").First(&updatedP)
	if updatedP.VillageID != v.ID {
		t.Errorf("Expected VillageID %d, got %d", v.ID, updatedP.VillageID)
	}

	// 3. 測試重複加入同一個
	err = JoinVillage("testplayer", v.ID)
	if err != ErrAlreadyInThis {
		t.Errorf("Expected ErrAlreadyInThis, got %v", err)
	}

	// 4. 測試加入不存在的莊頭
	err = JoinVillage("testplayer", 9999)
	if err != ErrVillageNotFound {
		t.Errorf("Expected ErrVillageNotFound, got %v", err)
	}
}

func TestGetVillageInfo(t *testing.T) {
	db := database.GetDB()
	v := model.Village{Name: "Info Village", Level: 2}
	db.Create(&v)

	// 加入兩個玩家
	db.Create(&model.Player{Username: "p1", VillageID: v.ID})
	db.Create(&model.Player{Username: "p2", VillageID: v.ID})

	info, pop, err := GetVillageInfo(v.ID)
	if err != nil {
		t.Fatalf("GetVillageInfo failed: %v", err)
	}

	if info.Name != "Info Village" {
		t.Errorf("Expected Name 'Info Village', got %s", info.Name)
	}
	if pop != 2 {
		t.Errorf("Expected population 2, got %d", pop)
	}
}
