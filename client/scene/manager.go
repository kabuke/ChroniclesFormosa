package scene

import "github.com/hajimehoshi/ebiten/v2"

// Scene 介面統一了所有遊戲畫面的生命週期
type Scene interface {
	Update() error
	Draw(screen *ebiten.Image)
	OnEnter() // 面板被切入時觸發（如：開始播放背景音樂、載入地圖）
	OnLeave() // 面板被切出時觸發（如：釋放資源）
}

// SceneManager 用來協調 Ebiten 到不同的畫面
type SceneManager struct {
	scenes  map[string]Scene
	current string
}

func NewSceneManager() *SceneManager {
	return &SceneManager{
		scenes: make(map[string]Scene),
	}
}

// Register 註冊一個場景，需要給他一個唯一識別碼 (例如: "Login", "Map")
func (sm *SceneManager) Register(name string, scene Scene) {
	sm.scenes[name] = scene
}

// SwitchTo 轉換到另外一個場景，自動呼叫 OnLeave 與 OnEnter
func (sm *SceneManager) SwitchTo(name string) {
	if sm.current != "" {
		if s, ok := sm.scenes[sm.current]; ok {
			s.OnLeave()
		}
	}
	sm.current = name
	if s, ok := sm.scenes[sm.current]; ok {
		s.OnEnter()
	}
}

// Current 獲取當前作用中的場景
func (sm *SceneManager) Current() Scene {
	return sm.scenes[sm.current]
}
