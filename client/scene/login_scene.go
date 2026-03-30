package scene

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// LoginScene 為進入點的第一個畫面
type LoginScene struct {
	manager *SceneManager
}

// NewLoginScene 建構子，吃 SceneManager 作為切換中樞
func NewLoginScene(m *SceneManager) *LoginScene {
	return &LoginScene{
		manager: m,
	}
}

// Update 是這個畫面活躍時專屬的 60FPS 邏輯更新層
func (s *LoginScene) Update() error {
	// 【測試用】按空白鍵切換至地圖
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		log.Println("[Login] Space Pressed -> Switching to Map!")
		s.manager.SwitchTo("Map")
	}
	return nil
}

// Draw 是這個畫面活躍時專屬的繪製層
func (s *LoginScene) Draw(screen *ebiten.Image) {
	// 先單純印出白字做雛形測試
	ebitenutil.DebugPrint(screen, "=== Chronicles Formosa (台灣三國誌) ===\n\n[ Login Scene ]\nPress SPACE to enter Map")
}

func (s *LoginScene) OnEnter() {
	log.Println("[SceneManager] Entered LoginScene")
}

func (s *LoginScene) OnLeave() {
	log.Println("[SceneManager] Left LoginScene")
}
