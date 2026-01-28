package dragdrop

import "testing"

type testDraggable struct {
	endCalled bool
}

type testDropTarget struct {
	dropped   bool
	previewed bool
	last      DropPosition
	data      DragData
}

func (d *testDraggable) DragStart() DragData {
	return DragData{Kind: "test"}
}

func (d *testDraggable) DragEnd(cancelled bool) {
	d.endCalled = cancelled
}

func (t *testDropTarget) CanDrop(data DragData) bool {
	return data.Kind == "test"
}

func (t *testDropTarget) Drop(data DragData, position DropPosition) {
	t.dropped = true
	t.data = data
	t.last = position
}

func (t *testDropTarget) DropPreview(data DragData, position DropPosition) {
	t.previewed = true
	t.data = data
	t.last = position
}

func TestDragDropInterfaces(t *testing.T) {
	var _ Draggable = (*testDraggable)(nil)
	var _ DropTarget = (*testDropTarget)(nil)

	dragger := &testDraggable{}
	data := dragger.DragStart()
	if data.Kind != "test" {
		t.Fatalf("unexpected drag data kind: %s", data.Kind)
	}
	dragger.DragEnd(true)
	if !dragger.endCalled {
		t.Fatalf("expected DragEnd to record cancellation")
	}

	target := &testDropTarget{}
	if !target.CanDrop(data) {
		t.Fatalf("expected target to accept drag data")
	}
	pos := DropPosition{X: 3, Y: 4}
	target.DropPreview(data, pos)
	if !target.previewed || target.last != pos {
		t.Fatalf("expected DropPreview to record position")
	}
	target.Drop(data, pos)
	if !target.dropped {
		t.Fatalf("expected Drop to be recorded")
	}
}
