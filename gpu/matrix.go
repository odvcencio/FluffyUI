package gpu

import "math"

// Matrix3 is a 3x3 matrix for 2D transforms.
type Matrix3 struct {
	m [9]float32
}

// Identity returns the identity matrix.
func Identity() Matrix3 {
	return Matrix3{m: [9]float32{1, 0, 0, 0, 1, 0, 0, 0, 1}}
}

// IsIdentity reports whether the matrix is the identity.
func (m Matrix3) IsIdentity() bool {
	return m.m == [9]float32{1, 0, 0, 0, 1, 0, 0, 0, 1}
}

// Mul multiplies two matrices.
func (m Matrix3) Mul(o Matrix3) Matrix3 {
	var r Matrix3
	r.m[0] = m.m[0]*o.m[0] + m.m[1]*o.m[3] + m.m[2]*o.m[6]
	r.m[1] = m.m[0]*o.m[1] + m.m[1]*o.m[4] + m.m[2]*o.m[7]
	r.m[2] = m.m[0]*o.m[2] + m.m[1]*o.m[5] + m.m[2]*o.m[8]

	r.m[3] = m.m[3]*o.m[0] + m.m[4]*o.m[3] + m.m[5]*o.m[6]
	r.m[4] = m.m[3]*o.m[1] + m.m[4]*o.m[4] + m.m[5]*o.m[7]
	r.m[5] = m.m[3]*o.m[2] + m.m[4]*o.m[5] + m.m[5]*o.m[8]

	r.m[6] = m.m[6]*o.m[0] + m.m[7]*o.m[3] + m.m[8]*o.m[6]
	r.m[7] = m.m[6]*o.m[1] + m.m[7]*o.m[4] + m.m[8]*o.m[7]
	r.m[8] = m.m[6]*o.m[2] + m.m[7]*o.m[5] + m.m[8]*o.m[8]
	return r
}

// Apply transforms a point.
func (m Matrix3) Apply(x, y float32) (float32, float32) {
	return m.m[0]*x + m.m[1]*y + m.m[2], m.m[3]*x + m.m[4]*y + m.m[5]
}

// Translate returns a translation matrix.
func Translate(x, y float32) Matrix3 {
	return Matrix3{m: [9]float32{1, 0, x, 0, 1, y, 0, 0, 1}}
}

// Rotate returns a rotation matrix (radians).
func Rotate(radians float32) Matrix3 {
	c := float32(math.Cos(float64(radians)))
	s := float32(math.Sin(float64(radians)))
	return Matrix3{m: [9]float32{c, -s, 0, s, c, 0, 0, 0, 1}}
}

// Scale returns a scale matrix.
func Scale(x, y float32) Matrix3 {
	return Matrix3{m: [9]float32{x, 0, 0, 0, y, 0, 0, 0, 1}}
}
