package scroll

import (
	"sort"

	"github.com/odvcencio/fluffy-ui/runtime"
)

// VirtualRenderFunc renders an item at the given index.
type VirtualRenderFunc func(index int, selected bool, ctx runtime.RenderContext)

// VirtualHeightFunc returns the height for an item index.
type VirtualHeightFunc func(index int) int

// VirtualItemFunc returns the item data for an index.
type VirtualItemFunc func(index int) any

// VirtualList provides windowed calculations for large lists.
// It implements VirtualContent plus optional sizing and indexing helpers.
type VirtualList struct {
	itemCount      int
	viewportHeight int
	fixedHeight    int
	itemHeight     VirtualHeightFunc
	overscan       int

	renderItem VirtualRenderFunc
	itemAt     VirtualItemFunc
	onSelect   func(index int)

	selected int
	offset   int

	prefix      []int
	prefixCount int
	prefixDirty bool
}

// NewVirtualList creates a virtual list with fixed item height.
func NewVirtualList(itemCount int, itemHeight int, render VirtualRenderFunc) *VirtualList {
	list := &VirtualList{
		itemCount:   itemCount,
		fixedHeight: itemHeight,
		renderItem:  render,
	}
	if list.itemCount < 0 {
		list.itemCount = 0
	}
	return list
}

// SetItemCount updates the item count and clamps offset/selection.
func (v *VirtualList) SetItemCount(count int) {
	if v == nil {
		return
	}
	if count < 0 {
		count = 0
	}
	if v.itemCount != count {
		v.itemCount = count
		v.invalidateHeights()
	}
	v.clampSelection(false)
	v.clampOffset()
}

// SetViewportHeight updates the viewport height and clamps the offset.
func (v *VirtualList) SetViewportHeight(height int) {
	if v == nil {
		return
	}
	if height < 0 {
		height = 0
	}
	if v.viewportHeight != height {
		v.viewportHeight = height
	}
	v.clampOffset()
}

// SetItemHeight sets a fixed item height.
func (v *VirtualList) SetItemHeight(height int) {
	if v == nil {
		return
	}
	v.fixedHeight = height
	v.itemHeight = nil
	v.invalidateHeights()
	v.clampOffset()
}

// SetItemHeightFunc sets a variable item height function.
func (v *VirtualList) SetItemHeightFunc(fn VirtualHeightFunc) {
	if v == nil {
		return
	}
	v.itemHeight = fn
	v.invalidateHeights()
	v.clampOffset()
}

// SetOverscan updates the overscan item count.
func (v *VirtualList) SetOverscan(count int) {
	if v == nil {
		return
	}
	if count < 0 {
		count = 0
	}
	v.overscan = count
}

// SetRenderItem updates the render callback.
func (v *VirtualList) SetRenderItem(fn VirtualRenderFunc) {
	if v == nil {
		return
	}
	v.renderItem = fn
}

// SetItemAt updates the item callback.
func (v *VirtualList) SetItemAt(fn VirtualItemFunc) {
	if v == nil {
		return
	}
	v.itemAt = fn
}

// SetOnSelection updates the selection callback.
func (v *VirtualList) SetOnSelection(fn func(index int)) {
	if v == nil {
		return
	}
	v.onSelect = fn
}

// SelectedIndex returns the selected index.
func (v *VirtualList) SelectedIndex() int {
	if v == nil {
		return 0
	}
	return v.selected
}

// SetSelected updates the selected index.
func (v *VirtualList) SetSelected(index int) {
	if v == nil {
		return
	}
	v.selected = index
	v.clampSelection(true)
}

// Offset returns the current scroll offset.
func (v *VirtualList) Offset() int {
	if v == nil {
		return 0
	}
	return v.offset
}

// ScrollToIndex scrolls to the given item index.
func (v *VirtualList) ScrollToIndex(index int) {
	if v == nil {
		return
	}
	v.ScrollToOffset(v.OffsetForIndex(index))
}

// ScrollToOffset scrolls to the given offset.
func (v *VirtualList) ScrollToOffset(pixels int) {
	if v == nil {
		return
	}
	v.offset = pixels
	v.clampOffset()
}

// ScrollBy scrolls the offset by the provided delta.
func (v *VirtualList) ScrollBy(delta int) {
	if v == nil || delta == 0 {
		return
	}
	v.ScrollToOffset(v.offset + delta)
}

// EnsureVisible scrolls just enough to keep the item fully visible.
func (v *VirtualList) EnsureVisible(index int) {
	if v == nil {
		return
	}
	count := v.count()
	if count == 0 {
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= count {
		index = count - 1
	}
	if v.viewportHeight <= 0 {
		return
	}
	itemTop := v.OffsetForIndex(index)
	itemHeight := v.ItemHeight(index)
	itemBottom := itemTop + itemHeight
	if itemTop < v.offset {
		v.offset = itemTop
		v.clampOffset()
		return
	}
	if itemBottom > v.offset+v.viewportHeight {
		v.offset = itemBottom - v.viewportHeight
		v.clampOffset()
	}
}

// GetVisibleRange returns the [start, end) item range for the current viewport,
// including overscan items.
func (v *VirtualList) GetVisibleRange() (start, end int) {
	if v == nil {
		return 0, 0
	}
	count := v.count()
	if count == 0 || v.viewportHeight <= 0 {
		return 0, 0
	}
	start = v.IndexForOffset(v.offset)
	endOffset := v.offset + v.viewportHeight - 1
	if endOffset < v.offset {
		endOffset = v.offset
	}
	end = v.IndexForOffset(endOffset) + 1
	if start < 0 {
		start = 0
	}
	if end > count {
		end = count
	}
	if v.overscan > 0 {
		start -= v.overscan
		end += v.overscan
		if start < 0 {
			start = 0
		}
		if end > count {
			end = count
		}
	}
	if end < start {
		end = start
	}
	return start, end
}

// MaxOffset returns the maximum scrollable offset.
func (v *VirtualList) MaxOffset() int {
	if v == nil {
		return 0
	}
	total := v.TotalHeight()
	if total <= 0 || v.viewportHeight <= 0 {
		return 0
	}
	maxOffset := total - v.viewportHeight
	if maxOffset < 0 {
		return 0
	}
	return maxOffset
}

// InvalidateHeights clears cached height data for variable-height lists.
func (v *VirtualList) InvalidateHeights() {
	if v == nil {
		return
	}
	v.invalidateHeights()
}

// ItemCount returns the item count.
func (v *VirtualList) ItemCount() int {
	if v == nil {
		return 0
	}
	return v.count()
}

// ItemHeight returns the height for the given item index.
func (v *VirtualList) ItemHeight(index int) int {
	if v == nil {
		return 0
	}
	if v.itemHeight != nil {
		height := v.itemHeight(index)
		if height < 0 {
			return 0
		}
		return height
	}
	if v.fixedHeight > 0 {
		return v.fixedHeight
	}
	return 1
}

// RenderItem renders the item at index.
func (v *VirtualList) RenderItem(index int, ctx runtime.RenderContext) {
	if v == nil || v.renderItem == nil {
		return
	}
	v.renderItem(index, index == v.selected, ctx)
}

// ItemAt returns the item data at the given index.
func (v *VirtualList) ItemAt(index int) any {
	if v == nil || v.itemAt == nil {
		return nil
	}
	return v.itemAt(index)
}

// TotalHeight returns the total height for all items.
func (v *VirtualList) TotalHeight() int {
	if v == nil {
		return 0
	}
	count := v.count()
	if count == 0 {
		return 0
	}
	if v.itemHeight == nil {
		if v.fixedHeight <= 0 {
			return 0
		}
		return v.fixedHeight * count
	}
	v.ensurePrefix()
	if len(v.prefix) == 0 {
		return 0
	}
	return v.prefix[len(v.prefix)-1]
}

// IndexForOffset maps a scroll offset to an item index.
func (v *VirtualList) IndexForOffset(offset int) int {
	if v == nil {
		return 0
	}
	count := v.count()
	if count == 0 {
		return 0
	}
	if offset <= 0 {
		return 0
	}
	if v.itemHeight == nil {
		height := v.fixedHeight
		if height <= 0 {
			return 0
		}
		index := offset / height
		if index < 0 {
			index = 0
		}
		if index >= count {
			index = count - 1
		}
		return index
	}
	v.ensurePrefix()
	if len(v.prefix) == 0 {
		return 0
	}
	total := v.prefix[len(v.prefix)-1]
	if offset >= total {
		return count - 1
	}
	idx := sort.Search(len(v.prefix), func(i int) bool {
		return v.prefix[i] > offset
	}) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= count {
		idx = count - 1
	}
	return idx
}

// OffsetForIndex maps an item index to its scroll offset.
func (v *VirtualList) OffsetForIndex(index int) int {
	if v == nil {
		return 0
	}
	count := v.count()
	if count == 0 {
		return 0
	}
	if index <= 0 {
		return 0
	}
	if index >= count {
		index = count - 1
	}
	if v.itemHeight == nil {
		height := v.fixedHeight
		if height <= 0 {
			return 0
		}
		return index * height
	}
	v.ensurePrefix()
	if len(v.prefix) == 0 || index >= len(v.prefix) {
		return 0
	}
	return v.prefix[index]
}

func (v *VirtualList) count() int {
	if v.itemCount < 0 {
		return 0
	}
	return v.itemCount
}

func (v *VirtualList) invalidateHeights() {
	v.prefixDirty = true
}

func (v *VirtualList) ensurePrefix() {
	if v == nil {
		return
	}
	if v.itemHeight == nil {
		return
	}
	count := v.count()
	if count < 0 {
		count = 0
	}
	if !v.prefixDirty && v.prefixCount == count {
		return
	}
	if count == 0 {
		v.prefix = v.prefix[:0]
		v.prefixCount = count
		v.prefixDirty = false
		return
	}
	if cap(v.prefix) < count+1 {
		v.prefix = make([]int, count+1)
	} else {
		v.prefix = v.prefix[:count+1]
	}
	v.prefix[0] = 0
	for i := 0; i < count; i++ {
		height := v.ItemHeight(i)
		if height < 0 {
			height = 0
		}
		v.prefix[i+1] = v.prefix[i] + height
	}
	v.prefixCount = count
	v.prefixDirty = false
}

func (v *VirtualList) clampSelection(notify bool) {
	count := v.count()
	if count == 0 {
		v.selected = 0
		return
	}
	if v.selected < 0 {
		v.selected = 0
	}
	if v.selected >= count {
		v.selected = count - 1
	}
	if notify && v.onSelect != nil {
		v.onSelect(v.selected)
	}
}

func (v *VirtualList) clampOffset() {
	maxOffset := v.MaxOffset()
	if v.offset < 0 {
		v.offset = 0
	}
	if v.offset > maxOffset {
		v.offset = maxOffset
	}
}

var _ VirtualContent = (*VirtualList)(nil)
var _ VirtualSizer = (*VirtualList)(nil)
var _ VirtualIndexer = (*VirtualList)(nil)
