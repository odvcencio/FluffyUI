//go:build !js

package widgettest

import (
	"fmt"
	"testing"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/runtime"
)

// mockWidget is a minimal widget with configurable accessibility properties.
type mockWidget struct {
	accessibility.Base
	bounds  runtime.Rect
	width   int
	height  int
	panics  bool
}

func (m *mockWidget) Measure(c runtime.Constraints) runtime.Size {
	w := m.width
	if w > c.MaxWidth {
		w = c.MaxWidth
	}
	h := m.height
	if h > c.MaxHeight {
		h = c.MaxHeight
	}
	return runtime.Size{Width: w, Height: h}
}

func (m *mockWidget) Layout(bounds runtime.Rect) {
	m.bounds = bounds
}

func (m *mockWidget) Render(ctx runtime.RenderContext) {
	if m.panics {
		panic("intentional panic in render")
	}
}

func (m *mockWidget) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}

func (m *mockWidget) Bounds() runtime.Rect {
	return m.bounds
}

// focusableMockWidget is a mock widget that also implements runtime.Focusable.
type focusableMockWidget struct {
	mockWidget
	canFocus bool
	focused  bool
}

func (f *focusableMockWidget) CanFocus() bool  { return f.canFocus }
func (f *focusableMockWidget) Focus()          { f.focused = true }
func (f *focusableMockWidget) Blur()           { f.focused = false }
func (f *focusableMockWidget) IsFocused() bool { return f.focused }

// bindableMockWidget is a mock widget that also implements runtime.Bindable.
type bindableMockWidget struct {
	mockWidget
	bound bool
}

func (b *bindableMockWidget) Bind(services runtime.Services) {
	b.bound = true
}

// bareWidget is a widget that does NOT implement accessibility.Accessible.
type bareWidget struct {
	bounds runtime.Rect
}

func (b *bareWidget) Measure(c runtime.Constraints) runtime.Size {
	return runtime.Size{Width: 5, Height: 1}
}

func (b *bareWidget) Layout(bounds runtime.Rect) {
	b.bounds = bounds
}

func (b *bareWidget) Render(ctx runtime.RenderContext) {}

func (b *bareWidget) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}

// --- Tests ---

func TestAssertRole_Pass(t *testing.T) {
	w := &mockWidget{width: 10, height: 1}
	w.Role = accessibility.RoleButton
	AssertRole(t, w, "button")
}

func TestAssertRole_Mismatch(t *testing.T) {
	w := &mockWidget{width: 10, height: 1}
	w.Role = accessibility.RoleCheckbox

	ft := &fakeTB{}
	AssertRole(ft, w, "button")
	if !ft.failed {
		t.Error("expected AssertRole to fail on role mismatch")
	}
}

func TestAssertRole_NotAccessible(t *testing.T) {
	w := &bareWidget{}

	ft := &fakeTB{}
	AssertRole(ft, w, "button")
	if !ft.failed {
		t.Error("expected AssertRole to fail when widget is not Accessible")
	}
}

func TestAssertLabel_Pass(t *testing.T) {
	w := &mockWidget{width: 10, height: 1}
	w.Label = "Save"
	AssertLabel(t, w, "Save")
}

func TestAssertLabel_Mismatch(t *testing.T) {
	w := &mockWidget{width: 10, height: 1}
	w.Label = "Save"

	ft := &fakeTB{}
	AssertLabel(ft, w, "Cancel")
	if !ft.failed {
		t.Error("expected AssertLabel to fail on label mismatch")
	}
}

func TestAssertLabel_NotAccessible(t *testing.T) {
	w := &bareWidget{}

	ft := &fakeTB{}
	AssertLabel(ft, w, "anything")
	if !ft.failed {
		t.Error("expected AssertLabel to fail when widget is not Accessible")
	}
}

func TestAssertFocusable_Pass(t *testing.T) {
	w := &focusableMockWidget{canFocus: true}
	AssertFocusable(t, w)
}

func TestAssertFocusable_NotFocusable(t *testing.T) {
	w := &mockWidget{width: 10, height: 1}

	ft := &fakeTB{}
	AssertFocusable(ft, w)
	if !ft.failed {
		t.Error("expected AssertFocusable to fail when widget is not Focusable")
	}
}

func TestAssertFocusable_CanFocusFalse(t *testing.T) {
	w := &focusableMockWidget{canFocus: false}

	ft := &fakeTB{}
	AssertFocusable(ft, w)
	if !ft.failed {
		t.Error("expected AssertFocusable to fail when CanFocus returns false")
	}
}

func TestAssertBindable_Pass(t *testing.T) {
	w := &bindableMockWidget{}
	AssertBindable(t, w)
}

func TestAssertBindable_NotBindable(t *testing.T) {
	w := &mockWidget{width: 10, height: 1}

	ft := &fakeTB{}
	AssertBindable(ft, w)
	if !ft.failed {
		t.Error("expected AssertBindable to fail when widget is not Bindable")
	}
}

func TestAssertMeasure_Pass(t *testing.T) {
	w := &mockWidget{width: 20, height: 5}
	AssertMeasure(t, w, 10, 3)
}

func TestAssertMeasure_TooSmallWidth(t *testing.T) {
	w := &mockWidget{width: 5, height: 5}

	ft := &fakeTB{}
	AssertMeasure(ft, w, 10, 3)
	if !ft.failed {
		t.Error("expected AssertMeasure to fail when width is too small")
	}
}

func TestAssertMeasure_TooSmallHeight(t *testing.T) {
	w := &mockWidget{width: 20, height: 1}

	ft := &fakeTB{}
	AssertMeasure(ft, w, 10, 3)
	if !ft.failed {
		t.Error("expected AssertMeasure to fail when height is too small")
	}
}

func TestAssertRenders_Pass(t *testing.T) {
	w := &mockWidget{width: 10, height: 1}
	AssertRenders(t, w, 10, 1)
}

func TestAssertRenders_Panics(t *testing.T) {
	w := &mockWidget{width: 10, height: 1, panics: true}

	ft := &fakeTB{}
	AssertRenders(ft, w, 10, 1)
	if !ft.failed {
		t.Error("expected AssertRenders to fail when widget panics during render")
	}
}

func TestAssertValue_Pass(t *testing.T) {
	w := &mockWidget{width: 10, height: 1}
	w.Value = &accessibility.ValueInfo{Text: "50%", Current: 50, Min: 0, Max: 100}
	AssertValue(t, w, "50%")
}

func TestAssertValue_Mismatch(t *testing.T) {
	w := &mockWidget{width: 10, height: 1}
	w.Value = &accessibility.ValueInfo{Text: "50%"}

	ft := &fakeTB{}
	AssertValue(ft, w, "75%")
	if !ft.failed {
		t.Error("expected AssertValue to fail on value text mismatch")
	}
}

func TestAssertValue_NilValue(t *testing.T) {
	w := &mockWidget{width: 10, height: 1}

	ft := &fakeTB{}
	AssertValue(ft, w, "anything")
	if !ft.failed {
		t.Error("expected AssertValue to fail when value is nil")
	}
}

func TestAssertValue_NotAccessible(t *testing.T) {
	w := &bareWidget{}

	ft := &fakeTB{}
	AssertValue(ft, w, "anything")
	if !ft.failed {
		t.Error("expected AssertValue to fail when widget is not Accessible")
	}
}

func TestAssertState_Disabled(t *testing.T) {
	w := &mockWidget{width: 10, height: 1}
	w.State = accessibility.StateSet{Disabled: true}
	AssertState(t, w, func(s accessibility.StateSet) bool {
		return s.Disabled
	}, "Disabled == true")
}

func TestAssertState_CheckFails(t *testing.T) {
	w := &mockWidget{width: 10, height: 1}
	w.State = accessibility.StateSet{Disabled: false}

	ft := &fakeTB{}
	AssertState(ft, w, func(s accessibility.StateSet) bool {
		return s.Disabled
	}, "Disabled == true")
	if !ft.failed {
		t.Error("expected AssertState to fail when state check returns false")
	}
}

func TestAssertState_Expanded(t *testing.T) {
	w := &mockWidget{width: 10, height: 1}
	expanded := true
	w.State = accessibility.StateSet{Expanded: &expanded}
	AssertState(t, w, func(s accessibility.StateSet) bool {
		return s.Expanded != nil && *s.Expanded
	}, "Expanded == true")
}

func TestAssertState_Checked(t *testing.T) {
	w := &mockWidget{width: 10, height: 1}
	checked := true
	w.State = accessibility.StateSet{Checked: &checked}
	AssertState(t, w, func(s accessibility.StateSet) bool {
		return s.Checked != nil && *s.Checked
	}, "Checked == true")
}

func TestAssertState_NotAccessible(t *testing.T) {
	w := &bareWidget{}

	ft := &fakeTB{}
	AssertState(ft, w, func(s accessibility.StateSet) bool {
		return s.Disabled
	}, "Disabled == true")
	if !ft.failed {
		t.Error("expected AssertState to fail when widget is not Accessible")
	}
}

// --- fakeTB captures test failures without aborting ---

type fakeTB struct {
	testing.TB
	failed   bool
	messages []string
}

func (f *fakeTB) Helper() {}

func (f *fakeTB) Errorf(format string, args ...interface{}) {
	f.failed = true
	f.messages = append(f.messages, fmt.Sprintf(format, args...))
}

func (f *fakeTB) Fatalf(format string, args ...interface{}) {
	f.failed = true
	f.messages = append(f.messages, fmt.Sprintf(format, args...))
}

func (f *fakeTB) Log(args ...interface{}) {}

func (f *fakeTB) Logf(format string, args ...interface{}) {}
