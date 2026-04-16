package handler

import (
	"log"

	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/logic/social"
	"github.com/kabuke/ChroniclesFormosa/server/logic/stamina"
	village_logic "github.com/kabuke/ChroniclesFormosa/server/logic/village"
	"github.com/kabuke/ChroniclesFormosa/server/repo"
	"github.com/kabuke/ChroniclesFormosa/server/session"
)

// HandleVillageAction 處理所有莊頭相關請求
func HandleVillageAction(action *pb.VillageAction, s *session.UserSession) {
	switch req := action.Action.(type) {
	case *pb.VillageAction_InfoReq:
		handleVillageInfoReq(req.InfoReq, s)
	case *pb.VillageAction_JoinReq:
		handleVillageJoinReq(req.JoinReq, s)
	case *pb.VillageAction_ElectReq:
		handleVillageElectReq(req.ElectReq, s)
	case *pb.VillageAction_ImpeachReq:
		handleVillageImpeachReq(req.ImpeachReq, s)
	case *pb.VillageAction_MembersReq:
		handleVillageMemberListReq(req.MembersReq, s)
	case *pb.VillageAction_StabilityReq:
		handleVillageStabilityReq(req.StabilityReq, s)
	case *pb.VillageAction_ListReq:
		handleVillageListReq(req.ListReq, s)
	case *pb.VillageAction_AssignRoleReq:
		handleVillageAssignRoleReq(req.AssignRoleReq, s)
	default:
		log.Println("[VillageHandler] Unhandled VillageAction:", req)
	}
}

// handleVillageListReq 獲取全服庄頭清單
func handleVillageListReq(req *pb.VillageListReq, s *session.UserSession) {
	villages, err := village_logic.GetAllVillages()
	if err != nil {
		log.Printf("[VillageHandler] ListReq failed: %v", err)
		return
	}
	resp := &pb.Envelope{
		Payload: &pb.Envelope_Village{
			Village: &pb.VillageAction{
				Action: &pb.VillageAction_ListResp{
					ListResp: &pb.VillageListResp{Villages: villages},
				},
			},
		},
	}
	s.QueueMessage(resp)
	if s.TriggerFlush != nil { s.TriggerFlush() }
}

func handleVillageStabilityReq(req *pb.VillageStabilityReq, s *session.UserSession) {
	pRepo := repo.NewPlayerRepo()
	p, _ := pRepo.FindByUsername(s.Username)
	if !stamina.ConsumeStamina(p, 10) {
		sendSystemMsg(s, "精力不足，無法執行維穩操作")
		return
	}

	msg, err := social.HandleStabilityOperation(s.Username, req.VillageId, req.Type)
	if err != nil {
		sendSystemMsg(s, "操作失敗："+err.Error())
		return
	}
	_ = pRepo.Update(p)
	stamina.SyncStamina(s, p)
	broadcastSystemMsg(msg)
}

func handleVillageMemberListReq(req *pb.VillageMemberListReq, s *session.UserSession) {
	members, err := village_logic.GetVillageMembers(req.VillageId)
	if err != nil {
		log.Printf("[VillageHandler] MembersReq failed: %v", err)
		return
	}
	resp := &pb.Envelope{
		Payload: &pb.Envelope_Village{
			Village: &pb.VillageAction{
				Action: &pb.VillageAction_MembersResp{
					MembersResp: &pb.VillageMemberListResp{VillageId: req.VillageId, Members: members},
				},
			},
		},
	}
	s.QueueMessage(resp)
	if s.TriggerFlush != nil { s.TriggerFlush() }
}

func handleVillageInfoReq(req *pb.VillageInfoReq, s *session.UserSession) {
	targetID := req.VillageId
	if targetID == 0 { targetID = s.VillageID }
	if targetID == 0 { 
		handleVillageListReq(&pb.VillageListReq{}, s)
		return
	}
	village, population, err := village_logic.GetVillageInfo(targetID)
	if err != nil {
		log.Printf("[VillageHandler] InfoReq failed: %v", err)
		return
	}
	resp := &pb.Envelope{
		Payload: &pb.Envelope_Village{
			Village: &pb.VillageAction{
				Action: &pb.VillageAction_InfoResp{
					InfoResp: &pb.VillageInfoResp{
						VillageId: village.ID, Name: village.Name, Level: village.Level,
						Population: population, MaxPopulation: 100, Headman: village.Headman,
						Wood: village.Wood, Food: village.Food, Iron: village.Iron,
						Soldiers: village.Soldiers,
					},
				},
			},
		},
	}
	s.QueueMessage(resp)
	if s.TriggerFlush != nil { s.TriggerFlush() }
}

func handleVillageJoinReq(req *pb.VillageJoinReq, s *session.UserSession) {
	err := village_logic.JoinVillage(s.Username, req.VillageId)
	if err != nil {
		sendSystemMsg(s, "加入失敗："+err.Error())
		return
	}
	s.VillageID = req.VillageId
	sendSystemMsg(s, "成功加入莊頭！從此落葉歸根。")
	handleVillageInfoReq(&pb.VillageInfoReq{VillageId: req.VillageId}, s)
}

func handleVillageElectReq(req *pb.VillageElectReq, s *session.UserSession) {
	pRepo := repo.NewPlayerRepo()
	p, _ := pRepo.FindByUsername(s.Username)
	if !stamina.ConsumeStamina(p, 5) {
		sendSystemMsg(s, "精力不足，無法發起推舉")
		return
	}
	msg, err := village_logic.ElectHeadman(s.Username, req.VillageId)
	if err != nil {
		sendSystemMsg(s, "推舉失敗："+err.Error())
		return
	}
	_ = pRepo.Update(p)
	stamina.SyncStamina(s, p)
	broadcastSystemMsg(msg)
	handleVillageInfoReq(&pb.VillageInfoReq{VillageId: req.VillageId}, s)
}

func handleVillageImpeachReq(req *pb.VillageImpeachReq, s *session.UserSession) {
	msg, err := village_logic.ImpeachHeadman(s.Username, req.VillageId)
	if err != nil {
		sendSystemMsg(s, "彈劾失敗："+err.Error())
		return
	}
	broadcastSystemMsg(msg)
	handleVillageInfoReq(&pb.VillageInfoReq{VillageId: req.VillageId}, s)
}

func handleVillageAssignRoleReq(req *pb.VillageAssignRoleReq, s *session.UserSession) {
	msg, err := village_logic.AssignRole(s.Username, req.TargetUsername, req.VillageId, req.TargetRole)
	
	resp := &pb.VillageAssignRoleResp{
		Success: err == nil,
		Message: msg,
	}
	
	if err != nil {
		resp.Message = err.Error()
	} else {
		// 若成功，向本庄廣播該事件
		broadcastSystemMsg(msg)
		// 通知客戶端刷新成員列表
		handleVillageMemberListReq(&pb.VillageMemberListReq{VillageId: req.VillageId}, s)
	}

	env := &pb.Envelope{
		Payload: &pb.Envelope_Village{
			Village: &pb.VillageAction{
				Action: &pb.VillageAction_AssignRoleResp{
					AssignRoleResp: resp,
				},
			},
		},
	}
	s.QueueMessage(env)
	if s.TriggerFlush != nil { s.TriggerFlush() }
}

func sendSystemMsg(s *session.UserSession, msg string) {
	env := &pb.Envelope{
		Payload: &pb.Envelope_Chat{
			Chat: &pb.ChatMessage{Channel: pb.ChatChannelType_CHANNEL_PRIVATE, Sender: "【系統】", Content: msg},
		},
	}
	s.QueueMessage(env)
	if s.TriggerFlush != nil { s.TriggerFlush() }
}

func broadcastSystemMsg(msg string) {
	env := &pb.Envelope{
		Payload: &pb.Envelope_Chat{
			Chat: &pb.ChatMessage{Channel: pb.ChatChannelType_CHANNEL_GLOBAL, Sender: "廟口說書人", Content: msg},
		},
	}
	session.GetManager().AddToForwardQueue(env)
}
