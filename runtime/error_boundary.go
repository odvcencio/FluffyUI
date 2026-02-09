package runtime

import (
	"fmt"

	"github.com/odvcencio/fluffyui/backend"
)

// ErrorBoundary wraps a widget and catches panics during Render and HandleMessage.
// When a panic is caught, the boundary renders a red error message instead of the
// child widget. Use Reset to clear the error and retry.
type ErrorBoundary struct {
	child  Widget
	err    error
	bounds Rect
}

// NewErrorBoundary creates an ErrorBoundary wrapping the given child widget.
func NewErrorBoundary(child Widget) *ErrorBoundary {
	return &ErrorBoundary{child: child}
}

// Measure delegates to the child widget, catching any panics.
// Returns zero if the child is nil, or the minimum constraint size
// if in an error state.
func (eb *ErrorBoundary) Measure(constraints Constraints) Size {
	if eb.child == nil {
		return Size{}
	}
	if eb.err != nil {
		return constraints.MinSize()
	}

	var result Size
	var caught error
	func() {
		defer func() {
			if r := recover(); r != nil {
				caught = fmt.Errorf("panic: %v", r)
			}
		}()
		result = eb.child.Measure(constraints)
	}()

	if caught != nil {
		eb.err = caught
		return constraints.MinSize()
	}
	return result
}

// Layout stores the bounds and delegates to the child widget, catching
// any panics. If in an error state, only stores the bounds without delegating.
func (eb *ErrorBoundary) Layout(bounds Rect) {
	eb.bounds = bounds
	if eb.err != nil || eb.child == nil {
		return
	}

	func() {
		defer func() {
			if r := recover(); r != nil {
				eb.err = fmt.Errorf("panic: %v", r)
			}
		}()
		eb.child.Layout(bounds)
	}()
}

// Render draws the child widget, catching any panics. If the boundary is in an
// error state, it renders a red error message instead.
func (eb *ErrorBoundary) Render(ctx RenderContext) {
	if eb.err != nil {
		eb.renderError(ctx)
		return
	}
	if eb.child == nil {
		return
	}

	var caught error
	func() {
		defer func() {
			if r := recover(); r != nil {
				caught = fmt.Errorf("panic: %v", r)
			}
		}()
		eb.child.Render(ctx)
	}()

	if caught != nil {
		eb.err = caught
		eb.renderError(ctx)
	}
}

// HandleMessage dispatches a message to the child widget, catching any panics.
// If the boundary is in an error state, messages are silently consumed.
func (eb *ErrorBoundary) HandleMessage(msg Message) HandleResult {
	if eb.err != nil {
		return Unhandled()
	}
	if eb.child == nil {
		return Unhandled()
	}

	var result HandleResult
	var caught error
	func() {
		defer func() {
			if r := recover(); r != nil {
				caught = fmt.Errorf("panic: %v", r)
			}
		}()
		result = eb.child.HandleMessage(msg)
	}()

	if caught != nil {
		eb.err = caught
		return Unhandled()
	}
	return result
}

// Bounds returns the assigned bounds.
func (eb *ErrorBoundary) Bounds() Rect {
	return eb.bounds
}

// ChildWidgets returns the child widget for container traversal.
func (eb *ErrorBoundary) ChildWidgets() []Widget {
	if eb.child == nil {
		return nil
	}
	return []Widget{eb.child}
}

// Error returns the captured error, or nil if none.
func (eb *ErrorBoundary) Error() error {
	return eb.err
}

// Reset clears the error state so the child widget will be rendered again.
func (eb *ErrorBoundary) Reset() {
	eb.err = nil
}

// CanFocus delegates to the child if it implements Focusable.
func (eb *ErrorBoundary) CanFocus() bool {
	if f, ok := eb.child.(Focusable); ok {
		return f.CanFocus()
	}
	return false
}

// Focus delegates to the child if it implements Focusable.
func (eb *ErrorBoundary) Focus() {
	if f, ok := eb.child.(Focusable); ok {
		f.Focus()
	}
}

// Blur delegates to the child if it implements Focusable.
func (eb *ErrorBoundary) Blur() {
	if f, ok := eb.child.(Focusable); ok {
		f.Blur()
	}
}

// IsFocused delegates to the child if it implements Focusable.
func (eb *ErrorBoundary) IsFocused() bool {
	if f, ok := eb.child.(Focusable); ok {
		return f.IsFocused()
	}
	return false
}

// renderError fills the bounds with a red background and writes the error text.
func (eb *ErrorBoundary) renderError(ctx RenderContext) {
	errStyle := backend.DefaultStyle().
		Foreground(backend.ColorWhite).
		Background(backend.ColorRed)

	// Fill the entire bounds area with the error style.
	ctx.Buffer.Fill(eb.bounds, ' ', errStyle)

	// Write the error message on the first line within bounds.
	msg := "Error: " + eb.err.Error()
	// Truncate message to fit within bounds width.
	if len(msg) > eb.bounds.Width && eb.bounds.Width > 0 {
		msg = msg[:eb.bounds.Width]
	}
	if eb.bounds.Width > 0 && eb.bounds.Height > 0 {
		ctx.Buffer.SetString(eb.bounds.X, eb.bounds.Y, msg, errStyle)
	}
}
