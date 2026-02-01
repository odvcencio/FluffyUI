package mcp

import "testing"

func TestResourceHelpers(t *testing.T) {
	widgets := []WidgetInfo{
		{ID: "layer0:btn:0:save"},
		{ID: "layer0:btn:1:save#2"},
		{ID: "layer1:label:0:title"},
	}
	index := explicitIndexFromWidgets(widgets)
	if len(index["save"]) != 2 {
		t.Fatalf("expected explicit index")
	}
	ref := resourceRef{uri: "fluffy://widget/layer0:btn:0:save", id: "layer0:btn:0:save"}
	if widgetIDChangedURI(ref, index) != "" {
		t.Fatalf("expected no unique replacement")
	}
	index = map[string][]string{"save": {"layer0:btn:2:save"}}
	if widgetIDChangedURI(ref, index) != "fluffy://widget/layer0:btn:2:save" {
		t.Fatalf("expected widget uri replacement")
	}
	ref.subresource = "value"
	if widgetIDChangedURI(ref, index) != "fluffy://widget/layer0:btn:2:save/value" {
		t.Fatalf("expected subresource uri replacement")
	}
}

func TestChangedChildrenAndParseRefs(t *testing.T) {
	if changedChildren(WidgetChange{}) {
		t.Fatalf("expected no changed children")
	}
	change := WidgetChange{Changes: map[string]ValueChange{"children_ids": {}}}
	if !changedChildren(change) {
		t.Fatalf("expected changed children")
	}

	refs, needsText, needsClipboard := parseResourceRefs([]string{"fluffy://screen", "fluffy://clipboard", "fluffy://widget/id"})
	if len(refs) != 3 || !needsText || !needsClipboard {
		t.Fatalf("unexpected ref parsing")
	}
}

func TestComputeChangedLayers(t *testing.T) {
	before := Snapshot{LayerCount: 1}
	after := Snapshot{LayerCount: 2}
	diff := Diff{LayerCountChanged: true, WidgetsAdded: []string{"layer1:label:0:title"}}
	changed := computeChangedLayers(before, after, diff)
	if !changed[0] || !changed[1] {
		t.Fatalf("expected layer changes")
	}
}
