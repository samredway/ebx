package main

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebx/assetmgr"
	"github.com/samredway/ebx/camera"
	"github.com/samredway/ebx/engine"
	gameassets "github.com/samredway/ebx/examples/top-down/assets"
)

// ExampleScene demonstrates using the topdown.BaseScene for rapid prototyping
type ExampleScene struct {
	engine.BaseScene
	assets    *assetmgr.Assets
	tilemap   *assetmgr.TileMap
	entities  *engine.EntityManager
	renderSys *engine.RenderSystem
	moveSys   *engine.MovementSystem
}

// OnEnter sets up the scene by initializing base systems and creating entities
func (es *ExampleScene) OnEnter() {
	// Load game assets and tilemap -------------------------------------------
	es.assets = assetmgr.NewAssets()
	es.assets.LoadTileSetFromFS(gameassets.GameFS, "Dungeon_floor", "DungeonFloors.png", 16, 16)
	var err error
	es.tilemap, err = assetmgr.NewTileMapFromTmx(gameassets.GameFS, "example.tmx", es.assets)
	if err != nil {
		panic(fmt.Errorf("Unable to load tilemap %w", err))
	}
	err = es.assets.LoadSpriteSheetFromFS(gameassets.GameFS, "Player", "Player_sprites.png", 16, 16)
	if err != nil {
		panic(fmt.Errorf("Unable to load player sprite sheet"))
	}

	// Create player enity -----------------------------------------------------
	player := NewPlayer(es.assets)

	// Create entity manager and add player
	es.entities = engine.NewEntityManager()
	es.entities.Add(player)

	// Init systems ------------------------------------------------------------
	mapWidth := es.tilemap.MapWidth * es.tilemap.TileWidth
	mapHeight := es.tilemap.MapHeight * es.tilemap.TileHeight
	bounds := image.Rect(0, 0, mapWidth, mapHeight)
	cam := camera.NewCamera(es.Viewport, bounds)
	cam.Zoom = 2.0
	es.renderSys = engine.NewRenderSystem(es.entities, cam, player, es.tilemap)
	es.moveSys = engine.NewMovementSystem(es.entities, es.tilemap, 1)
}

func (es *ExampleScene) Update(dt float64) (engine.Scene, error) {
	es.entities.Update(dt)
	es.moveSys.Update(dt)
	es.entities.RemoveDead()
	return nil, nil
}

func (es *ExampleScene) Draw(screen *ebiten.Image) {
	es.renderSys.Draw(screen)
}
