package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebx/assets"
)

// Scene is a level or view like a menu screen for example that has its own
// behviour. If you return a Scene from Update the Game will load in the
// new scene.
type Scene interface {
	OnEnter()
	OnExit()
	Draw(*ebiten.Image)
	Update(float64) Scene
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
	fps := ebiten.ActualFPS()
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
