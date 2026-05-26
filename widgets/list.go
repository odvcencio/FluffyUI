package widgets

import (
	"fmt"
	"html/template"
	"strings"

	"m31labs.dev/fluffyui/accessibility"
	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/scroll"
	"m31labs.dev/fluffyui/state"
	"m31labs.dev/fluffyui/terminal"
)

// RenderFunc renders an item.
type RenderFunc[T any] func(item T, index int, selected bool, ctx runtime.RenderContext)

// HTMLItemRenderer renders a list item as HTML.
type HTMLItemRenderer[T any] func(item T, index int) runtime.HTML

// ListAdapter provides data for list widgets.
type ListAdapter[T any] interface {
	Count() int
	Item(index int) T
	Render(item T, index int, selected bool, ctx runtime.RenderContext)
}

// SliceAdapter adapts a slice to a ListAdapter.
type SliceAdapter[T any] struct {
	items  []T
	render RenderFunc[T]
}

// NewSliceAdapter creates a slice adapter.
func NewSliceAdapter[T any](items []T, render RenderFunc[T]) ListAdapter[T] {
	return &SliceAdapter[T]{items: items, render: render}
}

// Count returns the item count.
func (s *SliceAdapter[T]) Count() int {
	if s == nil {
		return 0
	}
	return len(s.items)
}

// Item returns the item at index.
func (s *SliceAdapter[T]) Item(index int) T {
	var zero T
	if s == nil || index < 0 || index >= len(s.items) {
		return zero
	}
	return s.items[index]
}

// Render renders the item.
func (s *SliceAdapter[T]) Render(item T, index int, selected bool, ctx runtime.RenderContext) {
	if s == nil || s.render == nil {
		return
	}
	s.render(item, index, selected, ctx)
}

// SignalAdapter adapts a signal slice to a ListAdapter.
type SignalAdapter[T any] struct {
	items  *state.Signal[[]T]
	render RenderFunc[T]
}

// NewSignalAdapter creates a signal adapter.
func NewSignalAdapter[T any](items *state.Signal[[]T], render RenderFunc[T]) ListAdapter[T] {
	return &SignalAdapter[T]{items: items, render: render}
}

// Count returns the item count.
func (s *SignalAdapter[T]) Count() int {
	if s == nil || s.items == nil {
		return 0
	}
	return len(s.items.Get())
}

// Item returns an item.
func (s *SignalAdapter[T]) Item(index int) T {
	var zero T
	if s == nil || s.items == nil {
		return zero
	}
	items := s.items.Get()
	if index < 0 || index >= len(items) {
		return zero
	}
	return items[index]
}

// Render draws an item.
func (s *SignalAdapter[T]) Render(item T, index int, selected bool, ctx runtime.RenderContext) {
	if s == nil || s.render == nil {
		return
	}
	s.render(item, index, selected, ctx)
}

// MutableListAdapter extends ListAdapter with mutation support for drag-and-drop reordering.
type MutableListAdapter[T any] interface {
	ListAdapter[T]
	// Move moves the item at fromIndex to toIndex, shifting other items as needed.
	Move(fromIndex, toIndex int)
}

// MutableSliceAdapter adapts a mutable slice to a MutableListAdapter.
type MutableSliceAdapter[T any] struct {
	SliceAdapter[T]
}

// NewMutableSliceAdapter creates a mutable slice adapter.
func NewMutableSliceAdapter[T any](items []T, render RenderFunc[T]) *MutableSliceAdapter[T] {
	return &MutableSliceAdapter[T]{
		SliceAdapter: SliceAdapter[T]{items: items, render: render},
	}
}

// Move moves an item from one index to another.
func (m *MutableSliceAdapter[T]) Move(fromIndex, toIndex int) {
	if m == nil || fromIndex < 0 || toIndex < 0 {
		return
	}
	if fromIndex >= len(m.items) || toIndex >= len(m.items) {
		return
	}
	if fromIndex == toIndex {
		return
	}
	item := m.items[fromIndex]
	// Remove from old position.
	m.items = append(m.items[:fromIndex], m.items[fromIndex+1:]...)
	// Insert at new position.
	m.items = append(m.items[:toIndex], append([]T{item}, m.items[toIndex:]...)...)
}

// Items returns the current slice (useful for reading back after reorder).
func (m *MutableSliceAdapter[T]) Items() []T {
	if m == nil {
		return nil
	}
	return m.items
}

// ListOnDrop is called when a drop completes on a list.
// It receives the source index, destination index, and the item that was moved.
type ListOnDrop[T any] func(fromIndex, toIndex int, item T)

// List renders a list of items.
type List[T any] struct {
	FocusableBase
	adapter         ListAdapter[T]
	selected        int
	offset          int
	onSelect        func(index int, item T)
	onDrop          ListOnDrop[T]
	label           string
	labelTrimmed    string
	descriptionCache string
	cachedCount     int
	style           backend.Style
	selectedStyle   backend.Style
	dragStyle       backend.Style
	services         runtime.Services
	htmlItemRenderer HTMLItemRenderer[T]

	// Drag-and-drop state
	draggable  bool
	dragging   bool
	dragOrigin int // original index of the dragged item
	dropIndex  int // current drop target index
}

// NewList creates a list widget.
func NewList[T any](adapter ListAdapter[T]) *List[T] {
	list := &List[T]{
		adapter:       adapter,
		selected:      0,
		label:         "List",
		labelTrimmed:  "List",
		style:         backend.DefaultStyle(),
		selectedStyle: backend.DefaultStyle().Reverse(true),
		dragStyle:     backend.DefaultStyle().Reverse(true).Bold(true),
		cachedCount:   -1,
		dragOrigin:    -1,
		dropIndex:     -1,
	}
	list.Base.Role = accessibility.RoleList
	list.syncA11y()
	return list
}

// SetStyle updates the list base style.
func (l *List[T]) SetStyle(style backend.Style) {
	if l == nil {
		return
	}
	l.style = style
}

// SetSelectedStyle updates the selected row style.
func (l *List[T]) SetSelectedStyle(style backend.Style) {
	if l == nil {
		return
	}
	l.selectedStyle = style
}

// StyleType returns the selector type name.
func (l *List[T]) StyleType() string {
	return "List"
}

// SetOnSelect registers a selection handler.
func (l *List[T]) SetOnSelect(fn func(index int, item T)) {
	if l == nil {
		return
	}
	l.onSelect = fn
}

// OnSelect registers a selection handler on the list.
//
// Deprecated: Use SetOnSelect instead. This method will be removed in v1.0.
func (l *List[T]) OnSelect(fn func(index int, item T)) {
	l.SetOnSelect(fn)
}

// SetLabel updates the accessibility label.
func (l *List[T]) SetLabel(label string) {
	if l == nil {
		return
	}
	l.label = label
	l.labelTrimmed = strings.TrimSpace(label)
	l.syncA11y()
}

// SetDraggable enables or disables drag-and-drop reordering.
// When enabled, the user can press Ctrl+G to grab a selected item,
// use arrow keys to move the drop indicator, Enter to drop (reorder),
// or Escape to cancel.
func (l *List[T]) SetDraggable(enabled bool) {
	if l == nil {
		return
	}
	l.draggable = enabled
}

// IsDraggable reports whether drag-and-drop reordering is enabled.
func (l *List[T]) IsDraggable() bool {
	if l == nil {
		return false
	}
	return l.draggable
}

// IsDragging reports whether a drag operation is currently in progress.
func (l *List[T]) IsDragging() bool {
	if l == nil {
		return false
	}
	return l.dragging
}

// DragOrigin returns the index of the item being dragged, or -1 if no drag is active.
func (l *List[T]) DragOrigin() int {
	if l == nil || !l.dragging {
		return -1
	}
	return l.dragOrigin
}

// DropIndex returns the current drop target index, or -1 if no drag is active.
func (l *List[T]) DropIndex() int {
	if l == nil || !l.dragging {
		return -1
	}
	return l.dropIndex
}

// SetOnDrop registers a handler called when a drop completes.
func (l *List[T]) SetOnDrop(fn ListOnDrop[T]) {
	if l == nil {
		return
	}
	l.onDrop = fn
}

// SetDragStyle updates the style used for the item being dragged.
func (l *List[T]) SetDragStyle(style backend.Style) {
	if l == nil {
		return
	}
	l.dragStyle = style
}

// SetHTMLItemRenderer sets a custom HTML renderer for list items.
func (l *List[T]) SetHTMLItemRenderer(r HTMLItemRenderer[T]) {
	l.htmlItemRenderer = r
}

// RenderHTML renders the list as a static HTML unordered list.
func (l *List[T]) RenderHTML(ctx runtime.HTMLContext) runtime.HTML {
	var b strings.Builder
	b.WriteString(`<ul class="fluffy-List">`)
	for i := 0; i < l.adapter.Count(); i++ {
		item := l.adapter.Item(i)
		if l.htmlItemRenderer != nil {
			b.WriteString(string(l.htmlItemRenderer(item, i)))
		} else {
			b.WriteString(fmt.Sprintf("<li>%s</li>",
				template.HTMLEscapeString(fmt.Sprint(item))))
		}
	}
	b.WriteString("</ul>")
	return runtime.HTML(b.String())
}

// Measure returns the desired size.
func (l *List[T]) Measure(constraints runtime.Constraints) runtime.Size {
	return l.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		count := 0
		if l != nil && l.adapter != nil {
			count = l.adapter.Count()
		}
		height := min(count, contentConstraints.MaxHeight)
		if height <= 0 {
			height = contentConstraints.MinHeight
		}
		return contentConstraints.Constrain(runtime.Size{Width: contentConstraints.MaxWidth, Height: height})
	})
}

// Render draws list items.
func (l *List[T]) Render(ctx runtime.RenderContext) {
	if l == nil || l.adapter == nil {
		return
	}
	l.syncA11y()
	outer := l.bounds
	content := l.ContentBounds()
	if outer.Width <= 0 || outer.Height <= 0 {
		return
	}
	baseStyle := mergeBackendStyles(resolveBaseStyle(ctx, l, backend.DefaultStyle(), false), l.style)
	selectedStyle := mergeBackendStyles(baseStyle, l.selectedStyle)
	ctx.Buffer.Fill(outer, ' ', baseStyle)
	if content.Width <= 0 || content.Height <= 0 {
		return
	}
	count := l.adapter.Count()
	if l.selected < 0 {
		l.selected = 0
	}
	if l.selected >= count {
		l.selected = count - 1
	}
	if l.selected < l.offset {
		l.offset = l.selected
	}
	if l.selected >= l.offset+content.Height {
		l.offset = l.selected - content.Height + 1
	}
	dragIndicatorStyle := mergeBackendStyles(baseStyle, l.dragStyle)
	for i := 0; i < content.Height; i++ {
		index := l.offset + i
		if index < 0 || index >= count {
			break
		}
		item := l.adapter.Item(index)
		rowBounds := runtime.Rect{X: content.X, Y: content.Y + i, Width: content.Width, Height: 1}
		rowCtx := ctx.Sub(rowBounds)
		if l.dragging && index == l.dragOrigin {
			// The item being dragged gets the drag style.
			ctx.Buffer.Fill(rowBounds, ' ', dragIndicatorStyle)
			l.adapter.Render(item, index, true, rowCtx)
		} else if l.dragging && index == l.dropIndex {
			// The drop target position gets highlighted.
			ctx.Buffer.Fill(rowBounds, ' ', selectedStyle)
			l.adapter.Render(item, index, true, rowCtx)
		} else if index == l.selected && !l.dragging {
			ctx.Buffer.Fill(rowBounds, ' ', selectedStyle)
			l.adapter.Render(item, index, true, rowCtx)
		} else {
			l.adapter.Render(item, index, false, rowCtx)
		}
	}
}

// HandleMessage handles navigation and drag-and-drop.
func (l *List[T]) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if l == nil || !l.focused || l.adapter == nil {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	count := l.adapter.Count()
	if count == 0 {
		return runtime.Unhandled()
	}

	// Handle drag-in-progress keys first.
	if l.dragging {
		return l.handleDragKey(key, count)
	}

	switch key.Key {
	case terminal.KeyUp:
		l.setSelected(l.selected - 1)
		return runtime.Handled()
	case terminal.KeyDown:
		l.setSelected(l.selected + 1)
		return runtime.Handled()
	case terminal.KeyPageUp:
		l.setSelected(l.selected - l.bounds.Height)
		return runtime.Handled()
	case terminal.KeyPageDown:
		l.setSelected(l.selected + l.bounds.Height)
		return runtime.Handled()
	case terminal.KeyHome:
		l.setSelected(0)
		return runtime.Handled()
	case terminal.KeyEnd:
		l.setSelected(count - 1)
		return runtime.Handled()
	case terminal.KeyEnter:
		item := l.adapter.Item(l.selected)
		if l.onSelect != nil {
			l.onSelect(l.selected, item)
		}
		return runtime.Handled()
	case terminal.KeyCtrlG:
		if l.draggable {
			return l.startDrag()
		}
	}
	return runtime.Unhandled()
}

// startDrag begins a drag operation on the currently selected item.
func (l *List[T]) startDrag() runtime.HandleResult {
	count := l.adapter.Count()
	if count == 0 || l.selected < 0 || l.selected >= count {
		return runtime.Unhandled()
	}
	l.dragging = true
	l.dragOrigin = l.selected
	l.dropIndex = l.selected
	item := l.adapter.Item(l.selected)
	label := fmt.Sprint(item)
	data := runtime.DragData{
		Source: l,
		Kind:   "list-item",
		Value:  item,
		Label:  label,
	}
	if announcer := l.services.Announcer(); announcer != nil {
		announcer.Announce(fmt.Sprintf("Grabbed item %d: %s", l.selected+1, label), accessibility.PriorityAssertive)
	}
	return runtime.WithCommand(runtime.DragStartMsg{Data: data})
}

// handleDragKey processes keys while a drag operation is in progress.
func (l *List[T]) handleDragKey(key runtime.KeyMsg, count int) runtime.HandleResult {
	switch key.Key {
	case terminal.KeyUp:
		if l.dropIndex > 0 {
			l.dropIndex--
			l.announceDropPosition()
		}
		return runtime.Handled()
	case terminal.KeyDown:
		if l.dropIndex < count-1 {
			l.dropIndex++
			l.announceDropPosition()
		}
		return runtime.Handled()
	case terminal.KeyHome:
		l.dropIndex = 0
		l.announceDropPosition()
		return runtime.Handled()
	case terminal.KeyEnd:
		l.dropIndex = count - 1
		l.announceDropPosition()
		return runtime.Handled()
	case terminal.KeyEnter, terminal.KeyCtrlD:
		return l.completeDrop()
	case terminal.KeyEscape:
		return l.cancelDrag()
	}
	return runtime.Handled() // Swallow other keys during drag.
}

// completeDrop finishes the drag by reordering the list.
func (l *List[T]) completeDrop() runtime.HandleResult {
	from := l.dragOrigin
	to := l.dropIndex
	l.dragging = false
	dragOrigin := l.dragOrigin
	l.dragOrigin = -1
	l.dropIndex = -1

	item := l.adapter.Item(from)

	// Perform the reorder if the adapter supports it.
	if mutable, ok := l.adapter.(MutableListAdapter[T]); ok && from != to {
		mutable.Move(from, to)
	}

	l.selected = to
	l.syncA11y()

	if l.onDrop != nil {
		l.onDrop(dragOrigin, to, item)
	}

	data := runtime.DragData{
		Source: l,
		Kind:   "list-item",
		Value:  item,
		Label:  fmt.Sprint(item),
	}
	return runtime.WithCommand(runtime.DragEndMsg{
		Data:      data,
		Target:    l,
		Cancelled: false,
	})
}

// cancelDrag cancels the current drag operation.
func (l *List[T]) cancelDrag() runtime.HandleResult {
	item := l.adapter.Item(l.dragOrigin)
	l.selected = l.dragOrigin
	l.dragging = false
	l.dragOrigin = -1
	l.dropIndex = -1
	l.syncA11y()

	data := runtime.DragData{
		Source: l,
		Kind:   "list-item",
		Value:  item,
		Label:  fmt.Sprint(item),
	}
	return runtime.WithCommand(runtime.DragEndMsg{
		Data:      data,
		Target:    nil,
		Cancelled: true,
	})
}

func (l *List[T]) announceDropPosition() {
	if announcer := l.services.Announcer(); announcer != nil {
		item := l.adapter.Item(l.dropIndex)
		announcer.Announce(fmt.Sprintf("Drop position %d: %v", l.dropIndex+1, item), accessibility.PriorityPolite)
	}
}

// CanDrop returns true if this list accepts the given drag data.
func (l *List[T]) CanDrop(data runtime.DragData) bool {
	if l == nil || !l.draggable {
		return false
	}
	return data.Kind == "list-item"
}

// OnDrop handles an external drop onto this list.
func (l *List[T]) OnDrop(data runtime.DragData) bool {
	if l == nil || !l.draggable {
		return false
	}
	return data.Kind == "list-item"
}

func (l *List[T]) setSelected(index int) {
	if l == nil || l.adapter == nil {
		return
	}
	count := l.adapter.Count()
	if count == 0 {
		l.selected = 0
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= count {
		index = count - 1
	}
	prev := l.selected
	l.selected = index
	l.syncA11y()
	if prev != index {
		if announcer := l.services.Announcer(); announcer != nil {
			item := l.adapter.Item(l.selected)
			announcer.Announce(fmt.Sprintf("item %d, %v", index+1, item), accessibility.PriorityPolite)
		}
	}
	if l.onSelect != nil {
		l.onSelect(l.selected, l.adapter.Item(l.selected))
	}
}

// SetSelected updates the selected index.
func (l *List[T]) SetSelected(index int) {
	if l == nil {
		return
	}
	l.setSelected(index)
	l.Invalidate()
}

// SelectedIndex returns the current selection index.
func (l *List[T]) SelectedIndex() int {
	if l == nil {
		return 0
	}
	return l.selected
}

// SelectedItem returns the selected item.
func (l *List[T]) SelectedItem() (T, bool) {
	var zero T
	if l == nil || l.adapter == nil {
		return zero, false
	}
	if l.selected < 0 || l.selected >= l.adapter.Count() {
		return zero, false
	}
	return l.adapter.Item(l.selected), true
}

// ScrollBy scrolls selection by delta.
func (l *List[T]) ScrollBy(dx, dy int) {
	if l == nil || l.adapter == nil {
		return
	}
	if dy == 0 {
		return
	}
	l.setSelected(l.selected + dy)
	l.Invalidate()
}

// ScrollTo scrolls to an absolute index.
func (l *List[T]) ScrollTo(x, y int) {
	if l == nil || l.adapter == nil {
		return
	}
	l.setSelected(y)
	l.Invalidate()
}

// PageBy scrolls by a number of pages.
func (l *List[T]) PageBy(pages int) {
	if l == nil || l.adapter == nil {
		return
	}
	pageSize := l.bounds.Height
	if pageSize < 1 {
		pageSize = 1
	}
	l.setSelected(l.selected + pages*pageSize)
	l.Invalidate()
}

// ScrollToStart scrolls to the first item.
func (l *List[T]) ScrollToStart() {
	if l == nil || l.adapter == nil {
		return
	}
	l.setSelected(0)
	l.Invalidate()
}

// ScrollToEnd scrolls to the last item.
func (l *List[T]) ScrollToEnd() {
	if l == nil || l.adapter == nil {
		return
	}
	l.setSelected(l.adapter.Count() - 1)
	l.Invalidate()
}

func (l *List[T]) syncA11y() {
	if l == nil || l.adapter == nil {
		return
	}
	if l.Base.Role == "" {
		l.Base.Role = accessibility.RoleList
	}
	l.Base.Orientation = "vertical"
	label := l.labelTrimmed
	if label == "" {
		label = "List"
	}
	l.Base.Label = label
	count := l.adapter.Count()
	if count != l.cachedCount {
		l.descriptionCache = fmt.Sprintf("%d items", count)
		l.cachedCount = count
	}
	l.Base.Description = l.descriptionCache
	if count > 0 && l.selected >= 0 && l.selected < count {
		item := l.adapter.Item(l.selected)
		l.Base.Value = &accessibility.ValueInfo{Text: fmt.Sprint(item)}
	} else {
		l.Base.Value = nil
	}
}

// Bind attaches app services.
func (l *List[T]) Bind(services runtime.Services) {
	if l == nil {
		return
	}
	l.services = services
}

// Unbind releases app services.
func (l *List[T]) Unbind() {
	if l == nil {
		return
	}
	l.services = runtime.Services{}
}

var _ scroll.Controller = (*List[any])(nil)

var _ runtime.Widget = (*List[any])(nil)
var _ runtime.Focusable = (*List[any])(nil)
var _ runtime.Bindable = (*List[any])(nil)
var _ runtime.Unbindable = (*List[any])(nil)
var _ runtime.DragSource = (*List[any])(nil)
var _ runtime.DropTarget = (*List[any])(nil)
