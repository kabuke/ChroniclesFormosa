package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/kabuke/ChroniclesFormosa/client/asset"
)

type GuiKeyboard struct {
	Visible bool
	Input   string
	OnEnter func(string)
}

var GlobalKeyboard = &GuiKeyboard{
	Visible: false,
}

func (k *GuiKeyboard) Show() { k.Visible = true }
func (k *GuiKeyboard) Hide() { k.Visible = false }

func (k *GuiKeyboard) Update() {
	if !k.Visible {
		return
	}

	chars := ebiten.AppendInputChars(nil)
	for _, c := range chars {
		k.Input += string(c)
	}

	// 修正：連續刪除邏輯
	if isKeyRepeating(ebiten.KeyBackspace) && len(k.Input) > 0 {
		k.Input = k.Input[:len(k.Input)-1]
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if k.OnEnter != nil {
			k.OnEnter(k.Input)
		}
		k.Hide()
	}
}

func (k *GuiKeyboard) Draw(screen *ebiten.Image) {
	if !k.Visible {
		return
	}

	w, h := screen.Size()
	kbH := 200
	kbW := 400
	x := (w - kbW) / 2
	y := h - kbH - 20

	DrawFilledRoundedRect(screen, float32(x), float32(y), float32(kbW), float32(kbH), 10, color.RGBA{30, 30, 30, 240})
	
	text.Draw(screen, "VIRTUAL KEYBOARD", asset.DefaultFont, x+20, y+30, color.White)
	text.Draw(screen, "> "+k.Input+"_", asset.DefaultFont, x+20, y+80, color.RGBA{200, 200, 100, 255})
	text.Draw(screen, "Press ENTER to confirm", asset.DefaultFont, x+20, y+170, color.RGBA{150, 150, 150, 255})
}
