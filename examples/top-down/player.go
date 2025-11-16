package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/samredway/ebx/assetmgr"
	"github.com/samredway/ebx/engine"
	"github.com/samredway/ebx/geom"
)

// pScript implements the engine.Script interface for player-specific behavior
type pScript struct {
	time       float64
	animRate   float64
	curAnim    string
	curFrame   int
	animations map[string][]*ebiten.Image
}

func (ps *pScript) Update(e *engine.Entity, dt float64) {
	direction := geom.Vec2I{}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		direction.Y -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		direction.Y += 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		direction.X -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		direction.X += 1
	}

	// update desired movement
	e.Movement.DesiredDir = direction

	// update animation frame
	ps.updateAnimations(e.Movement, dt)

	// update Render.Img
	e.Render.Img = ps.animations[ps.curAnim][ps.curFrame]
}

func (ps *pScript) updateAnimations(m *engine.MovementComponent, dt float64) {
	var nextAnim string
	var dirString string

	switch {
	case m.FacingDir.Y < 0:
		dirString = "up"
	case m.FacingDir.Y > 0:
		dirString = "down"
	case m.FacingDir.X < 0:
		dirString = "left"
	case m.FacingDir.X > 0:
		dirString = "right"
	default:
		dirString = "down"
	}

	if m.IsMoving {
		nextAnim = "walk_" + dirString
	} else {
		nextAnim = "idle_" + dirString
	}

	if nextAnim != ps.curAnim {
		ps.curFrame = 0
		ps.curAnim = nextAnim
	} else {
		ps.time += dt
		if ps.time >= ps.animRate {
			ps.time = 0.0
			ps.curFrame++
		}
	}

	// Wrap the frame to 0
	ps.curFrame %= len(ps.animations[ps.curAnim])
}

func newPScript(assets *assetmgr.Assets) *pScript {
	a := map[string][]*ebiten.Image{}

	// Setup animations
	anims, err := assets.GetSpriteSheet("Player")
	if err != nil {
		panic("Error retrieving spritesheet 'Player'")
	}

	// The spritesheet is a 16x16 tile grid with 18 tiles per row.
	// Each animation row has 6 frames laid out like:
	//
	//   E S E E S E E S E E S E E S E E S E
	//     ^   ^   ^   ^   ^   ^
	//     6 sprites, each 3 tiles apart, starting at column 1
	//
	// Rows with sprites are at tile rows 1,4,7,10,13,16,19,22,...
	// After splitting into tiles without filtering, the frame indices
	// we care about in the flat `anims` slice are:
	//
	//   idle_down  : 19  + 3*k, k=0..5
	//   idle_up    : 73  + 3*k
	//   idle_right : 127 + 3*k
	//   idle_left  : 181 + 3*k
	//   walk_down  : 235 + 3*k
	//   walk_up    : 289 + 3*k
	//   walk_right : 343 + 3*k
	//   walk_left  : 397 + 3*k
	//
	row := func(start int) []*ebiten.Image {
		frames := make([]*ebiten.Image, 0, 6)
		for i := 0; i < 6; i++ {
			frames = append(frames, anims[start+i*3])
		}
		return frames
	}

	a["idle_down"] = row(19)
	a["idle_up"] = row(73)
	a["idle_right"] = row(127)
	a["idle_left"] = row(181)

	a["walk_down"] = row(235)
	a["walk_up"] = row(289)
	a["walk_right"] = row(343)
	a["walk_left"] = row(397)

	return &pScript{
		animRate:   0.15,
		curFrame:   0,
		curAnim:    "idle_down",
		animations: a,
	}
}

func NewPlayer(assets *assetmgr.Assets) *engine.Entity {
	pPos := &engine.PositionComponent{
		Vec2: geom.Vec2{X: 100, Y: 200},
		Size: geom.Size{W: 16, H: 16},
	}

	pMov := &engine.MovementComponent{Speed: 200}

	sprites, err := assets.GetSpriteSheet("Player")
	if err != nil {
		panic(fmt.Errorf("Unable to load sprites %w", err))
	}

	pRen := &engine.RenderComponent{Img: sprites[0]}

	player := &engine.Entity{
		Name:     "Player",
		Position: pPos,
		Movement: pMov,
		Render:   pRen,
		Script:   newPScript(assets),
	}

	return player
}
