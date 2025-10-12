package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebx/assets"
)

// Entity is just an ID that allows groups of components to be matched together
type EntityId int

type IdGen struct {
	last EntityId
}

func (ig *IdGen) Next() EntityId {
	ig.last++
	return ig.last
}

// Scene is a level or view like a menu screen for example that has its own
// behviour. If you return a Scene from Update the Game will load in the
// new scene.
type Scene interface {
	OnEnter()
	OnExit()
	Draw(*ebiten.Image)
	Update(float64) Scene
}

// SceneBase is a template for how scene operates with its required hooks this can
// be embedded into new scenes and its methods overriden as desired
type SceneBase struct {
	Ids       IdGen
	PosStore  *PositionStore
	RenderSys *RenderSystem
}

// OnEnter is called on each scene load and should be used for setup like creating
// components and adding them to their relevant systems
func (sb *SceneBase) OnEnter() {
	sb.Ids = IdGen{}
	sb.PosStore = NewPositionStore()
	sb.RenderSys = NewRenderSystem(sb.PosStore)
}

// OnExit is called when the scene is removed from current and allows exit transitions
// and clean up
func (sb *SceneBase) OnExit() {}

// Update us used primarily to run the relevant systems update methods
func (sb *SceneBase) Update(dt float64) Scene {
	return nil
}

// Draw will primarily run the scenes systems Draw methods
func (sb *SceneBase) Draw(screen *ebiten.Image) {
	sb.RenderSys.Draw(screen)
}

// Game object implements ebiten.Game interface
type Game struct {
	curr    Scene
	assets  *assets.Assets
	screenW int
	screenH int
}

// NewGame returns a Game object that can run in Ebiten.
// You can must pass in a Scene argument that is your opening scene along with
// an Assets object which contains all the assets your game requires
func NewGame(scene Scene, assets *assets.Assets, screenW, screenH int) *Game {
	scene.OnEnter()
	return &Game{
		curr:    scene,
		assets:  assets,
		screenW: screenW,
		screenH: screenH,
	}
}

func (g *Game) Update() error {
	fps := float64(ebiten.TPS())
	dt := 1 / fps
	scene := g.curr.Update(dt)
	if scene != nil {
		g.curr.OnExit()
		g.curr = scene
		g.curr.OnEnter()
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.curr.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.screenW, g.screenH
}
