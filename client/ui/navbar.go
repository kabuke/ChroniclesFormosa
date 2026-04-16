package ui

import (
	"fmt"
	"image/color"

	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/kabuke/ChroniclesFormosa/client/asset"
)

// Navbar 管理頂部狀態欄
type Navbar struct {
	SceneName   string
	RTT         int64
	Connected   bool
	Stamina     int32
	MaxStamina  int32
	FactionBuff float64
	Username    string // 當前登入的使用者帳號

	TimeSyncStr string
	WarningMsg  string
	WarningTime time.Time
}

var GlobalNavbar = &Navbar{
	SceneName:   "Initializing...",
	Connected:   false,
	MaxStamina:  100,
	FactionBuff: 1.0,
}

// Draw 繪製導航條
func (n *Navbar) Draw(screen *ebiten.Image) {
	w, _ := screen.Size()
	barH := 32.0

	// 1. 繪製背景底板 (墨色半透明)
	vector.DrawFilledRect(screen, 0, 0, float32(w), float32(barH), color.RGBA{0, 0, 0, 200}, true)
	
	// 2. 繪製底部裝飾線
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
	text.Draw(screen, infoText, asset.DefaultFont, 10, 22, color.White)

	// 3.1 繪製遊戲內時間
	if n.TimeSyncStr != "" {
		text.Draw(screen, n.TimeSyncStr, asset.DefaultFont, int(w)-550, 22, color.RGBA{200, 220, 255, 255})
	}

	// 3.5 繪製天災預警 (置中閃爍)
	if n.WarningMsg != "" {
		if time.Since(n.WarningTime) > 60*time.Second {
			n.WarningMsg = "" // 60秒後消失 (或者可以保留到天災真正發生)
		} else {
			// Blink every 500ms
			if (time.Now().UnixMilli()/500)%2 == 0 {
				// Use bright yellow/red to contrast with black
				warnColor := color.RGBA{255, 200, 50, 255}
				// Draw slightly below navbar or inside
				text.Draw(screen, "⚠️ "+n.WarningMsg, asset.DefaultFont, 10, 55, warnColor)
			}
		}
	}

	// 4. 繪製精力值與 Buff (右側)
	if n.Connected {
		buffText := ""
		if n.FactionBuff > 1.0 {
			buffText = fmt.Sprintf(" [天命 x%.1f]", n.FactionBuff)
		}
		staminaText := fmt.Sprintf("⚡ %d/%d%s", n.Stamina, n.MaxStamina, buffText)
		text.Draw(screen, staminaText, asset.DefaultFont, w-220, 22, ColorFactionQing)
	}
}

func (n *Navbar) SetStatus(scene string, connected bool, rtt int64) {
	n.SceneName = scene
	n.Connected = connected
	n.RTT = rtt
}

func (n *Navbar) UpdateStamina(cur, max int32) {
	n.Stamina = cur
	n.MaxStamina = max
}

func (n *Navbar) UpdateBuff(val float64) {
	n.FactionBuff = val
}

func (n *Navbar) SetWarning(msg string) {
	n.WarningMsg = msg
	n.WarningTime = time.Now()
}
