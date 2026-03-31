package scene

import (
	"log"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/kabuke/ChroniclesFormosa/client/i18n"
	"github.com/kabuke/ChroniclesFormosa/client/network"
	"github.com/kabuke/ChroniclesFormosa/client/ui"
	pb "github.com/kabuke/ChroniclesFormosa/resource"
)

// LoginScene 為進入點的第一個畫面
type LoginScene struct {
	manager        *SceneManager
	net            *network.NetworkClient
	loginForm      *ui.GenericForm
	registerForm   *ui.GenericForm
	isRegisterMode bool
}

// NewLoginScene 建構子
func NewLoginScene(m *SceneManager, net *network.NetworkClient) *LoginScene {
	s := &LoginScene{
		manager:        m,
		net:            net,
		isRegisterMode: false,
	}

	// 初始化登入表單
	s.loginForm = &ui.GenericForm{
		Title: i18n.Global.GetText("LOGIN_TITLE"),
		Fields: []*ui.FormField{
			{Label: i18n.Global.GetText("ACCOUNT"), Value: ""},
			{Label: i18n.Global.GetText("PASSWORD"), Value: "", IsPassword: true},
		},
		OnSubmit: s.onLoginSubmit,
	}

	// 初始化註冊表單
	s.registerForm = &ui.GenericForm{
		Title: i18n.Global.GetText("REGISTER_TITLE"),
		Fields: []*ui.FormField{
			{Label: i18n.Global.GetText("ACCOUNT"), Value: ""},
			{Label: i18n.Global.GetText("PASSWORD"), Value: "", IsPassword: true},
			{Label: i18n.Global.GetText("CONFIRM_PASSWORD"), Value: "", IsPassword: true},
			{Label: i18n.Global.GetText("NICKNAME"), Value: ""},
			{Label: i18n.Global.GetText("FACTION"), Value: "1"}, // 預設 1=清軍
		},
		OnSubmit: s.onRegisterSubmit,
	}

	return s
}

func (s *LoginScene) onLoginSubmit(data map[string]string) {
	username := data[i18n.Global.GetText("ACCOUNT")]
	password := data[i18n.Global.GetText("PASSWORD")]

	log.Printf("[Login] Attempting login for %s...", username)

	env := &pb.Envelope{
		Payload: &pb.Envelope_Login{
			Login: &pb.Login{
				Username: username,
				Password: password,
			},
		},
	}
	s.net.SendEnvelope(env)
	ui.GlobalToastManager.Info(i18n.Global.GetText("VERIFYING"))
}

func (s *LoginScene) onRegisterSubmit(data map[string]string) {
	username := data[i18n.Global.GetText("ACCOUNT")]
	password := data[i18n.Global.GetText("PASSWORD")]
	confirm := data[i18n.Global.GetText("CONFIRM_PASSWORD")]
	nickname := data[i18n.Global.GetText("NICKNAME")]
	factionStr := data[i18n.Global.GetText("FACTION")]

	factionID, _ := strconv.Atoi(factionStr)

	log.Printf("[Register] Attempting signup for %s...", username)

	env := &pb.Envelope{
		Payload: &pb.Envelope_Register{
			Register: &pb.Register{
				Username:        username,
				Password:        password,
				ConfirmPassword: confirm,
				Nickname:        nickname,
				FactionId:       int32(factionID),
			},
		},
	}
	s.net.SendEnvelope(env)
	ui.GlobalToastManager.Info("Signing up...")
}

func (s *LoginScene) SwitchToLoginMode() {
	s.isRegisterMode = false
	s.loginForm.Clear()
	s.registerForm.Clear()
}

// Update 是這個畫面活躍時專屬的 60FPS 邏輯更新層
func (s *LoginScene) Update() error {
	// 切換模式 (按 Esc 或特定按鈕切換，這裡改用右側功能鍵控制)
	if inpututil.IsKeyJustPressed(ebiten.KeyControl) || inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		s.isRegisterMode = !s.isRegisterMode
		s.loginForm.Clear()
		s.registerForm.Clear()
	}

	sw, sh := ebiten.WindowSize()
	if s.isRegisterMode {
		s.registerForm.Update(sw, sh)
	} else {
		s.loginForm.Update(sw, sh)
	}
	return nil
}

// Draw 是這個畫面活躍時專屬的繪製層
func (s *LoginScene) Draw(screen *ebiten.Image) {
	screen.Fill(ui.ColorPaperWhite)

	if s.isRegisterMode {
		s.registerForm.Draw(screen)
		ebitenutil.DebugPrintAt(screen, "Press F1 to switch to LOGIN", 10, 50)
	} else {
		s.loginForm.Draw(screen)
		ebitenutil.DebugPrintAt(screen, "Press F1 to switch to SIGNUP", 10, 50)
	}
}

func (s *LoginScene) OnEnter() {
	log.Println("[SceneManager] Entered LoginScene")
}

func (s *LoginScene) OnLeave() {
	log.Println("[SceneManager] Left LoginScene")
}
