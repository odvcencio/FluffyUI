package widgets

import (
	"testing"

	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

func testCommands() []Command {
	return []Command{
		{Name: "Save", Description: "Save file", Shortcut: "Ctrl+S"},
		{Name: "Open", Description: "Open file", Shortcut: "Ctrl+O"},
		{Name: "Quit", Description: "Quit application", Shortcut: "Ctrl+Q"},
		{Name: "Search", Description: "Find text", Shortcut: "Ctrl+F"},
		{Name: "Save As", Description: "Save file as", Shortcut: "Ctrl+Shift+S"},
	}
}

func TestCommandPalette_New(t *testing.T) {
	cp := NewCommandPalette(testCommands())
	if cp == nil {
		t.Fatal("expected non-nil command palette")
	}
	if cp.AccessibleRole() != "combobox" {
		t.Fatalf("expected role combobox, got %q", cp.AccessibleRole())
	}
	if !cp.CanFocus() {
		t.Fatal("expected command palette to be focusable")
	}
	if cp.Open() {
		t.Fatal("expected palette to start closed")
	}
}

func TestCommandPalette_OpenClose(t *testing.T) {
	cp := NewCommandPalette(testCommands())
	cp.SetOpen(true)
	if !cp.Open() {
		t.Fatal("expected palette to be open")
	}
	cp.SetOpen(false)
	if cp.Open() {
		t.Fatal("expected palette to be closed")
	}
}

func TestCommandPalette_FilteredCount(t *testing.T) {
	cp := NewCommandPalette(testCommands())
	cp.SetOpen(true)
	if cp.FilteredCount() != 5 {
		t.Fatalf("expected 5 filtered commands, got %d", cp.FilteredCount())
	}
}

func TestCommandPalette_QueryFiltering(t *testing.T) {
	cp := NewCommandPalette(testCommands())
	cp.SetOpen(true)
	cp.Focus()

	// Type "sa" to filter
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 's'})
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'a'})
	if cp.Query() != "sa" {
		t.Fatalf("expected query 'sa', got %q", cp.Query())
	}

	// Should match "Save" and "Save As" and "Search" doesn't match "sa"
	if cp.FilteredCount() != 2 {
		t.Fatalf("expected 2 filtered commands for 'sa', got %d", cp.FilteredCount())
	}

	// Backspace
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyBackspace})
	if cp.Query() != "s" {
		t.Fatalf("expected query 's' after backspace, got %q", cp.Query())
	}
	// "s" matches Save, Search, Save As
	if cp.FilteredCount() != 3 {
		t.Fatalf("expected 3 filtered commands for 's', got %d", cp.FilteredCount())
	}
}

func TestCommandPalette_KeyNavigation(t *testing.T) {
	cp := NewCommandPalette(testCommands())
	cp.SetOpen(true)
	cp.Focus()

	// Down
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if cp.SelectedIndex() != 1 {
		t.Fatalf("expected selected 1 after down, got %d", cp.SelectedIndex())
	}

	// Up
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if cp.SelectedIndex() != 0 {
		t.Fatalf("expected selected 0 after up, got %d", cp.SelectedIndex())
	}

	// Up at 0 should stay at 0
	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if cp.SelectedIndex() != 0 {
		t.Fatalf("expected selected still 0, got %d", cp.SelectedIndex())
	}
}

func TestCommandPalette_Execute(t *testing.T) {
	executed := false
	cmds := []Command{
		{Name: "Test", Action: func() { executed = true }},
	}
	cp := NewCommandPalette(cmds)
	cp.SetOpen(true)
	cp.Focus()

	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if !executed {
		t.Fatal("expected command action to execute")
	}
	if cp.Open() {
		t.Fatal("expected palette to close after execution")
	}
}

func TestCommandPalette_EscapeCloses(t *testing.T) {
	cp := NewCommandPalette(testCommands())
	cp.SetOpen(true)
	cp.Focus()

	cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})
	if cp.Open() {
		t.Fatal("expected palette to close on escape")
	}
}

func TestCommandPalette_Measure(t *testing.T) {
	cp := NewCommandPalette(testCommands())
	cp.SetOpen(true)
	size := cp.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 40})
	if size.Width < 10 {
		t.Fatalf("expected width >= 10, got %d", size.Width)
	}
	if size.Height < 4 {
		t.Fatalf("expected height >= 4, got %d", size.Height)
	}
}

func TestCommandPalette_MeasureClosed(t *testing.T) {
	cp := NewCommandPalette(testCommands())
	size := cp.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 40})
	if size.Width != 0 || size.Height != 0 {
		t.Fatalf("expected zero size when closed, got %dx%d", size.Width, size.Height)
	}
}

func TestCommandPalette_StyleType(t *testing.T) {
	cp := NewCommandPalette(testCommands())
	if cp.StyleType() != "CommandPalette" {
		t.Fatalf("expected StyleType 'CommandPalette', got %q", cp.StyleType())
	}
}

func TestCommandPalette_UnfocusedNoHandle(t *testing.T) {
	cp := NewCommandPalette(testCommands())
	cp.SetOpen(true)
	result := cp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if result.Handled {
		t.Fatal("expected unhandled when not focused")
	}
}
