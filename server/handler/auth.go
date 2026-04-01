package handler

import (
	"log"

	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/repo"
	"github.com/kabuke/ChroniclesFormosa/server/session"
	"github.com/kabuke/ChroniclesFormosa/server/model"
	"github.com/kabuke/ChroniclesFormosa/server/logic/stamina"
	"github.com/kabuke/ChroniclesFormosa/common/crypto"
)

func HandleLogin(req *pb.Login, s *session.UserSession) {
	pRepo := repo.NewPlayerRepo()
	player, err := pRepo.FindByUsername(req.Username)
	if err != nil {
		sendLoginResp(s, false, "找不到使用者")
		return
	}

	// 修正：對輸入密碼進行雜湊後再與資料庫比對
	inputHash := crypto.HashSHA256(req.Password)
	if player.PasswordHash != inputHash {
		sendLoginResp(s, false, "密碼錯誤")
		return
	}

	s.SetUsername(player.Username)
	s.FactionID = player.FactionID
	s.VillageID = player.VillageID 

	sendLoginResp(s, true, "登入成功！")
	
	// 🇹🇼 核心修復：登入後立即同步資料庫中的真實精力值
	stamina.SyncStamina(s, player)
	
	log.Printf("[Auth] User %s logged in. Village: %d, Stamina: %d", player.Username, player.VillageID, player.Stamina)
}

func HandleRegister(req *pb.Register, s *session.UserSession) {
	if req.Password != req.ConfirmPassword {
		sendLoginResp(s, false, "兩次密碼不一致")
		return
	}

	pRepo := repo.NewPlayerRepo()
	if _, err := pRepo.FindByUsername(req.Username); err == nil {
		sendLoginResp(s, false, "使用者名稱已存在")
		return
	}

	// 註冊時也進行雜湊
	player := &model.Player{
		Username:     req.Username,
		PasswordHash: crypto.HashSHA256(req.Password),
		Nickname:     req.Nickname,
		FactionID:    0, // 初始無陣營，加入庄頭後由庄頭決定
		Stamina:      100,
	}

	if err := pRepo.Create(player); err != nil {
		sendLoginResp(s, false, "註冊失敗："+err.Error())
		return
	}

	sendLoginResp(s, true, "註冊成功！請重新登入。")
}

func sendLoginResp(s *session.UserSession, success bool, msg string) {
	env := &pb.Envelope{
		Payload: &pb.Envelope_LoginResponse{
			LoginResponse: &pb.LoginResponse{
				Success: success,
				Message: msg,
			},
		},
	}
	s.QueueMessage(env)
	if s.TriggerFlush != nil { s.TriggerFlush() }
}
