package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

type EnemyType int

const (
	TypeChaser EnemyType = iota
	TypeCharger
	TypeBoss
	TypeElite
)

type Enemy struct {
	x, y     float64
	vx, vy   float64
	hp       int
	maxHP    int
	size     float64
	speed    float64
	etype    EnemyType
	state    int // 0: idle/move, 1: charging
	timer    int
	rotation float64
	active   bool
}

var enemyPool []*Enemy

func init() {
	enemyPool = make([]*Enemy, 0, 200)
	for i := 0; i < 200; i++ {
		enemyPool = append(enemyPool, &Enemy{})
	}
}
func getEnemyFromPool() *Enemy {
	if len(enemyPool) > 0 {
		e := enemyPool[len(enemyPool)-1]
		enemyPool = enemyPool[:len(enemyPool)-1]
		// Reset state
		e.state = 0
		e.timer = 0
		e.vx = 0
		e.vy = 0
		e.active = true
		return e
	}
	return &Enemy{active: true}
}
func (e *Enemy) Deactivate() {
	enemyPool = append(enemyPool, e)
}
func NewEnemy(minutes int) *Enemy {
	side := rand.Intn(4)
	var x, y float64
	switch side {
	case 0:
		x, y = rand.Float64()*screenWidth, -30
	case 1:
		x, y = rand.Float64()*screenWidth, screenHeight+30
	case 2:
		x, y = -30, rand.Float64()*screenHeight
	case 3:
		x, y = screenWidth+30, rand.Float64()*screenHeight
	}
	etype := TypeChaser
	if rand.Float64() > 0.7 {
		etype = TypeCharger
	}
	if minutes > 0 && rand.Float64() > 0.8 {
		etype = TypeElite
	}
	// Stats scaling
	baseSpeed := 2.5
	hp := 1
	size := 30.0
	if etype == TypeCharger {
		baseSpeed = 1.5
	}
	if etype == TypeElite {
		baseSpeed = 3.0 + rand.Float64()*2.0
		hp = rand.Intn(3) + 3
		size = 60 + rand.Float64()*30
	}
	e := getEnemyFromPool()
	e.x = x
	e.y = y
	e.hp = hp
	e.maxHP = hp
	e.size = size
	e.speed = baseSpeed
	e.etype = etype
	e.rotation = math.Atan2(screenHeight/2-y, screenWidth/2-x)
	return e
}
func NewBoss(minutes int) *Enemy {
	side := rand.Intn(4)
	var x, y float64
	switch side {
	case 0:
		x, y = rand.Float64()*screenWidth, -50
	case 1:
		x, y = rand.Float64()*screenWidth, screenHeight+50
	case 2:
		x, y = -50, rand.Float64()*screenHeight
	case 3:
		x, y = screenWidth+50, rand.Float64()*screenHeight
	}
	// Health: (Random + high default) * minutes passed
	health := (rand.Intn(10) + 20) * (minutes + 1)
	speed := 1.5
	e := getEnemyFromPool()
	e.x = x
	e.y = y
	e.hp = health
	e.maxHP = health
	e.size = 100 + (float64(minutes) * 10)
	e.speed = speed
	e.etype = TypeBoss
	e.rotation = math.Atan2(screenHeight/2-y, screenWidth/2-x)
	return e
}
func (e *Enemy) Update(px, py float64, isWaveMode bool) {
	effectiveSpeed := e.speed
	if isWaveMode {
		beat := GetBeatPulse()
		effectiveSpeed *= (1.0 + beat*0.65) // Pulse speed up to 65% faster on beat in Wave Mode
	}

	switch e.etype {
	case TypeChaser:
		angle := math.Atan2(py-e.y, px-e.x)
		e.vx = math.Cos(angle) * effectiveSpeed
		e.vy = math.Sin(angle) * effectiveSpeed
	case TypeCharger:
		if e.state == 0 { // Move towards player slowly
			angle := math.Atan2(py-e.y, px-e.x)
			e.vx = math.Cos(angle) * effectiveSpeed
			e.vy = math.Sin(angle) * effectiveSpeed
			e.timer++
			if e.timer > 120 {
				e.state = 1
				e.timer = 0
				// Lock in direction
				e.vx *= 8.0
				e.vy *= 8.0
			}
		} else { // Charging
			e.timer++
			if e.timer > 30 {
				e.state = 0
				e.timer = 0
			}
		}
	case TypeBoss, TypeElite:
		// Slow but unstoppable or fast and random
		angle := math.Atan2(py-e.y, px-e.x)
		e.vx = math.Cos(angle) * effectiveSpeed
		e.vy = math.Sin(angle) * effectiveSpeed
	}
	e.x += e.vx
	e.y += e.vy
	// Update rotation based on movement direction
	if e.vx != 0 || e.vy != 0 {
		e.rotation = math.Atan2(e.vy, e.vx)
	}
}
func (e *Enemy) Draw(screen *ebiten.Image, showUI bool) {
	clr := color.RGBA{255, 50, 50, 255} // Default Red
	switch e.etype {
	case TypeCharger:
		if e.state == 1 {
			clr = color.RGBA{255, 255, 50, 255} // Yellow
		} else {
			clr = color.RGBA{255, 128, 50, 255} // Orange
		}
	case TypeBoss:
		clr = color.RGBA{200, 50, 255, 255} // Purple
	case TypeElite:
		if e.hp > 5 || e.maxHP > 5 {
			clr = color.RGBA{200, 50, 255, 255} // Purple Void
		} else {
			clr = color.RGBA{50, 255, 100, 255} // Green
		}
	}
	if e.etype == TypeBoss {
		DrawNeonBoss(screen, e.x, e.y, e.size, clr, e.rotation+math.Pi/2)
	} else {
		DrawNeonAlien(screen, e.x, e.y, e.size, clr)
	}
	if showUI {
		// Draw health bar
		barW := e.size * 1.2
		barH := 4.0
		bx := e.x - barW/2
		by := e.y - e.size/2 - 10
		// Background
		sharedDrawOp.GeoM.Reset()
		sharedDrawOp.ColorScale.Reset()
		sharedDrawOp.GeoM.Scale(barW, barH)
		sharedDrawOp.GeoM.Translate(bx, by)
		sharedDrawOp.ColorScale.ScaleWithColor(color.RGBA{50, 50, 50, 200})
		screen.DrawImage(whitePixel, &sharedDrawOp)
		// Health
		healthRatio := float64(e.hp) / float64(e.maxHP)
		sharedDrawOp.GeoM.Reset()
		sharedDrawOp.ColorScale.Reset()
		sharedDrawOp.GeoM.Scale(barW*healthRatio, barH)
		sharedDrawOp.GeoM.Translate(bx, by)
		sharedDrawOp.ColorScale.ScaleWithColor(color.RGBA{255, 50, 50, 255})
		screen.DrawImage(whitePixel, &sharedDrawOp)
		// Numbers
		DrawNeonText(screen, fmt.Sprintf("%d/%d", e.hp, e.maxHP), bx+barW/2, by-10, 12, color.RGBA{255, 255, 255, 200})
	}
}
