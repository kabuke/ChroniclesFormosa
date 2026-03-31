package ui

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/kabuke/ChroniclesFormosa/client/asset"
	"github.com/kabuke/ChroniclesFormosa/client/i18n"
)

type FormField struct {
	Label      string
	Value      string
	IsPassword bool
}

type GenericForm struct {
	Title     string
	Fields    []*FormField
	ActiveIdx int
	OnSubmit  func(map[string]string)
}

// isKeyRepeating 輔助函數：處理按鍵重複邏輯 (連續刪除)
func isKeyRepeating(key ebiten.Key) bool {
	d := inpututil.KeyPressDuration(key)
	if d == 1 {
		return true
	}
	if d >= 30 && d%2 == 0 {
		return true
	}
	return false
}

func (f *GenericForm) Update(sw, sh int) {
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) || inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		f.ActiveIdx = (f.ActiveIdx + 1) % len(f.Fields)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		f.ActiveIdx = (f.ActiveIdx - 1 + len(f.Fields)) % len(f.Fields)
	}

	fw, fh := float32(340), float32(280)
	fx := (float32(sw) - fw) / 2
	fy := (float32(sh) - fh) / 2

	mx, my := ebiten.CursorPosition()
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		for i := range f.Fields {
			fieldY := fy + 80 + float32(i*70)
			if float32(mx) >= fx+20 && float32(mx) <= fx+320 &&
				float32(my) >= fieldY && float32(my) <= fieldY+35 {
				f.ActiveIdx = i
			}
		}
	}

	chars := ebiten.AppendInputChars(nil)
	if len(chars) > 0 {
		f.Fields[f.ActiveIdx].Value += string(chars)
	}

	// 修正：連續刪除邏輯
	if isKeyRepeating(ebiten.KeyBackspace) && len(f.Fields[f.ActiveIdx].Value) > 0 {
		f.Fields[f.ActiveIdx].Value = f.Fields[f.ActiveIdx].Value[:len(f.Fields[f.ActiveIdx].Value)-1]
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if f.OnSubmit != nil {
			results := make(map[string]string)
			for _, field := range f.Fields {
				results[field.Label] = field.Value
			}
			f.OnSubmit(results)
		}
	}
}

func (f *GenericForm) Clear() {
	for _, field := range f.Fields {
		field.Value = ""
	}
	f.ActiveIdx = 0
}

func (f *GenericForm) Draw(screen *ebiten.Image) {
	sw, sh := screen.Size()
	fw, fh := float32(340), float32(280)
	fx := (float32(sw) - fw) / 2
	fy := (float32(sh) - fh) / 2

	DrawFilledRoundedRect(screen, fx, fy, fw, fh, 12, ColorPaperWhite)
	text.Draw(screen, f.Title, asset.DefaultFont, int(fx)+30, int(fy)+35, ColorInkBlack)

	for i, field := range f.Fields {
		y := fy + 80 + float32(i*70)
		text.Draw(screen, field.Label, asset.DefaultFont, int(fx)+25, int(y)-10, ColorInkBlack)

		boxColor := color.RGBA{255, 255, 255, 255}
		if i == f.ActiveIdx {
			boxColor = color.RGBA{255, 255, 200, 255}
		}
		DrawFilledRoundedRect(screen, fx+20, y, fw-40, 35, 5, boxColor)
		
		strokeColor := ColorInkPale
		if i == f.ActiveIdx {
			strokeColor = ColorFactionMing
		}
		vector.StrokeRect(screen, fx+20, y, fw-40, 35, 2, strokeColor, true)

		displayVal := field.Value
		if field.IsPassword {
			displayVal = ""
			for j := 0; j < len(field.Value); j++ { displayVal += "*" }
		}
		if i == f.ActiveIdx && (time.Now().UnixMilli()/500)%2 == 0 {
			displayVal += "_"
		}
		if displayVal != "" {
			text.Draw(screen, displayVal, asset.DefaultFont, int(fx)+30, int(y)+24, ColorInkBlack)
		}
	}
	
	prompt := i18n.Global.GetText("ENTER_TO_LOGIN")
	text.Draw(screen, prompt, asset.DefaultFont, int(fx)+30, int(fy+fh)-20, ColorInkPale)
}
