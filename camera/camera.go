package camera

import (
	"image"

	"github.com/samredway/ebx/geom"
)

type Camera struct {
	geom.Vec2                    // X, Y world coord
	viewW, viewH int             // viewport size
	bounds       image.Rectangle // Bounding box of whole world
}

func NewCamera(coords geom.Vec2, w, h int, bounds image.Rectangle) *Camera {
	return &Camera{coords, w, h, bounds}
}

// CenterOn centres the camera on the given position
func (c *Camera) CentreOn(pos geom.Vec2) {
	c.X = pos.X - (float64(c.viewW) / 2)
	c.Y = pos.Y - (float64(c.viewH) / 2)
	c.clamp()
}

// Apply calculates a screen position from a world position
func (c *Camera) Apply(pos geom.Vec2) geom.Vec2 {
	return geom.Vec2{X: pos.X - c.X, Y: pos.Y - c.Y}
}

// clamp keeps the camera inside world bounds
func (c *Camera) clamp() {
	maxX := float64(c.bounds.Max.X) - float64(c.viewW)
	maxY := float64(c.bounds.Max.Y) - float64(c.viewH)

	if c.X < float64(c.bounds.Min.X) {
		c.X = float64(c.bounds.Min.X)
	}
	if c.X > maxX {
		c.X = maxX
	}
	if c.Y < float64(c.bounds.Min.Y) {
		c.Y = float64(c.bounds.Min.Y)
	}
	if c.Y > maxY {
		c.Y = maxY
	}
}
