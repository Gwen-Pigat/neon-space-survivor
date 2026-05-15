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
	particles []*Particle
	pool      []*Particle
}

func NewParticleSystem() *ParticleSystem {
	ps := &ParticleSystem{
		particles: make([]*Particle, 0, 2000),
		pool:      make([]*Particle, 0, 2000),
	}
	// Pre-fill pool
	for i := 0; i < 2000; i++ {
		ps.pool = append(ps.pool, &Particle{})
	}
	return ps
}
func (ps *ParticleSystem) Add(x, y, vx, vy float64, clr color.RGBA, decay float64) {
	var p *Particle
	if len(ps.pool) > 0 {
		p = ps.pool[len(ps.pool)-1]
		ps.pool = ps.pool[:len(ps.pool)-1]
	} else {
		p = &Particle{}
	}
	p.x = x
	p.y = y
	p.vx = vx
	p.vy = vy
	p.life = 1.0
	p.decay = decay
	p.color = clr
	p.size = rand.Float64()*2 + 1
	p.active = true
	ps.particles = append(ps.particles, p)
}
func (ps *ParticleSystem) Update() {
	for i := 0; i < len(ps.particles); i++ {
		p := ps.particles[i]
		p.x += p.vx
		p.y += p.vy
		p.life -= p.decay
		if p.life <= 0 {
			p.active = false
			ps.pool = append(ps.pool, p)
			// Swap-and-pop
			ps.particles[i] = ps.particles[len(ps.particles)-1]
			ps.particles = ps.particles[:len(ps.particles)-1]
			i--
		}
	}
}
func (ps *ParticleSystem) Draw(screen *ebiten.Image) {
	if len(ps.particles) == 0 {
		return
	}

	for _, p := range ps.particles {
		op := &ebiten.DrawImageOptions{}
		s := p.size * p.life
		op.GeoM.Scale(s, s)
		op.GeoM.Translate(p.x, p.y)

		// Apply life to alpha
		alpha := float32(p.color.A) * float32(p.life) / 255.0
		op.ColorScale.Scale(
			float32(p.color.R)/255,
			float32(p.color.G)/255,
			float32(p.color.B)/255,
			alpha,
		)

		screen.DrawImage(particleImg, op)
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
