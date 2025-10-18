package main

import (
	_ "embed"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebx/assetmgr"
	"github.com/samredway/ebx/camera"
	"github.com/samredway/ebx/engine"
	gameassets "github.com/samredway/ebx/examples/top-down/assets"
	"github.com/samredway/ebx/geom"
)

// ExampleScene is a template for how scene operates with its required hooks
type ExampleScene struct {
	viewport     geom.Size
	ids          engine.IdGen
	camera       *camera.Camera
	posStore     *engine.PositionStore
	renderSys    *engine.RenderSystem
	moveSys      *engine.MovementSystem
	userInputSys *engine.UserInputSystem
	tileMap      *assetmgr.TileMap
	assets       *assetmgr.Assets
}

// OnEnter is called on each scene load and should be used for setup like creating
// components and adding them to their relevant systems
func (es *ExampleScene) OnEnter() {
	es.ids = engine.IdGen{}
	es.posStore = engine.NewPositionStore()

	// Create assets and load tilemap (tilemap will load its own tilesets automatically)
	es.assets = assetmgr.NewAssets()
	es.tileMap = assetmgr.NewTileMapFromTmx(
		gameassets.GameFS,
		"example.tmx",
		es.assets,
	)

	// Setup camera
	es.camera = camera.NewCamera(
		es.viewport,
		image.Rect(
			0,
			0,
			es.tileMap.MapSize().W*es.tileMap.TileW(),
			es.tileMap.MapSize().H*es.tileMap.TileH(),
		),
	)

	// Setup core systems
	es.renderSys = engine.NewRenderSystem(
		es.posStore,
		es.camera,
		es.tileMap,
	)
	es.moveSys = engine.NewMovementSystem(es.posStore, es.tileMap, 1)
	es.userInputSys = &engine.UserInputSystem{}

	// Create entities
	NewPlayer(es.ids, es.renderSys, es.posStore, es.moveSys, es.userInputSys)
}

// OnExit is called when the scene is removed from current and allows exit transitions
// and clean up
func (es *ExampleScene) OnExit() {}

// Update us used primarily to run the relevant systems update methods
func (es *ExampleScene) Update(dt float64) engine.Scene {
	es.userInputSys.Update(dt)
	es.moveSys.Update(dt)
	es.renderSys.Update(dt)
	return nil
}

// Draw will primarily run the scenes systems Draw methods
func (es *ExampleScene) Draw(screen *ebiten.Image) {
	es.renderSys.Draw(screen)
}

// Set the view port size
func (es *ExampleScene) SetViewport(view geom.Size) {
	es.viewport = view
}
