package assets

import (
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

type Assets struct {
	imgs         map[string]*ebiten.Image
	tileMaps     map[string][]*ebiten.Image
	spriteSheets map[string][]*ebiten.Image
}

func NewAssets() *Assets {
	imgs := map[string]*ebiten.Image{}
	return &Assets{imgs: imgs}
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
	b := sheet.Bounds()
	w := b.Dx()
	h := b.Dy()

	var tiles []*ebiten.Image
	for y := 0; y < h; y += tileSize {
		for x := 0; x < w; x += tileSize {
			sub := sheet.SubImage(image.Rect(x, y, x+tileSize, y+tileSize)).(*ebiten.Image)
			tiles = append(tiles, sub)
		}
	}
	a.tileMaps[name] = tiles
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
