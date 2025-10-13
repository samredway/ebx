// Package assetmgr provides loading and management of images, sprite sheets,
// and Tiled (.tmx) maps for use with Ebiten.
package assetmgr

import (
	"fmt"
	"github.com/Rulox/ebitmx"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebx/geom"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
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

func (a *Assets) LoadTileSetFromPath(name, path string, tileSize int) {
	sheet := loadEbitenImage(path)
	a.tiles[name] = splitSheet(sheet, tileSize)
}

func (a *Assets) LoadSpriteSheetFromPath(name, path string, frameSize int) {
	sheet := loadEbitenImage(path)
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

func loadEbitenImage(path string) *ebiten.Image {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	return ebiten.NewImageFromImage(img)
}

// TileMap represents a whole tilemap - world or level. Currently it is designed
// to work by loading .tmx files (created in the free and open source Tiled level
// editor)
type TileMap struct {
	tileSize int       // Assume tiles are square
	mapSize  geom.Size // World width and height in tiles
	layers   [][]int   // Each layer is flat []int of tiles
}

// NewTileMapFromTmx loads in the level from a .tmx file (made in Tiled tile editor)
func NewTileMapFromTmx(pathToTmx string, assets Assets) *TileMap {
	m, err := ebitmx.GetEbitenMap(pathToTmx)
	if err != nil {
		panic("Tilemap not found")
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
