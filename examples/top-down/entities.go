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
// unique in that the movement component is handled by the input system so
// normall this function should only be called once in a game
func NewPlayer(
	idGen engine.IdGen,
	rs *engine.RenderSystem,
	pos *engine.PositionStore,
	mov *engine.MovementSystem,
) {
	// Setup player entity
	pId := idGen.Next()

	// Player Image
	pImg := ebiten.NewImage(30, 30)
	pImg.Fill(color.RGBA{80, 200, 120, 255})
	pRc := &engine.RenderComponent{
		ComponentBase: engine.ComponentBase{EntityId: pId},
		Img:           pImg,
	}
	rs.Attach(pRc)

	// Player pos
	pPos := &engine.PositionComponent{
		ComponentBase: engine.ComponentBase{EntityId: pId},
		Vec2:          geom.Vec2{X: 100, Y: 100},
	}
	pos.Attach(pPos)

	// Player movement
	pMov := &engine.MovementComponent{
		ComponentBase: engine.ComponentBase{EntityId: pId},
		Speed:         200.0,
		Direction:     geom.Vec2{X: 1, Y: 0},
	}
	mov.Attach(pMov)
}
