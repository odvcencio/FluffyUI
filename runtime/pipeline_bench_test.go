package runtime

import (
	"strconv"
	"testing"
	"time"

	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/backend/sim"
)

func BenchmarkRenderPipelineDepth10(b *testing.B) {
	benchmarkRenderPipeline(b, 10)
}

func BenchmarkRenderPipelineDepth100(b *testing.B) {
	benchmarkRenderPipeline(b, 100)
}

func BenchmarkRenderPipelineDepth1000(b *testing.B) {
	benchmarkRenderPipeline(b, 1000)
}

func benchmarkRenderPipeline(b *testing.B, widgetCount int) {
	height := max(24, widgetCount+2)
	be := sim.New(160, height)
	if err := be.Init(); err != nil {
		b.Fatalf("init backend: %v", err)
	}
	defer be.Fini()

	frame := uint64(0)
	root := buildPipelineTree(widgetCount, &frame)

	app := NewApp(AppConfig{Backend: be, Root: root})
	w, h := be.Size()
	app.screen = NewScreen(w, h)
	app.screen.SetServices(app.Services())
	app.screen.SetRoot(root)

	var relayoutTotal time.Duration
	var renderTotal time.Duration
	var diffTotal time.Duration
	var dirtyTotal int

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		frame++

		relayoutStart := time.Now()
		app.screen.relayout()
		relayoutTotal += time.Since(relayoutStart)

		renderStart := time.Now()
		app.screen.Render()
		renderTotal += time.Since(renderStart)

		buf := app.screen.Buffer()
		diffStart := time.Now()
		dirty := 0
		buf.ForEachDirtyCell(func(x, y int, cell Cell) {
			dirty++
		})
		buf.ClearDirty()
		diffTotal += time.Since(diffStart)
		dirtyTotal += dirty
	}

	iterations := float64(b.N)
	b.ReportMetric(float64(relayoutTotal.Nanoseconds())/iterations, "relayout_ns/op")
	b.ReportMetric(float64(renderTotal.Nanoseconds())/iterations, "render_ns/op")
	b.ReportMetric(float64(diffTotal.Nanoseconds())/iterations, "diff_ns/op")
	b.ReportMetric(float64(dirtyTotal)/iterations, "dirty_cells/op")
}

type pipelineLeaf struct {
	bounds Rect
	lineA  string
	lineB  string
	frame  *uint64
}

func (w *pipelineLeaf) Measure(c Constraints) Size {
	return c.Constrain(Size{Width: len(w.lineA), Height: 1})
}

func (w *pipelineLeaf) Layout(bounds Rect) {
	w.bounds = bounds
}

func (w *pipelineLeaf) Render(ctx RenderContext) {
	if w == nil || ctx.Buffer == nil {
		return
	}
	if w.bounds.Width <= 0 || w.bounds.Height <= 0 {
		return
	}
	line := w.lineA
	if w.frame != nil && *w.frame%2 == 1 {
		line = w.lineB
	}
	ctx.Buffer.SetString(w.bounds.X, w.bounds.Y, line, backend.DefaultStyle())
}

func (w *pipelineLeaf) HandleMessage(msg Message) HandleResult {
	return Unhandled()
}

func buildPipelineTree(count int, frame *uint64) Widget {
	children := make([]FlexChild, 0, count)
	for i := 0; i < count; i++ {
		prefix := "node-" + strconv.Itoa(i) + "-"
		leaf := &pipelineLeaf{
			lineA: prefix + "A",
			lineB: prefix + "B",
			frame: frame,
		}
		children = append(children, Fixed(leaf))
	}
	return VBox(children...)
}

var _ Widget = (*pipelineLeaf)(nil)
