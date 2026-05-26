package widgets

import (
	"fmt"
	"strings"
	"testing"

	flufftest "m31labs.dev/fluffyui/testing"
)

// buildLargeTree creates a tree with the given branching factor and depth.
// Total nodes = (factor^depth - 1) / (factor - 1) for factor > 1.
func buildLargeTree(label string, factor, depth int) *TreeNode {
	node := &TreeNode{Label: label, Expanded: true}
	if depth <= 0 {
		return node
	}
	for i := 0; i < factor; i++ {
		child := buildLargeTree(fmt.Sprintf("%s.%d", label, i), factor, depth-1)
		node.Children = append(node.Children, child)
	}
	return node
}

func countNodes(node *TreeNode) int {
	if node == nil {
		return 0
	}
	count := 1
	if node.Expanded {
		for _, child := range node.Children {
			count += countNodes(child)
		}
	}
	return count
}

func TestTree_VirtualScroll_LargeTree(t *testing.T) {
	// Build a tree with branching factor 10 and depth 4 => ~11,111 nodes
	root := buildLargeTree("root", 10, 4)
	totalNodes := countNodes(root)
	if totalNodes < 10000 {
		t.Fatalf("expected at least 10000 nodes, got %d", totalNodes)
	}

	tree := NewTree(root)
	tree.SetVirtualScroll(true)

	output := flufftest.RenderToString(tree, 40, 25)

	// Root should be visible
	if !strings.Contains(output, "root") {
		t.Fatalf("expected 'root' in output")
	}

	// Check visible range is small relative to total
	start, end := tree.VisibleRange()
	rendered := end - start
	if rendered > 35 {
		t.Fatalf("expected ~27 rows rendered, got %d (total nodes: %d)", rendered, totalNodes)
	}
	if start != 0 {
		t.Fatalf("expected visible start 0, got %d", start)
	}
}

func TestTree_VirtualScroll_ScrollToMiddle(t *testing.T) {
	// Build a wide flat tree: root with 10,000 children
	root := &TreeNode{Label: "root", Expanded: true}
	for i := 0; i < 10000; i++ {
		root.Children = append(root.Children, &TreeNode{Label: fmt.Sprintf("child-%d", i)})
	}

	tree := NewTree(root)
	tree.SetVirtualScroll(true)

	// Select a node in the middle
	tree.ScrollTo(0, 5000)

	output := flufftest.RenderToString(tree, 30, 20)

	// The selected node should be visible
	if !strings.Contains(output, "child-4999") {
		t.Fatalf("expected 'child-4999' in output after scrolling to middle:\n%s", output)
	}

	// child-0 should NOT be visible
	if strings.Contains(output, "child-0 ") {
		t.Fatalf("did not expect 'child-0' when scrolled to middle")
	}

	start, end := tree.VisibleRange()
	rendered := end - start
	if rendered > 30 {
		t.Fatalf("expected ~22 rows rendered, got %d", rendered)
	}
}

func TestTree_VirtualScroll_DefaultOff(t *testing.T) {
	root := &TreeNode{Label: "root"}
	tree := NewTree(root)

	if tree.VirtualScroll() {
		t.Fatal("expected virtual scroll off by default")
	}

	tree.SetVirtualScroll(true)
	if !tree.VirtualScroll() {
		t.Fatal("expected virtual scroll on after enable")
	}

	tree.SetVirtualScroll(false)
	if tree.VirtualScroll() {
		t.Fatal("expected virtual scroll off after disable")
	}
}

func TestTree_VirtualScroll_ExpandCollapse(t *testing.T) {
	// Build tree: root -> 100 children, each with 100 grandchildren
	root := &TreeNode{Label: "root", Expanded: true}
	for i := 0; i < 100; i++ {
		child := &TreeNode{Label: fmt.Sprintf("child-%d", i), Expanded: false}
		for j := 0; j < 100; j++ {
			child.Children = append(child.Children, &TreeNode{Label: fmt.Sprintf("gc-%d-%d", i, j)})
		}
		root.Children = append(root.Children, child)
	}

	tree := NewTree(root)
	tree.SetVirtualScroll(true)

	// Initially, only root + 100 children = 101 rows
	flufftest.RenderToString(tree, 30, 20)
	start1, end1 := tree.VisibleRange()
	rendered1 := end1 - start1

	// Now expand first child — adds 100 grandchildren
	root.Children[0].Expanded = true
	tree.SetRoot(root) // triggers flatDirty

	flufftest.RenderToString(tree, 30, 20)
	start2, end2 := tree.VisibleRange()
	rendered2 := end2 - start2

	// Both renders should show roughly the same number of rendered rows
	// despite the second having many more total nodes
	if rendered1 > 30 {
		t.Fatalf("expected ~22 rendered rows before expand, got %d", rendered1)
	}
	if rendered2 > 30 {
		t.Fatalf("expected ~22 rendered rows after expand, got %d", rendered2)
	}
	_ = start1
	_ = start2
}

func TestTree_VirtualScroll_DeepNesting(t *testing.T) {
	// Build a deeply nested tree: depth 1000
	root := &TreeNode{Label: "level-0", Expanded: true}
	current := root
	for i := 1; i <= 1000; i++ {
		child := &TreeNode{Label: fmt.Sprintf("level-%d", i), Expanded: true}
		current.Children = append(current.Children, child)
		current = child
	}

	tree := NewTree(root)
	tree.SetVirtualScroll(true)

	output := flufftest.RenderToString(tree, 60, 20)

	// Root should be visible
	if !strings.Contains(output, "level-0") {
		t.Fatalf("expected 'level-0' in output")
	}

	start, end := tree.VisibleRange()
	rendered := end - start
	if rendered > 30 {
		t.Fatalf("expected ~22 rendered rows with deep nesting, got %d", rendered)
	}
}

func TestTree_VirtualScroll_Overscan(t *testing.T) {
	root := &TreeNode{Label: "root", Expanded: true}
	for i := 0; i < 1000; i++ {
		root.Children = append(root.Children, &TreeNode{Label: fmt.Sprintf("node-%d", i)})
	}

	tree := NewTree(root)
	tree.SetVirtualScroll(true)
	tree.SetVirtualOverscan(10)

	tree.ScrollTo(0, 500)
	flufftest.RenderToString(tree, 30, 20)

	start, end := tree.VisibleRange()
	// With overscan 10, range should extend beyond the viewport
	viewportRows := 20
	if end-start < viewportRows {
		t.Fatalf("expected rendered range > viewport (%d), got %d", viewportRows, end-start)
	}
	if end-start > viewportRows+25 {
		t.Fatalf("expected rendered range reasonable with overscan 10, got %d", end-start)
	}
}

func TestTree_VirtualScroll_NilReceiver(t *testing.T) {
	var tree *Tree
	tree.SetVirtualScroll(true)
	tree.SetVirtualOverscan(5)
	if tree.VirtualScroll() {
		t.Fatal("expected false from nil receiver")
	}
	start, end := tree.VisibleRange()
	if start != 0 || end != 0 {
		t.Fatalf("expected (0,0) from nil receiver, got (%d,%d)", start, end)
	}
}

func TestTree_VirtualScroll_EmptyTree(t *testing.T) {
	tree := NewTree(nil)
	tree.SetVirtualScroll(true)

	// Should not panic
	output := flufftest.RenderToString(tree, 20, 10)
	_ = output
}
