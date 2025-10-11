package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebx/geom"
)

// ComponentBase gives base functionality for all components
type ComponentBase struct {
	EntityId EntityId
}

func (cb ComponentBase) GetEntityId() EntityId {
	return cb.EntityId
}

// PositionComponent holds entity's position coords only
type PositionComponent struct {
	ComponentBase
	geom.Vec2
}

// MovementComponent holds entity's velocity
type MovementComponent struct {
	ComponentBase
	Speed     float64
	Direction geom.Vec2 // X, Y can be (-1 to 1)
}

// RenderComponent holds current image
type RenderComponenet struct {
	ComponentBase
	img *ebiten.Image
}
