package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/kabuke/ChroniclesFormosa/client/asset"
)

type TensionLevel string

const (
	TensionPeace  TensionLevel = "PEACE"
	TensionUneasy TensionLevel = "UNEASY"
	TensionTense  TensionLevel = "TENSE"
	TensionRiot   TensionLevel = "RIOT"
)

type TensionMeter struct {
	Value int32
	Level TensionLevel
	X, Y, W, H float32
}

var GlobalTensionMeter = &TensionMeter{
	Value: 0,
	Level: TensionPeace,
	W:     200,
	H:     45,
}

func (m *TensionMeter) Set(val int32, level string) {
	m.Value = val
	m.Level = TensionLevel(level)
}

func (m *TensionMeter) Update() {
	sw, _ := ebiten.WindowSize()
	m.X = float32(sw) - m.W - 20
	m.Y = 50 
}

func (m *TensionMeter) Draw(screen *ebiten.Image) {
	// 1. 背景底板 (半透明墨色)
	DrawFilledRoundedRect(screen, m.X, m.Y, m.W, m.H, 6, color.RGBA{0, 0, 0, 180})

	// 2. 引信進度條
	barX, barY := m.X+10, m.Y+28
	barW := m.W - 20
	barH := float32(6)

	// 進度條底色
	vector.DrawFilledRect(screen, barX, barY, barW, barH, color.RGBA{60, 60, 60, 255}, true)

	fuzeColor := color.RGBA{100, 255, 100, 255} 
	statusText := "社會安穩"
	
	switch m.Level {
	case TensionUneasy:
		fuzeColor = color.RGBA{255, 255, 100, 255}
		statusText = "人心浮動"
	case TensionTense:
		fuzeColor = color.RGBA{255, 150, 50, 255}
		statusText = "劍拔弩張"
	case TensionRiot:
		fuzeColor = color.RGBA{255, 50, 50, 255}
		statusText = "械鬥邊緣"
	}

	// 繪製填充部分
	fillW := barW * (float32(m.Value) / 100.0)
	if fillW > 0 {
		vector.DrawFilledRect(screen, barX, barY, fillW, barH, fuzeColor, true)
	}
	
	// 🇹🇼 核心修復：引信火花 (Spark) 圓球
	// 即使 Value=0 也繪製一個暗色圓點，讓玩家知道指標在哪
	sparkX := barX + fillW
	sparkColor := color.RGBA{255, 255, 0, 255} // 明亮黃色
	if m.Value == 0 { sparkColor = color.RGBA{100, 100, 0, 255} } // 熄滅感
	
	vector.DrawFilledCircle(screen, sparkX, barY+barH/2, 5, sparkColor, true)
	// 增加一個外發光效果
	if m.Value > 30 {
		vector.StrokeCircle(screen, sparkX, barY+barH/2, 7, 1, color.RGBA{255, 200, 0, 100}, true)
	}

	// 3. 文字標籤
	text.Draw(screen, "族群緊張儀", asset.DefaultFont, int(m.X)+10, int(m.Y)+20, color.White)
	text.Draw(screen, statusText, asset.DefaultFont, int(m.X)+120, int(m.Y)+20, fuzeColor)
}
