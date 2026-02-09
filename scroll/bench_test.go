package scroll

import (
	"testing"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
)

// BenchmarkVirtualListSetSelected_10K measures SetSelected on a 10K-item list
// with fixed height. This exercises clamping and selection bookkeeping.
func BenchmarkVirtualListSetSelected_10K(b *testing.B) {
	list := NewVirtualList(10_000, 1, nil)
	list.SetViewportHeight(50)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list.SetSelected(i % 10_000)
	}
}

// BenchmarkVirtualListSetSelected_100K measures SetSelected on a 100K-item list.
func BenchmarkVirtualListSetSelected_100K(b *testing.B) {
	list := NewVirtualList(100_000, 1, nil)
	list.SetViewportHeight(50)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list.SetSelected(i % 100_000)
	}
}

// BenchmarkVirtualListScrollToIndex_10K measures ScrollToIndex which converts
// an item index to a pixel offset and then clamps the scroll position.
func BenchmarkVirtualListScrollToIndex_10K(b *testing.B) {
	list := NewVirtualList(10_000, 1, nil)
	list.SetViewportHeight(50)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list.ScrollToIndex(i % 10_000)
	}
}

// BenchmarkVirtualListScrollToIndex_100K measures ScrollToIndex on 100K items.
func BenchmarkVirtualListScrollToIndex_100K(b *testing.B) {
	list := NewVirtualList(100_000, 1, nil)
	list.SetViewportHeight(50)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list.ScrollToIndex(i % 100_000)
	}
}

// BenchmarkVirtualListGetVisibleRange_10K measures GetVisibleRange which
// computes the [start, end) item range for the viewport.
func BenchmarkVirtualListGetVisibleRange_10K(b *testing.B) {
	list := NewVirtualList(10_000, 1, nil)
	list.SetViewportHeight(50)
	list.SetOverscan(5)
	list.ScrollToIndex(5000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = list.GetVisibleRange()
	}
}

// BenchmarkVirtualListGetVisibleRange_100K measures GetVisibleRange on 100K items.
func BenchmarkVirtualListGetVisibleRange_100K(b *testing.B) {
	list := NewVirtualList(100_000, 1, nil)
	list.SetViewportHeight(50)
	list.SetOverscan(5)
	list.ScrollToIndex(50_000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = list.GetVisibleRange()
	}
}

// BenchmarkVirtualListRenderVisible_10K measures rendering the visible items
// in a 10K-item list. This exercises the full render pipeline per visible row.
func BenchmarkVirtualListRenderVisible_10K(b *testing.B) {
	benchmarkRenderVisible(b, 10_000, 50, false)
}

// BenchmarkVirtualListRenderVisible_100K measures rendering visible items
// in a 100K-item list.
func BenchmarkVirtualListRenderVisible_100K(b *testing.B) {
	benchmarkRenderVisible(b, 100_000, 50, false)
}

// BenchmarkVirtualListVariableHeight_10K measures operations on a variable-height
// 10K list, exercising the prefix sum calculation and binary search.
func BenchmarkVirtualListVariableHeight_10K(b *testing.B) {
	benchmarkRenderVisible(b, 10_000, 50, true)
}

func benchmarkRenderVisible(b *testing.B, itemCount, viewportHeight int, variableHeight bool) {
	style := backend.DefaultStyle()
	renderFn := func(index int, selected bool, ctx runtime.RenderContext) {
		ctx.Buffer.SetString(ctx.Bounds.X, ctx.Bounds.Y, "item", style)
	}
	list := NewVirtualList(itemCount, 1, renderFn)
	if variableHeight {
		list.SetItemHeightFunc(func(index int) int {
			// Heights alternate between 1 and 2.
			if index%3 == 0 {
				return 2
			}
			return 1
		})
	}
	list.SetViewportHeight(viewportHeight)
	list.SetOverscan(3)
	list.ScrollToIndex(itemCount / 2)

	buf := runtime.NewBuffer(80, viewportHeight)
	bounds := runtime.Rect{X: 0, Y: 0, Width: 80, Height: viewportHeight}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start, end := list.GetVisibleRange()
		y := 0
		for idx := start; idx < end && y < viewportHeight; idx++ {
			h := list.ItemHeight(idx)
			ctx := runtime.RenderContext{
				Buffer: buf,
				Bounds: runtime.Rect{X: bounds.X, Y: bounds.Y + y, Width: bounds.Width, Height: h},
			}
			list.RenderItem(idx, ctx)
			y += h
		}
	}
}

// BenchmarkVirtualListEnsureVisible measures EnsureVisible which adjusts
// scroll offset to keep a given item fully in view.
func BenchmarkVirtualListEnsureVisible(b *testing.B) {
	list := NewVirtualList(10_000, 1, nil)
	list.SetViewportHeight(50)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list.EnsureVisible(i % 10_000)
	}
}

// BenchmarkVirtualListTotalHeight_VariableHeight measures TotalHeight
// which forces prefix sum computation for variable-height items.
func BenchmarkVirtualListTotalHeight_VariableHeight(b *testing.B) {
	list := NewVirtualList(10_000, 1, nil)
	list.SetItemHeightFunc(func(index int) int {
		if index%3 == 0 {
			return 2
		}
		return 1
	})
	list.SetViewportHeight(50)

	// First call builds the prefix sum; subsequent calls are cached.
	_ = list.TotalHeight()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = list.TotalHeight()
	}
}
