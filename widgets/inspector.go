package widgets

import (
	"fmt"
	"strings"

	"m31labs.dev/fluffyui/accessibility"
	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/terminal"
)

// inspectorEntry holds metadata for one widget in the flattened tree.
type inspectorEntry struct {
	widget  runtime.Widget
	depth   int
	label   string
	role    string
	bounds  runtime.Rect
	focused bool
}

// Inspector displays a two-column overlay showing the widget tree and
// properties of the selected widget. It can inspect any widget tree
// provided via SetRoot.
type Inspector struct {
	FocusableBase
	root     runtime.Widget // root widget to inspect
	visible  bool
	selected int // index in flattened tree
	entries  []inspectorEntry
	scroll   int // scroll offset for tree view
	services runtime.Services
}

// InspectorOption configures an Inspector.
type InspectorOption = Option[Inspector]

// NewInspector creates a new Inspector widget. If root is non-nil the
// tree is built immediately.
func NewInspector(root runtime.Widget, opts ...InspectorOption) *Inspector {
	ins := &Inspector{
		root:    root,
		visible: true,
	}
	ins.Base.Role = accessibility.RoleGroup
	ins.Base.Label = "Widget Inspector"
	for _, opt := range opts {
		if opt != nil {
			opt(ins)
		}
	}
	if root != nil {
		ins.buildTree()
	}
	return ins
}

// SetRoot sets the root widget to inspect and rebuilds the tree.
func (ins *Inspector) SetRoot(w runtime.Widget) {
	if ins == nil {
		return
	}
	ins.root = w
	ins.buildTree()
}

// Toggle flips visibility.
func (ins *Inspector) Toggle() {
	if ins == nil {
		return
	}
	ins.visible = !ins.visible
}

// Visible reports whether the inspector is visible.
func (ins *Inspector) Visible() bool {
	if ins == nil {
		return false
	}
	return ins.visible
}

// Refresh rebuilds the tree from the current root.
func (ins *Inspector) Refresh() {
	if ins == nil {
		return
	}
	ins.buildTree()
}

// Selected returns the index of the currently selected entry.
func (ins *Inspector) Selected() int {
	if ins == nil {
		return 0
	}
	return ins.selected
}

// Entries returns the number of entries in the flattened tree.
func (ins *Inspector) Entries() int {
	if ins == nil {
		return 0
	}
	return len(ins.entries)
}

// SelectedEntry returns the inspectorEntry at the current selection, or a
// zero value if out of range.
func (ins *Inspector) SelectedEntry() inspectorEntry {
	if ins == nil || ins.selected < 0 || ins.selected >= len(ins.entries) {
		return inspectorEntry{}
	}
	return ins.entries[ins.selected]
}

// buildTree walks the root widget recursively and builds a flat list.
func (ins *Inspector) buildTree() {
	ins.entries = ins.entries[:0]
	if ins.root == nil {
		ins.selected = 0
		ins.scroll = 0
		return
	}
	visited := map[runtime.Widget]struct{}{}
	ins.walkTree(ins.root, 0, visited)
	if ins.selected >= len(ins.entries) {
		if len(ins.entries) > 0 {
			ins.selected = len(ins.entries) - 1
		} else {
			ins.selected = 0
		}
	}
}

func (ins *Inspector) walkTree(w runtime.Widget, depth int, visited map[runtime.Widget]struct{}) {
	if w == nil {
		return
	}
	if _, seen := visited[w]; seen {
		return
	}
	visited[w] = struct{}{}

	entry := inspectorEntry{
		widget: w,
		depth:  depth,
		label:  widgetTypeName(w),
	}

	if acc, ok := w.(accessibility.Accessible); ok && acc != nil {
		entry.role = string(acc.AccessibleRole())
		lbl := strings.TrimSpace(acc.AccessibleLabel())
		if lbl != "" {
			entry.label += " (" + lbl + ")"
		}
	}

	if bp, ok := w.(runtime.BoundsProvider); ok {
		entry.bounds = bp.Bounds()
	}

	if f, ok := w.(runtime.Focusable); ok {
		entry.focused = f.IsFocused()
	}

	ins.entries = append(ins.entries, entry)

	if cp, ok := w.(runtime.ChildProvider); ok {
		for _, child := range cp.ChildWidgets() {
			ins.walkTree(child, depth+1, visited)
		}
	}
}

// widgetTypeName returns a short type name like "Button" from "*widgets.Button".
func widgetTypeName(w runtime.Widget) string {
	name := fmt.Sprintf("%T", w)
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[idx+1:]
	}
	return strings.TrimPrefix(name, "*")
}

// Measure takes all available space.
func (ins *Inspector) Measure(constraints runtime.Constraints) runtime.Size {
	if ins == nil {
		return runtime.Size{}
	}
	return constraints.MaxSize()
}

// Layout stores bounds.
func (ins *Inspector) Layout(bounds runtime.Rect) {
	if ins == nil {
		return
	}
	ins.Base.Layout(bounds)
}

// Render draws the two-column inspector overlay.
func (ins *Inspector) Render(ctx runtime.RenderContext) {
	if ins == nil || !ins.visible {
		return
	}
	bounds := ins.Bounds()
	if bounds.Width < 4 || bounds.Height < 2 {
		return
	}

	bgStyle := backend.DefaultStyle().Dim(true)
	selectedStyle := backend.DefaultStyle().Reverse(true)
	headerStyle := backend.DefaultStyle().Reverse(true)
	propLabelStyle := backend.DefaultStyle().Dim(true)
	propValueStyle := backend.DefaultStyle()

	// Clear background
	ctx.Buffer.Fill(bounds, ' ', bgStyle)

	// Draw border box
	ctx.Buffer.DrawBox(bounds, bgStyle)

	// Title
	title := " Widget Inspector "
	if bounds.Width > len(title)+2 {
		ctx.Buffer.SetString(bounds.X+1, bounds.Y, title, headerStyle)
	}

	innerX := bounds.X + 1
	innerY := bounds.Y + 1
	innerW := bounds.Width - 2
	innerH := bounds.Height - 2
	if innerW <= 0 || innerH <= 0 {
		return
	}

	// Split: left ~40%, right ~60%
	leftW := innerW * 2 / 5
	if leftW < 10 {
		leftW = innerW
	}
	rightW := innerW - leftW
	rightX := innerX + leftW

	// Draw separator between columns
	if rightW > 0 && leftW < innerW {
		for row := 0; row < innerH; row++ {
			ctx.Buffer.Set(rightX, innerY+row, '|', bgStyle)
		}
		rightX++
		rightW--
	}

	// --- Left column: tree view ---
	ins.adjustScroll(innerH)
	visibleCount := len(ins.entries) - ins.scroll
	if visibleCount > innerH {
		visibleCount = innerH
	}

	for i := 0; i < visibleCount; i++ {
		idx := ins.scroll + i
		entry := ins.entries[idx]
		y := innerY + i

		indent := strings.Repeat("  ", entry.depth)
		prefix := indent
		if entry.focused {
			prefix += "*"
		}
		line := prefix + entry.label

		st := bgStyle
		if idx == ins.selected {
			st = selectedStyle
		}
		writePadded(ctx.Buffer, innerX, y, leftW, line, st)
	}

	// --- Right column: properties of selected widget ---
	if rightW <= 0 || ins.selected < 0 || ins.selected >= len(ins.entries) {
		return
	}
	entry := ins.entries[ins.selected]
	props := ins.buildProperties(entry)

	for i, prop := range props {
		if i >= innerH {
			break
		}
		y := innerY + i
		label := prop.key + ": "
		ctx.Buffer.SetString(rightX, y, clipString(label, rightW), propLabelStyle)
		valueX := rightX + textWidth(label)
		remaining := rightW - textWidth(label)
		if remaining > 0 {
			ctx.Buffer.SetString(valueX, y, clipString(prop.value, remaining), propValueStyle)
		}
	}
}

type propertyPair struct {
	key   string
	value string
}

// buildProperties gathers display properties from the selected entry.
func (ins *Inspector) buildProperties(entry inspectorEntry) []propertyPair {
	var props []propertyPair

	props = append(props, propertyPair{"Type", widgetTypeName(entry.widget)})

	if entry.role != "" {
		props = append(props, propertyPair{"Role", entry.role})
	}

	props = append(props, propertyPair{"Bounds", fmt.Sprintf("(%d,%d %dx%d)",
		entry.bounds.X, entry.bounds.Y, entry.bounds.Width, entry.bounds.Height)})

	props = append(props, propertyPair{"Focused", fmt.Sprintf("%v", entry.focused)})

	acc, ok := entry.widget.(accessibility.Accessible)
	if !ok || acc == nil {
		return props
	}

	label := strings.TrimSpace(acc.AccessibleLabel())
	if label != "" {
		props = append(props, propertyPair{"Label", label})
	}

	desc := strings.TrimSpace(acc.AccessibleDescription())
	if desc != "" {
		props = append(props, propertyPair{"Description", desc})
	}

	state := acc.AccessibleState()
	if parts := state.Strings(); len(parts) > 0 {
		props = append(props, propertyPair{"State", strings.Join(parts, ", ")})
	}

	if val := acc.AccessibleValue(); val != nil {
		text := strings.TrimSpace(val.Text)
		if text != "" {
			props = append(props, propertyPair{"Value", text})
		} else {
			props = append(props, propertyPair{"Value", fmt.Sprintf("%.2f (%.2f-%.2f)", val.Current, val.Min, val.Max)})
		}
	}

	if live := acc.AccessibleLive(); live != "" {
		props = append(props, propertyPair{"Live", string(live)})
	}

	if lm := acc.AccessibleLandmark(); lm != "" {
		props = append(props, propertyPair{"Landmark", string(lm)})
	}

	if lb := strings.TrimSpace(acc.AccessibleLabelledBy()); lb != "" {
		props = append(props, propertyPair{"LabelledBy", lb})
	}

	if db := strings.TrimSpace(acc.AccessibleDescribedBy()); db != "" {
		props = append(props, propertyPair{"DescribedBy", db})
	}

	if ctrl := strings.TrimSpace(acc.AccessibleControls()); ctrl != "" {
		props = append(props, propertyPair{"Controls", ctrl})
	}

	if level := acc.AccessibleLevel(); level > 0 {
		props = append(props, propertyPair{"Level", fmt.Sprintf("%d", level)})
	}

	if orient := acc.AccessibleOrientation(); orient != "" {
		props = append(props, propertyPair{"Orientation", orient})
	}

	if ks := acc.AccessibleKeyShortcuts(); ks != "" {
		props = append(props, propertyPair{"KeyShortcuts", ks})
	}

	return props
}

// adjustScroll ensures the selected entry is visible.
func (ins *Inspector) adjustScroll(visibleHeight int) {
	if visibleHeight <= 0 {
		return
	}
	if ins.selected < ins.scroll {
		ins.scroll = ins.selected
	}
	if ins.selected >= ins.scroll+visibleHeight {
		ins.scroll = ins.selected - visibleHeight + 1
	}
	if ins.scroll < 0 {
		ins.scroll = 0
	}
}

// HandleMessage processes keyboard input.
func (ins *Inspector) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if ins == nil || !ins.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}

	switch key.Key {
	case terminal.KeyEscape:
		ins.Toggle()
		return runtime.Handled()
	case terminal.KeyUp:
		if ins.selected > 0 {
			ins.selected--
			ins.announceSelected()
			return runtime.Handled()
		}
	case terminal.KeyDown:
		if ins.selected < len(ins.entries)-1 {
			ins.selected++
			ins.announceSelected()
			return runtime.Handled()
		}
	case terminal.KeyRune:
		if key.Rune == 'r' || key.Rune == 'R' {
			ins.Refresh()
			return runtime.Handled()
		}
	}
	return runtime.Unhandled()
}

// announceSelected announces the currently selected widget.
func (ins *Inspector) announceSelected() {
	if ins.selected < 0 || ins.selected >= len(ins.entries) {
		return
	}
	ann := ins.services.Announcer()
	if ann == nil {
		return
	}
	entry := ins.entries[ins.selected]
	msg := entry.label
	if entry.role != "" {
		msg += ", " + entry.role
	}
	ann.Announce(msg, accessibility.PriorityPolite)
}

// Bind attaches app services.
func (ins *Inspector) Bind(services runtime.Services) {
	if ins == nil {
		return
	}
	ins.services = services
}

// Unbind releases app services.
func (ins *Inspector) Unbind() {
	if ins == nil {
		return
	}
	ins.services = runtime.Services{}
}

var _ runtime.Widget = (*Inspector)(nil)
var _ runtime.Focusable = (*Inspector)(nil)
var _ runtime.Bindable = (*Inspector)(nil)
var _ runtime.Unbindable = (*Inspector)(nil)
