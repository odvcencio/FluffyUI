//go:build linux

package atspi

import (
	"fmt"
	"sync"
)

// AccessibleNode represents a node in the accessible tree.
type AccessibleNode struct {
	widget   Accessible
	path     string
	parent   *AccessibleNode
	children []*AccessibleNode
	index    int
}

// NewAccessibleNode creates a new accessible node.
func NewAccessibleNode(widget Accessible, path string, parent *AccessibleNode, index int) *AccessibleNode {
	return &AccessibleNode{
		widget: widget,
		path:   path,
		parent: parent,
		index:  index,
	}
}

// Widget returns the underlying accessible widget.
func (n *AccessibleNode) Widget() Accessible {
	return n.widget
}

// Path returns the D-Bus object path for this node.
func (n *AccessibleNode) Path() string {
	return n.path
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

// AT-SPI Accessible interface implementation

// GetName returns the accessible name.
func (n *AccessibleNode) GetName() string {
	if n.widget == nil {
		return ""
	}
	return n.widget.AccessibleLabel()
}

// GetDescription returns the accessible description.
func (n *AccessibleNode) GetDescription() string {
	if n.widget == nil {
		return ""
	}
	return n.widget.AccessibleDescription()
}

// GetRole returns the AT-SPI role constant.
func (n *AccessibleNode) GetRole() uint32 {
	if n.widget == nil {
		return roleUnknown
	}
	return roleToATSPI(n.widget.AccessibleRole())
}

// GetRoleName returns the role name string.
func (n *AccessibleNode) GetRoleName() string {
	if n.widget == nil {
		return ""
	}
	return n.widget.AccessibleRole()
}

// GetState returns the AT-SPI state set.
func (n *AccessibleNode) GetState() []uint32 {
	if n.widget == nil {
		return []uint32{0, 0}
	}
	return stateToATSPI(n.widget.AccessibleState())
}

// GetChildCount returns the number of children.
func (n *AccessibleNode) GetChildCount() int32 {
	return int32(len(n.children))
}

// GetChildAtIndex returns the child at the given index.
func (n *AccessibleNode) GetChildAtIndex(index int32) *AccessibleNode {
	if index < 0 || int(index) >= len(n.children) {
		return nil
	}
	return n.children[index]
}

// GetParent returns the parent reference.
func (n *AccessibleNode) GetParent() (string, string) {
	if n.parent == nil {
		return "", ""
	}
	return atspiRegistry, n.parent.path
}

// GetIndexInParent returns the index of this node in parent.
func (n *AccessibleNode) GetIndexInParent() int32 {
	return int32(n.index)
}

// AT-SPI role constants
const (
	roleUnknown       uint32 = 0
	rolePushButton    uint32 = 62
	roleCheckBox      uint32 = 22
	roleRadioButton   uint32 = 65
	roleEntry         uint32 = 30
	roleList          uint32 = 48
	roleListItem      uint32 = 49
	roleTable         uint32 = 81
	roleTableRow      uint32 = 82
	roleTableCell     uint32 = 83
	roleSlider        uint32 = 75
	roleTree          uint32 = 87
	roleTreeItem      uint32 = 88
	roleMenu          uint32 = 53
	roleMenuItem      uint32 = 54
	rolePageTab       uint32 = 61
	rolePageTabList   uint32 = 60
	rolePanel         uint32 = 59
	roleDialog        uint32 = 25
	roleAlert         uint32 = 5
	roleStatusBar     uint32 = 77
	roleProgressBar   uint32 = 63
	roleFrame         uint32 = 35
	roleLabel         uint32 = 47
	roleApplication   uint32 = 7
	roleWindow        uint32 = 91
	roleFiller        uint32 = 33
	roleImage         uint32 = 43
)

// roleToATSPI converts a role string to AT-SPI role constant.
func roleToATSPI(role string) uint32 {
	switch role {
	case "button":
		return rolePushButton
	case "checkbox":
		return roleCheckBox
	case "radio":
		return roleRadioButton
	case "textbox":
		return roleEntry
	case "list":
		return roleList
	case "listitem":
		return roleListItem
	case "table":
		return roleTable
	case "row":
		return roleTableRow
	case "cell":
		return roleTableCell
	case "slider":
		return roleSlider
	case "tree":
		return roleTree
	case "treeitem":
		return roleTreeItem
	case "menu":
		return roleMenu
	case "menuitem":
		return roleMenuItem
	case "tab":
		return rolePageTab
	case "tablist":
		return rolePageTabList
	case "tabpanel":
		return rolePanel
	case "dialog":
		return roleDialog
	case "alert":
		return roleAlert
	case "status":
		return roleStatusBar
	case "progressbar":
		return roleProgressBar
	case "group":
		return roleFrame
	case "text":
		return roleLabel
	case "chart":
		return roleImage
	case "application":
		return roleApplication
	case "window":
		return roleWindow
	default:
		return roleFiller
	}
}

// AT-SPI state bits (these are bit positions in a 64-bit state set)
// States 0-31 are in the low word, states 32-63 are in the high word.
const (
	stateActive       = 1
	stateChecked      = 3
	stateEditable     = 6
	stateEnabled      = 7
	stateExpanded     = 9
	stateFocusable    = 10
	stateFocused      = 11
	stateRequired     = 25
	stateSelected     = 26
	stateSensitive    = 27
	stateShowing      = 28
	stateVisible      = 32 // In high word
	stateInvalidEntry = 45 // In high word
)

// stateToATSPI converts StateSet to AT-SPI state bits.
func stateToATSPI(state StateSet) []uint32 {
	// AT-SPI uses two 32-bit words for state
	var low, high uint32

	// Default states - visible and showing
	high |= 1 << (stateVisible - 32) // stateVisible is bit 32, so shift by 0 in high word
	low |= 1 << stateShowing

	if !state.Disabled {
		low |= 1 << stateEnabled
		low |= 1 << stateSensitive
	}

	if state.Checked != nil && *state.Checked {
		low |= 1 << stateChecked
	}

	if state.Expanded != nil && *state.Expanded {
		low |= 1 << stateExpanded
	}

	if state.Selected {
		low |= 1 << stateSelected
	}

	if !state.ReadOnly {
		low |= 1 << stateEditable
	}

	if state.Required {
		low |= 1 << stateRequired
	}

	if state.Invalid {
		high |= 1 << (stateInvalidEntry - 32)
	}

	return []uint32{low, high}
}

// AccessibleTree manages the hierarchy of accessible nodes.
type AccessibleTree struct {
	mu      sync.RWMutex
	root    *AccessibleNode
	nodes   map[Accessible]*AccessibleNode
	pathMap map[string]*AccessibleNode
}

// NewAccessibleTree creates a new accessible tree from the app.
func NewAccessibleTree(app BridgeApp) *AccessibleTree {
	tree := &AccessibleTree{
		nodes:   make(map[Accessible]*AccessibleNode),
		pathMap: make(map[string]*AccessibleNode),
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

// PathFor returns the D-Bus path for an accessible.
func (t *AccessibleTree) PathFor(widget Accessible) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if node, ok := t.nodes[widget]; ok {
		return node.path
	}
	return ""
}

// NodeFor returns the node for an accessible.
func (t *AccessibleTree) NodeFor(widget Accessible) *AccessibleNode {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.nodes[widget]
}

// NodeAtPath returns the node at the given path.
func (t *AccessibleTree) NodeAtPath(path string) *AccessibleNode {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.pathMap[path]
}

// Rebuild reconstructs the tree from the app.
func (t *AccessibleTree) Rebuild(app BridgeApp) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Clear existing tree
	t.nodes = make(map[Accessible]*AccessibleNode)
	t.pathMap = make(map[string]*AccessibleNode)

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

	t.root = t.buildNode(app, rootWidget, "", nil, 0)
}

// buildNode recursively builds the node tree.
func (t *AccessibleTree) buildNode(app BridgeApp, widget Accessible, basePath string, parent *AccessibleNode, index int) *AccessibleNode {
	if widget == nil {
		return nil
	}

	path := fmt.Sprintf("%s/%d", basePath, index)
	if basePath == "" {
		path = "/root"
	}

	node := NewAccessibleNode(widget, path, parent, index)
	t.nodes[widget] = node
	t.pathMap[path] = node

	// Build children
	children := app.ChildAccessibles(widget)
	for i, child := range children {
		childNode := t.buildNode(app, child, path, node, i)
		if childNode != nil {
			node.children = append(node.children, childNode)
		}
	}

	return node
}
