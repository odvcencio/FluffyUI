package scroll

import "testing"

func TestVirtualListFixedHeightIndexing(t *testing.T) {
	list := NewVirtualList(10, 2, nil)
	list.SetViewportHeight(6)

	if got := list.TotalHeight(); got != 20 {
		t.Fatalf("total height = %d, want 20", got)
	}
	if got := list.IndexForOffset(0); got != 0 {
		t.Fatalf("index for offset 0 = %d, want 0", got)
	}
	if got := list.IndexForOffset(3); got != 1 {
		t.Fatalf("index for offset 3 = %d, want 1", got)
	}
	if got := list.IndexForOffset(100); got != 9 {
		t.Fatalf("index for offset 100 = %d, want 9", got)
	}
	if got := list.OffsetForIndex(4); got != 8 {
		t.Fatalf("offset for index 4 = %d, want 8", got)
	}
}

func TestVirtualListVisibleRangeOverscan(t *testing.T) {
	list := NewVirtualList(100, 1, nil)
	list.SetViewportHeight(10)
	list.SetOverscan(2)
	list.ScrollToOffset(5)

	start, end := list.GetVisibleRange()
	if start != 3 || end != 17 {
		t.Fatalf("visible range = %d..%d, want 3..17", start, end)
	}
}

func TestVirtualListVariableHeights(t *testing.T) {
	heights := []int{1, 2, 3, 4}
	list := NewVirtualList(len(heights), 0, nil)
	list.SetItemHeightFunc(func(index int) int {
		return heights[index]
	})

	if got := list.TotalHeight(); got != 10 {
		t.Fatalf("total height = %d, want 10", got)
	}
	if got := list.IndexForOffset(1); got != 1 {
		t.Fatalf("index for offset 1 = %d, want 1", got)
	}
	if got := list.IndexForOffset(3); got != 2 {
		t.Fatalf("index for offset 3 = %d, want 2", got)
	}
	if got := list.OffsetForIndex(3); got != 6 {
		t.Fatalf("offset for index 3 = %d, want 6", got)
	}
}

func TestVirtualListEnsureVisible(t *testing.T) {
	list := NewVirtualList(20, 1, nil)
	list.SetViewportHeight(5)

	list.ScrollToOffset(0)
	list.EnsureVisible(10)
	if got := list.Offset(); got != 6 {
		t.Fatalf("offset after ensure visible 10 = %d, want 6", got)
	}

	list.EnsureVisible(2)
	if got := list.Offset(); got != 2 {
		t.Fatalf("offset after ensure visible 2 = %d, want 2", got)
	}
}

func TestVirtualListSelectionClamp(t *testing.T) {
	list := NewVirtualList(5, 1, nil)
	seen := -1
	list.SetOnSelection(func(index int) {
		seen = index
	})
	list.SetSelected(10)
	if got := list.SelectedIndex(); got != 4 {
		t.Fatalf("selected index = %d, want 4", got)
	}
	if seen != 4 {
		t.Fatalf("selection callback = %d, want 4", seen)
	}
}
