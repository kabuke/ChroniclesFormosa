package handler

import (
	"log"

	pb "github.com/kabuke/ChroniclesFormosa/resource"
	village_logic "github.com/kabuke/ChroniclesFormosa/server/logic/village"
	"github.com/kabuke/ChroniclesFormosa/server/session"
)

// HandleVillageAction 處理所有莊頭相關請求 (VillageAction oneof)
func HandleVillageAction(action *pb.VillageAction, s *session.UserSession) {
	switch req := action.Action.(type) {
	case *pb.VillageAction_InfoReq:
		handleVillageInfoReq(req.InfoReq, s)
	case *pb.VillageAction_JoinReq:
		handleVillageJoinReq(req.JoinReq, s)
	case *pb.VillageAction_ElectReq:
		handleVillageElectReq(req.ElectReq, s)
	default:
		log.Println("[VillageHandler] Unhandled VillageAction:", req)
	}
}

// handleVillageInfoReq: 查詢村莊詳細資訊 (包含動態人口統計)
func handleVillageInfoReq(req *pb.VillageInfoReq, s *session.UserSession) {
	village, population, err := village_logic.GetVillageInfo(req.VillageId)
	if err != nil {
		log.Printf("[VillageHandler] InfoReq failed: %v", err)
		return
	}

	resp := &pb.Envelope{
		Payload: &pb.Envelope_Village{
			Village: &pb.VillageAction{
				Action: &pb.VillageAction_InfoResp{
					InfoResp: &pb.VillageInfoResp{
						VillageId:     village.ID,
						Name:          village.Name,
						Level:         village.Level,
						Population:    population, // int32
						MaxPopulation: 100,        // Phase 1 預設最多 100 人
						Headman:       "",         // Phase 1 尚未實作頭目選舉
					},
				},
			},
		},
	}
	s.QueueMessage(resp)
	if s.TriggerFlush != nil {
		s.TriggerFlush()
	}
}

// handleVillageJoinReq: 驗證玩家身分並寫入資料庫歸屬派系
func handleVillageJoinReq(req *pb.VillageJoinReq, s *session.UserSession) {
	sendResp := func(success bool, msg string) {
		resp := &pb.Envelope{
			Payload: &pb.Envelope_Village{
				Village: &pb.VillageAction{
					Action: &pb.VillageAction_JoinResp{
						JoinResp: &pb.VillageJoinResp{
							Success: success,
							Message: msg,
						},
					},
				},
			},
		}
		s.QueueMessage(resp)
		if s.TriggerFlush != nil {
			s.TriggerFlush()
		}
	}

	// 委派至 Logic 層：處理 DB 驗證與更新
	err := village_logic.JoinVillage(s.Username, req.VillageId)
	if err != nil {
		sendResp(false, err.Error())
		return
	}

	sendResp(true, "成功加入莊頭！")
}

// handleVillageElectReq 預防性拋出未實作之例外 (Phase 2 的範圍)
func handleVillageElectReq(req *pb.VillageElectReq, s *session.UserSession) {
	resp := &pb.Envelope{
		Payload: &pb.Envelope_Village{
			Village: &pb.VillageAction{
				Action: &pb.VillageAction_ElectResp{
					ElectResp: &pb.VillageElectResp{
						Success: false,
						Message: "選舉與報名參選功能尚未開放，敬請期待 Phase 2",
					},
				},
			},
		},
	}
	s.QueueMessage(resp)
	if s.TriggerFlush != nil {
		s.TriggerFlush()
	}
}
