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
var OnAssignRoleSubmit func(targetUsername string, targetRole int32)

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
			if p.isPointIn(float32(mx), float32(my), p.X+320, btnY, 60, 35) {
				if OnImpeachSubmit != nil { OnImpeachSubmit() }
			}
			// [捐獻]
			if GlobalReliefPanel.IsAffected && p.isPointIn(float32(mx), float32(my), p.X+390, btnY, 60, 35) {
				if GlobalNavbar.Stamina >= 10 {
					GlobalReliefPanel.Show(0, p.Village.VillageId, p.Village.Name, p.Village.Population) // 點擊捐獻時開啟賑災面板
					p.Hide()
				}
			}

			// [人事任命區塊] - 若為庄長且選中某人
			if p.Village != nil && p.Village.Headman == GlobalNavbar.Username && p.SelectedIdx != -1 {
				m := p.Members[p.SelectedIdx]
				if m.Username != p.Village.Headman {
					assignY := p.Y + 160 + float32((p.SelectedIdx-p.ScrollIdx)*30) + 30
					if assignY <= p.Y+400 { // 確定還在清單顯示範圍內
						if p.isPointIn(float32(mx), float32(my), p.X+40, assignY, 80, 25) { if OnAssignRoleSubmit != nil { OnAssignRoleSubmit(m.Username, 2) } }
						if p.isPointIn(float32(mx), float32(my), p.X+130, assignY, 80, 25) { if OnAssignRoleSubmit != nil { OnAssignRoleSubmit(m.Username, 3) } }
						if p.isPointIn(float32(mx), float32(my), p.X+220, assignY, 80, 25) { if OnAssignRoleSubmit != nil { OnAssignRoleSubmit(m.Username, 4) } }
						if p.isPointIn(float32(mx), float32(my), p.X+310, assignY, 60, 25) { if OnAssignRoleSubmit != nil { OnAssignRoleSubmit(m.Username, 0) } }
					}
				}
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
		roleName := "普通族民"
		switch m.Role {
		case 1: roleName = "庄長"
		case 2: roleName = "墾首"
		case 3: roleName = "武師"
		case 4: roleName = "商賈"
		}
		if p.Village != nil && m.Username == p.Village.Headman { roleName = "庄長" }
		
		text.Draw(screen, fmt.Sprintf("%s (%s) [%s]", m.Nickname, m.Username, roleName), asset.DefaultFont, int(p.X)+30, itemY, clr)

		// 畫出任命按鈕
		if idx == p.SelectedIdx && p.Village != nil && p.Village.Headman == GlobalNavbar.Username && m.Username != p.Village.Headman {
			// 在選中的項目下方浮現指派按鈕
			assignY := float32(itemY) + 12
			p.drawButtonSmall(screen, p.X+40, assignY, 80, "指派墾首", color.RGBA{139, 69, 19, 255}, false)
			p.drawButtonSmall(screen, p.X+130, assignY, 80, "指派武師", color.RGBA{178, 34, 34, 255}, false)
			p.drawButtonSmall(screen, p.X+220, assignY, 80, "指派商賈", color.RGBA{184, 134, 11, 255}, false)
			p.drawButtonSmall(screen, p.X+310, assignY, 60, "解職", color.RGBA{50, 50, 50, 255}, false)
		}
	}

	btnY := p.Y + p.H - 65
	p.drawButton(screen, p.X+20, btnY, 90, "推舉", color.RGBA{100, 100, 100, 255}, GlobalNavbar.Stamina < 5)
	p.drawButton(screen, p.X+120, btnY, 90, "辦桌", color.RGBA{178, 34, 34, 255}, GlobalNavbar.Stamina < 10)
	p.drawButton(screen, p.X+220, btnY, 90, "祭祀", color.RGBA{46, 139, 87, 255}, GlobalNavbar.Stamina < 10)
	p.drawButton(screen, p.X+320, btnY, 60, "彈劾", color.RGBA{50, 50, 50, 255}, false)
	if GlobalReliefPanel.IsAffected {
		p.drawButton(screen, p.X+390, btnY, 60, "捐獻", color.RGBA{0, 150, 255, 255}, GlobalNavbar.Stamina < 10)
	}
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

func (p *VillagePanel) drawButtonSmall(screen *ebiten.Image, x, y, w float32, label string, clr color.Color, disabled bool) {
	finalClr := clr
	if disabled { finalClr = color.RGBA{150, 150, 150, 255} }
	DrawFilledRoundedRect(screen, x, y, w, 25, 4, finalClr)
	text.Draw(screen, label, asset.DefaultFont, int(x+w/2)-20, int(y)+17, color.White)
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
