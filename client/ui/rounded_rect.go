package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/kabuke/ChroniclesFormosa/client/asset"
)

// DrawFilledRoundedRect 繪製填充的圓角矩形
func DrawFilledRoundedRect(screen *ebiten.Image, x, y, w, h, r float32, clr color.Color) {
	vector.DrawFilledRect(screen, x+r, y, w-2*r, h, clr, true)
	vector.DrawFilledRect(screen, x, y+r, w, h-2*r, clr, true)
	vector.DrawFilledCircle(screen, x+r, y+r, r, clr, true)
	vector.DrawFilledCircle(screen, x+w-r, y+r, r, clr, true)
	vector.DrawFilledCircle(screen, x+r, y+h-r, r, clr, true)
	vector.DrawFilledCircle(screen, x+w-r, y+h-r, r, clr, true)
}

// DrawCloseButton 繪製右上角 X 按鈕
func DrawCloseButton(screen *ebiten.Image, panelX, panelY, panelW float32) {
	size := float32(24)
	bx := panelX + panelW - size - 10
	by := panelY + 10
	
	vector.DrawFilledCircle(screen, bx+size/2, by+size/2, size/2, color.RGBA{180, 0, 0, 200}, true)
	text.Draw(screen, "X", asset.DefaultFont, int(bx)+7, int(by)+18, color.White)
}

// IsCloseButtonClicked 判定是否點擊了關閉按鈕
func IsCloseButtonClicked(mx, my int, panelX, panelY, panelW float32) bool {
	size := float32(24)
	bx := panelX + panelW - size - 10
	by := panelY + 10
	return float32(mx) >= bx && float32(mx) <= bx+size && float32(my) >= by && float32(my) <= by+size
}
