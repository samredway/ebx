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
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samredway/ebitmx"
)

// ----------------------------------------------------------------------------
// Assets
// ----------------------------------------------------------------------------

type Assets struct {
	imgs    map[string]*ebiten.Image
	tiles   map[string][]*ebiten.Image
	sprites map[string][]*ebiten.Image
}

func (a *Assets) GetImage(imgName string) (*ebiten.Image, error) {
	img, ok := a.imgs[imgName]
	if !ok {
		return nil, fmt.Errorf("no image with name %s", imgName)
	}
	return img, nil
}

func (a *Assets) AddImage(imgName string, img *ebiten.Image) {
	a.imgs[imgName] = img
}

// LoadTileSetFromFS will load form the given file system (see doc at top of file)
// name: name of the tileset as stated in your .tmx file
// path: path within your fs.FS object to png file
// frameW, frameH: the tile size in px
func (a *Assets) LoadTileSetFromFS(fsys fs.FS, name, path string, frameW, frameH int) error {
	sheet, err := loadEbitenImage(fsys, path)
	if err != nil {
		return fmt.Errorf("failed to load tileset %s: %w", name, err)
	}
	tiles, err := splitSheet(sheet, frameW, frameH)
	if err != nil {
		return fmt.Errorf("failed to split tileset %s: %w", name, err)
	}
	a.tiles[name] = tiles
	return nil
}

func (a *Assets) GetTileSet(name string) ([]*ebiten.Image, error) {
	tileSet, ok := a.tiles[name]
	if !ok {
		return nil, fmt.Errorf("no tileset with name %s", name)
	}
	return tileSet, nil
}

// LoadSpriteSheetFromFS loads a spritesheet from the filesystem object passed in
func (a *Assets) LoadSpriteSheetFromFS(fsys fs.FS, name, path string, frameW, frameH int) error {
	sheet, err := loadEbitenImage(fsys, path)
	if err != nil {
		return fmt.Errorf("failed to load sprite sheet %s: %w", path, err)
	}
	sprites, err := splitSheet(sheet, frameW, frameH)
	if err != nil {
		return fmt.Errorf("failed to split sprite sheet %s: %w", path, err)
	}
	a.sprites[name] = sprites
	return nil
}

func (a *Assets) GetSpriteSheet(name string) ([]*ebiten.Image, error) {
	spriteSheet, ok := a.sprites[name]
	if !ok {
		return nil, fmt.Errorf("no sprite sheet with name %s", name)
	}
	return spriteSheet, nil
}

// NewAssets is constructor for Assets
func NewAssets() *Assets {
	return &Assets{
		imgs:    map[string]*ebiten.Image{},
		tiles:   map[string][]*ebiten.Image{},
		sprites: map[string][]*ebiten.Image{},
	}
}

func splitSheet(sheet *ebiten.Image, frameW, frameH int) ([]*ebiten.Image, error) {
	b := sheet.Bounds()
	w := b.Dx()
	h := b.Dy()

	if w%frameW != 0 || h%frameH != 0 {
		return nil, fmt.Errorf("sheet dimensions (%dx%d) not divisible by frame dimensions (%dx%d)", w, h, frameW, frameH)
	}

	var tiles []*ebiten.Image
	for y := 0; y < h; y += frameH {
		for x := 0; x < w; x += frameW {
			sub := sheet.SubImage(image.Rect(x, y, x+frameW, y+frameH)).(*ebiten.Image)
			tiles = append(tiles, sub)
		}
	}
	return tiles, nil
}

func loadEbitenImage(fsys fs.FS, path string) (*ebiten.Image, error) {
	f, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	img, _, err := image.Decode(bytes.NewReader(f))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image %s: %w", path, err)
	}

	return ebiten.NewImageFromImage(img), nil
}

// ----------------------------------------------------------------------------
// Tilesets
// ----------------------------------------------------------------------------

// FirstGid represents the first global tile ID for a tileset in a TMX map
type FirstGid int

// TilesetInfo stores metadata about a tileset referenced in the map
type TilesetInfo struct {
	imgSource string // Path to the image file
	tileW     int    // Tile width
	tileH     int    // Tile height
}

// TilesetManager manages tileset metadata and tile ID resolution
type TilesetManager struct {
	infos  map[FirstGid]TilesetInfo // Tileset metadata keyed by firstGid
	assets *Assets                  // Reference to assets for loading tile images
}

// Add registers a tileset with its firstGid
func (ts *TilesetManager) Add(firstGid FirstGid, info TilesetInfo) {
	ts.infos[firstGid] = info
}

// GetImageForTileId returns the tile image for a given global tile ID
func (ts *TilesetManager) GetImageForTileId(globalId int) (*ebiten.Image, error) {
	if globalId == 0 {
		return nil, nil // 0 means empty tile
	}

	// Find which tileset this ID belongs to by checking firstGids in descending order
	var matchingFirstGid FirstGid
	for firstGid := range ts.infos {
		if globalId >= int(firstGid) && firstGid > matchingFirstGid {
			matchingFirstGid = firstGid
		}
	}

	if matchingFirstGid == 0 {
		return nil, fmt.Errorf("no tileset found for tile ID %d", globalId)
	}

	info := ts.infos[matchingFirstGid]
	localId := globalId - int(matchingFirstGid)

	// Get the tileset by image filename
	imgFilename := filepath.Base(info.imgSource)
	tileSet, err := ts.assets.GetTileSet(imgFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to get tileset %s: %w", imgFilename, err)
	}

	if localId >= len(tileSet) {
		return nil, fmt.Errorf("tile ID %d (local %d) out of range for tileset %s (has %d tiles)", globalId, localId, info.imgSource, len(tileSet))
	}

	return tileSet[localId], nil
}

// NewTilesetManager creates a new Tilesets manager
func NewTilesetManager(assets *Assets) *TilesetManager {
	return &TilesetManager{
		infos:  map[FirstGid]TilesetInfo{},
		assets: assets,
	}
}

// ----------------------------------------------------------------------------
// TileMap
// ----------------------------------------------------------------------------

// TileMap represents a whole tilemap - world or level. Currently it is designed
// to work by loading .tmx files (created in the free and open source Tiled level
// editor) and has a dependendency on ebitmx
// Note that Assets.tiles[name] will load tiles in the same order as Tiled, however
// tiled uses ids from 1 not 0 so the ids of the tiles in each layer will be the
// same as the index + 1 in Assets.tiles
type TileMap struct {
	*ebitmx.EbitenMap                 // Embedded map data from ebitmx
	tilesets          *TilesetManager // Tileset manager
}

// NumLayers returs the number of layers in the tilemap
func (tm *TileMap) NumLayers() int { return len(tm.Layers) }

// GetImageById returns the tile image for a given global tile ID
func (tm *TileMap) GetImageById(globalId int) (*ebiten.Image, error) {
	return tm.tilesets.GetImageForTileId(globalId)
}

// OverlapsTiles returns true if a position overlaps any tiles in a given layer
// used to check collision for example
func (tm *TileMap) OverlapsTiles(x, y, w, h float64, layer int) (bool, error) {
	if layer < 0 || layer >= len(tm.Layers) {
		return false, fmt.Errorf("invalid layer index: %d (map has %d layers)", layer, len(tm.Layers))
	}

	tw := float64(tm.TileWidth)
	th := float64(tm.TileHeight)

	tx0 := int(math.Floor(x / tw))
	ty0 := int(math.Floor(y / th))
	tx1 := int(math.Floor((x+w-1)/tw)) + 1 // exclusive Max
	ty1 := int(math.Floor((y+h-1)/th)) + 1

	// outside = collide with world bounds
	if tx1 <= 0 || ty1 <= 0 || tx0 >= tm.MapWidth || ty0 >= tm.MapHeight {
		return true, nil
	}
	if tx0 < 0 {
		tx0 = 0
	}
	if ty0 < 0 {
		ty0 = 0
	}
	if tx1 > tm.MapWidth {
		tx1 = tm.MapWidth
	}
	if ty1 > tm.MapHeight {
		ty1 = tm.MapHeight
	}

	rowW := tm.MapWidth
	data := tm.Layers[layer]
	for ty := ty0; ty < ty1; ty++ {
		base := ty * rowW
		for tx := tx0; tx < tx1; tx++ {
			if data[base+tx] != 0 {
				return true, nil
			}
		}
	}
	return false, nil
}

// ForEachIn allows user to run a function (for example to render) each tile within
// the bounds (in terms of tilesx and tilesy coords) of a rect
func (tm *TileMap) ForEachIn(area image.Rectangle, layer int, fn func(tx, ty, id int)) error {
	if layer < 0 || layer >= len(tm.Layers) {
		return fmt.Errorf("invalid layer index: %d (map has %d layers)", layer, len(tm.Layers))
	}

	// clamp to map bounds
	if area.Min.X < 0 {
		area.Min.X = 0
	}
	if area.Min.Y < 0 {
		area.Min.Y = 0
	}
	if area.Max.X > tm.MapWidth {
		area.Max.X = tm.MapWidth
	}
	if area.Max.Y > tm.MapHeight {
		area.Max.Y = tm.MapHeight
	}

	data := tm.Layers[layer]
	w := tm.MapWidth
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
	return nil
}

func (tm *TileMap) loadTilesets(fsys fs.FS, tmxDir string, tilesetRefs []ebitmx.TilesetRef) error {
	for _, tsRef := range tilesetRefs {
		info, err := tm.loadTileset(fsys, tmxDir, tsRef)
		if err != nil {
			return err
		}
		tm.tilesets.Add(FirstGid(tsRef.FirstGid), info)
	}
	return nil
}

func (tm *TileMap) loadTileset(fsys fs.FS, tmxDir string, tsRef ebitmx.TilesetRef) (TilesetInfo, error) {
	tsxPath := resolvePath(tmxDir, tsRef.Source)

	tsxBytes, err := fs.ReadFile(fsys, tsxPath)
	if err != nil {
		return TilesetInfo{}, fmt.Errorf("failed to read TSX file %s: %w", tsxPath, err)
	}

	tileset, err := ebitmx.ParseTSX(tsxBytes)
	if err != nil {
		return TilesetInfo{}, fmt.Errorf("failed to parse TSX file %s: %w", tsxPath, err)
	}

	imgPath := resolvePath(tmxDir, tileset.Image.Source)
	imgFilename := filepath.Base(imgPath)

	if err := tm.tilesets.assets.LoadTileSetFromFS(fsys, imgFilename, imgPath, tileset.TileWidth, tileset.TileHeight); err != nil {
		return TilesetInfo{}, fmt.Errorf("failed to load tileset image %s: %w", imgPath, err)
	}

	return TilesetInfo{
		imgSource: tileset.Image.Source,
		tileW:     tileset.TileWidth,
		tileH:     tileset.TileHeight,
	}, nil
}

// NewTileMapFromTmx loads in the level from a .tmx file (made in Tiled tile editor)
// It automatically parses referenced .tsx files and loads all tilesets
func NewTileMapFromTmx(fsys fs.FS, pathToTmx string, assets *Assets) (*TileMap, error) {
	m, err := ebitmx.GetEbitenMapFromFS(fsys, pathToTmx)
	if err != nil {
		return nil, fmt.Errorf("failed to load TMX file %s: %w", pathToTmx, err)
	}

	tileMap := &TileMap{
		EbitenMap: m,
		tilesets:  NewTilesetManager(assets),
	}

	tmxDir := normalizeTmxDir(pathToTmx)
	if err := tileMap.loadTilesets(fsys, tmxDir, m.Tilesets); err != nil {
		return nil, fmt.Errorf("failed to load tilesets for %s: %w", pathToTmx, err)
	}

	return tileMap, nil
}

func resolvePath(baseDir, path string) string {
	if baseDir == "" {
		return path
	}
	return filepath.Join(baseDir, path)
}

func normalizeTmxDir(pathToTmx string) string {
	tmxDir := filepath.Dir(pathToTmx)
	if tmxDir == "." {
		return ""
	}
	return tmxDir
}
