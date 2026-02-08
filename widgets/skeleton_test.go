package widgets

import (
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/runtime"
)

func TestSkeleton_NewSkeleton(t *testing.T) {
	s := NewSkeleton(20, 3)
	if s == nil {
		t.Fatal("expected non-nil skeleton")
	}
	if s.AccessibleRole() != "status" {
		t.Fatalf("expected role status, got %q", s.AccessibleRole())
	}
	if s.AccessibleLive() != "polite" {
		t.Fatalf("expected live polite, got %q", s.AccessibleLive())
	}
	if !s.AccessibleState().Busy {
		t.Fatal("expected busy state to be true")
	}
	if s.AccessibleLabel() != "Loading" {
		t.Fatalf("expected label 'Loading', got %q", s.AccessibleLabel())
	}
	if !s.IsAnimating() {
		t.Fatal("expected animate to be true by default")
	}
}

func TestSkeleton_Measure(t *testing.T) {
	s := NewSkeleton(15, 2)
	size := s.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 10})
	if size.Width != 15 {
		t.Fatalf("expected width 15, got %d", size.Width)
	}
	if size.Height != 2 {
		t.Fatalf("expected height 2, got %d", size.Height)
	}
}

func TestSkeleton_Animate(t *testing.T) {
	s := NewSkeleton(10, 1)
	if s.Frame() != 0 {
		t.Fatalf("expected initial frame 0, got %d", s.Frame())
	}

	// Tick should advance frame.
	result := s.HandleMessage(runtime.TickMsg{Time: time.Now()})
	if !result.Handled {
		t.Fatal("expected tick to be handled")
	}
	if s.Frame() != 1 {
		t.Fatalf("expected frame 1 after tick, got %d", s.Frame())
	}

	// Multiple ticks.
	s.HandleMessage(runtime.TickMsg{Time: time.Now()})
	s.HandleMessage(runtime.TickMsg{Time: time.Now()})
	if s.Frame() != 3 {
		t.Fatalf("expected frame 3, got %d", s.Frame())
	}

	// Disable animation.
	s.SetAnimate(false)
	result = s.HandleMessage(runtime.TickMsg{Time: time.Now()})
	if result.Handled {
		t.Fatal("expected tick to be unhandled when animation disabled")
	}
	if s.Frame() != 3 {
		t.Fatalf("expected frame unchanged at 3, got %d", s.Frame())
	}
}

func TestSkeleton_DefaultSize(t *testing.T) {
	// Negative dimensions should be clamped.
	s := NewSkeleton(-5, 0)
	size := s.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 10})
	if size.Width < 1 {
		t.Fatalf("expected width >= 1, got %d", size.Width)
	}
	if size.Height < 1 {
		t.Fatalf("expected height >= 1, got %d", size.Height)
	}
}

func TestSkeleton_NonKeyUnhandled(t *testing.T) {
	s := NewSkeleton(10, 1)
	result := s.HandleMessage(runtime.KeyMsg{})
	if result.Handled {
		t.Fatal("expected key message to be unhandled by skeleton")
	}
}
