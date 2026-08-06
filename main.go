package main

import (
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"runtime/debug"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 1920
	screenHeight = 1080
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			errStr := fmt.Sprintf("PANIC: %v\nStack:\n%s", r, debug.Stack())
			_ = os.WriteFile("game_error.log", []byte(errStr), 0644)
			log.Fatal(errStr)
		}
	}()

	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("Neon Space Survivor")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetFullscreen(true)
	InitAudio()
	LoadSplash()
	game := NewGame(true)
	if err := ebiten.RunGame(game); err != nil {
		_ = os.WriteFile("game_error.log", []byte(err.Error()), 0644)
		log.Fatal(err)
	}
}
