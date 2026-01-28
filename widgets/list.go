package widgets

import (
	"fmt"
	"strings"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/scroll"
	"github.com/odvcencio/fluffyui/state"
	"github.com/odvcencio/fluffyui/terminal"
)

// RenderFunc renders an item.
type RenderFunc[T any] func(item T, index int, selected bool, ctx runtime.RenderContext)

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

// List renders a list of items.
type List[T any] struct {
	FocusableBase
	adapter       ListAdapter[T]
	selected      int
	offset        int
	onSelect      func(index int, item T)
	label         string
	style         backend.Style
	selectedStyle backend.Style
}

// NewList creates a list widget.
func NewList[T any](adapter ListAdapter[T]) *List[T] {
	list := &List[T]{
		adapter:       adapter,
		selected:      0,
		label:         "List",
		style:         backend.DefaultStyle(),
		selectedStyle: backend.DefaultStyle().Reverse(true),
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

// OnSelect registers a selection handler.
func (l *List[T]) OnSelect(fn func(index int, item T)) {
	if l == nil {
		return
	}
	l.onSelect = fn
}

// SetLabel updates the accessibility label.
func (l *List[T]) SetLabel(label string) {
	if l == nil {
		return
	}
	l.label = label
	l.syncA11y()
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
	for i := 0; i < content.Height; i++ {
		index := l.offset + i
		if index < 0 || index >= count {
			break
		}
		item := l.adapter.Item(index)
		rowBounds := runtime.Rect{X: content.X, Y: content.Y + i, Width: content.Width, Height: 1}
		rowCtx := ctx.Sub(rowBounds)
		if index == l.selected {
			ctx.Buffer.Fill(rowBounds, ' ', selectedStyle)
		}
		l.adapter.Render(item, index, index == l.selected, rowCtx)
	}
}

// HandleMessage handles navigation.
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
	}
	return runtime.Unhandled()
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
	l.selected = index
	l.syncA11y()
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
	label := strings.TrimSpace(l.label)
	if label == "" {
		label = "List"
	}
	l.Base.Label = label
	count := l.adapter.Count()
	l.Base.Description = fmt.Sprintf("%d items", count)
	if count > 0 && l.selected >= 0 && l.selected < count {
		item := l.adapter.Item(l.selected)
		l.Base.Value = &accessibility.ValueInfo{Text: fmt.Sprint(item)}
	} else {
		l.Base.Value = nil
	}
}

var _ scroll.Controller = (*List[any])(nil)
