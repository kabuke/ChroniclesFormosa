package scene

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Camera 負責視野的平移與縮放
type Camera struct {
	X, Y  float64
	Zoom  float64
	lastX int
	lastY int
}

func NewCamera() *Camera {
	return &Camera{
		X:    0,
		Y:    0,
		Zoom: 1.0,
	}
}

func (c *Camera) Update() {
	// 1. 處理滑鼠滾輪縮放
	_, wy := ebiten.Wheel()
	if wy > 0 {
		c.Zoom *= 1.1
	} else if wy < 0 {
		c.Zoom /= 1.1
	}

	// 限制縮放範圍
	if c.Zoom < 0.1 { c.Zoom = 0.1 }
	if c.Zoom > 5.0 { c.Zoom = 5.0 }

	// 2. 處理滑鼠右鍵拖拽平移
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		mx, my := ebiten.CursorPosition()
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
			c.lastX, c.lastY = mx, my
		} else {
			dx, dy := mx-c.lastX, my-c.lastY
			c.X -= float64(dx) / c.Zoom
			c.Y -= float64(dy) / c.Zoom
			c.lastX, c.lastY = mx, my
		}
	}
}

// WorldToScreen 將世界座標轉換為螢幕座標
func (c *Camera) WorldToScreen(wx, wy float64) (float64, float64) {
	sw, sh := ebiten.WindowSize()
	sx := (wx-c.X)*c.Zoom + float64(sw)/2
	sy := (wy-c.Y)*c.Zoom + float64(sh)/2
	return sx, sy
}

// ScreenToWorld 將螢幕座標轉換為世界座標
func (c *Camera) ScreenToWorld(sx, sy float64) (float64, float64) {
	sw, sh := ebiten.WindowSize()
	wx := (sx-float64(sw)/2)/c.Zoom + c.X
	wy := (sy-float64(sh)/2)/c.Zoom + c.Y
	return wx, wy
}
