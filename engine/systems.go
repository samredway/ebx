package engine

import "github.com/samredway/ebx/collections"

// PositionStore has no update but acts as a store for position component
// that gets updated by movement system and read from by some others like
// render and collision
type PositionStore struct {
	positions map[EntityId]*PositionComponent
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

// RenderSystem

// MovementSystem handles updating position component for corresponding entity
// based on movement data
type MovementSystem struct {
	ps        *PositionStore
	movements []MovementComponent
	remove    collections.Set[EntityId]
}

func (ms *MovementSystem) Attach(comps ...MovementComponent) {
	ms.movements = append(ms.movements, comps...)
}

func (ms *MovementSystem) Detach(ids ...EntityId) {
	for _, id := range ids {
		ms.remove.Add(id)
	}
}

func (ms *MovementSystem) Update(dt float64) {
	// Point a new slice handle to the underlaying movements array but with 0 length
	active := ms.movements[:0]

	for _, m := range ms.movements {
		// Remove and compact those on the remove list
		if ms.remove.Has(m.EntityId) {
			ms.remove.Remove(m.EntityId)
			continue
		}
		active = append(active, m)

		// Handle movement updates
		pos := ms.ps.GetPosition(m.EntityId)
		pos.X += m.Direction.X * m.Speed * dt
		pos.Y += m.Direction.Y * m.Speed * dt
	}

	// Remove any potential ids that do no longer exist
	ms.remove.Clear()

	// Reattach the slice handle to the underlying array so it re-indexes and gets
	// the correct new length
	ms.movements = active
}
