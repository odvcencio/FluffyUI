package widgets

import (
	"strings"
	"testing"

	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/terminal"
)

func TestStatusBar_NewStatusBar(t *testing.T) {
	sb := NewStatusBar(
		StatusBarItem{Key: "branch", Text: "main", Alignment: StatusBarLeft},
		StatusBarItem{Key: "errors", Text: "0 errors", Alignment: StatusBarRight},
	)
	if sb == nil {
		t.Fatal("expected non-nil status bar")
	}
	items := sb.Items()
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Key != "branch" {
		t.Fatalf("expected first item key 'branch', got %q", items[0].Key)
	}
	if items[1].Key != "errors" {
		t.Fatalf("expected second item key 'errors', got %q", items[1].Key)
	}
	// All items should be visible by default.
	for i, item := range items {
		if !item.Visible {
			t.Fatalf("expected item %d to be visible", i)
		}
	}
}

func TestStatusBar_NewStatusBarEmpty(t *testing.T) {
	sb := NewStatusBar()
	if sb == nil {
		t.Fatal("expected non-nil status bar")
	}
	if len(sb.Items()) != 0 {
		t.Fatalf("expected 0 items, got %d", len(sb.Items()))
	}
}

func TestStatusBar_AddItem(t *testing.T) {
	sb := NewStatusBar()
	sb.AddItem(StatusBarItem{Key: "line", Text: "Ln 42", Alignment: StatusBarRight})
	items := sb.Items()
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Key != "line" {
		t.Fatalf("expected key 'line', got %q", items[0].Key)
	}
	if items[0].Text != "Ln 42" {
		t.Fatalf("expected text 'Ln 42', got %q", items[0].Text)
	}
	if !items[0].Visible {
		t.Fatal("expected added item to be visible")
	}
}

func TestStatusBar_RemoveItem(t *testing.T) {
	sb := NewStatusBar(
		StatusBarItem{Key: "a", Text: "A"},
		StatusBarItem{Key: "b", Text: "B"},
		StatusBarItem{Key: "c", Text: "C"},
	)
	sb.RemoveItem("b")
	items := sb.Items()
	if len(items) != 2 {
		t.Fatalf("expected 2 items after removal, got %d", len(items))
	}
	for _, item := range items {
		if item.Key == "b" {
			t.Fatal("item 'b' should have been removed")
		}
	}
}

func TestStatusBar_RemoveItemClampsSelected(t *testing.T) {
	sb := NewStatusBar(
		StatusBarItem{Key: "a", Text: "A"},
		StatusBarItem{Key: "b", Text: "B"},
	)
	sb.Focus()
	// Select the last item.
	sb.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if sb.Selected() != 1 {
		t.Fatalf("expected selected 1, got %d", sb.Selected())
	}
	// Remove last item; selected should clamp.
	sb.RemoveItem("b")
	if sb.Selected() != 0 {
		t.Fatalf("expected selected clamped to 0, got %d", sb.Selected())
	}
}

func TestStatusBar_UpdateItem(t *testing.T) {
	sb := NewStatusBar(
		StatusBarItem{Key: "branch", Text: "main"},
	)
	sb.UpdateItem("branch", "develop")
	items := sb.Items()
	if items[0].Text != "develop" {
		t.Fatalf("expected text 'develop', got %q", items[0].Text)
	}
}

func TestStatusBar_UpdateItemNonexistent(t *testing.T) {
	sb := NewStatusBar(
		StatusBarItem{Key: "branch", Text: "main"},
	)
	// Should not panic.
	sb.UpdateItem("nonexistent", "test")
	items := sb.Items()
	if items[0].Text != "main" {
		t.Fatalf("expected text unchanged, got %q", items[0].Text)
	}
}

func TestStatusBar_SetItemVisible(t *testing.T) {
	sb := NewStatusBar(
		StatusBarItem{Key: "a", Text: "A"},
		StatusBarItem{Key: "b", Text: "B"},
	)
	sb.SetItemVisible("a", false)
	visible := sb.visibleItems()
	if len(visible) != 1 {
		t.Fatalf("expected 1 visible item, got %d", len(visible))
	}
	if visible[0].Key != "b" {
		t.Fatalf("expected visible item key 'b', got %q", visible[0].Key)
	}

	// Make it visible again.
	sb.SetItemVisible("a", true)
	visible = sb.visibleItems()
	if len(visible) != 2 {
		t.Fatalf("expected 2 visible items, got %d", len(visible))
	}
}

func TestStatusBar_SetBackground(t *testing.T) {
	sb := NewStatusBar()
	custom := backend.DefaultStyle().Reverse(true)
	sb.SetBackground(custom)
	if sb.bgStyle != custom {
		t.Fatal("expected background style to be updated")
	}
}

func TestStatusBar_Measure(t *testing.T) {
	sb := NewStatusBar(
		StatusBarItem{Key: "a", Text: "Hello"},
		StatusBarItem{Key: "b", Text: "World"},
	)
	size := sb.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 10})
	// "Hello | World" = 5 + 3 + 5 = 13
	if size.Width != 13 {
		t.Fatalf("expected width 13, got %d", size.Width)
	}
	if size.Height != 1 {
		t.Fatalf("expected height 1, got %d", size.Height)
	}
}

func TestStatusBar_MeasureEmpty(t *testing.T) {
	sb := NewStatusBar()
	size := sb.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 10})
	if size.Width < 1 {
		t.Fatalf("expected minimum width 1, got %d", size.Width)
	}
	if size.Height != 1 {
		t.Fatalf("expected height 1, got %d", size.Height)
	}
}

func TestStatusBar_RenderLeftRight(t *testing.T) {
	sb := NewStatusBar(
		StatusBarItem{Key: "branch", Text: "main", Alignment: StatusBarLeft},
		StatusBarItem{Key: "errors", Text: "0 err", Alignment: StatusBarRight},
	)
	output := renderStatusBar(sb, 40)
	// Left item should appear at the start.
	if !strings.HasPrefix(output, "main") {
		t.Fatalf("expected left item 'main' at start, got %q", output)
	}
	// Right item should appear at the end.
	if !strings.HasSuffix(strings.TrimRight(output, " "), "0 err") {
		t.Fatalf("expected right item '0 err' at end, got %q", output)
	}
}

func TestStatusBar_RenderCenter(t *testing.T) {
	sb := NewStatusBar(
		StatusBarItem{Key: "center", Text: "MID", Alignment: StatusBarCenter},
	)
	output := renderStatusBar(sb, 20)
	// The center item should be somewhere in the middle, not at position 0.
	idx := strings.Index(output, "MID")
	if idx < 0 {
		t.Fatalf("expected output to contain 'MID', got %q", output)
	}
	// Should be roughly centered.
	expectedCenter := (20 - 3) / 2
	if idx < expectedCenter-2 || idx > expectedCenter+2 {
		t.Fatalf("expected 'MID' near center (~%d), found at %d in %q", expectedCenter, idx, output)
	}
}

func TestStatusBar_RenderAllAlignments(t *testing.T) {
	sb := NewStatusBar(
		StatusBarItem{Key: "left", Text: "L", Alignment: StatusBarLeft},
		StatusBarItem{Key: "center", Text: "C", Alignment: StatusBarCenter},
		StatusBarItem{Key: "right", Text: "R", Alignment: StatusBarRight},
	)
	output := renderStatusBar(sb, 30)
	lIdx := strings.Index(output, "L")
	cIdx := strings.Index(output, "C")
	rIdx := strings.LastIndex(output, "R")
	if lIdx < 0 || cIdx < 0 || rIdx < 0 {
		t.Fatalf("expected all items in output, got %q", output)
	}
	if lIdx >= cIdx {
		t.Fatalf("expected left before center: L@%d C@%d", lIdx, cIdx)
	}
	if cIdx >= rIdx {
		t.Fatalf("expected center before right: C@%d R@%d", cIdx, rIdx)
	}
}

func TestStatusBar_KeyboardNavigation(t *testing.T) {
	sb := NewStatusBar(
		StatusBarItem{Key: "a", Text: "A"},
		StatusBarItem{Key: "b", Text: "B"},
		StatusBarItem{Key: "c", Text: "C"},
	)
	sb.Focus()

	// Initial selection at 0.
	if sb.Selected() != 0 {
		t.Fatalf("expected initial selected 0, got %d", sb.Selected())
	}

	// Right arrow moves selection forward.
	result := sb.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if !result.Handled {
		t.Fatal("expected Right to be handled")
	}
	if sb.Selected() != 1 {
		t.Fatalf("expected selected 1 after Right, got %d", sb.Selected())
	}

	// Another right.
	sb.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if sb.Selected() != 2 {
		t.Fatalf("expected selected 2, got %d", sb.Selected())
	}

	// Right at the end should not be handled.
	result = sb.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if result.Handled {
		t.Fatal("expected Right at end to be unhandled")
	}

	// Left arrow.
	result = sb.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})
	if !result.Handled {
		t.Fatal("expected Left to be handled")
	}
	if sb.Selected() != 1 {
		t.Fatalf("expected selected 1 after Left, got %d", sb.Selected())
	}

	// Left at the start should not be handled.
	sb.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})
	result = sb.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})
	if result.Handled {
		t.Fatal("expected Left at start to be unhandled")
	}
}

func TestStatusBar_KeyboardTab(t *testing.T) {
	sb := NewStatusBar(
		StatusBarItem{Key: "a", Text: "A"},
		StatusBarItem{Key: "b", Text: "B"},
	)
	sb.Focus()

	result := sb.HandleMessage(runtime.KeyMsg{Key: terminal.KeyTab})
	if !result.Handled {
		t.Fatal("expected Tab to be handled")
	}
	if sb.Selected() != 1 {
		t.Fatalf("expected selected 1 after Tab, got %d", sb.Selected())
	}
}

func TestStatusBar_KeyboardHomeEnd(t *testing.T) {
	sb := NewStatusBar(
		StatusBarItem{Key: "a", Text: "A"},
		StatusBarItem{Key: "b", Text: "B"},
		StatusBarItem{Key: "c", Text: "C"},
	)
	sb.Focus()

	// End.
	result := sb.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnd})
	if !result.Handled {
		t.Fatal("expected End to be handled")
	}
	if sb.Selected() != 2 {
		t.Fatalf("expected selected 2 after End, got %d", sb.Selected())
	}

	// Home.
	result = sb.HandleMessage(runtime.KeyMsg{Key: terminal.KeyHome})
	if !result.Handled {
		t.Fatal("expected Home to be handled")
	}
	if sb.Selected() != 0 {
		t.Fatalf("expected selected 0 after Home, got %d", sb.Selected())
	}
}

func TestStatusBar_KeyboardEnter(t *testing.T) {
	sb := NewStatusBar(
		StatusBarItem{Key: "a", Text: "A"},
		StatusBarItem{Key: "b", Text: "B", OnClick: nil},
	)
	sb.Focus()

	var clicked bool
	sb.items[0].OnClick = func() { clicked = true }

	result := sb.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if !result.Handled {
		t.Fatal("expected Enter to be handled")
	}
	if !clicked {
		t.Fatal("expected OnClick to be called")
	}
}

func TestStatusBar_UnfocusedNoHandle(t *testing.T) {
	sb := NewStatusBar(
		StatusBarItem{Key: "a", Text: "A"},
	)
	// Do not focus.
	result := sb.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if result.Handled {
		t.Fatal("expected unhandled when not focused")
	}
}

func TestStatusBar_PriorityTruncation(t *testing.T) {
	sb := NewStatusBar(
		StatusBarItem{Key: "important", Text: "IMPORTANT", Priority: 10},
		StatusBarItem{Key: "filler", Text: "FILLER-DATA", Priority: 1},
		StatusBarItem{Key: "critical", Text: "CRIT", Priority: 100},
	)

	// Width enough for all: "IMPORTANT | FILLER-DATA | CRIT" = 9+3+11+3+4 = 30
	fitting := sb.fittingItems(30)
	if len(fitting) != 3 {
		t.Fatalf("expected 3 items when enough space, got %d", len(fitting))
	}

	// Width only enough for two: drop lowest priority ("filler", priority=1).
	// "IMPORTANT | CRIT" = 9+3+4 = 16
	fitting = sb.fittingItems(16)
	if len(fitting) != 2 {
		t.Fatalf("expected 2 items when space limited, got %d", len(fitting))
	}
	for _, item := range fitting {
		if item.Key == "filler" {
			t.Fatal("expected 'filler' (lowest priority) to be dropped")
		}
	}

	// Very narrow: only the highest priority survives.
	fitting = sb.fittingItems(4)
	if len(fitting) != 1 {
		t.Fatalf("expected 1 item when very narrow, got %d", len(fitting))
	}
	if fitting[0].Key != "critical" {
		t.Fatalf("expected 'critical' (highest priority) to survive, got %q", fitting[0].Key)
	}
}

func TestStatusBar_PriorityPreservesOrder(t *testing.T) {
	sb := NewStatusBar(
		StatusBarItem{Key: "a", Text: "AAA", Priority: 5},
		StatusBarItem{Key: "b", Text: "BBB", Priority: 10},
		StatusBarItem{Key: "c", Text: "CCC", Priority: 5},
	)

	// Drop one: lowest priority items are "a" and "c" (both 5).
	// Stable sort drops "a" first.
	fitting := sb.fittingItems(9) // "BBB | CCC" = 3+3+3 = 9
	if len(fitting) != 2 {
		t.Fatalf("expected 2 items, got %d", len(fitting))
	}
	// "b" should come before "c" (original order).
	if fitting[0].Key != "b" || fitting[1].Key != "c" {
		t.Fatalf("expected [b, c], got [%s, %s]", fitting[0].Key, fitting[1].Key)
	}
}

func TestStatusBar_AccessibilityRole(t *testing.T) {
	sb := NewStatusBar(
		StatusBarItem{Key: "a", Text: "A"},
	)
	if sb.AccessibleRole() != "toolbar" {
		t.Fatalf("expected role 'toolbar', got %q", sb.AccessibleRole())
	}
}

func TestStatusBar_AccessibilityLabel(t *testing.T) {
	sb := NewStatusBar()
	if sb.AccessibleLabel() != "Status Bar" {
		t.Fatalf("expected label 'Status Bar', got %q", sb.AccessibleLabel())
	}
}

func TestStatusBar_AccessibilityOrientation(t *testing.T) {
	sb := NewStatusBar()
	if sb.AccessibleOrientation() != "horizontal" {
		t.Fatalf("expected orientation 'horizontal', got %q", sb.AccessibleOrientation())
	}
}

func TestStatusBar_CanFocus(t *testing.T) {
	sb := NewStatusBar()
	if !sb.CanFocus() {
		t.Fatal("expected status bar to be focusable")
	}
}

func TestStatusBar_NilSafety(t *testing.T) {
	var sb *StatusBar
	// These should not panic.
	sb.AddItem(StatusBarItem{})
	sb.RemoveItem("x")
	sb.UpdateItem("x", "y")
	sb.SetItemVisible("x", false)
	sb.SetBackground(backend.DefaultStyle())
	if sb.Items() != nil {
		t.Fatal("expected nil items from nil status bar")
	}
	if sb.Selected() != 0 {
		t.Fatal("expected 0 from nil status bar Selected()")
	}
	result := sb.HandleMessage(runtime.KeyMsg{})
	if result.Handled {
		t.Fatal("expected unhandled from nil status bar")
	}
}

func TestStatusBar_ItemSeparator(t *testing.T) {
	sb := NewStatusBar(
		StatusBarItem{Key: "a", Text: "A"},
		StatusBarItem{Key: "b", Text: "B"},
	)
	output := renderStatusBar(sb, 20)
	if !strings.Contains(output, " | ") {
		t.Fatalf("expected separator ' | ' in output, got %q", output)
	}
}

func TestStatusBar_SelectedHighlight(t *testing.T) {
	sb := NewStatusBar(
		StatusBarItem{Key: "a", Text: "X"},
		StatusBarItem{Key: "b", Text: "Y"},
	)
	sb.Focus()
	// Move to item "b".
	sb.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})

	// Render and check that the selected item has reverse style.
	buf := runtime.NewBuffer(20, 1)
	sb.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 1})
	sb.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 1})
	sb.Render(runtime.RenderContext{Buffer: buf})

	// Find the cell for "Y" (the selected item).
	// "X | Y" -> Y is at position 4.
	cell := buf.Get(4, 0)
	if cell.Rune != 'Y' {
		t.Fatalf("expected 'Y' at position 4, got %q", string(cell.Rune))
	}
	if cell.Style.Attributes()&backend.AttrReverse == 0 {
		t.Fatal("expected selected item to have reverse attribute")
	}

	// First item should NOT have reverse.
	cell0 := buf.Get(0, 0)
	if cell0.Style.Attributes()&backend.AttrReverse != 0 {
		t.Fatal("expected unselected item to NOT have reverse attribute")
	}
}

func TestStatusBar_MouseClick(t *testing.T) {
	var clicked string
	sb := NewStatusBar(
		StatusBarItem{Key: "a", Text: "AA", OnClick: func() { clicked = "a" }},
		StatusBarItem{Key: "b", Text: "BB", OnClick: func() { clicked = "b" }},
	)

	// Layout the widget.
	sb.Measure(runtime.Constraints{MaxWidth: 40, MaxHeight: 1})
	sb.Layout(runtime.Rect{X: 0, Y: 0, Width: 40, Height: 1})

	// Click on first item (position 0).
	result := sb.HandleMessage(runtime.MouseMsg{
		X:      0,
		Y:      0,
		Button: runtime.MouseLeft,
		Action: runtime.MousePress,
	})
	if !result.Handled {
		t.Fatal("expected mouse click to be handled")
	}
	if clicked != "a" {
		t.Fatalf("expected 'a' clicked, got %q", clicked)
	}

	// Click on second item: "AA | BB" -> BB starts at position 5.
	result = sb.HandleMessage(runtime.MouseMsg{
		X:      5,
		Y:      0,
		Button: runtime.MouseLeft,
		Action: runtime.MousePress,
	})
	if !result.Handled {
		t.Fatal("expected mouse click on second item to be handled")
	}
	if clicked != "b" {
		t.Fatalf("expected 'b' clicked, got %q", clicked)
	}
}

func TestStatusBar_MouseClickOutside(t *testing.T) {
	sb := NewStatusBar(
		StatusBarItem{Key: "a", Text: "A"},
	)
	sb.Measure(runtime.Constraints{MaxWidth: 40, MaxHeight: 1})
	sb.Layout(runtime.Rect{X: 0, Y: 0, Width: 40, Height: 1})

	// Click outside the item text area.
	result := sb.HandleMessage(runtime.MouseMsg{
		X:      30,
		Y:      0,
		Button: runtime.MouseLeft,
		Action: runtime.MousePress,
	})
	if result.Handled {
		t.Fatal("expected click outside items to be unhandled")
	}
}

func TestStatusBar_StyleType(t *testing.T) {
	sb := NewStatusBar()
	if sb.StyleType() != "StatusBar" {
		t.Fatalf("expected StyleType 'StatusBar', got %q", sb.StyleType())
	}
}

func TestStatusBar_BindUnbind(t *testing.T) {
	sb := NewStatusBar()
	services := runtime.Services{}
	sb.Bind(services)
	sb.Unbind()
	// Should not panic on nil.
	var nilSB *StatusBar
	nilSB.Bind(services)
	nilSB.Unbind()
}

func TestStatusBar_SyncA11y(t *testing.T) {
	sb := NewStatusBar(
		StatusBarItem{Key: "a", Text: "Alpha"},
		StatusBarItem{Key: "b", Text: "Beta"},
	)
	sb.syncA11y()

	desc := sb.AccessibleDescription()
	if !strings.Contains(desc, "Alpha") || !strings.Contains(desc, "Beta") {
		t.Fatalf("expected description to contain item texts, got %q", desc)
	}

	val := sb.AccessibleValue()
	if val == nil || val.Text != "Alpha" {
		t.Fatalf("expected value text 'Alpha', got %v", val)
	}
}

// renderStatusBar is a helper to render a StatusBar to a string.
func renderStatusBar(sb *StatusBar, width int) string {
	buf := runtime.NewBuffer(width, 1)
	sb.Measure(runtime.Constraints{MaxWidth: width, MaxHeight: 1})
	sb.Layout(runtime.Rect{X: 0, Y: 0, Width: width, Height: 1})
	sb.Render(runtime.RenderContext{Buffer: buf})

	var out strings.Builder
	for x := 0; x < width; x++ {
		cell := buf.Get(x, 0)
		if cell.Rune == 0 {
			out.WriteRune(' ')
		} else {
			out.WriteRune(cell.Rune)
		}
	}
	return out.String()
}
