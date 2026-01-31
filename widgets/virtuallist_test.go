package widgets

import (
	"strings"
	"testing"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/scroll"
	"github.com/odvcencio/fluffyui/terminal"
	flufftest "github.com/odvcencio/fluffyui/testing"
)

type testVirtualAdapter struct {
	items   []string
	renders int
}

func (a *testVirtualAdapter) Count() int { return len(a.items) }

func (a *testVirtualAdapter) Item(index int) string { return a.items[index] }

func (a *testVirtualAdapter) Render(item string, index int, selected bool, ctx runtime.RenderContext) {
	a.renders++
	prefix := "  "
	if selected {
		prefix = "> "
	}
	ctx.Buffer.SetString(ctx.Bounds.X, ctx.Bounds.Y, prefix+item, backend.DefaultStyle())
}

func (a *testVirtualAdapter) FixedItemHeight() int { return 1 }

type testVirtualWidgetAdapter struct {
	items    []string
	updates  int
	resets   int
	created  int
	lastText string
}

func (a *testVirtualWidgetAdapter) Count() int { return len(a.items) }

func (a *testVirtualWidgetAdapter) Item(index int) string { return a.items[index] }

func (a *testVirtualWidgetAdapter) Render(item string, index int, selected bool, ctx runtime.RenderContext) {
	// Render path is unused when widget factory is active.
}

func (a *testVirtualWidgetAdapter) FixedItemHeight() int { return 1 }

func (a *testVirtualWidgetAdapter) NewWidget() runtime.Widget {
	a.created++
	return NewLabel("")
}

func (a *testVirtualWidgetAdapter) UpdateWidget(widget runtime.Widget, item string, index int, selected bool) {
	a.updates++
	a.lastText = item
	if label, ok := widget.(*Label); ok {
		label.SetText(item)
	}
}

func (a *testVirtualWidgetAdapter) ResetWidget(widget runtime.Widget) {
	a.resets++
	if label, ok := widget.(*Label); ok {
		label.SetText("")
	}
}

func TestVirtualListSelectionAndScrolling(t *testing.T) {
	adapter := &testVirtualAdapter{items: []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon"}}
	list := NewVirtualList[string](adapter)
	list.SetLabel("Items")
	list.SetBehavior(scroll.ScrollBehavior{Vertical: scroll.ScrollAuto, Horizontal: scroll.ScrollNever, MouseWheel: 2, PageSize: 1})
	list.SetItemHeight(1)

	list.Layout(runtime.Rect{X: 0, Y: 0, Width: 16, Height: 3})
	list.SetSelected(2)
	if got := list.SelectedIndex(); got != 2 {
		t.Fatalf("expected selected index 2, got %d", got)
	}
	if item, ok := list.SelectedItem(); !ok || item != "Gamma" {
		t.Fatalf("expected selected item Gamma, got %v (ok=%v)", item, ok)
	}

	list.ScrollBy(0, 1)
	if got := list.SelectedIndex(); got != 3 {
		t.Fatalf("expected scroll by to select 3, got %d", got)
	}
	list.ScrollTo(0, 1)
	if got := list.SelectedIndex(); got != 1 {
		t.Fatalf("expected scroll to select 1, got %d", got)
	}
	list.ScrollToStart()
	if got := list.SelectedIndex(); got != 0 {
		t.Fatalf("expected scroll to start to select 0, got %d", got)
	}
	list.ScrollToEnd()
	if got := list.SelectedIndex(); got != len(adapter.items)-1 {
		t.Fatalf("expected scroll to end to select last, got %d", got)
	}

	list.ScrollToIndex(2)
	list.ScrollToOffset(1)
	if list.Offset() < 0 {
		t.Fatalf("expected non-negative offset")
	}
	list.PageBy(-1)
}

func TestVirtualListHandleMessageAndLazyLoad(t *testing.T) {
	adapter := &testVirtualAdapter{items: []string{"One", "Two", "Three", "Four", "Five"}}
	list := NewVirtualList[string](adapter)
	list.SetBehavior(scroll.ScrollBehavior{Vertical: scroll.ScrollAuto, Horizontal: scroll.ScrollNever, MouseWheel: 3, PageSize: 1})
	list.Layout(runtime.Rect{X: 0, Y: 0, Width: 12, Height: 3})
	list.Focus()

	var selected int
	list.SetOnSelect(func(index int, item string) {
		selected = index
	})

	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if list.SelectedIndex() != 1 {
		t.Fatalf("expected key down to select 1, got %d", list.SelectedIndex())
	}
	list.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if selected != 1 {
		t.Fatalf("expected enter to trigger select 1, got %d", selected)
	}

	list.HandleMessage(runtime.MouseMsg{Button: runtime.MouseWheelDown})
	if list.SelectedIndex() != 4 {
		t.Fatalf("expected mouse wheel down to select 4, got %d", list.SelectedIndex())
	}

	lazyCalls := 0
	list.SetLazyLoadThreshold(0)
	list.SetLazyLoad(func(start, end, total int) {
		lazyCalls++
		if total != len(adapter.items) {
			t.Fatalf("expected total %d, got %d", len(adapter.items), total)
		}
	})

	out := flufftest.RenderToString(list, 12, 3)
	if lazyCalls == 0 {
		t.Fatalf("expected lazy load to fire")
	}
	if !strings.Contains(out, ">") {
		t.Fatalf("expected selection indicator in render output, got:\n%s", out)
	}
}

func TestVirtualListWidgetFactoryPooling(t *testing.T) {
	adapter := &testVirtualWidgetAdapter{items: []string{"A", "B", "C", "D"}}
	list := NewVirtualList[string](adapter)
	list.SetOverscan(0)
	list.SetWidgetPoolMax(1)
	list.Layout(runtime.Rect{X: 0, Y: 0, Width: 8, Height: 1})
	list.Bind(runtime.Services{})

	_ = flufftest.RenderToString(list, 8, 1)
	if adapter.updates == 0 || adapter.created == 0 {
		t.Fatalf("expected widget factory to update/create widgets")
	}
	list.ScrollTo(0, 2)
	_ = flufftest.RenderToString(list, 8, 1)
	if adapter.resets == 0 {
		t.Fatalf("expected widget pool reset to be called")
	}

	list.Unbind()
}

func TestVirtualListUseAdapterHeights(t *testing.T) {
	adapter := &testVirtualAdapter{items: []string{"A", "B"}}
	list := NewVirtualList[string](adapter)
	list.SetItemHeight(2)
	if !list.manualHeight {
		t.Fatalf("expected manual height to be enabled")
	}
	list.UseAdapterHeights()
	if list.manualHeight {
		t.Fatalf("expected manual height to be disabled")
	}
	if list.list.ItemHeight(0) != 1 {
		t.Fatalf("expected adapter height 1, got %d", list.list.ItemHeight(0))
	}
}
