package ui

import (
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/kabuke/ChroniclesFormosa/client/asset"
)

type KeyboardMode int
const (
	ModeChat KeyboardMode = iota
	ModeLogin
	ModeRegister
)

type GuiKeyboard struct {
	Visible bool
	Mode    KeyboardMode
	
	User string // 作為通用輸入欄位
	Pass string
	Nick string
	
	FocusIdx int 
	OnEnter  func()
	
	X, Y, W, H float32
	tick int
}

var GlobalKeyboard = &GuiKeyboard{
	Visible: false,
	W: 300,
	H: 220,
}

func (k *GuiKeyboard) Update() {
	if !k.Visible { return }
	k.tick++

	sw, sh := ebiten.WindowSize()
	k.X = (float32(sw) - k.W) / 2
	k.Y = (float32(sh) - k.H) / 2

	// 1. 切換焦點 (僅在非聊天模式有效)
	if k.Mode != ModeChat {
		if inpututil.IsKeyJustPressed(ebiten.KeyTab) || inpututil.IsKeyJustPressed(ebiten.KeyDown) {
			k.FocusIdx = (k.FocusIdx + 1) % 3
			if k.Mode == ModeLogin && k.FocusIdx > 1 { k.FocusIdx = 0 }
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
			k.FocusIdx--
			if k.FocusIdx < 0 {
				if k.Mode == ModeRegister { k.FocusIdx = 2 } else { k.FocusIdx = 1 }
			}
		}
	} else {
		k.FocusIdx = 0 // 聊天模式永遠鎖定在 User 欄位
	}

	// 2. 滑鼠點擊切換焦點
	mx, my := ebiten.CursorPosition()
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if k.Mode != ModeChat {
			if k.isPointIn(float32(mx), float32(my), k.X+20, k.Y+40, 260, 30) { k.FocusIdx = 0 }
			if k.isPointIn(float32(mx), float32(my), k.X+20, k.Y+90, 260, 30) { k.FocusIdx = 1 }
			if k.Mode == ModeRegister && k.isPointIn(float32(mx), float32(my), k.X+20, k.Y+140, 260, 30) { k.FocusIdx = 2 }
		}
	}

	// 3. 接收輸入
	chars := ebiten.AppendInputChars(nil)
	for _, c := range chars {
		switch k.FocusIdx {
		case 0: k.User += string(c)
		case 1: k.Pass += string(c)
		case 2: k.Nick += string(c)
		}
	}

	// 4. 刪除
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		switch k.FocusIdx {
		case 0: if len(k.User) > 0 { k.User = k.User[:len(k.User)-1] }
		case 1: if len(k.Pass) > 0 { k.Pass = k.Pass[:len(k.Pass)-1] }
		case 2: if len(k.Nick) > 0 { k.Nick = k.Nick[:len(k.Nick)-1] }
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if k.OnEnter != nil { k.OnEnter() }
	}
}

func (k *GuiKeyboard) isPointIn(mx, my, x, y, w, h float32) bool {
	return mx >= x && mx <= x+w && my >= y && my <= y+h
}

func (k *GuiKeyboard) Draw(screen *ebiten.Image) {
	if !k.Visible { return }

	if k.Mode == ModeChat {
		// 聊天模式：畫在底部或中心的一個小長條
		k.H = 80
		DrawFilledRoundedRect(screen, k.X, k.Y, k.W, k.H, 8, color.RGBA{245, 230, 200, 220})
		k.drawInput(screen, k.X+20, k.Y+40, "Chat Message", k.User, true)
	} else {
		// 登入註冊：完整面板
		k.H = 220
		DrawFilledRoundedRect(screen, k.X, k.Y, k.W, k.H, 8, color.RGBA{245, 230, 200, 220})
		k.drawInput(screen, k.X+20, k.Y+40, "Account", k.User, k.FocusIdx == 0)
		k.drawInput(screen, k.X+20, k.Y+90, "Password", strings.Repeat("*", len(k.Pass)), k.FocusIdx == 1)
		if k.Mode == ModeRegister {
			k.drawInput(screen, k.X+20, k.Y+140, "Nickname", k.Nick, k.FocusIdx == 2)
		}
	}
}

func (k *GuiKeyboard) drawInput(screen *ebiten.Image, x, y float32, label, val string, focus bool) {
	text.Draw(screen, label, asset.DefaultFont, int(x), int(y)-10, ColorInkBlack)
	bgClr := color.RGBA{255, 255, 255, 100}
	if focus { bgClr = color.RGBA{255, 255, 255, 200} }
	DrawFilledRoundedRect(screen, x, y, 260, 30, 4, bgClr)
	text.Draw(screen, val, asset.DefaultFont, int(x)+10, int(y)+20, ColorInkBlack)
	if focus && (k.tick/30%2 == 0) {
		text.Draw(screen, "|", asset.DefaultFont, int(x)+10+len(val)*8, int(y)+20, ColorInkBlack)
	}
}

func (k *GuiKeyboard) Show(mode KeyboardMode) { 
	k.Mode = mode
	k.Visible = true 
}
func (k *GuiKeyboard) Hide() { k.Visible = false }
