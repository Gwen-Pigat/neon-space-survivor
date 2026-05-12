package main

import (
	"fmt"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	shipImage    *ebiten.Image
	bulletImage  *ebiten.Image
	glowImage    *ebiten.Image
	splashImage  *ebiten.Image
	particleImg  *ebiten.Image
	alienImage   *ebiten.Image
	bossImage    *ebiten.Image
)

func init() {
	// These are now mostly fallbacks as we use vectors
	shipImage = ebiten.NewImage(32, 32)
	bulletImage = ebiten.NewImage(8, 8)
	particleImg = ebiten.NewImage(4, 4)
	particleImg.Fill(color.White)
	glowImage = createGlowTexture(64, color.RGBA{0, 255, 255, 100})
	alienImage = ebiten.NewImage(32, 32)
	bossImage = ebiten.NewImage(64, 64)
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
	vector.StrokePath(screen, &path, &vector.StrokeOptions{Width: 2}, dop)
	
	ebitenutil.DrawRect(screen, float64(xf-4), float64(yf-2), 2, 2, color.White)
	ebitenutil.DrawRect(screen, float64(xf+2), float64(yf-2), 2, 2, color.White)
}

func DrawNeonBoss(screen *ebiten.Image, x, y, size float64, clr color.RGBA, rotation float64) {
	s := float32(size)
	xf, yf := float32(x), float32(y)
	
	cos := float32(math.Cos(rotation))
	sin := float32(math.Sin(rotation))

	drawLineRot := func(x1, y1, x2, y2 float32) {
		rx1 := x1*cos - y1*sin + xf
		ry1 := x1*sin + y1*cos + yf
		rx2 := x2*cos - y2*sin + xf
		ry2 := x2*sin + y2*cos + yf
		vector.StrokeLine(screen, rx1, ry1, rx2, ry2, 4, clr, true)
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
	
	ebitenutil.DrawRect(screen, float64(xf-5), float64(yf-5), 10, 10, color.White)
}

func DrawNeonShip(screen *ebiten.Image, x, y, size float64, clr color.RGBA, rotation float64) {
	s := float32(size)
	xf, yf := float32(x), float32(y)
	cos := float32(math.Cos(rotation))
	sin := float32(math.Sin(rotation))

	drawVec := func(x1, y1, x2, y2 float32) {
		rx1 := x1*cos - y1*sin + xf
		ry1 := x1*sin + y1*cos + yf
		rx2 := x2*cos - y2*sin + xf
		ry2 := x2*sin + y2*cos + yf
		vector.StrokeLine(screen, rx1, ry1, rx2, ry2, 3, clr, true)
	}

	drawVec(0, -s/2, -s/3, s/2)
	drawVec(-s/3, s/2, 0, s/3)
	drawVec(0, s/3, s/3, s/2)
	drawVec(s/3, s/2, 0, -s/2)
	drawVec(-s/6, s/3, s/6, s/3)
	
	ebitenutil.DrawRect(screen, float64(xf-2), float64(yf-float32(size)/6), 4, 4, color.White)
}

func DrawNeonBullet(screen *ebiten.Image, x, y, size float64, clr color.RGBA) {
	xf, yf := float32(x), float32(y)
	s := float32(size)
	ebitenutil.DrawRect(screen, float64(xf-s/2), float64(yf-s/2), size, size, color.White)
	vector.StrokeCircle(screen, xf, yf, s, 2, clr, true)
}

func DrawNeonTitle(screen *ebiten.Image, x, y float64) {
	xf, yf := float32(x), float32(y)
	s := float32(80) // Larger letters
	spacing := float32(100)

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
	drawLetterI(screen, xf+0.2*spacing, yf+s*2.4, s, clr3)
	drawLetterV(screen, xf+0.8*spacing, yf+s*2.4, s, clr3)
	drawLetterO(screen, xf+1.8*spacing, yf+s*2.4, s, clr3)
	drawLetterR(screen, xf+2.8*spacing, yf+s*2.4, s, clr3)
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
	vector.StrokeLine(s, x, y, x, y+size, 4, clr, true)
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

func DrawNeonText(screen *ebiten.Image, text string, x, y float64, size float64, clr color.RGBA) {
	xf, yf := float32(x), float32(y)
	s := float32(size)
	spacing := s * 0.8
	offset := -float32(len(text)-1) * spacing / 2
	for i, char := range text {
		cx := xf + offset + float32(i)*spacing
		switch char {
		case 'N': drawLetterN(screen, cx, yf, s, clr)
		case 'E': drawLetterE(screen, cx, yf, s, clr)
		case 'O': drawLetterO(screen, cx, yf, s, clr)
		case 'S': drawLetterS(screen, cx, yf, s, clr)
		case 'P': drawLetterP(screen, cx, yf, s, clr)
		case 'A': drawLetterA(screen, cx, yf, s, clr)
		case 'C': drawLetterC(screen, cx, yf, s, clr)
		case 'U': drawLetterU(screen, cx, yf, s, clr)
		case 'R': drawLetterR(screen, cx, yf, s, clr)
		case 'V': drawLetterV(screen, cx, yf, s, clr)
		case 'I': drawLetterI(screen, cx, yf, s, clr)
		case 'T': drawLetterT(screen, cx, yf, s, clr)
		case 'L': drawLetterL(screen, cx, yf, s, clr)
		case 'Y': drawLetterY(screen, cx, yf, s, clr)
		}
	}
}

