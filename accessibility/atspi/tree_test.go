package atspi

import (
	"testing"
)

func TestAccessibleTreeBuild(t *testing.T) {
	root := &mockAccessible{role: "window", label: "Main"}
	child1 := &mockAccessible{role: "button", label: "Button 1"}
	child2 := &mockAccessible{role: "button", label: "Button 2"}

	app := &mockApp{
		name: "TestApp",
		root: root,
		children: map[Accessible][]Accessible{
			root: {child1, child2},
		},
	}

	tree := NewAccessibleTree(app)
	if tree == nil {
		t.Fatal("NewAccessibleTree returned nil")
	}

	rootNode := tree.Root()
	if rootNode == nil {
		t.Fatal("tree root is nil")
	}

	if rootNode.Widget() != root {
		t.Error("root widget mismatch")
	}

	if len(rootNode.Children()) != 2 {
		t.Errorf("expected 2 children, got %d", len(rootNode.Children()))
	}
}

func TestAccessibleTreeNilApp(t *testing.T) {
	tree := NewAccessibleTree(nil)
	if tree == nil {
		t.Fatal("NewAccessibleTree returned nil")
	}

	if tree.Root() != nil {
		t.Error("expected nil root for nil app")
	}
}

func TestAccessibleTreeNilRoot(t *testing.T) {
	app := &mockApp{name: "Test", root: nil}
	tree := NewAccessibleTree(app)

	if tree.Root() != nil {
		t.Error("expected nil root when app has nil root")
	}
}

func TestAccessibleTreePathFor(t *testing.T) {
	root := &mockAccessible{role: "window", label: "Main"}
	child := &mockAccessible{role: "button", label: "Button"}

	app := &mockApp{
		name: "TestApp",
		root: root,
		children: map[Accessible][]Accessible{
			root: {child},
		},
	}

	tree := NewAccessibleTree(app)

	rootPath := tree.PathFor(root)
	if rootPath != "/root" {
		t.Errorf("expected root path '/root', got %q", rootPath)
	}

	childPath := tree.PathFor(child)
	if childPath != "/root/0" {
		t.Errorf("expected child path '/root/0', got %q", childPath)
	}

	// Unknown widget
	unknown := &mockAccessible{label: "Unknown"}
	unknownPath := tree.PathFor(unknown)
	if unknownPath != "" {
		t.Errorf("expected empty path for unknown widget, got %q", unknownPath)
	}
}

func TestAccessibleTreeNodeFor(t *testing.T) {
	root := &mockAccessible{role: "window", label: "Main"}
	child := &mockAccessible{role: "button", label: "Button"}

	app := &mockApp{
		name: "TestApp",
		root: root,
		children: map[Accessible][]Accessible{
			root: {child},
		},
	}

	tree := NewAccessibleTree(app)

	rootNode := tree.NodeFor(root)
	if rootNode == nil {
		t.Fatal("NodeFor(root) returned nil")
	}

	childNode := tree.NodeFor(child)
	if childNode == nil {
		t.Fatal("NodeFor(child) returned nil")
	}

	if childNode.Parent() != rootNode {
		t.Error("child parent should be root")
	}

	unknown := &mockAccessible{label: "Unknown"}
	if tree.NodeFor(unknown) != nil {
		t.Error("expected nil for unknown widget")
	}
}

func TestAccessibleTreeNodeAtPath(t *testing.T) {
	root := &mockAccessible{role: "window", label: "Main"}
	child := &mockAccessible{role: "button", label: "Button"}

	app := &mockApp{
		name: "TestApp",
		root: root,
		children: map[Accessible][]Accessible{
			root: {child},
		},
	}

	tree := NewAccessibleTree(app)

	rootNode := tree.NodeAtPath("/root")
	if rootNode == nil {
		t.Fatal("NodeAtPath('/root') returned nil")
	}

	childNode := tree.NodeAtPath("/root/0")
	if childNode == nil {
		t.Fatal("NodeAtPath('/root/0') returned nil")
	}

	if tree.NodeAtPath("/nonexistent") != nil {
		t.Error("expected nil for nonexistent path")
	}
}

func TestAccessibleTreeRebuild(t *testing.T) {
	root := &mockAccessible{role: "window", label: "Main"}
	child := &mockAccessible{role: "button", label: "Button"}

	app := &mockApp{
		name: "TestApp",
		root: root,
		children: map[Accessible][]Accessible{
			root: {child},
		},
	}

	tree := NewAccessibleTree(app)

	// Verify initial state
	if len(tree.Root().Children()) != 1 {
		t.Error("expected 1 child initially")
	}

	// Add another child and rebuild
	child2 := &mockAccessible{role: "button", label: "Button 2"}
	app.children[root] = append(app.children[root], child2)

	tree.Rebuild(app)

	if len(tree.Root().Children()) != 2 {
		t.Error("expected 2 children after rebuild")
	}
}

func TestAccessibleNodeProperties(t *testing.T) {
	checked := true
	expanded := false
	widget := &mockAccessible{
		role:        "checkbox",
		label:       "Accept Terms",
		description: "You must accept to continue",
		state: StateSet{
			Checked:  &checked,
			Expanded: &expanded,
			Selected: true,
			Disabled: true,
		},
		value: &ValueInfo{Text: "Selected"},
	}

	node := NewAccessibleNode(widget, "/test", nil, 0)

	if node.GetName() != "Accept Terms" {
		t.Errorf("GetName = %q, want 'Accept Terms'", node.GetName())
	}

	if node.GetDescription() != "You must accept to continue" {
		t.Errorf("GetDescription = %q, want description", node.GetDescription())
	}

	if node.GetRole() != roleCheckBox {
		t.Errorf("GetRole = %d, want %d", node.GetRole(), roleCheckBox)
	}

	if node.GetRoleName() != "checkbox" {
		t.Errorf("GetRoleName = %q, want 'checkbox'", node.GetRoleName())
	}

	if node.Path() != "/test" {
		t.Errorf("Path = %q, want '/test'", node.Path())
	}

	if node.Index() != 0 {
		t.Errorf("Index = %d, want 0", node.Index())
	}

	if node.Parent() != nil {
		t.Error("expected nil parent")
	}
}

func TestAccessibleNodeNilWidget(t *testing.T) {
	node := NewAccessibleNode(nil, "/test", nil, 0)

	if node.GetName() != "" {
		t.Error("expected empty name for nil widget")
	}

	if node.GetDescription() != "" {
		t.Error("expected empty description for nil widget")
	}

	if node.GetRole() != roleUnknown {
		t.Errorf("expected unknown role for nil widget, got %d", node.GetRole())
	}

	if node.GetRoleName() != "" {
		t.Error("expected empty role name for nil widget")
	}

	state := node.GetState()
	if len(state) != 2 {
		t.Error("expected 2-element state array")
	}
}

func TestAccessibleNodeChildren(t *testing.T) {
	parent := &mockAccessible{role: "list", label: "List"}
	child1 := &mockAccessible{role: "listitem", label: "Item 1"}
	child2 := &mockAccessible{role: "listitem", label: "Item 2"}

	app := &mockApp{
		name: "TestApp",
		root: parent,
		children: map[Accessible][]Accessible{
			parent: {child1, child2},
		},
	}

	tree := NewAccessibleTree(app)
	parentNode := tree.Root()

	if parentNode.GetChildCount() != 2 {
		t.Errorf("GetChildCount = %d, want 2", parentNode.GetChildCount())
	}

	child := parentNode.GetChildAtIndex(0)
	if child == nil {
		t.Fatal("GetChildAtIndex(0) returned nil")
	}
	if child.GetName() != "Item 1" {
		t.Errorf("child name = %q, want 'Item 1'", child.GetName())
	}

	if parentNode.GetChildAtIndex(-1) != nil {
		t.Error("expected nil for negative index")
	}

	if parentNode.GetChildAtIndex(100) != nil {
		t.Error("expected nil for out of bounds index")
	}
}

func TestAccessibleNodeParent(t *testing.T) {
	parent := &mockAccessible{role: "window", label: "Main"}
	child := &mockAccessible{role: "button", label: "Button"}

	app := &mockApp{
		name: "TestApp",
		root: parent,
		children: map[Accessible][]Accessible{
			parent: {child},
		},
	}

	tree := NewAccessibleTree(app)
	childNode := tree.NodeFor(child)

	bus, path := childNode.GetParent()
	if bus == "" {
		t.Error("expected non-empty bus name")
	}
	if path == "" {
		t.Error("expected non-empty parent path")
	}

	if childNode.GetIndexInParent() != 0 {
		t.Errorf("GetIndexInParent = %d, want 0", childNode.GetIndexInParent())
	}
}

func TestAccessibleNodeParentNil(t *testing.T) {
	node := NewAccessibleNode(&mockAccessible{}, "/root", nil, 0)

	bus, path := node.GetParent()
	if bus != "" || path != "" {
		t.Error("expected empty parent info for root node")
	}
}

func TestDeepTree(t *testing.T) {
	root := &mockAccessible{role: "window", label: "Window"}
	level1 := &mockAccessible{role: "group", label: "Level 1"}
	level2 := &mockAccessible{role: "group", label: "Level 2"}
	level3 := &mockAccessible{role: "button", label: "Deep Button"}

	app := &mockApp{
		name: "TestApp",
		root: root,
		children: map[Accessible][]Accessible{
			root:   {level1},
			level1: {level2},
			level2: {level3},
		},
	}

	tree := NewAccessibleTree(app)

	// Verify paths
	if tree.PathFor(root) != "/root" {
		t.Errorf("root path = %q", tree.PathFor(root))
	}
	if tree.PathFor(level1) != "/root/0" {
		t.Errorf("level1 path = %q", tree.PathFor(level1))
	}
	if tree.PathFor(level2) != "/root/0/0" {
		t.Errorf("level2 path = %q", tree.PathFor(level2))
	}
	if tree.PathFor(level3) != "/root/0/0/0" {
		t.Errorf("level3 path = %q", tree.PathFor(level3))
	}

	// Verify parent chain
	deepNode := tree.NodeFor(level3)
	if deepNode.Parent().Widget() != level2 {
		t.Error("level3 parent should be level2")
	}
	if deepNode.Parent().Parent().Widget() != level1 {
		t.Error("level2 parent should be level1")
	}
}
