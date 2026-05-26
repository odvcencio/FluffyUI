package widgets

import (
	"testing"

	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/terminal"
)

func TestSidebarNew(t *testing.T) {
	s := NewSidebar(
		SidebarItem{Label: "Home", Icon: "H"},
		SidebarItem{Label: "Settings"},
	)
	if s == nil {
		t.Fatal("NewSidebar returned nil")
	}
	if s.StyleType() != "Sidebar" {
		t.Errorf("StyleType() = %q, want %q", s.StyleType(), "Sidebar")
	}
	if !s.Expanded() {
		t.Error("Sidebar should be expanded by default")
	}
	if s.Selected() != 0 {
		t.Errorf("Initial Selected = %d, want 0", s.Selected())
	}
}

func TestSidebarMeasure(t *testing.T) {
	s := NewSidebar(
		SidebarItem{Label: "Home"},
		SidebarItem{Label: "About"},
		SidebarItem{Label: "Contact"},
	)

	constraints := runtime.Constraints{MinWidth: 0, MinHeight: 0, MaxWidth: 30, MaxHeight: 10}
	size := s.Measure(constraints)

	if size.Height != 3 {
		t.Errorf("Height = %d, want 3", size.Height)
	}
	if size.Width != 30 {
		t.Errorf("Width = %d, want 30 (fill available)", size.Width)
	}
}

func TestSidebarMeasureFixedWidth(t *testing.T) {
	s := NewSidebar(SidebarItem{Label: "Home"})
	s.SetWidth(20)

	constraints := runtime.Constraints{MinWidth: 0, MinHeight: 0, MaxWidth: 80, MaxHeight: 10}
	size := s.Measure(constraints)

	if size.Width != 20 {
		t.Errorf("Width = %d, want 20 (fixed)", size.Width)
	}
}

func TestSidebarMeasureCollapsed(t *testing.T) {
	s := NewSidebar(SidebarItem{Label: "Home"})
	s.SetExpanded(false)

	constraints := runtime.Constraints{MinWidth: 0, MinHeight: 0, MaxWidth: 30, MaxHeight: 10}
	size := s.Measure(constraints)

	if size.Width != 0 || size.Height != 0 {
		t.Errorf("Collapsed size = %dx%d, want 0x0", size.Width, size.Height)
	}
}

func TestSidebarKeyboardUpDown(t *testing.T) {
	s := NewSidebar(
		SidebarItem{Label: "A"},
		SidebarItem{Label: "B"},
		SidebarItem{Label: "C"},
	)
	s.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 10})
	s.Focus()

	if s.Selected() != 0 {
		t.Fatalf("Initial selected = %d, want 0", s.Selected())
	}

	// Down
	result := s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if !result.Handled {
		t.Error("Down key should be handled")
	}
	if s.Selected() != 1 {
		t.Errorf("After Down, selected = %d, want 1", s.Selected())
	}

	// Down again
	s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if s.Selected() != 2 {
		t.Errorf("After second Down, selected = %d, want 2", s.Selected())
	}

	// Down at end - clamp
	s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if s.Selected() != 2 {
		t.Errorf("Down at end should stay 2, got %d", s.Selected())
	}

	// Up
	s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if s.Selected() != 1 {
		t.Errorf("After Up, selected = %d, want 1", s.Selected())
	}

	// Up at start
	s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if s.Selected() != 0 {
		t.Errorf("Up at start should stay 0, got %d", s.Selected())
	}
}

func TestSidebarHomeEnd(t *testing.T) {
	s := NewSidebar(
		SidebarItem{Label: "A"},
		SidebarItem{Label: "B"},
		SidebarItem{Label: "C"},
	)
	s.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 10})
	s.Focus()

	// End
	s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnd})
	if s.Selected() != 2 {
		t.Errorf("After End, selected = %d, want 2", s.Selected())
	}

	// Home
	s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyHome})
	if s.Selected() != 0 {
		t.Errorf("After Home, selected = %d, want 0", s.Selected())
	}
}

func TestSidebarEnterCallsOnSelect(t *testing.T) {
	activated := ""
	s := NewSidebar(
		SidebarItem{Label: "Home", OnSelect: func() { activated = "home" }},
		SidebarItem{Label: "Settings", OnSelect: func() { activated = "settings" }},
	)
	s.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 10})
	s.Focus()

	// Move to Settings
	s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})

	// Enter
	result := s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if !result.Handled {
		t.Error("Enter should be handled")
	}
	if activated != "settings" {
		t.Errorf("activated = %q, want %q", activated, "settings")
	}
}

func TestSidebarExpandCollapse(t *testing.T) {
	s := NewSidebar(
		SidebarItem{
			Label: "Docs",
			Children: []SidebarItem{
				{Label: "API"},
				{Label: "Guides"},
			},
		},
		SidebarItem{Label: "About"},
	)
	s.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 10})
	s.Focus()

	// Initially Docs is not expanded, so flat list = [Docs, About]
	rows := s.flatten()
	if len(rows) != 2 {
		t.Fatalf("Flat rows = %d, want 2 (unexpanded)", len(rows))
	}

	// Press Right to expand Docs
	s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	s.flatDirty = true // force re-flatten
	rows = s.flatten()
	if len(rows) != 4 {
		t.Fatalf("Flat rows after expand = %d, want 4", len(rows))
	}
	if rows[1].item.Label != "API" {
		t.Errorf("Row 1 = %q, want %q", rows[1].item.Label, "API")
	}

	// Press Left to collapse Docs
	s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})
	s.flatDirty = true
	rows = s.flatten()
	if len(rows) != 2 {
		t.Fatalf("Flat rows after collapse = %d, want 2", len(rows))
	}
}

func TestSidebarEnterToggleExpand(t *testing.T) {
	s := NewSidebar(
		SidebarItem{
			Label: "Docs",
			Children: []SidebarItem{
				{Label: "API"},
			},
		},
	)
	s.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 10})
	s.Focus()

	// Enter on parent should toggle expand
	s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	s.flatDirty = true
	rows := s.flatten()
	if len(rows) != 2 {
		t.Fatalf("After Enter expand, flat rows = %d, want 2", len(rows))
	}

	// Enter again to collapse
	s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	s.flatDirty = true
	rows = s.flatten()
	if len(rows) != 1 {
		t.Fatalf("After Enter collapse, flat rows = %d, want 1", len(rows))
	}
}

func TestSidebarNestedNavigation(t *testing.T) {
	s := NewSidebar(
		SidebarItem{
			Label:    "Docs",
			Expanded: true,
			Children: []SidebarItem{
				{Label: "API"},
				{Label: "Guides"},
			},
		},
		SidebarItem{Label: "About"},
	)
	s.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 10})
	s.Focus()

	// Flat: [Docs, API, Guides, About]
	rows := s.flatten()
	if len(rows) != 4 {
		t.Fatalf("Flat rows = %d, want 4", len(rows))
	}

	// Navigate down to "Guides" (index 2)
	s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if s.Selected() != 2 {
		t.Errorf("selected = %d, want 2", s.Selected())
	}
	item := s.SelectedItem()
	if item == nil || item.Label != "Guides" {
		t.Errorf("SelectedItem = %v, want Guides", item)
	}
}

func TestSidebarIcon(t *testing.T) {
	s := NewSidebar(
		SidebarItem{Label: "Home", Icon: "H"},
	)
	rows := s.flatten()
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(rows))
	}
	if rows[0].item.Icon != "H" {
		t.Errorf("Icon = %q, want %q", rows[0].item.Icon, "H")
	}
}

func TestSidebarBadge(t *testing.T) {
	s := NewSidebar(
		SidebarItem{Label: "Inbox", Badge: "5"},
	)
	rows := s.flatten()
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(rows))
	}
	if rows[0].item.Badge != "5" {
		t.Errorf("Badge = %q, want %q", rows[0].item.Badge, "5")
	}
}

func TestSidebarWidthAccessor(t *testing.T) {
	s := NewSidebar()
	if s.Width() != 0 {
		t.Errorf("Default Width = %d, want 0", s.Width())
	}
	s.SetWidth(25)
	if s.Width() != 25 {
		t.Errorf("Width = %d, want 25", s.Width())
	}
	s.SetWidth(-1) // invalid
	if s.Width() != 0 {
		t.Errorf("Width after invalid set = %d, want 0", s.Width())
	}
}

func TestSidebarCollapsedIgnoresKeys(t *testing.T) {
	s := NewSidebar(
		SidebarItem{Label: "Home"},
	)
	s.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 10})
	s.Focus()
	s.SetExpanded(false)

	result := s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if result.Handled {
		t.Error("Collapsed sidebar should not handle keys")
	}
}

func TestSidebarUnfocusedIgnoresKeys(t *testing.T) {
	s := NewSidebar(
		SidebarItem{Label: "Home"},
	)
	s.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 10})
	// Not focused

	result := s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if result.Handled {
		t.Error("Unfocused sidebar should not handle keys")
	}
}

func TestSidebarA11y(t *testing.T) {
	s := NewSidebar(
		SidebarItem{Label: "Home"},
		SidebarItem{Label: "Settings"},
	)
	s.syncA11y()

	if s.Base.Role != "navigation" {
		t.Errorf("Role = %q, want %q", s.Base.Role, "navigation")
	}
	if s.Base.Label != "Sidebar" {
		t.Errorf("Label = %q, want %q", s.Base.Label, "Sidebar")
	}
	if s.Base.Orientation != "vertical" {
		t.Errorf("Orientation = %q, want %q", s.Base.Orientation, "vertical")
	}
	if s.Base.Description != "2 items" {
		t.Errorf("Description = %q, want %q", s.Base.Description, "2 items")
	}
}

func TestSidebarA11yCustomLabel(t *testing.T) {
	s := NewSidebar()
	s.SetLabel("Main Navigation")
	if s.Base.Label != "Main Navigation" {
		t.Errorf("Label = %q, want %q", s.Base.Label, "Main Navigation")
	}
}

func TestSidebarA11yExpandedState(t *testing.T) {
	s := NewSidebar(
		SidebarItem{
			Label:    "Docs",
			Expanded: true,
			Children: []SidebarItem{
				{Label: "API"},
			},
		},
	)
	s.syncA11y()

	// Selected is Docs which has children and is expanded
	if s.Base.State.Expanded == nil || !*s.Base.State.Expanded {
		t.Error("Expected State.Expanded = true for parent item")
	}
}

func TestSidebarA11yBadgeInValue(t *testing.T) {
	s := NewSidebar(
		SidebarItem{Label: "Inbox", Badge: "3"},
	)
	s.syncA11y()

	if s.Base.Value == nil {
		t.Fatal("Value should not be nil")
	}
	if s.Base.Value.Text != "Inbox (3)" {
		t.Errorf("Value.Text = %q, want %q", s.Base.Value.Text, "Inbox (3)")
	}
}

func TestSidebarBindUnbind(t *testing.T) {
	s := NewSidebar()
	svc := runtime.Services{}
	s.Bind(svc)
	s.Unbind()
	// Should not panic
}

func TestSidebarNilGuards(t *testing.T) {
	var s *Sidebar

	if s.Selected() != 0 {
		t.Error("nil Selected should return 0")
	}
	if s.Expanded() {
		t.Error("nil Expanded should return false")
	}
	if s.Width() != 0 {
		t.Error("nil Width should return 0")
	}
	if s.SelectedItem() != nil {
		t.Error("nil SelectedItem should return nil")
	}
	if s.GetItems() != nil {
		t.Error("nil GetItems should return nil")
	}

	// Should not panic
	s.SetItems()
	s.SetWidth(10)
	s.SetExpanded(true)
	s.SetLabel("X")
	s.SetStyle(backend.DefaultStyle())
	s.SetSelectedStyle(backend.DefaultStyle())
	s.ScrollBy(0, 1)
	s.ScrollTo(0, 0)
	s.ScrollToStart()
	s.ScrollToEnd()
	s.PageBy(1)
	s.Bind(runtime.Services{})
	s.Unbind()
	s.Render(runtime.RenderContext{})
	s.syncA11y()
}

func TestSidebarEmptyHandleMessage(t *testing.T) {
	s := NewSidebar()
	s.Focus()
	result := s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if result.Handled {
		t.Error("Empty sidebar should not handle key messages")
	}
}

func TestSidebarSelectedItem(t *testing.T) {
	s := NewSidebar(
		SidebarItem{Label: "Home"},
		SidebarItem{Label: "Settings"},
	)
	s.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 10})
	s.Focus()

	item := s.SelectedItem()
	if item == nil {
		t.Fatal("SelectedItem should not be nil")
	}
	if item.Label != "Home" {
		t.Errorf("SelectedItem = %q, want %q", item.Label, "Home")
	}

	s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	item = s.SelectedItem()
	if item == nil || item.Label != "Settings" {
		t.Errorf("SelectedItem after Down = %v, want Settings", item)
	}
}

func TestSidebarSetItems(t *testing.T) {
	s := NewSidebar(SidebarItem{Label: "A"})
	s.SetItems(SidebarItem{Label: "X"}, SidebarItem{Label: "Y"})

	items := s.GetItems()
	if len(items) != 2 {
		t.Fatalf("len = %d, want 2", len(items))
	}
	if items[0].Label != "X" {
		t.Errorf("items[0] = %q, want %q", items[0].Label, "X")
	}
}

func TestSidebarDepthIndent(t *testing.T) {
	s := NewSidebar()
	if s.indent(0) != "" {
		t.Errorf("indent(0) = %q, want %q", s.indent(0), "")
	}
	if s.indent(1) != "  " {
		t.Errorf("indent(1) = %q, want %q", s.indent(1), "  ")
	}
	if s.indent(3) != "      " {
		t.Errorf("indent(3) = %q, want %q", s.indent(3), "      ")
	}
}

func TestSidebarLeftOnLeafNoOp(t *testing.T) {
	s := NewSidebar(
		SidebarItem{Label: "Leaf"},
	)
	s.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 10})
	s.Focus()

	result := s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})
	if !result.Handled {
		t.Error("Left should still be handled (just a no-op for leaves)")
	}
}

func TestSidebarRightOnLeafNoOp(t *testing.T) {
	s := NewSidebar(
		SidebarItem{Label: "Leaf"},
	)
	s.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 10})
	s.Focus()

	result := s.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if !result.Handled {
		t.Error("Right should still be handled (just a no-op for leaves)")
	}
}

func TestSidebarScrollBy(t *testing.T) {
	s := NewSidebar(
		SidebarItem{Label: "A"},
		SidebarItem{Label: "B"},
		SidebarItem{Label: "C"},
	)
	s.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 2})

	s.ScrollBy(0, 2)
	if s.Selected() != 2 {
		t.Errorf("After ScrollBy(0,2), selected = %d, want 2", s.Selected())
	}
}
