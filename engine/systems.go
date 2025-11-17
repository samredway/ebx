package engine

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebx/assetmgr"
	"github.com/samredway/ebx/camera"
	"github.com/samredway/ebx/geom"
)

// collisionEpsilon is a tiny offset to prevent floating-point precision issues
// when resolving collisions, avoiding player jitter against walls.
const collisionEpsilon = 0.001

// RenderSystem gets run in the Scene.Draw() method
type RenderSystem struct {
	entities  *EntityManager
	camera    *camera.Camera
	tileMap   *assetmgr.TileMap
	camTarget *Entity // Entity for camera to center on (usaully Player)
}

// Draw draws entities and tiles to screen
func (rs *RenderSystem) Draw(screen *ebiten.Image) {
	if rs.camTarget.Position == nil && rs.camTarget == nil {
		panic("Camera target has not been set")
	}
	rs.camera.CentreOn(rs.camTarget.Position.Vec2)

	// Draw tiles first
	rs.drawTiles(screen)

	// Draw entities
	rs.entities.Each(func(e *Entity) {
		if e.Position == nil || e.Render == nil {
			return
		}
		if e.Render.Img == nil {
			panic(fmt.Errorf("Entity %s does not have image", e.Name))
		}
		rs.drawToScreen(e.Position.Vec2, e.Render.Img, screen)
	})
}

func (rs *RenderSystem) drawTiles(screen *ebiten.Image) {
	// Find the rectangle that the viewport covers as a rect on the tileMap
	// by coverting world cooridanates to tile coords
	offsetX := int(rs.camera.X)
	offsetY := int(rs.camera.Y)

	// Account for zoom when calculating visible area
	viewportWorldW := int(float64(rs.camera.Viewport().W) / rs.camera.Zoom)
	viewportWorldH := int(float64(rs.camera.Viewport().H) / rs.camera.Zoom)

	tx0 := offsetX / rs.tileMap.TileWidth
	tx1 := (offsetX+viewportWorldW)/rs.tileMap.TileWidth + 1
	ty0 := offsetY / rs.tileMap.TileHeight
	ty1 := (offsetY+viewportWorldH)/rs.tileMap.TileHeight + 1

	viewRect := image.Rect(tx0, ty0, tx1, ty1)

	// Iterate layers and render
	for layer := range rs.tileMap.NumLayers() {
		err := rs.tileMap.ForEachIn(viewRect, layer, func(tx, ty, id int) {
			worldCoords := geom.Vec2{
				X: float64(tx * rs.tileMap.TileWidth),
				Y: float64(ty * rs.tileMap.TileHeight),
			}
			img, err := rs.tileMap.GetImageById(id)
			if err != nil {
				panic(fmt.Sprintf("Failed to get tile image for ID %d at (%d, %d): %v", id, tx, ty, err))
			}
			if img != nil {
				rs.drawToScreen(worldCoords, img, screen)
			}
		})
		if err != nil {
			panic(fmt.Sprintf("Failed to iterate tiles in layer %d: %v", layer, err))
		}
	}
}

func (rs *RenderSystem) drawToScreen(
	worldCoords geom.Vec2,
	img *ebiten.Image,
	screen *ebiten.Image,
) {
	screenCoords := rs.camera.Apply(worldCoords)
	imgW := float64(img.Bounds().Dx()) * rs.camera.Zoom
	imgH := float64(img.Bounds().Dy()) * rs.camera.Zoom
	viewW := float64(rs.camera.Viewport().W)
	viewH := float64(rs.camera.Viewport().H)

	// Skip anything outside the visible screen
	if screenCoords.X < -imgW || screenCoords.X > viewW ||
		screenCoords.Y < -imgH || screenCoords.Y > viewH {
		return
	}

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(rs.camera.Zoom, rs.camera.Zoom)
	opts.GeoM.Translate(screenCoords.X, screenCoords.Y)
	screen.DrawImage(img, opts)
}

func NewRenderSystem(
	ents *EntityManager,
	cam *camera.Camera,
	camT *Entity,
	tiles *assetmgr.TileMap,
) *RenderSystem {
	return &RenderSystem{
		entities:  ents,
		camera:    cam,
		camTarget: camT,
		tileMap:   tiles,
	}
}

// MovementSystem handles updating position component for corresponding entity
// based on movement data.
// Checks whether a movement is possible by looking at tile map before moving
type MovementSystem struct {
	entities       *EntityManager
	tileMap        *assetmgr.TileMap
	collisionLayer int
}

func (ms *MovementSystem) Update(dt float64) {
	tw := float64(ms.tileMap.TileWidth)
	th := float64(ms.tileMap.TileHeight)

	ms.entities.Each(func(e *Entity) {
		m := e.Movement
		pos := e.Position

		if m == nil || pos == nil {
			return
		}

		// Check if there's any desired movement
		if m.DesiredDir.X == 0 && m.DesiredDir.Y == 0 {
			m.IsMoving = false
			return
		}

		// Normalize desired direction to prevent faster diagonal movement
		dir := geom.Vec2{X: float64(m.DesiredDir.X), Y: float64(m.DesiredDir.Y)}
		dir = geom.Normalize(dir)

		// Calculate velocity
		dx := dir.X * m.Speed * dt
		dy := dir.Y * m.Speed * dt

		// Store old position to detect actual movement
		oldX, oldY := pos.X, pos.Y

		// move X, then Y (axis-separated → natural sliding)
		// If no collision component, move freely without collision checks
		if e.Collision == nil {
			pos.X += dx
			pos.Y += dy
			m.IsMoving = true
			m.FacingDir = m.DesiredDir
			return
		}

		newX, newY := ms.resolveXAxis(pos.X, pos.Y, float64(e.Collision.Size.W), float64(e.Collision.Size.H), dx, tw, e.Collision.Offset)
		newX, newY = ms.resolveYAxis(newX, newY, float64(e.Collision.Size.W), float64(e.Collision.Size.H), dy, th, e.Collision.Offset)

		// Update position
		pos.X, pos.Y = newX, newY

		// Calculate actual movement to determine if entity is moving
		actualDX := newX - oldX
		actualDY := newY - oldY

		// Update IsMoving based on whether position actually changed
		m.IsMoving = (actualDX != 0 || actualDY != 0)

		// Update FacingDir to actual movement direction (or preserve if no movement)
		if m.IsMoving {
			// Convert actual movement to unit vector
			if actualDX > 0 {
				m.FacingDir.X = 1
			} else if actualDX < 0 {
				m.FacingDir.X = -1
			} else {
				m.FacingDir.X = 0
			}

			if actualDY > 0 {
				m.FacingDir.Y = 1
			} else if actualDY < 0 {
				m.FacingDir.Y = -1
			} else {
				m.FacingDir.Y = 0
			}
		}
	})
}

// resolveXAxis moves along the X axis and clamps on collision.
// It uses "predict and correct" logic:
//  1. Calculate the new position (newX) after moving by dx
//  2. Check if that position would overlap any tiles
//  3. If yes, "push back" to the edge of the blocking tile
//
// Returns the resolved (x, y) position.
func (ms *MovementSystem) resolveXAxis(posX, posY, w, h, dx, tileW float64, colOffset geom.Vec2) (float64, float64) {
	// Try to move to the new X position
	newX := posX + dx

	overlaps, err := ms.tileMap.OverlapsTiles(newX+colOffset.X, posY+colOffset.Y, w, h, ms.collisionLayer)
	if err != nil {
		panic(fmt.Sprintf("Failed to check tile collision: %v", err))
	}
	if overlaps {
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
//
// Returns the resolved (x, y) position.
func (ms *MovementSystem) resolveYAxis(posX, posY, w, h, dy, tileH float64, colOffset geom.Vec2) (float64, float64) {
	// Try to move to the new Y position
	newY := posY + dy

	overlaps, err := ms.tileMap.OverlapsTiles(posX+colOffset.X, newY+colOffset.Y, w, h, ms.collisionLayer)
	if err != nil {
		panic(fmt.Sprintf("Failed to check tile collision: %v", err))
	}
	if overlaps {
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

func NewMovementSystem(ents *EntityManager, tiles *assetmgr.TileMap, collLayer int) *MovementSystem {
	return &MovementSystem{
		entities:       ents,
		tileMap:        tiles,
		collisionLayer: collLayer,
	}
}
