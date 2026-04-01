package social

import (
	"errors"

	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/database"
	"github.com/kabuke/ChroniclesFormosa/server/model"
	"github.com/kabuke/ChroniclesFormosa/server/repo"
	"github.com/kabuke/ChroniclesFormosa/server/session"
)

var (
	ErrVillageRequired = errors.New("此動作需要加入庄頭才能執行")
)

// HandleDiplomacyRequest 處理外交邏輯
func HandleDiplomacyRequest(s *session.UserSession, req *pb.DiplomacyReq) (string, error) {
	pRepo := repo.NewPlayerRepo()
	vRepo := repo.NewVillageRepo()

	p, err := pRepo.FindByUsername(s.Username)
	if err != nil {
		return "", err
	}

	if p.VillageID == 0 && req.Type != pb.DiplomacyType_DIPLO_BLOOD_BROTHER {
		return "", ErrVillageRequired
	}

	// 目標庄頭
	targetV, err := vRepo.FindByID(req.TargetVillageId)
	if err != nil || targetV == nil {
		return "", errors.New("目標庄頭不存在")
	}

	switch req.Type {
	case pb.DiplomacyType_DIPLO_ALLIANCE, pb.DiplomacyType_DIPLO_MARRIAGE:
		if targetV.Headman == "" {
			return "", errors.New("該庄頭目前無庄長，無法進行外交洽談")
		}

		targetS := session.GetManager().GetSessionByUsername(targetV.Headman)
		if targetS == nil {
			return "", errors.New("目標庄長目前不在線")
		}

		// 推送請求給目標庄長
		env := &pb.Envelope{
			Payload: &pb.Envelope_Diplomacy{
				Diplomacy: &pb.DiplomacyAction{
					Action: &pb.DiplomacyAction_Req{
						Req: &pb.DiplomacyReq{
							Type:            req.Type,
							TargetVillageId: p.VillageID,
							TargetPlayerId:  int64(p.ID),
						},
					},
				},
			},
		}
		targetS.QueueMessage(env)
		if targetS.TriggerFlush != nil { go targetS.TriggerFlush() }

		return "外交請求已送達，等待對方回覆中...", nil

	case pb.DiplomacyType_DIPLO_RECONCILE:
		// 理番和議：消耗 100 糧食 (模擬換取鹽布)，提升民忠與產出
		v, _ := vRepo.FindByID(p.VillageID)
		if v != nil {
			if v.Food < 100 { return "", errors.New("資源不足 (需 100 糧食模擬鹽布)") }
			v.Food -= 100
			v.Loyalty += 15
			v.Soldiers += 50 // 獲得 50 名山林勇士
			if v.Loyalty > 100 { v.Loyalty = 100 }
			_ = vRepo.Update(v)
		}
		return "理番和議達成，獲得山林勇士效忠與物資交換。", nil

	default:
		return "外交官正在洽談中...", nil
	}
}

// IsAllied 檢查兩庄頭是否為盟友
func IsAllied(v1, v2 int64) bool {
	if v1 == v2 { return true }
	var count int64
	database.GetDB().Model(&model.DiplomacyRelation{}).
		Where("type = ? AND ((source_id = ? AND target_id = ?) OR (source_id = ? AND target_id = ?))", 
			int32(pb.DiplomacyType_DIPLO_ALLIANCE), v1, v2, v2, v1).
		Count(&count)
	return count > 0
}
