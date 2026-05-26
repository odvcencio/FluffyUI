package main

import (
	"testing"

	mcp "m31labs.dev/fluffyui/agent/mcp"
)

func TestDiffSnapshotsAdded(t *testing.T) {
	before := mcp.Snapshot{
		Widgets: []mcp.WidgetInfo{
			{ID: "w1", Role: "button", Label: "A"},
		},
	}
	after := mcp.Snapshot{
		Widgets: []mcp.WidgetInfo{
			{ID: "w1", Role: "button", Label: "A"},
			{ID: "w2", Role: "button", Label: "B"},
		},
	}
	diff := diffSnapshots(before, after)
	if len(diff.WidgetsAdded) != 1 || diff.WidgetsAdded[0] != "w2" {
		t.Errorf("expected w2 added, got %v", diff.WidgetsAdded)
	}
}

func TestDiffSnapshotsRemoved(t *testing.T) {
	before := mcp.Snapshot{
		Widgets: []mcp.WidgetInfo{
			{ID: "w1", Role: "button", Label: "A"},
			{ID: "w2", Role: "button", Label: "B"},
		},
	}
	after := mcp.Snapshot{
		Widgets: []mcp.WidgetInfo{
			{ID: "w1", Role: "button", Label: "A"},
		},
	}
	diff := diffSnapshots(before, after)
	if len(diff.WidgetsRemoved) != 1 || diff.WidgetsRemoved[0] != "w2" {
		t.Errorf("expected w2 removed, got %v", diff.WidgetsRemoved)
	}
}

func TestDiffSnapshotsModified(t *testing.T) {
	before := mcp.Snapshot{
		Widgets: []mcp.WidgetInfo{
			{ID: "w1", Role: "button", Label: "Save"},
		},
	}
	after := mcp.Snapshot{
		Widgets: []mcp.WidgetInfo{
			{ID: "w1", Role: "button", Label: "Saved!"},
		},
	}
	diff := diffSnapshots(before, after)
	if len(diff.WidgetsModified) != 1 {
		t.Fatalf("expected 1 modified, got %d", len(diff.WidgetsModified))
	}
	if diff.WidgetsModified[0].ID != "w1" {
		t.Errorf("expected w1, got %s", diff.WidgetsModified[0].ID)
	}
	if _, ok := diff.WidgetsModified[0].Changes["label"]; !ok {
		t.Error("expected label change")
	}
}

func TestDiffSnapshotsFocusChanged(t *testing.T) {
	before := mcp.Snapshot{FocusedID: "w1"}
	after := mcp.Snapshot{FocusedID: "w2"}
	diff := diffSnapshots(before, after)
	if !diff.FocusChanged {
		t.Error("expected focus changed")
	}
}

func TestDiffSnapshotsARIAFields(t *testing.T) {
	before := mcp.Snapshot{
		Widgets: []mcp.WidgetInfo{
			{ID: "w1", Role: "status", Live: "polite"},
		},
	}
	after := mcp.Snapshot{
		Widgets: []mcp.WidgetInfo{
			{ID: "w1", Role: "status", Live: "assertive"},
		},
	}
	diff := diffSnapshots(before, after)
	if len(diff.WidgetsModified) != 1 {
		t.Fatalf("expected 1 modified, got %d", len(diff.WidgetsModified))
	}
	if _, ok := diff.WidgetsModified[0].Changes["live"]; !ok {
		t.Error("expected live change in diff")
	}
}

func TestBoolPtrEqual(t *testing.T) {
	tr := true
	fa := false
	if !boolPtrEqual(nil, nil) {
		t.Error("nil == nil should be true")
	}
	if boolPtrEqual(nil, &tr) {
		t.Error("nil != &true")
	}
	if boolPtrEqual(&tr, &fa) {
		t.Error("true != false")
	}
	if !boolPtrEqual(&tr, &tr) {
		t.Error("true == true")
	}
}
