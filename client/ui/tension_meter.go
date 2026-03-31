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
	
	// UI 座標
	X, Y, W, H float32
}

var GlobalTensionMeter = &TensionMeter{
	Value: 0,
	Level: TensionPeace,
	W:     200,
	H:     40,
}

func (m *TensionMeter) Set(val int32, level string) {
	m.Value = val
	m.Level = TensionLevel(level)
}

func (m *TensionMeter) Update() {
	sw, _ := ebiten.WindowSize()
	m.X = float32(sw) - m.W - 20
	m.Y = 50 // 位於 Navbar 下方
}

func (m *TensionMeter) Draw(screen *ebiten.Image) {
	// 1. 繪製背景
	DrawFilledRoundedRect(screen, m.X, m.Y, m.W, m.H, 5, color.RGBA{0, 0, 0, 150})

	// 2. 繪製「引信」進度條 (Fuze)
	barX, barY := m.X+10, m.Y+22
	barW := m.W - 20
	barH := float32(8)

	// 底色 (灰色)
	vector.DrawFilledRect(screen, barX, barY, barW, barH, color.RGBA{50, 50, 50, 255}, true)

	// 引信顏色隨緊張度變化
	fuzeColor := color.RGBA{100, 200, 100, 255} // Peace: Green
	statusText := "社會安穩"
	
	switch m.Level {
	case TensionUneasy:
		fuzeColor = color.RGBA{200, 200, 100, 255} // Yellow
		statusText = "人心浮動"
	case TensionTense:
		fuzeColor = color.RGBA{200, 100, 50, 255}  // Orange
		statusText = "劍拔弩張"
	case TensionRiot:
		fuzeColor = color.RGBA{200, 0, 0, 255}    // Red
		statusText = "械鬥邊緣"
	}

	// 填充引信 (百分比)
	fillW := barW * (float32(m.Value) / 100.0)
	vector.DrawFilledRect(screen, barX, barY, fillW, barH, fuzeColor, true)
	
	// 引信火花 (Spark)
	if m.Value > 0 {
		sparkX := barX + fillW
		vector.DrawFilledCircle(screen, sparkX, barY+barH/2, 5, color.RGBA{255, 255, 0, 255}, true)
	}

	// 3. 文字標籤
	text.Draw(screen, "族群緊張儀", asset.DefaultFont, int(m.X)+10, int(m.Y)+18, color.White)
	text.Draw(screen, statusText, asset.DefaultFont, int(m.X)+120, int(m.Y)+18, fuzeColor)
}
