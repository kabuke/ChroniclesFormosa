package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/kabuke/ChroniclesFormosa/client/asset"
)

type ActionMenu struct {
	Visible bool
	X, Y, W, H float32
}

var GlobalActionMenu = &ActionMenu{
	Visible: true,
	W: 120,
	H: 160,
}

var OnMenuAction func(action string)

func (m *ActionMenu) Update() {
	if !m.Visible { return }
	sw, sh := ebiten.WindowSize()
	m.X = float32(sw) - m.W - 10
	m.Y = float32(sh) - m.H - 10

	mx, my := ebiten.CursorPosition()
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		labels := []string{"庄頭事務", "外交合約", "情報紀錄", "系統設定"}
		for i, label := range labels {
			btnY := m.Y + 10 + float32(i*35)
			if float32(mx) >= m.X+5 && float32(mx) <= m.X+m.W-5 &&
				float32(my) >= btnY && float32(my) <= btnY+30 {
				if OnMenuAction != nil {
					OnMenuAction(label)
				}
				break
			}
		}
	}
}

func (m *ActionMenu) Draw(screen *ebiten.Image) {
	if !m.Visible { return }
	DrawFilledRoundedRect(screen, m.X, m.Y, m.W, m.H, 8, color.RGBA{0, 0, 0, 180})
	labels := []string{"庄頭事務", "外交合約", "情報紀錄", "系統設定"}
	for i, label := range labels {
		btnY := m.Y + 10 + float32(i*35)
		DrawFilledRoundedRect(screen, m.X+5, btnY, m.W-10, 30, 4, color.RGBA{60, 60, 60, 200})
		text.Draw(screen, label, asset.DefaultFont, int(m.X)+15, int(btnY)+20, color.White)
	}
}
