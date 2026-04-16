package handler

import (
	"os"
	"testing"

	"github.com/kabuke/ChroniclesFormosa/server/database"
	"github.com/kabuke/ChroniclesFormosa/server/model"
	"github.com/kabuke/ChroniclesFormosa/server/session"
	pb "github.com/kabuke/ChroniclesFormosa/resource"
)

func TestMain(m *testing.M) {
	_ = database.InitDB(":memory:")
	database.GetDB().AutoMigrate(&model.Player{}, &model.Village{})
	
	code := m.Run()
	os.Exit(code)
}

func TestHandleRegisterAndLogin(t *testing.T) {
	// 1. 建立一個虛擬 Session
	sess := &session.UserSession{
		SessionID: "test-session",
		NextSeq:   1,
	}

	// 用於捕捉回傳的封包
	var lastEnv *pb.Envelope
	sess.TriggerFlush = func() {}

	// 2. 測試註冊
	regReq := &pb.Register{
		Username:        "testuser123",
		Password:        "password123",
		ConfirmPassword: "password123",
		Nickname:        "Tester",
	}
	HandleRegister(regReq, sess)

	lastEnv = sess.Outbox[len(sess.Outbox)-1]
	if lastEnv == nil || lastEnv.GetLoginResponse() == nil {
		t.Fatal("Expected LoginResponse after registration")
	}
	if !lastEnv.GetLoginResponse().Success {
		t.Errorf("Registration failed: %s", lastEnv.GetLoginResponse().Message)
	}

	// 3. 測試登入
	loginReq := &pb.Login{
		Username: "testuser123",
		Password: "password123",
	}
	HandleLogin(loginReq, sess)

	var loginEnv *pb.Envelope
	for i := len(sess.Outbox) - 1; i >= 0; i-- {
		if sess.Outbox[i].GetLoginResponse() != nil {
			loginEnv = sess.Outbox[i]
			break
		}
	}
	if loginEnv == nil {
		t.Fatalf("Login failed: loginEnv is nil, HandleLogin did not output LoginResponse")
	} else if !loginEnv.GetLoginResponse().Success {
		t.Errorf("Login failed: %s", loginEnv.GetLoginResponse().Message)
	}
	if sess.Username != "testuser123" {
		t.Errorf("Expected session username 'testuser123', got '%s'", sess.Username)
	}

	// 4. 測試登入失敗 (密碼錯誤)
	loginReqWrong := &pb.Login{
		Username: "testuser123",
		Password: "wrongpassword",
	}
	HandleLogin(loginReqWrong, sess)
	for i := len(sess.Outbox) - 1; i >= 0; i-- {
		if sess.Outbox[i].GetLoginResponse() != nil {
			loginEnv = sess.Outbox[i]
			break
		}
	}
	if loginEnv.GetLoginResponse().Success {
		t.Error("Expected login failure for wrong password")
	}
}
