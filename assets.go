package main

import (
	"fmt"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	shipImage   *ebiten.Image
	bulletImage *ebiten.Image
	glowImage   *ebiten.Image
	splashImage *ebiten.Image
	particleImg *ebiten.Image
	alienImage  *ebiten.Image
	bossImage   *ebiten.Image
	// Templates for pre-rendered neon shapes (drawn in white for ColorScale)
	shipTemplate   *ebiten.Image
	alienTemplate  *ebiten.Image
	bossTemplate   *ebiten.Image
	bulletTemplate *ebiten.Image
	charTemplates  map[rune]*ebiten.Image
	whitePixel     *ebiten.Image
)

func init() {
	whitePixel = ebiten.NewImage(1, 1)
	whitePixel.Fill(color.White)
	// These are now mostly fallbacks as we use vectors
	shipImage = ebiten.NewImage(32, 32)
	bulletImage = ebiten.NewImage(8, 8)
	particleImg = ebiten.NewImage(4, 4)
	particleImg.Fill(color.White)
	glowImage = createGlowTexture(64, color.RGBA{0, 255, 255, 100})
	alienImage = ebiten.NewImage(32, 32)
	bossImage = ebiten.NewImage(64, 64)
	// Pre-render templates
	preRenderTemplates()
}
func preRenderTemplates() {
	white := color.RGBA{255, 255, 255, 255}
	// Ship (using size 32 as base)
	shipTemplate = ebiten.NewImage(128, 128)
	drawNeonShipRaw(shipTemplate, 64, 64, 64, white, 0)
	// Alien (using size 30 as base)
	alienTemplate = ebiten.NewImage(128, 128)
	drawNeonAlienRaw(alienTemplate, 64, 64, 60, white)
	// Boss (using size 100 as base)
	bossTemplate = ebiten.NewImage(512, 512)
	drawNeonBossRaw(bossTemplate, 256, 256, 200, white, 0)
	// Bullet (using size 4 as base)
	bulletTemplate = ebiten.NewImage(32, 32)
	drawNeonBulletRaw(bulletTemplate, 16, 16, 8, white)
	// Characters (A-Z, 0-9, etc.)
	charTemplates = make(map[rune]*ebiten.Image)
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789:!?.+-|/% "
	for _, char := range chars {
		img := ebiten.NewImage(32, 32)
		// Letter is size/2 wide (12) and size high (24).
		// To center in 32x32: x = (32-12)/2 = 10, y = (32-24)/2 = 4
		drawLetterRaw(img, char, 10, 4, 24, white)
		charTemplates[char] = img
	}
}
func drawLetterRaw(screen *ebiten.Image, char rune, x, y, size float32, clr color.RGBA) {
	switch char {
	case 'A':
		drawLetterA(screen, x, y, size, clr)
	case 'B':
		drawLetterB(screen, x, y, size, clr)
	case 'C':
		drawLetterC(screen, x, y, size, clr)
	case 'D':
		drawLetterD(screen, x, y, size, clr)
	case 'E':
		drawLetterE(screen, x, y, size, clr)
	case 'F':
		drawLetterF(screen, x, y, size, clr)
	case 'G':
		drawLetterG(screen, x, y, size, clr)
	case 'H':
		drawLetterH(screen, x, y, size, clr)
	case 'I':
		drawLetterI(screen, x, y, size, clr)
	case 'J':
		drawLetterJ(screen, x, y, size, clr)
	case 'K':
		drawLetterK(screen, x, y, size, clr)
	case 'L':
		drawLetterL(screen, x, y, size, clr)
	case 'M':
		drawLetterM(screen, x, y, size, clr)
	case 'N':
		drawLetterN(screen, x, y, size, clr)
	case 'O':
		drawLetterO(screen, x, y, size, clr)
	case 'P':
		drawLetterP(screen, x, y, size, clr)
	case 'Q':
		drawLetterQ(screen, x, y, size, clr)
	case 'R':
		drawLetterR(screen, x, y, size, clr)
	case 'S':
		drawLetterS(screen, x, y, size, clr)
	case 'T':
		drawLetterT(screen, x, y, size, clr)
	case 'U':
		drawLetterU(screen, x, y, size, clr)
	case 'V':
		drawLetterV(screen, x, y, size, clr)
	case 'W':
		drawLetterW(screen, x, y, size, clr)
	case 'X':
		drawLetterX(screen, x, y, size, clr)
	case 'Y':
		drawLetterY(screen, x, y, size, clr)
	case 'Z':
		drawLetterZ(screen, x, y, size, clr)
	case '0':
		drawDigit0(screen, x, y, size, clr)
	case '1':
		drawDigit1(screen, x, y, size, clr)
	case '2':
		drawDigit2(screen, x, y, size, clr)
	case '3':
		drawDigit3(screen, x, y, size, clr)
	case '4':
		drawDigit4(screen, x, y, size, clr)
	case '5':
		drawDigit5(screen, x, y, size, clr)
	case '6':
		drawDigit6(screen, x, y, size, clr)
	case '7':
		drawDigit7(screen, x, y, size, clr)
	case '8':
		drawDigit8(screen, x, y, size, clr)
	case '9':
		drawDigit9(screen, x, y, size, clr)
	case ':':
		drawColon(screen, x, y, size, clr)
	case '!':
		drawExclamation(screen, x, y, size, clr)
	case '?':
		drawQuestion(screen, x, y, size, clr)
	case '.':
		drawDot(screen, x, y, size, clr)
	case '+':
		drawPlus(screen, x, y, size, clr)
	case '-':
		drawMinus(screen, x, y, size, clr)
	case '|':
		drawPipe(screen, x, y, size, clr)
	case '/':
		drawSlash(screen, x, y, size, clr)
	case '%':
		drawPercent(screen, x, y, size, clr)
	}
}
func LoadSplash() {
	var err error
	splashImage, _, err = ebitenutil.NewImageFromFile("static/images/splashscreen.png")
	if err != nil {
		fmt.Printf("Failed to load splash image: %v\n", err)
		splashImage = ebiten.NewImage(screenWidth, screenHeight)
		splashImage.Fill(color.RGBA{50, 50, 70, 255})
	} else {
		fmt.Println("Splash image loaded successfully")
	}
}
func createGlowTexture(size int, clr color.RGBA) *ebiten.Image {
	img := ebiten.NewImage(size, size)
	center := float64(size) / 2
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - center
			dy := float64(y) - center
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < center {
				alpha := float64(clr.A) * math.Pow(1-dist/center, 2)
				img.Set(x, y, color.RGBA{clr.R, clr.G, clr.B, uint8(alpha)})
			}
		}
	}
	return img
}
func DrawNeonAlien(screen *ebiten.Image, x, y, size float64, clr color.RGBA) {
	op := &ebiten.DrawImageOptions{}
	scale := size / 60.0
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(x-64*scale, y-64*scale)
	op.ColorScale.ScaleWithColor(clr)
	screen.DrawImage(alienTemplate, op)
}
func drawNeonAlienRaw(screen *ebiten.Image, x, y, size float64, clr color.RGBA) {
	s := float32(size)
	xf, yf := float32(x), float32(y)
	path := vector.Path{}
	path.MoveTo(xf, yf-s/2)
	path.LineTo(xf+s/4, yf-s/4)
	path.LineTo(xf+s/2, yf)
	path.LineTo(xf+s/4, yf+s/4)
	path.LineTo(xf, yf+s/2)
	path.LineTo(xf-s/4, yf+s/4)
	path.LineTo(xf-s/2, yf)
	path.LineTo(xf-s/4, yf-s/4)
	path.Close()
	dop := &vector.DrawPathOptions{}
	dop.ColorScale.ScaleWithColor(clr)
	vector.StrokePath(screen, &path, &vector.StrokeOptions{Width: 3}, dop)
	// Eyes
	vector.DrawFilledRect(screen, xf-4, yf-2, 2, 2, color.White, true)
	vector.DrawFilledRect(screen, xf+2, yf-2, 2, 2, color.White, true)
}
func DrawNeonBoss(screen *ebiten.Image, x, y, size float64, clr color.RGBA, rotation float64) {
	op := &ebiten.DrawImageOptions{}
	scale := size / 200.0
	op.GeoM.Translate(-256, -256)
	op.GeoM.Rotate(rotation)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(clr)
	screen.DrawImage(bossTemplate, op)
}
func drawNeonBossRaw(screen *ebiten.Image, x, y, size float64, clr color.RGBA, rotation float64) {
	s := float32(size)
	xf, yf := float32(x), float32(y)
	cos := float32(math.Cos(rotation))
	sin := float32(math.Sin(rotation))
	drawLineRot := func(x1, y1, x2, y2 float32) {
		rx1 := x1*cos - y1*sin + xf
		ry1 := x1*sin + y1*cos + yf
		rx2 := x2*cos - y2*sin + xf
		ry2 := x2*sin + y2*cos + yf
		vector.StrokeLine(screen, rx1, ry1, rx2, ry2, 6, clr, true)
	}
	drawLineRot(-s/4, -s/4, 0, -s/2)
	drawLineRot(0, -s/2, s/4, -s/4)
	drawLineRot(s/4, -s/4, s/4, s/4)
	drawLineRot(s/4, s/4, 0, s/2)
	drawLineRot(0, s/2, -s/4, s/4)
	drawLineRot(-s/4, s/4, -s/4, -s/4)
	drawLineRot(-s/4, -s/4, -s/2, -s/2)
	drawLineRot(-s/2, -s/2, -s/2, 0)
	drawLineRot(s/4, -s/4, s/2, -s/2)
	drawLineRot(s/2, -s/2, s/2, 0)
	drawLineRot(-s/6, s/2, -s/4, s/1.5)
	drawLineRot(s/6, s/2, s/4, s/1.5)
	vector.DrawFilledRect(screen, xf-5, yf-5, 10, 10, color.White, true)
}
func DrawNeonShip(screen *ebiten.Image, x, y, size float64, clr color.RGBA, rotation float64) {
	op := &ebiten.DrawImageOptions{}
	scale := size / 64.0
	op.GeoM.Translate(-64, -64)
	op.GeoM.Rotate(rotation)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(clr)
	screen.DrawImage(shipTemplate, op)
}
func drawNeonShipRaw(screen *ebiten.Image, x, y, size float64, clr color.RGBA, rotation float64) {
	s := float32(size)
	xf, yf := float32(x), float32(y)
	cos := float32(math.Cos(rotation))
	sin := float32(math.Sin(rotation))
	drawVec := func(x1, y1, x2, y2 float32) {
		rx1 := x1*cos - y1*sin + xf
		ry1 := x1*sin + y1*cos + yf
		rx2 := x2*cos - y2*sin + xf
		ry2 := x2*sin + y2*cos + yf
		vector.StrokeLine(screen, rx1, ry1, rx2, ry2, 5, clr, true)
	}
	drawVec(0, -s/2, -s/3, s/2)
	drawVec(-s/3, s/2, 0, s/3)
	drawVec(0, s/3, s/3, s/2)
	drawVec(s/3, s/2, 0, -s/2)
	drawVec(-s/6, s/3, s/6, s/3)
	vector.DrawFilledRect(screen, xf-4, yf-float32(size)/6, 8, 8, color.White, true)
}
func DrawNeonBullet(screen *ebiten.Image, x, y, size float64, clr color.RGBA) {
	op := &ebiten.DrawImageOptions{}
	scale := size / 8.0
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(x-16*scale, y-16*scale)
	op.ColorScale.ScaleWithColor(clr)
	screen.DrawImage(bulletTemplate, op)
}
func drawNeonBulletRaw(screen *ebiten.Image, x, y, size float64, clr color.RGBA) {
	xf, yf := float32(x), float32(y)
	s := float32(size)
	vector.DrawFilledRect(screen, xf-s/2, yf-s/2, s, s, color.White, true)
	vector.StrokeCircle(screen, xf, yf, s, 3, clr, true)
}
func DrawNeonTitle(screen *ebiten.Image, x, y float64) {
	xf, yf := float32(x), float32(y)
	s := float32(60) // Reduced from 80
	spacing := s * 0.8
	clr1 := color.RGBA{0, 255, 255, 255}
	drawLetterN(screen, xf-1.5*spacing, yf, s, clr1)
	drawLetterE(screen, xf-0.5*spacing, yf, s, clr1)
	drawLetterO(screen, xf+0.5*spacing, yf, s, clr1)
	drawLetterN(screen, xf+1.5*spacing, yf, s, clr1)
	clr2 := color.RGBA{255, 200, 0, 255}
	drawLetterS(screen, xf-2*spacing, yf+s*1.2, s, clr2)
	drawLetterP(screen, xf-spacing, yf+s*1.2, s, clr2)
	drawLetterA(screen, xf, yf+s*1.2, s, clr2)
	drawLetterC(screen, xf+spacing, yf+s*1.2, s, clr2)
	drawLetterE(screen, xf+2*spacing, yf+s*1.2, s, clr2)
	clr3 := color.RGBA{255, 0, 255, 255}
	drawLetterS(screen, xf-3.5*spacing, yf+s*2.4, s, clr3)
	drawLetterU(screen, xf-2.5*spacing, yf+s*2.4, s, clr3)
	drawLetterR(screen, xf-1.5*spacing, yf+s*2.4, s, clr3)
	drawLetterV(screen, xf-0.5*spacing, yf+s*2.4, s, clr3)
	drawLetterI(screen, xf+0.5*spacing, yf+s*2.4, s, clr3)
	drawLetterV(screen, xf+1.5*spacing, yf+s*2.4, s, clr3)
	drawLetterO(screen, xf+2.5*spacing, yf+s*2.4, s, clr3)
	drawLetterR(screen, xf+3.5*spacing, yf+s*2.4, s, clr3)
}
func drawLetterN(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x, y, x, y+size, 4, clr, true)
	vector.StrokeLine(s, x, y, x+size/2, y+size, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y, x+size/2, y+size, 4, clr, true)
}
func drawLetterE(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x, y, x+size/2, y, 4, clr, true)
	vector.StrokeLine(s, x, y+size/2, x+size/2, y+size/2, 4, clr, true)
	vector.StrokeLine(s, x, y+size, x+size/2, y+size, 4, clr, true)
	vector.StrokeLine(s, x, y, x, y+size, 4, clr, true)
}
func drawLetterO(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x, y, x+size/2, y, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y, x+size/2, y+size, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y+size, x, y+size, 4, clr, true)
	vector.StrokeLine(s, x, y+size, x, y, 4, clr, true)
}
func drawLetterS(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x, y, x+size/2, y, 4, clr, true)
	vector.StrokeLine(s, x, y, x, y+size/2, 4, clr, true)
	vector.StrokeLine(s, x, y+size/2, x+size/2, y+size/2, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y+size/2, x+size/2, y+size, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y+size, x, y+size, 4, clr, true)
}
func drawLetterP(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x, y, x, y+size, 4, clr, true)
	vector.StrokeLine(s, x, y, x+size/2, y, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y, x+size/2, y+size/2, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y+size/2, x, y+size/2, 4, clr, true)
}
func drawLetterA(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x, y+size, x+size/4, y, 4, clr, true)
	vector.StrokeLine(s, x+size/4, y, x+size/2, y+size, 4, clr, true)
	vector.StrokeLine(s, x+size/8, y+size/2, x+3*size/8, y+size/2, 4, clr, true)
}
func drawLetterC(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x+size/2, y, x, y, 4, clr, true)
	vector.StrokeLine(s, x, y, x, y+size, 4, clr, true)
	vector.StrokeLine(s, x, y+size, x+size/2, y+size, 4, clr, true)
}
func drawLetterU(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x, y, x, y+size, 4, clr, true)
	vector.StrokeLine(s, x, y+size, x+size/2, y+size, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y+size, x+size/2, y, 4, clr, true)
}
func drawLetterR(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x, y, x, y+size, 4, clr, true)
	vector.StrokeLine(s, x, y, x+size/2, y, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y, x+size/2, y+size/2, 4, clr, true)
	vector.StrokeLine(s, x, y+size/2, x+size/2, y+size/2, 4, clr, true)
	vector.StrokeLine(s, x, y+size/2, x+size/2, y+size, 4, clr, true)
}
func drawLetterV(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x, y, x+size/4, y+size, 4, clr, true)
	vector.StrokeLine(s, x+size/4, y+size, x+size/2, y, 4, clr, true)
}
func drawLetterI(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x+size/4, y, x+size/4, y+size, 4, clr, true)
}
func drawLetterT(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x, y, x+size/2, y, 4, clr, true)
	vector.StrokeLine(s, x+size/4, y, x+size/4, y+size, 4, clr, true)
}
func drawLetterL(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x, y, x, y+size, 4, clr, true)
	vector.StrokeLine(s, x, y+size, x+size/2, y+size, 4, clr, true)
}
func drawLetterY(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x, y, x+size/4, y+size/2, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y, x+size/4, y+size/2, 4, clr, true)
	vector.StrokeLine(s, x+size/4, y+size/2, x+size/4, y+size, 4, clr, true)
}
func drawLetterB(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x, y, x, y+size, 4, clr, true)
	vector.StrokeLine(s, x, y, x+size/2, y, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y, x+size/2, y+size/2, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y+size/2, x, y+size/2, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y+size/2, x+size/2, y+size, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y+size, x, y+size, 4, clr, true)
}
func drawLetterD(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x, y, x, y+size, 4, clr, true)
	vector.StrokeLine(s, x, y, x+size/4, y, 4, clr, true)
	vector.StrokeLine(s, x+size/4, y, x+size/2, y+size/4, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y+size/4, x+size/2, y+3*size/4, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y+3*size/4, x+size/4, y+size, 4, clr, true)
	vector.StrokeLine(s, x+size/4, y+size, x, y+size, 4, clr, true)
}
func drawLetterM(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x, y, x, y+size, 4, clr, true)
	vector.StrokeLine(s, x, y, x+size/4, y+size/2, 4, clr, true)
	vector.StrokeLine(s, x+size/4, y+size/2, x+size/2, y, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y, x+size/2, y+size, 4, clr, true)
}
func drawLetterQ(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	drawLetterO(s, x, y, size, clr)
	vector.StrokeLine(s, x+size/4, y+3*size/4, x+size/2, y+size, 4, clr, true)
}
func drawLetterH(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x, y, x, y+size, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y, x+size/2, y+size, 4, clr, true)
	vector.StrokeLine(s, x, y+size/2, x+size/2, y+size/2, 4, clr, true)
}
func drawLetterF(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x, y, x, y+size, 4, clr, true)
	vector.StrokeLine(s, x, y, x+size/2, y, 4, clr, true)
	vector.StrokeLine(s, x, y+size/2, x+size/2, y+size/2, 4, clr, true)
}
func drawLetterG(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x+size/2, y, x, y, 4, clr, true)
	vector.StrokeLine(s, x, y, x, y+size, 4, clr, true)
	vector.StrokeLine(s, x, y+size, x+size/2, y+size, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y+size, x+size/2, y+size/2, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y+size/2, x+size/4, y+size/2, 4, clr, true)
}
func drawLetterJ(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x+size/2, y, x+size/2, y+size, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y+size, x, y+size, 4, clr, true)
	vector.StrokeLine(s, x, y+size, x, y+3*size/4, 4, clr, true)
}
func drawLetterK(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x, y, x, y+size, 4, clr, true)
	vector.StrokeLine(s, x, y+size/2, x+size/2, y, 4, clr, true)
	vector.StrokeLine(s, x, y+size/2, x+size/2, y+size, 4, clr, true)
}
func drawLetterW(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x, y, x, y+size, 4, clr, true)
	vector.StrokeLine(s, x, y+size, x+size/4, y+size/2, 4, clr, true)
	vector.StrokeLine(s, x+size/4, y+size/2, x+size/2, y+size, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y+size, x+size/2, y, 4, clr, true)
}
func drawLetterX(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x, y, x+size/2, y+size, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y, x, y+size, 4, clr, true)
}
func drawLetterZ(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x, y, x+size/2, y, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y, x, y+size, 4, clr, true)
	vector.StrokeLine(s, x, y+size, x+size/2, y+size, 4, clr, true)
}
func DrawNeonText(screen *ebiten.Image, text string, x, y float64, size float64, clr color.RGBA) {
	text = strings.ToUpper(text)
	xf, yf := float32(x), float32(y)
	s := float32(size)
	spacing := s * 0.8
	offset := -float32(len(text)-1) * spacing / 2
	for i, char := range text {
		if img, ok := charTemplates[char]; ok {
			op := &ebiten.DrawImageOptions{}
			scale := float64(s / 24.0)
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(float64(xf+offset+float32(i)*spacing)-16*scale, float64(yf)-16*scale)
			op.ColorScale.ScaleWithColor(clr)
			screen.DrawImage(img, op)
		}
	}
}
func drawDigit0(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	drawLetterO(s, x, y, size, clr)
	vector.StrokeLine(s, x+size/2, y, x, y+size, 2, clr, true)
}
func drawDigit1(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x+size/4, y, x+size/4, y+size, 4, clr, true)
}
func drawDigit2(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x, y, x+size/2, y, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y, x+size/2, y+size/2, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y+size/2, x, y+size/2, 4, clr, true)
	vector.StrokeLine(s, x, y+size/2, x, y+size, 4, clr, true)
	vector.StrokeLine(s, x, y+size, x+size/2, y+size, 4, clr, true)
}
func drawDigit3(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x, y, x+size/2, y, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y, x+size/2, y+size, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y+size/2, x, y+size/2, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y+size, x, y+size, 4, clr, true)
}
func drawDigit4(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x, y, x, y+size/2, 4, clr, true)
	vector.StrokeLine(s, x, y+size/2, x+size/2, y+size/2, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y, x+size/2, y+size, 4, clr, true)
}
func drawDigit5(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x+size/2, y, x, y, 4, clr, true)
	vector.StrokeLine(s, x, y, x, y+size/2, 4, clr, true)
	vector.StrokeLine(s, x, y+size/2, x+size/2, y+size/2, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y+size/2, x+size/2, y+size, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y+size, x, y+size, 4, clr, true)
}
func drawDigit6(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	drawDigit5(s, x, y, size, clr)
	vector.StrokeLine(s, x, y+size/2, x, y+size, 4, clr, true)
}
func drawDigit7(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x, y, x+size/2, y, 4, clr, true)
	vector.StrokeLine(s, x+size/2, y, x+size/2, y+size, 4, clr, true)
}
func drawDigit8(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	drawLetterO(s, x, y, size, clr)
	vector.StrokeLine(s, x, y+size/2, x+size/2, y+size/2, 4, clr, true)
}
func drawDigit9(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	drawDigit4(s, x, y, size, clr)
	vector.StrokeLine(s, x, y, x+size/2, y, 4, clr, true)
}
func drawColon(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.DrawFilledRect(s, x+size/4-1, y+size/4, 2, 2, clr, true)
	vector.DrawFilledRect(s, x+size/4-1, y+3*size/4, 2, 2, clr, true)
}
func drawExclamation(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x+size/4, y, x+size/4, y+size*0.7, 2, clr, true)
	vector.DrawFilledRect(s, x+size/4-1, y+size*0.85, 2, 2, clr, true)
}
func drawQuestion(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x, y, x+size/2, y, 2, clr, true)
	vector.StrokeLine(s, x+size/2, y, x+size/2, y+size/2, 2, clr, true)
	vector.StrokeLine(s, x+size/2, y+size/2, x+size/4, y+size/2, 2, clr, true)
	vector.StrokeLine(s, x+size/4, y+size/2, x+size/4, y+size*0.7, 2, clr, true)
	vector.DrawFilledRect(s, x+size/4-1, y+size*0.85, 2, 2, clr, true)
}
func drawDot(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.DrawFilledRect(s, x+size/4-1, y+size-2, 2, 2, clr, true)
}
func drawPlus(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x, y+size/2, x+size/2, y+size/2, 2, clr, true)
	vector.StrokeLine(s, x+size/4, y+size/4, x+size/4, y+3*size/4, 2, clr, true)
}
func drawMinus(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x, y+size/2, x+size/2, y+size/2, 2, clr, true)
}
func drawPipe(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x+size/4, y, x+size/4, y+size, 2, clr, true)
}
func drawSlash(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x+size/2, y, x, y+size, 2, clr, true)
}
func drawPercent(s *ebiten.Image, x, y, size float32, clr color.RGBA) {
	vector.StrokeLine(s, x+size/2, y, x, y+size, 2, clr, true)
	vector.DrawFilledRect(s, x, y, 2, 2, clr, true)
	vector.DrawFilledRect(s, x+size/2-2, y+size-2, 2, 2, clr, true)
}
