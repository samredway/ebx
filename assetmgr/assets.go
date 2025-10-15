// Package assetmgr provides functions for loading and managing images,
// sprite sheets, and Tiled (.tmx) maps for use with Ebiten.
//
// It expects your assets to be embedded in the binary and accepts any io/fs.FS
// implementation to locate and read those files. For quick testing, you can
// use the default filesystem by passing os.DirFS(".").
//
// A typical setup for embedding looks like this:
//
//	package assets
//
//	import "embed"
//
//	//go:embed *.tmx *.png
//	var GameFS embed.FS
//
// You can then pass assets.GameFS to the assetmgr functions when loading data.

package assetmgr

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"

	"github.com/Rulox/ebitmx"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebx/geom"
)

type Assets struct {
	imgs    map[string]*ebiten.Image
	tiles   map[string][]*ebiten.Image
	sprites map[string][]*ebiten.Image
}

func NewAssets() *Assets {
	return &Assets{
		imgs:    map[string]*ebiten.Image{},
		tiles:   map[string][]*ebiten.Image{},
		sprites: map[string][]*ebiten.Image{},
	}
}

func (a *Assets) GetImage(imgName string) (*ebiten.Image, error) {
	img, ok := a.imgs[imgName]
	if !ok {
		panic("no image with name " + imgName)
	}
	return img, nil
}

func (a *Assets) AddImage(imgName string, img *ebiten.Image) {
	a.imgs[imgName] = img
}

func (a *Assets) LoadTileSetFromFS(fsys fs.FS, name, path string, tileSize int) {
	sheet := loadEbitenImage(fsys, path)
	a.tiles[name] = splitSheet(sheet, tileSize)
}

func (a *Assets) GetTileSet(name string) []*ebiten.Image {
	tileSet, ok := a.tiles[name]
	if !ok {
		panic(fmt.Sprintf("No tileset with name %s", name))
	}
	return tileSet
}

func (a *Assets) LoadSpriteSheetFromFS(fsys fs.FS, name, path string, frameSize int) {
	sheet := loadEbitenImage(fsys, path)
	a.sprites[name] = splitSheet(sheet, frameSize)
}

func splitSheet(sheet *ebiten.Image, frameSize int) []*ebiten.Image {
	b := sheet.Bounds()
	w := b.Dx()
	h := b.Dy()

	if w%frameSize != 0 || h%frameSize != 0 {
		panic("Assets sheet not divisible by frame size")
	}

	var tiles []*ebiten.Image
	for y := 0; y < h; y += frameSize {
		for x := 0; x < w; x += frameSize {
			sub := sheet.SubImage(image.Rect(x, y, x+frameSize, y+frameSize)).(*ebiten.Image)
			tiles = append(tiles, sub)
		}
	}
	return tiles
}

func loadEbitenImage(fsys fs.FS, path string) *ebiten.Image {
	f, err := fs.ReadFile(fsys, path)
	if err != nil {
		panic(fmt.Sprintf("Unable to load file from path %s got ERROR: %v", path, err))
	}

	img, _, err := image.Decode(bytes.NewReader(f))
	if err != nil {
		panic(fmt.Sprintf("Unable to decode bytes for image %s ERROR: %v", path, err))
	}

	return ebiten.NewImageFromImage(img)
}

// TileMap represents a whole tilemap - world or level. Currently it is designed
// to work by loading .tmx files (created in the free and open source Tiled level
// editor)
// Note that Assets.tiles[name] will load tiles in the same order as Tiled, however
// tiled uses ids from 1 not 0 so the ids of the tiles in each layer will be the
// same as the index + 1 in Assets.tiles
type TileMap struct {
	tileSize int       // Assume tiles are square
	mapSize  geom.Size // World width and height in tiles
	layers   [][]int   // Each layer is flat []int of tiles
}

// NewTileMapFromTmx loads in the level from a .tmx file (made in Tiled tile editor)
func NewTileMapFromTmx(fsys fs.FS, pathToTmx string, assets Assets) *TileMap {
	m, err := ebitmx.GetEbitenMapFromFS(fsys, pathToTmx)
	if err != nil {
		panic(fmt.Sprintf("Tilemap not found with at path %s", pathToTmx))
	}

	return &TileMap{
		tileSize: m.TileHeight, // Assume tiles are square
		mapSize:  geom.Size{W: m.MapWidth, H: m.MapHeight},
		layers:   m.Layers,
	}
}

func (tm *TileMap) TileSize() int      { return tm.tileSize }
func (tm *TileMap) NumLayers() int     { return len(tm.layers) }
func (tm *TileMap) MapSize() geom.Size { return tm.mapSize }

// IsColliding checks if the given coordinates are in world bounds (assumes a
// collision if an out of bounds coord is given) or if there is a tile of this
// layer that is at this tile mape coord
func (tm *TileMap) IsColliding(collRect image.Rectangle, layer int) bool {
	// check world bounds
	if collRect.Min.X < 0 || collRect.Max.X > tm.mapSize.W {
		return true
	}
	if collRect.Min.Y < 0 || collRect.Max.Y > tm.mapSize.H {
		return true
	}
	// find overlapping tiles that have non 0 values

	// TOP LEFT can only collide with upper and left tiles
	tl := collRect.Min.Y*tm.mapSize.W + collRect.Min.X
	if tm.layers[layer][tl] != 0 {
		return true
	}
	// TOP RIGHT can only collide with up and right tiles
	tr := collRect.Min.Y*tm.mapSize.W + collRect.Max.X
	if tm.layers[layer][tr] != 0 {
		return true
	}
	// BOTTOM LEFT can only collide with upper and left tiles
	bl := collRect.Max.Y*tm.mapSize.W + collRect.Min.X
	if tm.layers[layer][bl] != 0 {
		return true
	}
	// BOTTOM RIGHT can only collide with up and right tiles
	br := collRect.Max.Y*tm.mapSize.W + collRect.Max.X
	if tm.layers[layer][br] != 0 {
		return true
	}
	return false
}

func (tm *TileMap) WorldCoordsToTileCoords(wc geom.Vec2) geom.Vec2I {
	return geom.Vec2I{
		X: int(wc.X / float64(tm.tileSize)),
		Y: int(wc.Y / float64(tm.tileSize)),
	}
}

// ForEachIn allows user to run a function (for example to render) each tile within
// the bounds (in terms of tilesx and tilesy coords) of a rect
func (tm *TileMap) ForEachIn(area image.Rectangle, layer int, fn func(tx, ty, id int)) {
	if layer < 0 || layer >= len(tm.layers) {
		panic(fmt.Sprintf("invalid layer index: %d", layer))
	}

	// clamp to map bounds
	if area.Min.X < 0 {
		area.Min.X = 0
	}
	if area.Min.Y < 0 {
		area.Min.Y = 0
	}
	if area.Max.X > tm.mapSize.W {
		area.Max.X = tm.mapSize.W
	}
	if area.Max.Y > tm.mapSize.H {
		area.Max.Y = tm.mapSize.H
	}

	data := tm.layers[layer]
	w := tm.mapSize.W
	for ty := area.Min.Y; ty < area.Max.Y; ty++ {
		row := ty * w
		for tx := area.Min.X; tx < area.Max.X; tx++ {
			id := data[row+tx]
			if id == 0 {
				continue
			}
			fn(tx, ty, id)
		}
	}
}
