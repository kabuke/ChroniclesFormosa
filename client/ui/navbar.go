package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/kabuke/ChroniclesFormosa/client/asset"
)

// Navbar 管理頂部狀態欄
type Navbar struct {
	SceneName string
	RTT       int64
	Connected bool
}

var GlobalNavbar = &Navbar{
	SceneName: "Initializing...",
	Connected: false,
}

// Draw 繪製導航條
func (n *Navbar) Draw(screen *ebiten.Image) {
	w, _ := screen.Size()
	barH := 32.0

	// 1. 繪製背景底板 (墨色半透明)
	vector.DrawFilledRect(screen, 0, 0, float32(w), float32(barH), color.RGBA{0, 0, 0, 180}, true)
	
	// 2. 繪製底部裝飾線 (宣紙白/金)
	lineColor := ColorPaperWhite
	if n.Connected {
		lineColor = ColorNightGold
	}
	vector.StrokeLine(screen, 0, float32(barH), float32(w), float32(barH), 1, lineColor, true)

	// 3. 繪製文字資訊
	statusStr := "OFFLINE"
	if n.Connected {
		statusStr = fmt.Sprintf("ONLINE (%dms)", n.RTT)
	}

	infoText := fmt.Sprintf("【 %s 】  %s", n.SceneName, statusStr)
	
	// 使用 TrueType 字體繪製
	text.Draw(screen, infoText, asset.DefaultFont, 10, 22, color.White)
}

func (n *Navbar) SetStatus(scene string, connected bool, rtt int64) {
	n.SceneName = scene
	n.Connected = connected
	n.RTT = rtt
}
