package disaster

import (
	"fmt"
	"log"

	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/logic/stamina"
	"github.com/kabuke/ChroniclesFormosa/server/repo"
	"github.com/kabuke/ChroniclesFormosa/server/session"
)

// HandleReliefDonate 處理庄民捐獻精力
func HandleReliefDonate(s *session.UserSession, req *pb.ReliefDonateReq) (*pb.Envelope, error) {
	if s.VillageID == 0 {
		return nil, fmt.Errorf("未加入庄頭")
	}

	pRepo := repo.NewPlayerRepo()
	p, err := pRepo.FindByUsername(s.Username)
	if err != nil { return nil, err }

	// 檢查精力並扣除
	if !stamina.ConsumeStamina(p, req.ResourceAmount) {
		return nil, fmt.Errorf("精力值不足")
	}
	_ = pRepo.Update(p)
	stamina.SyncStamina(s, p)

	log.Printf("[Relief] 玩家 %s 捐獻了 %d 精力用於救災", p.Nickname, req.ResourceAmount)

	// 這裡可以把捐獻紀錄存起來，或是透過頻道通知庄長
	return nil, nil // 捐獻成功，不一定要立即回傳特定 UI，可以透過 Toast 處理
}

// HandleReliefRouteSubmit 處理庄長提交牛車路線
func HandleReliefRouteSubmit(s *session.UserSession, req *pb.ReliefRouteSubmit) (*pb.Envelope, error) {
	if s.VillageID == 0 {
		return nil, fmt.Errorf("未加入庄頭")
	}

	vRepo := repo.NewVillageRepo()
	v, err := vRepo.FindByID(s.VillageID)
	if err != nil { return nil, err }

	if v.Headman != s.Username {
		return nil, fmt.Errorf("只有庄長能提交救災路線")
	}

	// 1. 評分演算法：根據路線節點數量 (越少越好) 與 覆蓋的災區數量
	waypoints := req.Waypoints
	score := 100
	if len(waypoints) > 10 {
		score -= (len(waypoints) - 10) * 5
	}
	if score < 0 { score = 0 }

	var grade pb.ReliefGrade
	if score >= 90 {
		grade = pb.ReliefGrade_GRADE_PERFECT
	} else if score >= 60 {
		grade = pb.ReliefGrade_GRADE_GOOD
	} else {
		grade = pb.ReliefGrade_GRADE_FAIL
	}

	// 2. 獎勵計算
	reward := int32(score * 10) // 給予糧食或銀兩
	if grade == pb.ReliefGrade_GRADE_PERFECT {
		log.Printf("[Relief] 庄長 %s 完美救災！獲得額外 BUFF", s.Username)
		v.Food += int64(reward)
		v.Loyalty += 20
		if v.Loyalty > 100 { v.Loyalty = 100 }
	} else if grade == pb.ReliefGrade_GRADE_GOOD {
		v.Food += int64(reward / 2)
		v.Loyalty += 10
	} else {
		v.Loyalty -= 20
	}
	if v.Loyalty < 0 { v.Loyalty = 0 }
	_ = vRepo.Update(v)

	log.Printf("[Relief] 庄頭 %s 救災結算：分數 %d，評級 %v", v.Name, score, grade)

	// 3. 回傳結算結果
	resp := &pb.Envelope{
		Payload: &pb.Envelope_Disaster{
			Disaster: &pb.DisasterAction{
				Action: &pb.DisasterAction_ReliefResult{
					ReliefResult: &pb.ReliefResult{
						Success: true,
						Score:   int32(score),
						Reward:  reward,
						Grade:   grade,
					},
				},
			},
		},
	}

	// 廣播給同庄頭的玩家
	broadcastReliefResult(s.VillageID, resp)

	return nil, nil // 因為已經手動廣播了，不需額外回傳給發送者
}

func broadcastReliefResult(villageID int64, env *pb.Envelope) {
	sm := session.GetManager()
	for _, s := range sm.GetAllSessions() {
		if s.VillageID == villageID {
			s.QueueMessage(env)
			if s.TriggerFlush != nil { go s.TriggerFlush() }
		}
	}
}
