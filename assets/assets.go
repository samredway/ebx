package assets

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

type Assets struct {
	imgs map[string]*ebiten.Image
}

func NewAssets() *Assets {
	imgs := map[string]*ebiten.Image{}
	return &Assets{imgs: imgs}
}

func (a *Assets) GetImage(imgName string) (*ebiten.Image, error) {
	img, ok := a.imgs[imgName]
	if !ok {
		return nil, fmt.Errorf("no image with name %q", imgName)
	}
	return img, nil
}

func (a *Assets) AddImage(imgName string, img *ebiten.Image) {
	a.imgs[imgName] = img
}
