package widgets

import (
	"testing"

	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
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
