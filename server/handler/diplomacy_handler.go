package handler

import (
	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/logic/social"
	"github.com/kabuke/ChroniclesFormosa/server/session"
)

// HandleDiplomacyAction 處理外交請求封包
func HandleDiplomacyAction(action *pb.DiplomacyAction, s *session.UserSession) {
	req := action.GetReq()
	if req == nil || s.Username == "" {
		return
	}

	msg, err := social.HandleDiplomacyRequest(s, req)
	
	success := true
	if err != nil {
		success = false
		msg = err.Error()
	}

	resp := &pb.Envelope{
		Payload: &pb.Envelope_Diplomacy{
			Diplomacy: &pb.DiplomacyAction{
				Action: &pb.DiplomacyAction_Resp{
					Resp: &pb.DiplomacyResp{
						Success:    success,
						ResultDesc: msg,
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
