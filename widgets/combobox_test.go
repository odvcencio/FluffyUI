package widgets

import (
	"strings"
	"testing"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
	fluffytest "github.com/odvcencio/fluffyui/testing"
)

func TestCombobox_New(t *testing.T) {
	cb := NewCombobox("Apple", "Banana", "Cherry")
	if cb == nil {
		t.Fatal("expected non-nil Combobox")
	}
	if cb.Text() != "" {
		t.Fatalf("expected empty text, got %q", cb.Text())
	}
	if len(cb.Options()) != 3 {
		t.Fatalf("expected 3 options, got %d", len(cb.Options()))
	}
	if !cb.CanFocus() {
		t.Fatal("expected combobox to be focusable")
	}
}

func TestCombobox_Role(t *testing.T) {
	cb := NewCombobox("One", "Two")
	if cb.AccessibleRole() != accessibility.RoleCombobox {
		t.Fatalf("expected role combobox, got %q", cb.AccessibleRole())
	}
	if cb.AccessibleHasPopup() != "listbox" {
		t.Fatalf("expected HasPopup listbox, got %q", cb.AccessibleHasPopup())
	}
}

func TestCombobox_TypeFilters(t *testing.T) {
	cb := NewCombobox("Apple", "Banana", "Apricot")
	cb.Focus()

	// Type 'a' to filter
	cb.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'a'})
	if cb.Text() != "a" {
		t.Fatalf("expected text 'a', got %q", cb.Text())
	}
	// filtered should contain Apple, Banana (has 'a'), and Apricot
	if len(cb.filtered) != 3 {
		t.Fatalf("expected 3 filtered options, got %d", len(cb.filtered))
	}
}

func TestCombobox_ArrowAndEnterSelects(t *testing.T) {
	cb := NewCombobox("Apple", "Banana", "Cherry")
	cb.Focus()

	var changed string
	cb.SetOnChange(func(s string) { changed = s })

	// Open dropdown
	cb.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if !cb.open {
		t.Fatal("expected dropdown to open on Down")
	}

	// Move down and select
	cb.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	result := cb.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if !result.Handled {
		t.Fatal("expected Enter to be handled")
	}
	if cb.Text() != "Banana" {
		t.Fatalf("expected text 'Banana', got %q", cb.Text())
	}
	if changed != "Banana" {
		t.Fatalf("expected onChange 'Banana', got %q", changed)
	}
}

func TestCombobox_Measure(t *testing.T) {
	cb := NewCombobox("Apple", "Banana")
	size := cb.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 10})
	if size.Width < 10 {
		t.Fatalf("expected width >= 10, got %d", size.Width)
	}
	if size.Height < 1 {
		t.Fatalf("expected height >= 1, got %d", size.Height)
	}
}

func TestCombobox_Render(t *testing.T) {
	cb := NewCombobox("Apple", "Banana")
	output := fluffytest.RenderToString(cb, 20, 1)
	if !strings.Contains(output, "v]") {
		t.Fatalf("expected dropdown indicator, got %q", output)
	}
}

func TestCombobox_SetText(t *testing.T) {
	cb := NewCombobox("Apple", "Banana")
	cb.SetText("Ban")
	if cb.Text() != "Ban" {
		t.Fatalf("expected text 'Ban', got %q", cb.Text())
	}
	if len(cb.filtered) != 1 {
		t.Fatalf("expected 1 filtered, got %d", len(cb.filtered))
	}
}

func TestCombobox_SetOptions(t *testing.T) {
	cb := NewCombobox("A", "B")
	cb.SetOptions([]string{"X", "Y", "Z"})
	if len(cb.Options()) != 3 {
		t.Fatalf("expected 3 options, got %d", len(cb.Options()))
	}
}

func TestCombobox_EscapeCloses(t *testing.T) {
	cb := NewCombobox("Apple", "Banana")
	cb.Focus()
	cb.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if !cb.open {
		t.Fatal("expected open")
	}
	cb.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})
	if cb.open {
		t.Fatal("expected closed after Escape")
	}
}

func TestCombobox_Backspace(t *testing.T) {
	cb := NewCombobox("Apple", "Banana")
	cb.Focus()
	cb.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'a'})
	cb.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'p'})
	if cb.Text() != "ap" {
		t.Fatalf("expected 'ap', got %q", cb.Text())
	}
	cb.HandleMessage(runtime.KeyMsg{Key: terminal.KeyBackspace})
	if cb.Text() != "a" {
		t.Fatalf("expected 'a' after backspace, got %q", cb.Text())
	}
}
