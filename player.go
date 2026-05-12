package main

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

type Player struct {
	x, y        float64
	vx, vy      float64
	rotation    float64
	cooldown    int
	hp          float64
	maxHP       float64
	invulFrames int
	blinkFrames int
	tick        int
}

func NewPlayer() *Player {
	return &Player{
		x:     screenWidth / 2,
		y:     screenHeight / 2,
		hp:    100,
		maxHP: 100,
	}
}

func (p *Player) Update(ps *ParticleSystem, bullets *[]*Bullet, enemies []*Enemy, autoAim bool, fireRate int) {
	// Movement
	accel := 0.5
	friction := 0.95

	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		p.vy -= accel
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		p.vy += accel
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		p.vx -= accel
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		p.vx += accel
	}

	p.vx *= friction
	p.vy *= friction
	p.x += p.vx
	p.y += p.vy

	// Keep on screen
	if p.x < 0 {
		p.x = 0
		p.vx = 0
	}
	if p.x > screenWidth {
		p.x = screenWidth
		p.vx = 0
	}
	if p.y < 0 {
		p.y = 0
		p.vy = 0
	}
	if p.y > screenHeight {
		p.y = screenHeight
		p.vy = 0
	}

	// Rotation
	targetX, targetY := p.x, p.y
	found := false

	// Find nearest enemy for auto-aim
	minDist := 1000000.0
	if autoAim {
		for _, e := range enemies {
			dx := e.x - p.x
			dy := e.y - p.y
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < minDist {
				minDist = dist
				targetX, targetY = e.x, e.y
				found = true
			}
		}
	}

	var targetRot float64
	if found {
		targetRot = math.Atan2(targetY-p.y, targetX-p.x) + math.Pi/2
	} else {
		mx, my := ebiten.CursorPosition()
		targetRot = math.Atan2(float64(my)-p.y, float64(mx)-p.x) + math.Pi/2
	}
	p.rotation = lerpAngle(p.rotation, targetRot, 0.15)

	// Shooting (Auto-shot)
	if p.cooldown > 0 {
		p.cooldown--
	}
	if p.cooldown == 0 {
		*bullets = append(*bullets, NewBullet(p.x, p.y, p.rotation-math.Pi/2))
		p.cooldown = fireRate
	}

	// Health regeneration
	if p.hp < p.maxHP {
		p.hp += 0.05
		if p.hp > p.maxHP {
			p.hp = p.maxHP
		}
	}

	// Timers
	if p.invulFrames > 0 {
		p.invulFrames--
	}
	if p.blinkFrames > 0 {
		p.blinkFrames--
	}
	p.tick++

	// Thruster particles
	if math.Abs(p.vx) > 0.5 || math.Abs(p.vy) > 0.5 {
		ps.Add(p.x, p.y, -p.vx*0.5+randFloat(-1, 1), -p.vy*0.5+randFloat(-1, 1), color.RGBA{0, 255, 255, 255}, 0.05)
	}
}

func (p *Player) TakeDamage(amount float64) bool {
	if p.invulFrames > 0 {
		return false
	}

	p.hp -= amount
	p.invulFrames = 60
	p.blinkFrames = 60

	if p.hp < 0 {
		p.hp = 0
	}
	return true
}

func (p *Player) Draw(screen *ebiten.Image) {
	// Blinking effect
	if p.blinkFrames > 0 && (p.blinkFrames/5)%2 == 0 {
		return
	}

	// Color shift based on health
	ratio := p.hp / p.maxHP
	if ratio < 0 {
		ratio = 0
	}

	var clr color.RGBA
	if ratio > 0.5 {
		// Mostly Cyan, slightly shifting to White
		clr = color.RGBA{
			R: uint8(255 * (1 - ratio) * 2), // Should be 0 at 1.0, increase as it drops to 0.5
			G: 255,
			B: 255,
			A: 255,
		}
	} else {
		// Shifting from Cyan/White to Red
		clr = color.RGBA{
			R: 255,
			G: uint8(255 * ratio * 2),
			B: uint8(255 * ratio * 2),
			A: 255,
		}
	}

	// Danger pulse at low health
	if ratio < 0.25 {
		pulse := math.Sin(float64(p.tick)*0.2)*0.5 + 0.5
		clr.R = uint8(200 + 55*pulse)
		clr.G = uint8(50 * pulse)
		clr.B = uint8(50 * pulse)
	}

	DrawNeonShip(screen, p.x, p.y, 32, clr, p.rotation)
}

type Bullet struct {
	x, y     float64
	vx, vy   float64
	rotation float64
	life     int
}

func NewBullet(x, y, rot float64) *Bullet {
	speed := 15.0
	return &Bullet{
		x:        x,
		y:        y,
		vx:       math.Cos(rot) * speed,
		vy:       math.Sin(rot) * speed,
		rotation: rot,
		life:     120, // Long enough to cross screen
	}
}

func (b *Bullet) Update() bool {
	b.x += b.vx
	b.y += b.vy
	b.life--
	return b.life > 0
}

func (b *Bullet) Draw(screen *ebiten.Image) {
	DrawNeonBullet(screen, b.x, b.y, 4, color.RGBA{255, 255, 0, 255})
}

func randFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func lerpAngle(current, target, factor float64) float64 {
	diff := target - current
	for diff > math.Pi {
		diff -= 2 * math.Pi
	}
	for diff < -math.Pi {
		diff += 2 * math.Pi
	}
	return current + diff*factor
}
