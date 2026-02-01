package widgets

import (
	"strconv"
	"testing"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
)

func BenchmarkLabelRender(b *testing.B) {
	label := NewLabel("Benchmark", WithLabelStyle(backend.DefaultStyle().Bold(true)))
	benchmarkRender(b, label, 20, 1)
}

func BenchmarkListRender(b *testing.B) {
	items := make([]string, 200)
	for i := range items {
		items[i] = "Item " + strconv.Itoa(i)
	}
	adapter := NewSliceAdapter(items, func(item string, index int, selected bool, ctx runtime.RenderContext) {
		ctx.Buffer.SetString(ctx.Bounds.X, ctx.Bounds.Y, item, backend.DefaultStyle())
	})
	list := NewList(adapter)
	benchmarkRender(b, list, 30, 10)
}

func BenchmarkTableRender(b *testing.B) {
	table := NewTable(
		TableColumn{Title: "Name"},
		TableColumn{Title: "Value"},
		TableColumn{Title: "Status"},
	)
	rows := make([][]string, 120)
	for i := range rows {
		rows[i] = []string{"Row " + strconv.Itoa(i), strconv.Itoa(i), "OK"}
	}
	table.SetRows(rows)
	benchmarkRender(b, table, 40, 12)
}

func BenchmarkScrollViewRender(b *testing.B) {
	text := NewText("Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8\nLine 9\nLine 10")
	view := NewScrollView(text)
	benchmarkRender(b, view, 30, 5)
}

func BenchmarkGridRender1000Labels(b *testing.B) {
	const rows = 25
	const cols = 40
	grid := NewGrid(rows, cols)
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			grid.Add(NewLabel("x"), r, c, 1, 1)
		}
	}
	benchmarkRenderWithFPS(b, grid, 80, 25)
}

func benchmarkRender(b *testing.B, w runtime.Widget, width, height int) {
	buf := runtime.NewBuffer(width, height)
	constraints := runtime.Constraints{MaxWidth: width, MaxHeight: height}
	w.Measure(constraints)
	w.Layout(runtime.Rect{X: 0, Y: 0, Width: width, Height: height})
	ctx := runtime.RenderContext{Buffer: buf, Bounds: runtime.Rect{X: 0, Y: 0, Width: width, Height: height}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Render(ctx)
	}
}

func benchmarkRenderWithFPS(b *testing.B, w runtime.Widget, width, height int) {
	buf := runtime.NewBuffer(width, height)
	constraints := runtime.Constraints{MaxWidth: width, MaxHeight: height}
	w.Measure(constraints)
	w.Layout(runtime.Rect{X: 0, Y: 0, Width: width, Height: height})
	ctx := runtime.RenderContext{Buffer: buf, Bounds: runtime.Rect{X: 0, Y: 0, Width: width, Height: height}}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Render(ctx)
	}
	b.StopTimer()

	elapsed := b.Elapsed().Seconds()
	if elapsed > 0 {
		b.ReportMetric(float64(b.N)/elapsed, "frames/s")
	}
}
