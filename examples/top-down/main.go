package main

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebx/assets"
	"github.com/samredway/ebx/engine"
	"github.com/samredway/ebx/geom"
)

type ExampleScene struct {
	engine.SceneBase
}

func (es *ExampleScene) OnEnter() {
	es.SceneBase.OnEnter()

	// Setup player entity
	pId := es.Ids.Next()

	// Player Image
	pImg := ebiten.NewImage(30, 30)
	pImg.Fill(color.RGBA{80, 200, 120, 255})
	pRc := &engine.RenderComponent{
		ComponentBase: engine.ComponentBase{EntityId: pId},
		Img:           pImg,
	}
	es.RenderSys.Attach(pRc)

	// Player pos
	pPos := &engine.PositionComponent{
		ComponentBase: engine.ComponentBase{EntityId: pId},
		Vec2:          geom.Vec2{X: 100, Y: 100},
	}
	es.PosStore.Attach(pPos)

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
