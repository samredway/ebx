package camera

import (
	"image"

	"github.com/samredway/ebx/geom"
)

// Camera is a simple cam with functionality to translate wolrd coords to
// viewport coords
type Camera struct {
	geom.Vec2                 // X, Y world coord
	viewport  geom.Size       // viewport size
	bounds    image.Rectangle // Bounding box of whole world
}

// NewCamera creates a new camera at 0,0 that can be set to a position later
// when CenterOn gets called
func NewCamera(viewport geom.Size, bounds image.Rectangle) *Camera {
	pos := geom.Vec2{X: 0.0, Y: 0.0}
	return &Camera{Vec2: pos, viewport: viewport, bounds: bounds}
}

// CenterOn centres the camera on the given position
func (c *Camera) CentreOn(pos geom.Vec2) {
	c.X = pos.X - (float64(c.viewport.W) / 2)
	c.Y = pos.Y - (float64(c.viewport.H) / 2)
	c.clamp()
}

// Apply calculates a screen position from a world position
func (c *Camera) Apply(pos geom.Vec2) geom.Vec2 {
	return geom.Vec2{X: pos.X - c.X, Y: pos.Y - c.Y}
}

// clamp keeps the camera inside world bounds
func (c *Camera) clamp() {
	maxX := float64(c.bounds.Max.X) - float64(c.viewport.W)
	maxY := float64(c.bounds.Max.Y) - float64(c.viewport.H)

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
