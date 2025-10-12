package engine

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebx/camera"
	"github.com/samredway/ebx/geom"
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
	SetViewport(geom.Size)
}

// SceneBase is a template for how scene operates with its required hooks this can
// be embedded into new scenes and its methods overriden as desired
type SceneBase struct {
	Viewport     geom.Size
	Ids          IdGen
	Camera       *camera.Camera
	PosStore     *PositionStore
	RenderSys    *RenderSystem
	MoveSys      *MovementSystem
	UserInputSys *UserInputSystem
}

// OnEnter is called on each scene load and should be used for setup like creating
// components and adding them to their relevant systems
func (sb *SceneBase) OnEnter() {
	sb.Ids = IdGen{}
	// Create a camera with a default worlsize of Viewport for now. When the tile map
	// is done can add a proper world bounds
	sb.Camera = camera.NewCamera(
		sb.Viewport,
		image.Rect(0, 0, sb.Viewport.W, sb.Viewport.H),
	)
	sb.PosStore = NewPositionStore()
	sb.RenderSys = NewRenderSystem(sb.PosStore, sb.Camera)
	sb.MoveSys = NewMovementSystem(sb.PosStore)
	sb.UserInputSys = &UserInputSystem{}
}

// OnExit is called when the scene is removed from current and allows exit transitions
// and clean up
func (sb *SceneBase) OnExit() {}

// Update us used primarily to run the relevant systems update methods
func (sb *SceneBase) Update(dt float64) Scene {
	sb.UserInputSys.Update(dt)
	sb.MoveSys.Update(dt)
	sb.RenderSys.Update(dt)
	return nil
}

// Draw will primarily run the scenes systems Draw methods
func (sb *SceneBase) Draw(screen *ebiten.Image) {
	sb.RenderSys.Draw(screen)
}

// Set the view port size
func (sb *SceneBase) SetViewport(view geom.Size) {
	sb.Viewport = view
}

// Game object implements ebiten.Game interface
type Game struct {
	curr     Scene
	viewport geom.Size
}

// NewGame returns a Game object that can run in Ebiten.
// You can must pass in a Scene argument that is your opening scene along with
// an Assets object which contains all the assets your game requires
func NewGame(scene Scene, screenW, screenH int) *Game {
	viewport := geom.Size{W: screenW, H: screenH}
	scene.SetViewport(viewport)
	scene.OnEnter()
	return &Game{
		curr:     scene,
		viewport: viewport,
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
	return g.viewport.W, g.viewport.H
}
