package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/kabuke/ChroniclesFormosa/client/asset"
	"github.com/kabuke/ChroniclesFormosa/client/i18n"
	pb "github.com/kabuke/ChroniclesFormosa/resource"
)

type IntelPanel struct {
	Visible bool
	History []*pb.ChatMessage
	
	X, Y, W, H float32
}

var GlobalIntelPanel = &IntelPanel{
	Visible: false,
	W:       400,
	H:       400,
}

func (p *IntelPanel) AddIntel(msg *pb.ChatMessage) {
	p.History = append(p.History, msg)
	if len(p.History) > 20 {
		p.History = p.History[1:]
	}
}

func (p *IntelPanel) Update() {
	if !p.Visible { return }
	sw, sh := ebiten.WindowSize()
	p.X = (float32(sw) - p.W) / 2
	p.Y = (float32(sh) - p.H) / 2

	mx, my := ebiten.CursorPosition()
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if IsCloseButtonClicked(mx, my, p.X, p.Y, p.W) {
			p.Hide()
		}
	}
}

func (p *IntelPanel) Draw(screen *ebiten.Image) {
	if !p.Visible { return }

	DrawFilledRoundedRect(screen, p.X, p.Y, p.W, p.H, 12, ColorPaperWhite)
	DrawCloseButton(screen, p.X, p.Y, p.W)
	
	text.Draw(screen, i18n.Global.GetText("INTEL_PANEL_TITLE"), asset.DefaultFont, int(p.X)+20, int(p.Y)+40, ColorInkBlack)
	vector.DrawFilledRect(screen, p.X+20, p.Y+55, p.W-40, 2, ColorInkBlack, true)

	if len(p.History) == 0 {
		text.Draw(screen, "目前尚無傳聞...", asset.DefaultFont, int(p.X)+20, int(p.Y)+100, ColorInkPale)
	}

	for i, m := range p.History {
		itemY := int(p.Y) + 90 + i*50
		timeStr := "傳聞"
		txt := fmt.Sprintf("[%s] %s", timeStr, m.Content)
		
		// 包裝文字顯示（簡單處理換行）
		displayTxt := txt
		if len(displayTxt) > 40 { displayTxt = displayTxt[:40] + "..." }
		
		text.Draw(screen, displayTxt, asset.DefaultFont, int(p.X)+25, itemY, color.RGBA{139, 0, 0, 255})
	}

	text.Draw(screen, "按 [ESC] 關閉", asset.DefaultFont, int(p.X)+20, int(p.Y+p.H)-20, ColorInkPale)
}

func (p *IntelPanel) Hide() { p.Visible = false }
func (p *IntelPanel) Show() { p.Visible = true }
