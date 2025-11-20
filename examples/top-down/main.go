package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebx/engine"
	"github.com/samredway/ebx/geom"

	"github.com/samredway/ebx/examples/top-down/scenes"
)

const (
	screenW = 640
	screenH = 480
)

func main() {
	ebiten.SetWindowSize(screenW, screenH)
	ebiten.SetWindowTitle("Top down Example")
	scene := &scenes.ExampleScene{}
	err := ebiten.RunGame(engine.NewGame(scene, geom.Size{W: screenW, H: screenH}))
	if err != nil {
		log.Fatal(err)
	}
}
