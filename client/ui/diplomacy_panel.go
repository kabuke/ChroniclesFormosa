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

type DiplomacyPanel struct {
	Visible     bool
	AllVillages []*pb.VillageSummary
	SelectedIdx int
	ScrollIdx   int // 滾動偏移
	
	X, Y, W, H float32
}

var GlobalDiplomacyPanel = &DiplomacyPanel{
	Visible:     false,
	W:           400,
	H:           480,
	SelectedIdx: -1,
}

var OnDiplomacySubmit func(targetID int64, dType pb.DiplomacyType)

func (p *DiplomacyPanel) Update() {
	if !p.Visible { return }

	sw, sh := ebiten.WindowSize()
	p.X = (float32(sw) - p.W) / 2
	p.Y = (float32(sh) - p.H) / 2

	// 處理滾輪滾動
	_, wy := ebiten.Wheel()
	if wy > 0 && p.ScrollIdx > 0 { p.ScrollIdx-- }
	if wy < 0 && p.ScrollIdx < len(p.AllVillages)-6 { p.ScrollIdx++ }

	mx, my := ebiten.CursorPosition()
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		// 0. 關閉判定
		if IsCloseButtonClicked(mx, my, p.X, p.Y, p.W) { p.Hide(); return }

		// 1. 選擇目標 (僅顯示 6 個)
		for i := 0; i < 6; i++ {
			idx := i + p.ScrollIdx
			if idx >= len(p.AllVillages) { break }
			
			itemY := p.Y + 80 + float32(i*45)
			if p.isPointIn(float32(mx), float32(my), p.X+20, itemY, p.W-40, 40) {
				p.SelectedIdx = idx
			}
		}

		// 2. 按鈕判定
		if p.SelectedIdx != -1 {
			targetID := p.AllVillages[p.SelectedIdx].VillageId
			btnY := p.Y + p.H - 65
			if p.isPointIn(float32(mx), float32(my), p.X+20, btnY, 175, 35) {
				if OnDiplomacySubmit != nil { OnDiplomacySubmit(targetID, pb.DiplomacyType_DIPLO_ALLIANCE) }
			}
			if p.isPointIn(float32(mx), float32(my), p.X+205, btnY, 175, 35) {
				if OnDiplomacySubmit != nil { OnDiplomacySubmit(targetID, pb.DiplomacyType_DIPLO_MARRIAGE) }
			}
		}
	}
}

func (p *DiplomacyPanel) isPointIn(mx, my, x, y, w, h float32) bool {
	return mx >= x && mx <= x+w && my >= y && my <= y+h
}

func (p *DiplomacyPanel) Draw(screen *ebiten.Image) {
	if !p.Visible { return }

	DrawFilledRoundedRect(screen, p.X, p.Y, p.W, p.H, 12, ColorPaperWhite)
	DrawCloseButton(screen, p.X, p.Y, p.W)
	
	text.Draw(screen, i18n.Global.GetText("DIPLO_TITLE"), asset.DefaultFont, int(p.X)+20, int(p.Y)+40, ColorInkBlack)
	text.Draw(screen, "(滾輪查看更多)", asset.DefaultFont, int(p.X)+200, int(p.Y)+40, ColorInkPale)

	for i := 0; i < 6; i++ {
		idx := i + p.ScrollIdx
		if idx >= len(p.AllVillages) { break }
		v := p.AllVillages[idx]

		itemY := int(p.Y) + 80 + i*45
		clr := ColorInkBlack
		bgAlpha := uint8(10)
		if idx == p.SelectedIdx {
			clr = ColorFactionMing
			bgAlpha = 30
		}
		vector.DrawFilledRect(screen, p.X+20, float32(itemY), p.W-40, 40, color.RGBA{0, 0, 0, bgAlpha}, true)
		
		faction := "無主"
		if v.FactionId == 1 { faction = "清軍" }
		if v.FactionId == 2 { faction = "義軍" }
		if v.FactionId == 3 { faction = "原民" }

		txt := fmt.Sprintf("%s (%s) - 人口: %d", v.Name, faction, v.Population)
		text.Draw(screen, txt, asset.DefaultFont, int(p.X)+35, itemY+25, clr)
	}

	btnY := p.Y + p.H - 65
	btnClr := color.RGBA{150, 150, 150, 255}
	if p.SelectedIdx != -1 { btnClr = ColorFactionMing }

	p.drawButton(screen, p.X+20, btnY, 175, i18n.Global.GetText("DIPLO_ALLIANCE"), btnClr)
	p.drawButton(screen, p.X+205, btnY, 175, i18n.Global.GetText("DIPLO_MARRIAGE"), btnClr)
}

func (p *DiplomacyPanel) drawButton(screen *ebiten.Image, x, y, w float32, label string, clr color.Color) {
	DrawFilledRoundedRect(screen, x, y, w, 35, 5, clr)
	text.Draw(screen, label, asset.DefaultFont, int(x+w/2)-35, int(y)+24, color.White)
}

func (p *DiplomacyPanel) Show(villages []*pb.VillageSummary) {
	p.AllVillages = villages
	p.Visible = true
	p.SelectedIdx = -1
	p.ScrollIdx = 0
}

func (p *DiplomacyPanel) Hide() { p.Visible = false }
