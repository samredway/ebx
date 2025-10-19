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

// AnimationState represents a state in the animation state machine
// Users define their own states as strings (e.g., "idle", "walk", "jump", "attack")
type AnimationState string

// AnimationStateMachine manages animation state transitions
type AnimationStateMachine struct {
	currentState        AnimationState
	transitions         map[AnimationState][]AnimationTransition
	directionalStates   map[AnimationState]bool // States that use direction in animation name
}

// AnimationTransition defines a transition from one state to another
type AnimationTransition struct {
	to        AnimationState
	condition func(*StateComponent) bool
	priority  int // Higher priority transitions are checked first
}

func NewAnimationStateMachine(initialState AnimationState) *AnimationStateMachine {
	return &AnimationStateMachine{
		currentState:      initialState,
		transitions:       map[AnimationState][]AnimationTransition{},
		directionalStates: map[AnimationState]bool{},
	}
}

// SetDirectional marks a state as directional (animation name includes direction)
// By default, all states are directional. Call this with false for states like "death"
// that should not include direction in the animation name.
func (sm *AnimationStateMachine) SetDirectional(state AnimationState, directional bool) {
	sm.directionalStates[state] = directional
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
	// Check if this state uses direction (default is true if not specified)
	directional, exists := sm.directionalStates[state]
	if !exists {
		directional = true // Default: use direction
	}
	
	if !directional {
		// Non-directional animation (e.g., "death")
		return string(state)
	}
	
	// Directional animation: state + direction
	dirStr := DirectionToString(dir)
	return fmt.Sprintf("%s_%s", state, dirStr)
}

// DirectionToString converts a Vec2I direction to a string suffix for animation names
// This is exported so users can use it when building custom animation logic
func DirectionToString(dir geom.Vec2I) string {
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
