package style

import "math"

// easingClamp restricts t to the [0, 1] range before applying an easing function.
func easingClamp(t float64) float64 {
	if t <= 0 {
		return 0
	}
	if t >= 1 {
		return 1
	}
	return t
}

// --- Linear ---

func easeLinear(t float64) float64 {
	return easingClamp(t)
}

// --- Quadratic ---

func easeInQuad(t float64) float64 {
	t = easingClamp(t)
	return t * t
}

func easeOutQuad(t float64) float64 {
	t = easingClamp(t)
	return t * (2 - t)
}

func easeInOutQuad(t float64) float64 {
	t = easingClamp(t)
	if t < 0.5 {
		return 2 * t * t
	}
	return -1 + (4-2*t)*t
}

// --- Cubic ---

func easeInCubic(t float64) float64 {
	t = easingClamp(t)
	return t * t * t
}

func easeOutCubic(t float64) float64 {
	t = easingClamp(t)
	t--
	return t*t*t + 1
}

func easeInOutCubic(t float64) float64 {
	t = easingClamp(t)
	if t < 0.5 {
		return 4 * t * t * t
	}
	t = 2*t - 2
	return 0.5*t*t*t + 1
}

// --- Quartic ---

func easeInQuart(t float64) float64 {
	t = easingClamp(t)
	return t * t * t * t
}

func easeOutQuart(t float64) float64 {
	t = easingClamp(t)
	t--
	return 1 - t*t*t*t
}

func easeInOutQuart(t float64) float64 {
	t = easingClamp(t)
	if t < 0.5 {
		return 8 * t * t * t * t
	}
	t = t - 1
	return 1 - 8*t*t*t*t
}

// --- Quintic ---

func easeInQuint(t float64) float64 {
	t = easingClamp(t)
	return t * t * t * t * t
}

func easeOutQuint(t float64) float64 {
	t = easingClamp(t)
	t--
	return t*t*t*t*t + 1
}

func easeInOutQuint(t float64) float64 {
	t = easingClamp(t)
	if t < 0.5 {
		return 16 * t * t * t * t * t
	}
	t = 2*t - 2
	return 0.5*t*t*t*t*t + 1
}

// --- Sinusoidal ---

func easeInSine(t float64) float64 {
	t = easingClamp(t)
	return 1 - math.Cos(t*math.Pi/2)
}

func easeOutSine(t float64) float64 {
	t = easingClamp(t)
	return math.Sin(t * math.Pi / 2)
}

func easeInOutSine(t float64) float64 {
	t = easingClamp(t)
	return -0.5 * (math.Cos(math.Pi*t) - 1)
}

// --- Exponential ---

func easeInExpo(t float64) float64 {
	t = easingClamp(t)
	if t == 0 {
		return 0
	}
	return math.Pow(2, 10*(t-1))
}

func easeOutExpo(t float64) float64 {
	t = easingClamp(t)
	if t == 1 {
		return 1
	}
	return 1 - math.Pow(2, -10*t)
}

func easeInOutExpo(t float64) float64 {
	t = easingClamp(t)
	if t == 0 {
		return 0
	}
	if t == 1 {
		return 1
	}
	if t < 0.5 {
		return 0.5 * math.Pow(2, 20*t-10)
	}
	return 1 - 0.5*math.Pow(2, -20*t+10)
}

// --- Circular ---

func easeInCirc(t float64) float64 {
	t = easingClamp(t)
	return 1 - math.Sqrt(1-t*t)
}

func easeOutCirc(t float64) float64 {
	t = easingClamp(t)
	t--
	return math.Sqrt(1 - t*t)
}

func easeInOutCirc(t float64) float64 {
	t = easingClamp(t)
	t *= 2
	if t < 1 {
		return -0.5 * (math.Sqrt(1-t*t) - 1)
	}
	t -= 2
	return 0.5 * (math.Sqrt(1-t*t) + 1)
}

// --- Elastic ---

func easeInElastic(t float64) float64 {
	t = easingClamp(t)
	if t == 0 || t == 1 {
		return t
	}
	p := 0.3
	s := p / 4
	t--
	return -math.Pow(2, 10*t) * math.Sin((t-s)*2*math.Pi/p)
}

func easeOutElastic(t float64) float64 {
	t = easingClamp(t)
	if t == 0 || t == 1 {
		return t
	}
	p := 0.3
	s := p / 4
	return math.Pow(2, -10*t)*math.Sin((t-s)*2*math.Pi/p) + 1
}

func easeInOutElastic(t float64) float64 {
	t = easingClamp(t)
	if t == 0 || t == 1 {
		return t
	}
	p := 0.3 * 1.5
	s := p / 4
	t = t*2 - 1
	if t < 0 {
		return -0.5 * math.Pow(2, 10*t) * math.Sin((t-s)*2*math.Pi/p)
	}
	return 0.5*math.Pow(2, -10*t)*math.Sin((t-s)*2*math.Pi/p) + 1
}

// --- Back (overshoot) ---

const backOvershoot = 1.70158

func easeInBack(t float64) float64 {
	t = easingClamp(t)
	return t * t * ((backOvershoot+1)*t - backOvershoot)
}

func easeOutBack(t float64) float64 {
	t = easingClamp(t)
	t--
	return t*t*((backOvershoot+1)*t+backOvershoot) + 1
}

func easeInOutBack(t float64) float64 {
	t = easingClamp(t)
	s := backOvershoot * 1.525
	t *= 2
	if t < 1 {
		return 0.5 * t * t * ((s+1)*t - s)
	}
	t -= 2
	return 0.5 * (t*t*((s+1)*t+s) + 2)
}

// --- Bounce ---

func easeOutBounce(t float64) float64 {
	t = easingClamp(t)
	if t < 1.0/2.75 {
		return 7.5625 * t * t
	}
	if t < 2.0/2.75 {
		t -= 1.5 / 2.75
		return 7.5625*t*t + 0.75
	}
	if t < 2.5/2.75 {
		t -= 2.25 / 2.75
		return 7.5625*t*t + 0.9375
	}
	t -= 2.625 / 2.75
	return 7.5625*t*t + 0.984375
}

func easeInBounce(t float64) float64 {
	t = easingClamp(t)
	return 1 - easeOutBounce(1-t)
}

func easeInOutBounce(t float64) float64 {
	t = easingClamp(t)
	if t < 0.5 {
		return 0.5 * easeInBounce(2*t)
	}
	return 0.5*easeOutBounce(2*t-1) + 0.5
}

// EasingByName returns the EasingFunc constant for a given CSS-style name.
// Supported names include the standard CSS names ("linear", "ease-in",
// "ease-out", "ease-in-out") and extended names ("ease-in-quad",
// "ease-out-cubic", "ease-in-elastic", "ease-out-bounce", etc.).
// If the name is not recognized, EaseLinear is returned.
func EasingByName(name string) EasingFunc {
	ef, ok := easingNameMap[name]
	if !ok {
		return EaseLinear
	}
	return ef
}

// easingNameMap maps CSS-style easing names to EasingFunc constants.
var easingNameMap = map[string]EasingFunc{
	"linear":           EaseLinear,
	"ease-in":          EaseIn,
	"ease-out":         EaseOut,
	"ease-in-out":      EaseInOut,
	"ease-in-quad":     EaseInQuad,
	"ease-out-quad":    EaseOutQuad,
	"ease-in-out-quad": EaseInOutQuad,
	"ease-in-cubic":     EaseInCubic,
	"ease-out-cubic":    EaseOutCubic,
	"ease-in-out-cubic": EaseInOutCubic,
	"ease-in-quart":     EaseInQuart,
	"ease-out-quart":    EaseOutQuart,
	"ease-in-out-quart": EaseInOutQuart,
	"ease-in-quint":     EaseInQuint,
	"ease-out-quint":    EaseOutQuint,
	"ease-in-out-quint": EaseInOutQuint,
	"ease-in-sine":      EaseInSine,
	"ease-out-sine":     EaseOutSine,
	"ease-in-out-sine":  EaseInOutSine,
	"ease-in-expo":      EaseInExpo,
	"ease-out-expo":     EaseOutExpo,
	"ease-in-out-expo":  EaseInOutExpo,
	"ease-in-circ":      EaseInCirc,
	"ease-out-circ":     EaseOutCirc,
	"ease-in-out-circ":  EaseInOutCirc,
	"ease-in-elastic":      EaseInElastic,
	"ease-out-elastic":     EaseOutElastic,
	"ease-in-out-elastic":  EaseInOutElastic,
	"ease-in-back":      EaseInBack,
	"ease-out-back":     EaseOutBack,
	"ease-in-out-back":  EaseInOutBack,
	"ease-in-bounce":      EaseInBounce,
	"ease-out-bounce":     EaseOutBounce,
	"ease-in-out-bounce":  EaseInOutBounce,
}

// easingFuncTable maps each EasingFunc constant to its implementation function.
// This is used by applyEasing in transition.go.
var easingFuncTable = map[EasingFunc]func(float64) float64{
	EaseLinear:       easeLinear,
	EaseIn:           easeInQuad,
	EaseOut:          easeOutQuad,
	EaseInOut:        easeInOutQuad,
	EaseInQuad:       easeInQuad,
	EaseOutQuad:      easeOutQuad,
	EaseInOutQuad:    easeInOutQuad,
	EaseInCubic:      easeInCubic,
	EaseOutCubic:     easeOutCubic,
	EaseInOutCubic:   easeInOutCubic,
	EaseInQuart:      easeInQuart,
	EaseOutQuart:     easeOutQuart,
	EaseInOutQuart:   easeInOutQuart,
	EaseInQuint:      easeInQuint,
	EaseOutQuint:     easeOutQuint,
	EaseInOutQuint:   easeInOutQuint,
	EaseInSine:       easeInSine,
	EaseOutSine:      easeOutSine,
	EaseInOutSine:    easeInOutSine,
	EaseInExpo:       easeInExpo,
	EaseOutExpo:      easeOutExpo,
	EaseInOutExpo:    easeInOutExpo,
	EaseInCirc:       easeInCirc,
	EaseOutCirc:      easeOutCirc,
	EaseInOutCirc:    easeInOutCirc,
	EaseInElastic:    easeInElastic,
	EaseOutElastic:   easeOutElastic,
	EaseInOutElastic: easeInOutElastic,
	EaseInBack:       easeInBack,
	EaseOutBack:      easeOutBack,
	EaseInOutBack:    easeInOutBack,
	EaseInBounce:     easeInBounce,
	EaseOutBounce:    easeOutBounce,
	EaseInOutBounce:  easeInOutBounce,
}
