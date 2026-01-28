package widgets

import (
	"fmt"
	"strings"

	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/scroll"
	"github.com/odvcencio/fluffy-ui/terminal"
)

// VirtualListAdapter provides data for a virtual list.
type VirtualListAdapter[T any] interface {
	Count() int
	Item(index int) T
	Render(item T, index int, selected bool, ctx runtime.RenderContext)
}

// VirtualListHeightProvider optionally provides variable item heights.
type VirtualListHeightProvider interface {
	ItemHeight(index int) int
}

// VirtualListFixedHeightProvider optionally provides a fixed item height.
type VirtualListFixedHeightProvider interface {
	FixedItemHeight() int
}

// VirtualListWidgetFactory provides pooled widget rendering.
type VirtualListWidgetFactory[T any] interface {
	NewWidget() runtime.Widget
	UpdateWidget(widget runtime.Widget, item T, index int, selected bool)
}

// VirtualListWidgetResetter resets widgets before reusing them.
type VirtualListWidgetResetter interface {
	ResetWidget(widget runtime.Widget)
}

// VirtualList renders large datasets efficiently using virtualization.
type VirtualList[T any] struct {
	FocusableBase
	adapter       VirtualListAdapter[T]
	list          *scroll.VirtualList
	label         string
	style         backend.Style
	selectedStyle backend.Style
	behavior      scroll.ScrollBehavior
	manualHeight  bool
	onSelect      func(index int, item T)
	services      runtime.Services
	widgetFactory VirtualListWidgetFactory[T]
	itemPool      *runtime.WidgetPool[runtime.Widget]
	itemCache     map[int]runtime.Widget
	poolMax       int
}

// NewVirtualList creates a virtual list widget.
func NewVirtualList[T any](adapter VirtualListAdapter[T]) *VirtualList[T] {
	v := &VirtualList[T]{
		adapter:       adapter,
		label:         "Virtual List",
		style:         backend.DefaultStyle(),
		selectedStyle: backend.DefaultStyle().Reverse(true),
		behavior:      scroll.ScrollBehavior{Vertical: scroll.ScrollAuto, Horizontal: scroll.ScrollNever, MouseWheel: 3, PageSize: 1},
	}
	v.list = scroll.NewVirtualList(0, 1, nil)
	v.list.SetOverscan(2)
	v.list.SetRenderItem(func(index int, selected bool, ctx runtime.RenderContext) {
		v.renderIndex(index, selected, ctx)
	})
	v.list.SetItemAt(func(index int) any {
		if v.adapter == nil {
			return nil
		}
		return v.adapter.Item(index)
	})
	v.list.SetOnSelection(func(index int) {
		v.handleSelection(index)
	})
	v.Base.Role = accessibility.RoleList
	v.syncA11y()
	v.configureWidgetFactory()
	v.applyHeightProvider()
	return v
}

// SetAdapter replaces the data adapter.
func (v *VirtualList[T]) SetAdapter(adapter VirtualListAdapter[T]) {
	if v == nil {
		return
	}
	v.adapter = adapter
	v.configureWidgetFactory()
	v.applyHeightProvider()
	v.syncA11y()
	v.Invalidate()
}

// SetStyle updates the list base style.
func (v *VirtualList[T]) SetStyle(style backend.Style) {
	if v == nil {
		return
	}
	v.style = style
}

// SetSelectedStyle updates the selected row style.
func (v *VirtualList[T]) SetSelectedStyle(style backend.Style) {
	if v == nil {
		return
	}
	v.selectedStyle = style
}

// SetLabel updates the accessibility label.
func (v *VirtualList[T]) SetLabel(label string) {
	if v == nil {
		return
	}
	v.label = label
	v.syncA11y()
}

// SetOverscan updates the number of extra items to render.
func (v *VirtualList[T]) SetOverscan(count int) {
	if v == nil || v.list == nil {
		return
	}
	v.list.SetOverscan(count)
}

// SetBehavior updates scroll behavior.
func (v *VirtualList[T]) SetBehavior(behavior scroll.ScrollBehavior) {
	if v == nil {
		return
	}
	v.behavior = behavior
}

// SetWidgetPoolMax limits the pooled widget count when using widget factories.
func (v *VirtualList[T]) SetWidgetPoolMax(max int) {
	if v == nil {
		return
	}
	v.poolMax = max
	v.configureWidgetFactory()
}

// SetItemHeight sets a fixed item height for faster indexing.
func (v *VirtualList[T]) SetItemHeight(height int) {
	if v == nil || v.list == nil {
		return
	}
	v.manualHeight = true
	v.list.SetItemHeight(height)
}

// SetItemHeightFunc sets a variable height function.
func (v *VirtualList[T]) SetItemHeightFunc(fn func(index int) int) {
	if v == nil || v.list == nil {
		return
	}
	v.manualHeight = true
	v.list.SetItemHeightFunc(fn)
}

// UseAdapterHeights reverts to adapter-provided heights when available.
func (v *VirtualList[T]) UseAdapterHeights() {
	if v == nil {
		return
	}
	v.manualHeight = false
	v.applyHeightProvider()
}

// OnSelect registers a selection handler.
func (v *VirtualList[T]) OnSelect(fn func(index int, item T)) {
	if v == nil {
		return
	}
	v.onSelect = fn
}

// SelectedIndex returns the current selection.
func (v *VirtualList[T]) SelectedIndex() int {
	if v == nil || v.list == nil {
		return 0
	}
	return v.list.SelectedIndex()
}

// SetSelected updates the selected index.
func (v *VirtualList[T]) SetSelected(index int) {
	if v == nil || v.list == nil {
		return
	}
	v.refreshMetrics(v.ContentBounds())
	v.list.SetSelected(index)
	v.list.EnsureVisible(v.list.SelectedIndex())
	v.handleSelection(v.list.SelectedIndex())
	v.Invalidate()
}

// SelectedItem returns the selected item.
func (v *VirtualList[T]) SelectedItem() (T, bool) {
	var zero T
	if v == nil || v.list == nil || v.adapter == nil {
		return zero, false
	}
	index := v.list.SelectedIndex()
	if index < 0 || index >= v.adapter.Count() {
		return zero, false
	}
	return v.adapter.Item(index), true
}

// ScrollToIndex scrolls to the specified index.
func (v *VirtualList[T]) ScrollToIndex(index int) {
	if v == nil || v.list == nil {
		return
	}
	v.refreshMetrics(v.ContentBounds())
	v.list.ScrollToIndex(index)
	v.Invalidate()
}

// ScrollToOffset scrolls to the specified offset in rows/pixels.
func (v *VirtualList[T]) ScrollToOffset(offset int) {
	if v == nil || v.list == nil {
		return
	}
	v.refreshMetrics(v.ContentBounds())
	v.list.ScrollToOffset(offset)
	v.Invalidate()
}

// Offset returns the current scroll offset.
func (v *VirtualList[T]) Offset() int {
	if v == nil || v.list == nil {
		return 0
	}
	return v.list.Offset()
}

// Measure returns the desired size.
func (v *VirtualList[T]) Measure(constraints runtime.Constraints) runtime.Size {
	return v.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		if v == nil {
			return contentConstraints.MinSize()
		}
		width := contentConstraints.MaxWidth
		if width <= 0 {
			width = contentConstraints.MinWidth
		}
		height := contentConstraints.MaxHeight
		if height <= 0 {
			height = contentConstraints.MinHeight
		}
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: height})
	})
}

// Layout stores bounds and updates viewport height.
func (v *VirtualList[T]) Layout(bounds runtime.Rect) {
	v.FocusableBase.Layout(bounds)
	if v == nil || v.list == nil {
		return
	}
	content := v.ContentBounds()
	v.list.SetViewportHeight(content.Height)
}

// Bind attaches app services.
func (v *VirtualList[T]) Bind(services runtime.Services) {
	if v == nil {
		return
	}
	v.services = services
	for _, widget := range v.itemCache {
		runtime.BindTree(widget, services)
	}
}

// Unbind releases app services.
func (v *VirtualList[T]) Unbind() {
	if v == nil {
		return
	}
	for _, widget := range v.itemCache {
		runtime.UnbindTree(widget)
	}
	v.services = runtime.Services{}
}

// Render draws the list items.
func (v *VirtualList[T]) Render(ctx runtime.RenderContext) {
	if v == nil || v.list == nil {
		return
	}
	v.syncA11y()
	outer := v.bounds
	content := v.ContentBounds()
	if outer.Width <= 0 || outer.Height <= 0 {
		return
	}
	baseStyle := mergeBackendStyles(resolveBaseStyle(ctx, v, backend.DefaultStyle(), false), v.style)
	ctx.Buffer.Fill(outer, ' ', baseStyle)
	if content.Width <= 0 || content.Height <= 0 {
		return
	}
	v.refreshMetrics(content)
	if v.widgetFactory != nil {
		v.renderWidgetItems(ctx, content)
		return
	}
	v.renderAdapterItems(ctx, content)
}

// HandleMessage handles navigation input.
func (v *VirtualList[T]) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if v == nil || v.list == nil {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if ok {
		if !v.focused {
			return runtime.Unhandled()
		}
		switch key.Key {
		case terminal.KeyUp:
			v.ScrollBy(0, -1)
			return runtime.Handled()
		case terminal.KeyDown:
			v.ScrollBy(0, 1)
			return runtime.Handled()
		case terminal.KeyPageUp:
			v.PageBy(-1)
			return runtime.Handled()
		case terminal.KeyPageDown:
			v.PageBy(1)
			return runtime.Handled()
		case terminal.KeyHome:
			v.ScrollToStart()
			return runtime.Handled()
		case terminal.KeyEnd:
			v.ScrollToEnd()
			return runtime.Handled()
		case terminal.KeyEnter:
			v.handleSelection(v.list.SelectedIndex())
			return runtime.Handled()
		}
	}
	if mouse, ok := msg.(runtime.MouseMsg); ok {
		if mouse.Button == runtime.MouseWheelUp {
			v.ScrollBy(0, -v.behavior.MouseWheel)
			return runtime.Handled()
		}
		if mouse.Button == runtime.MouseWheelDown {
			v.ScrollBy(0, v.behavior.MouseWheel)
			return runtime.Handled()
		}
	}
	return runtime.Unhandled()
}

// ScrollBy scrolls selection by delta rows.
func (v *VirtualList[T]) ScrollBy(dx, dy int) {
	if v == nil || v.list == nil || dy == 0 {
		return
	}
	v.refreshMetrics(v.ContentBounds())
	v.list.SetSelected(v.list.SelectedIndex() + dy)
	v.list.EnsureVisible(v.list.SelectedIndex())
	v.handleSelection(v.list.SelectedIndex())
	v.Invalidate()
}

// ScrollTo scrolls to an absolute index.
func (v *VirtualList[T]) ScrollTo(x, y int) {
	if v == nil || v.list == nil {
		return
	}
	v.refreshMetrics(v.ContentBounds())
	v.list.SetSelected(y)
	v.list.EnsureVisible(v.list.SelectedIndex())
	v.handleSelection(v.list.SelectedIndex())
	v.Invalidate()
}

// PageBy scrolls selection by a number of pages.
func (v *VirtualList[T]) PageBy(pages int) {
	if v == nil || v.list == nil {
		return
	}
	v.refreshMetrics(v.ContentBounds())
	pageSize := v.pageSize()
	if pageSize < 1 {
		pageSize = 1
	}
	v.list.SetSelected(v.list.SelectedIndex() + pages*pageSize)
	v.list.EnsureVisible(v.list.SelectedIndex())
	v.handleSelection(v.list.SelectedIndex())
	v.Invalidate()
}

// ScrollToStart selects the first item.
func (v *VirtualList[T]) ScrollToStart() {
	if v == nil || v.list == nil {
		return
	}
	v.refreshMetrics(v.ContentBounds())
	v.list.SetSelected(0)
	v.list.EnsureVisible(v.list.SelectedIndex())
	v.handleSelection(v.list.SelectedIndex())
	v.Invalidate()
}

// ScrollToEnd selects the last item.
func (v *VirtualList[T]) ScrollToEnd() {
	if v == nil || v.list == nil {
		return
	}
	v.refreshMetrics(v.ContentBounds())
	count := 0
	if v.adapter != nil {
		count = v.adapter.Count()
	}
	if count <= 0 {
		v.list.SetSelected(0)
	} else {
		v.list.SetSelected(count - 1)
	}
	v.list.EnsureVisible(v.list.SelectedIndex())
	v.handleSelection(v.list.SelectedIndex())
	v.Invalidate()
}

func (v *VirtualList[T]) pageSize() int {
	if v == nil {
		return 1
	}
	bounds := v.ContentBounds()
	if v.behavior.PageSize > 0 {
		return int(float64(bounds.Height) * v.behavior.PageSize)
	}
	if bounds.Height > 0 {
		return bounds.Height
	}
	return 1
}

func (v *VirtualList[T]) refreshMetrics(content runtime.Rect) {
	if v == nil || v.list == nil {
		return
	}
	if v.adapter != nil {
		v.list.SetItemCount(v.adapter.Count())
	} else {
		v.list.SetItemCount(0)
	}
	v.list.SetViewportHeight(content.Height)
}

func (v *VirtualList[T]) renderAdapterItems(ctx runtime.RenderContext, content runtime.Rect) {
	offset := v.list.Offset()
	start, end := v.list.GetVisibleRange()
	if end <= start {
		return
	}
	for i := start; i < end; i++ {
		itemHeight := v.list.ItemHeight(i)
		if itemHeight <= 0 {
			continue
		}
		itemY := content.Y + (v.list.OffsetForIndex(i) - offset)
		itemBounds := runtime.Rect{X: content.X, Y: itemY, Width: content.Width, Height: itemHeight}
		if itemBounds.Y >= content.Y+content.Height {
			break
		}
		if itemBounds.Y+itemBounds.Height <= content.Y {
			continue
		}
		v.list.RenderItem(i, ctx.Sub(itemBounds))
	}
}

func (v *VirtualList[T]) renderWidgetItems(ctx runtime.RenderContext, content runtime.Rect) {
	if v.adapter == nil || v.widgetFactory == nil {
		return
	}
	offset := v.list.Offset()
	start, end := v.list.GetVisibleRange()
	if end <= start {
		return
	}
	v.pruneItemCache(start, end)
	for i := start; i < end; i++ {
		itemHeight := v.list.ItemHeight(i)
		if itemHeight <= 0 {
			continue
		}
		itemY := content.Y + (v.list.OffsetForIndex(i) - offset)
		itemBounds := runtime.Rect{X: content.X, Y: itemY, Width: content.Width, Height: itemHeight}
		if itemBounds.Y >= content.Y+content.Height {
			break
		}
		if itemBounds.Y+itemBounds.Height <= content.Y {
			continue
		}
		widget := v.itemCache[i]
		if widget == nil {
			widget = v.acquireItem()
			if widget == nil {
				continue
			}
			if v.itemCache == nil {
				v.itemCache = make(map[int]runtime.Widget)
			}
			v.itemCache[i] = widget
		}
		item := v.adapter.Item(i)
		v.widgetFactory.UpdateWidget(widget, item, i, i == v.list.SelectedIndex())
		widget.Layout(itemBounds)
		if childCtx, ok := ctx.SubVisible(itemBounds); ok {
			widget.Render(childCtx)
		}
	}
}

func (v *VirtualList[T]) renderIndex(index int, selected bool, ctx runtime.RenderContext) {
	if v == nil {
		return
	}
	if v.adapter == nil {
		return
	}
	baseStyle := mergeBackendStyles(resolveBaseStyle(ctx, v, backend.DefaultStyle(), false), v.style)
	if selected {
		selectedStyle := mergeBackendStyles(baseStyle, v.selectedStyle)
		ctx.Buffer.Fill(ctx.Bounds, ' ', selectedStyle)
	}
	item := v.adapter.Item(index)
	v.adapter.Render(item, index, selected, ctx)
}

func (v *VirtualList[T]) handleSelection(index int) {
	if v == nil {
		return
	}
	if v.adapter != nil && index >= 0 && index < v.adapter.Count() && v.onSelect != nil {
		v.onSelect(index, v.adapter.Item(index))
	}
	v.syncA11y()
}

func (v *VirtualList[T]) applyHeightProvider() {
	if v == nil || v.list == nil || v.manualHeight {
		return
	}
	if v.adapter == nil {
		v.list.SetItemHeight(1)
		return
	}
	if fixed, ok := any(v.adapter).(VirtualListFixedHeightProvider); ok {
		v.list.SetItemHeight(fixed.FixedItemHeight())
		return
	}
	if variable, ok := any(v.adapter).(VirtualListHeightProvider); ok {
		v.list.SetItemHeightFunc(variable.ItemHeight)
		return
	}
	v.list.SetItemHeight(1)
}

func (v *VirtualList[T]) configureWidgetFactory() {
	if v == nil {
		return
	}
	v.releaseItemCache()
	v.widgetFactory = nil
	v.itemPool = nil
	if v.adapter == nil {
		return
	}
	factory, ok := any(v.adapter).(VirtualListWidgetFactory[T])
	if !ok {
		return
	}
	v.widgetFactory = factory
	resetFn := func(widget runtime.Widget) {
		if resetter, ok := any(v.adapter).(VirtualListWidgetResetter); ok {
			resetter.ResetWidget(widget)
		}
	}
	v.itemPool = runtime.NewWidgetPool(factory.NewWidget, resetFn, v.poolMax)
}

func (v *VirtualList[T]) acquireItem() runtime.Widget {
	if v == nil || v.widgetFactory == nil {
		return nil
	}
	var widget runtime.Widget
	if v.itemPool != nil {
		widget = v.itemPool.Acquire()
	} else {
		widget = v.widgetFactory.NewWidget()
	}
	if widget != nil {
		runtime.BindTree(widget, v.services)
	}
	return widget
}

func (v *VirtualList[T]) pruneItemCache(start, end int) {
	if v == nil || len(v.itemCache) == 0 {
		return
	}
	for index, widget := range v.itemCache {
		if index >= start && index < end {
			continue
		}
		v.releaseItem(index, widget)
	}
}

func (v *VirtualList[T]) releaseItem(index int, widget runtime.Widget) {
	if widget == nil {
		delete(v.itemCache, index)
		return
	}
	runtime.UnbindTree(widget)
	if v.itemPool != nil {
		v.itemPool.Release(widget)
	}
	delete(v.itemCache, index)
}

func (v *VirtualList[T]) releaseItemCache() {
	if v == nil || len(v.itemCache) == 0 {
		return
	}
	for index, widget := range v.itemCache {
		v.releaseItem(index, widget)
	}
}

func (v *VirtualList[T]) syncA11y() {
	if v == nil {
		return
	}
	if v.Base.Role == "" {
		v.Base.Role = accessibility.RoleList
	}
	label := strings.TrimSpace(v.label)
	if label == "" {
		label = "Virtual List"
	}
	v.Base.Label = label
	count := 0
	if v.adapter != nil {
		count = v.adapter.Count()
	}
	v.Base.Description = fmt.Sprintf("%d items", count)
	if count > 0 && v.list != nil {
		index := v.list.SelectedIndex()
		if index >= 0 && index < count && v.adapter != nil {
			item := v.adapter.Item(index)
			v.Base.Value = &accessibility.ValueInfo{Text: fmt.Sprint(item)}
		} else {
			v.Base.Value = nil
		}
	} else {
		v.Base.Value = nil
	}
}

// StyleType returns the selector type name.
func (v *VirtualList[T]) StyleType() string {
	return "VirtualList"
}

var _ scroll.Controller = (*VirtualList[any])(nil)
var _ runtime.Widget = (*VirtualList[any])(nil)
