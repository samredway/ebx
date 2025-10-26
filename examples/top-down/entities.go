package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebx/assetmgr"
	"github.com/samredway/ebx/engine"
	"github.com/samredway/ebx/geom"
	"github.com/samredway/ebx/topdown"
)

// NewPlayer generates a new Player entity prefab with animations.
// This is a complete prefab that sets up animations, components, and input handling.
func NewPlayer(
	idGen engine.IdGen,
	assets *assetmgr.Assets,
	animLibrary *engine.AnimationLibrary,
	render *engine.RenderSystem,
	pos *engine.PositionStore,
	anim *engine.AnimationSystem,
	mov *engine.MovementSystem,
	inp *engine.UserInputSystem,
) engine.EntityId {
	// Load player spritesheets
	// Character_Idle.png and Character_Walk.png are both 4x4 grids (160x192 pixels, 40x48 per frame)
	// Layout: Row 0 = left, Row 1 = right, Row 2 = up, Row 3 = down
	idleSheet, _ := assets.GetSpriteSheet("Character_Idle.png")
	walkSheet, _ := assets.GetSpriteSheet("Character_Walk.png")

	// Setup player animations
	playerAnims := map[string][]*ebiten.Image{
		"idle_left":  idleSheet[0:4],   // First row (frames 0-3)
		"idle_right": idleSheet[4:8],   // Second row (frames 4-7)
		"idle_up":    idleSheet[8:12],  // Third row (frames 8-11)
		"idle_down":  idleSheet[12:16], // Fourth row (frames 12-15)
		"walk_left":  walkSheet[0:4],
		"walk_right": walkSheet[4:8],
		"walk_up":    walkSheet[8:12],
		"walk_down":  walkSheet[12:16],
	}
	topdown.SetupAnimations(animLibrary, "player", playerAnims, 0.15, true)

	// Create default/fallback image (green square for visibility)
	defaultImg := ebiten.NewImage(16, 16)
	defaultImg.Fill(color.RGBA{80, 200, 120, 255})

	// Create player entity using topdown helper
	pId := topdown.CreateMovingCharacter(
		idGen,
		"player",
		geom.Vec2{X: 100, Y: 100}, // Starting position
		geom.Size{W: 16, H: 16},   // Size
		200.0,                      // Speed
		defaultImg,                 // Fallback image (used if animation missing)
		anim,
		mov,
		render,
		pos,
	)

	// Attach player to camera
	render.SetCamTarget(pId)

	// Attach player movement to input system (player-specific)
	pMov := mov.GetMovement(pId)
	inp.Attach(pMov)

	return pId
}
