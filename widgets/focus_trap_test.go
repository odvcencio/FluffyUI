package widgets

import (
	"testing"

	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/terminal"
)

// ftTestFocusable is a simple focusable widget for focus trap/skip nav testing.
type ftTestFocusable struct {
	FocusableBase
	name string
}

func newFTTestFocusable(name string) *ftTestFocusable {
	w := &ftTestFocusable{name: name}
	w.Base.Label = name
	return w
}

func (w *ftTestFocusable) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.Constrain(runtime.Size{Width: 10, Height: 1})
}

func (w *ftTestFocusable) Render(ctx runtime.RenderContext) {
	bounds := w.ContentBounds()
	if bounds.Width > 0 && bounds.Height > 0 {
		ctx.Buffer.SetString(bounds.X, bounds.Y, w.name, ctx.Buffer.Get(0, 0).Style)
	}
}

// ftTestContainer holds multiple children for focus trap testing.
type ftTestContainer struct {
	Base
	children []runtime.Widget
}

func newFTTestContainer(children ...runtime.Widget) *ftTestContainer {
	return &ftTestContainer{children: children}
}

func (c *ftTestContainer) ChildWidgets() []runtime.Widget {
	return c.children
}

func (c *ftTestContainer) Measure(constraints runtime.Constraints) runtime.Size {
	h := 0
	w := 0
	for _, child := range c.children {
		size := child.Measure(constraints)
		h += size.Height
		if size.Width > w {
			w = size.Width
		}
	}
	return constraints.Constrain(runtime.Size{Width: w, Height: h})
}

func (c *ftTestContainer) Layout(bounds runtime.Rect) {
	c.Base.Layout(bounds)
	y := bounds.Y
	for _, child := range c.children {
		size := child.Measure(runtime.Constraints{MaxWidth: bounds.Width, MaxHeight: bounds.Height})
		child.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: size.Height})
		y += size.Height
	}
}

func (c *ftTestContainer) Render(ctx runtime.RenderContext) {
	for _, child := range c.children {
		runtime.RenderChild(ctx, child)
	}
}

func TestFocusTrap_NewFocusTrap(t *testing.T) {
	child := newFTTestFocusable("btn1")
	ft := NewFocusTrap(child)

	if ft == nil {
		t.Fatal("expected non-nil FocusTrap")
	}
	if !ft.Active() {
		t.Error("expected FocusTrap to be active by default")
	}
	if ft.Child() != child {
		t.Error("expected child to match")
	}
	if ft.Base.Role != "dialog" {
		t.Errorf("expected role dialog, got %s", string(ft.Base.Role))
	}
}

func TestFocusTrap_SetActive(t *testing.T) {
	ft := NewFocusTrap(newFTTestFocusable("btn"))
	ft.SetActive(false)
	if ft.Active() {
		t.Error("expected FocusTrap to be inactive after SetActive(false)")
	}
	ft.SetActive(true)
	if !ft.Active() {
		t.Error("expected FocusTrap to be active after SetActive(true)")
	}
}

func TestFocusTrap_FocusStaysWithinTrap(t *testing.T) {
	btn1 := newFTTestFocusable("btn1")
	btn2 := newFTTestFocusable("btn2")
	btn3 := newFTTestFocusable("btn3")
	container := newFTTestContainer(btn1, btn2, btn3)
	ft := NewFocusTrap(container)

	// Layout and initial setup
	ft.Layout(runtime.Rect{X: 0, Y: 0, Width: 80, Height: 24})
	ft.Render(runtime.RenderContext{Buffer: runtime.NewBuffer(80, 24)})

	// After render+refreshScope, scope should have registered widgets.
	// The container itself is registered (Base satisfies Focusable) plus
	// the 3 buttons, totaling 4.
	scope := ft.Scope()
	if scope.Count() != 4 {
		t.Fatalf("expected 4 registered widgets in scope (1 container + 3 buttons), got %d", scope.Count())
	}

	// First focusable widget (btn1) should be focused (AutoFocusFirst).
	// The container is registered but CanFocus()=false, so it is skipped.
	current := scope.Current()
	if current != btn1 {
		t.Error("expected btn1 to be focused initially")
	}

	// Tab forward: btn1 -> btn2
	result := ft.HandleMessage(runtime.KeyMsg{Key: terminal.KeyTab})
	if !result.Handled {
		t.Error("expected Tab to be handled")
	}
	if scope.Current() != btn2 {
		t.Error("expected btn2 to be focused after first Tab")
	}

	// Tab forward: btn2 -> btn3
	ft.HandleMessage(runtime.KeyMsg{Key: terminal.KeyTab})
	if scope.Current() != btn3 {
		t.Error("expected btn3 to be focused after second Tab")
	}

	// Tab forward: btn3 -> btn1 (wraps around)
	ft.HandleMessage(runtime.KeyMsg{Key: terminal.KeyTab})
	if scope.Current() != btn1 {
		t.Error("expected focus to wrap around to btn1 after third Tab")
	}
}

func TestFocusTrap_FocusWrapsBackward(t *testing.T) {
	btn1 := newFTTestFocusable("btn1")
	btn2 := newFTTestFocusable("btn2")
	container := newFTTestContainer(btn1, btn2)
	ft := NewFocusTrap(container)

	ft.Layout(runtime.Rect{X: 0, Y: 0, Width: 80, Height: 24})
	ft.Render(runtime.RenderContext{Buffer: runtime.NewBuffer(80, 24)})

	scope := ft.Scope()
	// btn1 is focused initially
	if scope.Current() != btn1 {
		t.Error("expected btn1 to be focused initially")
	}

	// Shift+Tab from btn1 -> btn2 (wraps backward)
	result := ft.HandleMessage(runtime.KeyMsg{Key: terminal.KeyTab, Shift: true})
	if !result.Handled {
		t.Error("expected Shift+Tab to be handled")
	}
	if scope.Current() != btn2 {
		t.Error("expected focus to wrap backward to btn2")
	}

	// Shift+Tab from btn2 -> btn1
	ft.HandleMessage(runtime.KeyMsg{Key: terminal.KeyTab, Shift: true})
	if scope.Current() != btn1 {
		t.Error("expected focus to move to btn1")
	}
}

func TestFocusTrap_InactiveLetsKeyThrough(t *testing.T) {
	btn := newFTTestFocusable("btn")
	ft := NewFocusTrap(btn)
	ft.SetActive(false)

	// When inactive, Tab should be passed to child (which returns Unhandled)
	result := ft.HandleMessage(runtime.KeyMsg{Key: terminal.KeyTab})
	if result.Handled {
		t.Error("expected Tab to pass through when trap is inactive")
	}
}

func TestFocusTrap_OtherKeysPassToChild(t *testing.T) {
	btn := newFTTestFocusable("btn")
	btn.Focus()
	ft := NewFocusTrap(btn)

	// Non-Tab keys should pass through to the child
	result := ft.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})
	// The child's HandleMessage returns Unhandled for Escape
	if result.Handled {
		t.Error("expected Escape to pass through to child")
	}
}

func TestFocusTrap_AccessibilityRole(t *testing.T) {
	ft := NewFocusTrap(newFTTestFocusable("btn"))
	if ft.AccessibleRole() != "dialog" {
		t.Errorf("expected role 'dialog', got %q", ft.AccessibleRole())
	}
}

func TestFocusTrap_AnnouncementLabel(t *testing.T) {
	ft := NewFocusTrap(newFTTestFocusable("btn"))
	ft.SetLabel("Login Form")
	if ft.Base.Label != "Login Form" {
		t.Errorf("expected label 'Login Form', got %q", ft.Base.Label)
	}
}

func TestFocusTrap_SetChild(t *testing.T) {
	btn1 := newFTTestFocusable("btn1")
	btn2 := newFTTestFocusable("btn2")
	ft := NewFocusTrap(btn1)

	ft.SetChild(btn2)
	if ft.Child() != btn2 {
		t.Error("expected child to be updated")
	}
}

func TestFocusTrap_StyleType(t *testing.T) {
	ft := NewFocusTrap(newFTTestFocusable("btn"))
	if ft.StyleType() != "FocusTrap" {
		t.Errorf("expected StyleType 'FocusTrap', got %q", ft.StyleType())
	}
}

func TestFocusTrap_ChildWidgets(t *testing.T) {
	btn := newFTTestFocusable("btn")
	ft := NewFocusTrap(btn)

	children := ft.ChildWidgets()
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}
	if children[0] != btn {
		t.Error("expected child to match")
	}
}

func TestFocusTrap_NilChild(t *testing.T) {
	ft := NewFocusTrap(nil)
	children := ft.ChildWidgets()
	if children != nil {
		t.Error("expected nil children for nil child")
	}

	result := ft.HandleMessage(runtime.KeyMsg{Key: terminal.KeyTab})
	if result.Handled {
		t.Error("expected unhandled with nil child")
	}
}

func TestFocusTrap_Measure(t *testing.T) {
	btn := newFTTestFocusable("btn1")
	ft := NewFocusTrap(btn)

	size := ft.Measure(runtime.Loose(80, 24))
	childSize := btn.Measure(runtime.Loose(80, 24))
	if size != childSize {
		t.Errorf("expected FocusTrap size %v to equal child size %v", size, childSize)
	}
}

func TestFocusTrap_MeasureNilChild(t *testing.T) {
	ft := NewFocusTrap(nil)
	size := ft.Measure(runtime.Loose(80, 24))
	if size.Width != 0 || size.Height != 0 {
		t.Errorf("expected zero size for nil child, got %v", size)
	}
}

func TestFocusTrap_BindUnbind(t *testing.T) {
	ft := NewFocusTrap(newFTTestFocusable("btn"))

	// Bind should not panic
	ft.Bind(runtime.Services{})
	// Unbind should not panic
	ft.Unbind()
}

func TestFocusTrap_ReactivateRefreshes(t *testing.T) {
	btn1 := newFTTestFocusable("btn1")
	btn2 := newFTTestFocusable("btn2")
	container := newFTTestContainer(btn1, btn2)
	ft := NewFocusTrap(container)

	ft.Layout(runtime.Rect{X: 0, Y: 0, Width: 80, Height: 24})
	ft.Render(runtime.RenderContext{Buffer: runtime.NewBuffer(80, 24)})

	// Deactivate
	ft.SetActive(false)

	// Reactivate - should refresh scope
	ft.SetActive(true)
	scope := ft.Scope()
	// 1 container + 2 buttons = 3 registered widgets
	if scope.Count() != 3 {
		t.Fatalf("expected 3 registered widgets after reactivation, got %d", scope.Count())
	}
}

func TestFocusTrap_PathSegment(t *testing.T) {
	ft := NewFocusTrap(newFTTestFocusable("btn"))
	if seg := ft.PathSegment(nil); seg != "FocusTrap" {
		t.Errorf("expected PathSegment 'FocusTrap', got %q", seg)
	}
}

func TestFocusTrap_NilReceiver(t *testing.T) {
	var ft *FocusTrap

	// All nil receiver methods should not panic
	ft.SetActive(true)
	ft.SetLabel("test")
	ft.SetChild(nil)
	ft.Bind(runtime.Services{})
	ft.Unbind()

	if ft.Active() {
		t.Error("nil FocusTrap should return false for Active()")
	}
	if ft.Child() != nil {
		t.Error("nil FocusTrap should return nil Child()")
	}
	if ft.Scope() != nil {
		t.Error("nil FocusTrap should return nil Scope()")
	}
}
