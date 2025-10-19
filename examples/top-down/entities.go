package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebx/engine"
	"github.com/samredway/ebx/geom"
)

// NOTE: in ebx an entity is just a collection of components and an id so
// this func creates a new id and writes the components to their relveant
// systems

// NewPlayer generates a new Player entity. Player is usually going to be
// unique in that the state component is handled by the input system so
// normally this function should only be called once in a game
func NewPlayer(
	idGen engine.IdGen,
	render *engine.RenderSystem,
	pos *engine.PositionStore,
	state *engine.StateStore,
	anim *engine.AnimationSystem,
	mov *engine.MovementSystem,
	inp *engine.UserInputSystem,
) {
	// Setup player entity
	pId := idGen.Next()

	// Player render component with fallback blue square
	fallbackImg := ebiten.NewImage(16, 16)
	fallbackImg.Fill(color.RGBA{80, 200, 120, 255})
	pRc := &engine.RenderComponent{
		ComponentBase: engine.ComponentBase{EntityId: pId},
		Img:           fallbackImg, // Fallback if animation system returns nil
	}
	render.Attach(pRc)

	// Player pos (using 16x16 for now until we get proper sized sprites)
	pPos := &engine.PositionComponent{
		ComponentBase: engine.ComponentBase{EntityId: pId},
		Vec2:          geom.Vec2{X: 100, Y: 100},
		Size:          geom.Size{W: 16, H: 16},
	}
	pos.Attach(pPos)

	// Attach to camera
	render.SetCamTarget(pId)

	// Player state
	pState := engine.NewDefaultStateComponent(pId)
	state.Attach(pState)

	// Player animation (CurrentFrame will be set by AnimationSystem based on FirstFrame)
	pAnim := &engine.AnimationComponent{
		ComponentBase: engine.ComponentBase{EntityId: pId},
		CurrentAnim:   "idle_down",
		CurrentFrame:  12, // Start at FirstFrame of idle_down animation
		ElapsedTime:   0,
		Playing:       true,
	}
	anim.Attach(pAnim)

	// Player movement
	pMov := &engine.MovementComponent{
		ComponentBase: engine.ComponentBase{EntityId: pId},
		Speed:         200.0,
	}
	mov.Attach(pMov)

	// Attach Player to the UserInputSystem
	inp.Attach(pState)
}
