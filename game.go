package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Star struct {
	x, y  float64
	speed float64
	size  float64
	color color.RGBA
}

type GameMode int

const (
	ModeRegular GameMode = iota
	ModeWave
	ModeBossRush
)

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
	event8         bool
	event8Started  bool
	isFinalBoss    bool
	showUI         bool
	autoAim        bool
	victory        bool
	started        bool
	kills          int
	splashTimer    int
	isPaused       bool
	menuSelection  int
	gameMode       GameMode
	modeSelection  int
	// Cached images
	textOverlay     *ebiten.Image
	pauseOverlay    *ebiten.Image
	gameOverOverlay *ebiten.Image
	victoryOverlay  *ebiten.Image
	logoImg         *ebiten.Image
	promptImg       *ebiten.Image
	wipeTimer       int
	overdriveTimer  int
}

func NewGame(start bool) *Game {
	if start {
		LoadSplash()
	}
	g := &Game{
		player:          NewPlayer(),
		particles:       NewParticleSystem(),
		stars:           make([]Star, 150),
		offscreen:       ebiten.NewImage(screenWidth, screenHeight),
		scores:          NewHighscoreManager(),
		showUI:          true,
		autoAim:         false,
		started:         !start, // If start is false (reset), then started is true
		textOverlay:     ebiten.NewImage(screenWidth, screenHeight),
		pauseOverlay:    ebiten.NewImage(screenWidth, screenHeight),
		gameOverOverlay: ebiten.NewImage(screenWidth, screenHeight),
		victoryOverlay:  ebiten.NewImage(screenWidth, screenHeight),
		logoImg:         ebiten.NewImage(screenWidth, screenHeight),
		promptImg:       ebiten.NewImage(screenWidth, screenHeight),
	}
	g.pauseOverlay.Fill(color.RGBA{0, 0, 0, 180})
	g.gameOverOverlay.Fill(color.RGBA{0, 0, 0, 180})
	g.victoryOverlay.Fill(color.RGBA{0, 50, 0, 180})
	// Pre-render logo and prompt if possible, or just use them as targets
	DrawNeonTitle(g.logoImg, float64(screenWidth)/2, 100)
	DrawNeonText(g.promptImg, "PRESS ENTER TO PLAY", float64(screenWidth)/2, 900, 30, color.RGBA{255, 255, 255, 255})
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

func (g *Game) HandleEscape() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) && g.started && !g.gameOver && !g.victory {
		g.isPaused = !g.isPaused
		g.menuSelection = 0
	}
	if g.isPaused {
		if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
			g.menuSelection--
			if g.menuSelection < 0 {
				g.menuSelection = 2
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
			g.menuSelection++
			if g.menuSelection > 3 {
				g.menuSelection = 0
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			switch g.menuSelection {
			case 0: // Resume
				g.isPaused = false
			case 1: // Reset
				g.isPaused = false
				g.Reset()
			case 2: // Main Menu
				g.Reset()
				g.started = false
				g.gameOver = false
				g.victory = false
				g.saved = false
				g.playerName = ""
				g.timer = 0
				g.splashTimer = 0
				g.gameMode = ModeRegular
				g.modeSelection = 0
				g.wipeTimer = 0
				g.overdriveTimer = 0
			case 3: // Quit
				return ebiten.Termination
			}
		}
		return nil
	}
	return nil
}

func (g *Game) HandleEndGame() error {
	if g.gameOver || g.victory {
		g.shakeFrames = 0
		g.overdriveTimer = 0
		if !g.saved {
			// Handle name input
			g.playerName += string(ebiten.AppendInputChars(nil))
			if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(g.playerName) > 0 {
				g.playerName = g.playerName[:len(g.playerName)-1]
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyEnter) && len(g.playerName) > 0 {
				g.scores.AddScore(g.playerName, g.score, g.timer/60, g.victory, g.kills, int(g.gameMode))
				g.saved = true
			}
		} else {
			if inpututil.IsKeyJustPressed(ebiten.KeyR) {
				g.Reset()
			}
		}
		return nil
	}
	return nil
}

func (g *Game) HandleKeys() {
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
}

func (g *Game) WipeEnnemies(includeBosses bool) {
	if len(g.enemies) > 0 {
		PlayExplosion()
	}
	newEnemies := g.enemies[:0]
	for _, e := range g.enemies {
		if includeBosses || e.etype != TypeBoss {
			g.particles.Explosion(e.x, e.y, color.RGBA{255, 100, 255, 255}, 15, 1.0)
			e.Deactivate()
		} else {
			newEnemies = append(newEnemies, e)
		}
	}
	g.enemies = newEnemies
}

func (g *Game) Update() error {
	if err := g.HandleEscape(); err != nil {
		return err
	}
	if g.isPaused {
		return nil
	}

	if err := g.HandleEndGame(); err != nil {
		return err
	}
	if g.gameOver || g.victory {
		return nil
	}

	UpdateMusic(g.timer, g.started)
	if g.shakeFrames > 0 {
		g.shakeFrames--
	}
	g.HandleKeys()

	// Handle Splash Screen
	if !g.started {
		g.splashTimer++
		if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
			g.modeSelection--
			if g.modeSelection < 0 {
				g.modeSelection = 2
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
			g.modeSelection++
			if g.modeSelection > 2 {
				g.modeSelection = 0
			}
		}
		if g.splashTimer > 10 && inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.started = true
			g.gameMode = GameMode(g.modeSelection)
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
	if g.wipeTimer > 0 {
		g.wipeTimer--
		g.WipeEnnemies(false)
	}
	useAutoAim := g.autoAim
	if g.overdriveTimer > 0 {
		g.overdriveTimer--
		useAutoAim = true
		g.shakeFrames = 2
		g.shakeIntensity = 20.0 // Much more intense
	}
	minutes := g.timer / 3600
	if minutes < 10 && g.overdriveTimer <= 0 {
		g.shakeIntensity = 0
	}
	// Calculate fire rate based on minutes (starts at 12, decreases by 1 every minute, min 3)
	fireRate := 12 - minutes
	if minutes >= 8 {
		fireRate = 2 // Massive boost at minute 8
	} else if fireRate < 3 {
		fireRate = 3
	}
	if g.overdriveTimer > 0 {
		fireRate = 1 // Off the roof
	}
	g.player.Update(g.particles, &g.bullets, g.enemies, useAutoAim, fireRate)
	g.particles.Update()
	// Mode specific boss spawning
	if g.gameMode == ModeBossRush {
		if g.timer > 0 && g.timer%600 == 0 { // Boss every 10 seconds
			g.enemies = append(g.enemies, NewBoss(minutes))
			PlayBossAppears()
			g.shakeFrames = 20
		}
	} else {
		// Spawn boss every minute
		if g.timer > 0 && g.timer%3600 == 0 && minutes < 11 { // 60fps * 60s
			g.enemies = append(g.enemies, NewBoss(minutes))
			PlayBossAppears()
			g.shakeFrames = 20
		}
	}
	// Minute 11 Final Boss
	if minutes == 11 && !g.isFinalBoss {
		g.isFinalBoss = true
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
			// Swap-and-pop (pooling handled inside Update)
			g.bullets[i] = g.bullets[len(g.bullets)-1]
			g.bullets = g.bullets[:len(g.bullets)-1]
			i--
		}
	}
	// Spawn enemies
	g.spawnTimer++
	spawnThreshold := 60 - (minutes * 5)
	if spawnThreshold < 25 {
		spawnThreshold = 25
	}
	// Minute 8 Event: Gradual Shake then Swarm
	if minutes == 8 {
		framesInMinute := g.timer % 3600
		if framesInMinute < 300 { // 5 seconds duration
			g.WipeEnnemies(true)
			StopMusic()
			StartAlarm()
			g.shakeFrames = 2
			g.shakeIntensity = 30.0 // Shaking a lot
			spawnThreshold = 999999 // No spawn
			g.event8 = true
		} else if framesInMinute >= 300 && g.event8 {
			// Transition at exactly 5 seconds (runs once)
			StopAlarm()
			StartMusic()
			g.event8 = false

			for i := 0; i < 3; i++ {
				g.enemies = append(g.enemies, NewBoss(minutes))
			}
			PlayBossAppears()
		}
		if framesInMinute >= 300 {
			spawnThreshold = 15 // Swarm intensity resumes
		}
	} else if minutes > 8 {
		spawnThreshold = 15
	}

	if minutes >= 8 && minutes < 10 && !g.event8 {
		g.shakeFrames = 2
		g.shakeIntensity = 10.0 // Shaking a lot
	}

	// Minute 10 Wipe and Void Wave
	if minutes >= 10 {
		g.HandleMinutesEndgame(minutes)
	} else if minutes < 10 && g.spawnTimer > spawnThreshold && g.wipeTimer <= 0 {
		if g.gameMode == ModeBossRush {
			// No small enemies in Boss Rush
			g.spawnTimer = 0
		} else {
			// More conservative wave scaling
			count := 1 + (minutes / 4)
			if g.gameMode == ModeWave {
				count *= 2 // Double enemies in Wave Mode
			}
			if minutes > 2 && g.timer%180 == 0 { // Small chance for double wave after 2 mins
				count += 2
			}
			for i := 0; i < count; i++ {
				g.enemies = append(g.enemies, NewEnemy(minutes))
			}
			g.spawnTimer = 0
		}
	}
	g.HandleCollisionsAndEnnemies(minutes)
	return nil
}

func (g *Game) HandleMinutesEndgame(minutes int) {
	var spawnTimer int
	if minutes == 10 {
		framesInMinute := g.timer % 3600
		g.shakeIntensity = 1.0 + (float64(framesInMinute)/3600.0)*15.0
		if g.shakeFrames < 2 {
			g.shakeFrames = 2
		}
		spawnTimer = 20
	} else {
		spawnTimer = 40
		g.shakeIntensity = 10.0
	}
	if !g.event10 {
		g.WipeEnnemies(true)
		StartAlarm()
		g.shakeFrames = 120 // Initial big blast shake
		g.event10 = true
	}
	// Special spawn for minute 10: Purple Elites
	if g.spawnTimer > spawnTimer {
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
}

func (g *Game) HandleCollisionsAndEnnemies(minutes int) error {
	// Update enemies
	for i := 0; i < len(g.enemies); i++ {
		e := g.enemies[i]
		e.Update(g.player.x, g.player.y)
		// Collision with player
		dx := e.x - g.player.x
		dy := e.y - g.player.y
		radius := e.size/2 + 10
		if dx*dx+dy*dy < radius*radius {
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
				// Deactivate and Swap-and-pop (Bosses don't die on collision)
				if e.etype != TypeBoss {
					e.Deactivate()
					g.enemies[i] = g.enemies[len(g.enemies)-1]
					g.enemies = g.enemies[:len(g.enemies)-1]
					i--
					continue
				}
			}
		}
		// Collision with bullets
		for j := 0; j < len(g.bullets); j++ {
			b := g.bullets[j]
			bdx := e.x - b.x
			bdy := e.y - b.y
			bRadius := e.size * 0.8
			if bdx*bdx+bdy*bdy < bRadius*bRadius {
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
					if e.etype == TypeBoss {
						if g.isFinalBoss {
							g.victory = true
							PlayVictory()
						}
						g.wipeTimer = 120 // 2 second wipe
					}
					g.particles.Explosion(e.x, e.y, explClr, 20, explSize)
					PlayExplosion()
					// Deactivate and Swap-and-pop
					e.Deactivate()
					g.enemies[i] = g.enemies[len(g.enemies)-1]
					g.enemies = g.enemies[:len(g.enemies)-1]
					g.score += scoreGain
					g.kills++
					if g.kills > 0 && g.kills%150 == 0 {
						g.overdriveTimer = 300 // 5 second overdrive
						g.shakeFrames = 100
						g.shakeIntensity = 20.0
						PlayOverdrive()
					} else {
						newShake := int(10 * explSize)
						if newShake > g.shakeFrames {
							g.shakeFrames = newShake
						}
					}
					i--
				}
				// Deactivate bullet (swap-and-pop handled here)
				b.active = false
				bulletPool = append(bulletPool, b)
				g.bullets[j] = g.bullets[len(g.bullets)-1]
				g.bullets = g.bullets[:len(g.bullets)-1]
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
		for i := range g.stars {
			s := &g.stars[i]
			sharedDrawOp.GeoM.Reset()
			sharedDrawOp.ColorScale.Reset()
			sharedDrawOp.GeoM.Scale(s.size, s.size)
			sharedDrawOp.GeoM.Translate(s.x, s.y)
			sharedDrawOp.ColorScale.ScaleWithColor(s.color)
			g.offscreen.DrawImage(whitePixel, &sharedDrawOp)
		}
		sharedDrawOp.GeoM.Reset()
		sharedDrawOp.ColorScale.Reset()
		screen.DrawImage(g.offscreen, &sharedDrawOp)
		// Progressive Fade-in for Splash Image and UI
		alpha := float64(g.splashTimer) / 240.0 // 4 second slow fade
		if alpha > 1.0 {
			alpha = 1.0
		}
		// Draw Splash Image with Alpha
		if splashImage != nil {
			sharedDrawOp.GeoM.Reset()
			sharedDrawOp.ColorScale.Reset()
			sw, sh := splashImage.Bounds().Dx(), splashImage.Bounds().Dy()
			scaleX := float64(screenWidth) / float64(sw)
			scaleY := float64(screenHeight) / float64(sh)
			scale := scaleX
			if scaleY > scale {
				scale = scaleY
			}
			sharedDrawOp.GeoM.Scale(scale, scale)
			sharedDrawOp.GeoM.Translate(float64(screenWidth)/2-float64(sw)*scale/2, float64(screenHeight)/2-float64(sh)*scale/2)
			sharedDrawOp.ColorScale.ScaleWithColor(color.RGBA{255, 255, 255, uint8(255 * alpha)})
			screen.DrawImage(splashImage, &sharedDrawOp)
		}
		// Draw Dark Overlay for text contrast
		if alpha > 0.5 {
			g.textOverlay.Fill(color.RGBA{0, 0, 0, uint8(120 * (alpha - 0.5) * 2)})
			screen.DrawImage(g.textOverlay, nil)
		}
		// Draw Logo with Alpha - Positioned at top
		sharedDrawOp.GeoM.Reset()
		sharedDrawOp.ColorScale.Reset()
		sharedDrawOp.ColorScale.ScaleWithColor(color.RGBA{255, 255, 255, uint8(255 * alpha)})
		screen.DrawImage(g.logoImg, &sharedDrawOp)
		// Draw Game Mode Menu with Alpha - Positioned at bottom
		if g.splashTimer > 120 { // Start showing after 2s
			enterAlpha := (float64(g.splashTimer) - 120.0) / 120.0
			if enterAlpha > 1.0 {
				enterAlpha = 1.0
			}
			modes := []string{"REGULAR MODE", "WAVE MODE", "BOSS RUSH"}
			for i, mode := range modes {
				clr := color.RGBA{200, 200, 200, uint8(150 * enterAlpha)}
				size := 20.0
				if i == g.modeSelection {
					pulse := (math.Sin(float64(g.splashTimer)*0.1) + 1.0) / 2.0
					clr = color.RGBA{uint8(100 + 155*pulse), 255, 255, uint8(255 * enterAlpha)}
					size = 28.0
				}
				DrawNeonText(screen, mode, float64(screenWidth)/2, float64(750+i*60), size, clr)
			}
		}
		return
	}
	bgColor := color.RGBA{2, 2, 10, 255}
	if g.isFinalBoss {
		bgColor = color.RGBA{40, 10, 10, 255}
	}
	g.offscreen.Fill(bgColor)
	// Parallax Stars
	for i := range g.stars {
		s := &g.stars[i]
		sharedDrawOp.GeoM.Reset()
		sharedDrawOp.ColorScale.Reset()
		sharedDrawOp.GeoM.Scale(s.size, s.size)
		sharedDrawOp.GeoM.Translate(s.x, s.y)
		sharedDrawOp.ColorScale.ScaleWithColor(s.color)
		g.offscreen.DrawImage(whitePixel, &sharedDrawOp)
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
	sharedDrawOp.GeoM.Reset()
	sharedDrawOp.ColorScale.Reset()
	if g.shakeFrames > 0 {
		intensity := g.shakeIntensity
		if intensity == 0 {
			intensity = 10.0
		}
		sharedDrawOp.GeoM.Translate((rand.Float64()-0.5)*intensity, (rand.Float64()-0.5)*intensity)
	}
	screen.DrawImage(g.offscreen, &sharedDrawOp)
	if g.isPaused {
		screen.DrawImage(g.pauseOverlay, nil)
		DrawNeonText(screen, "PAUSED", float64(screenWidth)/2, 300, 60, color.RGBA{0, 255, 255, 255})
		options := []string{"RESUME", "RESTART", "MAIN MENU", "QUIT"}
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
	DrawNeonText(screen, s, 300, 20, 15, color.RGBA{255, 255, 255, 200})
	if g.overdriveTimer > 0 {
		pulse := math.Sin(float64(g.timer)*0.2)*0.5 + 0.5
		DrawNeonText(screen, "OVERDRIVE ACTIVE", float64(screenWidth)/2, 100, 30, color.RGBA{255, uint8(255 * pulse), 0, 255})
	}
	displayHealth := "ON"
	if !g.showUI {
		displayHealth = "OFF"
	}
	DrawNeonText(screen, "DISPLAY HEALTH [I]: "+displayHealth, 130, 60, 12, color.RGBA{255, 255, 255, 150})
	// Player HP Bar at top center
	barW := 400.0
	barH := 20.0
	bx := float64(screenWidth)/2 - barW/2
	by := 20.0
	// BG
	bgOp := &ebiten.DrawImageOptions{}
	bgOp.GeoM.Scale(barW, barH)
	bgOp.GeoM.Translate(bx, by)
	bgOp.ColorScale.ScaleWithColor(color.RGBA{50, 0, 0, 150})
	screen.DrawImage(whitePixel, bgOp)
	// Fill
	hpRatio := g.player.hp / g.player.maxHP
	if hpRatio < 0 {
		hpRatio = 0
	}
	hpClr := color.RGBA{uint8(255 * (1 - hpRatio)), uint8(255 * hpRatio), uint8(255 * hpRatio), 200}

	fillOp := &ebiten.DrawImageOptions{}
	fillOp.GeoM.Scale(barW*hpRatio, barH)
	fillOp.GeoM.Translate(bx, by)
	fillOp.ColorScale.ScaleWithColor(hpClr)
	screen.DrawImage(whitePixel, fillOp)
	// Text
	hpStr := fmt.Sprintf("HULL INTEGRITY: %d%%", int(hpRatio*100))
	DrawNeonText(screen, hpStr, bx+barW/2, by+10, 10, color.RGBA{255, 255, 255, 255})
	// Draw Top 5 Highscores
	DrawNeonText(screen, "TOP SCORES", 80, 100, 15, color.RGBA{0, 255, 255, 255})
	top := g.scores.GetTop(5, int(g.gameMode))
	for i, entry := range top {
		timeStr := fmt.Sprintf("%02d:%02d", entry.Time/60, entry.Time%60)
		// Colored indicator: Green for Win, Red for Loss
		indicatorClr := color.RGBA{255, 50, 50, 255}
		if entry.Victory {
			indicatorClr = color.RGBA{50, 255, 100, 255}
		}

		indOp := &ebiten.DrawImageOptions{}
		indOp.GeoM.Scale(10, 10)
		indOp.GeoM.Translate(20, float64(135+i*30))
		indOp.ColorScale.ScaleWithColor(indicatorClr)
		screen.DrawImage(whitePixel, indOp)

		modeChar := "R"
		switch entry.GameMode {
		case int(ModeWave):
			modeChar = "W"
		case int(ModeBossRush):
			modeChar = "B"
		}

		y := float64(140 + i*30)
		clr := color.RGBA{255, 255, 255, 200}
		DrawNeonTextLeft(screen, fmt.Sprintf("%d.", i+1), 40, y, 12, clr)
		DrawNeonTextLeft(screen, entry.Name, 75, y, 12, clr)
		DrawNeonTextLeft(screen, fmt.Sprintf("%06d", entry.Score), 190, y, 12, clr)
		DrawNeonTextLeft(screen, timeStr, 270, y, 12, clr)
		DrawNeonTextLeft(screen, fmt.Sprintf("K:%d", entry.Kills), 335, y, 12, clr)
		DrawNeonTextLeft(screen, "M:"+modeChar, 400, y, 12, clr)
	}
	if g.gameOver || g.victory {
		// Dim the background
		if g.victory {
			screen.DrawImage(g.victoryOverlay, nil)
		} else {
			screen.DrawImage(g.gameOverOverlay, nil)
		}
		mainMsg := "NEON SPACE SURVIVOR"
		statusMsg := "GAME OVER"
		if g.victory {
			statusMsg = "MISSION ACCOMPLISHED"
		}
		DrawNeonText(screen, mainMsg, screenWidth/2, screenHeight/2-60, 40, color.RGBA{255, 255, 255, 255})
		if !g.saved {
			DrawNeonText(screen, "NEW HIGH SCORE!", screenWidth/2, screenHeight/2-20, 25, color.RGBA{255, 255, 0, 255})
			DrawNeonText(screen, "ENTER YOUR NAME:", screenWidth/2, screenHeight/2+10, 20, color.RGBA{255, 255, 255, 200})
			DrawNeonText(screen, "> "+g.playerName+"_", screenWidth/2, screenHeight/2+40, 22, color.RGBA{0, 255, 255, 255})
			DrawNeonText(screen, "PRESS ENTER TO SAVE", screenWidth/2, screenHeight/2+80, 18, color.RGBA{255, 255, 255, 150})
		} else {
			DrawNeonText(screen, statusMsg, screenWidth/2, screenHeight/2, 35, color.RGBA{255, 255, 255, 255})
			DrawNeonText(screen, "SCORE SAVED!", screenWidth/2, screenHeight/2+30, 20, color.RGBA{0, 255, 100, 255})
			DrawNeonText(screen, "PRESS 'R' TO RESTART", screenWidth/2, screenHeight/2+80, 18, color.RGBA{255, 255, 255, 150})
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
