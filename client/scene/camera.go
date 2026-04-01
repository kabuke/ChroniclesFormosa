package scene

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kabuke/ChroniclesFormosa/client/ui"
)

type Camera struct {
	X, Y float64 
	Zoom float64
	
	isDragging bool
	lastMouseX, lastMouseY int
}

func NewCamera() *Camera {
	return &Camera{X: 1600, Y: 1600, Zoom: 1.0} 
}

func (c *Camera) Update() {
	mx, my := ebiten.CursorPosition()

	// 右鍵拖拽
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		if !c.isDragging {
			c.isDragging = true
			c.lastMouseX, c.lastMouseY = mx, my
		} else {
			dx := float64(mx - c.lastMouseX)
			dy := float64(my - c.lastMouseY)
			c.X -= dx / c.Zoom
			c.Y -= dy / c.Zoom
			c.lastMouseX, c.lastMouseY = mx, my
		}
	} else {
		c.isDragging = false
	}

	// 滾輪縮放 (當沒有開啟彈窗時才縮放地圖，解決滾動衝突)
	if !ui.GlobalVillagePanel.Visible && !ui.GlobalDiplomacyPanel.Visible && !ui.GlobalIntelPanel.Visible {
		_, wy := ebiten.Wheel()
		if wy > 0 { c.Zoom *= 1.1 }
		if wy < 0 { c.Zoom /= 1.1 }
		if c.Zoom < 0.1 { c.Zoom = 0.1 }
		if c.Zoom > 3.0 { c.Zoom = 3.0 }
	}

	// 修改為上下左右鍵移動
	speed := 10.0 / c.Zoom
	if ebiten.IsKeyPressed(ebiten.KeyUp)    { c.Y -= speed }
	if ebiten.IsKeyPressed(ebiten.KeyDown)  { c.Y += speed }
	if ebiten.IsKeyPressed(ebiten.KeyLeft)  { c.X -= speed }
	if ebiten.IsKeyPressed(ebiten.KeyRight) { c.X += speed }
}

func (c *Camera) WorldToScreen(wx, wy float64) (float64, float64) {
	sw, sh := ebiten.WindowSize()
	sx := (wx-c.X)*c.Zoom + float64(sw)/2
	sy := (wy-c.Y)*c.Zoom + float64(sh)/2
	return sx, sy
}
