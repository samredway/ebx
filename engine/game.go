package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebx/geom"
)

// PositionComponent holds entity's position coords only
type PositionComponent struct {
	geom.Vec2 // X, Y
	geom.Size // W, H
}

// MovementComponent holds entity's movement state
type MovementComponent struct {
	Speed      float64
	DesiredDir geom.Vec2I // Direction intent (-1, 0, 1) - set by input system
	FacingDir  geom.Vec2I // Actual direction (-1, 0, 1) - set by movement system
	IsMoving   bool       // Whether entity moved this frame - set by movement system
}

// RenderComponent holds current image
type RenderComponent struct {
	Img *ebiten.Image
}

// Used to give entity specific custom behaviour to manage stuff like animations
// inputs/AI etc
type Script interface {
	Update(*Entity, float64)
}

// Entity game entity type
type Entity struct {
	Name     string
	Position *PositionComponent
	Movement *MovementComponent
	Render   *RenderComponent
	Script   Script
	Dead     bool
}

// EntityManager is a deliberately small abstraction to handle game entities
type EntityManager struct {
	entities []*Entity
}

// Add adds new entity
func (em *EntityManager) Add(e *Entity) {
	em.entities = append(em.entities, e)
}

// Each is a safe way for systems to run updates on the entity list
func (em *EntityManager) Each(fn func(*Entity)) {
	for _, e := range em.entities {
		fn(e)
	}
}

func (em *EntityManager) Update(dt float64) {
	em.Each(func(e *Entity) {
		if e.Script != nil {
			e.Script.Update(e, dt)
		}
	})
}

// RemoveDead removes all entityes marked Dead
func (em *EntityManager) RemoveDead() {
	alive := em.entities[:0]
	for _, e := range em.entities {
		if !e.Dead {
			alive = append(alive, e)
		}
	}
	em.entities = alive
}

func NewEntityManager() *EntityManager {
	return &EntityManager{entities: []*Entity{}}
}

// Scene is a level or view like a menu screen for example that has its own
// behviour. If you return a Scene from Update the Game will load in the
// new scene.
type Scene interface {
	OnEnter()
	OnExit()
	Draw(*ebiten.Image)
	Update(float64) (Scene, error)
	SetViewport(geom.Size)
}

// Game object implements ebiten.Game interface
type Game struct {
	curr     Scene
	viewport geom.Size
}

func (g *Game) Update() error {
	fps := float64(ebiten.TPS())
	dt := 1 / fps
	scene, err := g.curr.Update(dt)
	if scene != nil {
		g.curr.OnExit()
		g.curr = scene
		g.curr.OnEnter()
	}
	return err
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.curr.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.viewport.W, g.viewport.H
}

// NewGame returns a Game object that can run in Ebiten.
// You can must pass in a Scene argument that is your opening scene along with
// an Assets object which contains all the assets your game requires
func NewGame(scene Scene, viewport geom.Size) *Game {
	scene.SetViewport(viewport)
	scene.OnEnter()
	return &Game{
		curr:     scene,
		viewport: viewport,
	}
}
