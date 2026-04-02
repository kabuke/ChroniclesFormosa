package handler

import (
	"log"

	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/session"
)

// HandleEnvelope 是伺服器端所有業務邏輯的總入口
func HandleEnvelope(env *pb.Envelope, s *session.UserSession) {
	// 基礎可靠性更新
	if env.Header != nil {
		s.UpdateMaxClientSeq(env.Header.Seq)
		s.Acknowledge(env.Header.Ack)
	}

	switch payload := env.Payload.(type) {
	case *pb.Envelope_Login:
		HandleLogin(payload.Login, s)
	case *pb.Envelope_Register:
		HandleRegister(payload.Register, s)
	case *pb.Envelope_Ping:
		handlePing(payload.Ping, s)
	case *pb.Envelope_Chat:
		HandleChatRequest(payload.Chat, s)
	case *pb.Envelope_Village:
		HandleVillageAction(payload.Village, s)
	case *pb.Envelope_Diplomacy:
		HandleDiplomacyAction(payload.Diplomacy, s)
	case *pb.Envelope_MoveReq:
		HandleAoiUpdate(payload.MoveReq, s)
	case *pb.Envelope_Disaster:
		HandleDisasterAction(payload.Disaster, s)
	default:
		log.Printf("[Handler] Received unhandled Envelope Action: %v", payload)
	}
}

func handlePing(ping *pb.Heartbeat, s *session.UserSession) {
	resp := &pb.Envelope{
		Payload: &pb.Envelope_Pong{
			Pong: &pb.Heartbeat{
				Timestamp: ping.Timestamp,
			},
		},
	}
	s.QueueMessage(resp)
	if s.TriggerFlush != nil {
		s.TriggerFlush()
	}
}
