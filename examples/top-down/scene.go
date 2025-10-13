package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebx/assetmgr"
	"github.com/samredway/ebx/camera"
	"github.com/samredway/ebx/engine"
	"github.com/samredway/ebx/geom"
	"image"
)

// ExampleScene is a template for how scene operates with its required hooks
type ExampleScene struct {
	Viewport     geom.Size
	Ids          engine.IdGen
	Camera       *camera.Camera
	PosStore     *engine.PositionStore
	RenderSys    *engine.RenderSystem
	MoveSys      *engine.MovementSystem
	UserInputSys *engine.UserInputSystem
	TileMap      *assetmgr.TileMap
}

// OnEnter is called on each scene load and should be used for setup like creating
// components and adding them to their relevant systems
func (es *ExampleScene) OnEnter() {
	es.Ids = engine.IdGen{}
	// Create a camera with a default worlsize of Viewport for now. When the tile map
	// is done can add a proper world bounds
	es.Camera = camera.NewCamera(
		es.Viewport,
		image.Rect(0, 0, es.Viewport.W, es.Viewport.H),
	)
	es.PosStore = engine.NewPositionStore()
	es.RenderSys = engine.NewRenderSystem(es.PosStore, es.Camera, es.TileMap)
	es.MoveSys = engine.NewMovementSystem(es.PosStore)
	es.UserInputSys = &engine.UserInputSystem{}

	// Create entities
	NewPlayer(es.Ids, es.RenderSys, es.PosStore, es.MoveSys, es.UserInputSys)
}

// OnExit is called when the scene is removed from current and allows exit transitions
// and clean up
func (es *ExampleScene) OnExit() {}

// Update us used primarily to run the relevant systems update methods
func (es *ExampleScene) Update(dt float64) engine.Scene {
	es.UserInputSys.Update(dt)
	es.MoveSys.Update(dt)
	es.RenderSys.Update(dt)
	return nil
}

// Draw will primarily run the scenes systems Draw methods
func (es *ExampleScene) Draw(screen *ebiten.Image) {
	es.RenderSys.Draw(screen)
}

// Set the view port size
func (es *ExampleScene) SetViewport(view geom.Size) {
	es.Viewport = view
}
