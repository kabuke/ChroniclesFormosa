package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/kabuke/ChroniclesFormosa/client/asset"
)

var (
	ColorPaperWhite = color.RGBA{245, 230, 200, 255}
	ColorInkBlack   = color.RGBA{44, 24, 16, 255}
	ColorInkPale    = color.RGBA{44, 24, 16, 180}
	ColorNightGold  = color.RGBA{180, 150, 80, 255}
	
	ColorFactionQing = color.RGBA{178, 34, 34, 255}  // 紅 (清軍)
	ColorFactionMing = color.RGBA{46, 139, 87, 255}  // 綠 (義軍)
	ColorFactionIndi = color.RGBA{218, 165, 32, 255} // 橘 (原民)
)

// DrawText 繪製帶陰影的標準 UI 文字
func DrawText(screen *ebiten.Image, str string, x, y int, clr color.Color) {
	// 陰影
	text.Draw(screen, str, asset.DefaultFont, x+1, y+1, color.RGBA{0, 0, 0, 200})
	text.Draw(screen, str, asset.DefaultFont, x, y, clr)
}
