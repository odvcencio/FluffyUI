//go:build darwin

package nsaccessibility

import (
	"sync"
)

// AccessibleNode represents a node in the accessible tree.
type AccessibleNode struct {
	widget   Accessible
	parent   *AccessibleNode
	children []*AccessibleNode
	index    int
}

// NewAccessibleNode creates a new accessible node.
func NewAccessibleNode(widget Accessible, parent *AccessibleNode, index int) *AccessibleNode {
	return &AccessibleNode{
		widget: widget,
		parent: parent,
		index:  index,
	}
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

// NSAccessibility attribute accessors

// Label returns the accessibility label (title).
func (n *AccessibleNode) Label() string {
	if n.widget == nil {
		return ""
	}
	return n.widget.AccessibleLabel()
}

// Description returns the accessibility description (help text).
func (n *AccessibleNode) Description() string {
	if n.widget == nil {
		return ""
	}
	return n.widget.AccessibleDescription()
}

// Role returns the NSAccessibility role.
func (n *AccessibleNode) Role() string {
	if n.widget == nil {
		return "AXUnknown"
	}
	return roleToNS(n.widget.AccessibleRole())
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

// IsExpanded returns the expanded state (nil if not applicable).
func (n *AccessibleNode) IsExpanded() *bool {
	if n.widget == nil {
		return nil
	}
	return n.widget.AccessibleState().Expanded
}

// roleToNS converts a role string to NSAccessibility role.
func roleToNS(role string) string {
	switch role {
	case "button":
		return "AXButton"
	case "checkbox":
		return "AXCheckBox"
	case "radio":
		return "AXRadioButton"
	case "textbox":
		return "AXTextField"
	case "list":
		return "AXList"
	case "listitem":
		return "AXStaticText" // List items are typically static text
	case "table":
		return "AXTable"
	case "row":
		return "AXRow"
	case "cell":
		return "AXCell"
	case "slider":
		return "AXSlider"
	case "tree":
		return "AXOutline"
	case "treeitem":
		return "AXRow"
	case "menu":
		return "AXMenu"
	case "menuitem":
		return "AXMenuItem"
	case "tab":
		return "AXRadioButton" // Tabs use radio button semantics
	case "tablist":
		return "AXTabGroup"
	case "tabpanel":
		return "AXGroup"
	case "dialog":
		return "AXSheet"
	case "alert":
		return "AXSheet"
	case "status":
		return "AXStaticText"
	case "progressbar":
		return "AXProgressIndicator"
	case "group":
		return "AXGroup"
	case "text":
		return "AXStaticText"
	case "chart":
		return "AXImage"
	case "application":
		return "AXApplication"
	case "window":
		return "AXWindow"
	default:
		return "AXGroup"
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
