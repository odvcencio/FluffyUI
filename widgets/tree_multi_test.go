package widgets

import (
	"sort"
	"testing"

	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/terminal"
)

// newTestTree creates a simple tree for testing:
//
//	root (expanded)
//	  child-0
//	  child-1 (expanded)
//	    grandchild-0
//	    grandchild-1
//	  child-2
func newTestTree() *Tree {
	root := &TreeNode{
		Label:    "root",
		Expanded: true,
		Children: []*TreeNode{
			{Label: "child-0"},
			{
				Label:    "child-1",
				Expanded: true,
				Children: []*TreeNode{
					{Label: "grandchild-0"},
					{Label: "grandchild-1"},
				},
			},
			{Label: "child-2"},
		},
	}
	tree := NewTree(root)
	tree.Focus()
	tree.Measure(runtime.Constraints{MaxWidth: 40, MaxHeight: 20})
	tree.Layout(runtime.Rect{X: 0, Y: 0, Width: 40, Height: 20})
	return tree
}

// flatPaths returns the paths of all visible rows in order.
func flatPaths(t *Tree) []string {
	rows := t.flatten()
	paths := make([]string, len(rows))
	for i, r := range rows {
		paths[i] = r.path
	}
	return paths
}

// ---------- Multi-selection tests ----------

func TestTree_SetMultiSelect(t *testing.T) {
	tree := newTestTree()
	if tree.MultiSelect() {
		t.Fatal("expected multi-select off by default")
	}
	tree.SetMultiSelect(true)
	if !tree.MultiSelect() {
		t.Fatal("expected multi-select on after SetMultiSelect(true)")
	}
	tree.SetMultiSelect(false)
	if tree.MultiSelect() {
		t.Fatal("expected multi-select off after SetMultiSelect(false)")
	}
}

func TestTree_SetMultiSelect_NilReceiver(t *testing.T) {
	var tree *Tree
	tree.SetMultiSelect(true) // should not panic
	if tree.MultiSelect() {
		t.Fatal("expected false from nil receiver")
	}
}

func TestTree_SpaceTogglesSelection(t *testing.T) {
	tree := newTestTree()
	tree.SetMultiSelect(true)

	// Focus is on root (index 0). Press Space to select it.
	result := tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: ' '})
	if !result.Handled {
		t.Fatal("expected Space to be handled")
	}
	selected := tree.SelectedNodes()
	if len(selected) != 1 || selected[0] != "root" {
		t.Fatalf("expected [root], got %v", selected)
	}

	// Press Space again to deselect.
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: ' '})
	selected = tree.SelectedNodes()
	if len(selected) != 0 {
		t.Fatalf("expected empty selection after toggle off, got %v", selected)
	}
}

func TestTree_SpaceIgnoredWithoutMultiSelect(t *testing.T) {
	tree := newTestTree()
	// multiSelect is off by default.
	result := tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: ' '})
	if result.Handled {
		t.Fatal("expected Space to be unhandled when multiSelect is off")
	}
}

func TestTree_SelectedNodes_ReturnsVisualOrder(t *testing.T) {
	tree := newTestTree()
	tree.SetMultiSelect(true)

	// Select nodes out of visual order.
	tree.SelectNode("root/child-2")
	tree.SelectNode("root")
	tree.SelectNode("root/child-1/grandchild-0")

	selected := tree.SelectedNodes()
	expected := []string{"root", "root/child-1/grandchild-0", "root/child-2"}
	if len(selected) != len(expected) {
		t.Fatalf("expected %d selected, got %d: %v", len(expected), len(selected), selected)
	}
	for i, want := range expected {
		if selected[i] != want {
			t.Fatalf("selected[%d] = %q, want %q", i, selected[i], want)
		}
	}
}

func TestTree_SelectNode_DeselectNode(t *testing.T) {
	tree := newTestTree()
	tree.SetMultiSelect(true)

	tree.SelectNode("root/child-0")
	tree.SelectNode("root/child-2")

	selected := tree.SelectedNodes()
	if len(selected) != 2 {
		t.Fatalf("expected 2 selected nodes, got %d", len(selected))
	}

	tree.DeselectNode("root/child-0")
	selected = tree.SelectedNodes()
	if len(selected) != 1 || selected[0] != "root/child-2" {
		t.Fatalf("expected [root/child-2] after deselect, got %v", selected)
	}
}

func TestTree_SelectNode_IgnoredWithoutMultiSelect(t *testing.T) {
	tree := newTestTree()
	tree.SelectNode("root/child-0")
	selected := tree.SelectedNodes()
	if len(selected) != 0 {
		t.Fatalf("expected no selections when multiSelect is off, got %v", selected)
	}
}

func TestTree_SelectAll(t *testing.T) {
	tree := newTestTree()
	tree.SetMultiSelect(true)
	tree.SelectAll()

	selected := tree.SelectedNodes()
	allPaths := flatPaths(tree)
	if len(selected) != len(allPaths) {
		t.Fatalf("expected %d selected (all), got %d", len(allPaths), len(selected))
	}
	// Verify they match.
	for i, want := range allPaths {
		if selected[i] != want {
			t.Fatalf("selected[%d] = %q, want %q", i, selected[i], want)
		}
	}
}

func TestTree_DeselectAll(t *testing.T) {
	tree := newTestTree()
	tree.SetMultiSelect(true)
	tree.SelectAll()
	tree.DeselectAll()

	selected := tree.SelectedNodes()
	if len(selected) != 0 {
		t.Fatalf("expected 0 selected after DeselectAll, got %d", len(selected))
	}
}

func TestTree_CtrlA_SelectsAll(t *testing.T) {
	tree := newTestTree()
	tree.SetMultiSelect(true)

	result := tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'a', Ctrl: true})
	if !result.Handled {
		t.Fatal("expected Ctrl+A to be handled")
	}

	selected := tree.SelectedNodes()
	allPaths := flatPaths(tree)
	if len(selected) != len(allPaths) {
		t.Fatalf("expected all %d nodes selected, got %d", len(allPaths), len(selected))
	}
}

func TestTree_CtrlA_IgnoredWithoutMultiSelect(t *testing.T) {
	tree := newTestTree()
	result := tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'a', Ctrl: true})
	if result.Handled {
		t.Fatal("expected Ctrl+A to be unhandled when multiSelect is off")
	}
}

func TestTree_ShiftDown_ExtendsSelection(t *testing.T) {
	tree := newTestTree()
	tree.SetMultiSelect(true)

	// Start at root (index 0). Shift+Down should select root and child-0.
	result := tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown, Shift: true})
	if !result.Handled {
		t.Fatal("expected Shift+Down to be handled")
	}

	selected := tree.SelectedNodes()
	sort.Strings(selected) // sort for deterministic comparison
	if len(selected) != 2 {
		t.Fatalf("expected 2 selected after Shift+Down, got %d: %v", len(selected), selected)
	}

	// Shift+Down again: now root, child-0, child-1 should be selected.
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown, Shift: true})
	selected = tree.SelectedNodes()
	if len(selected) != 3 {
		t.Fatalf("expected 3 selected after 2x Shift+Down, got %d: %v", len(selected), selected)
	}
}

func TestTree_ShiftUp_ExtendsSelection(t *testing.T) {
	tree := newTestTree()
	tree.SetMultiSelect(true)

	// Move to child-2 (last visible row, index 5).
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})

	// Shift+Up from child-2.
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp, Shift: true})
	selected := tree.SelectedNodes()
	if len(selected) != 2 {
		t.Fatalf("expected 2 selected after Shift+Up, got %d: %v", len(selected), selected)
	}
}

func TestTree_ShiftDown_IgnoredWithoutMultiSelect(t *testing.T) {
	tree := newTestTree()
	// Without multiSelect, Shift+Down should just move cursor normally.
	result := tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown, Shift: true})
	if !result.Handled {
		t.Fatal("expected Down to be handled")
	}
	selected := tree.SelectedNodes()
	if len(selected) != 0 {
		t.Fatalf("expected no multi-selection without multiSelect, got %v", selected)
	}
}

// ---------- ARIA accessibility tests ----------

func TestTree_ARIA_Multiselectable(t *testing.T) {
	tree := newTestTree()

	// Off by default.
	if tree.Base.State.Multiselectable {
		t.Fatal("expected Multiselectable false by default")
	}

	tree.SetMultiSelect(true)
	if !tree.Base.State.Multiselectable {
		t.Fatal("expected Multiselectable true after SetMultiSelect(true)")
	}

	tree.SetMultiSelect(false)
	if tree.Base.State.Multiselectable {
		t.Fatal("expected Multiselectable false after SetMultiSelect(false)")
	}
}

func TestTree_ARIA_SelectedState(t *testing.T) {
	tree := newTestTree()
	tree.SetMultiSelect(true)

	// Focus is on root but root is not multi-selected yet.
	if tree.Base.State.Selected {
		t.Fatal("expected Selected false when node is not in selectedNodes")
	}

	tree.SelectNode("root")
	tree.syncA11y()
	if !tree.Base.State.Selected {
		t.Fatal("expected Selected true after selecting the focused node")
	}
}

// ---------- Drag-and-drop tests ----------

func TestTree_SetDraggable(t *testing.T) {
	tree := newTestTree()
	if tree.IsDraggable() {
		t.Fatal("expected draggable false by default")
	}
	tree.SetDraggable(true)
	if !tree.IsDraggable() {
		t.Fatal("expected draggable true after SetDraggable(true)")
	}
	tree.SetDraggable(false)
	if tree.IsDraggable() {
		t.Fatal("expected draggable false after SetDraggable(false)")
	}
}

func TestTree_SetDraggable_NilReceiver(t *testing.T) {
	var tree *Tree
	tree.SetDraggable(true)
	if tree.IsDraggable() {
		t.Fatal("expected false from nil receiver")
	}
	if tree.IsDragging() {
		t.Fatal("expected false from nil receiver")
	}
	if tree.DragOrigin() != -1 {
		t.Fatalf("expected -1 from nil receiver, got %d", tree.DragOrigin())
	}
	if tree.DropIndex() != -1 {
		t.Fatalf("expected -1 from nil receiver, got %d", tree.DropIndex())
	}
}

func TestTree_CtrlG_StartsDrag(t *testing.T) {
	tree := newTestTree()
	tree.SetDraggable(true)

	// Move to child-0 (index 1).
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})

	result := tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlG})
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
	if startMsg.Data.Kind != "tree-node" {
		t.Fatalf("expected Kind 'tree-node', got %q", startMsg.Data.Kind)
	}
	if startMsg.Data.Label != "child-0" {
		t.Fatalf("expected Label 'child-0', got %q", startMsg.Data.Label)
	}
	if !tree.IsDragging() {
		t.Fatal("expected tree to be in dragging state")
	}
	if tree.DragOrigin() != 1 {
		t.Fatalf("expected drag origin 1, got %d", tree.DragOrigin())
	}
	if tree.DropIndex() != 1 {
		t.Fatalf("expected drop index 1, got %d", tree.DropIndex())
	}
}

func TestTree_CtrlG_NotDraggable(t *testing.T) {
	tree := newTestTree()
	result := tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlG})
	if result.Handled {
		t.Fatal("expected Ctrl+G to be unhandled when not draggable")
	}
}

func TestTree_DragArrowKeys_MoveDropIndicator(t *testing.T) {
	tree := newTestTree()
	tree.SetDraggable(true)

	// Start at root (index 0).
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlG})

	if tree.DropIndex() != 0 {
		t.Fatalf("expected drop index 0, got %d", tree.DropIndex())
	}

	// Move down.
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if tree.DropIndex() != 1 {
		t.Fatalf("expected drop index 1 after down, got %d", tree.DropIndex())
	}

	// Move down more.
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if tree.DropIndex() != 3 {
		t.Fatalf("expected drop index 3, got %d", tree.DropIndex())
	}

	// Move up.
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if tree.DropIndex() != 2 {
		t.Fatalf("expected drop index 2 after up, got %d", tree.DropIndex())
	}

	// Move up past beginning should clamp.
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if tree.DropIndex() != 0 {
		t.Fatalf("expected drop index 0 at boundary, got %d", tree.DropIndex())
	}
}

func TestTree_DragHomeEnd(t *testing.T) {
	tree := newTestTree()
	tree.SetDraggable(true)

	// Start drag at index 2 (child-1).
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlG})

	// End should jump to last.
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnd})
	rows := tree.flatten()
	if tree.DropIndex() != len(rows)-1 {
		t.Fatalf("expected drop index %d after End, got %d", len(rows)-1, tree.DropIndex())
	}

	// Home should jump to 0.
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyHome})
	if tree.DropIndex() != 0 {
		t.Fatalf("expected drop index 0 after Home, got %d", tree.DropIndex())
	}
}

func TestTree_EnterCompletesDrop(t *testing.T) {
	tree := newTestTree()
	tree.SetDraggable(true)

	var capturedSources []string
	var capturedTarget string
	tree.SetOnNodeDrop(func(sources []string, target string) {
		capturedSources = sources
		capturedTarget = target
	})

	// Start drag at root (index 0).
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlG})

	// Move to child-1 (index 2).
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})

	result := tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
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
		t.Fatal("expected Cancelled false")
	}
	if endMsg.Target != tree {
		t.Fatal("expected Target to be the tree")
	}
	if tree.IsDragging() {
		t.Fatal("expected dragging false after drop")
	}

	// Verify callback was invoked.
	if len(capturedSources) != 1 || capturedSources[0] != "root" {
		t.Fatalf("expected source [root], got %v", capturedSources)
	}
	if capturedTarget != "root/child-1" {
		t.Fatalf("expected target 'root/child-1', got %q", capturedTarget)
	}
}

func TestTree_EscapeCancelsDrag(t *testing.T) {
	tree := newTestTree()
	tree.SetDraggable(true)

	var callbackCalled bool
	tree.SetOnNodeDrop(func(sources []string, target string) {
		callbackCalled = true
	})

	// Start drag at child-0 (index 1).
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlG})

	// Move drop indicator.
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})

	// Cancel with Escape.
	result := tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})
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
		t.Fatal("expected Cancelled true")
	}
	if endMsg.Target != nil {
		t.Fatal("expected Target nil for cancelled drag")
	}
	if tree.IsDragging() {
		t.Fatal("expected dragging false after cancel")
	}
	// Callback should NOT be called on cancel.
	if callbackCalled {
		t.Fatal("expected onNodeDrop not called on cancel")
	}
	// Selected index should be restored to drag origin.
	if tree.selectedIndex != 1 {
		t.Fatalf("expected selected restored to 1, got %d", tree.selectedIndex)
	}
}

func TestTree_DragSwallowsOtherKeys(t *testing.T) {
	tree := newTestTree()
	tree.SetDraggable(true)
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlG})

	// PageDown should be swallowed.
	result := tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyPageDown})
	if !result.Handled {
		t.Fatal("expected PageDown to be swallowed during drag")
	}
}

func TestTree_DnD_WithMultipleSelectedNodes(t *testing.T) {
	tree := newTestTree()
	tree.SetMultiSelect(true)
	tree.SetDraggable(true)

	// Select child-0 and child-2.
	tree.SelectNode("root/child-0")
	tree.SelectNode("root/child-2")

	// Move cursor to child-0 (index 1) and start drag.
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	result := tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyCtrlG})
	if !result.Handled {
		t.Fatal("expected Ctrl+G to be handled")
	}

	startMsg, ok := result.Commands[0].(runtime.DragStartMsg)
	if !ok {
		t.Fatalf("expected DragStartMsg, got %T", result.Commands[0])
	}

	// The drag should carry the multi-selected paths.
	paths, ok := startMsg.Data.Value.([]string)
	if !ok {
		t.Fatalf("expected Value []string, got %T", startMsg.Data.Value)
	}
	if len(paths) != 2 {
		t.Fatalf("expected 2 paths in drag data, got %d: %v", len(paths), paths)
	}
	// Paths should be in visual order.
	if paths[0] != "root/child-0" || paths[1] != "root/child-2" {
		t.Fatalf("expected [root/child-0, root/child-2], got %v", paths)
	}

	// Now complete the drop at a target.
	var capturedSources []string
	var capturedTarget string
	tree.SetOnNodeDrop(func(sources []string, target string) {
		capturedSources = sources
		capturedTarget = target
	})

	// Move to grandchild-0 (index 3) and drop.
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})

	if len(capturedSources) != 2 {
		t.Fatalf("expected 2 source paths, got %d: %v", len(capturedSources), capturedSources)
	}
	if capturedTarget != "root/child-1/grandchild-0" {
		t.Fatalf("expected target 'root/child-1/grandchild-0', got %q", capturedTarget)
	}
}

func TestTree_CanDrop(t *testing.T) {
	tree := newTestTree()

	data := runtime.DragData{Kind: "tree-node"}
	if tree.CanDrop(data) {
		t.Fatal("expected CanDrop false when not draggable")
	}

	tree.SetDraggable(true)
	if !tree.CanDrop(data) {
		t.Fatal("expected CanDrop true for tree-node kind")
	}

	otherData := runtime.DragData{Kind: "list-item"}
	if tree.CanDrop(otherData) {
		t.Fatal("expected CanDrop false for non-matching kind")
	}
}

func TestTree_OnDrop_Interface(t *testing.T) {
	tree := newTestTree()

	data := runtime.DragData{Kind: "tree-node"}
	if tree.OnDrop(data) {
		t.Fatal("expected OnDrop false when not draggable")
	}

	tree.SetDraggable(true)
	if !tree.OnDrop(data) {
		t.Fatal("expected OnDrop true when draggable")
	}
}

func TestTree_DragOriginDropIndex_NoDrag(t *testing.T) {
	tree := newTestTree()
	if tree.DragOrigin() != -1 {
		t.Fatalf("expected DragOrigin -1, got %d", tree.DragOrigin())
	}
	if tree.DropIndex() != -1 {
		t.Fatalf("expected DropIndex -1, got %d", tree.DropIndex())
	}
}

func TestTree_NodePath_Structure(t *testing.T) {
	tree := newTestTree()
	paths := flatPaths(tree)
	expected := []string{
		"root",
		"root/child-0",
		"root/child-1",
		"root/child-1/grandchild-0",
		"root/child-1/grandchild-1",
		"root/child-2",
	}
	if len(paths) != len(expected) {
		t.Fatalf("expected %d paths, got %d: %v", len(expected), len(paths), paths)
	}
	for i, want := range expected {
		if paths[i] != want {
			t.Fatalf("paths[%d] = %q, want %q", i, paths[i], want)
		}
	}
}

func TestTree_MultiSelect_ClearsOnDisable(t *testing.T) {
	tree := newTestTree()
	tree.SetMultiSelect(true)
	tree.SelectAll()

	if len(tree.SelectedNodes()) == 0 {
		t.Fatal("expected nodes selected")
	}

	tree.SetMultiSelect(false)
	if len(tree.SelectedNodes()) != 0 {
		t.Fatalf("expected selections cleared when multiSelect disabled, got %v", tree.SelectedNodes())
	}
}

func TestTree_ShiftDown_AtBoundary(t *testing.T) {
	tree := newTestTree()
	tree.SetMultiSelect(true)

	// Move to last row.
	rows := tree.flatten()
	for i := 0; i < len(rows)-1; i++ {
		tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	}

	// Shift+Down at the last row should clamp.
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown, Shift: true})
	selected := tree.SelectedNodes()
	// Should select last row only (anchor=last, newIdx=last).
	if len(selected) != 1 {
		t.Fatalf("expected 1 selected at boundary, got %d: %v", len(selected), selected)
	}
}

func TestTree_SetDragStyle(t *testing.T) {
	tree := newTestTree()
	tree.SetDragStyle(tree.dragStyle) // should not panic
}

func TestTree_SetOnNodeDrop(t *testing.T) {
	tree := newTestTree()
	called := false
	tree.SetOnNodeDrop(func(sources []string, target string) {
		called = true
	})
	// Just verify it was set (callback tested in drop tests).
	if tree.onNodeDrop == nil {
		t.Fatal("expected onNodeDrop to be set")
	}
	_ = called
}

// Verify compile-time interface assertions.
func TestTree_InterfaceAssertions(t *testing.T) {
	tree := newTestTree()
	var _ runtime.DropTarget = tree
	var _ runtime.DragSource = tree
}
