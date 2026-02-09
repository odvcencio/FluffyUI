package widgets

import (
	"strings"
	"testing"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
	fluffytest "github.com/odvcencio/fluffyui/testing"
)

func TestSegmentedControl_New(t *testing.T) {
	sc := NewSegmentedControl("A", "B", "C")
	if sc == nil {
		t.Fatal("expected non-nil SegmentedControl")
	}
	if sc.Selected() != 0 {
		t.Fatalf("expected selected 0, got %d", sc.Selected())
	}
	if !sc.CanFocus() {
		t.Fatal("expected segmented control to be focusable")
	}
}

func TestSegmentedControl_Role(t *testing.T) {
	sc := NewSegmentedControl("A", "B")
	if sc.AccessibleRole() != accessibility.RoleRadioGroup {
		t.Fatalf("expected role radiogroup, got %q", sc.AccessibleRole())
	}
	if sc.AccessibleOrientation() != "horizontal" {
		t.Fatalf("expected horizontal orientation, got %q", sc.AccessibleOrientation())
	}
}

func TestSegmentedControl_SetSelected(t *testing.T) {
	var calledIdx int
	var calledLabel string
	sc := NewSegmentedControl("One", "Two", "Three")
	sc.SetOnChange(func(idx int, label string) {
		calledIdx = idx
		calledLabel = label
	})
	sc.SetSelected(2)
	if sc.Selected() != 2 {
		t.Fatalf("expected selected 2, got %d", sc.Selected())
	}
	if calledIdx != 2 || calledLabel != "Three" {
		t.Fatalf("expected onChange(2, Three), got (%d, %q)", calledIdx, calledLabel)
	}
}

func TestSegmentedControl_KeyboardNavigation(t *testing.T) {
	sc := NewSegmentedControl("A", "B", "C")
	sc.Focus()

	result := sc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if !result.Handled {
		t.Fatal("expected Right to be handled")
	}
	if sc.Selected() != 1 {
		t.Fatalf("expected selected 1, got %d", sc.Selected())
	}

	result = sc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if sc.Selected() != 2 {
		t.Fatalf("expected selected 2, got %d", sc.Selected())
	}

	// At end, Right should not go past
	result = sc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if sc.Selected() != 2 {
		t.Fatalf("expected selected still 2, got %d", sc.Selected())
	}

	result = sc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})
	if !result.Handled {
		t.Fatal("expected Left to be handled")
	}
	if sc.Selected() != 1 {
		t.Fatalf("expected selected 1, got %d", sc.Selected())
	}
}

func TestSegmentedControl_HomeEnd(t *testing.T) {
	sc := NewSegmentedControl("A", "B", "C")
	sc.Focus()
	sc.SetSelected(1)

	sc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnd})
	if sc.Selected() != 2 {
		t.Fatalf("expected End to go to 2, got %d", sc.Selected())
	}

	sc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyHome})
	if sc.Selected() != 0 {
		t.Fatalf("expected Home to go to 0, got %d", sc.Selected())
	}
}

func TestSegmentedControl_Measure(t *testing.T) {
	sc := NewSegmentedControl("One", "Two")
	size := sc.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 10})
	// [ One | Two ] = 13 chars
	if size.Width < 13 {
		t.Fatalf("expected width >= 13, got %d", size.Width)
	}
	if size.Height != 1 {
		t.Fatalf("expected height 1, got %d", size.Height)
	}
}

func TestSegmentedControl_Render(t *testing.T) {
	sc := NewSegmentedControl("Alpha", "Beta")
	output := fluffytest.RenderToString(sc, 30, 1)
	if !strings.Contains(output, "Alpha") {
		t.Fatalf("expected 'Alpha' in output, got %q", output)
	}
	if !strings.Contains(output, "Beta") {
		t.Fatalf("expected 'Beta' in output, got %q", output)
	}
	if !strings.Contains(output, "|") {
		t.Fatalf("expected separator '|' in output, got %q", output)
	}
}

func TestSegmentedControl_A11yValue(t *testing.T) {
	sc := NewSegmentedControl("X", "Y", "Z")
	sc.SetSelected(1)
	val := sc.AccessibleValue()
	if val == nil {
		t.Fatal("expected non-nil value")
	}
	if val.Text != "Y" {
		t.Fatalf("expected value text 'Y', got %q", val.Text)
	}
}

func TestSegmentedControl_UnfocusedNoHandle(t *testing.T) {
	sc := NewSegmentedControl("A", "B")
	result := sc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if result.Handled {
		t.Fatal("expected unhandled when not focused")
	}
}

func TestSegmentedControl_InvalidIndex(t *testing.T) {
	sc := NewSegmentedControl("A", "B")
	sc.SetSelected(-1)
	if sc.Selected() != 0 {
		t.Fatalf("expected selected to remain 0, got %d", sc.Selected())
	}
	sc.SetSelected(5)
	if sc.Selected() != 0 {
		t.Fatalf("expected selected to remain 0, got %d", sc.Selected())
	}
}
