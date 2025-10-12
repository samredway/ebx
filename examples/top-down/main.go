package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebx/assets"
	"github.com/samredway/ebx/engine"
)

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Top down Example")
	ebiten.SetTPS(60)
	scene := &ExampleScene{}
	assets := assets.NewAssets()
	if err := ebiten.RunGame(engine.NewGame(scene, assets, 640, 480)); err != nil {
		log.Fatal(err)
	}
}
