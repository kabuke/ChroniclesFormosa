package handler

import (
	"log"

	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/session"
)

// HandleEnvelope 處理業務邏輯分發
func HandleEnvelope(env *pb.Envelope, s *session.UserSession) {
	// 基礎測試：如果是 Chat 封包，交給 SessionManager 轉發廣播
	switch payload := env.Payload.(type) {
	case *pb.Envelope_Chat:
		HandleChatRequest(payload.Chat, s)

	case *pb.Envelope_Login:
		HandleLogin(payload.Login, s)

	case *pb.Envelope_Register:
		HandleRegister(payload.Register, s)

	case *pb.Envelope_Village:
		HandleVillageAction(payload.Village, s)

	case *pb.Envelope_MoveReq:
		HandleAoiUpdate(payload.MoveReq, s)

	default:
		log.Println("Received unhandled Envelope Action:", env.Payload)
	}
}
