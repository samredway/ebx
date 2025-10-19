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
	geom.Vec2 // X, Y
	geom.Size // W, H
}

// MovementComponent holds entity's velocity
type MovementComponent struct {
	ComponentBase
	Speed float64
}

// RenderComponent holds current image
type RenderComponent struct {
	ComponentBase
	Img *ebiten.Image
}

// StateComponent holds entity state that drives animation and behavior
type StateComponent struct {
	ComponentBase
	IsMoving    bool
	IsAttacking bool
	IsDead      bool
	FacingDir   geom.Vec2I // Unit vector direction entity is facing (-1, 0, 1)
}

// NewDefaultStateComponent creates a StateComponent with default values (facing down, not moving)
func NewDefaultStateComponent(id EntityId) *StateComponent {
	return &StateComponent{
		ComponentBase: ComponentBase{EntityId: id},
		IsMoving:      false,
		IsAttacking:   false,
		IsDead:        false,
		FacingDir:     geom.Vec2I{X: 0, Y: 1}, // Default facing down
	}
}

// AnimationComponent holds per-entity animation runtime state
type AnimationComponent struct {
	ComponentBase
	CurrentAnim  string  // Name of currently playing animation (e.g., "walk", "idle")
	CurrentFrame int     // Current frame index in the animation
	ElapsedTime  float64 // Time elapsed in current frame
	Playing      bool    // Whether animation is currently playing
}
