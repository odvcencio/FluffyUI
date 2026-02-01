package mcp

import (
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/agent"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
)

func TestSnapshotConversions(t *testing.T) {
	checked := true
	state := accessibility.StateSet{Checked: &checked, Selected: true}
	child := agent.WidgetInfo{ID: "layer0:label:0.0", Label: "Child"}
	root := agent.WidgetInfo{ID: "layer0:stack:0", Label: "Root", State: state, Children: []agent.WidgetInfo{child}}
	snap := agent.Snapshot{
		Timestamp:  time.Unix(1, 0),
		Width:      80,
		Height:     24,
		LayerCount: 1,
		FocusedID:  child.ID,
		Widgets:    []agent.WidgetInfo{root},
		Text:       "hello",
	}

	flat := snapshotFromAgent(snap, true)
	if flat.Dimensions.Width != 80 || flat.Dimensions.Height != 24 {
		t.Fatalf("unexpected dimensions")
	}
	if flat.Text != "hello" {
		t.Fatalf("expected text")
	}
	if len(flat.Widgets) != 2 {
		t.Fatalf("expected 2 widgets, got %d", len(flat.Widgets))
	}
	if flat.Widgets[1].ParentID != root.ID {
		t.Fatalf("expected parent id")
	}

	tree := treeSnapshotFromAgent(snap, true)
	if len(tree.Widgets) != 1 || len(tree.Widgets[0].Children) != 1 {
		t.Fatalf("expected tree widgets")
	}
	if tree.Text != "hello" {
		t.Fatalf("expected text")
	}
}

func TestRoleAndStateHelpers(t *testing.T) {
	if roleToMCP(accessibility.RoleButton) != "button" {
		t.Fatalf("role mapping failed")
	}
	if roleToMCP(accessibility.RoleGroup) != "container" {
		t.Fatalf("role mapping failed")
	}
	if roleToMCP("unknown") != "unknown" {
		t.Fatalf("role mapping default failed")
	}

	info := agent.WidgetInfo{Focused: true, State: accessibility.StateSet{Disabled: true}}
	state := stateFromAgent(info)
	if !state.Focused || !state.Disabled {
		t.Fatalf("state mapping failed")
	}
}

func TestWidgetIDParsing(t *testing.T) {
	if layerFromID("layer2:widget:0") != 2 {
		t.Fatalf("expected layer 2")
	}
	if layerFromID("bad") != 0 {
		t.Fatalf("expected default layer")
	}
	if explicitIDFromWidgetID("layer0:widget:0:submit#2") != "submit#2" {
		t.Fatalf("explicit id parse failed")
	}
	if explicitBaseID("submit#2") != "submit" {
		t.Fatalf("explicit base failed")
	}
	if explicitBaseID(" ") != "" {
		t.Fatalf("expected empty base id")
	}
}

func TestParseResourceURI(t *testing.T) {
	ref := parseResourceURI("fluffy://widget/foo/value")
	if ref.kind != resourceWidgetValue || ref.id != "foo" || ref.subresource != "value" {
		t.Fatalf("unexpected widget value ref: %#v", ref)
	}
	ref = parseResourceURI("fluffy://widget/foo/children")
	if ref.kind != resourceWidgetChildren || ref.subresource != "children" {
		t.Fatalf("unexpected widget children ref")
	}
	ref = parseResourceURI("fluffy://layer/3")
	if ref.kind != resourceLayer || ref.layer != 3 {
		t.Fatalf("unexpected layer ref")
	}
	ref = parseResourceURI("fluffy://screen")
	if ref.kind != resourceScreen {
		t.Fatalf("unexpected screen ref")
	}
	ref = parseResourceURI("http://screen")
	if ref.kind != resourceUnknown {
		t.Fatalf("expected unknown for invalid scheme")
	}
}

func TestWidgetIndexesAndSearch(t *testing.T) {
	widgets := []WidgetInfo{
		{ID: "layer0:button:0:submit", Label: "Submit"},
		{ID: "layer0:button:1:submit#2", Label: "Submit"},
		{ID: "layer1:label:0:title", Label: "Title"},
	}
	index := indexWidgets(widgets)
	if len(index) != 3 {
		t.Fatalf("expected index")
	}
	collected := collectWidgets(index, []string{"layer1:label:0:title"})
	if len(collected) != 1 {
		t.Fatalf("expected collected widget")
	}

	found := findWidgetByID(widgets, "submit", false)
	if found == nil || found.ID != widgets[0].ID {
		t.Fatalf("expected first submit")
	}
	strict := findWidgetByID(widgets, "submit", true)
	if strict != nil {
		t.Fatalf("expected strict match to fail")
	}
	found = findWidgetByID(widgets, widgets[2].ID, false)
	if found == nil || found.ID != widgets[2].ID {
		t.Fatalf("expected direct id match")
	}
}

func TestWalkDescendantsAndVisibility(t *testing.T) {
	widgets := []WidgetInfo{
		{ID: "layer0:root:0", ChildrenIDs: []string{"layer0:child:0.0"}},
		{ID: "layer0:child:0.0"},
	}
	index := indexWidgets(widgets)
	var out []WidgetInfo
	walkDescendants(index, []string{"layer0:root:0"}, &out)
	if len(out) != 2 {
		t.Fatalf("expected 2 descendants")
	}

	dims := Dimensions{Width: 10, Height: 10}
	bounds := Rect{X: -1, Y: -1, Width: 3, Height: 3}
	if !visibleBounds(bounds, dims) {
		t.Fatalf("expected visible bounds")
	}
	if visibleArea(WidgetInfo{Bounds: Rect{X: 20, Y: 20, Width: 1, Height: 1}}, dims) != 0 {
		t.Fatalf("expected zero visible area")
	}
}

func TestStateMatchesAndFocusNavigation(t *testing.T) {
	checked := true
	filter := StateSet{Checked: &checked}
	widget := StateSet{Checked: &checked}
	if !stateMatches(widget, filter) {
		t.Fatalf("expected state match")
	}
	unchecked := false
	filter.Checked = &unchecked
	if stateMatches(widget, filter) {
		t.Fatalf("expected state mismatch")
	}

	snap := Snapshot{
		FocusedID: "layer0:btn:0",
		Widgets: []WidgetInfo{
			{ID: "layer0:btn:0", Actions: []string{"focus"}},
			{ID: "layer0:btn:1", Actions: []string{"focus"}},
			{ID: "layer1:btn:0", Actions: []string{"focus"}},
		},
	}
	next := nextFocusable(snap, "layer0:btn:0", true)
	if next == nil || next.ID != "layer0:btn:1" {
		t.Fatalf("expected next focusable")
	}
	prev := nextFocusable(snap, "layer0:btn:0", false)
	if prev == nil || prev.ID != "layer0:btn:1" {
		t.Fatalf("expected wrap-around focusable")
	}
}

func TestCellInfoAndColors(t *testing.T) {
	style := backend.DefaultStyle().Foreground(backend.ColorRed).Background(backend.ColorRGB(1, 2, 3)).Bold(true).Underline(true)
	cell := backend.Cell{Rune: 'A', Style: style}
	info := cellInfoFromCell(cell, true)
	if info.Char != "A" {
		t.Fatalf("expected char")
	}
	if info.Style.Bold != true || info.Style.Underline != true {
		t.Fatalf("expected style attributes")
	}
	if colorValueFromColor(backend.ColorDefault) != "default" {
		t.Fatalf("expected default color string")
	}
	if colorValueFromColor(backend.ColorRed) != "red" {
		t.Fatalf("expected red")
	}
	if colorValueFromColor(backend.ColorRGB(10, 11, 12)) != "#0a0b0c" {
		t.Fatalf("expected rgb hex")
	}
}

func TestFindWidgetAtPosition(t *testing.T) {
	snap := Snapshot{
		Widgets: []WidgetInfo{
			{ID: "layer0:box:0", Bounds: Rect{X: 0, Y: 0, Width: 5, Height: 5}},
			{ID: "layer1:box:0", Bounds: Rect{X: 0, Y: 0, Width: 5, Height: 5}},
		},
	}
	found := findWidgetAtPosition(snap, 2, 2)
	if found == nil || found.ID != "layer1:box:0" {
		t.Fatalf("expected topmost widget")
	}
}

func TestRectFromRuntime(t *testing.T) {
	r := rectFromRuntime(runtime.Rect{X: 1, Y: 2, Width: 3, Height: 4})
	if r.X != 1 || r.Height != 4 {
		t.Fatalf("unexpected rect")
	}
}
