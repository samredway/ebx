package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebx/engine"
)

const (
	screenW = 640
	screenH = 480
)

func main() {
	ebiten.SetWindowSize(screenW, screenH)
	ebiten.SetWindowTitle("Top down Example")
	ebiten.SetTPS(60)
	scene := &ExampleScene{}
	if err := ebiten.RunGame(engine.NewGame(scene, screenW, screenH)); err != nil {
		log.Fatal(err)
	}
}
