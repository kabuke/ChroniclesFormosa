package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kabuke/ChroniclesFormosa/client/asset"
	"github.com/kabuke/ChroniclesFormosa/client/config"
	"github.com/kabuke/ChroniclesFormosa/client/i18n"
	"github.com/kabuke/ChroniclesFormosa/client/network"
	"github.com/kabuke/ChroniclesFormosa/client/scene"
	"github.com/kabuke/ChroniclesFormosa/client/ui"
	pb "github.com/kabuke/ChroniclesFormosa/resource"
)

// Game 實作了 ebiten.Game 介面
type Game struct {
	networkClient *network.NetworkClient
	sceneManager  *scene.SceneManager
	lastState     network.ClientState
}

// Update 每秒會被 Ebiten 呼叫 60 次
func (g *Game) Update() error {
	if g.networkClient != nil {
		g.networkClient.ProcessIncoming()
		
		if g.networkClient.State != g.lastState {
			switch g.networkClient.State {
			case network.StateConnected:
				ui.GlobalToastManager.Success(i18n.Global.GetText("STATUS_ONLINE"))
			case network.StateDisconnected:
				ui.GlobalToastManager.Error(i18n.Global.GetText("STATUS_OFFLINE"))
			case network.StateResuming:
				ui.GlobalToastManager.Warning(i18n.Global.GetText("STATUS_CONNECTING"))
			}
			g.lastState = g.networkClient.State
		}

		ui.GlobalNavbar.SetStatus(
			g.sceneManager.CurrentName(),
			g.networkClient.State != network.StateDisconnected,
			g.networkClient.RTT,
		)
	}

	ui.GlobalToastManager.Update()
	ui.GlobalKeyboard.Update()

	// 社會系統 UI 更新 (僅在大地圖)
	if g.sceneManager.CurrentName() == "Map" {
		ui.GlobalChatPanel.Visible = true
		ui.GlobalChatPanel.Update()
		ui.GlobalTensionMeter.Update()
	} else {
		ui.GlobalChatPanel.Visible = false
	}

	if current := g.sceneManager.Current(); current != nil {
		if err := current.Update(); err != nil {
			return err
		}
	}
	return nil
}

// Draw 每秒會被 Ebiten 畫 60 次
func (g *Game) Draw(screen *ebiten.Image) {
	if current := g.sceneManager.Current(); current != nil {
		current.Draw(screen)
	}
	ui.GlobalNavbar.Draw(screen)
	ui.GlobalToastManager.Draw(screen)
	ui.GlobalChatPanel.Draw(screen)
	ui.GlobalTensionMeter.Draw(screen)
	ui.GlobalKeyboard.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return config.AppConfig.ScreenWidth, config.AppConfig.ScreenHeight
}

func main() {
	config.LoadConfig()

	// 初始化資源與多語系
	i18n.Global.Init()
	if err := i18n.Global.LoadJSON(i18n.LangZhTW, "client/i18n/zh_TW.json"); err != nil {
		log.Printf("[i18n] ⚠️ Failed to load zh_TW.json")
	}
	i18n.Global.SetLanguage(i18n.LangZhTW)

	if err := asset.LoadAssets(); err != nil {
		log.Fatalf("[Asset] Critical error loading assets: %v", err)
	}

	log.Println("[Client] Start Network...")
	netClient := network.NewNetworkClient(config.AppConfig.ServerAddress)
	netClient.Connect()

	sm := scene.NewSceneManager()
	sm.Register("Login", scene.NewLoginScene(sm, netClient))
	sm.Register("Map", scene.NewMapScene(sm, netClient))
	sm.SwitchTo("Login")

	game := &Game{
		networkClient: netClient,
		sceneManager:  sm,
	}

	// 設定聊天傳送回調
	ui.OnChatSubmit = func(ch pb.ChatChannelType, content string) {
		netClient.SendEnvelope(&pb.Envelope{
			Payload: &pb.Envelope_Chat{
				Chat: &pb.ChatMessage{
					Channel: ch,
					Content: content,
				},
			},
		})
	}

	netClient.OnEnvelopeReceived = func(env *pb.Envelope) {
		// 1. 聊天訊息優先攔截
		if chat := env.GetChat(); chat != nil {
			ui.GlobalChatPanel.AddMessage(chat)
			return
		}

		// 2. 社會系統更新
		if tension := env.GetTension(); tension != nil {
			ui.GlobalTensionMeter.Set(tension.TensionValue, tension.VisualLevel)
			return
		}

		// 3. 其他業務回應
		if resp := env.GetLoginResponse(); resp != nil {
			if resp.Success {
				ui.GlobalToastManager.Success(resp.Message)
				if sm.CurrentName() == "Login" {
					if resp.Message == "註冊成功！請重新登入。" || resp.Message == "Signup Success! Please Login." {
						if loginScene, ok := sm.Current().(*scene.LoginScene); ok {
							loginScene.SwitchToLoginMode()
						}
					} else {
						sm.SwitchTo("Map")
					}
				}
			} else {
				ui.GlobalToastManager.Error(resp.Message)
			}
		}
	}

	ebiten.SetWindowSize(config.AppConfig.ScreenWidth, config.AppConfig.ScreenHeight)
	ebiten.SetWindowTitle("Chronicles Formosa (台灣三國誌)")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
