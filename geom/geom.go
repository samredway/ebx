package geom

import "math"

type Vec2 struct{ X, Y float64 }

// Normalize returns a unit-length vector pointing in the same direction as vec.
// If vec has zero length, it returns the zero vector unchanged.
func Normalize(vec Vec2) Vec2 {
	length := math.Hypot(vec.X, vec.Y)
	if length == 0 {
		return Vec2{}
	}
	return Vec2{
		X: vec.X / length,
		Y: vec.Y / length,
	}
}

type Size struct{ W, H int }
