package handler

import (
	"log"

	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/logic/disaster"
	"github.com/kabuke/ChroniclesFormosa/server/session"
)

func HandleDisasterAction(action *pb.DisasterAction, s *session.UserSession) {
	if s.Username == "" {
		return // 未登入不處理
	}

	if action == nil {
		return
	}

	var respEnv *pb.Envelope
	var err error

	switch act := action.Action.(type) {
	case *pb.DisasterAction_ReliefDonate:
		respEnv, err = disaster.HandleReliefDonate(s, act.ReliefDonate)
	case *pb.DisasterAction_ReliefRoute:
		respEnv, err = disaster.HandleReliefRouteSubmit(s, act.ReliefRoute)
	case *pb.DisasterAction_DebugTrigger:
		dt := act.DebugTrigger.DisasterType
		if dt == 0 {
			disaster.TriggerEarthquake(false)
		} else if dt == 1 {
			disaster.TriggerEarthquake(true)
		} else if dt == 2 {
			disaster.TriggerTyphoon()
		}
	default:
		log.Printf("[Handler] 未處理的 DisasterAction 類型: %T", act)
		return
	}

	if err != nil {
		log.Printf("[Handler] Disaster Error: %v", err)
		// 這裡可以考慮傳回一個錯誤的 Toast 給 Client
	}

	if respEnv != nil {
		s.QueueMessage(respEnv)
		if s.TriggerFlush != nil {
			go s.TriggerFlush()
		}
	}
}
