package style

import (
	"strings"
	"testing"
	"time"
)

func TestParseTransitionProperty(t *testing.T) {
	data := `Button { transition: foreground 200ms ease-in-out; }`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	btn := &testNode{typ: "Button"}
	resolved := sheet.Resolve(btn, nil)
	if len(resolved.Transitions) != 1 {
		t.Fatalf("got %d transitions, want 1", len(resolved.Transitions))
	}
	tr := resolved.Transitions[0]
	if tr.Property != "foreground" {
		t.Errorf("property = %q, want %q", tr.Property, "foreground")
	}
	if tr.Duration != 200*time.Millisecond {
		t.Errorf("duration = %v, want 200ms", tr.Duration)
	}
	if tr.Easing != EaseInOut {
		t.Errorf("easing = %v, want EaseInOut", tr.Easing)
	}
}

func TestParseMultipleTransitions(t *testing.T) {
	data := `Button { transition: foreground 200ms ease-in, background 300ms linear; }`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	btn := &testNode{typ: "Button"}
	resolved := sheet.Resolve(btn, nil)
	if len(resolved.Transitions) != 2 {
		t.Fatalf("got %d transitions, want 2", len(resolved.Transitions))
	}

	tr0 := resolved.Transitions[0]
	if tr0.Property != "foreground" || tr0.Duration != 200*time.Millisecond || tr0.Easing != EaseIn {
		t.Errorf("transition[0] = %+v", tr0)
	}

	tr1 := resolved.Transitions[1]
	if tr1.Property != "background" || tr1.Duration != 300*time.Millisecond || tr1.Easing != EaseLinear {
		t.Errorf("transition[1] = %+v", tr1)
	}
}

func TestParseMultipleTransitionDeclarations(t *testing.T) {
	data := `
Button {
	transition: foreground 200ms ease-in-out;
	transition: background 300ms linear;
}
`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	btn := &testNode{typ: "Button"}
	resolved := sheet.Resolve(btn, nil)
	if len(resolved.Transitions) != 2 {
		t.Fatalf("got %d transitions, want 2", len(resolved.Transitions))
	}
	if resolved.Transitions[0].Property != "foreground" {
		t.Errorf("transition[0].Property = %q, want %q", resolved.Transitions[0].Property, "foreground")
	}
	if resolved.Transitions[1].Property != "background" {
		t.Errorf("transition[1].Property = %q, want %q", resolved.Transitions[1].Property, "background")
	}
}

func TestParseDurationMilliseconds(t *testing.T) {
	d, err := parseDuration("200ms")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if d != 200*time.Millisecond {
		t.Errorf("got %v, want 200ms", d)
	}
}

func TestParseDurationSeconds(t *testing.T) {
	d, err := parseDuration("1.5s")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if d != 1500*time.Millisecond {
		t.Errorf("got %v, want 1.5s", d)
	}
}

func TestParseDurationZeroSeconds(t *testing.T) {
	d, err := parseDuration("0s")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if d != 0 {
		t.Errorf("got %v, want 0", d)
	}
}

func TestParseDurationInvalidUnit(t *testing.T) {
	_, err := parseDuration("200px")
	if err == nil {
		t.Fatal("expected error for invalid duration unit")
	}
}

func TestParseDurationNegative(t *testing.T) {
	_, err := parseDuration("-100ms")
	if err == nil {
		t.Fatal("expected error for negative duration")
	}
}

func TestParseEasingFunctions(t *testing.T) {
	tests := []struct {
		input string
		want  EasingFunc
	}{
		{"linear", EaseLinear},
		{"ease-in", EaseIn},
		{"ease-out", EaseOut},
		{"ease-in-out", EaseInOut},
		{"LINEAR", EaseLinear},
		{"Ease-In", EaseIn},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseEasing(tt.input)
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseEasingInvalid(t *testing.T) {
	_, err := parseEasing("cubic-bezier")
	if err == nil {
		t.Fatal("expected error for invalid easing")
	}
}

func TestParseTransitionAll(t *testing.T) {
	data := `Button { transition: all 200ms ease-in-out; }`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	btn := &testNode{typ: "Button"}
	resolved := sheet.Resolve(btn, nil)
	if len(resolved.Transitions) != 1 {
		t.Fatalf("got %d transitions, want 1", len(resolved.Transitions))
	}
	if resolved.Transitions[0].Property != "all" {
		t.Errorf("property = %q, want %q", resolved.Transitions[0].Property, "all")
	}
}

func TestParseTransitionDefaultEasing(t *testing.T) {
	data := `Button { transition: opacity 100ms; }`
	sheet, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	btn := &testNode{typ: "Button"}
	resolved := sheet.Resolve(btn, nil)
	if len(resolved.Transitions) != 1 {
		t.Fatalf("got %d transitions, want 1", len(resolved.Transitions))
	}
	if resolved.Transitions[0].Easing != EaseLinear {
		t.Errorf("default easing = %v, want EaseLinear", resolved.Transitions[0].Easing)
	}
}

func TestParseTransitionInvalidProperty(t *testing.T) {
	data := `Button { transition: border-radius 200ms; }`
	_, err := Parse(data)
	if err == nil {
		t.Fatal("expected error for unknown transition property")
	}
}

func TestTransitionManagerInterpolatesColors(t *testing.T) {
	tm := NewTransitionManager()

	oldStyle := Style{
		Foreground:  RGB(0, 0, 0),
		Transitions: []Transition{{Property: "foreground", Duration: 100 * time.Millisecond, Easing: EaseLinear}},
	}
	newStyle := Style{
		Foreground:  RGB(100, 200, 50),
		Transitions: []Transition{{Property: "foreground", Duration: 100 * time.Millisecond, Easing: EaseLinear}},
	}

	result := tm.Update(oldStyle, newStyle)

	// The transition just started, so the foreground should be near the old value (black),
	// not the new value yet.
	if result.Foreground == newStyle.Foreground {
		t.Error("expected interpolated foreground, not final value immediately")
	}

	// Verify there is an active transition.
	if !tm.HasActive() {
		t.Error("expected active transitions")
	}
}

func TestTransitionManagerInterpolatesOpacity(t *testing.T) {
	tm := NewTransitionManager()

	old := 0.2
	new := 0.8
	oldStyle := Style{
		Opacity:     &old,
		Transitions: []Transition{{Property: "opacity", Duration: 100 * time.Millisecond, Easing: EaseLinear}},
	}
	newStyle := Style{
		Opacity:     &new,
		Transitions: []Transition{{Property: "opacity", Duration: 100 * time.Millisecond, Easing: EaseLinear}},
	}

	result := tm.Update(oldStyle, newStyle)

	if result.Opacity == nil {
		t.Fatal("expected opacity to be set")
	}
	// Should be near the old value since transition just started.
	if *result.Opacity == new {
		t.Error("expected interpolated opacity, not final value immediately")
	}
	if *result.Opacity < 0.2 || *result.Opacity > 0.8 {
		t.Errorf("opacity %f out of expected range [0.2, 0.8]", *result.Opacity)
	}
}

func TestTransitionManagerCompletesAfterDuration(t *testing.T) {
	tm := NewTransitionManager()

	oldStyle := Style{
		Foreground:  RGB(0, 0, 0),
		Transitions: []Transition{{Property: "foreground", Duration: 50 * time.Millisecond, Easing: EaseLinear}},
	}
	newStyle := Style{
		Foreground:  RGB(255, 255, 255),
		Transitions: []Transition{{Property: "foreground", Duration: 50 * time.Millisecond, Easing: EaseLinear}},
	}

	// Start the transition.
	tm.Update(oldStyle, newStyle)

	// Wait for the transition to complete.
	time.Sleep(60 * time.Millisecond)

	now := time.Now()
	active := tm.Tick(now)
	if active {
		t.Error("expected no active transitions after duration elapsed")
	}
}

func TestTransitionManagerHandlesMultipleSimultaneous(t *testing.T) {
	tm := NewTransitionManager()

	old := 0.3
	new := 0.9
	oldStyle := Style{
		Foreground: RGB(0, 0, 0),
		Background: RGB(255, 255, 255),
		Opacity:    &old,
		Transitions: []Transition{
			{Property: "foreground", Duration: 100 * time.Millisecond, Easing: EaseLinear},
			{Property: "background", Duration: 100 * time.Millisecond, Easing: EaseLinear},
			{Property: "opacity", Duration: 100 * time.Millisecond, Easing: EaseLinear},
		},
	}
	newStyle := Style{
		Foreground: RGB(255, 0, 0),
		Background: RGB(0, 0, 255),
		Opacity:    &new,
		Transitions: []Transition{
			{Property: "foreground", Duration: 100 * time.Millisecond, Easing: EaseLinear},
			{Property: "background", Duration: 100 * time.Millisecond, Easing: EaseLinear},
			{Property: "opacity", Duration: 100 * time.Millisecond, Easing: EaseLinear},
		},
	}

	result := tm.Update(oldStyle, newStyle)

	// All three properties should be actively transitioning.
	if !tm.HasActive() {
		t.Error("expected active transitions")
	}

	// Foreground should not be the final value.
	if result.Foreground == newStyle.Foreground {
		t.Error("expected interpolated foreground")
	}
	// Background should not be the final value.
	if result.Background == newStyle.Background {
		t.Error("expected interpolated background")
	}
	// Opacity should not be the final value.
	if result.Opacity != nil && *result.Opacity == new {
		t.Error("expected interpolated opacity")
	}
}

func TestTransitionManagerNoTransitionConfig(t *testing.T) {
	tm := NewTransitionManager()

	oldStyle := Style{Foreground: RGB(0, 0, 0)}
	newStyle := Style{Foreground: RGB(255, 255, 255)}

	result := tm.Update(oldStyle, newStyle)

	// With no transitions configured, the new value should apply immediately.
	if result.Foreground != newStyle.Foreground {
		t.Errorf("expected immediate value, got %v", result.Foreground)
	}
	if tm.HasActive() {
		t.Error("expected no active transitions")
	}
}

func TestTransitionManagerNilReceiver(t *testing.T) {
	var tm *TransitionManager
	newStyle := Style{Foreground: RGB(255, 0, 0)}
	result := tm.Update(Style{}, newStyle)
	if result.Foreground != newStyle.Foreground {
		t.Error("nil TransitionManager should pass through newStyle")
	}
	if tm.Tick(time.Now()) {
		t.Error("nil TransitionManager Tick should return false")
	}
	if tm.HasActive() {
		t.Error("nil TransitionManager HasActive should return false")
	}
}

func TestTransitionManagerAllShorthand(t *testing.T) {
	tm := NewTransitionManager()

	old := 0.5
	new := 1.0
	oldStyle := Style{
		Foreground:  RGB(0, 0, 0),
		Background:  RGB(255, 255, 255),
		Opacity:     &old,
		Transitions: []Transition{{Property: "all", Duration: 100 * time.Millisecond, Easing: EaseLinear}},
	}
	newStyle := Style{
		Foreground:  RGB(255, 0, 0),
		Background:  RGB(0, 0, 255),
		Opacity:     &new,
		Transitions: []Transition{{Property: "all", Duration: 100 * time.Millisecond, Easing: EaseLinear}},
	}

	result := tm.Update(oldStyle, newStyle)

	// All should be transitioning.
	if result.Foreground == newStyle.Foreground {
		t.Error("expected interpolated foreground with 'all' shorthand")
	}
	if result.Background == newStyle.Background {
		t.Error("expected interpolated background with 'all' shorthand")
	}
}

func TestTransitionManagerSpacingInterpolation(t *testing.T) {
	tm := NewTransitionManager()

	oldStyle := Style{
		Padding:     Pad(0),
		Transitions: []Transition{{Property: "padding", Duration: 100 * time.Millisecond, Easing: EaseLinear}},
	}
	newStyle := Style{
		Padding:     Pad(10),
		Transitions: []Transition{{Property: "padding", Duration: 100 * time.Millisecond, Easing: EaseLinear}},
	}

	result := tm.Update(oldStyle, newStyle)

	if result.Padding == nil {
		t.Fatal("expected padding to be set")
	}
	// Should be between 0 and 10.
	if result.Padding.Top == 10 {
		t.Error("expected interpolated padding, not final value")
	}
}

func TestTransitionIsZero(t *testing.T) {
	s := Style{}
	if !s.IsZero() {
		t.Error("empty style should be zero")
	}
	s.Transitions = []Transition{{Property: "foreground", Duration: 200 * time.Millisecond}}
	if s.IsZero() {
		t.Error("style with transitions should not be zero")
	}
}

func TestTransitionMerge(t *testing.T) {
	base := Style{}
	override := Style{
		Transitions: []Transition{{Property: "foreground", Duration: 200 * time.Millisecond}},
	}
	merged := base.Merge(override)
	if len(merged.Transitions) != 1 {
		t.Errorf("merged transitions = %d, want 1", len(merged.Transitions))
	}
}

func TestApplyEasing(t *testing.T) {
	// EaseLinear
	if v := applyEasing(EaseLinear, 0.5); v != 0.5 {
		t.Errorf("linear(0.5) = %f, want 0.5", v)
	}
	// EaseIn should be less than linear at 0.5
	if v := applyEasing(EaseIn, 0.5); v >= 0.5 {
		t.Errorf("ease-in(0.5) = %f, want < 0.5", v)
	}
	// EaseOut should be greater than linear at 0.5
	if v := applyEasing(EaseOut, 0.5); v <= 0.5 {
		t.Errorf("ease-out(0.5) = %f, want > 0.5", v)
	}
	// Boundaries
	if v := applyEasing(EaseInOut, 0); v != 0 {
		t.Errorf("ease-in-out(0) = %f, want 0", v)
	}
	if v := applyEasing(EaseInOut, 1); v != 1 {
		t.Errorf("ease-in-out(1) = %f, want 1", v)
	}
}

func TestLerpColor(t *testing.T) {
	a := RGB(0, 0, 0)
	b := RGB(100, 200, 50)

	// At t=0, should be a.
	c := lerpColor(a, b, 0)
	if c != a {
		t.Errorf("lerpColor(t=0) = %v, want %v", c, a)
	}

	// At t=1, should be b.
	c = lerpColor(a, b, 1)
	if c != b {
		t.Errorf("lerpColor(t=1) = %v, want %v", c, b)
	}

	// At t=0.5, should be midpoint.
	c = lerpColor(a, b, 0.5)
	expected := RGB(50, 100, 25)
	if c != expected {
		t.Errorf("lerpColor(t=0.5) = %v, want %v", c, expected)
	}
}

func TestLerpFloat(t *testing.T) {
	if v := lerpFloat(0, 1, 0.5); v != 0.5 {
		t.Errorf("lerpFloat(0,1,0.5) = %f, want 0.5", v)
	}
	if v := lerpFloat(10, 20, 0); v != 10 {
		t.Errorf("lerpFloat(10,20,0) = %f, want 10", v)
	}
	if v := lerpFloat(10, 20, 1); v != 20 {
		t.Errorf("lerpFloat(10,20,1) = %f, want 20", v)
	}
}

func TestLerpSpacing(t *testing.T) {
	a := Spacing{Top: 0, Right: 0, Bottom: 0, Left: 0}
	b := Spacing{Top: 10, Right: 20, Bottom: 30, Left: 40}

	s := lerpSpacing(a, b, 0.5)
	if s.Top != 5 || s.Right != 10 || s.Bottom != 15 || s.Left != 20 {
		t.Errorf("lerpSpacing(0.5) = %+v, want {5 10 15 20}", s)
	}
}

func TestColorToRGB(t *testing.T) {
	// RGB mode
	r, g, b := colorToRGB(RGB(100, 150, 200))
	if r != 100 || g != 150 || b != 200 {
		t.Errorf("RGB = (%d,%d,%d), want (100,150,200)", r, g, b)
	}

	// 16-color mode (red = index 1)
	r, g, b = colorToRGB(ColorRed)
	if r != 170 || g != 0 || b != 0 {
		t.Errorf("16-color red = (%d,%d,%d), want (170,0,0)", r, g, b)
	}

	// 256-color mode
	r, g, b = colorToRGB(Color256(232)) // first greyscale
	if r != 8 || g != 8 || b != 8 {
		t.Errorf("256-color 232 = (%d,%d,%d), want (8,8,8)", r, g, b)
	}
}

func TestParseTransitionPropertyAliases(t *testing.T) {
	tests := []struct {
		input    string
		wantProp string
	}{
		{`Button { transition: color 200ms; }`, "color"},
		{`Button { transition: fg 200ms; }`, "fg"},
		{`Button { transition: bg 200ms; }`, "bg"},
	}
	for _, tt := range tests {
		t.Run(tt.wantProp, func(t *testing.T) {
			sheet, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			btn := &testNode{typ: "Button"}
			resolved := sheet.Resolve(btn, nil)
			if len(resolved.Transitions) != 1 {
				t.Fatalf("got %d transitions, want 1", len(resolved.Transitions))
			}
			if resolved.Transitions[0].Property != tt.wantProp {
				t.Errorf("property = %q, want %q", resolved.Transitions[0].Property, tt.wantProp)
			}
		})
	}
}

func TestTransitionManagerSameValueNoTransition(t *testing.T) {
	tm := NewTransitionManager()

	style := Style{
		Foreground:  RGB(100, 100, 100),
		Transitions: []Transition{{Property: "foreground", Duration: 100 * time.Millisecond, Easing: EaseLinear}},
	}

	// Same old and new -- no transition needed.
	tm.Update(style, style)

	if tm.HasActive() {
		t.Error("no transition should start when values are equal")
	}
}

func TestParseTransitionEmpty(t *testing.T) {
	data := `Button { transition: ; }`
	_, err := Parse(data)
	if err == nil {
		t.Fatal("expected error for empty transition value")
	}
	if !strings.Contains(err.Error(), "empty transition") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestParseTransitionMissingDuration(t *testing.T) {
	data := `Button { transition: foreground; }`
	_, err := Parse(data)
	if err == nil {
		t.Fatal("expected error for missing duration")
	}
}
