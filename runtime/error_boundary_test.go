package runtime

import (
	"testing"

	"m31labs.dev/fluffyui/backend"
)

// panickingWidget panics in the specified method(s).
type panickingWidget struct {
	bounds       Rect
	panicRender  bool
	panicHandle  bool
	renderCalled bool
	handleCalled bool
}

func (p *panickingWidget) Measure(c Constraints) Size {
	return Size{Width: 10, Height: 3}
}

func (p *panickingWidget) Layout(bounds Rect) {
	p.bounds = bounds
}

func (p *panickingWidget) Render(ctx RenderContext) {
	p.renderCalled = true
	if p.panicRender {
		panic("render exploded")
	}
	ctx.Buffer.Fill(p.bounds, 'X', backend.DefaultStyle())
}

func (p *panickingWidget) HandleMessage(msg Message) HandleResult {
	p.handleCalled = true
	if p.panicHandle {
		panic("handle exploded")
	}
	return Handled()
}

// focusablePanickingWidget supports focus delegation testing.
type focusablePanickingWidget struct {
	panickingWidget
	focused  bool
	canFocus bool
}

func (f *focusablePanickingWidget) CanFocus() bool  { return f.canFocus }
func (f *focusablePanickingWidget) Focus()           { f.focused = true }
func (f *focusablePanickingWidget) Blur()            { f.focused = false }
func (f *focusablePanickingWidget) IsFocused() bool  { return f.focused }

// simpleMsg is a minimal Message for testing.
type ebTestMsg struct{}

func (ebTestMsg) isMessage() {}

func TestErrorBoundary_RenderPanicCaptured(t *testing.T) {
	child := &panickingWidget{panicRender: true}
	eb := NewErrorBoundary(child)

	buf := NewBuffer(40, 10)
	bounds := Rect{X: 0, Y: 0, Width: 40, Height: 10}
	eb.Layout(bounds)

	ctx := RenderContext{
		Buffer: buf,
		Bounds: bounds,
	}

	// Should NOT panic.
	eb.Render(ctx)

	if eb.Error() == nil {
		t.Fatal("expected error to be captured after render panic")
	}
	if eb.Error().Error() != "panic: render exploded" {
		t.Errorf("unexpected error message: %s", eb.Error().Error())
	}
	if !child.renderCalled {
		t.Error("expected child.Render to have been called")
	}
}

func TestErrorBoundary_HandleMessagePanicCaptured(t *testing.T) {
	child := &panickingWidget{panicHandle: true}
	eb := NewErrorBoundary(child)

	bounds := Rect{X: 0, Y: 0, Width: 40, Height: 10}
	eb.Layout(bounds)

	msg := ebTestMsg{}
	// Should NOT panic.
	result := eb.HandleMessage(msg)

	if eb.Error() == nil {
		t.Fatal("expected error to be captured after handle panic")
	}
	if eb.Error().Error() != "panic: handle exploded" {
		t.Errorf("unexpected error message: %s", eb.Error().Error())
	}
	if result.Handled {
		t.Error("expected unhandled result after panic")
	}
	if !child.handleCalled {
		t.Error("expected child.HandleMessage to have been called")
	}
}

func TestErrorBoundary_Reset(t *testing.T) {
	child := &panickingWidget{panicRender: true}
	eb := NewErrorBoundary(child)

	buf := NewBuffer(40, 10)
	bounds := Rect{X: 0, Y: 0, Width: 40, Height: 10}
	eb.Layout(bounds)

	ctx := RenderContext{
		Buffer: buf,
		Bounds: bounds,
	}

	// Trigger the panic.
	eb.Render(ctx)
	if eb.Error() == nil {
		t.Fatal("expected error after render panic")
	}

	// Reset should clear the error.
	eb.Reset()
	if eb.Error() != nil {
		t.Fatal("expected nil error after Reset")
	}

	// Fix the child so it no longer panics.
	child.panicRender = false
	child.renderCalled = false

	// Should render the child again without error.
	eb.Render(ctx)
	if eb.Error() != nil {
		t.Fatalf("expected no error after reset and fix, got: %v", eb.Error())
	}
	if !child.renderCalled {
		t.Error("expected child.Render to be called after reset")
	}
}

func TestErrorBoundary_NormalWidgetPassthrough(t *testing.T) {
	child := &panickingWidget{} // no panics configured
	eb := NewErrorBoundary(child)

	buf := NewBuffer(20, 5)
	bounds := Rect{X: 0, Y: 0, Width: 20, Height: 5}

	// Measure delegates.
	size := eb.Measure(Tight(20, 5))
	if size.Width != 10 || size.Height != 3 {
		t.Errorf("unexpected measure size: %v", size)
	}

	// Layout delegates.
	eb.Layout(bounds)
	if child.bounds != bounds {
		t.Errorf("expected child bounds %v, got %v", bounds, child.bounds)
	}
	if eb.Bounds() != bounds {
		t.Errorf("expected ErrorBoundary bounds %v, got %v", bounds, eb.Bounds())
	}

	// Render passes through.
	ctx := RenderContext{
		Buffer: buf,
		Bounds: bounds,
	}
	eb.Render(ctx)
	if eb.Error() != nil {
		t.Fatalf("expected no error, got: %v", eb.Error())
	}
	if !child.renderCalled {
		t.Error("expected child.Render to be called")
	}

	// HandleMessage passes through.
	result := eb.HandleMessage(ebTestMsg{})
	if !result.Handled {
		t.Error("expected handled result from normal child")
	}
	if !child.handleCalled {
		t.Error("expected child.HandleMessage to be called")
	}
}

func TestErrorBoundary_ErrorRenderShowsMessage(t *testing.T) {
	child := &panickingWidget{panicRender: true}
	eb := NewErrorBoundary(child)

	buf := NewBuffer(40, 5)
	bounds := Rect{X: 0, Y: 0, Width: 40, Height: 5}
	eb.Layout(bounds)

	ctx := RenderContext{
		Buffer: buf,
		Bounds: bounds,
	}

	// First render triggers panic.
	eb.Render(ctx)
	if eb.Error() == nil {
		t.Fatal("expected error after render panic")
	}

	// Second render should show the error message.
	eb.Render(ctx)

	// Check that the error text was written at position (0, 0).
	errStyle := backend.DefaultStyle().
		Foreground(backend.ColorWhite).
		Background(backend.ColorRed)

	// The first cell should have the error style.
	cell := buf.Get(0, 0)
	if cell.Style != errStyle {
		t.Error("expected error cell style (white on red)")
	}
	// The cell should contain 'E' from "Error: panic: render exploded".
	if cell.Rune != 'E' {
		t.Errorf("expected 'E' at (0,0), got %q", cell.Rune)
	}
}

func TestErrorBoundary_HandleMessageInErrorState(t *testing.T) {
	child := &panickingWidget{panicRender: true}
	eb := NewErrorBoundary(child)

	buf := NewBuffer(40, 5)
	bounds := Rect{X: 0, Y: 0, Width: 40, Height: 5}
	eb.Layout(bounds)

	// Trigger error via render panic.
	ctx := RenderContext{Buffer: buf, Bounds: bounds}
	eb.Render(ctx)
	if eb.Error() == nil {
		t.Fatal("expected error")
	}

	// In error state, HandleMessage should return unhandled without calling child.
	child.handleCalled = false
	result := eb.HandleMessage(ebTestMsg{})
	if result.Handled {
		t.Error("expected unhandled in error state")
	}
	if child.handleCalled {
		t.Error("child.HandleMessage should not be called in error state")
	}
}

func TestErrorBoundary_FocusDelegation(t *testing.T) {
	child := &focusablePanickingWidget{canFocus: true}
	eb := NewErrorBoundary(child)

	if !eb.CanFocus() {
		t.Error("expected CanFocus to return true")
	}

	eb.Focus()
	if !child.focused {
		t.Error("expected child to be focused after Focus()")
	}
	if !eb.IsFocused() {
		t.Error("expected IsFocused to return true")
	}

	eb.Blur()
	if child.focused {
		t.Error("expected child to be blurred after Blur()")
	}
	if eb.IsFocused() {
		t.Error("expected IsFocused to return false after Blur()")
	}
}

func TestErrorBoundary_FocusNonFocusableChild(t *testing.T) {
	child := &panickingWidget{} // does not implement Focusable
	eb := NewErrorBoundary(child)

	if eb.CanFocus() {
		t.Error("expected CanFocus to return false for non-focusable child")
	}
	if eb.IsFocused() {
		t.Error("expected IsFocused to return false for non-focusable child")
	}

	// These should not panic.
	eb.Focus()
	eb.Blur()
}

func TestErrorBoundary_NilChild(t *testing.T) {
	eb := NewErrorBoundary(nil)

	// Measure should return zero.
	size := eb.Measure(Tight(20, 5))
	if size.Width != 0 || size.Height != 0 {
		t.Errorf("expected zero size for nil child, got %v", size)
	}

	// Layout should not panic.
	eb.Layout(Rect{X: 0, Y: 0, Width: 20, Height: 5})

	// Render should not panic.
	buf := NewBuffer(20, 5)
	ctx := RenderContext{Buffer: buf, Bounds: Rect{X: 0, Y: 0, Width: 20, Height: 5}}
	eb.Render(ctx)

	// HandleMessage should return unhandled.
	result := eb.HandleMessage(ebTestMsg{})
	if result.Handled {
		t.Error("expected unhandled for nil child")
	}

	if eb.Error() != nil {
		t.Error("expected no error for nil child")
	}
}

func TestErrorBoundary_ChildWidgets(t *testing.T) {
	child := &panickingWidget{}
	eb := NewErrorBoundary(child)

	children := eb.ChildWidgets()
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}
	if children[0] != child {
		t.Error("expected child widget to be returned")
	}

	// Nil child case.
	eb2 := NewErrorBoundary(nil)
	if ch := eb2.ChildWidgets(); ch != nil {
		t.Errorf("expected nil children for nil child, got %v", ch)
	}
}
