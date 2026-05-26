package widgets

import (
	"strings"
	"testing"

	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/state"
	"m31labs.dev/fluffyui/terminal"
	fluffytest "m31labs.dev/fluffyui/testing"
)

func TestNumberInput_New(t *testing.T) {
	sig := state.NewSignal(42.0)
	n := NewNumberInput(sig)
	if n == nil {
		t.Fatal("expected non-nil NumberInput")
	}
	if n.Value() != 42 {
		t.Fatalf("expected value 42, got %v", n.Value())
	}
	if n.Min() != 0 {
		t.Fatalf("expected min 0, got %v", n.Min())
	}
	if n.Max() != 100 {
		t.Fatalf("expected max 100, got %v", n.Max())
	}
	if n.Step() != 1 {
		t.Fatalf("expected step 1, got %v", n.Step())
	}
	if !n.CanFocus() {
		t.Fatal("expected NumberInput to be focusable")
	}
}

func TestNumberInput_Role(t *testing.T) {
	sig := state.NewSignal(0.0)
	n := NewNumberInput(sig)
	if n.AccessibleRole() != "spinbutton" {
		t.Fatalf("expected role spinbutton, got %q", n.AccessibleRole())
	}
	vi := n.AccessibleValue()
	if vi == nil {
		t.Fatal("expected non-nil ValueInfo")
	}
	if vi.Min != 0 || vi.Max != 100 {
		t.Fatalf("expected ValueInfo min=0 max=100, got min=%v max=%v", vi.Min, vi.Max)
	}
}

func TestNumberInput_WithOptions(t *testing.T) {
	sig := state.NewSignal(5.0)
	n := NewNumberInput(sig,
		WithNumberRange(1, 50),
		WithNumberStep(0.5),
	)
	if n.Min() != 1 {
		t.Fatalf("expected min 1, got %v", n.Min())
	}
	if n.Max() != 50 {
		t.Fatalf("expected max 50, got %v", n.Max())
	}
	if n.Step() != 0.5 {
		t.Fatalf("expected step 0.5, got %v", n.Step())
	}
}

func TestNumberInput_SetValue_Clamp(t *testing.T) {
	sig := state.NewSignal(0.0)
	n := NewNumberInput(sig, WithNumberRange(0, 10))

	n.SetValue(5)
	if n.Value() != 5 {
		t.Fatalf("expected value 5, got %v", n.Value())
	}

	// Clamp above max.
	n.SetValue(20)
	if n.Value() != 10 {
		t.Fatalf("expected value clamped to 10, got %v", n.Value())
	}

	// Clamp below min.
	n.SetValue(-5)
	if n.Value() != 0 {
		t.Fatalf("expected value clamped to 0, got %v", n.Value())
	}
}

func TestNumberInput_OnChange(t *testing.T) {
	sig := state.NewSignal(0.0)
	var changed float64
	n := NewNumberInput(sig, WithNumberOnChange(func(v float64) {
		changed = v
	}))

	n.SetValue(7)
	if changed != 7 {
		t.Fatalf("expected onChange with 7, got %v", changed)
	}
}

func TestNumberInput_KeyUpDown(t *testing.T) {
	sig := state.NewSignal(5.0)
	n := NewNumberInput(sig, WithNumberRange(0, 10), WithNumberStep(2))
	n.Focus()

	// Up increments.
	n.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if n.Value() != 7 {
		t.Fatalf("expected 7 after up, got %v", n.Value())
	}

	// Down decrements.
	n.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if n.Value() != 5 {
		t.Fatalf("expected 5 after down, got %v", n.Value())
	}

	// Home sets to min.
	n.HandleMessage(runtime.KeyMsg{Key: terminal.KeyHome})
	if n.Value() != 0 {
		t.Fatalf("expected 0 at home, got %v", n.Value())
	}

	// End sets to max.
	n.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnd})
	if n.Value() != 10 {
		t.Fatalf("expected 10 at end, got %v", n.Value())
	}
}

func TestNumberInput_KeyDigits(t *testing.T) {
	sig := state.NewSignal(0.0)
	n := NewNumberInput(sig, WithNumberRange(0, 100))
	n.Focus()

	// Type "42" then Enter.
	n.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: '4'})
	n.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: '2'})
	n.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if n.Value() != 42 {
		t.Fatalf("expected 42 after typing, got %v", n.Value())
	}
}

func TestNumberInput_UnfocusedNoHandle(t *testing.T) {
	sig := state.NewSignal(0.0)
	n := NewNumberInput(sig)
	result := n.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if result.Handled {
		t.Fatal("expected unhandled when not focused")
	}
}

func TestNumberInput_Measure(t *testing.T) {
	sig := state.NewSignal(42.0)
	n := NewNumberInput(sig)
	size := n.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 10})
	if size.Width < 5 {
		t.Fatalf("expected width >= 5, got %d", size.Width)
	}
	if size.Height != 1 {
		t.Fatalf("expected height 1, got %d", size.Height)
	}
}

func TestNumberInput_Render(t *testing.T) {
	sig := state.NewSignal(7.0)
	n := NewNumberInput(sig)
	output := fluffytest.RenderToString(n, 20, 1)
	if !strings.Contains(output, "7") {
		t.Fatalf("expected output to contain '7', got %q", output)
	}
}
