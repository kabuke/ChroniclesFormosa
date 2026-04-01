package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/kabuke/ChroniclesFormosa/client/asset"
	"github.com/kabuke/ChroniclesFormosa/client/i18n"
)

type ConfirmDialog struct {
	Visible bool
	Title   string
	Message string
	
	X, Y, W, H float32
	OnAccept   func()
	OnReject   func()
}

var GlobalConfirmDialog = &ConfirmDialog{
	Visible: false,
	W:       320,
	H:       180,
}

func (d *ConfirmDialog) Update() {
	if !d.Visible { return }

	sw, sh := ebiten.WindowSize()
	d.X = (float32(sw) - d.W) / 2
	d.Y = (float32(sh) - d.H) / 2

	mx, my := ebiten.CursorPosition()
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if IsCloseButtonClicked(mx, my, d.X, d.Y, d.W) {
			d.Hide()
			return
		}
		btnY := d.Y + d.H - 55
		// [接受]
		if d.isPointIn(float32(mx), float32(my), d.X+20, btnY, 130, 35) {
			if d.OnAccept != nil { d.OnAccept() }
			d.Hide()
		}
		// [拒絕]
		if d.isPointIn(float32(mx), float32(my), d.X+170, btnY, 130, 35) {
			if d.OnReject != nil { d.OnReject() }
			d.Hide()
		}
	}
}

func (d *ConfirmDialog) isPointIn(mx, my, x, y, w, h float32) bool {
	return mx >= x && mx <= x+w && my >= y && my <= y+h
}

func (d *ConfirmDialog) Draw(screen *ebiten.Image) {
	if !d.Visible { return }

	DrawFilledRoundedRect(screen, d.X, d.Y, d.W, d.H, 10, ColorPaperWhite)
	
	text.Draw(screen, d.Title, asset.DefaultFont, int(d.X)+20, int(d.Y)+35, ColorInkBlack)
	text.Draw(screen, d.Message, asset.DefaultFont, int(d.X)+20, int(d.Y)+70, ColorInkPale)

	btnY := d.Y + d.H - 55
	// 接受按鈕 (綠)
	DrawFilledRoundedRect(screen, d.X+20, btnY, 130, 35, 5, color.RGBA{46, 139, 87, 255})
	text.Draw(screen, i18n.Global.GetText("ACCEPT"), asset.DefaultFont, int(d.X)+65, int(btnY)+24, color.White)

	// 拒絕按鈕 (紅)
	DrawFilledRoundedRect(screen, d.X+170, btnY, 130, 35, 5, color.RGBA{178, 34, 34, 255})
	text.Draw(screen, i18n.Global.GetText("REJECT"), asset.DefaultFont, int(d.X)+215, int(btnY)+24, color.White)
}

func (d *ConfirmDialog) Show(title, msg string, onAccept, onReject func()) {
	d.Title = title
	d.Message = msg
	d.OnAccept = onAccept
	d.OnReject = onReject
	d.Visible = true
}

func (d *ConfirmDialog) Hide() { d.Visible = false }
