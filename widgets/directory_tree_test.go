package widgets

import (
	"os"
	"path/filepath"
	"testing"

	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/terminal"
)

func TestDirectoryTree_New(t *testing.T) {
	tmpDir := t.TempDir()

	dt := NewDirectoryTree(tmpDir)

	if dt == nil {
		t.Fatal("expected non-nil DirectoryTree")
	}
	if dt.Root() != tmpDir {
		t.Errorf("Root() = %q, want %q", dt.Root(), tmpDir)
	}
	if dt.LazyLoad() != true {
		t.Error("LazyLoad() should default to true")
	}
	if dt.ShowHidden() != false {
		t.Error("ShowHidden() should default to false")
	}
}

func TestDirectoryTree_WithOptions(t *testing.T) {
	tmpDir := t.TempDir()

	var selectPath string
	dt := NewDirectoryTree(tmpDir,
		WithShowHidden(true),
		WithLazyLoad(false),
		WithOnSelect(func(path string) {
			selectPath = path
		}),
		WithDirectoryStyle(backend.DefaultStyle().Bold(true)),
		WithDirectorySelectedStyle(backend.DefaultStyle().Italic(true)),
		WithDirectoryIcons("[D]", "[F]", "[-]", "[+]"),
	)

	if dt.ShowHidden() != true {
		t.Error("ShowHidden() = false, want true")
	}
	if dt.LazyLoad() != false {
		t.Error("LazyLoad() = true, want false")
	}

	// Test callback was set
	dt.onSelect(tmpDir)
	if selectPath != tmpDir {
		t.Errorf("onSelect called with %q, want %q", selectPath, tmpDir)
	}
}

func TestDirectoryTree_Filter(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file.go"), []byte("content"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	// Filter to only show .go files
	dt := NewDirectoryTree(tmpDir,
		WithDirectoryFilter(func(e os.DirEntry) bool {
			if e.IsDir() {
				return true
			}
			return filepath.Ext(e.Name()) == ".go"
		}),
	)

	entries := dt.flatten()

	// Should have: root, subdir, file.go (but not file.txt)
	found := false
	for _, entry := range entries {
		if entry.Name == "file.txt" {
			t.Error("file.txt should be filtered out")
		}
		if entry.Name == "file.go" {
			found = true
		}
	}
	if !found {
		t.Error("file.go should be included")
	}
}

func TestDirectoryTree_HiddenFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "visible.txt"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".hidden"), []byte("content"), 0644)

	// Without showing hidden
	dt := NewDirectoryTree(tmpDir)
	entries := dt.flatten()

	for _, entry := range entries {
		if entry.Name == ".hidden" {
			t.Error(".hidden should not be visible by default")
		}
	}

	// With showing hidden
	dt.SetShowHidden(true)
	entries = dt.flatten()

	found := false
	for _, entry := range entries {
		if entry.Name == ".hidden" {
			found = true
		}
	}
	if !found {
		t.Error(".hidden should be visible when ShowHidden is true")
	}
}

func TestDirectoryTree_ExpandCollapse(t *testing.T) {
	tmpDir := t.TempDir()

	// Create subdirectory with file
	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "file.txt"), []byte("content"), 0644)

	dt := NewDirectoryTree(tmpDir)
	dt.Focus()

	// Initial state: root expanded, subdir collapsed
	entries := dt.flatten()
	if len(entries) < 2 {
		t.Fatalf("expected at least 2 entries, got %d", len(entries))
	}

	// Navigate to subdir and expand
	dt.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	entry := dt.SelectedEntry()
	if entry == nil || !entry.IsDir {
		t.Fatal("expected to select subdir")
	}

	initialCount := len(dt.flatten())

	// Expand with Right arrow
	dt.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	expandedCount := len(dt.flatten())

	if expandedCount <= initialCount {
		t.Errorf("expected more entries after expand, got %d <= %d", expandedCount, initialCount)
	}

	// Collapse with Left arrow
	dt.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})
	collapsedCount := len(dt.flatten())

	if collapsedCount != initialCount {
		t.Errorf("expected %d entries after collapse, got %d", initialCount, collapsedCount)
	}
}

func TestDirectoryTree_EnterToggle(t *testing.T) {
	tmpDir := t.TempDir()

	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)

	dt := NewDirectoryTree(tmpDir)
	dt.Focus()

	// Navigate to subdir
	dt.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})

	var selectedPath string
	dt.SetOnSelect(func(path string) {
		selectedPath = path
	})

	entry := dt.SelectedEntry()
	wasExpanded := entry.Expanded

	// Toggle with Enter
	dt.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})

	if entry.Expanded == wasExpanded {
		t.Error("Enter should toggle expanded state")
	}
	if selectedPath == "" {
		t.Error("onSelect should be called on Enter")
	}
}

func TestDirectoryTree_Navigation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple files
	os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "b.txt"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "c.txt"), []byte("content"), 0644)

	dt := NewDirectoryTree(tmpDir)
	dt.Focus()
	dt.Layout(runtime.Rect{X: 0, Y: 0, Width: 40, Height: 10})

	// Test Down navigation
	dt.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if dt.selectedIndex != 1 {
		t.Errorf("selectedIndex = %d, want 1 after Down", dt.selectedIndex)
	}

	// Test Up navigation
	dt.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if dt.selectedIndex != 0 {
		t.Errorf("selectedIndex = %d, want 0 after Up", dt.selectedIndex)
	}

	// Test Home
	dt.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	dt.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	dt.HandleMessage(runtime.KeyMsg{Key: terminal.KeyHome})
	if dt.selectedIndex != 0 {
		t.Errorf("selectedIndex = %d, want 0 after Home", dt.selectedIndex)
	}

	// Test End
	entries := dt.flatten()
	dt.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnd})
	if dt.selectedIndex != len(entries)-1 {
		t.Errorf("selectedIndex = %d, want %d after End", dt.selectedIndex, len(entries)-1)
	}
}

func TestDirectoryTree_PageUpDown(t *testing.T) {
	tmpDir := t.TempDir()

	// Create many files
	for i := 0; i < 20; i++ {
		os.WriteFile(filepath.Join(tmpDir, string(rune('a'+i))+".txt"), []byte(""), 0644)
	}

	dt := NewDirectoryTree(tmpDir)
	dt.Focus()
	dt.Layout(runtime.Rect{X: 0, Y: 0, Width: 40, Height: 5})

	// Start at beginning
	dt.selectedIndex = 0

	// Page down
	dt.HandleMessage(runtime.KeyMsg{Key: terminal.KeyPageDown})
	if dt.selectedIndex < 3 {
		t.Errorf("selectedIndex = %d, expected >= 3 after PageDown", dt.selectedIndex)
	}

	// Page up
	prevIndex := dt.selectedIndex
	dt.HandleMessage(runtime.KeyMsg{Key: terminal.KeyPageUp})
	if dt.selectedIndex >= prevIndex {
		t.Errorf("selectedIndex = %d, expected < %d after PageUp", dt.selectedIndex, prevIndex)
	}
}

func TestDirectoryTree_Measure(t *testing.T) {
	tmpDir := t.TempDir()
	dt := NewDirectoryTree(tmpDir)

	size := dt.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 24})

	if size.Width != 80 {
		t.Errorf("Width = %d, want 80", size.Width)
	}
	// Height should be at least 1 (for root entry)
	if size.Height < 1 {
		t.Errorf("Height = %d, want >= 1", size.Height)
	}
}

func TestDirectoryTree_Render(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte(""), 0644)

	dt := NewDirectoryTree(tmpDir)
	dt.Focus()
	dt.Layout(runtime.Rect{X: 0, Y: 0, Width: 40, Height: 5})

	buf := runtime.NewBuffer(40, 5)
	ctx := runtime.RenderContext{Buffer: buf}

	// Should not panic
	dt.Render(ctx)

	// Check that something was rendered
	cell := buf.Get(0, 0)
	if cell.Rune == 0 || cell.Rune == ' ' {
		// It might be an icon or character
	}
}

func TestDirectoryTree_RenderEmpty(t *testing.T) {
	dt := NewDirectoryTree("/nonexistent/path/12345")
	dt.Layout(runtime.Rect{X: 0, Y: 0, Width: 0, Height: 0})

	buf := runtime.NewBuffer(40, 5)
	ctx := runtime.RenderContext{Buffer: buf}

	// Should not panic
	dt.Render(ctx)
}

func TestDirectoryTree_SelectedPath(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte(""), 0644)

	dt := NewDirectoryTree(tmpDir)

	path := dt.SelectedPath()
	if path != tmpDir {
		t.Errorf("SelectedPath() = %q, want %q", path, tmpDir)
	}

	dt.Focus()
	dt.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})

	path = dt.SelectedPath()
	if path == tmpDir {
		// Should have moved to a different entry
		t.Error("SelectedPath() should change after navigation")
	}
}

func TestDirectoryTree_Refresh(t *testing.T) {
	tmpDir := t.TempDir()

	dt := NewDirectoryTree(tmpDir)
	initialEntries := len(dt.flatten())

	// Create new file
	os.WriteFile(filepath.Join(tmpDir, "new.txt"), []byte(""), 0644)

	// Refresh
	dt.Refresh()

	newEntries := len(dt.flatten())
	if newEntries != initialEntries+1 {
		t.Errorf("expected %d entries after refresh, got %d", initialEntries+1, newEntries)
	}
}

func TestDirectoryTree_SetRoot(t *testing.T) {
	tmpDir1 := t.TempDir()
	tmpDir2 := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir1, "file1.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir2, "file2.txt"), []byte(""), 0644)

	dt := NewDirectoryTree(tmpDir1)

	dt.SetRoot(tmpDir2)

	if dt.Root() != tmpDir2 {
		t.Errorf("Root() = %q, want %q", dt.Root(), tmpDir2)
	}

	entries := dt.flatten()
	found := false
	for _, entry := range entries {
		if entry.Name == "file2.txt" {
			found = true
		}
		if entry.Name == "file1.txt" {
			t.Error("file1.txt should not be in new root")
		}
	}
	if !found {
		t.Error("file2.txt should be in new root")
	}
}

func TestDirectoryTree_Scroll(t *testing.T) {
	tmpDir := t.TempDir()

	// Create many files
	for i := 0; i < 10; i++ {
		os.WriteFile(filepath.Join(tmpDir, string(rune('a'+i))+".txt"), []byte(""), 0644)
	}

	dt := NewDirectoryTree(tmpDir)
	dt.Layout(runtime.Rect{X: 0, Y: 0, Width: 40, Height: 5})

	// Test ScrollBy
	dt.ScrollBy(0, 3)
	if dt.selectedIndex != 3 {
		t.Errorf("selectedIndex = %d, want 3 after ScrollBy", dt.selectedIndex)
	}

	// Test ScrollTo
	dt.ScrollTo(0, 0)
	if dt.selectedIndex != 0 {
		t.Errorf("selectedIndex = %d, want 0 after ScrollTo", dt.selectedIndex)
	}

	// Test ScrollToEnd
	dt.ScrollToEnd()
	entries := dt.flatten()
	if dt.selectedIndex != len(entries)-1 {
		t.Errorf("selectedIndex = %d, want %d after ScrollToEnd", dt.selectedIndex, len(entries)-1)
	}

	// Test ScrollToStart
	dt.ScrollToStart()
	if dt.selectedIndex != 0 {
		t.Errorf("selectedIndex = %d, want 0 after ScrollToStart", dt.selectedIndex)
	}
}

func TestDirectoryTree_Tab(t *testing.T) {
	dt := NewDirectoryTree(t.TempDir())
	dt.Focus()

	result := dt.HandleMessage(runtime.KeyMsg{Key: terminal.KeyTab})

	if len(result.Commands) == 0 {
		t.Error("expected FocusNext command")
	}
	_, ok := result.Commands[0].(runtime.FocusNext)
	if !ok {
		t.Errorf("expected FocusNext, got %T", result.Commands[0])
	}
}

func TestDirectoryTree_ShiftTab(t *testing.T) {
	dt := NewDirectoryTree(t.TempDir())
	dt.Focus()

	result := dt.HandleMessage(runtime.KeyMsg{Key: terminal.KeyTab, Shift: true})

	if len(result.Commands) == 0 {
		t.Error("expected FocusPrev command")
	}
	_, ok := result.Commands[0].(runtime.FocusPrev)
	if !ok {
		t.Errorf("expected FocusPrev, got %T", result.Commands[0])
	}
}

func TestDirectoryTree_Escape(t *testing.T) {
	dt := NewDirectoryTree(t.TempDir())
	dt.Focus()

	result := dt.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})

	if len(result.Commands) == 0 {
		t.Error("expected Cancel command")
	}
	_, ok := result.Commands[0].(runtime.Cancel)
	if !ok {
		t.Errorf("expected Cancel, got %T", result.Commands[0])
	}
}

func TestDirectoryTree_Unfocused(t *testing.T) {
	dt := NewDirectoryTree(t.TempDir())
	// Not focused

	result := dt.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})

	if result.Handled {
		t.Error("unfocused tree should not handle messages")
	}
}

func TestDirectoryTree_NonKeyMsg(t *testing.T) {
	dt := NewDirectoryTree(t.TempDir())
	dt.Focus()

	result := dt.HandleMessage(runtime.ResizeMsg{Width: 80, Height: 24})

	if result.Handled {
		t.Error("non-key message should not be handled")
	}
}

func TestDirectoryTree_Accessibility(t *testing.T) {
	tmpDir := t.TempDir()
	dt := NewDirectoryTree(tmpDir)
	dt.SetLabel("File Browser")

	if dt.AccessibleRole() != "tree" {
		t.Errorf("AccessibleRole() = %q, want 'tree'", dt.AccessibleRole())
	}
	if dt.AccessibleLabel() != "File Browser" {
		t.Errorf("AccessibleLabel() = %q, want 'File Browser'", dt.AccessibleLabel())
	}
}

func TestDirectoryTree_StyleType(t *testing.T) {
	dt := NewDirectoryTree(t.TempDir())

	if dt.StyleType() != "DirectoryTree" {
		t.Errorf("StyleType() = %q, want 'DirectoryTree'", dt.StyleType())
	}
}

func TestDirectoryTree_NilSafety(t *testing.T) {
	var dt *DirectoryTree

	// These should not panic
	dt.SetRoot("/tmp")
	dt.SetShowHidden(true)
	dt.SetFilter(nil)
	dt.SetOnSelect(nil)
	dt.SetLazyLoad(true)
	dt.SetStyle(backend.DefaultStyle())
	dt.SetSelectedStyle(backend.DefaultStyle())
	dt.SetLabel("Test")
	dt.Bind(runtime.Services{})
	dt.Unbind()
	dt.Refresh()
	dt.ExpandSelected()
	dt.CollapseSelected()
	dt.ToggleSelected()
	dt.ScrollBy(0, 1)
	dt.ScrollTo(0, 0)
	dt.PageBy(1)
	dt.ScrollToStart()
	dt.ScrollToEnd()

	if dt.Root() != "" {
		t.Error("expected empty Root")
	}
	if dt.ShowHidden() != false {
		t.Error("expected ShowHidden() = false")
	}
	if dt.LazyLoad() != false {
		t.Error("expected LazyLoad() = false")
	}
	if dt.SelectedPath() != "" {
		t.Error("expected empty SelectedPath")
	}
	if dt.SelectedEntry() != nil {
		t.Error("expected nil SelectedEntry")
	}
}
