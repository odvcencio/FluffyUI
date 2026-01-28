package widgets

import (
	"testing"

	"github.com/odvcencio/fluffyui/runtime"
)

func TestPopoverLayoutBelow(t *testing.T) {
	child := NewSimpleWidget()
	var got runtime.Rect
	child.MeasureFunc = func(runtime.Constraints) runtime.Size {
		return runtime.Size{Width: 10, Height: 3}
	}
	child.LayoutFunc = func(bounds runtime.Rect) {
		got = bounds
	}

	anchor := runtime.Rect{X: 5, Y: 2, Width: 10, Height: 1}
	popover := NewPopover(anchor, child, WithPopoverMatchAnchorWidth(true))
	popover.Layout(runtime.Rect{X: 0, Y: 0, Width: 40, Height: 20})

	if got.Y != anchor.Y+anchor.Height {
		t.Errorf("expected popover below anchor at y=%d, got %d", anchor.Y+anchor.Height, got.Y)
	}
	if got.Width != anchor.Width {
		t.Errorf("expected popover width %d, got %d", anchor.Width, got.Width)
	}
}

func TestPopoverLayoutAboveWhenNoSpace(t *testing.T) {
	child := NewSimpleWidget()
	var got runtime.Rect
	child.MeasureFunc = func(runtime.Constraints) runtime.Size {
		return runtime.Size{Width: 8, Height: 3}
	}
	child.LayoutFunc = func(bounds runtime.Rect) {
		got = bounds
	}

	anchor := runtime.Rect{X: 2, Y: 18, Width: 6, Height: 1}
	popover := NewPopover(anchor, child)
	popover.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 20})

	if got.Y >= anchor.Y {
		t.Errorf("expected popover to flip above anchor, got y=%d", got.Y)
	}
}
