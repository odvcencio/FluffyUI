package widgets

import (
	"fmt"
	"strings"
	"testing"

	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
)

type virtualListTestAdapter struct{}

func (virtualListTestAdapter) Count() int {
	return 5
}

func (virtualListTestAdapter) Item(index int) string {
	return fmt.Sprintf("row%d", index)
}

func (virtualListTestAdapter) Render(item string, index int, selected bool, ctx runtime.RenderContext) {
	ctx.Buffer.SetString(ctx.Bounds.X, ctx.Bounds.Y, item, backend.DefaultStyle())
}

func (virtualListTestAdapter) FixedItemHeight() int {
	return 1
}

func TestVirtualListRenderOffset(t *testing.T) {
	list := NewVirtualList[string](virtualListTestAdapter{})

	list.Measure(runtime.Constraints{MaxWidth: 8, MaxHeight: 2})
	list.Layout(runtime.Rect{X: 0, Y: 0, Width: 8, Height: 2})
	list.ScrollToOffset(1)

	buf := runtime.NewBuffer(8, 2)
	list.Render(runtime.RenderContext{Buffer: buf})

	line0 := captureRow(buf, 0, 8)
	if !strings.HasPrefix(line0, "row1") {
		t.Fatalf("line0 = %q, want prefix %q", line0, "row1")
	}
}

func captureRow(buf *runtime.Buffer, y, width int) string {
	var sb strings.Builder
	for x := 0; x < width; x++ {
		cell := buf.Get(x, y)
		r := cell.Rune
		if r == 0 {
			r = ' '
		}
		sb.WriteRune(r)
	}
	return sb.String()
}
