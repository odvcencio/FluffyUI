package widgets

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

func TestFilePicker_New(t *testing.T) {
	dir := t.TempDir()
	fp := NewFilePicker(dir)
	if fp == nil {
		t.Fatal("expected non-nil file picker")
	}
	if fp.AccessibleRole() != "tree" {
		t.Fatalf("expected role tree, got %q", fp.AccessibleRole())
	}
	if !fp.CanFocus() {
		t.Fatal("expected file picker to be focusable")
	}
	if fp.CurrentDir() != dir {
		t.Fatalf("expected dir %q, got %q", dir, fp.CurrentDir())
	}
}

func TestFilePicker_DefaultDir(t *testing.T) {
	fp := NewFilePicker("")
	if fp.CurrentDir() == "" {
		t.Fatal("expected non-empty default dir")
	}
}

func TestFilePicker_EntriesWithFiles(t *testing.T) {
	dir := t.TempDir()
	// Create test files
	_ = os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test"), 0o644)
	_ = os.Mkdir(filepath.Join(dir, "src"), 0o755)

	fp := NewFilePicker(dir)
	// +1 for ".."
	if fp.EntryCount() != 4 { // .., main.go, README.md, src
		t.Fatalf("expected 4 entries, got %d", fp.EntryCount())
	}
}

func TestFilePicker_HiddenFiles(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, ".hidden"), []byte(""), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "visible"), []byte(""), 0o644)

	// Without hidden
	fp := NewFilePicker(dir)
	if fp.EntryCount() != 2 { // .. + visible
		t.Fatalf("expected 2 entries without hidden, got %d", fp.EntryCount())
	}

	// With hidden
	fp2 := NewFilePicker(dir, WithFilePickerShowHidden(true))
	if fp2.EntryCount() != 3 { // .. + .hidden + visible
		t.Fatalf("expected 3 entries with hidden, got %d", fp2.EntryCount())
	}
}

func TestFilePicker_FileFilter(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "readme.md"), []byte("# test"), 0o644)
	_ = os.Mkdir(filepath.Join(dir, "src"), 0o755)

	fp := NewFilePicker(dir, WithFileFilter("*.go"))
	// Should show: .. + src (directories always pass) + main.go
	if fp.EntryCount() != 3 {
		t.Fatalf("expected 3 entries with *.go filter, got %d", fp.EntryCount())
	}
}

func TestFilePicker_KeyNavigation(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0o644)

	fp := NewFilePicker(dir)
	fp.Focus()

	// Start at 0 (..)
	if fp.SelectedIndex() != 0 {
		t.Fatalf("expected selected 0, got %d", fp.SelectedIndex())
	}

	// Down
	fp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if fp.SelectedIndex() != 1 {
		t.Fatalf("expected selected 1 after down, got %d", fp.SelectedIndex())
	}

	// Down again
	fp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if fp.SelectedIndex() != 2 {
		t.Fatalf("expected selected 2 after second down, got %d", fp.SelectedIndex())
	}

	// Up
	fp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if fp.SelectedIndex() != 1 {
		t.Fatalf("expected selected 1 after up, got %d", fp.SelectedIndex())
	}

	// Home
	fp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyHome})
	if fp.SelectedIndex() != 0 {
		t.Fatalf("expected selected 0 at home, got %d", fp.SelectedIndex())
	}

	// End
	fp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnd})
	if fp.SelectedIndex() != fp.EntryCount()-1 {
		t.Fatalf("expected selected at end, got %d", fp.SelectedIndex())
	}
}

func TestFilePicker_NavigateIntoDir(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "sub")
	_ = os.Mkdir(subdir, 0o755)
	_ = os.WriteFile(filepath.Join(subdir, "inner.txt"), []byte("inner"), 0o644)

	fp := NewFilePicker(dir)
	fp.Focus()

	// Move to the "sub" directory entry and enter it
	fp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown}) // Move to "sub"
	fp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})

	if fp.CurrentDir() != subdir {
		t.Fatalf("expected dir %q, got %q", subdir, fp.CurrentDir())
	}
}

func TestFilePicker_NavigateUp(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "sub")
	_ = os.Mkdir(subdir, 0o755)

	fp := NewFilePicker(subdir)
	fp.Focus()

	// Select ".." and press Enter
	fp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if fp.CurrentDir() != dir {
		t.Fatalf("expected parent dir %q, got %q", dir, fp.CurrentDir())
	}
}

func TestFilePicker_OnSelect(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "test.txt"), []byte("test"), 0o644)

	var selectedPath string
	fp := NewFilePicker(dir, WithOnFileSelect(func(path string) {
		selectedPath = path
	}))
	fp.Focus()

	// Move to file and select
	fp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown}) // test.txt
	fp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})

	expected := filepath.Join(dir, "test.txt")
	if selectedPath != expected {
		t.Fatalf("expected selected path %q, got %q", expected, selectedPath)
	}
}

func TestFilePicker_Measure(t *testing.T) {
	dir := t.TempDir()
	fp := NewFilePicker(dir)
	size := fp.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 40})
	if size.Width < 1 {
		t.Fatalf("expected width >= 1, got %d", size.Width)
	}
	if size.Height < 3 {
		t.Fatalf("expected height >= 3, got %d", size.Height)
	}
}

func TestFilePicker_StyleType(t *testing.T) {
	fp := NewFilePicker("")
	if fp.StyleType() != "FilePicker" {
		t.Fatalf("expected StyleType 'FilePicker', got %q", fp.StyleType())
	}
}

func TestFilePicker_UnfocusedNoHandle(t *testing.T) {
	fp := NewFilePicker("")
	result := fp.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if result.Handled {
		t.Fatal("expected unhandled when not focused")
	}
}

func TestFilePicker_FormatFileSize(t *testing.T) {
	tests := []struct {
		size     int64
		expected string
	}{
		{0, "0B"},
		{500, "500B"},
		{1024, "1.0K"},
		{1536, "1.5K"},
		{1048576, "1.0M"},
		{1073741824, "1.0G"},
	}
	for _, tt := range tests {
		got := formatFileSize(tt.size)
		if got != tt.expected {
			t.Errorf("formatFileSize(%d) = %q, want %q", tt.size, got, tt.expected)
		}
	}
}
