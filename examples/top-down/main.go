package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebx/engine"
	"github.com/samredway/ebx/geom"
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
	if err := ebiten.RunGame(engine.NewGame(scene, geom.Size{W: screenW, H: screenH})); err != nil {
		log.Fatal(err)
	}
}
