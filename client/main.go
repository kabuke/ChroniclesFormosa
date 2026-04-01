package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/kabuke/ChroniclesFormosa/client/asset"
	"github.com/kabuke/ChroniclesFormosa/client/config"
	"github.com/kabuke/ChroniclesFormosa/client/i18n"
	"github.com/kabuke/ChroniclesFormosa/client/network"
	"github.com/kabuke/ChroniclesFormosa/client/scene"
	"github.com/kabuke/ChroniclesFormosa/client/ui"
	pb "github.com/kabuke/ChroniclesFormosa/resource"
)

type MenuAction int
const (
	MenuNone MenuAction = iota
	MenuVillage
	MenuDiplomacy
)

type Game struct {
	networkClient      *network.NetworkClient
	sceneManager       *scene.SceneManager
	lastState          network.ClientState
	pendingMenuAction  MenuAction
}

func (g *Game) Update() error {
	if g.networkClient != nil {
		g.networkClient.ProcessIncoming()
		if g.networkClient.State != g.lastState {
			if g.networkClient.State == network.StateConnected { ui.GlobalToastManager.Success(i18n.Global.GetText("STATUS_ONLINE")) }
			g.lastState = g.networkClient.State
		}
		ui.GlobalNavbar.SetStatus(g.sceneManager.CurrentName(), g.networkClient.State != network.StateDisconnected, g.networkClient.RTT)
	}

	ui.GlobalToastManager.Update()
	ui.GlobalKeyboard.Update()

	if g.sceneManager.CurrentName() == "Map" {
		ui.GlobalChatPanel.Update()
		ui.GlobalTensionMeter.Update()
		ui.GlobalVillagePanel.Update()
		ui.GlobalDiplomacyPanel.Update()
		ui.GlobalConfirmDialog.Update()
		ui.GlobalIntelPanel.Update()
		ui.GlobalActionMenu.Update()

		if inpututil.IsKeyJustPressed(ebiten.KeyV) { g.handleMenuAction("庄頭事務") }
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) { 
			ui.GlobalVillagePanel.Hide() 
			ui.GlobalDiplomacyPanel.Hide()
			ui.GlobalIntelPanel.Hide()
		}
	}

	if current := g.sceneManager.Current(); current != nil {
		if err := current.Update(); err != nil { return err }
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if current := g.sceneManager.Current(); current != nil { current.Draw(screen) }
	ui.GlobalNavbar.Draw(screen)
	ui.GlobalToastManager.Draw(screen)
	ui.GlobalChatPanel.Draw(screen)
	ui.GlobalTensionMeter.Draw(screen)
	ui.GlobalVillagePanel.Draw(screen)
	ui.GlobalDiplomacyPanel.Draw(screen)
	ui.GlobalConfirmDialog.Draw(screen)
	ui.GlobalIntelPanel.Draw(screen)
	ui.GlobalActionMenu.Draw(screen)
	ui.GlobalKeyboard.Draw(screen)
}

func (g *Game) Layout(w, h int) (int, int) { return config.AppConfig.ScreenWidth, config.AppConfig.ScreenHeight }

func (g *Game) handleMenuAction(action string) {
	switch action {
	case "庄頭事務":
		g.pendingMenuAction = MenuVillage
		g.networkClient.SendEnvelope(&pb.Envelope{
			Payload: &pb.Envelope_Village{
				Village: &pb.VillageAction{Action: &pb.VillageAction_InfoReq{InfoReq: &pb.VillageInfoReq{VillageId: 0}}},
			},
		})
	case "外交合約":
		g.pendingMenuAction = MenuDiplomacy
		g.networkClient.SendEnvelope(&pb.Envelope{
			Payload: &pb.Envelope_Village{
				Village: &pb.VillageAction{Action: &pb.VillageAction_ListReq{ListReq: &pb.VillageListReq{}}},
			},
		})
		ui.GlobalDiplomacyPanel.Visible = true
	case "情報紀錄": ui.GlobalIntelPanel.Show()
	case "系統設定": ui.GlobalToastManager.Info("功能開發中，敬請期待")
	}
}

func main() {
	config.LoadConfig()
	i18n.Global.Init()
	i18n.Global.SetLanguage(i18n.LangZhTW)
	if err := asset.LoadAssets(); err != nil { log.Fatal(err) }
	asset.InitAudio()

	netClient := network.NewNetworkClient(config.AppConfig.ServerAddress)
	netClient.Connect()

	sm := scene.NewSceneManager()
	sm.Register("Login", scene.NewLoginScene(sm, netClient))
	mapScene := scene.NewMapScene(sm, netClient)
	sm.Register("Map", mapScene)
	sm.SwitchTo("Login")

	game := &Game{networkClient: netClient, sceneManager: sm}

	ui.OnChatSubmit = func(ch pb.ChatChannelType, content string) {
		netClient.SendEnvelope(&pb.Envelope{
			Payload: &pb.Envelope_Chat{Chat: &pb.ChatMessage{Channel: ch, Content: content}},
		})
	}
	ui.OnMenuAction = game.handleMenuAction
	ui.OnJoinSubmit = func(vID int64) {
		netClient.SendEnvelope(&pb.Envelope{
			Payload: &pb.Envelope_Village{
				Village: &pb.VillageAction{Action: &pb.VillageAction_JoinReq{JoinReq: &pb.VillageJoinReq{VillageId: vID}}},
			},
		})
	}
	ui.OnStabilitySubmit = func(opType pb.StabilityOpType) {
		netClient.SendEnvelope(&pb.Envelope{
			Payload: &pb.Envelope_Village{
				Village: &pb.VillageAction{Action: &pb.VillageAction_StabilityReq{StabilityReq: &pb.VillageStabilityReq{VillageId: ui.GlobalVillagePanel.Village.VillageId, Type: opType}}},
			},
		})
	}
	ui.OnElectSubmit = func() {
		netClient.SendEnvelope(&pb.Envelope{
			Payload: &pb.Envelope_Village{
				Village: &pb.VillageAction{Action: &pb.VillageAction_ElectReq{ElectReq: &pb.VillageElectReq{VillageId: ui.GlobalVillagePanel.Village.VillageId}}},
			},
		})
	}
	ui.OnImpeachSubmit = func() {
		netClient.SendEnvelope(&pb.Envelope{
			Payload: &pb.Envelope_Village{
				Village: &pb.VillageAction{Action: &pb.VillageAction_ImpeachReq{ImpeachReq: &pb.VillageImpeachReq{VillageId: ui.GlobalVillagePanel.Village.VillageId}}},
			},
		})
	}
	ui.OnDiplomacySubmit = func(targetID int64, dType pb.DiplomacyType) {
		netClient.SendEnvelope(&pb.Envelope{
			Payload: &pb.Envelope_Diplomacy{
				Diplomacy: &pb.DiplomacyAction{Action: &pb.DiplomacyAction_Req{Req: &pb.DiplomacyReq{Type: dType, TargetVillageId: targetID}}},
			},
		})
	}

	// 🇹🇼 核心修復：進入場景時的數據請求，不應設置選單目的
	scene.OnRequestVillages = func() {
		netClient.SendEnvelope(&pb.Envelope{
			Payload: &pb.Envelope_Village{
				Village: &pb.VillageAction{Action: &pb.VillageAction_ListReq{ListReq: &pb.VillageListReq{}}},
			},
		})
	}

	netClient.OnEnvelopeReceived = func(env *pb.Envelope) {
		if chat := env.GetChat(); chat != nil { 
			ui.GlobalChatPanel.AddMessage(chat)
			if chat.Sender == "廟口說書人" { 
				ui.GlobalIntelPanel.AddIntel(chat)
				ui.GlobalToastManager.Info("收到傳聞！")
			}
			return 
		}
		if tension := env.GetTension(); tension != nil { ui.GlobalTensionMeter.Set(tension.TensionValue, tension.VisualLevel); return }
		if stamina := env.GetStamina(); stamina != nil { ui.GlobalNavbar.UpdateStamina(stamina.Current, stamina.Max); return }
		if buff := env.GetFactionBuff(); buff != nil { ui.GlobalNavbar.UpdateBuff(buff.Multiplier); return }
		
		if vAct := env.GetVillage(); vAct != nil {
			switch act := vAct.Action.(type) {
			case *pb.VillageAction_ListResp:
				// 根據狀態分發，若為 MenuNone 則僅靜默同步地圖
				if game.pendingMenuAction == MenuDiplomacy {
					ui.GlobalDiplomacyPanel.Show(act.ListResp.Villages)
				} else if game.pendingMenuAction == MenuVillage {
					ui.GlobalVillagePanel.ShowList(act.ListResp.Villages)
				}
				game.pendingMenuAction = MenuNone
				mapScene.SyncVillages(act.ListResp.Villages)
			case *pb.VillageAction_InfoResp:
				ui.GlobalVillagePanel.ShowManage(act.InfoResp, nil)
				netClient.SendEnvelope(&pb.Envelope{
					Payload: &pb.Envelope_Village{
						Village: &pb.VillageAction{Action: &pb.VillageAction_MembersReq{MembersReq: &pb.VillageMemberListReq{VillageId: act.InfoResp.VillageId}}},
					},
				})
			case *pb.VillageAction_MembersResp:
				ui.GlobalVillagePanel.Members = act.MembersResp.Members
			case *pb.VillageAction_ElectResp:
				ui.GlobalToastManager.Info(act.ElectResp.Message)
				game.handleMenuAction("庄頭事務")
			case *pb.VillageAction_JoinResp:
				if act.JoinResp.Success { 
					ui.GlobalToastManager.Success(act.JoinResp.Message)
					game.handleMenuAction("庄頭事務") 
				} else {
					ui.GlobalToastManager.Error(act.JoinResp.Message)
				}
			}
			return
		}

		if resp := env.GetLoginResponse(); resp != nil {
			if resp.Success {
				ui.GlobalToastManager.Success(resp.Message)
				if sm.CurrentName() == "Login" { sm.SwitchTo("Map") }
			} else { ui.GlobalToastManager.Error(resp.Message) }
		}
	}

	ebiten.SetWindowSize(config.AppConfig.ScreenWidth, config.AppConfig.ScreenHeight)
	ebiten.SetWindowTitle("Chronicles Formosa")
	if err := ebiten.RunGame(game); err != nil { log.Fatal(err) }
}
