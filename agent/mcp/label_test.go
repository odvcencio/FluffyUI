package mcp

import "testing"

func TestLabelMatchesAndResolve(t *testing.T) {
	snap := Snapshot{
		Dimensions: Dimensions{Width: 10, Height: 10},
		FocusedID:  "layer0:btn:0:save",
		Widgets: []WidgetInfo{
			{ID: "layer0:btn:0:save", Label: "Save", Bounds: Rect{X: 0, Y: 0, Width: 2, Height: 1}},
			{ID: "layer1:btn:0:save#2", Label: "Save", Bounds: Rect{X: 0, Y: 0, Width: 2, Height: 1}},
			{ID: "layer0:label:1:other", Label: "Other", Bounds: Rect{X: 20, Y: 20, Width: 1, Height: 1}},
		},
	}
	matches := labelMatches(snap, "save")
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	chosen, reason := resolveLabelMatches(matches)
	if reason != "focused" || chosen.Widget.ID != "layer0:btn:0:save" {
		t.Fatalf("expected focused resolution")
	}

	snap.FocusedID = ""
	matches = labelMatches(snap, "save")
	chosen, reason = resolveLabelMatches(matches)
	if reason != "topmost_layer" || chosen.Widget.ID != "layer1:btn:0:save#2" {
		t.Fatalf("expected topmost layer resolution")
	}
}

func TestMatchSummaries(t *testing.T) {
	snap := Snapshot{
		Widgets: []WidgetInfo{
			{ID: "layer0:root:0", Label: "Root"},
			{ID: "layer0:child:0.0", Label: "Child", ParentID: "layer0:root:0"},
		},
	}
	matches := []labelMatch{
		{Widget: snap.Widgets[1], Layer: 0},
	}
	out := matchSummaries(matches, snap)
	if len(out) != 1 || out[0].Context != "Root" {
		t.Fatalf("expected parent context")
	}
}

func TestWidgetVisibilityHelpers(t *testing.T) {
	widget := WidgetInfo{Bounds: Rect{X: 0, Y: 0, Width: 2, Height: 2}}
	dims := Dimensions{Width: 1, Height: 1}
	if widgetVisible(widget, dims) {
		t.Fatalf("expected not fully visible")
	}
	if visibleArea(widget, dims) != 1 {
		t.Fatalf("expected visible area 1")
	}
}
