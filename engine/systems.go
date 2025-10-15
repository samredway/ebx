package engine

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebx/assetmgr"
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
	positions *PositionStore
	camera    *camera.Camera
	tileMap   *assetmgr.TileMap
	tileSet   []*ebiten.Image
	camTarget EntityId
}

func NewRenderSystem(
	pos *PositionStore,
	cam *camera.Camera,
	tileMap *assetmgr.TileMap,
	tileSet []*ebiten.Image,
) *RenderSystem {
	return &RenderSystem{
		SystemBase: NewSystemBase[*RenderComponent](),
		positions:  pos,
		camera:     cam,
		tileMap:    tileMap,
		tileSet:    tileSet,
	}
}

func (rs *RenderSystem) Draw(screen *ebiten.Image) {
	pPos := rs.positions.GetPosition(rs.camTarget)
	rs.camera.CentreOn(pPos.Vec2)

	// Draw tiles first -----

	// Find the rectangle that the viewport covers as a rect on the tileMap
	// by coverting world cooridanates to tile coords
	offsetX := int(rs.camera.X)
	offsetY := int(rs.camera.Y)

	tx0 := offsetX / rs.tileMap.TileSize()
	tx1 := (offsetX+rs.camera.Viewport().W)/rs.tileMap.TileSize() + 1
	ty0 := offsetY / rs.tileMap.TileSize()
	ty1 := (offsetY+rs.camera.Viewport().H)/rs.tileMap.TileSize() + 1

	viewRect := image.Rect(tx0, ty0, tx1, ty1)

	// Iterate layers and render
	for layer := range rs.tileMap.NumLayers() {
		rs.tileMap.ForEachIn(viewRect, layer, func(tx, ty, id int) {
			worldCoords := geom.Vec2{
				X: float64(tx * rs.tileMap.TileSize()),
				Y: float64(ty * rs.tileMap.TileSize()),
			}
			img := rs.tileSet[id-1]
			rs.drawToScreen(worldCoords, img, screen)
		})
	}

	// Draw enitities -----
	for _, r := range rs.components {
		pos := rs.positions.GetPosition(r.GetEntityId())
		rs.drawToScreen(pos.Vec2, r.Img, screen)
	}
}

func (rs *RenderSystem) drawToScreen(
	worldCoords geom.Vec2,
	img *ebiten.Image,
	screen *ebiten.Image,
) {
	screenCoords := rs.camera.Apply(worldCoords)
	imgW := float64(img.Bounds().Dx())
	imgH := float64(img.Bounds().Dy())
	viewW := float64(rs.camera.Viewport().W)
	viewH := float64(rs.camera.Viewport().H)

	// Skip anything outside the visible screen
	if screenCoords.X < -imgW || screenCoords.X > viewW ||
		screenCoords.Y < -imgH || screenCoords.Y > viewH {
		return
	}

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(screenCoords.X, screenCoords.Y)
	screen.DrawImage(img, opts)
}

func (rs *RenderSystem) SetCamTarget(id EntityId) {
	rs.camTarget = id
}

// MovementSystem handles updating position component for corresponding entity
// based on movement data
type MovementSystem struct {
	*SystemBase[*MovementComponent]
	pos            *PositionStore
	tileMap        *assetmgr.TileMap
	collisionLayer int // TODO: crude to pass this here? thinking ...
}

func NewMovementSystem(pos *PositionStore, tileMap *assetmgr.TileMap, collLayer int) *MovementSystem {
	return &MovementSystem{
		SystemBase:     NewSystemBase[*MovementComponent](),
		pos:            pos,
		tileMap:        tileMap,
		collisionLayer: collLayer,
	}
}

func (ms *MovementSystem) Update(dt float64) {
	ms.SystemBase.Update(dt)
	for _, m := range ms.components {
		pos := ms.pos.GetPosition(m.GetEntityId())
		m.Direction = geom.Normalize(m.Direction)

		// calculate intent so we can check if it will remain in bounds
		intentToMove := geom.Vec2{
			X: pos.X + m.Direction.X*m.Speed*dt,
			Y: pos.Y + m.Direction.Y*m.Speed*dt,
		}
		intentTileCoords := ms.tileMap.WorldCoordsToTileCoords(intentToMove)

		// Check bounds and update player position
		if ms.tileMap.IsColliding(intentTileCoords, ms.collisionLayer) {
			return
		}
		pos.Vec2 = intentToMove
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
