package social

import (
	"testing"

	"github.com/kabuke/ChroniclesFormosa/server/database"
	"github.com/kabuke/ChroniclesFormosa/server/model"
	"github.com/kabuke/ChroniclesFormosa/server/session"
	pb "github.com/kabuke/ChroniclesFormosa/resource"
)

func TestHandleDiplomacyReconcile(t *testing.T) {
	database.InitDB(":memory:")
	db := database.GetDB()
	db.AutoMigrate(&model.Player{}, &model.Village{})
	session.GetManager()

	// 1. 準備庄頭與玩家 (庄長)
	v := model.Village{ID: 1, Name: "測試庄", Food: 200, Loyalty: 50, Soldiers: 0}
	db.Create(&v)
	p := model.Player{Username: "headman", VillageID: 1}
	db.Create(&p)

	s := &session.UserSession{Username: "headman", VillageID: 1}

	// 2. 執行理番請求
	req := &pb.DiplomacyReq{
		Type: pb.DiplomacyType_DIPLO_RECONCILE,
		TargetVillageId: 1, // 雖然理番通常對外，目前邏輯是作用於自身庄頭
	}

	msg, err := HandleDiplomacyRequest(s, req)
	if err != nil {
		t.Fatalf("外交請求失敗: %v", err)
	}

	t.Logf("外交回應: %s", msg)

	// 3. 驗證資源扣除與獎勵
	var updatedV model.Village
	db.First(&updatedV, 1)

	if updatedV.Food != 100 { t.Errorf("糧食應扣除 100，目前為: %d", updatedV.Food) }
	if updatedV.Loyalty <= 50 { t.Errorf("民忠應提升，目前為: %d", updatedV.Loyalty) }
	if updatedV.Soldiers != 50 { t.Errorf("應獲得 50 名山林勇士，目前為: %d", updatedV.Soldiers) }
}
