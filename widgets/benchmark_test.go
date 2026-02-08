package widgets

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

// ---------------------------------------------------------------------------
// BenchmarkLabelRender — render a simple label widget.
// ---------------------------------------------------------------------------

func BenchmarkLabelRender_Simple(b *testing.B) {
	label := NewLabel("Hello, FluffyUI!")
	benchmarkWidgetRender(b, label, 40, 1)
}

func BenchmarkLabelRender_Styled(b *testing.B) {
	label := NewLabel("Styled label text", WithLabelStyle(backend.DefaultStyle().Bold(true)))
	benchmarkWidgetRender(b, label, 40, 1)
}

func BenchmarkLabelRender_Long(b *testing.B) {
	// A 200-character label to stress the render path.
	text := ""
	for len(text) < 200 {
		text += "abcdefghij"
	}
	label := NewLabel(text)
	benchmarkWidgetRender(b, label, 200, 1)
}

// ---------------------------------------------------------------------------
// BenchmarkTableRender — render tables of varying row counts.
// ---------------------------------------------------------------------------

func BenchmarkTableRender_100rows(b *testing.B) {
	benchmarkTableN(b, 100)
}

func BenchmarkTableRender_500rows(b *testing.B) {
	benchmarkTableN(b, 500)
}

func BenchmarkTableRender_1000rows(b *testing.B) {
	benchmarkTableN(b, 1000)
}

func benchmarkTableN(b *testing.B, rowCount int) {
	table := NewTable(
		TableColumn{Title: "ID"},
		TableColumn{Title: "Name"},
		TableColumn{Title: "Status"},
		TableColumn{Title: "Score"},
	)
	rows := make([][]string, rowCount)
	for i := range rows {
		rows[i] = []string{
			strconv.Itoa(i),
			"Item-" + strconv.Itoa(i),
			"active",
			strconv.Itoa(i * 17 % 100),
		}
	}
	table.SetRows(rows)
	// Typical viewport height; table will render only visible rows.
	benchmarkWidgetRender(b, table, 80, 30)
}

// ---------------------------------------------------------------------------
// BenchmarkVirtualListRender — render a virtual list with 10,000 items
// (only visible items are rendered).
// ---------------------------------------------------------------------------

func BenchmarkVirtualListRender(b *testing.B) {
	for _, n := range []int{1000, 10000, 100000} {
		b.Run(fmt.Sprintf("items_%d", n), func(b *testing.B) {
			items := make([]string, n)
			for i := range items {
				items[i] = fmt.Sprintf("Item %d: some realistic content", i)
			}
			adapter := &benchVirtualAdapter{items: items}
			list := NewVirtualList[string](adapter)
			list.SetItemHeight(1)

			benchmarkWidgetRender(b, list, 60, 24)
		})
	}
}

// benchVirtualAdapter implements VirtualListAdapter for benchmarks.
type benchVirtualAdapter struct {
	items []string
}

func (a *benchVirtualAdapter) Count() int        { return len(a.items) }
func (a *benchVirtualAdapter) Item(i int) string  { return a.items[i] }
func (a *benchVirtualAdapter) FixedItemHeight() int { return 1 }

func (a *benchVirtualAdapter) Render(item string, index int, selected bool, ctx runtime.RenderContext) {
	prefix := "  "
	if selected {
		prefix = "> "
	}
	ctx.Buffer.SetString(ctx.Bounds.X, ctx.Bounds.Y, prefix+item, backend.DefaultStyle())
}

// ---------------------------------------------------------------------------
// BenchmarkInputHandleKey — process keystrokes in a focused input widget.
// ---------------------------------------------------------------------------

func BenchmarkInputHandleKey(b *testing.B) {
	b.Run("single_char", func(b *testing.B) {
		input := NewInput()
		input.Focus()
		msg := runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'a'}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			input.HandleMessage(msg)
		}
	})

	b.Run("backspace", func(b *testing.B) {
		input := NewInput()
		input.Focus()
		input.SetText("Hello World")

		charMsg := runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'x'}
		bsMsg := runtime.KeyMsg{Key: terminal.KeyBackspace}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			input.HandleMessage(charMsg)
			input.HandleMessage(bsMsg)
		}
	})

	b.Run("rapid_typing_50_chars", func(b *testing.B) {
		keys := []rune("The quick brown fox jumps over the lazy dog fast!!")

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			input := NewInput()
			input.Focus()
			for _, ch := range keys {
				input.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: ch})
			}
		}
	})
}

// ---------------------------------------------------------------------------
// BenchmarkInputRender — render an input widget.
// ---------------------------------------------------------------------------

func BenchmarkInputRender(b *testing.B) {
	input := NewInput()
	input.Focus()
	input.SetText("Some text content for rendering benchmark")
	benchmarkWidgetRender(b, input, 50, 1)
}

// ---------------------------------------------------------------------------
// BenchmarkGridRender — render a grid of labels.
// ---------------------------------------------------------------------------

func BenchmarkGridRender(b *testing.B) {
	for _, sz := range []struct {
		rows, cols int
	}{
		{5, 5},
		{25, 10},
		{50, 20},
	} {
		b.Run(fmt.Sprintf("%dx%d", sz.rows, sz.cols), func(b *testing.B) {
			grid := NewGrid(sz.rows, sz.cols)
			for r := 0; r < sz.rows; r++ {
				for c := 0; c < sz.cols; c++ {
					grid.Add(NewLabel(fmt.Sprintf("(%d,%d)", r, c)), r, c, 1, 1)
				}
			}
			benchmarkWidgetRender(b, grid, sz.cols*8, sz.rows)
		})
	}
}

// ---------------------------------------------------------------------------
// Helper: benchmarkWidgetRender runs the standard Measure+Layout+Render cycle.
// ---------------------------------------------------------------------------

func benchmarkWidgetRender(b *testing.B, w runtime.Widget, width, height int) {
	b.Helper()
	buf := runtime.NewBuffer(width, height)
	constraints := runtime.Constraints{MaxWidth: width, MaxHeight: height}
	w.Measure(constraints)
	w.Layout(runtime.Rect{X: 0, Y: 0, Width: width, Height: height})
	ctx := runtime.RenderContext{
		Buffer: buf,
		Bounds: runtime.Rect{X: 0, Y: 0, Width: width, Height: height},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Render(ctx)
	}
}
