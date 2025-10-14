package main

import (
	_ "embed"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebx/assetmgr"
	"github.com/samredway/ebx/camera"
	"github.com/samredway/ebx/engine"
	gameassets "github.com/samredway/ebx/examples/top-down/assets"
	"github.com/samredway/ebx/geom"
	"image"
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
	// Create a camera with a default worldsize of Viewport for now. When the tile map
	// is done can add a proper world bounds
	es.camera = camera.NewCamera(
		es.viewport,
		image.Rect(0, 0, es.viewport.W, es.viewport.H),
	)
	es.posStore = engine.NewPositionStore()

	// Load assets
	es.assets = assetmgr.NewAssets()
	es.assets.LoadTileSetFromFS(gameassets.GameFS, "DungeonTiles", "DungeonTiles.png", 16)
	es.tileMap = assetmgr.NewTileMapFromTmx(gameassets.GameFS, "example.tmx", *es.assets)
	es.renderSys = engine.NewRenderSystem(es.posStore, es.camera, es.tileMap, es.assets.GetTileSet("DungeonTiles"))
	es.moveSys = engine.NewMovementSystem(es.posStore)
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
