package scene

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// MapScene 為戰鬥大地圖與庄頭的管理畫面
type MapScene struct {
	manager *SceneManager
}

// NewMapScene 建構子
func NewMapScene(m *SceneManager) *MapScene {
	return &MapScene{
		manager: m,
	}
}

func (s *MapScene) Update() error {
	// 【測試用】按 ESC 退回登入畫面
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		log.Println("[Map] ESC Pressed -> Returning to Login!")
		s.manager.SwitchTo("Login")
	}
	return nil
}

func (s *MapScene) Draw(screen *ebiten.Image) {
	// 先單純印出白字做雛形測試
	ebitenutil.DebugPrint(screen, "[ Map Scene ]\n\nPlaying in vast seamless beautiful Formosa world...\nPress ESC to log out.")
}

func (s *MapScene) OnEnter() {
	log.Println("[SceneManager] Entered MapScene")
}

func (s *MapScene) OnLeave() {
	log.Println("[SceneManager] Left MapScene")
}
