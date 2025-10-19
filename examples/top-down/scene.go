package main

import (
	_ "embed"
	"fmt"
	"image"
	_ "image/png"

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
	stateStore   *engine.StateStore
	animLibrary  *engine.AnimationLibrary
	renderSys    *engine.RenderSystem
	animationSys *engine.AnimationSystem
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
	es.stateStore = engine.NewStateStore()
	es.animLibrary = engine.NewAnimationLibrary()

	// Create assets and load tilemap (tilemap will load its own tilesets automatically)
	es.assets = assetmgr.NewAssets()
	tileMap, err := assetmgr.NewTileMapFromTmx(
		gameassets.GameFS,
		"example.tmx",
		es.assets,
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to load tilemap: %v", err))
	}
	es.tileMap = tileMap

	// Load character spritesheets - first check dimensions
	err = es.assets.LoadSpriteSheetFromFS(gameassets.GameFS, "Character_Idle.png", 16, 16)
	if err != nil {
		panic(fmt.Sprintf("Failed to load Character_Idle.png: %v", err))
	}

	err = es.assets.LoadSpriteSheetFromFS(gameassets.GameFS, "Character_Walk.png", 16, 16)
	if err != nil {
		panic(fmt.Sprintf("Failed to load Character_Walk.png: %v", err))
	}

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
	es.camera.Zoom = 2.0

	// Setup animation definitions

	// Character_Idle.png: 4 rows x 4 columns
	// Row 0 = left, Row 1 = right, Row 2 = up, Row 3 = down
	idleSheet, _ := es.assets.GetSpriteSheet("Character_Idle.png")

	es.animLibrary.AddAnimation("idle_left", &engine.AnimationDef{
		Name:        "idle_left",
		SpriteSheet: idleSheet,
		FirstFrame:  0,
		LastFrame:   3,
		FrameTime:   0.15, // 150ms per frame
		Loop:        true,
	})
	es.animLibrary.AddAnimation("idle_right", &engine.AnimationDef{
		Name:        "idle_right",
		SpriteSheet: idleSheet,
		FirstFrame:  4,
		LastFrame:   7,
		FrameTime:   0.15,
		Loop:        true,
	})
	es.animLibrary.AddAnimation("idle_up", &engine.AnimationDef{
		Name:        "idle_up",
		SpriteSheet: idleSheet,
		FirstFrame:  8,
		LastFrame:   11,
		FrameTime:   0.15,
		Loop:        true,
	})
	es.animLibrary.AddAnimation("idle_down", &engine.AnimationDef{
		Name:        "idle_down",
		SpriteSheet: idleSheet,
		FirstFrame:  12,
		LastFrame:   15,
		FrameTime:   0.15,
		Loop:        true,
	})

	// Character_Idle.png: 4 rows x 4 columns
	// Row 0 = left, Row 1 = right, Row 2 = up, Row 3 = down
	walkSheet, _ := es.assets.GetSpriteSheet("Character_Walk.png")

	es.animLibrary.AddAnimation("walk_left", &engine.AnimationDef{
		Name:        "walk_left",
		SpriteSheet: walkSheet,
		FirstFrame:  0,
		LastFrame:   3,
		FrameTime:   0.15, // 150ms per frame
		Loop:        true,
	})
	es.animLibrary.AddAnimation("walk_right", &engine.AnimationDef{
		Name:        "walk_right",
		SpriteSheet: walkSheet,
		FirstFrame:  4,
		LastFrame:   7,
		FrameTime:   0.15,
		Loop:        true,
	})
	es.animLibrary.AddAnimation("walk_up", &engine.AnimationDef{
		Name:        "walk_up",
		SpriteSheet: walkSheet,
		FirstFrame:  8,
		LastFrame:   11,
		FrameTime:   0.15,
		Loop:        true,
	})
	es.animLibrary.AddAnimation("walk_down", &engine.AnimationDef{
		Name:        "walk_down",
		SpriteSheet: walkSheet,
		FirstFrame:  12,
		LastFrame:   15,
		FrameTime:   0.15,
		Loop:        true,
	})

	// Setup animation state machine for top-down game
	const (
		stateIdle engine.AnimationState = "idle"
		stateWalk engine.AnimationState = "walk"
	)
	
	animStateMachine := engine.NewAnimationStateMachine(stateIdle)
	// Simple transitions: idle <-> walk (no attack or death for now)
	animStateMachine.AddTransition(stateIdle, stateWalk, 
		func(s *engine.StateComponent) bool { return s.IsMoving }, 10)
	animStateMachine.AddTransition(stateWalk, stateIdle, 
		func(s *engine.StateComponent) bool { return !s.IsMoving }, 10)

	// Setup core systems
	es.animationSys = engine.NewAnimationSystem(es.stateStore, es.animLibrary, animStateMachine)
	es.renderSys = engine.NewRenderSystem(
		es.posStore,
		es.camera,
		es.tileMap,
		es.animationSys, // AnimationSystem implements AnimationProvider
	)
	es.moveSys = engine.NewMovementSystem(es.posStore, es.stateStore, es.tileMap, 1)
	es.userInputSys = &engine.UserInputSystem{}

	// Create entities
	NewPlayer(es.ids, es.renderSys, es.posStore, es.stateStore, es.animationSys, es.moveSys, es.userInputSys)
}

// OnExit is called when the scene is removed from current and allows exit transitions
// and clean up
func (es *ExampleScene) OnExit() {}

// Update us used primarily to run the relevant systems update methods
func (es *ExampleScene) Update(dt float64) engine.Scene {
	es.userInputSys.Update(dt)
	es.animationSys.Update(dt)
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
