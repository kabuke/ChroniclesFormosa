package ui

import (
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

type ScreenShake struct {
	intensity float64
	duration  float64
	timer     float64
}

var GlobalScreenShake = &ScreenShake{}

func (s *ScreenShake) Trigger(intensity float64, duration float64) {
	s.intensity = intensity
	s.duration = duration
	s.timer = duration
}

func (s *ScreenShake) Update() {
	if s.timer > 0 {
		s.timer -= 1.0 / 60.0 // Assuming 60 TPS
		if s.timer < 0 {
			s.timer = 0
			s.intensity = 0
		}
	}
}

func (s *ScreenShake) Apply(screen *ebiten.Image) {
	if s.timer > 0 {
		// Calculate shake offset
		progress := s.timer / s.duration
		currentIntensity := s.intensity * progress
		
		dx := (rand.Float64()*2 - 1) * currentIntensity
		dy := (rand.Float64()*2 - 1) * currentIntensity

		// Create a temporary image to hold the shaken screen
		tmp := ebiten.NewImage(screen.Size())
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(dx, dy)
		tmp.DrawImage(screen, op)
		
		// Draw back to screen
		screen.Clear()
		screen.DrawImage(tmp, nil)
	}
}
