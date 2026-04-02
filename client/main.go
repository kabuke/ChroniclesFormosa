package main

import (
	"fmt"
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
	ui.OnReliefDonateSubmit = func(amount int32) {
		netClient.SendEnvelope(&pb.Envelope{
			Payload: &pb.Envelope_Disaster{
				Disaster: &pb.DisasterAction{Action: &pb.DisasterAction_ReliefDonate{ReliefDonate: &pb.ReliefDonateReq{ResourceAmount: amount}}},
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

	ui.OnReliefSubmit = func(waypoints []ui.Waypoint) {
		var route []int64
		for _, w := range waypoints {
			route = append(route, int64(w.X*1000 + w.Y)) // Simple encoding
		}
		netClient.SendEnvelope(&pb.Envelope{
			Payload: &pb.Envelope_Disaster{
				Disaster: &pb.DisasterAction{
					Action: &pb.DisasterAction_ReliefRoute{
						ReliefRoute: &pb.ReliefRouteSubmit{Waypoints: route},
					},
				},
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
		
		if timeSync := env.GetTimeSync(); timeSync != nil {
			ui.GlobalNavbar.TimeSyncStr = fmt.Sprintf("%04d年%02d月%02d日 %s", timeSync.Year, timeSync.Month, timeSync.Day, timeSync.TimeOfDay)
			return
		}

		if dAct := env.GetDisaster(); dAct != nil {
			switch act := dAct.Action.(type) {
			case *pb.DisasterAction_Warning:
				ui.GlobalNavbar.SetWarning(act.Warning.Message)
				ui.GlobalToastManager.Warning(act.Warning.Message)
			case *pb.DisasterAction_Earthquake:
				mag := act.Earthquake.Magnitude
				if mag >= 3.0 {
					ui.GlobalScreenShake.Trigger(float64(mag*3.0), 3.0)
					ui.GlobalExplosion.Trigger(float64(config.AppConfig.ScreenWidth/2), float64(config.AppConfig.ScreenHeight/2), 50)
					ui.GlobalToastManager.Error(fmt.Sprintf("發生規模 %.1f 強烈地震！震央：%s，受害庄頭數：%d", mag, act.Earthquake.EpicenterName, len(act.Earthquake.AffectedVillages)))
				} else {
					// 輕微晃動 (無害)
					ui.GlobalScreenShake.Trigger(float64(mag*1.5), 1.0)
				}
			case *pb.DisasterAction_Typhoon:
				ui.GlobalTyphoon.SetActive(true, act.Typhoon.Intensity)
				ui.GlobalToastManager.Error(fmt.Sprintf("颱風登陸！路徑：%s，受害庄頭數：%d", act.Typhoon.PathDesc, len(act.Typhoon.AffectedVillages)))
			case *pb.DisasterAction_ReliefStart:
				affected := false
				for _, vID := range act.ReliefStart.AffectedVillages {
					if ui.GlobalVillagePanel.Village != nil && ui.GlobalVillagePanel.Village.VillageId == vID {
						affected = true
						break
					}
				}
				
				ui.GlobalReliefPanel.SetAffected(affected)

				if affected {
					ui.GlobalToastManager.Info("您的庄頭受災！進入救災階段。")
					if ui.GlobalVillagePanel.Village != nil && ui.GlobalVillagePanel.Village.Headman == sm.CurrentName() {
						ui.GlobalReliefPanel.Show(act.ReliefStart.DisasterId)
					}
				}
			case *pb.DisasterAction_ReliefResult:
				ui.GlobalReliefPanel.Hide()
				ui.GlobalReliefPanel.SetAffected(false)
				ui.GlobalTyphoon.SetActive(false, 0) // Clear typhoon if active
				msg := fmt.Sprintf("救災結算 - 評分: %d, 獲得資源: %d", act.ReliefResult.Score, act.ReliefResult.Reward)
				if act.ReliefResult.Success {
					ui.GlobalToastManager.Success(msg)
				} else {
					ui.GlobalToastManager.Error(msg)
				}
			}
			return
		}

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
