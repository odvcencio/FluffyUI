package widgets

import (
	"testing"

	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/terminal"
)

func TestRating_NewRating(t *testing.T) {
	r := NewRating(5)
	if r == nil {
		t.Fatal("expected non-nil rating")
	}
	if r.Value() != 0 {
		t.Fatalf("expected value 0, got %d", r.Value())
	}
	if r.Max() != 5 {
		t.Fatalf("expected max 5, got %d", r.Max())
	}
	if r.AccessibleRole() != "slider" {
		t.Fatalf("expected role slider, got %q", r.AccessibleRole())
	}
	vi := r.AccessibleValue()
	if vi == nil {
		t.Fatal("expected non-nil ValueInfo")
	}
	if vi.Max != 5 {
		t.Fatalf("expected ValueInfo.Max=5, got %v", vi.Max)
	}
}

func TestRating_SetValue(t *testing.T) {
	r := NewRating(5)
	var changed int
	r.SetOnChange(func(v int) {
		changed = v
	})

	r.SetValue(3)
	if r.Value() != 3 {
		t.Fatalf("expected value 3, got %d", r.Value())
	}
	if changed != 3 {
		t.Fatalf("expected onChange with 3, got %d", changed)
	}

	// Clamp to max.
	r.SetValue(10)
	if r.Value() != 5 {
		t.Fatalf("expected value clamped to 5, got %d", r.Value())
	}

	// Clamp to min.
	r.SetValue(-1)
	if r.Value() != 0 {
		t.Fatalf("expected value clamped to 0, got %d", r.Value())
	}

	// Setting same value should not trigger onChange.
	changed = -1
	r.SetValue(0)
	if changed != -1 {
		t.Fatal("expected no onChange when setting same value")
	}
}

func TestRating_KeyNavigation(t *testing.T) {
	r := NewRating(5)
	r.Focus()

	// Right increases.
	r.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if r.Value() != 1 {
		t.Fatalf("expected value 1 after right, got %d", r.Value())
	}

	r.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	r.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if r.Value() != 3 {
		t.Fatalf("expected value 3 after three rights, got %d", r.Value())
	}

	// Left decreases.
	r.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})
	if r.Value() != 2 {
		t.Fatalf("expected value 2 after left, got %d", r.Value())
	}

	// End sets to max.
	r.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnd})
	if r.Value() != 5 {
		t.Fatalf("expected value 5 at end, got %d", r.Value())
	}

	// Home sets to 0.
	r.HandleMessage(runtime.KeyMsg{Key: terminal.KeyHome})
	if r.Value() != 0 {
		t.Fatalf("expected value 0 at home, got %d", r.Value())
	}

	// Right at max should not exceed.
	r.SetValue(5)
	result := r.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if result.Handled {
		t.Fatal("expected right at max to be unhandled")
	}
	if r.Value() != 5 {
		t.Fatalf("expected value still 5, got %d", r.Value())
	}

	// Left at 0 should not go below.
	r.SetValue(0)
	result = r.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})
	if result.Handled {
		t.Fatal("expected left at 0 to be unhandled")
	}
}

func TestRating_Measure(t *testing.T) {
	r := NewRating(5)
	size := r.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 10})
	if size.Width != 5 {
		t.Fatalf("expected width 5, got %d", size.Width)
	}
	if size.Height != 1 {
		t.Fatalf("expected height 1, got %d", size.Height)
	}
}

func TestRating_RenderText(t *testing.T) {
	r := NewRating(5)
	r.SetValue(3)
	text := r.renderText()
	// Should contain 3 filled stars and 2 empty stars.
	if len([]rune(text)) != 5 {
		t.Fatalf("expected 5 runes, got %d in %q", len([]rune(text)), text)
	}
}

func TestRating_UnfocusedNoHandle(t *testing.T) {
	r := NewRating(5)
	result := r.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if result.Handled {
		t.Fatal("expected unhandled when not focused")
	}
}
