package tiles

import (
	"github.com/Rulox/ebitmx"
	// "github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebx/assets"
	"github.com/samredway/ebx/geom"
)

type TileMap struct {
	tileSize int     // Assume tiles are square
	mapW     int     // World width in tiles
	mapH     int     // World height in tiles
	layers   [][]int // Each layer is flat []int of tiles
}

// NewTileManager loads in the level as a .tmx file (made in Tiled tile editor)
func NewTileMap(pathToTmx string, assets *assets.Assets) *TileMap {
	m, err := ebitmx.GetEbitenMap(pathToTmx)
	if err != nil {
		panic("Tilemap not found")
	}

	return &TileMap{
		tileSize: m.TileHeight, // Assume tiles are square
		mapW:     m.MapWidth,
		mapH:     m.MapHeight,
		layers:   m.Layers,
	}
}

// Manaages rendering tile entities
// TODO will have a collection of tile entities
// Tiles system will be initialised with each tile having the correct coords and
// image name
// Will have logic for drawing the tiles
type TileSystem struct {
	tiles []Tile
	tm    *TileMap
}

type Tile struct {
	coords  geom.Vec2
	tile_id int
}

// // Draw tiles to screen
// func (tm *TileManager) Draw(screen *ebiten.Image, cam *Camera) {
// 	// Calculate visible tile range based on camera position
// 	startX := int(cam.X) / tm.tileSize
// 	startY := int(cam.Y) / tm.tileSize
// 	endX := startX + (cam.Width / tm.tileSize) + 2  // +2 for buffer
// 	endY := startY + (cam.Height / tm.tileSize) + 2 // +2 for buffer
//
// 	// Clamp to map bounds
// 	if startX < 0 {
// 		startX = 0
// 	}
// 	if startY < 0 {
// 		startY = 0
// 	}
// 	if endX > tm.mapW {
// 		endX = tm.mapW
// 	}
// 	if endY > tm.mapH {
// 		endY = tm.mapH
// 	}
//
// 	// Draw each layer
// 	for _, layer := range tm.layers {
// 		for y := startY; y < endY; y++ {
// 			for x := startX; x < endX; x++ {
// 				tileIndex := y*tm.mapW + x
// 				if tileIndex >= len(layer) {
// 					continue
// 				}
//
// 				tileID := layer[tileIndex]
// 				if tileID == 0 {
// 					continue // Skip empty tiles
// 				}
//
// 				// Calculate world position
// 				worldX := float64(x * tm.tileSize)
// 				worldY := float64(y * tm.tileSize)
//
// 				// Apply camera transform to get screen position
// 				screenPos := cam.Apply(Position{worldX, worldY})
//
// 				// Create draw options
// 				op := &ebiten.DrawImageOptions{}
// 				op.GeoM.Translate(screenPos.X, screenPos.Y)
//
// 				// Draw the tile (using Tile1 for all non-zero tiles)
// 				screen.DrawImage(tm.assets.GetImage(Tile1), op)
// 			}
// 		}
// 	}
// }
