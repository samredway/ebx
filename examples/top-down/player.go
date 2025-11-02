package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/samredway/ebx/assetmgr"
	"github.com/samredway/ebx/engine"
	"github.com/samredway/ebx/geom"
)

// pScript implements the engine.Script interface for player-specific behavior
type pScript struct{}

func (ps *pScript) Update(e *engine.Entity, dt float64) {
	direction := geom.Vec2I{}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		direction.Y -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		direction.Y += 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		direction.X -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		direction.X += 1
	}
	e.Movement.DesiredDir = direction
}

func NewPlayer(assets *assetmgr.Assets) *engine.Entity {
	pPos := &engine.PositionComponent{
		Vec2: geom.Vec2{X: 100, Y: 200},
		Size: geom.Size{W: 16, H: 16},
	}

	pMov := &engine.MovementComponent{Speed: 200}

	pIdle, err := assets.GetSpriteSheet("Character_Idle.png")
	if err != nil {
		panic(fmt.Errorf("Unable to load sprites %w", err))
	}

	pRen := &engine.RenderComponent{Img: pIdle[0]}

	player := &engine.Entity{
		Name:     "Player",
		Position: pPos,
		Movement: pMov,
		Render:   pRen,
		Script:   &pScript{},
	}

	return player
}
