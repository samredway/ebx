package assets

import (
	"github.com/hajimehoshi/ebiten/v2"
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
