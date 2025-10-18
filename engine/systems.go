package engine

import (
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebx/assetmgr"
	"github.com/samredway/ebx/camera"
	"github.com/samredway/ebx/collections"
	"github.com/samredway/ebx/geom"
)

// collisionEpsilon is a tiny offset to prevent floating-point precision issues
// when resolving collisions, avoiding player jitter against walls.
const collisionEpsilon = 0.001

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
	camTarget EntityId
}

func NewRenderSystem(
	pos *PositionStore,
	cam *camera.Camera,
	tileMap *assetmgr.TileMap,
) *RenderSystem {
	return &RenderSystem{
		SystemBase: NewSystemBase[*RenderComponent](),
		positions:  pos,
		camera:     cam,
		tileMap:    tileMap,
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

	tx0 := offsetX / rs.tileMap.TileW()
	tx1 := (offsetX+rs.camera.Viewport().W)/rs.tileMap.TileW() + 1
	ty0 := offsetY / rs.tileMap.TileH()
	ty1 := (offsetY+rs.camera.Viewport().H)/rs.tileMap.TileH() + 1

	viewRect := image.Rect(tx0, ty0, tx1, ty1)

	// Iterate layers and render
	for layer := range rs.tileMap.NumLayers() {
		rs.tileMap.ForEachIn(viewRect, layer, func(tx, ty, id int) {
			worldCoords := geom.Vec2{
				X: float64(tx * rs.tileMap.TileW()),
				Y: float64(ty * rs.tileMap.TileH()),
			}
			img := rs.tileMap.GetImageById(id)
			if img != nil {
				rs.drawToScreen(worldCoords, img, screen)
			}
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

	tw := float64(ms.tileMap.TileW())
	th := float64(ms.tileMap.TileH())

	for _, m := range ms.components {
		pos := ms.pos.GetPosition(m.GetEntityId())

		dir := geom.Normalize(m.Direction)
		dx := dir.X * m.Speed * dt
		dy := dir.Y * m.Speed * dt

		// move X, then Y (axis-separated → natural sliding)
		pos.X, pos.Y = ms.resolveXAxis(pos.X, pos.Y, float64(pos.W), float64(pos.H), dx, tw)
		pos.X, pos.Y = ms.resolveYAxis(pos.X, pos.Y, float64(pos.W), float64(pos.H), dy, th)
	}
}

// resolveXAxis moves along the X axis and clamps on collision.
// It uses "predict and correct" logic:
//  1. Calculate the new position (newX) after moving by dx
//  2. Check if that position would overlap any tiles
//  3. If yes, "push back" to the edge of the blocking tile
// Returns the resolved (x, y) position.
func (ms *MovementSystem) resolveXAxis(posX, posY, w, h, dx, tileW float64) (float64, float64) {
	// Try to move to the new X position
	newX := posX + dx
	
	if ms.tileMap.OverlapsTiles(newX, posY, w, h, ms.collisionLayer) {
		// We hit something! Need to push back to the edge of the blocking tile
		
		if dx > 0 {
			// Moving RIGHT - find the right edge of the entity and which tile column it's in
			rightEdge := newX + w
			blockingTileCol := math.Floor(rightEdge / tileW)
			// Push back: left edge of blocking tile minus our width minus safety gap
			newX = blockingTileCol*tileW - w - collisionEpsilon
			
		} else if dx < 0 {
			// Moving LEFT - find which tile column our left edge is in
			blockingTileCol := math.Floor(newX / tileW)
			// Push back: right edge of blocking tile plus safety gap
			newX = (blockingTileCol+1)*tileW + collisionEpsilon
		}
	}
	
	return newX, posY
}

// resolveYAxis moves along the Y axis and clamps on collision.
// It uses "predict and correct" logic:
//  1. Calculate the new position (newY) after moving by dy
//  2. Check if that position would overlap any tiles
//  3. If yes, "push back" to the edge of the blocking tile
// Returns the resolved (x, y) position.
func (ms *MovementSystem) resolveYAxis(posX, posY, w, h, dy, tileH float64) (float64, float64) {
	// Try to move to the new Y position
	newY := posY + dy
	
	if ms.tileMap.OverlapsTiles(posX, newY, w, h, ms.collisionLayer) {
		// We hit something! Need to push back to the edge of the blocking tile
		
		if dy > 0 {
			// Moving DOWN - find the bottom edge of the entity and which tile row it's in
			bottomEdge := newY + h
			blockingTileRow := math.Floor(bottomEdge / tileH)
			// Push back: top edge of blocking tile minus our height minus safety gap
			newY = blockingTileRow*tileH - h - collisionEpsilon
			
		} else if dy < 0 {
			// Moving UP - find which tile row our top edge is in
			blockingTileRow := math.Floor(newY / tileH)
			// Push back: bottom edge of blocking tile plus safety gap
			newY = (blockingTileRow+1)*tileH + collisionEpsilon
		}
	}
	
	return posX, newY
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
