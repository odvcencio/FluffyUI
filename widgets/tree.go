package widgets

import (
	"fmt"
	"strings"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/scroll"
	"github.com/odvcencio/fluffyui/terminal"
)

// TreeNode represents a node in a tree.
type TreeNode struct {
	Label    string
	Children []*TreeNode
	Expanded bool
}

// Tree renders a hierarchical tree.
type Tree struct {
	FocusableBase
	Root          *TreeNode
	selectedIndex int
	offset        int
	label         string
	style         backend.Style
	selectedStyle backend.Style
	indentCache   []string
	flatCache     []treeRow
	flatDirty     bool
	rootRef       *TreeNode
	services      runtime.Services

	// Virtual scrolling fields
	virtualScroll bool
	virtualList   *scroll.VirtualList
	virtualOvrscn int
	lastVisStart  int
	lastVisEnd    int
}

// NewTree creates a tree widget.
func NewTree(root *TreeNode) *Tree {
	tree := &Tree{
		Root:          root,
		selectedIndex: 0,
		label:         "Tree",
		style:         backend.DefaultStyle(),
		selectedStyle: backend.DefaultStyle().Reverse(true),
		flatDirty:     true,
		rootRef:       root,
	}
	tree.Base.Role = accessibility.RoleTree
	tree.syncA11y()
	return tree
}

// SetVirtualScroll enables or disables virtual scrolling for large trees.
// When enabled, only visible rows (plus overscan) are rendered each frame,
// which dramatically improves performance for trees with thousands of nodes.
func (t *Tree) SetVirtualScroll(enabled bool) {
	if t == nil {
		return
	}
	t.virtualScroll = enabled
	if enabled {
		t.ensureVirtualList()
	}
}

// VirtualScroll reports whether virtual scrolling is enabled.
func (t *Tree) VirtualScroll() bool {
	if t == nil {
		return false
	}
	return t.virtualScroll
}

// SetVirtualOverscan sets the number of extra rows to render above and below
// the visible area when virtual scrolling is enabled. Default is 2.
func (t *Tree) SetVirtualOverscan(count int) {
	if t == nil {
		return
	}
	if count < 0 {
		count = 0
	}
	t.virtualOvrscn = count
	if t.virtualList != nil {
		t.virtualList.SetOverscan(count)
	}
}

// VisibleRange returns the [start, end) flattened node indices currently rendered.
func (t *Tree) VisibleRange() (start, end int) {
	if t == nil {
		return 0, 0
	}
	return t.lastVisStart, t.lastVisEnd
}

func (t *Tree) ensureVirtualList() {
	if t.virtualList != nil {
		return
	}
	rows := t.flatten()
	t.virtualList = scroll.NewVirtualList(len(rows), 1, nil)
	overscan := t.virtualOvrscn
	if overscan <= 0 {
		overscan = 2
	}
	t.virtualList.SetOverscan(overscan)
}

func (t *Tree) syncVirtualList(rows []treeRow, viewportHeight int) {
	if t.virtualList == nil {
		return
	}
	t.virtualList.SetItemCount(len(rows))
	t.virtualList.SetViewportHeight(viewportHeight)
	t.virtualList.SetSelected(t.selectedIndex)
	t.virtualList.EnsureVisible(t.selectedIndex)
}

// SetStyle updates the base tree style.
func (t *Tree) SetStyle(style backend.Style) {
	if t == nil {
		return
	}
	t.style = style
}

// SetSelectedStyle updates the selected row style.
func (t *Tree) SetSelectedStyle(style backend.Style) {
	if t == nil {
		return
	}
	t.selectedStyle = style
}

// StyleType returns the selector type name.
func (t *Tree) StyleType() string {
	return "Tree"
}

// SetRoot updates the tree root and clears cached rows.
func (t *Tree) SetRoot(root *TreeNode) {
	if t == nil {
		return
	}
	t.Root = root
	t.rootRef = root
	t.flatDirty = true
	t.syncA11y()
}

// SetLabel updates the accessibility label.
func (t *Tree) SetLabel(label string) {
	if t == nil {
		return
	}
	t.label = label
	t.syncA11y()
}

// Measure returns desired size.
func (t *Tree) Measure(constraints runtime.Constraints) runtime.Size {
	return t.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		count := len(t.flatten())
		height := min(count, contentConstraints.MaxHeight)
		if height <= 0 {
			height = contentConstraints.MinHeight
		}
		return contentConstraints.Constrain(runtime.Size{Width: contentConstraints.MaxWidth, Height: height})
	})
}

// Render draws the tree.
func (t *Tree) Render(ctx runtime.RenderContext) {
	if t == nil {
		return
	}
	t.syncA11y()
	outer := t.bounds
	content := t.ContentBounds()
	if outer.Width <= 0 || outer.Height <= 0 {
		return
	}
	baseStyle := mergeBackendStyles(resolveBaseStyle(ctx, t, backend.DefaultStyle(), false), t.style)
	ctx.Buffer.Fill(outer, ' ', baseStyle)
	if content.Width <= 0 || content.Height <= 0 {
		return
	}
	rows := t.flatten()
	if len(rows) == 0 {
		return
	}
	if t.selectedIndex < 0 {
		t.selectedIndex = 0
	}
	if t.selectedIndex >= len(rows) {
		t.selectedIndex = len(rows) - 1
	}

	if t.virtualScroll {
		t.renderVirtualRows(ctx, content, rows, baseStyle)
		return
	}

	if t.selectedIndex < t.offset {
		t.offset = t.selectedIndex
	}
	if t.selectedIndex >= t.offset+content.Height {
		t.offset = t.selectedIndex - content.Height + 1
	}
	t.lastVisStart = t.offset
	t.lastVisEnd = t.offset + content.Height
	if t.lastVisEnd > len(rows) {
		t.lastVisEnd = len(rows)
	}
	for i := 0; i < content.Height; i++ {
		rowIndex := t.offset + i
		if rowIndex < 0 || rowIndex >= len(rows) {
			break
		}
		row := rows[rowIndex]
		style := baseStyle
		if rowIndex == t.selectedIndex {
			style = mergeBackendStyles(baseStyle, t.selectedStyle)
		}
		prefix := ""
		if len(row.node.Children) > 0 {
			if row.node.Expanded {
				prefix = "- "
			} else {
				prefix = "+ "
			}
		} else {
			prefix = "  "
		}
		indent := t.indent(row.depth)
		line := indent + prefix + row.node.Label
		line = truncateString(line, content.Width)
		writePadded(ctx.Buffer, content.X, content.Y+i, content.Width, line, style)
	}
}

func (t *Tree) renderVirtualRows(ctx runtime.RenderContext, content runtime.Rect, rows []treeRow, baseStyle backend.Style) {
	t.ensureVirtualList()
	t.syncVirtualList(rows, content.Height)

	start, end := t.virtualList.GetVisibleRange()
	t.offset = t.virtualList.Offset()
	t.lastVisStart = start
	t.lastVisEnd = end

	for i := start; i < end; i++ {
		if i < 0 || i >= len(rows) {
			continue
		}
		screenRow := i - t.offset
		if screenRow < 0 || screenRow >= content.Height {
			continue
		}
		row := rows[i]
		style := baseStyle
		if i == t.selectedIndex {
			style = mergeBackendStyles(baseStyle, t.selectedStyle)
		}
		prefix := ""
		if len(row.node.Children) > 0 {
			if row.node.Expanded {
				prefix = "- "
			} else {
				prefix = "+ "
			}
		} else {
			prefix = "  "
		}
		indent := t.indent(row.depth)
		line := indent + prefix + row.node.Label
		line = truncateString(line, content.Width)
		writePadded(ctx.Buffer, content.X, content.Y+screenRow, content.Width, line, style)
	}
}

// HandleMessage handles navigation and expansion.
func (t *Tree) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if t == nil || !t.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	rows := t.flatten()
	switch key.Key {
	case terminal.KeyUp:
		t.setSelected(t.selectedIndex-1, len(rows))
		return runtime.Handled()
	case terminal.KeyDown:
		t.setSelected(t.selectedIndex+1, len(rows))
		return runtime.Handled()
	case terminal.KeyLeft:
		if row := t.selectedRow(rows); row != nil && row.node.Expanded {
			row.node.Expanded = false
			t.flatDirty = true
			if announcer := t.services.Announcer(); announcer != nil {
				announcer.Announce(fmt.Sprintf("%s collapsed", row.node.Label), accessibility.PriorityPolite)
			}
		}
		return runtime.Handled()
	case terminal.KeyRight:
		if row := t.selectedRow(rows); row != nil && len(row.node.Children) > 0 {
			row.node.Expanded = true
			t.flatDirty = true
			if announcer := t.services.Announcer(); announcer != nil {
				announcer.Announce(fmt.Sprintf("%s expanded", row.node.Label), accessibility.PriorityPolite)
			}
		}
		return runtime.Handled()
	case terminal.KeyEnter:
		if row := t.selectedRow(rows); row != nil && len(row.node.Children) > 0 {
			row.node.Expanded = !row.node.Expanded
			t.flatDirty = true
			if announcer := t.services.Announcer(); announcer != nil {
				if row.node.Expanded {
					announcer.Announce(fmt.Sprintf("%s expanded", row.node.Label), accessibility.PriorityPolite)
				} else {
					announcer.Announce(fmt.Sprintf("%s collapsed", row.node.Label), accessibility.PriorityPolite)
				}
			}
		}
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

type treeRow struct {
	node  *TreeNode
	depth int
}

func (t *Tree) flatten() []treeRow {
	if t == nil || t.Root == nil {
		return nil
	}
	if t.rootRef != t.Root {
		t.rootRef = t.Root
		t.flatDirty = true
	}
	if !t.flatDirty {
		return t.flatCache
	}
	rows := t.flatCache[:0]
	var walk func(node *TreeNode, depth int)
	walk = func(node *TreeNode, depth int) {
		if node == nil {
			return
		}
		rows = append(rows, treeRow{node: node, depth: depth})
		if node.Expanded {
			for _, child := range node.Children {
				walk(child, depth+1)
			}
		}
	}
	walk(t.Root, 0)
	t.flatCache = rows
	t.flatDirty = false
	return t.flatCache
}

func (t *Tree) setSelected(index int, count int) {
	if count == 0 {
		t.selectedIndex = 0
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= count {
		index = count - 1
	}
	prev := t.selectedIndex
	t.selectedIndex = index
	t.syncA11y()
	if prev != index {
		if announcer := t.services.Announcer(); announcer != nil {
			rows := t.flatten()
			if row := t.selectedRow(rows); row != nil && row.node != nil {
				announcer.Announce(row.node.Label, accessibility.PriorityPolite)
			}
		}
	}
}

func (t *Tree) selectedRow(rows []treeRow) *treeRow {
	if t.selectedIndex < 0 || t.selectedIndex >= len(rows) {
		return nil
	}
	return &rows[t.selectedIndex]
}

// ScrollBy scrolls selection by delta.
func (t *Tree) ScrollBy(dx, dy int) {
	if t == nil || dy == 0 {
		return
	}
	rows := t.flatten()
	t.setSelected(t.selectedIndex+dy, len(rows))
	t.Invalidate()
}

// ScrollTo scrolls to an absolute row index.
func (t *Tree) ScrollTo(x, y int) {
	if t == nil {
		return
	}
	rows := t.flatten()
	t.setSelected(y, len(rows))
	t.Invalidate()
}

// PageBy scrolls by a number of pages.
func (t *Tree) PageBy(pages int) {
	if t == nil {
		return
	}
	rows := t.flatten()
	pageSize := t.bounds.Height
	if pageSize < 1 {
		pageSize = 1
	}
	t.setSelected(t.selectedIndex+pages*pageSize, len(rows))
	t.Invalidate()
}

func (t *Tree) indent(depth int) string {
	if depth <= 0 {
		return ""
	}
	if len(t.indentCache) == 0 {
		t.indentCache = []string{""}
	}
	for len(t.indentCache) <= depth {
		t.indentCache = append(t.indentCache, t.indentCache[len(t.indentCache)-1]+"  ")
	}
	return t.indentCache[depth]
}

// ScrollToStart scrolls to the first row.
func (t *Tree) ScrollToStart() {
	if t == nil {
		return
	}
	rows := t.flatten()
	t.setSelected(0, len(rows))
	t.Invalidate()
}

// ScrollToEnd scrolls to the last row.
func (t *Tree) ScrollToEnd() {
	if t == nil {
		return
	}
	rows := t.flatten()
	t.setSelected(len(rows)-1, len(rows))
	t.Invalidate()
}

func (t *Tree) syncA11y() {
	if t == nil {
		return
	}
	if t.Base.Role == "" {
		t.Base.Role = accessibility.RoleTree
	}
	label := strings.TrimSpace(t.label)
	if label == "" {
		label = "Tree"
	}
	t.Base.Label = label
	rows := t.flatten()
	t.Base.Description = fmt.Sprintf("%d items", len(rows))
	if row := t.selectedRow(rows); row != nil && row.node != nil {
		t.Base.Value = &accessibility.ValueInfo{Text: row.node.Label}
		t.Base.Level = row.depth + 1 // aria-level is 1-based
		if len(row.node.Children) > 0 {
			t.Base.State.Expanded = accessibility.BoolPtr(row.node.Expanded)
		} else {
			t.Base.State.Expanded = nil
		}
	} else {
		t.Base.Value = nil
		t.Base.State.Expanded = nil
		t.Base.Level = 0
	}
}

var _ scroll.Controller = (*Tree)(nil)

// Bind attaches app services.
func (t *Tree) Bind(services runtime.Services) {
	if t == nil {
		return
	}
	t.services = services
}

// Unbind releases app services.
func (t *Tree) Unbind() {
	if t == nil {
		return
	}
	t.services = runtime.Services{}
}

var _ runtime.Widget = (*Tree)(nil)
var _ runtime.Focusable = (*Tree)(nil)
var _ runtime.Bindable = (*Tree)(nil)
var _ runtime.Unbindable = (*Tree)(nil)
