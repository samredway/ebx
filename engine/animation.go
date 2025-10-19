package engine

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebx/geom"
)

// AnimationDef defines a single animation sequence (shared across entities)
type AnimationDef struct {
	Name        string
	SpriteSheet []*ebiten.Image // The full spritesheet (all frames)
	FirstFrame  int             // Index of first frame in this animation
	LastFrame   int             // Index of last frame in this animation
	FrameTime   float64         // Time per frame in seconds
	Loop        bool            // Whether animation loops
}

// AnimationLibrary stores reusable animation definitions
type AnimationLibrary struct {
	animations map[string]*AnimationDef
}

func NewAnimationLibrary() *AnimationLibrary {
	return &AnimationLibrary{
		animations: map[string]*AnimationDef{},
	}
}

func (al *AnimationLibrary) AddAnimation(name string, def *AnimationDef) {
	al.animations[name] = def
}

func (al *AnimationLibrary) GetAnimation(name string) *AnimationDef {
	return al.animations[name]
}

// AnimationProvider is an interface for getting current animation frame for an entity
type AnimationProvider interface {
	GetCurrentImage(id EntityId) *ebiten.Image
}

// AnimationState represents a state in the animation state machine
type AnimationState string

const (
	AnimStateIdle   AnimationState = "idle"
	AnimStateWalk   AnimationState = "walk"
	AnimStateAttack AnimationState = "attack"
	AnimStateDead   AnimationState = "dead"
)

// AnimationStateMachine manages animation state transitions
type AnimationStateMachine struct {
	currentState AnimationState
	transitions  map[AnimationState][]AnimationTransition
}

// AnimationTransition defines a transition from one state to another
type AnimationTransition struct {
	to        AnimationState
	condition func(*StateComponent) bool
	priority  int // Higher priority transitions are checked first
}

func NewAnimationStateMachine(initialState AnimationState) *AnimationStateMachine {
	return &AnimationStateMachine{
		currentState: initialState,
		transitions:  map[AnimationState][]AnimationTransition{},
	}
}

func (sm *AnimationStateMachine) AddTransition(from, to AnimationState, condition func(*StateComponent) bool, priority int) {
	sm.transitions[from] = append(sm.transitions[from], AnimationTransition{
		to:        to,
		condition: condition,
		priority:  priority,
	})
}

// Update checks transitions and returns the new state and animation name
func (sm *AnimationStateMachine) Update(state *StateComponent) (AnimationState, string) {
	// Check all transitions from current state (sorted by priority)
	transitions := sm.transitions[sm.currentState]
	
	// Find highest priority transition that's valid
	var bestTransition *AnimationTransition
	for i := range transitions {
		t := &transitions[i]
		if t.condition(state) {
			if bestTransition == nil || t.priority > bestTransition.priority {
				bestTransition = t
			}
		}
	}
	
	// Transition if we found one
	if bestTransition != nil {
		sm.currentState = bestTransition.to
	}
	
	// Build animation name from state + direction
	animName := sm.buildAnimationName(sm.currentState, state.FacingDir)
	
	return sm.currentState, animName
}

func (sm *AnimationStateMachine) buildAnimationName(state AnimationState, dir geom.Vec2I) string {
	// Convert direction to string
	dirStr := directionToString(dir)
	
	// Dead animation typically has no direction
	if state == AnimStateDead {
		return "death"
	}
	
	// Other states use direction
	return fmt.Sprintf("%s_%s", state, dirStr)
}

func directionToString(dir geom.Vec2I) string {
	// Prioritize cardinal directions (X over Y)
	if dir.X < 0 {
		return "left"
	}
	if dir.X > 0 {
		return "right"
	}
	if dir.Y < 0 {
		return "up"
	}
	if dir.Y > 0 {
		return "down"
	}
	// Default to down if no direction
	return "down"
}

// AnimationSystem updates animation state based on entity state
// It acts as the arbiter for which animation should play based on state priority
//
// NOTE: Unlike other systems, AnimationSystem uses a map for both iteration and lookup.
// This trades slightly slower iteration (map vs slice) for simpler code - no need to
// maintain two data structures in sync. Given our pointer-based architecture, the
// performance difference is negligible, and we need O(1) lookup for GetCurrentImage anyway.
type AnimationSystem struct {
	components   map[EntityId]*AnimationComponent
	stateStore   *StateStore
	library      *AnimationLibrary
	stateMachine *AnimationStateMachine
}

func NewAnimationSystem(stateStore *StateStore, library *AnimationLibrary, stateMachine *AnimationStateMachine) *AnimationSystem {
	return &AnimationSystem{
		components:   map[EntityId]*AnimationComponent{},
		stateStore:   stateStore,
		library:      library,
		stateMachine: stateMachine,
	}
}

func (as *AnimationSystem) Attach(comps ...*AnimationComponent) {
	for _, comp := range comps {
		as.components[comp.EntityId] = comp
	}
}

func (as *AnimationSystem) Detach(ids ...EntityId) {
	for _, id := range ids {
		delete(as.components, id)
	}
}

func (as *AnimationSystem) Update(dt float64) {
	for _, anim := range as.components {
		state := as.stateStore.GetState(anim.EntityId)

		// Use state machine to determine animation
		_, desiredAnim := as.stateMachine.Update(state)

		// Switch animation if needed (check this FIRST, before getting current animDef)
		if desiredAnim != anim.CurrentAnim {
			anim.CurrentAnim = desiredAnim
			animDef := as.library.GetAnimation(anim.CurrentAnim)
			if animDef == nil {
				// Animation doesn't exist, stop playing so fallback image is used
				anim.Playing = false
				continue
			}
			anim.CurrentFrame = animDef.FirstFrame
			anim.ElapsedTime = 0
			anim.Playing = true
		}

		// Get current animation definition for frame advancement
		animDef := as.library.GetAnimation(anim.CurrentAnim)
		if animDef == nil || !anim.Playing {
			// Animation doesn't exist or not playing, skip frame advancement
			continue
		}

		// Advance time
		anim.ElapsedTime += dt

		// Check if we need to advance frame
		if anim.ElapsedTime >= animDef.FrameTime {
			anim.ElapsedTime -= animDef.FrameTime
			anim.CurrentFrame++

			// Handle end of animation
			if anim.CurrentFrame > animDef.LastFrame {
				if animDef.Loop {
					anim.CurrentFrame = animDef.FirstFrame
				} else {
					anim.CurrentFrame = animDef.LastFrame
					anim.Playing = false
				}
			}
		}
	}
}

// GetCurrentImage implements AnimationProvider interface
func (as *AnimationSystem) GetCurrentImage(id EntityId) *ebiten.Image {
	anim, ok := as.components[id]
	if !ok || !anim.Playing {
		return nil
	}

	animDef := as.library.GetAnimation(anim.CurrentAnim)
	if animDef == nil {
		return nil
	}

	return animDef.SpriteSheet[anim.CurrentFrame]
}
