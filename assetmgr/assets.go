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
	"math"

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

func (a *Assets) LoadTileSetFromFS(fsys fs.FS, name, path string, frameW, frameH int) {
	sheet := loadEbitenImage(fsys, path)
	a.tiles[name] = splitSheet(sheet, frameW, frameH)
}

func (a *Assets) GetTileSet(name string) []*ebiten.Image {
	tileSet, ok := a.tiles[name]
	if !ok {
		panic(fmt.Sprintf("No tileset with name %s", name))
	}
	return tileSet
}

func (a *Assets) LoadSpriteSheetFromFS(fsys fs.FS, name, path string, frameW, frameH int) {
	sheet := loadEbitenImage(fsys, path)
	a.sprites[name] = splitSheet(sheet, frameW, frameH)
}

func splitSheet(sheet *ebiten.Image, frameW, frameH int) []*ebiten.Image {
	b := sheet.Bounds()
	w := b.Dx()
	h := b.Dy()

	if w%frameW != 0 || h%frameH != 0 {
		panic("Assets sheet not divisible by frame dimensions")
	}

	var tiles []*ebiten.Image
	for y := 0; y < h; y += frameH {
		for x := 0; x < w; x += frameW {
			sub := sheet.SubImage(image.Rect(x, y, x+frameW, y+frameH)).(*ebiten.Image)
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
	tileW   int       // Tile width in pixels
	tileH   int       // Tile height in pixels
	mapSize geom.Size // World width and height in tiles
	layers  [][]int   // Each layer is flat []int of tiles
}

// NewTileMapFromTmx loads in the level from a .tmx file (made in Tiled tile editor)
func NewTileMapFromTmx(fsys fs.FS, pathToTmx string, assets Assets) *TileMap {
	m, err := ebitmx.GetEbitenMapFromFS(fsys, pathToTmx)
	if err != nil {
		panic(fmt.Sprintf("Tilemap not found with at path %s", pathToTmx))
	}

	return &TileMap{
		tileW:   m.TileWidth,
		tileH:   m.TileHeight,
		mapSize: geom.Size{W: m.MapWidth, H: m.MapHeight},
		layers:  m.Layers,
	}
}

func (tm *TileMap) TileW() int         { return tm.tileW }
func (tm *TileMap) TileH() int         { return tm.tileH }
func (tm *TileMap) NumLayers() int     { return len(tm.layers) }
func (tm *TileMap) MapSize() geom.Size { return tm.mapSize }

func (tm *TileMap) OverlapsTiles(x, y, w, h float64, layer int) bool {
	if layer < 0 || layer >= len(tm.layers) {
		panic("invalid layer")
	}

	tw := float64(tm.TileW())
	th := float64(tm.TileH())

	tx0 := int(math.Floor(x / tw))
	ty0 := int(math.Floor(y / th))
	tx1 := int(math.Floor((x+w-1)/tw)) + 1 // exclusive Max
	ty1 := int(math.Floor((y+h-1)/th)) + 1

	// outside = collide with world bounds
	if tx1 <= 0 || ty1 <= 0 || tx0 >= tm.mapSize.W || ty0 >= tm.mapSize.H {
		return true
	}
	if tx0 < 0 {
		tx0 = 0
	}
	if ty0 < 0 {
		ty0 = 0
	}
	if tx1 > tm.mapSize.W {
		tx1 = tm.mapSize.W
	}
	if ty1 > tm.mapSize.H {
		ty1 = tm.mapSize.H
	}

	rowW := tm.mapSize.W
	data := tm.layers[layer]
	for ty := ty0; ty < ty1; ty++ {
		base := ty * rowW
		for tx := tx0; tx < tx1; tx++ {
			if data[base+tx] != 0 {
				return true
			}
		}
	}
	return false
}

func (tm *TileMap) WorldCoordsToTileCoords(wc geom.Vec2) geom.Vec2I {
	return geom.Vec2I{
		X: int(wc.X / float64(tm.tileW)),
		Y: int(wc.Y / float64(tm.tileH)),
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
