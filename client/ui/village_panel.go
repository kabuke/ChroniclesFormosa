package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/kabuke/ChroniclesFormosa/client/asset"
	pb "github.com/kabuke/ChroniclesFormosa/resource"
)

type VillagePanelMode int

const (
	ModeManage VillagePanelMode = iota
	ModeList
)

type VillagePanel struct {
	Visible bool
	Mode    VillagePanelMode
	
	Village *pb.VillageInfoResp
	Members []*pb.VillageMember
	AllVillages []*pb.VillageSummary
	
	X, Y, W, H float32
	SelectedIdx int
	ScrollIdx   int
}

var GlobalVillagePanel = &VillagePanel{
	Visible: false,
	W: 420,
	H: 520,
	SelectedIdx: -1,
}

var OnStabilitySubmit func(opType pb.StabilityOpType)
var OnElectSubmit func()
var OnJoinSubmit func(villageID int64)
var OnImpeachSubmit func()

func (p *VillagePanel) Update() {
	if !p.Visible { return }

	sw, sh := ebiten.WindowSize()
	p.X = (float32(sw) - p.W) / 2
	p.Y = (float32(sh) - p.H) / 2

	_, wy := ebiten.Wheel()
	if wy > 0 && p.ScrollIdx > 0 { p.ScrollIdx-- }
	if p.Mode == ModeManage {
		if wy < 0 && len(p.Members) > 8 && p.ScrollIdx < len(p.Members)-8 { p.ScrollIdx++ }
	} else {
		if wy < 0 && len(p.AllVillages) > 8 && p.ScrollIdx < len(p.AllVillages)-8 { p.ScrollIdx++ }
	}

	mx, my := ebiten.CursorPosition()
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if IsCloseButtonClicked(mx, my, p.X, p.Y, p.W) { p.Hide(); return }

		if p.Mode == ModeManage {
			for i := 0; i < 8; i++ {
				idx := i + p.ScrollIdx
				if idx >= len(p.Members) { break }
				itemY := p.Y + 160 + float32(i*30)
				if p.isPointIn(float32(mx), float32(my), p.X+20, itemY, p.W-40, 25) { p.SelectedIdx = idx }
			}
		} else {
			for i := 0; i < 8; i++ {
				idx := i + p.ScrollIdx
				if idx >= len(p.AllVillages) { break }
				itemY := p.Y + 80 + float32(i*45)
				if p.isPointIn(float32(mx), float32(my), p.X+20, itemY, p.W-40, 40) { p.SelectedIdx = idx }
			}
		}

		btnY := p.Y + p.H - 65
		if p.Mode == ModeManage {
			// [推舉]
			if p.isPointIn(float32(mx), float32(my), p.X+20, btnY, 90, 35) {
				if GlobalNavbar.Stamina >= 5 && OnElectSubmit != nil { OnElectSubmit() }
			}
			// [辦桌]
			if p.isPointIn(float32(mx), float32(my), p.X+120, btnY, 90, 35) {
				if GlobalNavbar.Stamina >= 10 && OnStabilitySubmit != nil { OnStabilitySubmit(pb.StabilityOpType_OP_BANQUET) }
			}
			// [祭祀]
			if p.isPointIn(float32(mx), float32(my), p.X+220, btnY, 90, 35) {
				if GlobalNavbar.Stamina >= 10 && OnStabilitySubmit != nil { OnStabilitySubmit(pb.StabilityOpType_OP_RITUAL) }
			}
			// [彈劾]
			if p.isPointIn(float32(mx), float32(my), p.X+320, btnY, 80, 35) {
				if OnImpeachSubmit != nil { OnImpeachSubmit() }
			}
		} else {
			if p.SelectedIdx != -1 && p.isPointIn(float32(mx), float32(my), p.X+20, btnY, p.W-40, 35) {
				if OnJoinSubmit != nil { OnJoinSubmit(p.AllVillages[p.SelectedIdx].VillageId) }
			}
		}
	}
}

func (p *VillagePanel) isPointIn(mx, my, x, y, w, h float32) bool {
	return mx >= x && mx <= x+w && my >= y && my <= y+h
}

func (p *VillagePanel) Draw(screen *ebiten.Image) {
	if !p.Visible { return }
	DrawFilledRoundedRect(screen, p.X, p.Y, p.W, p.H, 12, ColorPaperWhite)
	DrawCloseButton(screen, p.X, p.Y, p.W)
	if p.Mode == ModeManage { p.drawManage(screen) } else { p.drawList(screen) }
}

func (p *VillagePanel) drawManage(screen *ebiten.Image) {
	title := "庄頭事務"
	if p.Village != nil { title = "【 " + p.Village.Name + " 】" }
	text.Draw(screen, title, asset.DefaultFont, int(p.X)+20, int(p.Y)+40, ColorInkBlack)

	if p.Village != nil {
		info := fmt.Sprintf("等級: %d | 人口: %d | 庄長: %s", p.Village.Level, p.Village.Population, p.Village.Headman)
		text.Draw(screen, info, asset.DefaultFont, int(p.X)+20, int(p.Y)+70, ColorInkPale)
		res := fmt.Sprintf("木: %d | 糧: %d | 鐵: %d | 武: %d", p.Village.Wood, p.Village.Food, p.Village.Iron, p.Village.Soldiers)
		text.Draw(screen, res, asset.DefaultFont, int(p.X)+20, int(p.Y)+100, color.RGBA{139, 69, 19, 255})
	}

	vector.DrawFilledRect(screen, p.X+20, p.Y+120, p.W-40, 2, ColorInkBlack, true)
	text.Draw(screen, "成員清單 (滾輪滾動):", asset.DefaultFont, int(p.X)+20, int(p.Y)+145, ColorInkBlack)

	for i := 0; i < 8; i++ {
		idx := i + p.ScrollIdx
		if idx >= len(p.Members) { break }
		m := p.Members[idx]
		itemY := int(p.Y) + 175 + i*30
		clr := ColorInkBlack
		if idx == p.SelectedIdx {
			clr = ColorFactionMing
			vector.DrawFilledRect(screen, p.X+20, float32(itemY)-18, p.W-40, 25, color.RGBA{0, 0, 0, 20}, true)
		}
		roleName := "族民"
		if m.Role == 1 { roleName = "庄長" }
		text.Draw(screen, fmt.Sprintf("%s (%s) [%s]", m.Nickname, m.Username, roleName), asset.DefaultFont, int(p.X)+30, itemY, clr)
	}

	btnY := p.Y + p.H - 65
	p.drawButton(screen, p.X+20, btnY, 90, "推舉", color.RGBA{100, 100, 100, 255}, GlobalNavbar.Stamina < 5)
	p.drawButton(screen, p.X+120, btnY, 90, "辦桌", color.RGBA{178, 34, 34, 255}, GlobalNavbar.Stamina < 10)
	p.drawButton(screen, p.X+220, btnY, 90, "祭祀", color.RGBA{46, 139, 87, 255}, GlobalNavbar.Stamina < 10)
	p.drawButton(screen, p.X+320, btnY, 80, "彈劾", color.RGBA{50, 50, 50, 255}, false)
}

func (p *VillagePanel) drawList(screen *ebiten.Image) {
	text.Draw(screen, "選擇要加入的聚落 (共 16 庄):", asset.DefaultFont, int(p.X)+20, int(p.Y)+40, ColorInkBlack)
	for i := 0; i < 8; i++ {
		idx := i + p.ScrollIdx
		if idx >= len(p.AllVillages) { break }
		v := p.AllVillages[idx]
		itemY := int(p.Y) + 80 + i*45
		bgAlpha := uint8(10)
		if idx == p.SelectedIdx { bgAlpha = 30 }
		vector.DrawFilledRect(screen, p.X+20, float32(itemY), p.W-40, 40, color.RGBA{0, 0, 0, bgAlpha}, true)
		txt := fmt.Sprintf("%s (Lv.%d) | 人口: %d", v.Name, v.Level, v.Population)
		text.Draw(screen, txt, asset.DefaultFont, int(p.X)+35, itemY+25, ColorInkBlack)
	}
	btnY := p.Y + p.H - 65
	p.drawButton(screen, p.X+20, btnY, p.W-40, "確認加入所選庄頭", ColorFactionMing, p.SelectedIdx == -1)
}

func (p *VillagePanel) drawButton(screen *ebiten.Image, x, y, w float32, label string, clr color.Color, disabled bool) {
	finalClr := clr
	if disabled { finalClr = color.RGBA{150, 150, 150, 255} }
	DrawFilledRoundedRect(screen, x, y, w, 35, 5, finalClr)
	text.Draw(screen, label, asset.DefaultFont, int(x+w/2)-20, int(y)+24, color.White)
}

func (p *VillagePanel) ShowManage(v *pb.VillageInfoResp, members []*pb.VillageMember) {
	p.Mode = ModeManage
	p.Village = v
	p.Members = members
	p.Visible = true
	p.SelectedIdx = -1
	p.ScrollIdx = 0
}

func (p *VillagePanel) ShowList(villages []*pb.VillageSummary) {
	p.Mode = ModeList
	p.AllVillages = villages
	p.Visible = true
	p.SelectedIdx = -1
	p.ScrollIdx = 0
}

func (p *VillagePanel) Hide() { p.Visible = false }
