package widgets

import (
	"testing"

	"m31labs.dev/fluffyui/runtime"
)

func TestFixedSpacer_New(t *testing.T) {
	s := NewFixedSpacer(10, 5)
	if s == nil {
		t.Fatal("expected non-nil FixedSpacer")
	}
	if s.width != 10 {
		t.Fatalf("expected width 10, got %d", s.width)
	}
	if s.height != 5 {
		t.Fatalf("expected height 5, got %d", s.height)
	}
}

func TestFixedSpacer_Role(t *testing.T) {
	s := NewFixedSpacer(1, 1)
	if s.AccessibleRole() != "presentation" {
		t.Fatalf("expected role presentation, got %q", s.AccessibleRole())
	}
}

func TestFixedSpacer_Measure(t *testing.T) {
	s := NewFixedSpacer(10, 3)
	size := s.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 24})
	if size.Width != 10 {
		t.Fatalf("expected width 10, got %d", size.Width)
	}
	if size.Height != 3 {
		t.Fatalf("expected height 3, got %d", size.Height)
	}
}

func TestFixedSpacer_MeasureClampedByConstraints(t *testing.T) {
	s := NewFixedSpacer(100, 50)
	size := s.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 10})
	if size.Width != 20 {
		t.Fatalf("expected width clamped to 20, got %d", size.Width)
	}
	if size.Height != 10 {
		t.Fatalf("expected height clamped to 10, got %d", size.Height)
	}
}

func TestFixedSpacer_MeasureZero(t *testing.T) {
	s := NewFixedSpacer(0, 0)
	size := s.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 24})
	if size.Width != 0 {
		t.Fatalf("expected width 0, got %d", size.Width)
	}
	if size.Height != 0 {
		t.Fatalf("expected height 0, got %d", size.Height)
	}
}

func TestFixedSpacer_MeasureNilReceiver(t *testing.T) {
	var s *FixedSpacer
	size := s.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 24})
	// Nil spacer returns constraints.MinSize() which is {0, 0} by default.
	if size.Width != 0 || size.Height != 0 {
		t.Fatalf("expected zero size for nil receiver, got %dx%d", size.Width, size.Height)
	}
}

func TestFixedSpacer_HandleMessageUnhandled(t *testing.T) {
	s := NewFixedSpacer(10, 5)
	result := s.HandleMessage(runtime.KeyMsg{})
	if result.Handled {
		t.Fatal("expected spacer to return unhandled for all messages")
	}
}

func TestFixedSpacer_RenderNoPanic(t *testing.T) {
	s := NewFixedSpacer(5, 3)
	buf := runtime.NewBuffer(20, 10)
	ctx := runtime.RenderContext{
		Buffer: buf,
		Bounds: runtime.Rect{X: 0, Y: 0, Width: 20, Height: 10},
	}
	// Render should be a no-op and not panic.
	s.Render(ctx)
}

func TestFixedSpacer_StyleType(t *testing.T) {
	s := NewFixedSpacer(1, 1)
	if s.StyleType() != "FixedSpacer" {
		t.Fatalf("expected StyleType 'FixedSpacer', got %q", s.StyleType())
	}
}

func TestFixedSpacer_WidgetInterface(t *testing.T) {
	// Compile-time check is in spacer.go, but verify at runtime too.
	var w runtime.Widget = NewFixedSpacer(1, 1)
	_ = w // compile-time interface satisfaction check
}

func TestFixedSpacer_NegativeDimensions(t *testing.T) {
	// Negative dimensions should be clamped to 0 by constraints.
	s := NewFixedSpacer(-5, -3)
	size := s.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 24})
	// runtime.Constraints.Constrain clamps negative values to 0.
	if size.Width != 0 {
		t.Fatalf("expected width 0 for negative input, got %d", size.Width)
	}
	if size.Height != 0 {
		t.Fatalf("expected height 0 for negative input, got %d", size.Height)
	}
}
