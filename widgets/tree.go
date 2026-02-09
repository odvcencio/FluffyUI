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

// treePath builds the slash-separated path for a node during flattening.
// The root node has path equal to its label; children are "parent/child".
func treeBuildPath(parentPath, label string) string {
	if parentPath == "" {
		return label
	}
	return parentPath + "/" + label
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

	// Multi-selection fields
	multiSelect   bool
	selectedNodes map[string]bool // keyed by node path
	selAnchor     int            // anchor index for shift-range selection

	// Drag-and-drop fields
	draggable  bool
	dragging   bool
	dragOrigin int // flat index where the drag started
	dropIndex  int // current drop target flat index
	dragStyle  backend.Style
	onNodeDrop func(sourcePaths []string, targetPath string)
}

// NewTree creates a hierarchical tree view rooted at the given node.
// Nodes can be expanded/collapsed with Enter and navigated with arrow keys.
func NewTree(root *TreeNode) *Tree {
	tree := &Tree{
		Root:          root,
		selectedIndex: 0,
		label:         "Tree",
		style:         backend.DefaultStyle(),
		selectedStyle: backend.DefaultStyle().Reverse(true),
		flatDirty:     true,
		rootRef:       root,
		selectedNodes: make(map[string]bool),
		selAnchor:     -1,
		dragOrigin:    -1,
		dropIndex:     -1,
		dragStyle:     backend.DefaultStyle().Reverse(true).Bold(true),
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
		if t.dragging && rowIndex == t.dropIndex {
			// Drop indicator during drag.
			style = mergeBackendStyles(baseStyle, t.dragStyle)
		} else if t.dragging && rowIndex == t.dragOrigin {
			// Origin node during drag (dimmed).
			style = mergeBackendStyles(baseStyle, t.dragStyle)
		} else if t.isNodeSelected(rows, rowIndex) {
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
		if t.dragging && i == t.dropIndex {
			style = mergeBackendStyles(baseStyle, t.dragStyle)
		} else if t.dragging && i == t.dragOrigin {
			style = mergeBackendStyles(baseStyle, t.dragStyle)
		} else if t.isNodeSelected(rows, i) {
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

	// Handle drag-in-progress keys first.
	if t.dragging {
		return t.handleDragKey(key, rows)
	}

	switch key.Key {
	case terminal.KeyUp:
		if key.Shift && t.multiSelect {
			return t.handleShiftArrow(rows, -1)
		}
		t.setSelected(t.selectedIndex-1, len(rows))
		return runtime.Handled()
	case terminal.KeyDown:
		if key.Shift && t.multiSelect {
			return t.handleShiftArrow(rows, 1)
		}
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
	case terminal.KeyRune:
		// Space toggles multi-selection on the current node.
		if key.Rune == ' ' && t.multiSelect {
			t.toggleNodeSelection(rows, t.selectedIndex)
			t.selAnchor = t.selectedIndex
			t.syncA11y()
			if announcer := t.services.Announcer(); announcer != nil {
				if row := t.selectedRow(rows); row != nil {
					if t.selectedNodes[row.path] {
						announcer.Announce(fmt.Sprintf("%s selected", row.node.Label), accessibility.PriorityPolite)
					} else {
						announcer.Announce(fmt.Sprintf("%s deselected", row.node.Label), accessibility.PriorityPolite)
					}
				}
			}
			return runtime.Handled()
		}
		// Ctrl+A selects all visible nodes.
		if key.Rune == 'a' && key.Ctrl && t.multiSelect {
			t.SelectAll()
			if announcer := t.services.Announcer(); announcer != nil {
				announcer.Announce(fmt.Sprintf("All %d items selected", len(rows)), accessibility.PriorityPolite)
			}
			return runtime.Handled()
		}
	case terminal.KeyCtrlG:
		if t.draggable {
			return t.startDrag(rows)
		}
	}
	return runtime.Unhandled()
}

// handleShiftArrow extends the multi-selection range in the given direction.
func (t *Tree) handleShiftArrow(rows []treeRow, delta int) runtime.HandleResult {
	if len(rows) == 0 {
		return runtime.Handled()
	}
	// Establish anchor on first shift-select.
	if t.selAnchor < 0 {
		t.selAnchor = t.selectedIndex
	}
	newIdx := t.selectedIndex + delta
	if newIdx < 0 {
		newIdx = 0
	}
	if newIdx >= len(rows) {
		newIdx = len(rows) - 1
	}
	t.selectedIndex = newIdx
	// Clear and re-select the range from anchor to new index.
	t.selectedNodes = make(map[string]bool)
	t.selectRange(rows, t.selAnchor, newIdx)
	t.syncA11y()
	if announcer := t.services.Announcer(); announcer != nil {
		if row := t.selectedRow(rows); row != nil {
			announcer.Announce(row.node.Label, accessibility.PriorityPolite)
		}
	}
	return runtime.Handled()
}

// ---------- Drag key handlers ----------

// startDrag begins a drag operation on the currently selected node(s).
func (t *Tree) startDrag(rows []treeRow) runtime.HandleResult {
	if len(rows) == 0 || t.selectedIndex < 0 || t.selectedIndex >= len(rows) {
		return runtime.Unhandled()
	}
	t.dragging = true
	t.dragOrigin = t.selectedIndex
	t.dropIndex = t.selectedIndex

	row := rows[t.selectedIndex]
	label := row.node.Label
	// Build the source paths: selected nodes in multi-select, or just the focused node.
	var paths []string
	if t.multiSelect && len(t.selectedNodes) > 0 {
		paths = t.SelectedNodes()
	} else {
		paths = []string{row.path}
	}

	data := runtime.DragData{
		Source: t,
		Kind:   "tree-node",
		Value:  paths,
		Label:  label,
	}
	if announcer := t.services.Announcer(); announcer != nil {
		if len(paths) > 1 {
			announcer.Announce(fmt.Sprintf("Grabbed %d nodes", len(paths)), accessibility.PriorityAssertive)
		} else {
			announcer.Announce(fmt.Sprintf("Grabbed: %s", label), accessibility.PriorityAssertive)
		}
	}
	return runtime.WithCommand(runtime.DragStartMsg{Data: data})
}

// handleDragKey processes keys while a drag operation is in progress.
func (t *Tree) handleDragKey(key runtime.KeyMsg, rows []treeRow) runtime.HandleResult {
	count := len(rows)
	switch key.Key {
	case terminal.KeyUp:
		if t.dropIndex > 0 {
			t.dropIndex--
			t.announceDropPosition(rows)
		}
		return runtime.Handled()
	case terminal.KeyDown:
		if t.dropIndex < count-1 {
			t.dropIndex++
			t.announceDropPosition(rows)
		}
		return runtime.Handled()
	case terminal.KeyHome:
		t.dropIndex = 0
		t.announceDropPosition(rows)
		return runtime.Handled()
	case terminal.KeyEnd:
		if count > 0 {
			t.dropIndex = count - 1
		}
		t.announceDropPosition(rows)
		return runtime.Handled()
	case terminal.KeyEnter:
		return t.completeDrop(rows)
	case terminal.KeyEscape:
		return t.cancelDrag(rows)
	}
	return runtime.Handled() // Swallow other keys during drag.
}

// completeDrop finishes the drag by calling the onNodeDrop callback.
func (t *Tree) completeDrop(rows []treeRow) runtime.HandleResult {
	t.dragging = false
	origin := t.dragOrigin
	target := t.dropIndex
	t.dragOrigin = -1
	t.dropIndex = -1

	// Collect source paths.
	var sourcePaths []string
	if t.multiSelect && len(t.selectedNodes) > 0 {
		sourcePaths = t.SelectedNodes()
	} else if origin >= 0 && origin < len(rows) {
		sourcePaths = []string{rows[origin].path}
	}

	// Get target path.
	var targetPath string
	if target >= 0 && target < len(rows) {
		targetPath = rows[target].path
	}

	if t.onNodeDrop != nil && len(sourcePaths) > 0 && targetPath != "" {
		t.onNodeDrop(sourcePaths, targetPath)
	}

	t.selectedIndex = target
	t.syncA11y()

	data := runtime.DragData{
		Source: t,
		Kind:   "tree-node",
		Value:  sourcePaths,
		Label:  targetPath,
	}
	if announcer := t.services.Announcer(); announcer != nil {
		announcer.Announce(fmt.Sprintf("Dropped on %s", targetPath), accessibility.PriorityAssertive)
	}
	return runtime.WithCommand(runtime.DragEndMsg{Data: data, Target: t})
}

// cancelDrag aborts the current drag operation.
func (t *Tree) cancelDrag(rows []treeRow) runtime.HandleResult {
	origin := t.dragOrigin
	t.dragging = false
	t.dragOrigin = -1
	t.dropIndex = -1
	t.selectedIndex = origin
	t.syncA11y()

	var sourcePaths []string
	if origin >= 0 && origin < len(rows) {
		sourcePaths = []string{rows[origin].path}
	}

	data := runtime.DragData{
		Source: t,
		Kind:   "tree-node",
		Value:  sourcePaths,
	}
	if announcer := t.services.Announcer(); announcer != nil {
		announcer.Announce("Drag cancelled", accessibility.PriorityAssertive)
	}
	return runtime.WithCommand(runtime.DragEndMsg{Data: data, Cancelled: true})
}

// announceDropPosition announces the current drop target for screen readers.
func (t *Tree) announceDropPosition(rows []treeRow) {
	if t.dropIndex < 0 || t.dropIndex >= len(rows) {
		return
	}
	if announcer := t.services.Announcer(); announcer != nil {
		row := rows[t.dropIndex]
		announcer.Announce(fmt.Sprintf("Drop position: %s", row.node.Label), accessibility.PriorityPolite)
	}
}

type treeRow struct {
	node  *TreeNode
	depth int
	path  string // slash-separated path from root, e.g. "root/child/grandchild"
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
	var walk func(node *TreeNode, depth int, parentPath string)
	walk = func(node *TreeNode, depth int, parentPath string) {
		if node == nil {
			return
		}
		p := treeBuildPath(parentPath, node.Label)
		rows = append(rows, treeRow{node: node, depth: depth, path: p})
		if node.Expanded {
			for _, child := range node.Children {
				walk(child, depth+1, p)
			}
		}
	}
	walk(t.Root, 0, "")
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

// ---------- Multi-selection API ----------

// SetMultiSelect enables or disables multi-selection.
func (t *Tree) SetMultiSelect(enabled bool) {
	if t == nil {
		return
	}
	t.multiSelect = enabled
	if !enabled {
		t.selectedNodes = make(map[string]bool)
		t.selAnchor = -1
	}
	t.syncA11y()
}

// MultiSelect reports whether multi-selection is enabled.
func (t *Tree) MultiSelect() bool {
	if t == nil {
		return false
	}
	return t.multiSelect
}

// SelectedNodes returns the paths of all currently selected nodes, sorted by
// their position in the flattened tree (i.e. visual order).
func (t *Tree) SelectedNodes() []string {
	if t == nil || len(t.selectedNodes) == 0 {
		return nil
	}
	rows := t.flatten()
	result := make([]string, 0, len(t.selectedNodes))
	for _, row := range rows {
		if t.selectedNodes[row.path] {
			result = append(result, row.path)
		}
	}
	return result
}

// SelectNode selects the node at the given path.
func (t *Tree) SelectNode(path string) {
	if t == nil || !t.multiSelect {
		return
	}
	if t.selectedNodes == nil {
		t.selectedNodes = make(map[string]bool)
	}
	t.selectedNodes[path] = true
	t.syncA11y()
}

// DeselectNode deselects the node at the given path.
func (t *Tree) DeselectNode(path string) {
	if t == nil || !t.multiSelect {
		return
	}
	delete(t.selectedNodes, path)
	t.syncA11y()
}

// SelectAll selects all visible (flattened) nodes.
func (t *Tree) SelectAll() {
	if t == nil || !t.multiSelect {
		return
	}
	rows := t.flatten()
	if t.selectedNodes == nil {
		t.selectedNodes = make(map[string]bool, len(rows))
	}
	for _, row := range rows {
		t.selectedNodes[row.path] = true
	}
	t.syncA11y()
}

// DeselectAll clears all multi-selections.
func (t *Tree) DeselectAll() {
	if t == nil {
		return
	}
	t.selectedNodes = make(map[string]bool)
	t.selAnchor = -1
	t.syncA11y()
}

// isNodeSelected reports whether the node at the given flat index is selected
// (either by single cursor or multi-selection).
func (t *Tree) isNodeSelected(rows []treeRow, index int) bool {
	if index < 0 || index >= len(rows) {
		return false
	}
	if t.multiSelect && t.selectedNodes[rows[index].path] {
		return true
	}
	return index == t.selectedIndex
}

// toggleNodeSelection toggles the multi-selection state of the node at index.
func (t *Tree) toggleNodeSelection(rows []treeRow, index int) {
	if index < 0 || index >= len(rows) {
		return
	}
	path := rows[index].path
	if t.selectedNodes[path] {
		delete(t.selectedNodes, path)
	} else {
		if t.selectedNodes == nil {
			t.selectedNodes = make(map[string]bool)
		}
		t.selectedNodes[path] = true
	}
}

// selectRange selects all nodes between anchor and the given index (inclusive).
func (t *Tree) selectRange(rows []treeRow, fromIdx, toIdx int) {
	if fromIdx > toIdx {
		fromIdx, toIdx = toIdx, fromIdx
	}
	if t.selectedNodes == nil {
		t.selectedNodes = make(map[string]bool)
	}
	for i := fromIdx; i <= toIdx && i < len(rows); i++ {
		t.selectedNodes[rows[i].path] = true
	}
}

// ---------- Drag-and-drop API ----------

// SetDraggable enables or disables drag-and-drop for tree nodes.
func (t *Tree) SetDraggable(enabled bool) {
	if t == nil {
		return
	}
	t.draggable = enabled
}

// IsDraggable reports whether drag-and-drop is enabled.
func (t *Tree) IsDraggable() bool {
	if t == nil {
		return false
	}
	return t.draggable
}

// IsDragging reports whether a drag operation is in progress.
func (t *Tree) IsDragging() bool {
	if t == nil {
		return false
	}
	return t.dragging
}

// DragOrigin returns the flat index of the drag origin, or -1 if not dragging.
func (t *Tree) DragOrigin() int {
	if t == nil || !t.dragging {
		return -1
	}
	return t.dragOrigin
}

// DropIndex returns the current drop target flat index, or -1 if not dragging.
func (t *Tree) DropIndex() int {
	if t == nil || !t.dragging {
		return -1
	}
	return t.dropIndex
}

// SetDragStyle sets the style used for the drop indicator during drag.
func (t *Tree) SetDragStyle(style backend.Style) {
	if t == nil {
		return
	}
	t.dragStyle = style
}

// SetOnNodeDrop sets the callback invoked when nodes are dropped.
// sourcePaths are the paths of the dragged nodes; targetPath is the drop target.
func (t *Tree) SetOnNodeDrop(fn func(sourcePaths []string, targetPath string)) {
	if t == nil {
		return
	}
	t.onNodeDrop = fn
}

// CanDrop implements runtime.DropTarget. It accepts drops of kind "tree-node".
func (t *Tree) CanDrop(data runtime.DragData) bool {
	if t == nil || !t.draggable {
		return false
	}
	return data.Kind == "tree-node"
}

// OnDrop implements runtime.DropTarget. It accepts the drop.
func (t *Tree) OnDrop(data runtime.DragData) bool {
	if t == nil || !t.draggable {
		return false
	}
	return data.Kind == "tree-node"
}

func (t *Tree) syncA11y() {
	if t == nil {
		return
	}
	if t.Base.Role == "" {
		t.Base.Role = accessibility.RoleTree
	}
	t.Base.Orientation = "vertical"
	t.Base.State.Multiselectable = t.multiSelect
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
		// In multiSelect mode, reflect whether the focused node is selected.
		if t.multiSelect {
			t.Base.State.Selected = t.selectedNodes[row.path]
		} else {
			t.Base.State.Selected = true
		}
	} else {
		t.Base.Value = nil
		t.Base.State.Expanded = nil
		t.Base.Level = 0
		t.Base.State.Selected = false
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
var _ runtime.DropTarget = (*Tree)(nil)
var _ runtime.DragSource = (*Tree)(nil)
