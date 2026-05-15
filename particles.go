package main

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

type Particle struct {
	x, y   float64
	vx, vy float64
	life   float64 // 0 to 1
	decay  float64
	color  color.RGBA
	size   float64
	active bool
}
type ParticleSystem struct {
	particles []Particle
}

func NewParticleSystem() *ParticleSystem {
	return &ParticleSystem{
		particles: make([]Particle, 0, 2000),
	}
}
func (ps *ParticleSystem) Add(x, y, vx, vy float64, clr color.RGBA, decay float64) {
	if len(ps.particles) >= 2000 {
		return
	}
	ps.particles = append(ps.particles, Particle{
		x:      x,
		y:      y,
		vx:     vx,
		vy:     vy,
		life:   1.0,
		decay:  decay,
		color:  clr,
		size:   rand.Float64()*2 + 1,
		active: true,
	})
}
func (ps *ParticleSystem) Update() {
	for i := 0; i < len(ps.particles); i++ {
		p := &ps.particles[i]
		p.x += p.vx
		p.y += p.vy
		p.life -= p.decay
		if p.life <= 0 {
			// Swap-and-pop
			ps.particles[i] = ps.particles[len(ps.particles)-1]
			ps.particles = ps.particles[:len(ps.particles)-1]
			i--
		}
	}
}
var particleDrawOp ebiten.DrawImageOptions

func (ps *ParticleSystem) Draw(screen *ebiten.Image) {
	if len(ps.particles) == 0 {
		return
	}

	for i := range ps.particles {
		p := &ps.particles[i]
		particleDrawOp.GeoM.Reset()
		s := p.size * p.life
		particleDrawOp.GeoM.Scale(s, s)
		particleDrawOp.GeoM.Translate(p.x, p.y)

		alpha := float32(p.color.A) * float32(p.life) / 255.0
		particleDrawOp.ColorScale.Reset()
		particleDrawOp.ColorScale.Scale(
			float32(p.color.R)/255,
			float32(p.color.G)/255,
			float32(p.color.B)/255,
			alpha,
		)

		screen.DrawImage(particleImg, &particleDrawOp)
	}
}
func (ps *ParticleSystem) Explosion(x, y float64, clr color.RGBA, count int, sizeMultiplier float64) {
	for i := 0; i < count; i++ {
		angle := rand.Float64() * 2 * math.Pi
		speed := rand.Float64()*4*sizeMultiplier + 1
		pClr := clr
		if rand.Float64() > 0.5 {
			pClr = color.RGBA{255, 255, 255, 255} // Some white sparks
		}
		ps.Add(x, y, math.Cos(angle)*speed, math.Sin(angle)*speed, pClr, rand.Float64()*0.05+0.01)
	}
}
