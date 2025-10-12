package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebx/camera"
	"github.com/samredway/ebx/collections"
	"github.com/samredway/ebx/geom"
)

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

// SystemBase is a generic base for all system classes with the key methods for
// attach and detach definied and the slice of components required for interation
type SystemBase[C Component] struct {
	components []C
	remove     collections.Set[EntityId]
}

func NewSystemBase[C Component]() *SystemBase[C] {
	return &SystemBase[C]{
		components: []C{},
		remove:     collections.Set[EntityId]{},
	}
}

func (sb *SystemBase[C]) Attach(comp ...C) {
	sb.components = append(sb.components, comp...)
}

func (sb *SystemBase[C]) Detach(ids ...EntityId) {
	for _, id := range ids {
		sb.remove.Add(id)
	}
}

func (sb *SystemBase[C]) Update(dt float64) {
	// Point a new slice handle to the underlaying components array but with 0 length
	active := sb.components[:0]

	for _, c := range sb.components {
		// Remove and compact those on the remove list
		if sb.remove.Has(c.GetEntityId()) {
			sb.remove.Remove(c.GetEntityId())
			continue
		}
		active = append(active, c)
	}

	// Remove any potential ids that do no longer exist
	sb.remove.Clear()

	// Reattach the slice handle to the underlying array so it re-indexes and gets
	// the correct new length
	sb.components = active
}

// RenderSystem gets run in the Scene.Draw() method
type RenderSystem struct {
	*SystemBase[*RenderComponent]
	pos       *PositionStore
	cam       *camera.Camera
	camTarget EntityId
}

func NewRenderSystem(pos *PositionStore, cam *camera.Camera) *RenderSystem {
	return &RenderSystem{
		SystemBase: NewSystemBase[*RenderComponent](),
		pos:        pos,
		cam:        cam,
	}
}

func (rs *RenderSystem) Draw(screen *ebiten.Image) {
	pPos := rs.pos.GetPosition(rs.camTarget)
	rs.cam.CentreOn(pPos.Vec2)

	for _, r := range rs.components {
		pos := rs.pos.GetPosition(r.GetEntityId())
		opts := &ebiten.DrawImageOptions{}
		screenCoords := rs.cam.Apply(pos.Vec2)
		opts.GeoM.Translate(screenCoords.X, screenCoords.Y)
		screen.DrawImage(r.Img, opts)
	}
}

func (rs *RenderSystem) SetCamTarget(id EntityId) {
	rs.camTarget = id
}

// MovementSystem handles updating position component for corresponding entity
// based on movement data
type MovementSystem struct {
	*SystemBase[*MovementComponent]
	pos *PositionStore
}

func NewMovementSystem(pos *PositionStore) *MovementSystem {
	return &MovementSystem{
		SystemBase: NewSystemBase[*MovementComponent](),
		pos:        pos,
	}
}

func (ms *MovementSystem) Update(dt float64) {
	ms.SystemBase.Update(dt)

	for _, m := range ms.components {
		pos := ms.pos.GetPosition(m.GetEntityId())
		m.Direction = geom.Normalize(m.Direction)
		pos.X += m.Direction.X * m.Speed * dt
		pos.Y += m.Direction.Y * m.Speed * dt
	}
}

// InputSystem handles user input and applies it to a given movement component
// NOTE: InputSystem has a slightly different interface to other systems as it really
// .only handles one component although it could easily be updated to match the
// others later if required
// NOTE: This is probably not very extensible but its hard for me to think how to
// really generalise this right now. Will probalby come back to this and give it a
// bit more thought. We will want other types of system that can update movement and
// initialse other types of actions like shooting eg EnemyTypeXAiInputSystem. I think
// this is something I will iterate on and figure out as I go.
type UserInputSystem struct {
	PlayerMovement *MovementComponent
}

func (uis *UserInputSystem) Update(dt float64) {
	directionX := 0.0
	directionY := 0.0

	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		directionX -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		directionX += 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		directionY -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		directionY += 1
	}

	uis.PlayerMovement.Direction.X = directionX
	uis.PlayerMovement.Direction.Y = directionY
}

func (uis *UserInputSystem) Attach(mov *MovementComponent) {
	uis.PlayerMovement = mov
}
