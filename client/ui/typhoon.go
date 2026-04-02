package ui

import (
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type RainDrop struct {
	x, y, length, speed float64
}

type TyphoonSystem struct {
	drops     []RainDrop
	intensity float64
	active    bool
}

var GlobalTyphoon = &TyphoonSystem{}

func (t *TyphoonSystem) SetActive(active bool, intensity float32) {
	t.active = active
	t.intensity = float64(intensity)
	if active && len(t.drops) == 0 {
		for i := 0; i < 500; i++ {
			t.drops = append(t.drops, RainDrop{
				x: rand.Float64() * 2000,
				y: rand.Float64() * 1500,
				length: rand.Float64()*10 + 5,
				speed: rand.Float64()*15 + 15,
			})
		}
	} else if !active {
		t.drops = nil
	}
}

func (t *TyphoonSystem) Update() {
	if !t.active { return }
	for i := range t.drops {
		t.drops[i].x -= t.drops[i].speed * 0.5 // wind from right to left
		t.drops[i].y += t.drops[i].speed

		// Wrap around
		if t.drops[i].y > 1500 || t.drops[i].x < 0 {
			t.drops[i].y = -50
			t.drops[i].x = rand.Float64() * 2000
		}
	}
}

func (t *TyphoonSystem) Draw(screen *ebiten.Image) {
	if !t.active { return }
	w, h := screen.Size()

	// 1. Gray filter
	filterAlpha := uint8(50 * t.intensity)
	vector.DrawFilledRect(screen, 0, 0, float32(w), float32(h), color.RGBA{50, 60, 70, filterAlpha}, true)

	// 2. Rain drops
	rainColor := color.RGBA{150, 180, 200, 150}
	for _, d := range t.drops {
		if d.x > float64(w) || d.y > float64(h) { continue }
		vector.StrokeLine(screen, float32(d.x), float32(d.y), float32(d.x-d.length*0.5), float32(d.y+d.length), 1, rainColor, true)
	}
}
