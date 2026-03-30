package handler

import (
	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/aoi"
	"github.com/kabuke/ChroniclesFormosa/server/session"
)

// HandleAoiUpdate 攔截客戶端的移動位置報告，交給全域大區管理器派發
func HandleAoiUpdate(update *pb.ClientMoveReq, s *session.UserSession) {
	// 如果還沒登入，就不允許把假座標送進伺服器混淆 (防呆機制)
	if s.Username == "" {
		return
	}
	
	aoi.GetManager().MovePlayer(s.Username, update.X, update.Y, s)
}
