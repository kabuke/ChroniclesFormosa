package scene

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/kabuke/ChroniclesFormosa/client/i18n"
	"github.com/kabuke/ChroniclesFormosa/client/network"
	"github.com/kabuke/ChroniclesFormosa/client/ui"
	pb "github.com/kabuke/ChroniclesFormosa/resource"
)

type LoginScene struct {
	manager *SceneManager
	net     *network.NetworkClient
	isReg   bool
}

func NewLoginScene(m *SceneManager, net *network.NetworkClient) *LoginScene {
	return &LoginScene{manager: m, net: net}
}

func (s *LoginScene) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		s.isReg = !s.isReg
		if s.isReg {
			ui.GlobalKeyboard.Show(ui.ModeRegister)
		} else {
			ui.GlobalKeyboard.Show(ui.ModeLogin)
		}
		ui.GlobalKeyboard.FocusIdx = 0
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if s.isReg { s.handleRegister() } else { s.handleLogin() }
	}
	return nil
}

func (s *LoginScene) handleLogin() {
	s.net.SendEnvelope(&pb.Envelope{
		Payload: &pb.Envelope_Login{
			Login: &pb.Login{
				Username: ui.GlobalKeyboard.User,
				Password: ui.GlobalKeyboard.Pass,
			},
		},
	})
}

func (s *LoginScene) handleRegister() {
	s.net.SendEnvelope(&pb.Envelope{
		Payload: &pb.Envelope_Register{
			Register: &pb.Register{
				Username:        ui.GlobalKeyboard.User,
				Password:        ui.GlobalKeyboard.Pass,
				ConfirmPassword: ui.GlobalKeyboard.Pass,
				Nickname:        ui.GlobalKeyboard.Nick,
			},
		},
	})
}

func (s *LoginScene) Draw(screen *ebiten.Image) {
	screen.Fill(ui.ColorPaperWhite)
	
	title := i18n.Global.GetText("LOGIN_TITLE")
	if s.isReg { title = i18n.Global.GetText("REGISTER_TITLE") }
	ui.DrawText(screen, title, 50, 50, ui.ColorInkBlack)
	
	prompt := "按 F1 切換至註冊"
	if s.isReg { prompt = "按 F1 切換至登入" }
	ui.DrawText(screen, prompt, 50, 80, ui.ColorInkPale)
}

func (s *LoginScene) OnEnter() {
	ui.GlobalKeyboard.Show(ui.ModeLogin)
	ui.GlobalNavbar.SceneName = "Login"
}

func (s *LoginScene) OnLeave() { ui.GlobalKeyboard.Hide() }

func (s *LoginScene) SwitchToLoginMode() {
	s.isReg = false
	ui.GlobalKeyboard.Show(ui.ModeLogin)
}
