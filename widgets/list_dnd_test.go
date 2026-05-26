package widgets

import (
	"testing"

	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/terminal"
)

func newTestDraggableList() *List[string] {
	adapter := NewMutableSliceAdapter([]string{"A", "B", "C", "D", "E"}, nil)
	list := NewList[string](adapter)
	list.SetDraggable(true)
	list.Focus()
	list.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 10})
	list.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 10})
	return list
}

func TestList_SetDraggable(t *testing.T) {
	adapter := NewSliceAdapter([]string{"A", "B"}, nil)
	list := NewList(adapter)
	if list.IsDraggable() {
		t.Fatal("expected draggable to be false by default")
	}
	list.SetDraggable(true)
	if !list.IsDraggable() {
		t.Fatal("expected draggable to be true after SetDraggable(true)")
	}
	list.SetDraggable(false)
	if list.IsDraggable() {
		t.Fatal("expected draggable to be false after SetDraggable(false)")
	}
}

func TestList_CtrlG_StartsDrag(t *testing.T) {
	list := newTestDraggableList()
	// Select item at index 1 (B).
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if list.SelectedIndex() != 1 {
		t.Fatalf("expected selected 1, got %d", list.SelectedIndex())
	}

	// Start drag with Ctrl+G.
	result := list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlG})
	if !result.Handled {
		t.Fatal("expected Ctrl+G to be handled")
	}
	if len(result.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(result.Commands))
	}
	startMsg, ok := result.Commands[0].(runtime.DragStartMsg)
	if !ok {
		t.Fatalf("expected DragStartMsg, got %T", result.Commands[0])
	}
	if startMsg.Data.Kind != "list-item" {
		t.Fatalf("expected Kind 'list-item', got %q", startMsg.Data.Kind)
	}
	if startMsg.Data.Label != "B" {
		t.Fatalf("expected Label 'B', got %q", startMsg.Data.Label)
	}
	if !list.IsDragging() {
		t.Fatal("expected list to be in dragging state")
	}
	if list.DragOrigin() != 1 {
		t.Fatalf("expected drag origin 1, got %d", list.DragOrigin())
	}
	if list.DropIndex() != 1 {
		t.Fatalf("expected drop index 1, got %d", list.DropIndex())
	}
}

func TestList_CtrlG_NotDraggable(t *testing.T) {
	adapter := NewSliceAdapter([]string{"A", "B"}, nil)
	list := NewList(adapter)
	list.Focus()
	// Not draggable, so Ctrl+G should be unhandled.
	result := list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlG})
	if result.Handled {
		t.Fatal("expected Ctrl+G to be unhandled when not draggable")
	}
}

func TestList_DragArrowKeys_MoveDropIndicator(t *testing.T) {
	list := newTestDraggableList()
	// Select item 2 (C) and start drag.
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlG})

	if list.DropIndex() != 2 {
		t.Fatalf("expected drop index 2, got %d", list.DropIndex())
	}

	// Move down.
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if list.DropIndex() != 3 {
		t.Fatalf("expected drop index 3 after down, got %d", list.DropIndex())
	}

	// Move down again.
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if list.DropIndex() != 4 {
		t.Fatalf("expected drop index 4 after down, got %d", list.DropIndex())
	}

	// Move down at boundary should stay.
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if list.DropIndex() != 4 {
		t.Fatalf("expected drop index 4 at boundary, got %d", list.DropIndex())
	}

	// Move up.
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if list.DropIndex() != 3 {
		t.Fatalf("expected drop index 3 after up, got %d", list.DropIndex())
	}

	// Move up past boundary.
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if list.DropIndex() != 0 {
		t.Fatalf("expected drop index 0 at top boundary, got %d", list.DropIndex())
	}
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if list.DropIndex() != 0 {
		t.Fatalf("expected drop index 0, got %d", list.DropIndex())
	}
}

func TestList_DragHomeEnd(t *testing.T) {
	list := newTestDraggableList()
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlG})

	// Home should jump to 0.
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyHome})
	if list.DropIndex() != 0 {
		t.Fatalf("expected drop index 0 after Home, got %d", list.DropIndex())
	}

	// End should jump to last.
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnd})
	if list.DropIndex() != 4 {
		t.Fatalf("expected drop index 4 after End, got %d", list.DropIndex())
	}
}

func TestList_EnterCompletesDrop(t *testing.T) {
	list := newTestDraggableList()

	// Select "B" (index 1), start drag.
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlG})

	// Move down to index 3.
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})

	// Drop with Enter.
	result := list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if !result.Handled {
		t.Fatal("expected Enter to be handled during drag")
	}
	if len(result.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(result.Commands))
	}
	endMsg, ok := result.Commands[0].(runtime.DragEndMsg)
	if !ok {
		t.Fatalf("expected DragEndMsg, got %T", result.Commands[0])
	}
	if endMsg.Cancelled {
		t.Fatal("expected Cancelled to be false")
	}
	if endMsg.Target != list {
		t.Fatal("expected Target to be the list")
	}
	if !list.IsDragging() == true {
		// Should not be dragging after drop.
	}
	if list.IsDragging() {
		t.Fatal("expected dragging to be false after drop")
	}

	// Verify the adapter was reordered: B moved from 1 to 3.
	mutable := list.adapter.(*MutableSliceAdapter[string])
	items := mutable.Items()
	expected := []string{"A", "C", "D", "B", "E"}
	for i, want := range expected {
		if items[i] != want {
			t.Fatalf("items[%d] = %q, want %q (full: %v)", i, items[i], want, items)
		}
	}

	// Selected index should be at the drop position.
	if list.SelectedIndex() != 3 {
		t.Fatalf("expected selected 3 after drop, got %d", list.SelectedIndex())
	}
}

func TestList_EscapeCancelsDrag(t *testing.T) {
	list := newTestDraggableList()
	// Select "C" (index 2), start drag.
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlG})

	// Move to index 4.
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})

	// Cancel with Escape.
	result := list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})
	if !result.Handled {
		t.Fatal("expected Escape to be handled during drag")
	}
	if len(result.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(result.Commands))
	}
	endMsg, ok := result.Commands[0].(runtime.DragEndMsg)
	if !ok {
		t.Fatalf("expected DragEndMsg, got %T", result.Commands[0])
	}
	if !endMsg.Cancelled {
		t.Fatal("expected Cancelled to be true")
	}
	if endMsg.Target != nil {
		t.Fatal("expected Target to be nil for cancelled drag")
	}
	if list.IsDragging() {
		t.Fatal("expected dragging to be false after cancel")
	}

	// Verify the adapter was NOT reordered.
	mutable := list.adapter.(*MutableSliceAdapter[string])
	items := mutable.Items()
	expected := []string{"A", "B", "C", "D", "E"}
	for i, want := range expected {
		if items[i] != want {
			t.Fatalf("items[%d] = %q, want %q", i, items[i], want)
		}
	}

	// Selected should be restored to origin.
	if list.SelectedIndex() != 2 {
		t.Fatalf("expected selected 2 after cancel, got %d", list.SelectedIndex())
	}
}

func TestList_CtrlD_CompletesDrop(t *testing.T) {
	list := newTestDraggableList()
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown}) // select B
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlG})
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown}) // drop at C

	result := list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlD})
	if !result.Handled {
		t.Fatal("expected Ctrl+D to be handled during drag")
	}
	endMsg, ok := result.Commands[0].(runtime.DragEndMsg)
	if !ok {
		t.Fatalf("expected DragEndMsg, got %T", result.Commands[0])
	}
	if endMsg.Cancelled {
		t.Fatal("expected not cancelled")
	}
}

func TestList_OnDropCallback(t *testing.T) {
	list := newTestDraggableList()
	var dropFrom, dropTo int
	var dropItem string
	list.SetOnDrop(func(from, to int, item string) {
		dropFrom = from
		dropTo = to
		dropItem = item
	})

	// Drag B (index 1) to index 3.
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlG})
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})

	if dropFrom != 1 {
		t.Fatalf("expected dropFrom 1, got %d", dropFrom)
	}
	if dropTo != 3 {
		t.Fatalf("expected dropTo 3, got %d", dropTo)
	}
	if dropItem != "B" {
		t.Fatalf("expected dropItem 'B', got %q", dropItem)
	}
}

func TestList_DragSamePosition(t *testing.T) {
	list := newTestDraggableList()
	// Select A (index 0), start drag, drop at same position.
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlG})
	result := list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})

	endMsg, ok := result.Commands[0].(runtime.DragEndMsg)
	if !ok {
		t.Fatalf("expected DragEndMsg, got %T", result.Commands[0])
	}
	if endMsg.Cancelled {
		t.Fatal("expected not cancelled even for same position drop")
	}

	// Verify order unchanged.
	mutable := list.adapter.(*MutableSliceAdapter[string])
	items := mutable.Items()
	expected := []string{"A", "B", "C", "D", "E"}
	for i, want := range expected {
		if items[i] != want {
			t.Fatalf("items[%d] = %q, want %q", i, items[i], want)
		}
	}
}

func TestList_DragSwallowsOtherKeys(t *testing.T) {
	list := newTestDraggableList()
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlG})

	// PageUp/Down/etc. should be swallowed during drag.
	result := list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyPageDown})
	if !result.Handled {
		t.Fatal("expected other keys to be swallowed during drag")
	}
	result = list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyPageUp})
	if !result.Handled {
		t.Fatal("expected PageUp to be swallowed during drag")
	}
}

func TestList_DragOriginAndDropIndex_NoDrag(t *testing.T) {
	adapter := NewSliceAdapter([]string{"A"}, nil)
	list := NewList(adapter)
	if list.DragOrigin() != -1 {
		t.Fatalf("expected DragOrigin -1 when not dragging, got %d", list.DragOrigin())
	}
	if list.DropIndex() != -1 {
		t.Fatalf("expected DropIndex -1 when not dragging, got %d", list.DropIndex())
	}
}

func TestList_CanDrop(t *testing.T) {
	adapter := NewSliceAdapter([]string{"A"}, nil)
	list := NewList(adapter)

	data := runtime.DragData{Kind: "list-item"}
	if list.CanDrop(data) {
		t.Fatal("expected CanDrop false when not draggable")
	}

	list.SetDraggable(true)
	if !list.CanDrop(data) {
		t.Fatal("expected CanDrop true when draggable and Kind matches")
	}

	otherData := runtime.DragData{Kind: "tree-node"}
	if list.CanDrop(otherData) {
		t.Fatal("expected CanDrop false for non-matching Kind")
	}
}

func TestList_OnDrop_Interface(t *testing.T) {
	adapter := NewSliceAdapter([]string{"A"}, nil)
	list := NewList(adapter)

	data := runtime.DragData{Kind: "list-item"}
	if list.OnDrop(data) {
		t.Fatal("expected OnDrop false when not draggable")
	}

	list.SetDraggable(true)
	if !list.OnDrop(data) {
		t.Fatal("expected OnDrop true when draggable")
	}
}

func TestList_NilDraggable(t *testing.T) {
	var list *List[string]
	list.SetDraggable(true) // should not panic
	if list.IsDraggable() {
		t.Fatal("expected false for nil list")
	}
	if list.IsDragging() {
		t.Fatal("expected false for nil list")
	}
	if list.DragOrigin() != -1 {
		t.Fatal("expected -1 for nil list")
	}
	if list.DropIndex() != -1 {
		t.Fatal("expected -1 for nil list")
	}
}

func TestMutableSliceAdapter_Move(t *testing.T) {
	adapter := NewMutableSliceAdapter([]string{"A", "B", "C", "D", "E"}, nil)

	// Move B (1) to position 3.
	adapter.Move(1, 3)
	items := adapter.Items()
	expected := []string{"A", "C", "D", "B", "E"}
	for i, want := range expected {
		if items[i] != want {
			t.Fatalf("after Move(1,3): items[%d] = %q, want %q (full: %v)", i, items[i], want, items)
		}
	}
}

func TestMutableSliceAdapter_MoveToStart(t *testing.T) {
	adapter := NewMutableSliceAdapter([]string{"A", "B", "C", "D"}, nil)

	// Move D (3) to position 0.
	adapter.Move(3, 0)
	items := adapter.Items()
	expected := []string{"D", "A", "B", "C"}
	for i, want := range expected {
		if items[i] != want {
			t.Fatalf("after Move(3,0): items[%d] = %q, want %q (full: %v)", i, items[i], want, items)
		}
	}
}

func TestMutableSliceAdapter_MoveSameIndex(t *testing.T) {
	adapter := NewMutableSliceAdapter([]string{"A", "B", "C"}, nil)

	// Move to same position should be no-op.
	adapter.Move(1, 1)
	items := adapter.Items()
	expected := []string{"A", "B", "C"}
	for i, want := range expected {
		if items[i] != want {
			t.Fatalf("after Move(1,1): items[%d] = %q, want %q", i, items[i], want)
		}
	}
}

func TestMutableSliceAdapter_MoveOutOfBounds(t *testing.T) {
	adapter := NewMutableSliceAdapter([]string{"A", "B"}, nil)

	// Out of bounds should be no-op.
	adapter.Move(-1, 0)
	adapter.Move(0, -1)
	adapter.Move(5, 0)
	adapter.Move(0, 5)

	items := adapter.Items()
	if len(items) != 2 || items[0] != "A" || items[1] != "B" {
		t.Fatalf("expected unchanged items, got %v", items)
	}
}

func TestMutableSliceAdapter_NilMove(t *testing.T) {
	var adapter *MutableSliceAdapter[string]
	adapter.Move(0, 1) // should not panic
}

func TestMutableSliceAdapter_NilItems(t *testing.T) {
	var adapter *MutableSliceAdapter[string]
	if adapter.Items() != nil {
		t.Fatal("expected nil items from nil adapter")
	}
}

func TestMutableSliceAdapter_AsListAdapter(t *testing.T) {
	adapter := NewMutableSliceAdapter([]string{"X", "Y", "Z"}, nil)
	if adapter.Count() != 3 {
		t.Fatalf("expected Count 3, got %d", adapter.Count())
	}
	if adapter.Item(1) != "Y" {
		t.Fatalf("expected Item(1) 'Y', got %q", adapter.Item(1))
	}
}

func TestList_DragMoveUpward(t *testing.T) {
	list := newTestDraggableList()

	// Select D (index 3), start drag.
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlG})

	// Move up to index 1.
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})

	// Verify: D moved from index 3 to index 1.
	mutable := list.adapter.(*MutableSliceAdapter[string])
	items := mutable.Items()
	expected := []string{"A", "D", "B", "C", "E"}
	for i, want := range expected {
		if items[i] != want {
			t.Fatalf("items[%d] = %q, want %q (full: %v)", i, items[i], want, items)
		}
	}
}

func TestList_SetDragStyle(t *testing.T) {
	adapter := NewSliceAdapter([]string{"A"}, nil)
	list := NewList(adapter)
	list.SetDragStyle(list.dragStyle) // should not panic
}
