package main

import (
	_ "image/jpeg"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 1920
	screenHeight = 1080
)

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Neon Space Survivor")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)
	ebiten.SetFullscreen(true)
	InitAudio()
	LoadSplash()
	game := NewGame(true)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
