package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebx/geom"
)

// BaseScene provides default implementations of the Scene interface
// Embed this in your scene to avoid implementing empty methods
//
// Example:
//   type MyScene struct {
//       engine.BaseScene
//   }
//
//   func (s *MyScene) OnEnter() {
//       // Your setup code
//   }
//
//   func (s *MyScene) Update(dt float64) Scene {
//       // Your update code
//       return nil
//   }
//
//   func (s *MyScene) Draw(screen *ebiten.Image) {
//       // Your draw code
//   }
//
// OnExit and SetViewport are already implemented (empty/storing viewport)
type BaseScene struct {
	Viewport geom.Size
}

// OnEnter is called when the scene is loaded
// Override this to initialize your scene
func (bs *BaseScene) OnEnter() {}

// OnExit is called when the scene is removed
// Override this to clean up resources
func (bs *BaseScene) OnExit() {}

// Update is called every frame
// Override this to update your game logic
// Return a new Scene to switch scenes, or nil to stay on this scene
func (bs *BaseScene) Update(dt float64) Scene {
	return nil
}

// Draw is called every frame to render
// Override this to draw your scene
func (bs *BaseScene) Draw(screen *ebiten.Image) {}

// SetViewport is called by the engine to set the viewport size
// You usually don't need to override this - just access bs.Viewport
func (bs *BaseScene) SetViewport(view geom.Size) {
	bs.Viewport = view
}
