package main

import (
	"testing"

	mcp "m31labs.dev/fluffyui/agent/mcp"
)

func boolPtr(v bool) *bool { return &v }

func TestAnnounceFocusChange(t *testing.T) {
	before := mcp.Snapshot{
		FocusedID: "btn1",
		Widgets: []mcp.WidgetInfo{
			{ID: "btn1", Role: "button", Label: "Save"},
			{ID: "btn2", Role: "button", Label: "Cancel"},
		},
	}
	after := mcp.Snapshot{
		FocusedID: "btn2",
		Widgets: []mcp.WidgetInfo{
			{ID: "btn1", Role: "button", Label: "Save"},
			{ID: "btn2", Role: "button", Label: "Cancel"},
		},
	}
	diff := mcp.Diff{FocusChanged: true}
	utterances := Announce(before, after, diff)
	if len(utterances) == 0 {
		t.Fatal("expected focus change utterance")
	}
	if utterances[0].Priority != "assertive" {
		t.Errorf("expected assertive, got %s", utterances[0].Priority)
	}
	if utterances[0].Text != "Cancel, button" {
		t.Errorf("unexpected text: %q", utterances[0].Text)
	}
}

func TestAnnounceLiveRegionChange(t *testing.T) {
	before := mcp.Snapshot{
		Widgets: []mcp.WidgetInfo{
			{ID: "log1", Role: "list", Label: "Log", Live: "polite", Relevant: "text", Value: "5 entries"},
		},
	}
	after := mcp.Snapshot{
		Widgets: []mcp.WidgetInfo{
			{ID: "log1", Role: "list", Label: "Log", Live: "polite", Relevant: "text", Value: "6 entries"},
		},
	}
	diff := mcp.Diff{
		WidgetsModified: []mcp.WidgetChange{
			{ID: "log1", Changes: map[string]mcp.ValueChange{
				"value": {Old: "5 entries", New: "6 entries"},
			}},
		},
	}
	utterances := Announce(before, after, diff)
	if len(utterances) == 0 {
		t.Fatal("expected live region utterance")
	}
	if utterances[0].Priority != "polite" {
		t.Errorf("expected polite, got %s", utterances[0].Priority)
	}
}

func TestAnnounceAtomicRegion(t *testing.T) {
	before := mcp.Snapshot{
		Widgets: []mcp.WidgetInfo{
			{ID: "alert1", Role: "alert", Label: "Error", Live: "assertive", Relevant: "all", Atomic: true, Value: "old"},
		},
	}
	after := mcp.Snapshot{
		Widgets: []mcp.WidgetInfo{
			{ID: "alert1", Role: "alert", Label: "Error", Live: "assertive", Relevant: "all", Atomic: true, Value: "File not found"},
		},
	}
	diff := mcp.Diff{
		WidgetsModified: []mcp.WidgetChange{
			{ID: "alert1", Changes: map[string]mcp.ValueChange{
				"value": {Old: "old", New: "File not found"},
			}},
		},
	}
	utterances := Announce(before, after, diff)
	if len(utterances) == 0 {
		t.Fatal("expected atomic utterance")
	}
	// Atomic should re-announce the full widget
	if utterances[0].Text != "Error, alert, File not found" {
		t.Errorf("unexpected text: %q", utterances[0].Text)
	}
	if utterances[0].Priority != "assertive" {
		t.Errorf("expected assertive, got %s", utterances[0].Priority)
	}
}

func TestAnnounceSkipsBusy(t *testing.T) {
	before := mcp.Snapshot{
		Widgets: []mcp.WidgetInfo{
			{ID: "w1", Role: "status", Label: "Status", Live: "polite", Value: "old", State: mcp.StateSet{Busy: true}},
		},
	}
	after := mcp.Snapshot{
		Widgets: []mcp.WidgetInfo{
			{ID: "w1", Role: "status", Label: "Status", Live: "polite", Value: "new", State: mcp.StateSet{Busy: true}},
		},
	}
	diff := mcp.Diff{
		WidgetsModified: []mcp.WidgetChange{
			{ID: "w1", Changes: map[string]mcp.ValueChange{
				"value": {Old: "old", New: "new"},
			}},
		},
	}
	utterances := Announce(before, after, diff)
	if len(utterances) != 0 {
		t.Errorf("expected no utterances during busy, got %d", len(utterances))
	}
}

func TestAnnounceWidgetAdded(t *testing.T) {
	before := mcp.Snapshot{
		Widgets: []mcp.WidgetInfo{
			{ID: "list1", Role: "list", Label: "Todos", Live: "polite", Relevant: "additions"},
		},
	}
	after := mcp.Snapshot{
		Widgets: []mcp.WidgetInfo{
			{ID: "list1", Role: "list", Label: "Todos", Live: "polite", Relevant: "additions"},
			{ID: "item1", Role: "listitem", Label: "Buy milk", ParentID: "list1"},
		},
	}
	diff := mcp.Diff{
		WidgetsAdded: []string{"item1"},
	}
	utterances := Announce(before, after, diff)
	if len(utterances) == 0 {
		t.Fatal("expected addition utterance")
	}
	if utterances[0].Text != "Buy milk, listitem" {
		t.Errorf("unexpected text: %q", utterances[0].Text)
	}
}

func TestAnnounceWidgetRemoved(t *testing.T) {
	before := mcp.Snapshot{
		Widgets: []mcp.WidgetInfo{
			{ID: "list1", Role: "list", Label: "Todos", Live: "polite", Relevant: "all"},
			{ID: "item1", Role: "listitem", Label: "Buy milk", ParentID: "list1"},
		},
	}
	after := mcp.Snapshot{
		Widgets: []mcp.WidgetInfo{
			{ID: "list1", Role: "list", Label: "Todos", Live: "polite", Relevant: "all"},
		},
	}
	diff := mcp.Diff{
		WidgetsRemoved: []string{"item1"},
	}
	utterances := Announce(before, after, diff)
	if len(utterances) == 0 {
		t.Fatal("expected removal utterance")
	}
	if utterances[0].Text != "Removed Buy milk" {
		t.Errorf("unexpected text: %q", utterances[0].Text)
	}
}

func TestAnnounceNoLiveSkipped(t *testing.T) {
	before := mcp.Snapshot{
		Widgets: []mcp.WidgetInfo{
			{ID: "w1", Role: "text", Label: "Static", Value: "old"},
		},
	}
	after := mcp.Snapshot{
		Widgets: []mcp.WidgetInfo{
			{ID: "w1", Role: "text", Label: "Static", Value: "new"},
		},
	}
	diff := mcp.Diff{
		WidgetsModified: []mcp.WidgetChange{
			{ID: "w1", Changes: map[string]mcp.ValueChange{
				"value": {Old: "old", New: "new"},
			}},
		},
	}
	utterances := Announce(before, after, diff)
	if len(utterances) != 0 {
		t.Errorf("expected no utterances for non-live widget, got %d", len(utterances))
	}
}

func TestFormatWidget(t *testing.T) {
	tests := []struct {
		name   string
		widget mcp.WidgetInfo
		want   string
	}{
		{
			name:   "button",
			widget: mcp.WidgetInfo{Role: "button", Label: "Save"},
			want:   "Save, button",
		},
		{
			name:   "checkbox checked",
			widget: mcp.WidgetInfo{Role: "checkbox", Label: "Accept", State: mcp.StateSet{Checked: boolPtr(true)}},
			want:   "Accept, checkbox, checked",
		},
		{
			name:   "input with value",
			widget: mcp.WidgetInfo{Role: "input", Label: "Name", Value: "Alice"},
			want:   "Name, input, Alice",
		},
		{
			name:   "empty",
			widget: mcp.WidgetInfo{},
			want:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatWidget(tt.widget)
			if got != tt.want {
				t.Errorf("formatWidget() = %q, want %q", got, tt.want)
			}
		})
	}
}
