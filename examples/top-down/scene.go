package main

import (
	"fmt"

	gameassets "github.com/samredway/ebx/examples/top-down/assets"
	"github.com/samredway/ebx/topdown"
)

// ExampleScene demonstrates using the topdown.BaseScene for rapid prototyping
type ExampleScene struct {
	topdown.BaseScene
}

// OnEnter sets up the scene by initializing base systems and creating entities
func (es *ExampleScene) OnEnter() {
	// Initialize base scene (tilemap, camera, core systems)
	// Note: Viewport is already set by engine before OnEnter is called
	err := es.BaseScene.Init(
		gameassets.GameFS, // embedded filesystem
		"example.tmx",     // tilemap file
		1,                 // collision layer index
		2.0,               // camera zoom
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize scene: %v", err))
	}

	// Load character spritesheets
	err = es.Assets.LoadSpriteSheetFromFS(gameassets.GameFS, "Character_Idle.png", 16, 16)
	if err != nil {
		panic(fmt.Sprintf("Failed to load Character_Idle.png: %v", err))
	}
	err = es.Assets.LoadSpriteSheetFromFS(gameassets.GameFS, "Character_Walk.png", 16, 16)
	if err != nil {
		panic(fmt.Sprintf("Failed to load Character_Walk.png: %v", err))
	}

	// Setup animation state machine and initialize animation system
	animStateMachine := topdown.SetupCharacterStateMachine("player", es.MoveSys)
	es.InitAnimationSystem(animStateMachine)

	// Create player entity (prefab handles animation setup)
	_ = NewPlayer(
		es.IdGen,
		es.Assets,
		es.AnimLibrary,
		es.RenderSys,
		es.PosStore,
		es.AnimationSys,
		es.MoveSys,
		es.UserInputSys,
	)
}

// Note: OnExit, Update, Draw, and SetViewport are all handled by topdown.BaseScene
// Only override them if you need custom behavior
