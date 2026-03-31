package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// DrawFilledRoundedRect 繪製一個抗鋸齒的圓角矩形
func DrawFilledRoundedRect(screen *ebiten.Image, x, y, w, h, r float32, clr color.Color) {
	vector.DrawFilledRect(screen, x+r, y, w-2*r, h, clr, true)
	vector.DrawFilledRect(screen, x, y+r, w, h-2*r, clr, true)

	// 繪製四個角落的圓弧
	vector.DrawFilledCircle(screen, x+r, y+r, r, clr, true)
	vector.DrawFilledCircle(screen, x+w-r, y+r, r, clr, true)
	vector.DrawFilledCircle(screen, x+r, y+h-r, r, clr, true)
	vector.DrawFilledCircle(screen, x+w-r, y+h-r, r, clr, true)
}
