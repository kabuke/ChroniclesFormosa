package handler

import (
	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/aoi"
	"github.com/kabuke/ChroniclesFormosa/server/logic/stamina"
	"github.com/kabuke/ChroniclesFormosa/server/repo"
	"github.com/kabuke/ChroniclesFormosa/server/session"
)

// HandleAoiUpdate 攔截客戶端的移動位置報告，交給全域大區管理器派發
func HandleAoiUpdate(update *pb.ClientMoveReq, s *session.UserSession) {
	if s.Username == "" {
		return
	}

	// 1. 獲取玩家資料
	pRepo := repo.NewPlayerRepo()
	p, err := pRepo.FindByUsername(s.Username)
	if err != nil {
		return
	}

	// 2. 檢查精力值 (移動消耗 1 點)
	if !stamina.ConsumeStamina(p, 1) {
		// 精力不足，拒絕移動請求（或是回傳錯誤封包）
		return
	}

	// 3. 執行移動
	aoi.GetManager().MovePlayer(s.Username, update.X, update.Y, s)

	// 4. 更新 DB 與 同步前端
	_ = pRepo.Update(p)
	stamina.SyncStamina(s, p)
}
