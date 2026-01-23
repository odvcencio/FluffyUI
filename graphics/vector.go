package graphics

import "math"

func normalizeVec(v Vector2) (Vector2, bool) {
	length := math.Hypot(v.X, v.Y)
	if length == 0 {
		return Vector2{}, false
	}
	return Vector2{X: v.X / length, Y: v.Y / length}, true
}

func dotVec(a, b Vector2) float64 {
	return a.X*b.X + a.Y*b.Y
}

func crossVec(a, b Vector2) float64 {
	return a.X*b.Y - a.Y*b.X
}

func clampFloat(v, minValue, maxValue float64) float64 {
	if v < minValue {
		return minValue
	}
	if v > maxValue {
		return maxValue
	}
	return v
}
