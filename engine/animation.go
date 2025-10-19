package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
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

// AnimationLibrary stores reusable animation definitions for all entity types
// e.g., "player_idle_left", "goblin_walk_right", "boss_death"
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

// Update checks transitions and returns the current state
func (sm *AnimationStateMachine) Update(state *StateComponent) AnimationState {
	// Check all transitions from current state
	transitions := sm.transitions[sm.currentState]

	// Find highest priority valid transition
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

	return sm.currentState
}

// GetCurrentState returns the current state (useful for debugging or custom logic)
func (sm *AnimationStateMachine) GetCurrentState() AnimationState {
	return sm.currentState
}
