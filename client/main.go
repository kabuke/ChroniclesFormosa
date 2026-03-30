package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kabuke/ChroniclesFormosa/client/config"
	"github.com/kabuke/ChroniclesFormosa/client/network"
	"github.com/kabuke/ChroniclesFormosa/client/scene"
)

// Game 實作了 ebiten.Game 介面，是整個應用程式生命週期的最高管轄者
type Game struct {
	networkClient *network.NetworkClient
	sceneManager  *scene.SceneManager
}

// Update 每秒會被 Ebiten 呼叫 60 次 (TPS)
func (g *Game) Update() error {
	// 【安全排空】這個方法讓所有的網路回呼，都在這裡被消化完成，避免與渲染線程衝突
	if g.networkClient != nil {
		g.networkClient.ProcessIncoming()
	}

	// 委派更新邏輯給當前顯示的畫面
	if current := g.sceneManager.Current(); current != nil {
		if err := current.Update(); err != nil {
			return err
		}
	}
	return nil
}

// Draw 每秒會被 Ebiten 畫 60 次以呈現畫面
func (g *Game) Draw(screen *ebiten.Image) {
	if current := g.sceneManager.Current(); current != nil {
		current.Draw(screen)
	}
}

// Layout 這個方法決定了畫面怎麼去適應實際的視窗縮放
// 回傳的就是在 Ebiten 內的邏輯解析度。
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return config.AppConfig.ScreenWidth, config.AppConfig.ScreenHeight
}

func main() {
	// 1. 讀取基礎設定
	config.LoadConfig()

	// 2. 建立並啟動背景網路引擎
	log.Println("[Client] Start KCP Network Engine...")
	netClient := network.NewNetworkClient(config.AppConfig.ServerAddress)
	netClient.Connect() // 非阻塞方法

	// 3. 匯入場景切換器
	log.Println("[Client] Init Scene Manager...")
	sm := scene.NewSceneManager()
	sm.Register("Login", scene.NewLoginScene(sm))
	sm.Register("Map", scene.NewMapScene(sm))
	
	// 起始點先指定給 Login
	sm.SwitchTo("Login")

	// 4. 掛載 Game
	game := &Game{
		networkClient: netClient,
		sceneManager:  sm,
	}

	// 5. 設定作業系統視窗屬性
	ebiten.SetWindowSize(config.AppConfig.ScreenWidth, config.AppConfig.ScreenHeight)
	ebiten.SetWindowTitle("Chronicles Formosa (台灣三國誌)")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// 6. 燃燒你的生命吧，遊戲引擎！
	log.Println("[Client] Launching Ebiten Engine...")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
