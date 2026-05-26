package widgets

import (
	"testing"

	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/terminal"
)

func TestBreadcrumbMeasure(t *testing.T) {
	bc := NewBreadcrumb(
		BreadcrumbItem{Label: "Home"},
		BreadcrumbItem{Label: "Docs"},
		BreadcrumbItem{Label: "API"},
	)

	constraints := runtime.Constraints{MinWidth: 0, MinHeight: 0, MaxWidth: 100, MaxHeight: 10}
	size := bc.Measure(constraints)

	// "Home" + " > " + "Docs" + " > " + "API" = 4 + 3 + 4 + 3 + 3 = 17
	expectedWidth := 17
	if size.Width != expectedWidth {
		t.Errorf("Width = %d, want %d", size.Width, expectedWidth)
	}
	if size.Height != 1 {
		t.Errorf("Height = %d, want 1", size.Height)
	}
}

func TestBreadcrumbItemAtPosition(t *testing.T) {
	bc := NewBreadcrumb(
		BreadcrumbItem{Label: "Home"},
		BreadcrumbItem{Label: "Docs"},
		BreadcrumbItem{Label: "API"},
	)

	// Layout the breadcrumb
	bc.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 1})

	// Test positions
	// "Home" at 0-3
	// " > " at 4-6
	// "Docs" at 7-10
	// " > " at 11-13
	// "API" at 14-16
	tests := []struct {
		x, y      int
		wantIndex int
	}{
		{0, 0, 0},   // Start of "Home"
		{3, 0, 0},   // End of "Home"
		{4, 0, -1},  // Separator
		{7, 0, 1},   // Start of "Docs"
		{10, 0, 1},  // End of "Docs"
		{14, 0, 2},  // Start of "API"
		{16, 0, 2},  // End of "API"
		{17, 0, -1}, // Past end
		{0, 1, -1},  // Wrong row
	}

	for _, tt := range tests {
		got := bc.itemAtPosition(tt.x, tt.y)
		if got != tt.wantIndex {
			t.Errorf("itemAtPosition(%d, %d) = %d, want %d", tt.x, tt.y, got, tt.wantIndex)
		}
	}
}

func TestBreadcrumbClickHandler(t *testing.T) {
	clicked := -1
	bc := NewBreadcrumb(
		BreadcrumbItem{Label: "Home", OnClick: func() { clicked = 0 }},
		BreadcrumbItem{Label: "Docs", OnClick: func() { clicked = 1 }},
		BreadcrumbItem{Label: "API", OnClick: func() { clicked = 2 }},
	)

	bc.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 1})

	// Click on "Docs" (at position 7)
	msg := runtime.MouseMsg{X: 7, Y: 0, Button: runtime.MouseLeft, Action: runtime.MousePress}
	result := bc.HandleMessage(msg)

	if !result.Handled {
		t.Error("Click should be handled")
	}
	if clicked != 1 {
		t.Errorf("Clicked = %d, want 1 (Docs)", clicked)
	}
}

func TestBreadcrumbOnNavigate(t *testing.T) {
	navigated := -1
	bc := NewBreadcrumb(
		BreadcrumbItem{Label: "Home"},
		BreadcrumbItem{Label: "Docs"},
		BreadcrumbItem{Label: "API"},
	)
	bc.OnNavigate(func(index int) { navigated = index })

	bc.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 1})

	// Click on "API" (at position 14)
	msg := runtime.MouseMsg{X: 14, Y: 0, Button: runtime.MouseLeft, Action: runtime.MousePress}
	result := bc.HandleMessage(msg)

	if !result.Handled {
		t.Error("Click should be handled")
	}
	if navigated != 2 {
		t.Errorf("Navigated = %d, want 2 (API)", navigated)
	}
}

func TestBreadcrumbKeyboardNavigation(t *testing.T) {
	bc := NewBreadcrumb(
		BreadcrumbItem{Label: "Home"},
		BreadcrumbItem{Label: "Docs"},
		BreadcrumbItem{Label: "API"},
	)

	bc.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 1})
	bc.Focus()

	// Start at 0
	if bc.Selected() != 0 {
		t.Errorf("Initial selected = %d, want 0", bc.Selected())
	}

	// Press Right
	result := bc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if !result.Handled {
		t.Error("Right key should be handled")
	}
	if bc.Selected() != 1 {
		t.Errorf("After Right, selected = %d, want 1", bc.Selected())
	}

	// Press Right again
	bc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if bc.Selected() != 2 {
		t.Errorf("After second Right, selected = %d, want 2", bc.Selected())
	}

	// Press Right at end - should stay at 2
	bc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if bc.Selected() != 2 {
		t.Errorf("Right at end should stay at 2, got %d", bc.Selected())
	}

	// Press Left
	bc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})
	if bc.Selected() != 1 {
		t.Errorf("After Left, selected = %d, want 1", bc.Selected())
	}

	// Press Home
	bc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyHome})
	if bc.Selected() != 0 {
		t.Errorf("After Home, selected = %d, want 0", bc.Selected())
	}

	// Press End
	bc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnd})
	if bc.Selected() != 2 {
		t.Errorf("After End, selected = %d, want 2", bc.Selected())
	}
}

func TestBreadcrumbEnterKey(t *testing.T) {
	activated := -1
	bc := NewBreadcrumb(
		BreadcrumbItem{Label: "Home", OnClick: func() { activated = 0 }},
		BreadcrumbItem{Label: "Docs", OnClick: func() { activated = 1 }},
	)

	bc.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 1})
	bc.Focus()
	bc.selected = 1 // Select "Docs"

	result := bc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if !result.Handled {
		t.Error("Enter key should be handled")
	}
	if activated != 1 {
		t.Errorf("Activated = %d, want 1 (Docs)", activated)
	}
}

func TestBreadcrumbCustomSeparator(t *testing.T) {
	bc := NewBreadcrumb(
		BreadcrumbItem{Label: "A"},
		BreadcrumbItem{Label: "B"},
	)
	bc.SetSeparator(" / ")

	constraints := runtime.Constraints{MinWidth: 0, MinHeight: 0, MaxWidth: 100, MaxHeight: 10}
	size := bc.Measure(constraints)

	// "A" + " / " + "B" = 1 + 3 + 1 = 5
	if size.Width != 5 {
		t.Errorf("Width with custom separator = %d, want 5", size.Width)
	}
}

func TestBreadcrumbStyleType(t *testing.T) {
	bc := NewBreadcrumb()
	if got := bc.StyleType(); got != "Breadcrumb" {
		t.Errorf("StyleType() = %q, want %q", got, "Breadcrumb")
	}
}

func TestBreadcrumbCollapsible(t *testing.T) {
	bc := NewBreadcrumb(
		BreadcrumbItem{Label: "Home"},
		BreadcrumbItem{Label: "Products"},
		BreadcrumbItem{Label: "Electronics"},
		BreadcrumbItem{Label: "Phones"},
		BreadcrumbItem{Label: "Details"},
	)

	// Not collapsible by default
	if bc.Collapsible() {
		t.Error("Breadcrumb should not be collapsible by default")
	}

	// Enable collapsing
	bc.SetCollapsible(true)
	if !bc.Collapsible() {
		t.Error("Breadcrumb should be collapsible after SetCollapsible(true)")
	}

	// Default thresholds
	if bc.CollapseThreshold() != 4 {
		t.Errorf("Default CollapseThreshold = %d, want 4", bc.CollapseThreshold())
	}
	if bc.CollapseKeep() != 2 {
		t.Errorf("Default CollapseKeep = %d, want 2", bc.CollapseKeep())
	}
}

func TestBreadcrumbNeedsCollapse(t *testing.T) {
	bc := NewBreadcrumb(
		BreadcrumbItem{Label: "Home"},
		BreadcrumbItem{Label: "Products"},
		BreadcrumbItem{Label: "Electronics"},
		BreadcrumbItem{Label: "Phones"},
		BreadcrumbItem{Label: "Details"},
	)
	bc.SetCollapsible(true)

	// Full width: "Home" + " > " + "Products" + " > " + "Electronics" + " > " + "Phones" + " > " + "Details"
	// = 4 + 3 + 8 + 3 + 11 + 3 + 6 + 3 + 7 = 48
	fullWidth := bc.fullWidth()
	if fullWidth != 48 {
		t.Errorf("fullWidth = %d, want 48", fullWidth)
	}

	// Should not collapse when wide enough
	if bc.needsCollapse(100) {
		t.Error("Should not collapse when width=100")
	}

	// Should collapse when narrow
	if !bc.needsCollapse(30) {
		t.Error("Should collapse when width=30")
	}
}

func TestBreadcrumbCollapsedIndices(t *testing.T) {
	bc := NewBreadcrumb(
		BreadcrumbItem{Label: "Home"},     // 0
		BreadcrumbItem{Label: "Products"}, // 1
		BreadcrumbItem{Label: "Elec"},     // 2
		BreadcrumbItem{Label: "Phones"},   // 3
		BreadcrumbItem{Label: "Details"},  // 4
	)
	bc.SetCollapsible(true)
	bc.SetCollapseKeep(2)

	indices := bc.collapsedIndices()
	// Should be: [0, -1, 3, 4]  (first, ellipsis, last 2)
	expected := []int{0, -1, 3, 4}
	if len(indices) != len(expected) {
		t.Fatalf("collapsedIndices len = %d, want %d: %v", len(indices), len(expected), indices)
	}
	for i, v := range indices {
		if v != expected[i] {
			t.Errorf("collapsedIndices[%d] = %d, want %d", i, v, expected[i])
		}
	}
}

func TestBreadcrumbCollapsedIndicesKeep3(t *testing.T) {
	bc := NewBreadcrumb(
		BreadcrumbItem{Label: "A"},
		BreadcrumbItem{Label: "B"},
		BreadcrumbItem{Label: "C"},
		BreadcrumbItem{Label: "D"},
		BreadcrumbItem{Label: "E"},
		BreadcrumbItem{Label: "F"},
	)
	bc.SetCollapsible(true)
	bc.SetCollapseKeep(3)

	indices := bc.collapsedIndices()
	// Should be: [0, -1, 3, 4, 5]
	expected := []int{0, -1, 3, 4, 5}
	if len(indices) != len(expected) {
		t.Fatalf("collapsedIndices len = %d, want %d: %v", len(indices), len(expected), indices)
	}
	for i, v := range indices {
		if v != expected[i] {
			t.Errorf("collapsedIndices[%d] = %d, want %d", i, v, expected[i])
		}
	}
}

func TestBreadcrumbCollapseThresholdBelowMinimum(t *testing.T) {
	bc := NewBreadcrumb(
		BreadcrumbItem{Label: "Home"},
		BreadcrumbItem{Label: "About"},
		BreadcrumbItem{Label: "Contact"},
	)
	bc.SetCollapsible(true)

	// 3 items is below default threshold of 4, should not collapse
	if bc.needsCollapse(5) {
		t.Error("Should not collapse when item count < threshold")
	}
}

func TestBreadcrumbSetCollapseThreshold(t *testing.T) {
	bc := NewBreadcrumb()
	bc.SetCollapseThreshold(6)
	if bc.CollapseThreshold() != 6 {
		t.Errorf("CollapseThreshold = %d, want 6", bc.CollapseThreshold())
	}

	// Invalid: less than 2
	bc.SetCollapseThreshold(1)
	if bc.CollapseThreshold() != 6 {
		t.Errorf("CollapseThreshold should remain 6 after invalid set, got %d", bc.CollapseThreshold())
	}
}

func TestBreadcrumbSetCollapseKeep(t *testing.T) {
	bc := NewBreadcrumb()
	bc.SetCollapseKeep(3)
	if bc.CollapseKeep() != 3 {
		t.Errorf("CollapseKeep = %d, want 3", bc.CollapseKeep())
	}

	// Invalid: less than 1
	bc.SetCollapseKeep(0)
	if bc.CollapseKeep() != 3 {
		t.Errorf("CollapseKeep should remain 3 after invalid set, got %d", bc.CollapseKeep())
	}
}

func TestBreadcrumbARIACurrent(t *testing.T) {
	bc := NewBreadcrumb(
		BreadcrumbItem{Label: "Home"},
		BreadcrumbItem{Label: "Docs"},
	)
	bc.syncA11y()

	if bc.Base.Current != "page" {
		t.Errorf("Current = %q, want %q", bc.Base.Current, "page")
	}
}

func TestBreadcrumbARIACurrentEmpty(t *testing.T) {
	bc := NewBreadcrumb()
	bc.syncA11y()

	if bc.Base.Current != "" {
		t.Errorf("Current should be empty for no items, got %q", bc.Base.Current)
	}
}

func TestBreadcrumbARIANavRole(t *testing.T) {
	bc := NewBreadcrumb(BreadcrumbItem{Label: "Home"})
	bc.syncA11y()

	if bc.Base.Role != "navigation" {
		t.Errorf("Role = %q, want %q", bc.Base.Role, "navigation")
	}
	if bc.Base.Label != "Breadcrumbs" {
		t.Errorf("Label = %q, want %q", bc.Base.Label, "Breadcrumbs")
	}
}

func TestBreadcrumbBindUnbind(t *testing.T) {
	bc := NewBreadcrumb()
	svc := runtime.Services{}
	bc.Bind(svc)
	bc.Unbind()
	// Should not panic
}

func TestBreadcrumbNilGuards(t *testing.T) {
	var bc *Breadcrumb

	// All methods should handle nil gracefully
	if bc.Selected() != 0 {
		t.Error("nil Selected should return 0")
	}
	if bc.Collapsible() {
		t.Error("nil Collapsible should return false")
	}
	if bc.CollapseThreshold() != 4 {
		t.Errorf("nil CollapseThreshold = %d, want 4", bc.CollapseThreshold())
	}
	if bc.CollapseKeep() != 2 {
		t.Errorf("nil CollapseKeep = %d, want 2", bc.CollapseKeep())
	}

	bc.SetSeparator("/")
	bc.SetCollapsible(true)
	bc.SetCollapseThreshold(5)
	bc.SetCollapseKeep(3)
	bc.SetOnNavigate(nil)
	bc.Bind(runtime.Services{})
	bc.Unbind()
	bc.Render(runtime.RenderContext{})
	bc.syncA11y()
}

func TestBreadcrumbCollapsedClickItemAtPosition(t *testing.T) {
	bc := NewBreadcrumb(
		BreadcrumbItem{Label: "Home"},     // 0
		BreadcrumbItem{Label: "Products"}, // 1
		BreadcrumbItem{Label: "Elec"},     // 2
		BreadcrumbItem{Label: "Phones"},   // 3
		BreadcrumbItem{Label: "Details"},  // 4
	)
	bc.SetCollapsible(true)
	bc.SetCollapseKeep(2)

	// Layout with narrow width to trigger collapse
	bc.Layout(runtime.Rect{X: 0, Y: 0, Width: 30, Height: 1})

	// When collapsed: "Home" + " > " + "..." + " > " + "Phones" + " > " + "Details"
	// Positions:       0-3     4-6     7       8-10    11-16      17-19   20-26
	// The ellipsis is 1 char wide (unicode ...)

	// Clicking on "Home" (x=0) should return original index 0
	got := bc.itemAtPosition(0, 0)
	if got != 0 {
		t.Errorf("Click at x=0 should hit item 0, got %d", got)
	}

	// Clicking on the ellipsis should return -1 (not a real item)
	gotEllipsis := bc.itemAtPosition(7, 0)
	if gotEllipsis != -1 {
		t.Errorf("Click on ellipsis should return -1, got %d", gotEllipsis)
	}
}

func TestBreadcrumbPathString(t *testing.T) {
	bc := NewBreadcrumb(
		BreadcrumbItem{Label: "Home"},
		BreadcrumbItem{Label: "  "},
		BreadcrumbItem{Label: "Docs"},
	)
	path := bc.pathString()
	// Blank labels are skipped
	if path != "Home > Docs" {
		t.Errorf("pathString = %q, want %q", path, "Home > Docs")
	}
}

func TestBreadcrumbEmptyItems(t *testing.T) {
	bc := NewBreadcrumb()
	result := bc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if result.Handled {
		t.Error("Empty breadcrumb should not handle messages")
	}
}

func TestBreadcrumbUnfocusedKeysIgnored(t *testing.T) {
	bc := NewBreadcrumb(
		BreadcrumbItem{Label: "A"},
		BreadcrumbItem{Label: "B"},
	)
	bc.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 1})
	// Not focused
	result := bc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if result.Handled {
		t.Error("Unfocused breadcrumb should not handle key messages")
	}
}
