//go:build windows

package uiautomation

import (
	"sync"
)

// AccessibleNode represents a node in the accessible tree.
type AccessibleNode struct {
	widget   Accessible
	parent   *AccessibleNode
	children []*AccessibleNode
	index    int
	provider *IRawElementProviderSimple
}

// NewAccessibleNode creates a new accessible node.
func NewAccessibleNode(widget Accessible, parent *AccessibleNode, index int) *AccessibleNode {
	node := &AccessibleNode{
		widget: widget,
		parent: parent,
		index:  index,
	}
	// Create the UI Automation provider for this node
	node.provider = NewProvider(node)
	return node
}

// Widget returns the underlying accessible widget.
func (n *AccessibleNode) Widget() Accessible {
	return n.widget
}

// Parent returns the parent node.
func (n *AccessibleNode) Parent() *AccessibleNode {
	return n.parent
}

// Children returns the child nodes.
func (n *AccessibleNode) Children() []*AccessibleNode {
	return n.children
}

// Index returns the index of this node in its parent.
func (n *AccessibleNode) Index() int {
	return n.index
}

// Provider returns the UI Automation provider for this node.
func (n *AccessibleNode) Provider() *IRawElementProviderSimple {
	return n.provider
}

// UI Automation property accessors

// Name returns the element name.
func (n *AccessibleNode) Name() string {
	if n.widget == nil {
		return ""
	}
	return n.widget.AccessibleLabel()
}

// Description returns the element description.
func (n *AccessibleNode) Description() string {
	if n.widget == nil {
		return ""
	}
	return n.widget.AccessibleDescription()
}

// ControlType returns the UIA control type ID.
func (n *AccessibleNode) ControlType() int32 {
	if n.widget == nil {
		return UIA_CustomControlTypeId
	}
	return roleToUIA(n.widget.AccessibleRole())
}

// Value returns the current value.
func (n *AccessibleNode) Value() string {
	if n.widget == nil {
		return ""
	}
	if v := n.widget.AccessibleValue(); v != nil {
		return v.Text
	}
	return ""
}

// IsEnabled returns true if the element is enabled.
func (n *AccessibleNode) IsEnabled() bool {
	if n.widget == nil {
		return false
	}
	return !n.widget.AccessibleState().Disabled
}

// IsSelected returns true if the element is selected.
func (n *AccessibleNode) IsSelected() bool {
	if n.widget == nil {
		return false
	}
	return n.widget.AccessibleState().Selected
}

// ToggleState returns the toggle state (0=off, 1=on, 2=indeterminate).
func (n *AccessibleNode) ToggleState() int32 {
	if n.widget == nil {
		return 0
	}
	state := n.widget.AccessibleState()
	if state.Checked == nil {
		return 0
	}
	if *state.Checked {
		return 1
	}
	return 0
}

// ExpandCollapseState returns the expand/collapse state (0=collapsed, 1=expanded, 2=partially, 3=leaf).
func (n *AccessibleNode) ExpandCollapseState() int32 {
	if n.widget == nil {
		return 3 // Leaf
	}
	state := n.widget.AccessibleState()
	if state.Expanded == nil {
		return 3 // Leaf
	}
	if *state.Expanded {
		return 1 // Expanded
	}
	return 0 // Collapsed
}

// roleToUIA converts a role string to UIA control type ID.
func roleToUIA(role string) int32 {
	switch role {
	case "button":
		return UIA_ButtonControlTypeId
	case "checkbox":
		return UIA_CheckBoxControlTypeId
	case "radio":
		return UIA_RadioButtonControlTypeId
	case "textbox":
		return UIA_EditControlTypeId
	case "list":
		return UIA_ListControlTypeId
	case "listitem":
		return UIA_ListItemControlTypeId
	case "table":
		return UIA_TableControlTypeId
	case "row", "cell":
		return UIA_DataItemControlTypeId
	case "slider":
		return UIA_SliderControlTypeId
	case "tree":
		return UIA_TreeControlTypeId
	case "treeitem":
		return UIA_TreeItemControlTypeId
	case "menu":
		return UIA_MenuControlTypeId
	case "menuitem":
		return UIA_MenuItemControlTypeId
	case "tab":
		return UIA_TabItemControlTypeId
	case "tablist":
		return UIA_TabControlTypeId
	case "tabpanel":
		return UIA_PaneControlTypeId
	case "dialog", "alert":
		return UIA_WindowControlTypeId
	case "status":
		return UIA_TextControlTypeId
	case "progressbar":
		return UIA_ProgressBarControlTypeId
	case "group":
		return UIA_GroupControlTypeId
	case "text":
		return UIA_TextControlTypeId
	case "chart":
		return UIA_ImageControlTypeId
	case "application", "window":
		return UIA_WindowControlTypeId
	default:
		return UIA_CustomControlTypeId
	}
}

// AccessibleTree manages the hierarchy of accessible nodes.
type AccessibleTree struct {
	mu    sync.RWMutex
	root  *AccessibleNode
	nodes map[Accessible]*AccessibleNode
}

// NewAccessibleTree creates a new accessible tree from the app.
func NewAccessibleTree(app BridgeApp) *AccessibleTree {
	tree := &AccessibleTree{
		nodes: make(map[Accessible]*AccessibleNode),
	}

	if app != nil {
		tree.Rebuild(app)
	}

	return tree
}

// Root returns the root accessible node.
func (t *AccessibleTree) Root() *AccessibleNode {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.root
}

// NodeFor returns the node for an accessible.
func (t *AccessibleTree) NodeFor(widget Accessible) *AccessibleNode {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.nodes[widget]
}

// Rebuild reconstructs the tree from the app.
func (t *AccessibleTree) Rebuild(app BridgeApp) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Clear existing tree
	t.nodes = make(map[Accessible]*AccessibleNode)

	if app == nil {
		t.root = nil
		return
	}

	// Build root node
	rootWidget := app.RootAccessible()
	if rootWidget == nil {
		t.root = nil
		return
	}

	t.root = t.buildNode(app, rootWidget, nil, 0)
}

// buildNode recursively builds the node tree.
func (t *AccessibleTree) buildNode(app BridgeApp, widget Accessible, parent *AccessibleNode, index int) *AccessibleNode {
	if widget == nil {
		return nil
	}

	node := NewAccessibleNode(widget, parent, index)
	t.nodes[widget] = node

	// Build children
	children := app.ChildAccessibles(widget)
	for i, child := range children {
		childNode := t.buildNode(app, child, node, i)
		if childNode != nil {
			node.children = append(node.children, childNode)
		}
	}

	return node
}
