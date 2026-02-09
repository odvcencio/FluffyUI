package widgets

import (
	"testing"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

// testContainer is a simple container widget for testing tree walking.
type testContainer struct {
	Base
	children []runtime.Widget
}

func (c *testContainer) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (c *testContainer) Render(ctx runtime.RenderContext) {}

func (c *testContainer) ChildWidgets() []runtime.Widget {
	return c.children
}

// testLeaf is a simple leaf widget with no children.
type testLeaf struct {
	Base
}

func (l *testLeaf) Measure(constraints runtime.Constraints) runtime.Size {
	return runtime.Size{Width: 5, Height: 1}
}

func (l *testLeaf) Render(ctx runtime.RenderContext) {}

// testFocusableLeaf is a focusable leaf widget for testing.
type testFocusableLeaf struct {
	FocusableBase
}

func (l *testFocusableLeaf) Measure(constraints runtime.Constraints) runtime.Size {
	return runtime.Size{Width: 5, Height: 1}
}

func (l *testFocusableLeaf) Render(ctx runtime.RenderContext) {}

func TestInspector_NewInspector(t *testing.T) {
	ins := NewInspector(nil)
	if ins == nil {
		t.Fatal("NewInspector returned nil")
	}
	if ins.Entries() != 0 {
		t.Errorf("empty root should have 0 entries, got %d", ins.Entries())
	}
	if !ins.Visible() {
		t.Error("Inspector should be visible by default")
	}
	if ins.AccessibleRole() != accessibility.RoleGroup {
		t.Errorf("role = %q, want %q", ins.AccessibleRole(), accessibility.RoleGroup)
	}
	if ins.AccessibleLabel() != "Widget Inspector" {
		t.Errorf("label = %q, want %q", ins.AccessibleLabel(), "Widget Inspector")
	}
	if !ins.CanFocus() {
		t.Error("Inspector should be focusable")
	}
}

func TestInspector_NewInspector_WithRoot(t *testing.T) {
	leaf := &testLeaf{}
	ins := NewInspector(leaf)
	if ins.Entries() != 1 {
		t.Errorf("single leaf should have 1 entry, got %d", ins.Entries())
	}
}

func TestInspector_BuildTree(t *testing.T) {
	leaf1 := &testLeaf{}
	leaf1.Label = "Leaf1"
	leaf2 := &testLeaf{}
	leaf2.Label = "Leaf2"
	container := &testContainer{
		children: []runtime.Widget{leaf1, leaf2},
	}
	container.Label = "Root"

	ins := NewInspector(container)

	if ins.Entries() != 3 {
		t.Fatalf("expected 3 entries, got %d", ins.Entries())
	}

	// Verify root is first entry at depth 0
	e0 := ins.entries[0]
	if e0.depth != 0 {
		t.Errorf("root depth = %d, want 0", e0.depth)
	}
	if e0.widget != container {
		t.Error("root entry should reference the container widget")
	}

	// Children at depth 1
	e1 := ins.entries[1]
	if e1.depth != 1 {
		t.Errorf("child1 depth = %d, want 1", e1.depth)
	}
	if e1.widget != leaf1 {
		t.Error("child1 should reference leaf1")
	}

	e2 := ins.entries[2]
	if e2.depth != 1 {
		t.Errorf("child2 depth = %d, want 1", e2.depth)
	}
	if e2.widget != leaf2 {
		t.Error("child2 should reference leaf2")
	}
}

func TestInspector_BuildTree_DeepNesting(t *testing.T) {
	leaf := &testLeaf{}
	inner := &testContainer{children: []runtime.Widget{leaf}}
	outer := &testContainer{children: []runtime.Widget{inner}}

	ins := NewInspector(outer)

	if ins.Entries() != 3 {
		t.Fatalf("expected 3 entries, got %d", ins.Entries())
	}
	if ins.entries[0].depth != 0 {
		t.Errorf("outer depth = %d, want 0", ins.entries[0].depth)
	}
	if ins.entries[1].depth != 1 {
		t.Errorf("inner depth = %d, want 1", ins.entries[1].depth)
	}
	if ins.entries[2].depth != 2 {
		t.Errorf("leaf depth = %d, want 2", ins.entries[2].depth)
	}
}

func TestInspector_BuildTree_CycleProtection(t *testing.T) {
	c := &testContainer{}
	c.children = []runtime.Widget{c} // self-referencing cycle
	ins := NewInspector(c)
	if ins.Entries() != 1 {
		t.Errorf("cycle should produce 1 entry, got %d", ins.Entries())
	}
}

func TestInspector_BuildTree_AccessibleLabel(t *testing.T) {
	leaf := &testLeaf{}
	leaf.Label = "MyButton"
	leaf.Role = accessibility.RoleButton

	ins := NewInspector(leaf)
	if ins.Entries() != 1 {
		t.Fatalf("expected 1 entry, got %d", ins.Entries())
	}

	entry := ins.entries[0]
	if entry.role != "button" {
		t.Errorf("role = %q, want %q", entry.role, "button")
	}
	if entry.label != "testLeaf (MyButton)" {
		t.Errorf("label = %q, want %q", entry.label, "testLeaf (MyButton)")
	}
}

func TestInspector_BuildTree_FocusState(t *testing.T) {
	leaf := &testLeaf{}
	leaf.Base.Focus()

	_ = NewInspector(leaf)
	// testLeaf does not implement Focusable (CanFocus returns false from Base)
	// but IsFocused is on Base. The inspector checks runtime.Focusable.
	// Since testLeaf.CanFocus() returns false, it does not satisfy runtime.Focusable.
	// Let's use a testFocusableLeaf defined at package level instead.
	fl := &testFocusableLeaf{}
	fl.FocusableBase.Focus()

	ins2 := NewInspector(fl)
	if ins2.Entries() != 1 {
		t.Fatalf("expected 1 entry, got %d", ins2.Entries())
	}
	if !ins2.entries[0].focused {
		t.Error("focused leaf should have focused=true in entry")
	}
}

func TestInspector_Navigate(t *testing.T) {
	leaf1 := &testLeaf{}
	leaf2 := &testLeaf{}
	container := &testContainer{children: []runtime.Widget{leaf1, leaf2}}

	ins := NewInspector(container)
	ins.Focus()

	if ins.Selected() != 0 {
		t.Fatalf("initial selection = %d, want 0", ins.Selected())
	}

	// Down arrow
	result := ins.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if !result.Handled {
		t.Error("down arrow should be handled")
	}
	if ins.Selected() != 1 {
		t.Errorf("after down: selected = %d, want 1", ins.Selected())
	}

	// Down again
	result = ins.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if !result.Handled {
		t.Error("down arrow should be handled")
	}
	if ins.Selected() != 2 {
		t.Errorf("after second down: selected = %d, want 2", ins.Selected())
	}

	// Down at bottom should not handle
	result = ins.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if result.Handled {
		t.Error("down at bottom should not be handled")
	}
	if ins.Selected() != 2 {
		t.Errorf("selected should stay at 2, got %d", ins.Selected())
	}

	// Up arrow
	result = ins.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if !result.Handled {
		t.Error("up arrow should be handled")
	}
	if ins.Selected() != 1 {
		t.Errorf("after up: selected = %d, want 1", ins.Selected())
	}

	// Up again
	result = ins.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if !result.Handled {
		t.Error("up arrow should be handled")
	}
	if ins.Selected() != 0 {
		t.Errorf("after second up: selected = %d, want 0", ins.Selected())
	}

	// Up at top should not handle
	result = ins.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if result.Handled {
		t.Error("up at top should not be handled")
	}
}

func TestInspector_Toggle(t *testing.T) {
	ins := NewInspector(nil)
	ins.Focus()

	if !ins.Visible() {
		t.Error("should start visible")
	}

	// Escape toggles
	result := ins.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})
	if !result.Handled {
		t.Error("escape should be handled")
	}
	if ins.Visible() {
		t.Error("should be hidden after escape")
	}

	// Toggle again
	result = ins.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})
	if !result.Handled {
		t.Error("escape should be handled")
	}
	if !ins.Visible() {
		t.Error("should be visible after second escape")
	}

	// Toggle via method
	ins.Toggle()
	if ins.Visible() {
		t.Error("Toggle() should hide")
	}
	ins.Toggle()
	if !ins.Visible() {
		t.Error("Toggle() should show")
	}
}

func TestInspector_Measure(t *testing.T) {
	ins := NewInspector(nil)

	size := ins.Measure(runtime.Loose(80, 24))
	if size.Width != 80 || size.Height != 24 {
		t.Errorf("Measure = %dx%d, want 80x24", size.Width, size.Height)
	}

	size = ins.Measure(runtime.Tight(40, 10))
	if size.Width != 40 || size.Height != 10 {
		t.Errorf("Measure tight = %dx%d, want 40x10", size.Width, size.Height)
	}
}

func TestInspector_Refresh(t *testing.T) {
	leaf := &testLeaf{}
	container := &testContainer{children: []runtime.Widget{leaf}}

	ins := NewInspector(container)
	if ins.Entries() != 2 {
		t.Fatalf("initial entries = %d, want 2", ins.Entries())
	}

	// Add another child
	leaf2 := &testLeaf{}
	container.children = append(container.children, leaf2)

	// Entries should not have changed yet
	if ins.Entries() != 2 {
		t.Error("entries should not change without Refresh")
	}

	ins.Refresh()
	if ins.Entries() != 3 {
		t.Errorf("after Refresh: entries = %d, want 3", ins.Entries())
	}
}

func TestInspector_Refresh_Key(t *testing.T) {
	leaf := &testLeaf{}
	container := &testContainer{children: []runtime.Widget{leaf}}
	ins := NewInspector(container)
	ins.Focus()

	container.children = append(container.children, &testLeaf{})

	result := ins.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'r'})
	if !result.Handled {
		t.Error("'r' key should be handled")
	}
	if ins.Entries() != 3 {
		t.Errorf("after 'r' key: entries = %d, want 3", ins.Entries())
	}

	// Also test 'R'
	container.children = append(container.children, &testLeaf{})
	result = ins.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'R'})
	if !result.Handled {
		t.Error("'R' key should be handled")
	}
	if ins.Entries() != 4 {
		t.Errorf("after 'R' key: entries = %d, want 4", ins.Entries())
	}
}

func TestInspector_Render_Hidden(t *testing.T) {
	ins := NewInspector(nil)
	ins.visible = false
	ins.Layout(runtime.Rect{X: 0, Y: 0, Width: 80, Height: 24})

	buf := runtime.NewBuffer(80, 24)
	ctx := runtime.RenderContext{Buffer: buf}

	// Should not panic when hidden
	ins.Render(ctx)
}

func TestInspector_Render_Visible(t *testing.T) {
	leaf := &testLeaf{}
	leaf.Label = "TestWidget"
	ins := NewInspector(leaf)
	ins.Layout(runtime.Rect{X: 0, Y: 0, Width: 80, Height: 24})

	buf := runtime.NewBuffer(80, 24)
	ctx := runtime.RenderContext{Buffer: buf}

	// Should not panic when rendering
	ins.Render(ctx)

	// Verify the title was written - check for 'W' in "Widget Inspector"
	// Title starts at x=1, y=0 -> " Widget Inspector "
	found := false
	for x := 0; x < 80; x++ {
		cell := buf.Get(x, 0)
		if cell.Rune == 'W' {
			found = true
			break
		}
	}
	if !found {
		t.Error("title 'Widget Inspector' not found in top row")
	}
}

func TestInspector_Render_SmallBounds(t *testing.T) {
	ins := NewInspector(nil)
	ins.Layout(runtime.Rect{X: 0, Y: 0, Width: 2, Height: 1})

	buf := runtime.NewBuffer(2, 1)
	ctx := runtime.RenderContext{Buffer: buf}

	// Should not panic with very small bounds
	ins.Render(ctx)
}

func TestInspector_SetRoot(t *testing.T) {
	ins := NewInspector(nil)
	if ins.Entries() != 0 {
		t.Error("initially should have 0 entries")
	}

	leaf := &testLeaf{}
	ins.SetRoot(leaf)
	if ins.Entries() != 1 {
		t.Errorf("after SetRoot: entries = %d, want 1", ins.Entries())
	}

	ins.SetRoot(nil)
	if ins.Entries() != 0 {
		t.Errorf("after SetRoot(nil): entries = %d, want 0", ins.Entries())
	}
}

func TestInspector_SelectedEntry(t *testing.T) {
	leaf := &testLeaf{}
	leaf.Label = "MyLeaf"
	ins := NewInspector(leaf)

	entry := ins.SelectedEntry()
	if entry.widget != leaf {
		t.Error("SelectedEntry should return the leaf widget")
	}
}

func TestInspector_SelectedEntry_Empty(t *testing.T) {
	ins := NewInspector(nil)
	entry := ins.SelectedEntry()
	if entry.widget != nil {
		t.Error("SelectedEntry on empty tree should have nil widget")
	}
}

func TestInspector_NilSafety(t *testing.T) {
	var ins *Inspector

	// All methods should be safe on nil receiver
	ins.SetRoot(nil)
	ins.Toggle()
	ins.Refresh()

	if ins.Visible() {
		t.Error("nil inspector should not be visible")
	}
	if ins.Selected() != 0 {
		t.Error("nil inspector should return 0 for Selected")
	}
	if ins.Entries() != 0 {
		t.Error("nil inspector should return 0 for Entries")
	}

	size := ins.Measure(runtime.Loose(80, 24))
	if size.Width != 0 || size.Height != 0 {
		t.Error("nil inspector should return zero size")
	}

	ins.Layout(runtime.Rect{})
	ins.Bind(runtime.Services{})
	ins.Unbind()

	result := ins.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if result.Handled {
		t.Error("nil inspector should return unhandled")
	}
}

func TestInspector_UnhandledMessages(t *testing.T) {
	ins := NewInspector(nil)
	ins.Focus()

	// Non-key messages should be unhandled
	result := ins.HandleMessage(runtime.ResizeMsg{Width: 80, Height: 24})
	if result.Handled {
		t.Error("resize message should be unhandled")
	}

	// Unrecognized key should be unhandled
	result = ins.HandleMessage(runtime.KeyMsg{Key: terminal.KeyTab})
	if result.Handled {
		t.Error("tab key should be unhandled")
	}
}

func TestInspector_BuildProperties(t *testing.T) {
	leaf := &testLeaf{}
	leaf.Label = "TestLabel"
	leaf.Role = accessibility.RoleButton
	leaf.Description = "A test button"

	ins := NewInspector(leaf)
	entry := ins.entries[0]
	props := ins.buildProperties(entry)

	// Check that we get at least Type, Role, Bounds, Focused, Label, Description
	propMap := make(map[string]string)
	for _, p := range props {
		propMap[p.key] = p.value
	}

	if propMap["Type"] != "testLeaf" {
		t.Errorf("Type = %q, want testLeaf", propMap["Type"])
	}
	if propMap["Role"] != "button" {
		t.Errorf("Role = %q, want button", propMap["Role"])
	}
	if propMap["Label"] != "TestLabel" {
		t.Errorf("Label = %q, want TestLabel", propMap["Label"])
	}
	if propMap["Description"] != "A test button" {
		t.Errorf("Description = %q, want 'A test button'", propMap["Description"])
	}
	if _, ok := propMap["Bounds"]; !ok {
		t.Error("Bounds property missing")
	}
	if _, ok := propMap["Focused"]; !ok {
		t.Error("Focused property missing")
	}
}

func TestInspector_ScrollAdjust(t *testing.T) {
	// Build a tree with many entries
	children := make([]runtime.Widget, 20)
	for i := range children {
		children[i] = &testLeaf{}
	}
	container := &testContainer{children: children}

	ins := NewInspector(container)
	ins.Focus()

	// Navigate down many times
	for i := 0; i < 15; i++ {
		ins.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	}

	if ins.Selected() != 15 {
		t.Errorf("selected = %d, want 15", ins.Selected())
	}

	// Layout with small height to trigger scrolling
	ins.Layout(runtime.Rect{X: 0, Y: 0, Width: 80, Height: 7})

	buf := runtime.NewBuffer(80, 7)
	ctx := runtime.RenderContext{Buffer: buf}
	ins.Render(ctx)

	// scroll should have adjusted so selected is visible
	// innerH = 7-2 = 5, so scroll should be at least 15-5+1=11
	if ins.scroll < 11 {
		t.Errorf("scroll = %d, should be >= 11", ins.scroll)
	}
}

func TestInspector_SelectionClampedOnRefresh(t *testing.T) {
	leaf1 := &testLeaf{}
	leaf2 := &testLeaf{}
	container := &testContainer{children: []runtime.Widget{leaf1, leaf2}}

	ins := NewInspector(container)
	ins.Focus()

	// Navigate to the end
	ins.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	ins.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if ins.Selected() != 2 {
		t.Fatalf("selected = %d, want 2", ins.Selected())
	}

	// Remove a child and refresh
	container.children = []runtime.Widget{leaf1}
	ins.Refresh()

	if ins.Selected() >= ins.Entries() {
		t.Errorf("selected = %d should be < entries = %d", ins.Selected(), ins.Entries())
	}
}

func TestInspector_InterfaceConformance(t *testing.T) {
	// Verify compile-time interface satisfaction
	var _ runtime.Widget = (*Inspector)(nil)
	var _ runtime.Focusable = (*Inspector)(nil)
	var _ runtime.Bindable = (*Inspector)(nil)
	var _ runtime.Unbindable = (*Inspector)(nil)
}

func TestInspector_WidgetTypeName(t *testing.T) {
	leaf := &testLeaf{}
	name := widgetTypeName(leaf)
	if name != "testLeaf" {
		t.Errorf("widgetTypeName = %q, want %q", name, "testLeaf")
	}

	ins := NewInspector(nil)
	name = widgetTypeName(ins)
	if name != "Inspector" {
		t.Errorf("widgetTypeName = %q, want %q", name, "Inspector")
	}
}
