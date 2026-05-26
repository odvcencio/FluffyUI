package style

import (
	"math"
	"time"

	"m31labs.dev/fluffyui/compositor"
)

// TransitionManager tracks active transitions and interpolates between old
// and new style values. Each widget should own one TransitionManager.
type TransitionManager struct {
	active map[string]*activeTransition // property -> transition state
}

type activeTransition struct {
	from     any           // starting value
	to       any           // target value
	started  time.Time
	duration time.Duration
	easing   EasingFunc
}

// NewTransitionManager creates a new TransitionManager.
func NewTransitionManager() *TransitionManager {
	return &TransitionManager{
		active: make(map[string]*activeTransition),
	}
}

// Update compares oldStyle and newStyle, starts transitions for properties
// that changed (according to the transition declarations on newStyle), and
// returns the interpolated style for the current moment.
func (tm *TransitionManager) Update(oldStyle, newStyle Style) Style {
	if tm == nil {
		return newStyle
	}
	now := time.Now()

	// Build a map of transition configs from the new style.
	tconfigs := transitionConfigs(newStyle.Transitions)

	// Check each transitionable property for changes and start transitions.
	tm.checkColor("foreground", oldStyle.Foreground, newStyle.Foreground, tconfigs, now)
	tm.checkColor("background", oldStyle.Background, newStyle.Background, tconfigs, now)
	tm.checkOpacity("opacity", oldStyle.Opacity, newStyle.Opacity, tconfigs, now)
	tm.checkSpacing("padding", oldStyle.Padding, newStyle.Padding, tconfigs, now)
	tm.checkSpacing("margin", oldStyle.Margin, newStyle.Margin, tconfigs, now)

	// Apply interpolated values to a copy of newStyle.
	return tm.apply(newStyle, now)
}

// Tick advances all active transitions and reports whether any are still running.
func (tm *TransitionManager) Tick(now time.Time) bool {
	if tm == nil {
		return false
	}
	active := false
	for prop, at := range tm.active {
		elapsed := now.Sub(at.started)
		if elapsed >= at.duration {
			delete(tm.active, prop)
		} else {
			active = true
		}
	}
	return active
}

// HasActive reports whether any transitions are currently running.
func (tm *TransitionManager) HasActive() bool {
	if tm == nil {
		return false
	}
	return len(tm.active) > 0
}

// transitionConfigs builds a lookup map from property name to Transition config.
// The "all" shorthand expands to every transitionable property.
func transitionConfigs(transitions []Transition) map[string]Transition {
	if len(transitions) == 0 {
		return nil
	}
	m := make(map[string]Transition, len(transitions))
	for _, t := range transitions {
		if t.Property == "all" {
			for _, prop := range allTransitionableProperties {
				if _, exists := m[prop]; !exists {
					m[prop] = Transition{Property: prop, Duration: t.Duration, Easing: t.Easing}
				}
			}
		} else {
			prop := canonicalProperty(t.Property)
			m[prop] = t
		}
	}
	return m
}

var allTransitionableProperties = []string{
	"foreground", "background", "opacity", "padding", "margin",
}

// canonicalProperty normalises alias property names.
func canonicalProperty(name string) string {
	switch name {
	case "color", "fg":
		return "foreground"
	case "bg":
		return "background"
	default:
		return name
	}
}

// --- property-specific change detection and transition start ---

func (tm *TransitionManager) checkColor(prop string, old, new Color, configs map[string]Transition, now time.Time) {
	if old == new {
		return
	}
	cfg, ok := configs[prop]
	if !ok || cfg.Duration <= 0 {
		// No transition configured; remove any active transition so the new value applies immediately.
		delete(tm.active, prop)
		return
	}
	// If there is an active transition to the same target, keep it.
	if at, exists := tm.active[prop]; exists {
		if colorsEqual(at.to, new) {
			return
		}
		// Retarget: start from current interpolated value.
		old = tm.interpolatedColor(at, now)
	}
	tm.active[prop] = &activeTransition{
		from:     old,
		to:       new,
		started:  now,
		duration: cfg.Duration,
		easing:   cfg.Easing,
	}
}

func (tm *TransitionManager) checkOpacity(prop string, old, new *float64, configs map[string]Transition, now time.Time) {
	oldVal := opacityValue(old)
	newVal := opacityValue(new)
	if oldVal == newVal {
		return
	}
	cfg, ok := configs[prop]
	if !ok || cfg.Duration <= 0 {
		delete(tm.active, prop)
		return
	}
	if at, exists := tm.active[prop]; exists {
		if floatsEqual(at.to, newVal) {
			return
		}
		oldVal = tm.interpolatedFloat(at, now)
	}
	tm.active[prop] = &activeTransition{
		from:     oldVal,
		to:       newVal,
		started:  now,
		duration: cfg.Duration,
		easing:   cfg.Easing,
	}
}

func (tm *TransitionManager) checkSpacing(prop string, old, new *Spacing, configs map[string]Transition, now time.Time) {
	oldS := spacingOrZero(old)
	newS := spacingOrZero(new)
	if oldS == newS {
		return
	}
	cfg, ok := configs[prop]
	if !ok || cfg.Duration <= 0 {
		delete(tm.active, prop)
		return
	}
	if at, exists := tm.active[prop]; exists {
		if spacingsEqual(at.to, newS) {
			return
		}
		oldS = tm.interpolatedSpacing(at, now)
	}
	tm.active[prop] = &activeTransition{
		from:     oldS,
		to:       newS,
		started:  now,
		duration: cfg.Duration,
		easing:   cfg.Easing,
	}
}

// --- interpolation helpers ---

func (tm *TransitionManager) apply(s Style, now time.Time) Style {
	for prop, at := range tm.active {
		t := tm.progress(at, now)
		switch prop {
		case "foreground":
			s.Foreground = lerpColor(at.from.(Color), at.to.(Color), t)
		case "background":
			s.Background = lerpColor(at.from.(Color), at.to.(Color), t)
		case "opacity":
			v := lerpFloat(at.from.(float64), at.to.(float64), t)
			s.Opacity = &v
		case "padding":
			sp := lerpSpacing(at.from.(Spacing), at.to.(Spacing), t)
			s.Padding = &sp
		case "margin":
			sp := lerpSpacing(at.from.(Spacing), at.to.(Spacing), t)
			s.Margin = &sp
		}
	}
	return s
}

func (tm *TransitionManager) progress(at *activeTransition, now time.Time) float64 {
	elapsed := now.Sub(at.started)
	if elapsed <= 0 {
		return 0
	}
	if elapsed >= at.duration {
		return 1
	}
	t := float64(elapsed) / float64(at.duration)
	return applyEasing(at.easing, t)
}

func (tm *TransitionManager) interpolatedColor(at *activeTransition, now time.Time) Color {
	t := tm.progress(at, now)
	return lerpColor(at.from.(Color), at.to.(Color), t)
}

func (tm *TransitionManager) interpolatedFloat(at *activeTransition, now time.Time) float64 {
	t := tm.progress(at, now)
	return lerpFloat(at.from.(float64), at.to.(float64), t)
}

func (tm *TransitionManager) interpolatedSpacing(at *activeTransition, now time.Time) Spacing {
	t := tm.progress(at, now)
	return lerpSpacing(at.from.(Spacing), at.to.(Spacing), t)
}

// --- easing application ---

func applyEasing(fn EasingFunc, t float64) float64 {
	if impl, ok := easingFuncTable[fn]; ok {
		return impl(t)
	}
	return t // fallback to linear
}

// --- lerp functions ---

func lerpColor(a, b Color, t float64) Color {
	ar, ag, ab := colorToRGB(a)
	br, bg, bb := colorToRGB(b)
	return compositor.RGB(
		lerpUint8(ar, br, t),
		lerpUint8(ag, bg, t),
		lerpUint8(ab, bb, t),
	)
}

func lerpFloat(a, b, t float64) float64 {
	return a + (b-a)*t
}

func lerpSpacing(a, b Spacing, t float64) Spacing {
	return Spacing{
		Top:    lerpInt(a.Top, b.Top, t),
		Right:  lerpInt(a.Right, b.Right, t),
		Bottom: lerpInt(a.Bottom, b.Bottom, t),
		Left:   lerpInt(a.Left, b.Left, t),
	}
}

func lerpUint8(a, b uint8, t float64) uint8 {
	v := float64(a) + float64(int(b)-int(a))*t
	return uint8(math.Round(clamp(v, 0, 255)))
}

func lerpInt(a, b int, t float64) int {
	return int(math.Round(float64(a) + float64(b-a)*t))
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// colorToRGB converts any Color to R, G, B components.
// For 16-color and 256-color modes, a best-effort mapping is used.
func colorToRGB(c Color) (uint8, uint8, uint8) {
	switch c.Mode {
	case compositor.ColorModeRGB:
		r := uint8((c.Value >> 16) & 0xFF)
		g := uint8((c.Value >> 8) & 0xFF)
		b := uint8(c.Value & 0xFF)
		return r, g, b
	case compositor.ColorMode16:
		return ansi16ToRGB(c.Value)
	case compositor.ColorMode256:
		return ansi256ToRGB(uint8(c.Value))
	default:
		return 0, 0, 0
	}
}

// ansi16ToRGB provides approximate RGB values for the standard 16 ANSI colors.
func ansi16ToRGB(index uint32) (uint8, uint8, uint8) {
	// Standard terminal palette approximation.
	palette := [16][3]uint8{
		{0, 0, 0},       // 0  black
		{170, 0, 0},     // 1  red
		{0, 170, 0},     // 2  green
		{170, 170, 0},   // 3  yellow
		{0, 0, 170},     // 4  blue
		{170, 0, 170},   // 5  magenta
		{0, 170, 170},   // 6  cyan
		{170, 170, 170}, // 7  white
		{85, 85, 85},    // 8  bright black
		{255, 85, 85},   // 9  bright red
		{85, 255, 85},   // 10 bright green
		{255, 255, 85},  // 11 bright yellow
		{85, 85, 255},   // 12 bright blue
		{255, 85, 255},  // 13 bright magenta
		{85, 255, 255},  // 14 bright cyan
		{255, 255, 255}, // 15 bright white
	}
	if index > 15 {
		index = 0
	}
	return palette[index][0], palette[index][1], palette[index][2]
}

// ansi256ToRGB converts a 256-palette index to approximate RGB.
func ansi256ToRGB(index uint8) (uint8, uint8, uint8) {
	if index < 16 {
		return ansi16ToRGB(uint32(index))
	}
	if index >= 232 {
		// Greyscale ramp: 232-255 -> 8, 18, 28, ..., 238
		v := uint8(8 + int(index-232)*10)
		return v, v, v
	}
	// 6x6x6 color cube: 16-231
	idx := int(index) - 16
	b := uint8((idx % 6) * 51)
	g := uint8(((idx / 6) % 6) * 51)
	r := uint8((idx / 36) * 51)
	return r, g, b
}

// --- comparison helpers ---

func colorsEqual(a any, b Color) bool {
	ac, ok := a.(Color)
	if !ok {
		return false
	}
	return ac == b
}

func floatsEqual(a any, b float64) bool {
	af, ok := a.(float64)
	if !ok {
		return false
	}
	return af == b
}

func spacingsEqual(a any, b Spacing) bool {
	as, ok := a.(Spacing)
	if !ok {
		return false
	}
	return as == b
}

func opacityValue(p *float64) float64 {
	if p == nil {
		return 1.0
	}
	return *p
}

func spacingOrZero(p *Spacing) Spacing {
	if p == nil {
		return Spacing{}
	}
	return *p
}
