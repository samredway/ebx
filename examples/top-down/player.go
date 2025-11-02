package main

import (
	"fmt"
	"strings"

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

func NewPScript(assets *assetmgr.Assets) *pScript {
	a := map[string][]*ebiten.Image{}

	// Setup animations
	for _, name := range []string{"Idle", "Walk"} {
		anims, err := assets.GetSpriteSheet(fmt.Sprintf("Character_%s.png", name))
		if err != nil {
			panic(fmt.Errorf("Error getting sprite sheet: %w", err))
		}

		nameLower := strings.ToLower(name)

		a[fmt.Sprintf("%s_left", nameLower)] = anims[0:4]
		a[fmt.Sprintf("%s_right", nameLower)] = anims[4:8]
		a[fmt.Sprintf("%s_up", nameLower)] = anims[8:12]
		a[fmt.Sprintf("%s_down", nameLower)] = anims[12:16]
	}

	return &pScript{
		animRate:   0.15,
		curFrame:   0,
		curAnim:    "idle_down",
		animations: a,
	}
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

func NewPlayer(assets *assetmgr.Assets) *engine.Entity {
	pPos := &engine.PositionComponent{
		Vec2: geom.Vec2{X: 100, Y: 200},
		Size: geom.Size{W: 16, H: 16},
	}

	pMov := &engine.MovementComponent{Speed: 200}

	pIdle, err := assets.GetSpriteSheet("Character_Idle.png")
	if err != nil {
		panic(fmt.Errorf("Unable to load sprites %w", err))
	}

	pRen := &engine.RenderComponent{Img: pIdle[0]}

	player := &engine.Entity{
		Name:     "Player",
		Position: pPos,
		Movement: pMov,
		Render:   pRen,
		Script:   NewPScript(assets),
	}

	return player
}
