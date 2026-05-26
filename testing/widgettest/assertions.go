//go:build !js

package widgettest

import (
	"fmt"
	"testing"

	"m31labs.dev/fluffyui/accessibility"
	"m31labs.dev/fluffyui/runtime"
)

// AssertRole checks that a widget has the expected ARIA role.
// The widget must implement accessibility.Accessible.
func AssertRole(t testing.TB, w runtime.Widget, expected string) {
	t.Helper()
	acc, ok := w.(accessibility.Accessible)
	if !ok {
		t.Errorf("widget %T does not implement accessibility.Accessible", w)
		return
	}
	got := string(acc.AccessibleRole())
	if got != expected {
		t.Errorf("role mismatch: got %q, want %q", got, expected)
	}
}

// AssertLabel checks that a widget's accessibility label matches expected.
// The widget must implement accessibility.Accessible.
func AssertLabel(t testing.TB, w runtime.Widget, expected string) {
	t.Helper()
	acc, ok := w.(accessibility.Accessible)
	if !ok {
		t.Errorf("widget %T does not implement accessibility.Accessible", w)
		return
	}
	got := acc.AccessibleLabel()
	if got != expected {
		t.Errorf("label mismatch: got %q, want %q", got, expected)
	}
}

// AssertFocusable checks that a widget implements runtime.Focusable
// and reports CanFocus() == true.
func AssertFocusable(t testing.TB, w runtime.Widget) {
	t.Helper()
	f, ok := w.(runtime.Focusable)
	if !ok {
		t.Errorf("widget %T does not implement runtime.Focusable", w)
		return
	}
	if !f.CanFocus() {
		t.Errorf("widget %T implements Focusable but CanFocus() returned false", w)
	}
}

// AssertBindable checks that a widget implements runtime.Bindable.
func AssertBindable(t testing.TB, w runtime.Widget) {
	t.Helper()
	if _, ok := w.(runtime.Bindable); !ok {
		t.Errorf("widget %T does not implement runtime.Bindable", w)
	}
}

// AssertMeasure checks that a widget measures to at least the given minimum
// width and height. The widget is measured with unconstrained limits.
func AssertMeasure(t testing.TB, w runtime.Widget, minWidth, minHeight int) {
	t.Helper()
	size := w.Measure(runtime.Constraints{MaxWidth: 1000, MaxHeight: 1000})
	if size.Width < minWidth {
		t.Errorf("measure width too small: got %d, want >= %d", size.Width, minWidth)
	}
	if size.Height < minHeight {
		t.Errorf("measure height too small: got %d, want >= %d", size.Height, minHeight)
	}
}

// AssertRenders checks that a widget renders without panicking.
// It performs a full measure-layout-render cycle at the given dimensions.
func AssertRenders(t testing.TB, w runtime.Widget, width, height int) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("widget %T panicked during render: %v", w, r)
		}
	}()
	buf := runtime.NewBuffer(width, height)
	w.Measure(runtime.Constraints{MaxWidth: width, MaxHeight: height})
	w.Layout(runtime.Rect{X: 0, Y: 0, Width: width, Height: height})
	w.Render(runtime.RenderContext{Buffer: buf})
}

// AssertValue checks that a widget's accessibility value text matches expected.
// The widget must implement accessibility.Accessible and have a non-nil value.
func AssertValue(t testing.TB, w runtime.Widget, expected string) {
	t.Helper()
	acc, ok := w.(accessibility.Accessible)
	if !ok {
		t.Errorf("widget %T does not implement accessibility.Accessible", w)
		return
	}
	v := acc.AccessibleValue()
	if v == nil {
		t.Errorf("widget %T has nil AccessibleValue(), want text %q", w, expected)
		return
	}
	if v.Text != expected {
		t.Errorf("value text mismatch: got %q, want %q", v.Text, expected)
	}
}

// AssertState checks that a specific accessibility state condition holds.
// The check function receives the widget's StateSet and should return true
// if the condition is met. The desc parameter describes the expected state
// for the error message (e.g. "Disabled == true").
func AssertState(t testing.TB, w runtime.Widget, check func(accessibility.StateSet) bool, desc string) {
	t.Helper()
	acc, ok := w.(accessibility.Accessible)
	if !ok {
		t.Errorf("widget %T does not implement accessibility.Accessible", w)
		return
	}
	state := acc.AccessibleState()
	if !check(state) {
		t.Errorf("state check failed: expected %s; state: %s", desc, fmt.Sprintf("%+v", state))
	}
}
