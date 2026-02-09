package widgets

import (
	"testing"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

func TestColorPicker_New(t *testing.T) {
	cp := NewColorPicker()
	if cp == nil {
		t.Fatal("expected non-nil color picker")
	}
	if cp.AccessibleRole() != "group" {
		t.Fatalf("expected role group, got %q", cp.AccessibleRole())
	}
	if cp.SelectedIndex() != 0 {
		t.Fatalf("expected selected index 0, got %d", cp.SelectedIndex())
	}
	if !cp.CanFocus() {
		t.Fatal("expected color picker to be focusable")
	}
}

func TestColorPicker_DefaultPalette(t *testing.T) {
	cp := NewColorPicker()
	// Default palette has 24 colors
	if len(cp.palette) != 24 {
		t.Fatalf("expected 24 palette entries, got %d", len(cp.palette))
	}
}

func TestColorPicker_SetSelected(t *testing.T) {
	cp := NewColorPicker()
	var changedColor backend.Color
	cp.onChange = func(c backend.Color) {
		changedColor = c
	}

	cp.SetSelected(5)
	if cp.SelectedIndex() != 5 {
		t.Fatalf("expected selected 5, got %d", cp.SelectedIndex())
	}
	if changedColor != cp.palette[5].Color {
		t.Fatal("expected onChange to fire with correct color")
	}

	// Clamp to range
	cp.SetSelected(-1)
	if cp.SelectedIndex() != 0 {
		t.Fatalf("expected selected clamped to 0, got %d", cp.SelectedIndex())
	}

	cp.SetSelected(100)
	if cp.SelectedIndex() != len(cp.palette)-1 {
		t.Fatalf("expected selected clamped to max, got %d", cp.SelectedIndex())
	}
}

func TestColorPicker_KeyNavigation(t *testing.T) {
	cp := NewColorPicker()
	cp.Focus()

	// Right
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if cp.SelectedIndex() != 1 {
		t.Fatalf("expected selected 1 after right, got %d", cp.SelectedIndex())
	}

	// Down (moves by columns)
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if cp.SelectedIndex() != 1+colorPickerColumns {
		t.Fatalf("expected selected %d after down, got %d", 1+colorPickerColumns, cp.SelectedIndex())
	}

	// Up
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if cp.SelectedIndex() != 1 {
		t.Fatalf("expected selected 1 after up, got %d", cp.SelectedIndex())
	}

	// Left
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})
	if cp.SelectedIndex() != 0 {
		t.Fatalf("expected selected 0 after left, got %d", cp.SelectedIndex())
	}

	// Home
	cp.SetSelected(10)
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyHome})
	if cp.SelectedIndex() != 0 {
		t.Fatalf("expected selected 0 at home, got %d", cp.SelectedIndex())
	}

	// End
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnd})
	if cp.SelectedIndex() != len(cp.palette)-1 {
		t.Fatalf("expected selected at end, got %d", cp.SelectedIndex())
	}
}

func TestColorPicker_UnfocusedNoHandle(t *testing.T) {
	cp := NewColorPicker()
	result := cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if result.Handled {
		t.Fatal("expected unhandled when not focused")
	}
}

func TestColorPicker_Measure(t *testing.T) {
	cp := NewColorPicker()
	size := cp.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 20})
	// 6 columns * 4 = 24 width
	if size.Width != 24 {
		t.Fatalf("expected width 24, got %d", size.Width)
	}
	// 24 colors / 6 cols = 4 rows + 1 for display = 5
	if size.Height != 5 {
		t.Fatalf("expected height 5, got %d", size.Height)
	}
}

func TestColorPicker_StyleType(t *testing.T) {
	cp := NewColorPicker()
	if cp.StyleType() != "ColorPicker" {
		t.Fatalf("expected StyleType 'ColorPicker', got %q", cp.StyleType())
	}
}

func TestColorPicker_CustomPalette(t *testing.T) {
	colors := []backend.Color{backend.ColorRed, backend.ColorBlue}
	names := []string{"Red", "Blue"}
	cp := NewColorPicker(WithColorPickerPalette(colors, names))
	if len(cp.palette) != 2 {
		t.Fatalf("expected 2 palette entries, got %d", len(cp.palette))
	}
	if cp.palette[0].Name != "Red" {
		t.Fatalf("expected first color name 'Red', got %q", cp.palette[0].Name)
	}
}
