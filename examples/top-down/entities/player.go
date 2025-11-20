package entities

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/samredway/ebx/assetmgr"
	"github.com/samredway/ebx/engine"
	gameassets "github.com/samredway/ebx/examples/top-down/assets"
	"github.com/samredway/ebx/geom"
)

func NewPlayer(assets *assetmgr.Assets) *engine.Entity {
	// The sprite sheet is a collection of 16x16 sprites each on a 48x48 canvas
	// We want to render it as a 48x48 but set collision only to the inside 16x16
	err := assets.LoadSpriteSheetFromFS(
		gameassets.GameFS,
		"Player",
		"Player_sprites.png",
		48, 48,
	)
	if err != nil {
		panic(fmt.Errorf("Unable to load player sprite sheet"))
	}

	// Set position
	pPos := &engine.PositionComponent{
		Vec2: geom.Vec2{X: 100, Y: 200},
	}

	// Set collision component
	pCollision := &engine.CollisionComponent{
		Size:   geom.Size{W: 16, H: 16},
		Offset: geom.Vec2{X: 16, Y: 16},
	}

	// Set Movement
	pMov := &engine.MovementComponent{Speed: 150}

	sprites, err := assets.GetSpriteSheet("Player")
	if err != nil {
		panic(fmt.Errorf("Unable to load sprites %w", err))
	}

	pRen := &engine.RenderComponent{Img: sprites[0]}

	player := &engine.Entity{
		Name:      "Player",
		Position:  pPos,
		Movement:  pMov,
		Render:    pRen,
		Collision: pCollision,
		Script:    newPScript(assets),
	}

	return player
}

// pScript implements the engine.Script interface for player-specific behavior
type pScript struct {
	time       float64
	animRate   float64
	curAnim    string
	curFrame   int
	animations map[string][]*ebiten.Image
	attacking  bool
}

func (ps *pScript) Update(e *engine.Entity, dt float64) {
	if ebiten.IsKeyPressed(ebiten.KeySpace) || ps.attacking {
		// Set attacking to true and ignore any movement
		ps.attacking = true
		e.Movement.IsMoving = false
		e.Movement.DesiredDir = geom.Vec2I{X: 0, Y: 0}
		ps.updateAttackAnimations(e.Movement, dt)
	} else {
		e.Movement.DesiredDir = getDir()
		ps.updateWalkAnimations(e.Movement, dt)
	}

	// update Render.Img
	e.Render.Img = ps.animations[ps.curAnim][ps.curFrame]
}

func (ps *pScript) updateAttackAnimations(m *engine.MovementComponent, dt float64) {
	dirString := getDirString(m)

	// If it is the last frame automatically transition to idel
	var nextAnim string
	if ps.curFrame != 5 {
		nextAnim = "attack_" + dirString
	} else {
		ps.attacking = false
		nextAnim = "idle_" + dirString
	}

	tickAnimationFrame(ps, nextAnim, dt)
}

func (ps *pScript) updateWalkAnimations(m *engine.MovementComponent, dt float64) {
	dirString := getDirString(m)
	var nextAnim string

	if m.IsMoving {
		nextAnim = "walk_" + dirString
	} else {
		nextAnim = "idle_" + dirString
	}

	tickAnimationFrame(ps, nextAnim, dt)
}

func tickAnimationFrame(ps *pScript, nextAnim string, dt float64) {
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

func getDir() geom.Vec2I {
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
	return direction
}

func getDirString(m *engine.MovementComponent) string {
	switch {
	case m.FacingDir.Y < 0:
		return "up"
	case m.FacingDir.Y > 0:
		return "down"
	case m.FacingDir.X < 0:
		return "left"
	case m.FacingDir.X > 0:
		return "right"
	default:
		return "down"
	}
}

func newPScript(assets *assetmgr.Assets) *pScript {
	a := map[string][]*ebiten.Image{}

	// Setup animations
	anims, err := assets.GetSpriteSheet("Player")
	if err != nil {
		panic("Error retrieving spritesheet 'Player'")
	}

	a["idle_down"] = anims[0:6]
	a["idle_up"] = anims[6:12]
	a["idle_right"] = anims[12:18]
	a["idle_left"] = anims[18:24]

	a["walk_down"] = anims[24:30]
	a["walk_up"] = anims[30:36]
	a["walk_right"] = anims[36:42]
	a["walk_left"] = anims[42:48]

	a["attack_down"] = anims[96:102]
	a["attack_up"] = anims[102:108]
	a["attack_right"] = anims[108:114]
	a["attack_left"] = anims[114:120]

	return &pScript{
		animRate:   0.15,
		curFrame:   0,
		curAnim:    "idle_down",
		animations: a,
	}
}
