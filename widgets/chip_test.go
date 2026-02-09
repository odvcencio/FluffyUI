package widgets

import (
	"strings"
	"testing"

	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
	fluffytest "github.com/odvcencio/fluffyui/testing"
)

func TestChip_New(t *testing.T) {
	c := NewChip("Go")
	if c == nil {
		t.Fatal("expected non-nil Chip")
	}
	if c.Label() != "Go" {
		t.Fatalf("expected label 'Go', got %q", c.Label())
	}
	if c.Variant() != ChipDefault {
		t.Fatalf("expected default variant, got %d", c.Variant())
	}
	if c.Dismissible() {
		t.Fatal("expected non-dismissible by default")
	}
}

func TestChip_Role(t *testing.T) {
	c := NewChip("tag")
	if c.AccessibleRole() != "status" {
		t.Fatalf("expected role status, got %q", c.AccessibleRole())
	}
}

func TestChip_WithVariant(t *testing.T) {
	c := NewChip("ok", WithChipVariant(ChipSuccess))
	if c.Variant() != ChipSuccess {
		t.Fatalf("expected ChipSuccess, got %d", c.Variant())
	}
}

func TestChip_Dismissible(t *testing.T) {
	dismissed := false
	c := NewChip("tag", WithChipDismiss(func() {
		dismissed = true
	}))
	if !c.Dismissible() {
		t.Fatal("expected dismissible chip")
	}
	if !c.CanFocus() {
		t.Fatal("expected dismissible chip to be focusable")
	}
	if c.AccessibleDescription() != "dismissible, press Enter to dismiss" {
		t.Fatalf("expected dismiss description, got %q", c.AccessibleDescription())
	}

	// Dismiss via Enter.
	c.Focus()
	result := c.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if !result.Handled {
		t.Fatal("expected enter to be handled")
	}
	if !dismissed {
		t.Fatal("expected onDismiss to be called")
	}
}

func TestChip_NonDismissibleNotFocusable(t *testing.T) {
	c := NewChip("tag")
	if c.CanFocus() {
		t.Fatal("expected non-dismissible chip to not be focusable")
	}
}

func TestChip_NonDismissibleUnhandled(t *testing.T) {
	c := NewChip("tag")
	c.Focus()
	result := c.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if result.Handled {
		t.Fatal("expected unhandled for non-dismissible chip")
	}
}

func TestChip_Measure(t *testing.T) {
	c := NewChip("Go")
	size := c.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 10})
	// "[Go]" = 4 chars
	if size.Width != 4 {
		t.Fatalf("expected width 4 for '[Go]', got %d", size.Width)
	}
	if size.Height != 1 {
		t.Fatalf("expected height 1, got %d", size.Height)
	}
}

func TestChip_MeasureDismissible(t *testing.T) {
	c := NewChip("Go", WithChipDismiss(func() {}))
	size := c.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 10})
	// "[Go x]" where x is the unicode multiply sign
	if size.Width < 5 {
		t.Fatalf("expected width >= 5 for dismissible chip, got %d", size.Width)
	}
}

func TestChip_Render(t *testing.T) {
	c := NewChip("Go")
	output := fluffytest.RenderToString(c, 20, 1)
	if !strings.Contains(output, "Go") {
		t.Fatalf("expected output to contain 'Go', got %q", output)
	}
}

func TestChip_SetLabel(t *testing.T) {
	c := NewChip("old")
	c.SetLabel("new")
	if c.Label() != "new" {
		t.Fatalf("expected label 'new', got %q", c.Label())
	}
	if c.AccessibleLabel() != "new" {
		t.Fatalf("expected accessible label 'new', got %q", c.AccessibleLabel())
	}
}

func TestChip_SpaceDismiss(t *testing.T) {
	dismissed := false
	c := NewChip("x", WithChipDismiss(func() {
		dismissed = true
	}))
	c.Focus()
	result := c.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: ' '})
	if !result.Handled {
		t.Fatal("expected space to be handled")
	}
	if !dismissed {
		t.Fatal("expected onDismiss to be called with space")
	}
}
