package topdown

import (
	"embed"
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebx/assetmgr"
	"github.com/samredway/ebx/camera"
	"github.com/samredway/ebx/engine"
)

// BaseScene provides common setup for top-down games
// Embed this in your scene to get standard systems and boilerplate for free
//
// Standard usage pattern:
//   type MyScene struct {
//       topdown.BaseScene
//   }
//
//   func (s *MyScene) OnEnter() {
//       // 1. Initialize base systems (tilemap, camera, movement, etc.)
//       s.Init(fs, "map.tmx", 1, 2.0)
//
//       // 2. Load your spritesheets
//       s.Assets.LoadSpriteSheetFromFS(fs, "player.png", 32, 32)
//
//       // 3. Setup animations (adds to s.AnimLibrary)
//       animations := map[string][]*ebiten.Image{
//           "idle_left": idleSheet[0:4],
//           "walk_left": walkSheet[0:4],
//       }
//       topdown.SetupAnimations(s.AnimLibrary, "player", animations, 0.15, true)
//
//       // 4. Create state machine (uses s.MoveSys)
//       stateMachine := topdown.SetupCharacterStateMachine("player", s.MoveSys)
//       // Optional: Add custom transitions here
//       // stateMachine.AddTransition("walk_left", "attack_left", ...)
//
//       // 5. Initialize animation system
//       s.InitAnimationSystem(stateMachine)
//
//       // 6. Create your entities
//       NewPlayer(s.IdGen, s.Assets, s.AnimLibrary, s.RenderSys, ...)
//   }
//
// Note: Update, Draw, OnExit, and SetViewport are already implemented.
// Only override OnEnter (required) and optionally override others for custom behavior.
type BaseScene struct {
	engine.BaseScene // Embeds engine.BaseScene for Scene interface implementation
	IdGen            engine.IdGen
	Camera           *camera.Camera
	PosStore         *engine.PositionStore
	AnimLibrary      *engine.AnimationLibrary
	RenderSys        *engine.RenderSystem
	AnimationSys     *engine.AnimationSystem
	MoveSys          *engine.MovementSystem
	UserInputSys     *engine.UserInputSystem
	TileMap          *assetmgr.TileMap
	Assets           *assetmgr.Assets
}

// Init initializes the base scene with common systems and tilemap
// This is step 1 in the standard setup pattern (see BaseScene documentation)
//
// Creates: IdGen, PosStore, AnimLibrary, Assets, TileMap, Camera, MoveSys, UserInputSys
// Does NOT create: AnimationSys, RenderSys (call InitAnimationSystem after setting up animations)
//
// Note: Viewport is already set by the engine before OnEnter is called.
//
// Parameters:
//   - fs: Embedded filesystem containing assets
//   - tmxFile: Path to TMX tilemap file
//   - collisionLayer: Which layer index to use for collision (usually 1)
//   - zoom: Camera zoom level (e.g., 2.0 for 2x zoom)
func (bs *BaseScene) Init(
	fs embed.FS,
	tmxFile string,
	collisionLayer int,
	zoom float64,
) error {
	bs.IdGen = engine.IdGen{}
	bs.PosStore = engine.NewPositionStore()
	bs.AnimLibrary = engine.NewAnimationLibrary()

	// Create assets and load tilemap
	bs.Assets = assetmgr.NewAssets()
	tileMap, err := assetmgr.NewTileMapFromTmx(fs, tmxFile, bs.Assets)
	if err != nil {
		return fmt.Errorf("failed to load tilemap: %w", err)
	}
	bs.TileMap = tileMap

	// Setup camera (Viewport is already set by engine before OnEnter)
	bs.Camera = camera.NewCamera(
		bs.Viewport,
		image.Rect(
			0,
			0,
			bs.TileMap.MapSize().W*bs.TileMap.TileW(),
			bs.TileMap.MapSize().H*bs.TileMap.TileH(),
		),
	)
	bs.Camera.Zoom = zoom

	// Setup core systems
	bs.MoveSys = engine.NewMovementSystem(bs.PosStore, bs.TileMap, collisionLayer)
	bs.UserInputSys = &engine.UserInputSystem{}

	return nil
}

// InitAnimationSystem initializes the animation system with a state machine
// This is step 5 in the standard setup pattern (see BaseScene documentation)
//
// Call this after:
//   - Init() has been called
//   - Animations have been added to AnimLibrary (via SetupAnimations)
//   - State machine has been created (via SetupCharacterStateMachine or custom)
//
// Creates: AnimationSys, RenderSys
//
// Note: You can add custom transitions to the state machine before passing it here:
//   stateMachine := topdown.SetupCharacterStateMachine("player", bs.MoveSys)
//   stateMachine.AddTransition("walk_left", "attack_left", condition, priority)
//   bs.InitAnimationSystem(stateMachine)
func (bs *BaseScene) InitAnimationSystem(stateMachine *engine.AnimationStateMachine) {
	bs.AnimationSys = engine.NewAnimationSystem(bs.AnimLibrary, stateMachine)
	bs.RenderSys = engine.NewRenderSystem(
		bs.PosStore,
		bs.Camera,
		bs.TileMap,
		bs.AnimationSys,
	)
}

// Update runs all standard system updates
// Override this if you need custom update logic or additional systems
func (bs *BaseScene) Update(dt float64) engine.Scene {
	bs.UserInputSys.Update(dt)
	bs.AnimationSys.Update(dt)
	bs.MoveSys.Update(dt)
	bs.RenderSys.Update(dt)
	return nil
}

// Draw renders the scene
// Override this if you need custom rendering
func (bs *BaseScene) Draw(screen *ebiten.Image) {
	bs.RenderSys.Draw(screen)
}

// Note: OnEnter, OnExit, and SetViewport are inherited from engine.BaseScene
