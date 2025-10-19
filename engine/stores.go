package engine

import "fmt"

// StateStore has no update but acts as a store for state components
type StateStore struct {
	states map[EntityId]*StateComponent
}

func NewStateStore() *StateStore {
	return &StateStore{
		states: map[EntityId]*StateComponent{},
	}
}

func (ss *StateStore) GetState(id EntityId) *StateComponent {
	state, ok := ss.states[id]
	if !ok {
		panic(fmt.Sprintf("StateComponent does not exist for entity %d", id))
	}
	return state
}

func (ss *StateStore) Attach(comps ...*StateComponent) {
	for _, comp := range comps {
		ss.states[comp.EntityId] = comp
	}
}

func (ss *StateStore) Detach(comps ...*StateComponent) {
	for _, comp := range comps {
		delete(ss.states, comp.EntityId)
	}
}

// PositionStore has no update but acts as a store for position component
// that gets updated by movement system and read from by some others like
// render and collision
type PositionStore struct {
	positions map[EntityId]*PositionComponent
}

func NewPositionStore() *PositionStore {
	return &PositionStore{
		positions: map[EntityId]*PositionComponent{},
	}
}

func (ps *PositionStore) GetPosition(id EntityId) *PositionComponent {
	pos, ok := ps.positions[id]
	if !ok {
		panic("Position id does not exist")
	}
	return pos
}

func (ps *PositionStore) Attach(comps ...*PositionComponent) {
	for _, comp := range comps {
		ps.positions[comp.EntityId] = comp
	}
}

func (ps *PositionStore) Detach(comps ...*PositionComponent) {
	for _, comp := range comps {
		delete(ps.positions, comp.EntityId)
	}
}
