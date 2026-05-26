package widgets

import (
	"strconv"
	"testing"

	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/runtime"
)

// BenchmarkLabelBind measures the cost of binding a Label to services.
func BenchmarkLabelBind(b *testing.B) {
	label := NewLabel("Benchmark")
	services := runtime.Services{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		label.Bind(services)
	}
}

// BenchmarkLabelUnbind measures the cost of unbinding a Label.
func BenchmarkLabelUnbind(b *testing.B) {
	label := NewLabel("Benchmark")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		label.Unbind()
	}
}

// BenchmarkTableBind measures the cost of binding a Table to services.
func BenchmarkTableBind(b *testing.B) {
	table := NewTable(
		TableColumn{Title: "Name"},
		TableColumn{Title: "Value"},
		TableColumn{Title: "Status"},
	)
	rows := make([][]string, 100)
	for i := range rows {
		rows[i] = []string{"Row " + strconv.Itoa(i), strconv.Itoa(i), "OK"}
	}
	table.SetRows(rows)
	services := runtime.Services{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		table.Bind(services)
	}
}

// BenchmarkTableUnbind measures the cost of unbinding a Table.
func BenchmarkTableUnbind(b *testing.B) {
	table := NewTable(
		TableColumn{Title: "Name"},
		TableColumn{Title: "Value"},
	)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		table.Unbind()
	}
}

// BenchmarkLabelMeasure measures the cost of measuring a Label.
func BenchmarkLabelMeasure(b *testing.B) {
	label := NewLabel("Hello, World!", WithLabelStyle(backend.DefaultStyle().Bold(true)))
	constraints := runtime.Constraints{MaxWidth: 80, MaxHeight: 24}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = label.Measure(constraints)
	}
}

// BenchmarkTableMeasure measures the cost of measuring a Table with 100 rows.
func BenchmarkTableMeasure(b *testing.B) {
	table := NewTable(
		TableColumn{Title: "Name"},
		TableColumn{Title: "Value"},
		TableColumn{Title: "Status"},
	)
	rows := make([][]string, 100)
	for i := range rows {
		rows[i] = []string{"Row " + strconv.Itoa(i), strconv.Itoa(i), "OK"}
	}
	table.SetRows(rows)
	constraints := runtime.Constraints{MaxWidth: 80, MaxHeight: 24}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = table.Measure(constraints)
	}
}

// BenchmarkFlexLayout10Children measures layout of a VBox with 10 Label children.
func BenchmarkFlexLayout10Children(b *testing.B) {
	children := make([]runtime.FlexChild, 10)
	for i := range children {
		children[i] = runtime.Fixed(NewLabel("Label " + strconv.Itoa(i)))
	}
	flex := runtime.VBox(children...)
	constraints := runtime.Constraints{MaxWidth: 80, MaxHeight: 40}
	flex.Measure(constraints)
	bounds := runtime.Rect{X: 0, Y: 0, Width: 80, Height: 40}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		flex.Measure(constraints)
		flex.Layout(bounds)
	}
}

// BenchmarkFlexLayout50Children measures layout of a VBox with 50 Label children.
func BenchmarkFlexLayout50Children(b *testing.B) {
	children := make([]runtime.FlexChild, 50)
	for i := range children {
		children[i] = runtime.Fixed(NewLabel("Label " + strconv.Itoa(i)))
	}
	flex := runtime.VBox(children...)
	constraints := runtime.Constraints{MaxWidth: 80, MaxHeight: 80}
	flex.Measure(constraints)
	bounds := runtime.Rect{X: 0, Y: 0, Width: 80, Height: 80}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		flex.Measure(constraints)
		flex.Layout(bounds)
	}
}

// BenchmarkFullRenderCycleLabel measures the complete Measure+Layout+Render
// cycle for a single Label widget.
func BenchmarkFullRenderCycleLabel(b *testing.B) {
	label := NewLabel("Hello, Benchmark World!")
	width, height := 40, 1
	buf := runtime.NewBuffer(width, height)
	constraints := runtime.Constraints{MaxWidth: width, MaxHeight: height}
	bounds := runtime.Rect{X: 0, Y: 0, Width: width, Height: height}
	ctx := runtime.RenderContext{Buffer: buf, Bounds: bounds}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		label.Measure(constraints)
		label.Layout(bounds)
		label.Render(ctx)
	}
}

// BenchmarkFullRenderCycleTable measures the complete Measure+Layout+Render
// cycle for a Table with 50 rows.
func BenchmarkFullRenderCycleTable(b *testing.B) {
	table := NewTable(
		TableColumn{Title: "Name"},
		TableColumn{Title: "Value"},
	)
	rows := make([][]string, 50)
	for i := range rows {
		rows[i] = []string{"Item " + strconv.Itoa(i), strconv.Itoa(i * 10)}
	}
	table.SetRows(rows)

	width, height := 60, 20
	buf := runtime.NewBuffer(width, height)
	constraints := runtime.Constraints{MaxWidth: width, MaxHeight: height}
	bounds := runtime.Rect{X: 0, Y: 0, Width: width, Height: height}
	ctx := runtime.RenderContext{Buffer: buf, Bounds: bounds}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		table.Measure(constraints)
		table.Layout(bounds)
		table.Render(ctx)
	}
}

// BenchmarkBindTreeFlex measures BindTree on a Flex with 10 Label children,
// exercising the recursive bind walk.
func BenchmarkBindTreeFlex(b *testing.B) {
	children := make([]runtime.FlexChild, 10)
	for i := range children {
		children[i] = runtime.Fixed(NewLabel("Child " + strconv.Itoa(i)))
	}
	flex := runtime.VBox(children...)
	services := runtime.Services{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runtime.BindTree(flex, services)
	}
}

// BenchmarkUnbindTreeFlex measures UnbindTree on a Flex with 10 Label children.
func BenchmarkUnbindTreeFlex(b *testing.B) {
	children := make([]runtime.FlexChild, 10)
	for i := range children {
		children[i] = runtime.Fixed(NewLabel("Child " + strconv.Itoa(i)))
	}
	flex := runtime.VBox(children...)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runtime.UnbindTree(flex)
	}
}
