package entities

import (
	"fmt"
	"math"

	"github.com/samredway/ebx/assetmgr"
	"github.com/samredway/ebx/engine"
	gameassets "github.com/samredway/ebx/examples/top-down/assets"
	"github.com/samredway/ebx/geom"
)

const SlimeSpeed float64 = 50

// NewSlime generates a new Slime entity at the given position
func NewSlime(
	assets *assetmgr.Assets,
	spawnPos geom.Vec2,
	name string,
	player *engine.Entity,
	tilemap *assetmgr.TileMap,
	collisionLayer int,
) *engine.Entity {
	err := assets.LoadSpriteSheetFromFS(
		gameassets.GameFS,
		"Slime",
		"Slimes.png",
		16, 16,
	)
	if err != nil {
		panic(fmt.Errorf("Unable to load Slimes sprite sheet"))
	}

	// Set position
	sPos := &engine.PositionComponent{Vec2: spawnPos}

	// Set collision component
	sCol := &engine.CollisionComponent{
		Size: geom.Size{W: 16, H: 16},
	}

	// Set Movement
	sMov := &engine.MovementComponent{Speed: SlimeSpeed}

	sprites, err := assets.GetSpriteSheet("Slime")
	if err != nil {
		panic(fmt.Errorf("Unable to load sprites %w", err))
	}

	sRend := &engine.RenderComponent{Img: sprites[0]}

	return &engine.Entity{
		Name:      name,
		Position:  sPos,
		Movement:  sMov,
		Render:    sRend,
		Collision: sCol,
		Script:    newSlimeScript(assets, player, tilemap, collisionLayer),
	}
}

// slimeScript is the Script object for the Slime entity
type slimeScript struct {
	player         *engine.Entity
	tilemap        *assetmgr.TileMap
	collisionLayer int
	sightRange     float64 // Maximum distance slime can see
}

func (ss *slimeScript) Update(e *engine.Entity, dt float64) {
	// Reset movment
	e.Movement.DesiredDir.X = 0
	e.Movement.DesiredDir.Y = 0

	// Calculate player's center position (accounting for collision offset and size)
	playerCenter := ss.getPlayerCollisionPos()

	// If player is in line of sight then move towards the player
	if ss.hasLineOfSight(e.Position.Vec2, playerCenter) {
		setDir(e, playerCenter)
	}
}

func setDir(e *engine.Entity, pPos geom.Vec2) {
	dx := pPos.X - e.Position.X
	dy := pPos.Y - e.Position.Y

	// Set desired direction
	if dx > 0 {
		e.Movement.DesiredDir.X = 1
	}
	if dx < 0 {
		e.Movement.DesiredDir.X = -1
	}

	if dy > 0 {
		e.Movement.DesiredDir.Y = 1
	}
	if dy < 0 {
		e.Movement.DesiredDir.Y = -1
	}
}

// hasLineOfSight checks if there's a clear line of sight from slime to player
// by casting a ray and checking for wall collisions along the path
func (ss *slimeScript) hasLineOfSight(slimePos, playerPos geom.Vec2) bool {
	// Calculate direction vector from slime to player
	dx := playerPos.X - slimePos.X
	dy := playerPos.Y - slimePos.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	// Check if player is outside sight range
	if distance > ss.sightRange {
		return false
	}

	// Normalize direction
	dirX := dx / distance
	dirY := dy / distance

	// Sample points along the ray
	// Use half tile size as step size for good accuracy
	stepSize := float64(ss.tilemap.TileWidth) / 2.0
	numSteps := int(distance / stepSize)

	// Check each point along the line for wall collision
	for i := 1; i <= numSteps; i++ {
		t := float64(i) * stepSize
		checkX := slimePos.X + dirX*t
		checkY := slimePos.Y + dirY*t

		// Check if this point overlaps with a wall tile
		// Using a small point check (2x2 pixels)
		overlaps, err := ss.tilemap.OverlapsTiles(checkX, checkY, 2, 2, ss.collisionLayer)
		if err != nil {
			return false
		}
		if overlaps {
			return false // Wall blocks line of sight
		}
	}

	return true // No walls blocking
}

// getPlayerCollisionPos calculates the center of the player's collision box
func (ss *slimeScript) getPlayerCollisionPos() geom.Vec2 {
	centerX := ss.player.Position.X + ss.player.Collision.Offset.X
	centerY := ss.player.Position.Y + ss.player.Collision.Offset.Y
	return geom.Vec2{X: centerX, Y: centerY}
}

func newSlimeScript(assets *assetmgr.Assets, player *engine.Entity, tilemap *assetmgr.TileMap, collisionLayer int) *slimeScript {
	return &slimeScript{
		player:         player,
		tilemap:        tilemap,
		collisionLayer: collisionLayer,
		sightRange:     200.0, // Slime can see up to 200 pixels away
	}
}
