package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebx/assets"
	"github.com/samredway/ebx/engine"
)

type ExampleScene struct {
	engine.SceneBase
}

func (es *ExampleScene) OnEnter() {
	es.SceneBase.OnEnter()
	playerId := es.Ids.Next()
	// Create a render component
	// Transform component
	// Input hanlder
	// Should now be able to see player and move around.
	// Perhpas move player into prefabs?
}

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
