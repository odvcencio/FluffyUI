package graphics

import "math"

// Transform represents a 2D affine transform.
type Transform struct {
	A, B, C, D, E, F float64
}

// IdentityTransform returns the identity transform.
func IdentityTransform() Transform {
	return Transform{A: 1, D: 1}
}

// Apply applies the transform to a point.
func (t Transform) Apply(x, y float64) (float64, float64) {
	return t.A*x + t.C*y + t.E, t.B*x + t.D*y + t.F
}

// Mul composes this transform with another.
func (t Transform) Mul(o Transform) Transform {
	return Transform{
		A: t.A*o.A + t.C*o.B,
		B: t.B*o.A + t.D*o.B,
		C: t.A*o.C + t.C*o.D,
		D: t.B*o.C + t.D*o.D,
		E: t.A*o.E + t.C*o.F + t.E,
		F: t.B*o.E + t.D*o.F + t.F,
	}
}

// TranslateTransform returns a translation transform.
func TranslateTransform(dx, dy float64) Transform {
	return Transform{A: 1, D: 1, E: dx, F: dy}
}

// ScaleTransform returns a scaling transform.
func ScaleTransform(sx, sy float64) Transform {
	return Transform{A: sx, D: sy}
}

// RotateTransform returns a rotation transform.
func RotateTransform(angle float64) Transform {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	return Transform{A: cos, B: sin, C: -sin, D: cos}
}

// IsIdentity reports whether the transform is the identity.
func (t Transform) IsIdentity() bool {
	return t.A == 1 && t.B == 0 && t.C == 0 && t.D == 1 && t.E == 0 && t.F == 0
}
