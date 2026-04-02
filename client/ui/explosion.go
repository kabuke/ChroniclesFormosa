package ui

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Particle struct {
	x, y, vx, vy float64
	life         float64
	maxLife      float64
	color        color.RGBA
	size         float32
}

type ExplosionSystem struct {
	particles []Particle
}

var GlobalExplosion = &ExplosionSystem{}

func (e *ExplosionSystem) Trigger(x, y float64, count int) {
	for i := 0; i < count; i++ {
		angle := rand.Float64() * 2 * 3.14159
		speed := rand.Float64()*3 + 1
		life := rand.Float64()*30 + 30
		c := color.RGBA{100, 100, 100, 255} // Default rock color
		if rand.Float32() < 0.3 {
			c = color.RGBA{150, 150, 150, 255} // Lighter rock
		}
		
		e.particles = append(e.particles, Particle{
			x: x, y: y,
			vx: math.Cos(angle) * speed,
			vy: math.Sin(angle) * speed,
			life: life, maxLife: life,
			color: c,
			size: rand.Float32()*3 + 1,
		})
	}
}

func (e *ExplosionSystem) Update() {
	var active []Particle
	for _, p := range e.particles {
		p.x += p.vx
		p.y += p.vy
		p.vy += 0.1 // Gravity
		p.life--
		if p.life > 0 {
			active = append(active, p)
		}
	}
	e.particles = active
}

func (e *ExplosionSystem) Draw(screen *ebiten.Image) {
	for _, p := range e.particles {
		alpha := uint8((p.life / p.maxLife) * 255)
		c := color.RGBA{p.color.R, p.color.G, p.color.B, alpha}
		vector.DrawFilledRect(screen, float32(p.x), float32(p.y), p.size, p.size, c, true)
	}
}
