package topdown

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebx/engine"
	"github.com/samredway/ebx/geom"
)

// SetupAnimations adds multiple animations to the library from a map
// This is the most flexible way to setup animations - works for any entity type
//
// Parameters:
//   - library: The animation library to add animations to
//   - prefix: Name prefix for animations (e.g., "player", "goblin")
//   - animations: Map of animation names to frame slices (e.g., "idle_left" -> frames)
//   - frameTime: Time per frame in seconds
//   - loop: Whether animations should loop
//
// Example:
//   animations := map[string][]*ebiten.Image{
//       "idle_left":   idleSheet[0:4],
//       "idle_right":  idleSheet[4:8],
//       "walk_left":   walkSheet[0:4],
//       "attack_left": attackSheet[0:6],
//   }
//   topdown.SetupAnimations(library, "player", animations, 0.15, true)
//   // Creates: player_idle_left, player_idle_right, player_walk_left, player_attack_left
//
// Note: Any animation with nil or empty frames will be skipped.
// The entity will fall back to its RenderComponent.Img for missing animations.
func SetupAnimations(
	library *engine.AnimationLibrary,
	prefix string,
	animations map[string][]*ebiten.Image,
	frameTime float64,
	loop bool,
) {
	for animName, frames := range animations {
		if frames != nil && len(frames) > 0 {
			fullName := fmt.Sprintf("%s_%s", prefix, animName)
			library.AddAnimation(fullName, &engine.AnimationDef{
				Name:        fullName,
				SpriteSheet: frames,
				FirstFrame:  0,
				LastFrame:   len(frames) - 1,
				FrameTime:   frameTime,
				Loop:        loop,
			})
		}
	}
}

// SetupCharacterStateMachine creates a standard state machine for 4-directional movement
// Returns a configured state machine with idle/walk transitions
//
// Parameters:
//   - prefix: Name prefix for states (e.g., "player", "goblin")
//   - moveSys: Movement system to query for entity state
//
// Creates states:
//   - {prefix}_idle_left, {prefix}_idle_right, {prefix}_idle_up, {prefix}_idle_down
//   - {prefix}_walk_left, {prefix}_walk_right, {prefix}_walk_up, {prefix}_walk_down
func SetupCharacterStateMachine(
	prefix string,
	moveSys *engine.MovementSystem,
) *engine.AnimationStateMachine {
	// Define states
	stateIdleLeft := engine.AnimationState(fmt.Sprintf("%s_idle_left", prefix))
	stateIdleRight := engine.AnimationState(fmt.Sprintf("%s_idle_right", prefix))
	stateIdleUp := engine.AnimationState(fmt.Sprintf("%s_idle_up", prefix))
	stateIdleDown := engine.AnimationState(fmt.Sprintf("%s_idle_down", prefix))
	stateWalkLeft := engine.AnimationState(fmt.Sprintf("%s_walk_left", prefix))
	stateWalkRight := engine.AnimationState(fmt.Sprintf("%s_walk_right", prefix))
	stateWalkUp := engine.AnimationState(fmt.Sprintf("%s_walk_up", prefix))
	stateWalkDown := engine.AnimationState(fmt.Sprintf("%s_walk_down", prefix))

	// Create state machine starting at idle_down
	sm := engine.NewAnimationStateMachine(stateIdleDown)

	// Helper to check direction (pure directions only, not diagonals)
	isMovingLeft := func(id engine.EntityId) bool {
		m := moveSys.GetMovement(id)
		return m != nil && m.IsMoving && m.FacingDir.X < 0 && m.FacingDir.Y == 0
	}
	isMovingRight := func(id engine.EntityId) bool {
		m := moveSys.GetMovement(id)
		return m != nil && m.IsMoving && m.FacingDir.X > 0 && m.FacingDir.Y == 0
	}
	isMovingUp := func(id engine.EntityId) bool {
		m := moveSys.GetMovement(id)
		return m != nil && m.IsMoving && m.FacingDir.Y < 0 && m.FacingDir.X == 0
	}
	isMovingDown := func(id engine.EntityId) bool {
		m := moveSys.GetMovement(id)
		return m != nil && m.IsMoving && m.FacingDir.Y > 0 && m.FacingDir.X == 0
	}

	// Diagonal movement helpers (for when both axes are active)
	isMovingUpLeft := func(id engine.EntityId) bool {
		m := moveSys.GetMovement(id)
		return m != nil && m.IsMoving && m.FacingDir.X < 0 && m.FacingDir.Y < 0
	}
	isMovingUpRight := func(id engine.EntityId) bool {
		m := moveSys.GetMovement(id)
		return m != nil && m.IsMoving && m.FacingDir.X > 0 && m.FacingDir.Y < 0
	}
	isMovingDownLeft := func(id engine.EntityId) bool {
		m := moveSys.GetMovement(id)
		return m != nil && m.IsMoving && m.FacingDir.X < 0 && m.FacingDir.Y > 0
	}
	isMovingDownRight := func(id engine.EntityId) bool {
		m := moveSys.GetMovement(id)
		return m != nil && m.IsMoving && m.FacingDir.X > 0 && m.FacingDir.Y > 0
	}

	// Helper to check if stopped in a direction (any component in that direction)
	isStoppedLeft := func(id engine.EntityId) bool {
		m := moveSys.GetMovement(id)
		return m != nil && !m.IsMoving && m.FacingDir.X < 0
	}
	isStoppedRight := func(id engine.EntityId) bool {
		m := moveSys.GetMovement(id)
		return m != nil && !m.IsMoving && m.FacingDir.X > 0
	}
	isStoppedUp := func(id engine.EntityId) bool {
		m := moveSys.GetMovement(id)
		return m != nil && !m.IsMoving && m.FacingDir.Y < 0
	}
	isStoppedDown := func(id engine.EntityId) bool {
		m := moveSys.GetMovement(id)
		return m != nil && !m.IsMoving && m.FacingDir.Y > 0
	}

	// From ANY idle state, can transition to ANY walk state based on direction
	idleStates := []engine.AnimationState{stateIdleLeft, stateIdleRight, stateIdleUp, stateIdleDown}
	for _, idleState := range idleStates {
		sm.AddTransition(idleState, stateWalkLeft, isMovingLeft, 10)
		sm.AddTransition(idleState, stateWalkRight, isMovingRight, 10)
		sm.AddTransition(idleState, stateWalkUp, isMovingUp, 10)
		sm.AddTransition(idleState, stateWalkDown, isMovingDown, 10)
	}

	// From ANY walk state, can transition to ANY idle state based on direction
	walkStates := []engine.AnimationState{stateWalkLeft, stateWalkRight, stateWalkUp, stateWalkDown}
	for _, walkState := range walkStates {
		sm.AddTransition(walkState, stateIdleLeft, isStoppedLeft, 10)
		sm.AddTransition(walkState, stateIdleRight, isStoppedRight, 10)
		sm.AddTransition(walkState, stateIdleUp, isStoppedUp, 10)
		sm.AddTransition(walkState, stateIdleDown, isStoppedDown, 10)
	}

	// Walk-to-walk transitions (for direction changes while moving)
	// Check pure directions first (higher priority), then diagonals
	for _, walkState := range walkStates {
		// Pure cardinal directions (highest priority)
		sm.AddTransition(walkState, stateWalkUp, isMovingUp, 10)
		sm.AddTransition(walkState, stateWalkDown, isMovingDown, 10)
		sm.AddTransition(walkState, stateWalkLeft, isMovingLeft, 10)
		sm.AddTransition(walkState, stateWalkRight, isMovingRight, 10)

		// Diagonals - prioritize Y-axis (up/down) over X-axis (left/right)
		sm.AddTransition(walkState, stateWalkUp, isMovingUpLeft, 8)
		sm.AddTransition(walkState, stateWalkUp, isMovingUpRight, 8)
		sm.AddTransition(walkState, stateWalkDown, isMovingDownLeft, 8)
		sm.AddTransition(walkState, stateWalkDown, isMovingDownRight, 8)
	}

	return sm
}

// CreateMovingCharacter is a convenience function that creates a character entity with
// movement, animation, and rendering components all set up
//
// Parameters:
//   - idGen: Entity ID generator
//   - prefix: Name prefix for animations (e.g., "player", "goblin")
//   - pos: Starting position
//   - size: Entity size
//   - speed: Movement speed
//   - defaultImg: Default/fallback image (can be nil for transparent fallback)
//   - animSys: Animation system
//   - moveSys: Movement system
//   - renderSys: Render system
//   - posSys: Position store
//
// Returns the entity ID
//
// Note: The defaultImg is used as a fallback when:
//   - An animation doesn't exist in the library (returns nil)
//   - An animation is not playing
//   - During prototyping before animations are ready
// If defaultImg is nil, a transparent image of the specified size is created.
func CreateMovingCharacter(
	idGen engine.IdGen,
	prefix string,
	pos geom.Vec2,
	size geom.Size,
	speed float64,
	defaultImg *ebiten.Image,
	animSys *engine.AnimationSystem,
	moveSys *engine.MovementSystem,
	renderSys *engine.RenderSystem,
	posSys *engine.PositionStore,
) engine.EntityId {
	id := idGen.Next()

	// Position
	pPos := &engine.PositionComponent{
		ComponentBase: engine.ComponentBase{EntityId: id},
		Vec2:          pos,
		Size:          size,
	}
	posSys.Attach(pPos)

	// Movement
	pMov := &engine.MovementComponent{
		ComponentBase: engine.ComponentBase{EntityId: id},
		Speed:         speed,
		DesiredDir:    geom.Vec2I{X: 0, Y: 0},
		FacingDir:     geom.Vec2I{X: 0, Y: 1}, // Default facing down
		IsMoving:      false,
	}
	moveSys.Attach(pMov)

	// Animation
	pAnim := &engine.AnimationComponent{
		ComponentBase: engine.ComponentBase{EntityId: id},
		CurrentAnim:   fmt.Sprintf("%s_idle_down", prefix),
		CurrentFrame:  0, // Start at first frame
		ElapsedTime:   0,
		Playing:       true,
	}
	animSys.Attach(pAnim)

	// Render (with default/fallback image)
	if defaultImg == nil {
		// Create transparent fallback if none provided
		defaultImg = ebiten.NewImage(int(size.W), int(size.H))
	}
	pRc := &engine.RenderComponent{
		ComponentBase: engine.ComponentBase{EntityId: id},
		Img:           defaultImg,
	}
	renderSys.Attach(pRc)

	return id
}
