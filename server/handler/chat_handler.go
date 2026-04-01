package handler

import (
	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/logic/social"
	"github.com/kabuke/ChroniclesFormosa/server/logic/stamina"
	"github.com/kabuke/ChroniclesFormosa/server/repo"
	"github.com/kabuke/ChroniclesFormosa/server/session"
)

// HandleChatRequest 處理來自客戶端的聊天發送請求
func HandleChatRequest(req *pb.ChatMessage, s *session.UserSession) {
	if req.Content == "" || s.Username == "" {
		return
	}

	// 1. 獲取玩家
	pRepo := repo.NewPlayerRepo()
	p, err := pRepo.FindByUsername(s.Username)
	if err != nil {
		return
	}

	// 2. 檢查精力 (發言消耗 1 點)
	if !stamina.ConsumeStamina(p, 1) {
		return
	}

	// 3. 執行
	social.HandleChatSend(s, req)

	// 4. 更新
	_ = pRepo.Update(p)
	stamina.SyncStamina(s, p)
}
