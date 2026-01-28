package graphics

import "github.com/odvcencio/fluffyui/backend"

// Point represents a pixel coordinate.
type Point struct {
	X, Y int
}

// Vector2 represents a 2D vector with float precision.
type Vector2 struct {
	X, Y float64
}

// Color represents a terminal color.
type Color = backend.Color

// Pixel represents a single sub-cell point.
type Pixel struct {
	Set   bool
	Color Color
	Alpha float32
}

// Rect represents a rectangular region.
type Rect struct {
	X, Y          int
	Width, Height int
}
