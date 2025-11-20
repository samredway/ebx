package entities

import (
	"fmt"

	"github.com/samredway/ebx/assetmgr"
	"github.com/samredway/ebx/engine"
	gameassets "github.com/samredway/ebx/examples/top-down/assets"
	"github.com/samredway/ebx/geom"
)

// NewSlime generates a new Slime entity at the given position
func NewSlime(assets *assetmgr.Assets, pos geom.Vec2) *engine.Entity {
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
	sPos := &engine.PositionComponent{Vec2: pos}

	// Set collision component
	sCollision := &engine.CollisionComponent{
		Size: geom.Size{W: 16, H: 16},
	}

	// Set Movement
	sMov := &engine.MovementComponent{Speed: 150}

	sprites, err := assets.GetSpriteSheet("Slime")
	if err != nil {
		panic(fmt.Errorf("Unable to load sprites %w", err))
	}

	sRend := &engine.RenderComponent{Img: sprites[0]}

	return &engine.Entity{
		Name:      "Player",
		Position:  sPos,
		Movement:  sMov,
		Render:    sRend,
		Collision: sCollision,
		Script:    newSlimeScript(assets),
	}
}

// slimeScript is the Script object for the Slime entity
type slimeScript struct{}

func (ss *slimeScript) Update(e *engine.Entity, dt float64) {}

func newSlimeScript(assets *assetmgr.Assets) *slimeScript {
	return &slimeScript{}
}
