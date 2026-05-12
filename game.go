package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Star struct {
	x, y  float64
	speed float64
	size  float64
	color color.RGBA
}

type Game struct {
	player         *Player
	stars          []Star
	particles      *ParticleSystem
	bullets        []*Bullet
	enemies        []*Enemy
	score          int
	spawnTimer     int
	gameOver       bool
	shakeFrames    int
	shakeIntensity float64
	offscreen      *ebiten.Image
	scores         *HighscoreManager
	saved          bool
	playerName     string
	timer          int // Frames
	event10        bool
	isFinalBoss    bool
	showUI         bool
	autoAim        bool
	victory        bool
	started        bool
	kills          int
	splashTimer     int
	isPaused        bool
	menuSelection   int
}

func NewGame(start bool) *Game {
	if start {
		LoadSplash()
	}
	g := &Game{
		player:    NewPlayer(),
		particles: NewParticleSystem(),
		stars:     make([]Star, 150),
		offscreen: ebiten.NewImage(screenWidth, screenHeight),
		scores:    NewHighscoreManager(),
		showUI:    true,
		autoAim:   true,
		started:   !start, // If start is false (reset), then started is true
	}
	for i := range g.stars {
		g.stars[i] = Star{
			x:     rand.Float64() * screenWidth,
			y:     rand.Float64() * screenHeight,
			speed: rand.Float64()*2 + 0.5,
			size:  rand.Float64()*1.5 + 0.5,
			color: color.RGBA{150, 150, 255, uint8(rand.Intn(155) + 100)},
		}
	}
	return g
}

func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) && g.started && !g.gameOver && !g.victory {
		g.isPaused = !g.isPaused
		g.menuSelection = 0
	}

	if g.isPaused {
		if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
			g.menuSelection--
			if g.menuSelection < 0 { g.menuSelection = 2 }
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
			g.menuSelection++
			if g.menuSelection > 2 { g.menuSelection = 0 }
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			switch g.menuSelection {
			case 0: // Resume
				g.isPaused = false
			case 1: // Reset
				g.isPaused = false
				g.Reset()
			case 2: // Quit
				return ebiten.Termination
			}
		}
		return nil
	}
	UpdateMusic(g.timer, g.started)
	if g.shakeFrames > 0 {
		g.shakeFrames--
	}

	if g.gameOver || g.victory {
		if !g.saved {
			// Handle name input
			g.playerName += string(ebiten.AppendInputChars(nil))
			if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(g.playerName) > 0 {
				g.playerName = g.playerName[:len(g.playerName)-1]
			}

			if inpututil.IsKeyJustPressed(ebiten.KeyEnter) && len(g.playerName) > 0 {
				g.scores.AddScore(g.playerName, g.score, g.timer/60, g.victory, g.kills)
				g.saved = true
			}
		} else {
			if inpututil.IsKeyJustPressed(ebiten.KeyR) {
				g.Reset()
			}
		}
		return nil
	}

	// Debug: Skip minute
	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		g.timer += 3600
	}

	// Toggle UI
	if inpututil.IsKeyJustPressed(ebiten.KeyI) {
		g.showUI = !g.showUI
	}

	// Toggle Auto-aim
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyL) {
		g.autoAim = !g.autoAim
	}

	// Handle Splash Screen
	if !g.started {
		g.splashTimer++
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.started = true
		}
		for i := range g.stars {
			g.stars[i].y += g.stars[i].speed
			if g.stars[i].y > screenHeight {
				g.stars[i].y = 0
				g.stars[i].x = rand.Float64() * screenWidth
			}
		}
		return nil
	}

	g.timer++
	minutes := g.timer / 3600
	if minutes < 10 { g.shakeIntensity = 0 }

	// Calculate fire rate based on minutes (starts at 12, decreases by 1 every minute, min 3)
	fireRate := 12 - minutes
	if minutes >= 8 {
		fireRate = 2 // Massive boost at minute 8
	} else if fireRate < 3 {
		fireRate = 3
	}
	g.player.Update(g.particles, &g.bullets, g.enemies, g.autoAim, fireRate)
	g.particles.Update()

	// Spawn boss every minute
	if g.timer > 0 && g.timer%3600 == 0 && minutes < 11 { // 60fps * 60s
		g.enemies = append(g.enemies, NewBoss(minutes))
		PlayBossAppears()
		g.shakeFrames = 20
	}

	// Minute 11 Final Boss
	if minutes == 11 && !g.isFinalBoss {
		g.isFinalBoss = true
		ResetMusic()
		finalBoss := NewBoss(20)
		finalBoss.hp = 1000
		finalBoss.maxHP = 1000 // Ensure maxHP matches
		finalBoss.size = 300
		finalBoss.speed = 4
		g.enemies = append(g.enemies, finalBoss)
		PlayBossAppears()
		g.shakeFrames = 40
	}

	// Update bullets
	for i := 0; i < len(g.bullets); i++ {
		if !g.bullets[i].Update() {
			g.bullets = append(g.bullets[:i], g.bullets[i+1:]...)
			i--
		}
	}

	// Spawn enemies
	g.spawnTimer++
	spawnThreshold := 60 - (minutes * 5)
	if minutes >= 8 {
		spawnThreshold = 15 // Increased spawning at minute 8
	} else if spawnThreshold < 25 {
		spawnThreshold = 25
	}

	// Minute 10 Wipe and Void Wave
	if minutes >= 10 {
		if minutes == 10 {
			framesInMinute := g.timer % 3600
			g.shakeIntensity = 1.0 + (float64(framesInMinute)/3600.0)*15.0
			if g.shakeFrames < 2 {
				g.shakeFrames = 2
			}
		} else {
			g.shakeIntensity = 10.0
		}
		if !g.event10 {
			for _, e := range g.enemies {
				g.particles.Explosion(e.x, e.y, color.RGBA{255, 100, 255, 255}, 15, 1.0)
				PlayExplosion()
			}
			g.enemies = nil
			StartAlarm()
			g.shakeFrames = 120 // Initial big blast shake
			g.event10 = true
		}
		// Special spawn for minute 10: Purple Elites
		if g.spawnTimer > 40 {
			health := 5 + rand.Intn(11)         // 5 to 15 HP
			speed := 1.0 + rand.Float64()*3.5   // 2.0 to 4.5 Speed
			size := 60.0 + rand.Float64()*100.0 // 60 to 100 Size

			voidElite := NewEnemy(10)
			voidElite.etype = TypeElite
			voidElite.hp = health
			voidElite.maxHP = health
			voidElite.size = size
			voidElite.speed = speed
			g.enemies = append(g.enemies, voidElite)
			g.spawnTimer = 0
		}
		// Skip normal spawning during minute 10 event
	} else if minutes < 10 && g.spawnTimer > spawnThreshold {
		// More conservative wave scaling
		count := 1 + (minutes / 4)
		if minutes > 2 && g.timer%180 == 0 { // Small chance for double wave after 2 mins
			count += 2
		}
		for i := 0; i < count; i++ {
			g.enemies = append(g.enemies, NewEnemy(minutes))
		}
		g.spawnTimer = 0
	}

	// Update enemies
	for i := 0; i < len(g.enemies); i++ {
		e := g.enemies[i]
		e.Update(g.player.x, g.player.y)

		// Collision with player
		dx := e.x - g.player.x
		dy := e.y - g.player.y
		if math.Sqrt(dx*dx+dy*dy) < (e.size/2 + 10) {
			// Damage calculation
			damage := 15.0
			oneShot := false

			switch e.etype {
			case TypeBoss:
				if g.isFinalBoss {
					oneShot = true
				} else {
					damage = g.player.maxHP * 0.75
				}
			case TypeElite:
				if minutes >= 10 {
					damage = g.player.maxHP * 0.5
				} else {
					damage = 35.0
				}
			case TypeCharger:
				damage = 25.0
			}

			if oneShot {
				g.gameOver = true
					PlayDefeat()
				g.particles.Explosion(g.player.x, g.player.y, color.RGBA{0, 255, 255, 255}, 50, 2.0)
				return nil
			}

			if g.player.TakeDamage(damage) {
				g.particles.Explosion(e.x, e.y, color.RGBA{255, 100, 100, 255}, 15, 1.0)
				g.shakeFrames = 15
				g.shakeIntensity = 20.0
				PlayHit()
				if g.player.hp <= 0 {
					g.gameOver = true
					PlayDefeat()
					g.particles.Explosion(g.player.x, g.player.y, color.RGBA{0, 255, 255, 255}, 50, 2.0)
				}
				g.enemies = append(g.enemies[:i], g.enemies[i+1:]...)
				i--
				continue
			}
		}

		// Collision with bullets
		for j := 0; j < len(g.bullets); j++ {
			b := g.bullets[j]
			bdx := e.x - b.x
			bdy := e.y - b.y
			if math.Sqrt(bdx*bdx+bdy*bdy) < (e.size * 0.8) {
				explClr := color.RGBA{255, 100, 100, 255}
				explSize := 1.0
				scoreGain := 100
				if e.etype == TypeCharger {
					explClr = color.RGBA{255, 200, 50, 255}
					explSize = 1.5
					scoreGain = 250
				}
				if e.etype == TypeBoss {
					explClr = color.RGBA{255, 50, 255, 255}
					explSize = 3.0
					scoreGain = 5000
				}

				e.hp--
				if e.hp <= 0 {
					if e.etype == TypeBoss && g.isFinalBoss {
						g.victory = true
					PlayVictory()
					}
					g.particles.Explosion(e.x, e.y, explClr, 20, explSize)
					PlayExplosion()
					g.enemies = append(g.enemies[:i], g.enemies[i+1:]...)
					g.score += scoreGain
					g.kills++
					g.shakeFrames = int(10 * explSize)
					i--
				}
				g.bullets = append(g.bullets[:j], g.bullets[j+1:]...)
				break
			}
		}
	}

	// Update stars
	for i := range g.stars {
		g.stars[i].y += g.stars[i].speed
		if g.stars[i].y > screenHeight {
			g.stars[i].y = 0
			g.stars[i].x = rand.Float64() * screenWidth
		}
	}


	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if !g.started {
		// Pure black background for splash
		g.offscreen.Fill(color.Black)

		// Parallax Stars only
		for _, s := range g.stars {
			ebitenutil.DrawRect(g.offscreen, s.x, s.y, s.size, s.size, s.color)
		}

		op := &ebiten.DrawImageOptions{}
		screen.DrawImage(g.offscreen, op)

		// Progressive Fade-in for Splash Image and UI
		alpha := float64(g.splashTimer) / 240.0 // 4 second slow fade
		if alpha > 1.0 { alpha = 1.0 }

		// Draw Splash Image with Alpha
		if splashImage != nil {
			slop := &ebiten.DrawImageOptions{}
			sw, sh := splashImage.Bounds().Dx(), splashImage.Bounds().Dy()
			scaleX := float64(screenWidth) / float64(sw)
			scaleY := float64(screenHeight) / float64(sh)
			scale := scaleX
			if scaleY > scale { scale = scaleY }
			slop.GeoM.Scale(scale, scale)
			slop.GeoM.Translate(float64(screenWidth)/2-float64(sw)*scale/2, float64(screenHeight)/2-float64(sh)*scale/2)
			slop.ColorScale.ScaleWithColor(color.RGBA{255, 255, 255, uint8(255 * alpha)})
			screen.DrawImage(splashImage, slop)
		}

		// Draw Dark Overlay for text contrast
		if alpha > 0.5 {
			textOverlay := ebiten.NewImage(screenWidth, screenHeight)
			textOverlay.Fill(color.RGBA{0, 0, 0, uint8(120 * (alpha-0.5)*2)})
			screen.DrawImage(textOverlay, nil)
		}

		// Draw Logo with Alpha - Positioned at top
		logoImg := ebiten.NewImage(screenWidth, screenHeight)
		DrawNeonTitle(logoImg, float64(screenWidth)/2, 100)
		lop := &ebiten.DrawImageOptions{}
		lop.ColorScale.ScaleWithColor(color.RGBA{255, 255, 255, uint8(255 * alpha)})
		screen.DrawImage(logoImg, lop)

		// Draw Press Enter with Alpha - Positioned at bottom with larger neon text
		if g.splashTimer > 120 { // Start showing after 2s
			enterAlpha := (float64(g.splashTimer) - 120.0) / 120.0
			if enterAlpha > 1.0 { enterAlpha = 1.0 }
			pulse := (math.Sin(float64(g.splashTimer)*0.03) + 1.0) / 2.0
			enterAlpha *= (0.4 + 0.6*pulse)
			
			promptImg := ebiten.NewImage(screenWidth, screenHeight)
			DrawNeonText(promptImg, "PRESS ENTER TO PLAY", float64(screenWidth)/2, 900, 30, color.RGBA{255, 255, 255, 255})
			pop := &ebiten.DrawImageOptions{}
			pop.ColorScale.ScaleWithColor(color.RGBA{255, 255, 255, uint8(255 * enterAlpha)})
			screen.DrawImage(promptImg, pop)
		}
		return
	}

	bgColor := color.RGBA{2, 2, 10, 255}
	if g.isFinalBoss {
		bgColor = color.RGBA{40, 10, 10, 255}
	}
	g.offscreen.Fill(bgColor)

	for _, s := range g.stars {
		ebitenutil.DrawRect(g.offscreen, s.x, s.y, s.size, s.size, s.color)
	}

	g.particles.Draw(g.offscreen)
	for _, b := range g.bullets {
		b.Draw(g.offscreen)
	}
	for _, e := range g.enemies {
		e.Draw(g.offscreen, g.showUI)
	}

	if !g.gameOver {
		g.player.Draw(g.offscreen)
	}

	g.drawUI(g.offscreen)

	op := &ebiten.DrawImageOptions{}
	if g.shakeFrames > 0 {
		intensity := g.shakeIntensity
		if intensity == 0 { intensity = 10.0 }
		op.GeoM.Translate((rand.Float64()-0.5)*intensity, (rand.Float64()-0.5)*intensity)
	}
	screen.DrawImage(g.offscreen, op)

	if g.isPaused {
		overlay := ebiten.NewImage(screenWidth, screenHeight)
		overlay.Fill(color.RGBA{0, 0, 0, 180})
		screen.DrawImage(overlay, nil)

		DrawNeonText(screen, "PAUSED", float64(screenWidth)/2, 300, 60, color.RGBA{0, 255, 255, 255})

		options := []string{"RESUME", "RESTART", "QUIT"}
		for i, opt := range options {
			clr := color.RGBA{255, 255, 255, 150}
			size := 30.0
			if i == g.menuSelection {
				clr = color.RGBA{255, 200, 0, 255}
				size = 40.0
			}
			DrawNeonText(screen, opt, float64(screenWidth)/2, float64(500+i*100), size, clr)
		}
	}
}

func (g *Game) drawUI(screen *ebiten.Image) {
	s := fmt.Sprintf("SCORE: %06d | KILLS: %d | TIME: %02d:%02d", g.score, g.kills, (g.timer/60)/60, (g.timer/60)%60)
	ebitenutil.DebugPrintAt(screen, s, 20, 20)

	aimStatus := "ON"
	if !g.autoAim {
		aimStatus = "OFF (Mouse)"
	}
	ebitenutil.DebugPrintAt(screen, "AUTO-AIM [L]: "+aimStatus, 20, 40)

	displayHealth := "ON"
	if !g.showUI {
		displayHealth = "OFF"
	}
	ebitenutil.DebugPrintAt(screen, "DISPLAY HEALTH [I]: "+displayHealth, 20, 60)

	// Player HP Bar at top center
	barW := 400.0
	barH := 20.0
	bx := float64(screenWidth)/2 - barW/2
	by := 20.0

	// BG
	ebitenutil.DrawRect(screen, bx, by, barW, barH, color.RGBA{50, 0, 0, 150})
	// Fill
	hpRatio := g.player.hp / g.player.maxHP
	if hpRatio < 0 {
		hpRatio = 0
	}
	hpClr := color.RGBA{uint8(255 * (1 - hpRatio)), uint8(255 * hpRatio), uint8(255 * hpRatio), 200}
	ebitenutil.DrawRect(screen, bx, by, barW*hpRatio, barH, hpClr)

	// Text
	hpStr := fmt.Sprintf("HULL INTEGRITY: %d%%", int(hpRatio*100))
	ebitenutil.DebugPrintAt(screen, hpStr, int(bx+barW/2-60), int(by+2))

	// Draw Top 5 Highscores
	ebitenutil.DebugPrintAt(screen, "TOP SCORES:", 20, 100)
	top := g.scores.GetTop(5)
	for i, entry := range top {
		timeStr := fmt.Sprintf("%02d:%02d", entry.Time/60, entry.Time%60)

		// Colored indicator: Green for Win, Red for Loss
		indicatorClr := color.RGBA{255, 50, 50, 255}
		if entry.Victory {
			indicatorClr = color.RGBA{50, 255, 100, 255}
		}
		ebitenutil.DrawRect(screen, 10, float64(120+i*20), 5, 10, indicatorClr)

		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d. %-10s %06d [%s] K:%d", i+1, entry.Name, entry.Score, timeStr, entry.Kills), 20, 120+i*20)
	}

	if g.gameOver || g.victory {
		// Dim the background
		overlay := ebiten.NewImage(screenWidth, screenHeight)
		if g.victory {
			overlay.Fill(color.RGBA{0, 50, 0, 180}) // Greenish for victory
		} else {
			overlay.Fill(color.RGBA{0, 0, 0, 180}) // Black for loss
		}
		screen.DrawImage(overlay, nil)

		mainMsg := "NEON SPACE SURVIVOR"
		statusMsg := "GAME OVER"
		if g.victory {
			statusMsg = "MISSION ACCOMPLISHED"
		}

		ebitenutil.DebugPrintAt(screen, mainMsg, screenWidth/2-60, screenHeight/2-60)

		if !g.saved {
			ebitenutil.DebugPrintAt(screen, "NEW HIGH SCORE!", screenWidth/2-45, screenHeight/2-20)
			ebitenutil.DebugPrintAt(screen, "ENTER YOUR NAME:", screenWidth/2-55, screenHeight/2+10)
			ebitenutil.DebugPrintAt(screen, "> "+g.playerName+"_", screenWidth/2-40, screenHeight/2+30)
			ebitenutil.DebugPrintAt(screen, "PRESS ENTER TO SAVE", screenWidth/2-60, screenHeight/2+60)
		} else {
			ebitenutil.DebugPrintAt(screen, statusMsg, screenWidth/2-40, screenHeight/2)
			ebitenutil.DebugPrintAt(screen, "SCORE SAVED!", screenWidth/2-35, screenHeight/2+20)
			ebitenutil.DebugPrintAt(screen, "PRESS 'R' TO RESTART", screenWidth/2-65, screenHeight/2+60)
		}
	}
}

func drawWithGlow(screen *ebiten.Image, drawFunc func(*ebiten.Image), intensity float64) {
	// Simple bloom effect: draw multiple times with additive blending (if supported)
	// For now, we'll just draw the object normally.
	// To do real glow we'd need an offscreen buffer and blur.
	drawFunc(screen)
}

func (g *Game) Reset() {
	ResetMusic()
	newGame := NewGame(false)
	// Preserve the scores manager so we don't reload from disk unnecessarily
	// but the NewGame already reloads them, which is fine.
	*g = *newGame
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}
