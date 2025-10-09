# Ebiten Engine Extension (ebx)

This is my own personal collection of wrappers around ebiten to make 2D game development alittle smoother. 

My motivation for writing this is simply that there is no fully developed engine in Golang - which I prefer to write code in and I like having a collection of small libs that I can mix and match and load in as required keeping my binary very small and focused for the exact game idea I am workign on. I put this together so I could quickly build out and play with some prototypes with very little effort.

As such each feature can be loaded and used as is or you can load the entire collection from the root engine package.


## Key Features

- **TileMap**: I use Tiled the free and open source time map editor as a level designer. The TileMap obj holds the complete map along with funcs for laoding and managing the exported .tmx file with your level in.

- **Camera**: A simple 2D camera to make handling viewport offsets etc a little easer

- **AssetLoader**: Functionality for loading and storing assets (spritesheets, tilesets, backgrounds etc) in RAM at runtime.

- **Ecs**: Very basic ECS system. The aim here is not performance or cache consistency (not really an issue in most 2d games) but rather I just think this is a handy way to manage code and handle object composition.

- **InputSystem**: Handles user input

- **AnimationSystem**: Handles sprite animation

- **CollisionSystem**: Handles collision detection between objects

- **PlayerEntity**: Simple player entity which can be loaded in with a sprite sheet so as to quickly get a game running. 


## Examples

To see exaclty how to use this small library just check the `examples/` folder.
