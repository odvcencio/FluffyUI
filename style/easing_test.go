package style

import (
	"math"
	"testing"
)

const epsilon = 1e-9

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}

// TestEasingBoundaries verifies that all easing functions return 0 at t=0 and 1 at t=1.
func TestEasingBoundaries(t *testing.T) {
	funcs := []struct {
		name string
		fn   func(float64) float64
	}{
		{"easeLinear", easeLinear},
		{"easeInQuad", easeInQuad},
		{"easeOutQuad", easeOutQuad},
		{"easeInOutQuad", easeInOutQuad},
		{"easeInCubic", easeInCubic},
		{"easeOutCubic", easeOutCubic},
		{"easeInOutCubic", easeInOutCubic},
		{"easeInQuart", easeInQuart},
		{"easeOutQuart", easeOutQuart},
		{"easeInOutQuart", easeInOutQuart},
		{"easeInQuint", easeInQuint},
		{"easeOutQuint", easeOutQuint},
		{"easeInOutQuint", easeInOutQuint},
		{"easeInSine", easeInSine},
		{"easeOutSine", easeOutSine},
		{"easeInOutSine", easeInOutSine},
		{"easeInExpo", easeInExpo},
		{"easeOutExpo", easeOutExpo},
		{"easeInOutExpo", easeInOutExpo},
		{"easeInCirc", easeInCirc},
		{"easeOutCirc", easeOutCirc},
		{"easeInOutCirc", easeInOutCirc},
		{"easeInElastic", easeInElastic},
		{"easeOutElastic", easeOutElastic},
		{"easeInOutElastic", easeInOutElastic},
		{"easeInBack", easeInBack},
		{"easeOutBack", easeOutBack},
		{"easeInOutBack", easeInOutBack},
		{"easeInBounce", easeInBounce},
		{"easeOutBounce", easeOutBounce},
		{"easeInOutBounce", easeInOutBounce},
	}
	for _, tc := range funcs {
		t.Run(tc.name, func(t *testing.T) {
			v0 := tc.fn(0)
			if !approxEqual(v0, 0) {
				t.Errorf("%s(0) = %f, want 0", tc.name, v0)
			}
			v1 := tc.fn(1)
			if !approxEqual(v1, 1) {
				t.Errorf("%s(1) = %f, want 1", tc.name, v1)
			}
		})
	}
}

// TestEaseLinearMidpoint verifies linear interpolation at the midpoint.
func TestEaseLinearMidpoint(t *testing.T) {
	v := easeLinear(0.5)
	if !approxEqual(v, 0.5) {
		t.Errorf("easeLinear(0.5) = %f, want 0.5", v)
	}
}

// TestEaseInQuadMidpoint checks that EaseInQuad(0.5) == 0.25.
func TestEaseInQuadMidpoint(t *testing.T) {
	v := easeInQuad(0.5)
	if !approxEqual(v, 0.25) {
		t.Errorf("easeInQuad(0.5) = %f, want 0.25", v)
	}
}

// TestEaseOutQuadMidpoint checks that EaseOutQuad(0.5) == 0.75.
func TestEaseOutQuadMidpoint(t *testing.T) {
	v := easeOutQuad(0.5)
	if !approxEqual(v, 0.75) {
		t.Errorf("easeOutQuad(0.5) = %f, want 0.75", v)
	}
}

// TestEaseInOutQuadMidpoint checks that EaseInOutQuad(0.5) == 0.5.
func TestEaseInOutQuadMidpoint(t *testing.T) {
	v := easeInOutQuad(0.5)
	if !approxEqual(v, 0.5) {
		t.Errorf("easeInOutQuad(0.5) = %f, want 0.5", v)
	}
}

// TestEaseInCubicMidpoint checks that EaseInCubic(0.5) == 0.125.
func TestEaseInCubicMidpoint(t *testing.T) {
	v := easeInCubic(0.5)
	if !approxEqual(v, 0.125) {
		t.Errorf("easeInCubic(0.5) = %f, want 0.125", v)
	}
}

// TestEaseOutCubicMidpoint checks that EaseOutCubic(0.5) == 0.875.
func TestEaseOutCubicMidpoint(t *testing.T) {
	v := easeOutCubic(0.5)
	if !approxEqual(v, 0.875) {
		t.Errorf("easeOutCubic(0.5) = %f, want 0.875", v)
	}
}

// TestEaseInExpoMidpoint verifies exponential easing at the midpoint.
func TestEaseInExpoMidpoint(t *testing.T) {
	v := easeInExpo(0.5)
	// 2^(10*(0.5-1)) = 2^(-5) = 0.03125
	expected := math.Pow(2, -5)
	if !approxEqual(v, expected) {
		t.Errorf("easeInExpo(0.5) = %f, want %f", v, expected)
	}
}

// TestEaseOutExpoMidpoint verifies exponential easing at the midpoint.
func TestEaseOutExpoMidpoint(t *testing.T) {
	v := easeOutExpo(0.5)
	// 1 - 2^(-10*0.5) = 1 - 2^(-5) = 1 - 0.03125 = 0.96875
	expected := 1 - math.Pow(2, -5)
	if !approxEqual(v, expected) {
		t.Errorf("easeOutExpo(0.5) = %f, want %f", v, expected)
	}
}

// TestEaseOutBounceKnownValues checks the classic bounce at known thresholds.
func TestEaseOutBounceKnownValues(t *testing.T) {
	// At t=0.5 (first segment boundary: 1/2.75 ~ 0.3636)
	// t=0.5 falls in second segment
	v := easeOutBounce(0.5)
	if v <= 0 || v >= 1 {
		t.Errorf("easeOutBounce(0.5) = %f, expected in (0, 1)", v)
	}
	// At the exact start of second segment
	v = easeOutBounce(1.0 / 2.75)
	// 7.5625 * (1/2.75)^2 = 7.5625 * 0.132231... = 1.0
	if !approxEqual(v, 1.0) {
		// Actually 7.5625 * (1/2.75)^2 = 7.5625/7.5625 = 1.0
		t.Errorf("easeOutBounce(1/2.75) = %f, want 1.0", v)
	}
}

// TestEasingClampBelowZero verifies clamping for t < 0.
func TestEasingClampBelowZero(t *testing.T) {
	funcs := []struct {
		name string
		fn   func(float64) float64
	}{
		{"easeLinear", easeLinear},
		{"easeInQuad", easeInQuad},
		{"easeOutQuad", easeOutQuad},
		{"easeInOutQuad", easeInOutQuad},
		{"easeInCubic", easeInCubic},
		{"easeOutCubic", easeOutCubic},
		{"easeInExpo", easeInExpo},
		{"easeOutExpo", easeOutExpo},
		{"easeInElastic", easeInElastic},
		{"easeOutElastic", easeOutElastic},
		{"easeInBounce", easeInBounce},
		{"easeOutBounce", easeOutBounce},
	}
	for _, tc := range funcs {
		t.Run(tc.name, func(t *testing.T) {
			v := tc.fn(-0.5)
			if !approxEqual(v, 0) {
				t.Errorf("%s(-0.5) = %f, want 0", tc.name, v)
			}
			v2 := tc.fn(-100)
			if !approxEqual(v2, 0) {
				t.Errorf("%s(-100) = %f, want 0", tc.name, v2)
			}
		})
	}
}

// TestEasingClampAboveOne verifies clamping for t > 1.
func TestEasingClampAboveOne(t *testing.T) {
	funcs := []struct {
		name string
		fn   func(float64) float64
	}{
		{"easeLinear", easeLinear},
		{"easeInQuad", easeInQuad},
		{"easeOutQuad", easeOutQuad},
		{"easeInOutQuad", easeInOutQuad},
		{"easeInCubic", easeInCubic},
		{"easeOutCubic", easeOutCubic},
		{"easeInExpo", easeInExpo},
		{"easeOutExpo", easeOutExpo},
		{"easeInElastic", easeInElastic},
		{"easeOutElastic", easeOutElastic},
		{"easeInBounce", easeInBounce},
		{"easeOutBounce", easeOutBounce},
	}
	for _, tc := range funcs {
		t.Run(tc.name, func(t *testing.T) {
			v := tc.fn(1.5)
			if !approxEqual(v, 1) {
				t.Errorf("%s(1.5) = %f, want 1", tc.name, v)
			}
			v2 := tc.fn(100)
			if !approxEqual(v2, 1) {
				t.Errorf("%s(100) = %f, want 1", tc.name, v2)
			}
		})
	}
}

// TestEasingByNameKnown verifies that EasingByName returns the correct constants.
func TestEasingByNameKnown(t *testing.T) {
	tests := []struct {
		name string
		want EasingFunc
	}{
		{"linear", EaseLinear},
		{"ease-in", EaseIn},
		{"ease-out", EaseOut},
		{"ease-in-out", EaseInOut},
		{"ease-in-quad", EaseInQuad},
		{"ease-out-quad", EaseOutQuad},
		{"ease-in-out-quad", EaseInOutQuad},
		{"ease-in-cubic", EaseInCubic},
		{"ease-out-cubic", EaseOutCubic},
		{"ease-in-out-cubic", EaseInOutCubic},
		{"ease-in-quart", EaseInQuart},
		{"ease-out-quart", EaseOutQuart},
		{"ease-in-out-quart", EaseInOutQuart},
		{"ease-in-quint", EaseInQuint},
		{"ease-out-quint", EaseOutQuint},
		{"ease-in-out-quint", EaseInOutQuint},
		{"ease-in-sine", EaseInSine},
		{"ease-out-sine", EaseOutSine},
		{"ease-in-out-sine", EaseInOutSine},
		{"ease-in-expo", EaseInExpo},
		{"ease-out-expo", EaseOutExpo},
		{"ease-in-out-expo", EaseInOutExpo},
		{"ease-in-circ", EaseInCirc},
		{"ease-out-circ", EaseOutCirc},
		{"ease-in-out-circ", EaseInOutCirc},
		{"ease-in-elastic", EaseInElastic},
		{"ease-out-elastic", EaseOutElastic},
		{"ease-in-out-elastic", EaseInOutElastic},
		{"ease-in-back", EaseInBack},
		{"ease-out-back", EaseOutBack},
		{"ease-in-out-back", EaseInOutBack},
		{"ease-in-bounce", EaseInBounce},
		{"ease-out-bounce", EaseOutBounce},
		{"ease-in-out-bounce", EaseInOutBounce},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := EasingByName(tc.name)
			if got != tc.want {
				t.Errorf("EasingByName(%q) = %d, want %d", tc.name, got, tc.want)
			}
		})
	}
}

// TestEasingByNameUnknown verifies that unknown names return EaseLinear.
func TestEasingByNameUnknown(t *testing.T) {
	unknowns := []string{"cubic-bezier", "spring", "wobble", "", "EASE-IN"}
	for _, name := range unknowns {
		t.Run(name, func(t *testing.T) {
			got := EasingByName(name)
			if got != EaseLinear {
				t.Errorf("EasingByName(%q) = %d, want EaseLinear (%d)", name, got, EaseLinear)
			}
		})
	}
}

// TestApplyEasingDispatch verifies the applyEasing function dispatches correctly for all types.
func TestApplyEasingDispatch(t *testing.T) {
	// Linear should be 0.5 at midpoint
	if v := applyEasing(EaseLinear, 0.5); !approxEqual(v, 0.5) {
		t.Errorf("applyEasing(Linear, 0.5) = %f, want 0.5", v)
	}
	// EaseIn (alias for InQuad) should be 0.25 at midpoint
	if v := applyEasing(EaseIn, 0.5); !approxEqual(v, 0.25) {
		t.Errorf("applyEasing(EaseIn, 0.5) = %f, want 0.25", v)
	}
	// EaseOut (alias for OutQuad) should be 0.75 at midpoint
	if v := applyEasing(EaseOut, 0.5); !approxEqual(v, 0.75) {
		t.Errorf("applyEasing(EaseOut, 0.5) = %f, want 0.75", v)
	}
	// EaseInOut (alias for InOutQuad) should be 0.5 at midpoint
	if v := applyEasing(EaseInOut, 0.5); !approxEqual(v, 0.5) {
		t.Errorf("applyEasing(EaseInOut, 0.5) = %f, want 0.5", v)
	}
	// EaseInCubic should be 0.125 at midpoint
	if v := applyEasing(EaseInCubic, 0.5); !approxEqual(v, 0.125) {
		t.Errorf("applyEasing(EaseInCubic, 0.5) = %f, want 0.125", v)
	}
	// All types should return 0 at t=0 and 1 at t=1
	for ef := EaseLinear; ef <= EaseInOutBounce; ef++ {
		v0 := applyEasing(ef, 0)
		if !approxEqual(v0, 0) {
			t.Errorf("applyEasing(%d, 0) = %f, want 0", ef, v0)
		}
		v1 := applyEasing(ef, 1)
		if !approxEqual(v1, 1) {
			t.Errorf("applyEasing(%d, 1) = %f, want 1", ef, v1)
		}
	}
}

// TestApplyEasingUnknownFallback verifies that an unknown EasingFunc falls back to linear.
func TestApplyEasingUnknownFallback(t *testing.T) {
	unknown := EasingFunc(999)
	if v := applyEasing(unknown, 0.5); !approxEqual(v, 0.5) {
		t.Errorf("applyEasing(unknown, 0.5) = %f, want 0.5 (linear fallback)", v)
	}
}

// TestEaseInProperties checks that ease-in functions are monotonically increasing
// and below the linear curve in the first half.
func TestEaseInProperties(t *testing.T) {
	easeIns := []struct {
		name string
		fn   func(float64) float64
	}{
		{"easeInQuad", easeInQuad},
		{"easeInCubic", easeInCubic},
		{"easeInQuart", easeInQuart},
		{"easeInQuint", easeInQuint},
		{"easeInSine", easeInSine},
		{"easeInExpo", easeInExpo},
		{"easeInCirc", easeInCirc},
	}
	for _, tc := range easeIns {
		t.Run(tc.name, func(t *testing.T) {
			// At t=0.25, ease-in should be below linear
			v := tc.fn(0.25)
			if v >= 0.25 {
				t.Errorf("%s(0.25) = %f, want < 0.25", tc.name, v)
			}
			// Monotonically increasing check
			prev := 0.0
			for i := 1; i <= 100; i++ {
				tt := float64(i) / 100.0
				curr := tc.fn(tt)
				if curr < prev-epsilon {
					t.Errorf("%s not monotonic: f(%f) = %f < f(%f) = %f", tc.name, tt, curr, float64(i-1)/100.0, prev)
				}
				prev = curr
			}
		})
	}
}

// TestEaseOutProperties checks that ease-out functions are monotonically increasing
// and above the linear curve in the first half.
func TestEaseOutProperties(t *testing.T) {
	easeOuts := []struct {
		name string
		fn   func(float64) float64
	}{
		{"easeOutQuad", easeOutQuad},
		{"easeOutCubic", easeOutCubic},
		{"easeOutQuart", easeOutQuart},
		{"easeOutQuint", easeOutQuint},
		{"easeOutSine", easeOutSine},
		{"easeOutExpo", easeOutExpo},
		{"easeOutCirc", easeOutCirc},
	}
	for _, tc := range easeOuts {
		t.Run(tc.name, func(t *testing.T) {
			// At t=0.75, ease-out should be above linear
			v := tc.fn(0.75)
			if v <= 0.75 {
				t.Errorf("%s(0.75) = %f, want > 0.75", tc.name, v)
			}
			// Monotonically increasing check
			prev := 0.0
			for i := 1; i <= 100; i++ {
				tt := float64(i) / 100.0
				curr := tc.fn(tt)
				if curr < prev-epsilon {
					t.Errorf("%s not monotonic: f(%f) = %f < f(%f) = %f", tc.name, tt, curr, float64(i-1)/100.0, prev)
				}
				prev = curr
			}
		})
	}
}

// TestParseExtendedEasingNames verifies the FSS parser accepts extended easing names.
func TestParseExtendedEasingNames(t *testing.T) {
	tests := []struct {
		input string
		want  EasingFunc
	}{
		{"ease-in-cubic", EaseInCubic},
		{"ease-out-cubic", EaseOutCubic},
		{"ease-in-out-cubic", EaseInOutCubic},
		{"ease-in-expo", EaseInExpo},
		{"ease-out-expo", EaseOutExpo},
		{"ease-in-elastic", EaseInElastic},
		{"ease-out-elastic", EaseOutElastic},
		{"ease-in-bounce", EaseInBounce},
		{"ease-out-bounce", EaseOutBounce},
		{"ease-in-out-bounce", EaseInOutBounce},
		{"ease-in-back", EaseInBack},
		{"ease-out-back", EaseOutBack},
		{"ease-in-out-back", EaseInOutBack},
		{"ease-in-sine", EaseInSine},
		{"ease-out-sine", EaseOutSine},
		{"ease-in-circ", EaseInCirc},
		{"ease-out-circ", EaseOutCirc},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got, err := parseEasing(tc.input)
			if err != nil {
				t.Fatalf("parseEasing(%q) error: %v", tc.input, err)
			}
			if got != tc.want {
				t.Errorf("parseEasing(%q) = %d, want %d", tc.input, got, tc.want)
			}
		})
	}
}

// TestParseFSSWithExtendedEasing verifies full FSS parsing with extended easing names.
func TestParseFSSWithExtendedEasing(t *testing.T) {
	tests := []struct {
		fss  string
		want EasingFunc
	}{
		{`Button { transition: background 300ms ease-in-cubic; }`, EaseInCubic},
		{`Button { transition: foreground 200ms ease-out-bounce; }`, EaseOutBounce},
		{`Button { transition: opacity 100ms ease-in-out-elastic; }`, EaseInOutElastic},
		{`Button { transition: background 500ms ease-in-expo; }`, EaseInExpo},
	}
	for _, tc := range tests {
		t.Run(tc.fss, func(t *testing.T) {
			sheet, err := Parse(tc.fss)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			btn := &testNode{typ: "Button"}
			resolved := sheet.Resolve(btn, nil)
			if len(resolved.Transitions) != 1 {
				t.Fatalf("got %d transitions, want 1", len(resolved.Transitions))
			}
			if resolved.Transitions[0].Easing != tc.want {
				t.Errorf("easing = %d, want %d", resolved.Transitions[0].Easing, tc.want)
			}
		})
	}
}

// TestEasingSymmetry checks that InOut functions are symmetric around the midpoint.
func TestEasingSymmetry(t *testing.T) {
	inOutFuncs := []struct {
		name string
		fn   func(float64) float64
	}{
		{"easeInOutQuad", easeInOutQuad},
		{"easeInOutCubic", easeInOutCubic},
		{"easeInOutQuart", easeInOutQuart},
		{"easeInOutQuint", easeInOutQuint},
		{"easeInOutSine", easeInOutSine},
		{"easeInOutExpo", easeInOutExpo},
		{"easeInOutCirc", easeInOutCirc},
	}
	for _, tc := range inOutFuncs {
		t.Run(tc.name, func(t *testing.T) {
			// f(0.5) should be approximately 0.5
			v := tc.fn(0.5)
			if !approxEqual(v, 0.5) {
				t.Errorf("%s(0.5) = %f, want ~0.5", tc.name, v)
			}
			// f(t) + f(1-t) should be approximately 1
			for _, tt := range []float64{0.1, 0.2, 0.3, 0.4} {
				sum := tc.fn(tt) + tc.fn(1-tt)
				if !approxEqual(sum, 1.0) {
					t.Errorf("%s(%f) + %s(%f) = %f, want ~1.0", tc.name, tt, tc.name, 1-tt, sum)
				}
			}
		})
	}
}

// TestEasingFuncTableCompleteness verifies every EasingFunc constant has an entry.
func TestEasingFuncTableCompleteness(t *testing.T) {
	for ef := EaseLinear; ef <= EaseInOutBounce; ef++ {
		if _, ok := easingFuncTable[ef]; !ok {
			t.Errorf("easingFuncTable missing entry for EasingFunc %d", ef)
		}
	}
}

// TestEasingNameMapCompleteness verifies the name map covers all types.
func TestEasingNameMapCompleteness(t *testing.T) {
	// Every value in easingNameMap should have an entry in easingFuncTable.
	for name, ef := range easingNameMap {
		if _, ok := easingFuncTable[ef]; !ok {
			t.Errorf("easingNameMap[%q] = %d, but easingFuncTable has no entry", name, ef)
		}
	}
}
