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
	"encoding/xml"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"math"
	"path/filepath"

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

// TilesetInfo stores metadata about a tileset referenced in the map
type TilesetInfo struct {
	firstGid  int    // First global ID for this tileset
	tsxSource string // Path to the .tsx file
	imgSource string // Path to the image file (extracted from .tsx)
	tileW     int    // Tile width
	tileH     int    // Tile height
}

// tsxTileset represents the structure of a .tsx file
type tsxTileset struct {
	XMLName    xml.Name `xml:"tileset"`
	Name       string   `xml:"name,attr"`
	TileWidth  int      `xml:"tilewidth,attr"`
	TileHeight int      `xml:"tileheight,attr"`
	TileCount  int      `xml:"tilecount,attr"`
	Columns    int      `xml:"columns,attr"`
	Image      struct {
		Source string `xml:"source,attr"`
		Width  int    `xml:"width,attr"`
		Height int    `xml:"height,attr"`
	} `xml:"image"`
}

// parseTSX parses a .tsx file and returns the image source and tile dimensions
func parseTSX(fsys fs.FS, tsxPath string) (imgSource string, tileW, tileH int, err error) {
	data, err := fs.ReadFile(fsys, tsxPath)
	if err != nil {
		return "", 0, 0, fmt.Errorf("failed to read TSX file %s: %w", tsxPath, err)
	}

	var tileset tsxTileset
	if err := xml.Unmarshal(data, &tileset); err != nil {
		return "", 0, 0, fmt.Errorf("failed to parse TSX file %s: %w", tsxPath, err)
	}

	return tileset.Image.Source, tileset.TileWidth, tileset.TileHeight, nil
}

// TileMap represents a whole tilemap - world or level. Currently it is designed
// to work by loading .tmx files (created in the free and open source Tiled level
// editor)
// Note that Assets.tiles[name] will load tiles in the same order as Tiled, however
// tiled uses ids from 1 not 0 so the ids of the tiles in each layer will be the
// same as the index + 1 in Assets.tiles
type TileMap struct {
	tileW    int            // Tile width in pixels
	tileH    int            // Tile height in pixels
	mapSize  geom.Size      // World width and height in tiles
	layers   [][]int        // Each layer is flat []int of tiles
	assets   *Assets        // TileMap owns its assets
	tilesets []TilesetInfo  // Metadata about each tileset
}

// NewTileMapFromTmx loads in the level from a .tmx file (made in Tiled tile editor)
// It automatically parses referenced .tsx files and loads all tilesets
func NewTileMapFromTmx(fsys fs.FS, pathToTmx string, assets *Assets) *TileMap {
	m, err := ebitmx.GetEbitenMapFromFS(fsys, pathToTmx)
	if err != nil {
		panic(fmt.Sprintf("Tilemap not found with at path %s", pathToTmx))
	}
	
	// Get the directory containing the TMX file for resolving relative paths
	tmxDir := filepath.Dir(pathToTmx)
	if tmxDir == "." {
		tmxDir = ""
	}
	
	// Parse each tileset reference and load the images
	var tilesets []TilesetInfo
	for _, tsRef := range m.Tilesets {
		// Resolve TSX path relative to TMX file
		var tsxPath string
		if tmxDir == "" {
			tsxPath = tsRef.Source
		} else {
			tsxPath = filepath.Join(tmxDir, tsRef.Source)
		}
		
		// Parse the TSX file to get image source and tile dimensions
		imgSource, tileW, tileH, err := parseTSX(fsys, tsxPath)
		if err != nil {
			panic(fmt.Sprintf("Failed to parse tileset %s: %v", tsxPath, err))
		}
		
		// Resolve image path relative to TMX file
		var imgPath string
		if tmxDir == "" {
			imgPath = imgSource
		} else {
			imgPath = filepath.Join(tmxDir, imgSource)
		}
		
		// Use just the filename as the key (e.g., "DungeonFloors.png")
		imgFilename := filepath.Base(imgPath)
		
		// Load the tileset image
		assets.LoadTileSetFromFS(fsys, imgFilename, imgPath, tileW, tileH)
		
		// Store tileset metadata
		tilesets = append(tilesets, TilesetInfo{
			firstGid:  tsRef.FirstGid,
			tsxSource: tsRef.Source,
			imgSource: imgSource,
			tileW:     tileW,
			tileH:     tileH,
		})
	}

	return &TileMap{
		tileW:    m.TileWidth,
		tileH:    m.TileHeight,
		mapSize:  geom.Size{W: m.MapWidth, H: m.MapHeight},
		layers:   m.Layers,
		assets:   assets,
		tilesets: tilesets,
	}
}

func (tm *TileMap) TileW() int         { return tm.tileW }
func (tm *TileMap) TileH() int         { return tm.tileH }
func (tm *TileMap) NumLayers() int     { return len(tm.layers) }
func (tm *TileMap) MapSize() geom.Size { return tm.mapSize }

// GetImageById returns the tile image for a given global tile ID
func (tm *TileMap) GetImageById(globalId int) *ebiten.Image {
	if globalId == 0 {
		return nil // 0 means empty tile
	}
	
	// Find which tileset this ID belongs to (iterate backwards to find the highest matching firstGid)
	for i := len(tm.tilesets) - 1; i >= 0; i-- {
		ts := tm.tilesets[i]
		if globalId >= ts.firstGid {
			// Calculate local ID within this tileset
			localId := globalId - ts.firstGid
			
			// Get the tileset by image filename
			imgPath := filepath.Base(filepath.Join(filepath.Dir(""), ts.imgSource))
			tileSet := tm.assets.GetTileSet(imgPath)
			
			if localId >= len(tileSet) {
				panic(fmt.Sprintf("Tile ID %d (local %d) out of range for tileset %s", globalId, localId, ts.imgSource))
			}
			
			return tileSet[localId]
		}
	}
	
	panic(fmt.Sprintf("No tileset found for tile ID %d", globalId))
}

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
