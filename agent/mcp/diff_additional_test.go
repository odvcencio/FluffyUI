package mcp

import "testing"

func TestDiffSnapshotsAndWidgetChanges(t *testing.T) {
	before := Snapshot{
		Dimensions: Dimensions{Width: 10, Height: 10},
		LayerCount: 1,
		FocusedID:  "layer0:btn:0:save",
		Widgets: []WidgetInfo{
			{ID: "layer0:btn:0:save", Label: "Save", Bounds: Rect{X: 0, Y: 0, Width: 2, Height: 1}},
		},
		Text: "hello",
	}
	checked := true
	after := Snapshot{
		Dimensions: Dimensions{Width: 12, Height: 10},
		LayerCount: 2,
		FocusedID:  "layer0:btn:0:save",
		Widgets: []WidgetInfo{
			{ID: "layer0:btn:0:save", Label: "Save!", Bounds: Rect{X: 1, Y: 0, Width: 2, Height: 1}, State: StateSet{Checked: &checked}},
			{ID: "layer1:label:0:title", Label: "Title"},
		},
		Text: "hello world",
	}
	diff := diffSnapshots(before, after)
	if !diff.TextChanged || !diff.DimensionsChanged || !diff.LayerCountChanged {
		t.Fatalf("expected top-level changes")
	}
	if len(diff.WidgetsAdded) != 1 || diff.WidgetsAdded[0] != "layer1:label:0:title" {
		t.Fatalf("expected widget added")
	}
	if len(diff.WidgetsModified) != 1 {
		t.Fatalf("expected widget modified")
	}
	change := diff.WidgetsModified[0]
	if _, ok := change.Changes["label"]; !ok {
		t.Fatalf("expected label change")
	}
	if _, ok := change.Changes["state.checked"]; !ok {
		t.Fatalf("expected state checked change")
	}
}

func TestStateDiffHelpers(t *testing.T) {
	checked := true
	unchecked := false
	before := StateSet{Focused: true, Checked: &checked}
	after := StateSet{Focused: false, Checked: &unchecked}
	changes := stateDiff(before, after)
	if _, ok := changes["focused"]; !ok {
		t.Fatalf("expected focused change")
	}
	if _, ok := changes["checked"]; !ok {
		t.Fatalf("expected checked change")
	}
	if !boolPtrEqual(nil, nil) || boolPtrEqual(&checked, nil) || boolPtrEqual(nil, &checked) {
		t.Fatalf("boolPtrEqual mismatch")
	}
	if !boolPtrEqual(&checked, &checked) {
		t.Fatalf("boolPtrEqual true mismatch")
	}
	if boolPtrEqual(&checked, &unchecked) {
		t.Fatalf("boolPtrEqual should be false")
	}
	if !equalStringSlice([]string{"a"}, []string{"a"}) {
		t.Fatalf("expected equal string slice")
	}
	if equalStringSlice([]string{"a"}, []string{"b"}) {
		t.Fatalf("expected unequal string slice")
	}
}
