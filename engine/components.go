package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebx/geom"
)

// Component is a data only object that can be processed by System's. It should hold
// only the data that is required for a given task. For example movement may only
// require the X, Y and Speed values of an entity.
type Component interface {
	GetEntityId() EntityId
}

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
type RenderComponent struct {
	ComponentBase
	Img *ebiten.Image
}
