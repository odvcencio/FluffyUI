package animation

import "math"

// EasingFunc maps progress [0,1] to eased progress [0,1].
type EasingFunc func(t float64) float64

// Linear easing (no acceleration).
var Linear EasingFunc = func(t float64) float64 { return t }

// Quadratic easing.
var InQuad EasingFunc = func(t float64) float64 { return t * t }
var OutQuad EasingFunc = func(t float64) float64 { return t * (2 - t) }
var InOutQuad EasingFunc = func(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	return -1 + (4-2*t)*t
}

// Cubic easing.
var InCubic EasingFunc = func(t float64) float64 { return t * t * t }
var OutCubic EasingFunc = func(t float64) float64 {
	t--
	return t*t*t + 1
}
var InOutCubic EasingFunc = func(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	return (t-1)*(2*t-2)*(2*t-2) + 1
}

// Quartic easing.
var InQuart EasingFunc = func(t float64) float64 { return t * t * t * t }
var OutQuart EasingFunc = func(t float64) float64 {
	t--
	return 1 - t*t*t*t
}
var InOutQuart EasingFunc = func(t float64) float64 {
	if t < 0.5 {
		return 8 * t * t * t * t
	}
	t--
	return 1 - 8*t*t*t*t
}

// Exponential easing.
var InExpo EasingFunc = func(t float64) float64 {
	if t == 0 {
		return 0
	}
	return math.Pow(2, 10*(t-1))
}
var OutExpo EasingFunc = func(t float64) float64 {
	if t == 1 {
		return 1
	}
	return 1 - math.Pow(2, -10*t)
}

// Elastic easing.
var InElastic EasingFunc = func(t float64) float64 {
	if t == 0 || t == 1 {
		return t
	}
	return -math.Pow(2, 10*(t-1)) * math.Sin((t-1.1)*5*math.Pi)
}
var OutElastic EasingFunc = func(t float64) float64 {
	if t == 0 || t == 1 {
		return t
	}
	return math.Pow(2, -10*t)*math.Sin((t-0.1)*5*math.Pi) + 1
}

// Bounce easing.
var OutBounce EasingFunc = func(t float64) float64 {
	if t < 1/2.75 {
		return 7.5625 * t * t
	} else if t < 2/2.75 {
		t -= 1.5 / 2.75
		return 7.5625*t*t + 0.75
	} else if t < 2.5/2.75 {
		t -= 2.25 / 2.75
		return 7.5625*t*t + 0.9375
	}
	t -= 2.625 / 2.75
	return 7.5625*t*t + 0.984375
}
var InBounce EasingFunc = func(t float64) float64 {
	return 1 - OutBounce(1-t)
}

// Back easing (overshoot).
var InBack EasingFunc = func(t float64) float64 {
	s := 1.70158
	return t * t * ((s+1)*t - s)
}
var OutBack EasingFunc = func(t float64) float64 {
	s := 1.70158
	t--
	return t*t*((s+1)*t+s) + 1
}

// Easings provides lookup by name.
var Easings = map[string]EasingFunc{
	"linear":     Linear,
	"inQuad":     InQuad,
	"outQuad":    OutQuad,
	"inOutQuad":  InOutQuad,
	"inCubic":    InCubic,
	"outCubic":   OutCubic,
	"inOutCubic": InOutCubic,
	"inQuart":    InQuart,
	"outQuart":   OutQuart,
	"inOutQuart": InOutQuart,
	"inExpo":     InExpo,
	"outExpo":    OutExpo,
	"inElastic":  InElastic,
	"outElastic": OutElastic,
	"inBounce":   InBounce,
	"outBounce":  OutBounce,
	"inBack":     InBack,
	"outBack":    OutBack,
}
