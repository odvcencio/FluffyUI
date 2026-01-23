package animation

import "math"

// Animatable represents a value that can be interpolated.
type Animatable interface {
	Lerp(other Animatable, t float64) Animatable
}

// Float64 is an animatable float.
type Float64 float64

func (f Float64) Lerp(other Animatable, t float64) Animatable {
	o := other.(Float64)
	return Float64(float64(f) + (float64(o)-float64(f))*t)
}

// Int is an animatable integer.
type Int int

func (i Int) Lerp(other Animatable, t float64) Animatable {
	o := other.(Int)
	return Int(math.Round(float64(i) + float64(o-i)*t))
}

// Vec2 is an animatable 2D vector.
type Vec2 struct {
	X, Y float64
}

func (v Vec2) Lerp(other Animatable, t float64) Animatable {
	o := other.(Vec2)
	return Vec2{
		X: v.X + (o.X-v.X)*t,
		Y: v.Y + (o.Y-v.Y)*t,
	}
}

// AnimColor is an animatable RGB color.
type AnimColor struct {
	R, G, B uint8
}

func (c AnimColor) Lerp(other Animatable, t float64) Animatable {
	o := other.(AnimColor)
	return AnimColor{
		R: uint8(float64(c.R) + float64(int(o.R)-int(c.R))*t),
		G: uint8(float64(c.G) + float64(int(o.G)-int(c.G))*t),
		B: uint8(float64(c.B) + float64(int(o.B)-int(c.B))*t),
	}
}

// AnimRect is an animatable rectangle.
type AnimRect struct {
	X, Y, W, H float64
}

func (r AnimRect) Lerp(other Animatable, t float64) Animatable {
	o := other.(AnimRect)
	return AnimRect{
		X: r.X + (o.X-r.X)*t,
		Y: r.Y + (o.Y-r.Y)*t,
		W: r.W + (o.W-r.W)*t,
		H: r.H + (o.H-r.H)*t,
	}
}
