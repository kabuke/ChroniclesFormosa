package handler

import (
	"log"

	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/repo"
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
	villageRepo := repo.NewVillageRepo()
	village, err := villageRepo.FindByID(req.VillageId)

	if err != nil {
		log.Printf("[VillageHandler] InfoReq failed: %v", err)
		return
	}

	playerRepo := repo.NewPlayerRepo()
	// 即時從 DB 統計有多少名玩家村莊歸屬於此地
	population, _ := playerRepo.CountByVillageID(req.VillageId)

	resp := &pb.Envelope{
		Payload: &pb.Envelope_Village{
			Village: &pb.VillageAction{
				Action: &pb.VillageAction_InfoResp{
					InfoResp: &pb.VillageInfoResp{
						VillageId:     village.ID,
						Name:          village.Name,
						Level:         village.Level,
						Population:    int32(population),
						MaxPopulation: 100, // Phase 1 預設最多 100 人
						Headman:       "",  // Phase 1 尚未實作頭目選舉
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

	// 1. 防呆：必須擁有合法的 Username 身分（在 auth.go 內賦予）
	if s.Username == "" {
		sendResp(false, "您必須先登入帳號才能執行加入莊頭的操作。")
		return
	}

	// 2. 確保此村莊真實存在
	vRepo := repo.NewVillageRepo()
	_, err := vRepo.FindByID(req.VillageId)
	if err != nil {
		sendResp(false, "查無此莊頭，請確認莊頭編號。")
		return
	}

	// 3. 取得目前登入者的 DB 實體
	pRepo := repo.NewPlayerRepo()
	player, err := pRepo.FindByUsername(s.Username)
	if err != nil {
		sendResp(false, "無法取得您的角色資料庫紀錄。")
		return
	}

	// 4. 重複屬地判斷與防呆
	if player.VillageID == req.VillageId {
		sendResp(true, "您早已經是本莊頭的成員囉！")
		return
	}
	if player.VillageID != 0 {
		sendResp(false, "您已經是其他莊頭的成員了，請先尋求退出原莊頭。")
		return
	}

	// 5. 寫入變更
	player.VillageID = req.VillageId
	if err := pRepo.Update(player); err != nil {
		log.Printf("[VillageHandler] DB Update Error: %v", err)
		sendResp(false, "加入莊頭失敗，系統儲存發生錯誤。")
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
