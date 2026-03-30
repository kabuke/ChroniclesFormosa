package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/model"
	"github.com/kabuke/ChroniclesFormosa/server/repo"
	"github.com/kabuke/ChroniclesFormosa/server/session"
)

// hashPassword 使用 SHA-256 對密碼進行不可逆雜湊，回傳 Hex 字串
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// sendLoginResponse 輔助函數：發送回應並立即 Flush
func sendLoginResponse(s *session.UserSession, success bool, message string) {
	resp := &pb.Envelope{
		Payload: &pb.Envelope_LoginResponse{
			LoginResponse: &pb.LoginResponse{
				Success: success,
				Message: message,
			},
		},
	}
	s.QueueMessage(resp)
	if s.TriggerFlush != nil {
		s.TriggerFlush()
	}
}

// validateLength 負責長度檢查 (需求：長度必須在 8 到 32 之間)
func validateLength(str string) bool {
	length := len(str)
	return length >= 8 && length <= 32
}

// HandleRegister 處理帳號註冊邏輯
func HandleRegister(req *pb.Register, s *session.UserSession) {
	// 1. 防呆驗證
	if !validateLength(req.Username) {
		sendLoginResponse(s, false, "註冊失敗：帳號長度必須為 8 到 32 字元")
		return
	}
	if !validateLength(req.Password) {
		sendLoginResponse(s, false, "註冊失敗：密碼長度必須為 8 到 32 字元")
		return
	}
	if req.Password != req.ConfirmPassword {
		sendLoginResponse(s, false, "註冊失敗：密碼與確認密碼不相符")
		return
	}

	// 2. 雜湊密碼
	pHash := hashPassword(req.Password)

	// 3. 建立 Model
	player := &model.Player{
		Username:     req.Username,
		PasswordHash: pHash,
		FactionID:    req.FactionId,
	}

	// 4. 對接 DB
	playerRepo := repo.NewPlayerRepo()
	err := playerRepo.Create(player)
	if err != nil {
		sendLoginResponse(s, false, "註冊失敗：帳號可能已經存在")
		return
	}

	sendLoginResponse(s, true, "註冊成功！請重新登入。")
}

// HandleLogin 處理使用者登入邏輯
func HandleLogin(req *pb.Login, s *session.UserSession) {
	if !validateLength(req.Username) || !validateLength(req.Password) {
		sendLoginResponse(s, false, "登入失敗：帳號或密碼格式錯誤")
		return
	}

	playerRepo := repo.NewPlayerRepo()
	player, err := playerRepo.FindByUsername(req.Username)
	if err != nil {
		if errors.Is(err, repo.ErrPlayerNotFound) {
			// 為防止列舉攻擊，統一提示帳號或密碼錯誤
			sendLoginResponse(s, false, "登入失敗：帳號不存在或密碼錯誤")
		} else {
			sendLoginResponse(s, false, "登入失敗：系統內部錯誤")
		}
		return
	}

	// 驗證密碼 Hash
	pHash := hashPassword(req.Password)
	if player.PasswordHash != pHash {
		sendLoginResponse(s, false, "登入失敗：帳號不存在或密碼錯誤")
		return
	}

	// 登入成功，將 UserSession 綁定為此玩家 Username
	s.SetUsername(player.Username)

	// 更新最後登入時間
	player.LastLoginAt = time.Now()
	_ = playerRepo.Update(player)

	sendLoginResponse(s, true, "登入成功！歡迎回到《台灣三國誌》")
}
