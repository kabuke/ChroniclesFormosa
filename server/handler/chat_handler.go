package handler

import (
	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/logic/social"
	"github.com/kabuke/ChroniclesFormosa/server/session"
)

// HandleChatRequest 處理來自客戶端的聊天發送請求
func HandleChatRequest(req *pb.ChatMessage, s *session.UserSession) {
	// 如果訊息為空，直接無視
	if req.Content == "" {
		return
	}

	// 轉交給 Social Logic 處理路由
	social.HandleChatSend(s, req)
}
