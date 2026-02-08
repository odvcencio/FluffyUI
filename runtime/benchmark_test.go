package runtime

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/backend"
)

// ---------------------------------------------------------------------------
// BenchmarkFlexLayout — measure + layout a flex container with N children.
// ---------------------------------------------------------------------------

func BenchmarkFlexLayout(b *testing.B) {
	for _, n := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("children_%d", n), func(b *testing.B) {
			children := make([]FlexChild, n)
			for i := range children {
				children[i] = Fixed(newTestWidget(80, 1))
			}
			flex := VBox(children...)
			constraints := Loose(80, n+10)
			bounds := Rect{X: 0, Y: 0, Width: 80, Height: n + 10}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				flex.Measure(constraints)
				flex.Layout(bounds)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// BenchmarkMeasureCache — cache hits vs misses for the layout measure cache.
// ---------------------------------------------------------------------------

func BenchmarkMeasureCache(b *testing.B) {
	b.Run("hit", func(b *testing.B) {
		cache := NewMeasureCache()
		widget := newTestWidget(40, 10)
		constraints := Loose(80, 24)
		gen := cache.Generation()
		cache.Set(widget, constraints, Size{Width: 40, Height: 10}, gen)

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = cache.Get(widget, constraints, gen)
		}
	})

	b.Run("miss", func(b *testing.B) {
		cache := NewMeasureCache()
		widget := newTestWidget(40, 10)
		constraints := Loose(80, 24)
		// Don't populate the cache, so every lookup is a miss.

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			gen := cache.Generation()
			if _, ok := cache.Get(widget, constraints, gen); !ok {
				result := widget.Measure(constraints)
				cache.Set(widget, constraints, result, gen)
			}
			cache.Invalidate(widget)
		}
	})

	b.Run("mixed_100_widgets", func(b *testing.B) {
		cache := NewMeasureCache()
		widgets := make([]*testWidget, 100)
		for i := range widgets {
			widgets[i] = newTestWidget(20+i%40, 1+i%10)
		}
		constraints := Loose(80, 200)
		gen := cache.Generation()

		// Pre-populate half the cache.
		for i := 0; i < 50; i++ {
			cache.Set(widgets[i], constraints, widgets[i].Measure(constraints), gen)
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := widgets[i%100]
			if _, ok := cache.Get(w, constraints, gen); !ok {
				result := w.Measure(constraints)
				cache.Set(w, constraints, result, gen)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// BenchmarkMessageDispatch — send and receive messages through the priority queue.
// ---------------------------------------------------------------------------

func BenchmarkMessageDispatch(b *testing.B) {
	b.Run("send_high", func(b *testing.B) {
		q := NewPriorityQueue(1024)
		msg := KeyMsg{Key: 0, Rune: 'a'}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			q.TrySend(msg)
			// Drain to avoid filling the channel.
			select {
			case <-q.high:
			default:
			}
		}
	})

	b.Run("send_normal", func(b *testing.B) {
		q := NewPriorityQueue(1024)
		msg := ResizeMsg{Width: 80, Height: 24}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			q.TrySend(msg)
			select {
			case <-q.normal:
			default:
			}
		}
	})

	b.Run("send_low", func(b *testing.B) {
		q := NewPriorityQueue(1024)
		msg := TickMsg{Time: time.Now()}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			q.TrySend(msg)
			select {
			case <-q.low:
			default:
			}
		}
	})

	b.Run("roundtrip", func(b *testing.B) {
		q := NewPriorityQueue(1024)
		ctx := context.Background()
		msg := KeyMsg{Key: 0, Rune: 'x'}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			q.Send(msg)
			_, _ = q.Recv(ctx)
		}
	})

	b.Run("mixed_priority_burst", func(b *testing.B) {
		q := NewPriorityQueue(256)
		ctx := context.Background()
		highMsg := KeyMsg{Key: 0, Rune: 'k'}
		normalMsg := ResizeMsg{Width: 100, Height: 40}
		lowMsg := InvalidateMsg{}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Send a burst of mixed messages.
			q.Send(lowMsg)
			q.Send(normalMsg)
			q.Send(highMsg)
			// Drain all three.
			for j := 0; j < 3; j++ {
				_, _ = q.Recv(ctx)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// BenchmarkFocusTraversal — cycle focus next/prev through a widget tree.
// ---------------------------------------------------------------------------

func BenchmarkFocusTraversal(b *testing.B) {
	for _, n := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("next_%d_widgets", n), func(b *testing.B) {
			fs := NewFocusScope()
			for i := 0; i < n; i++ {
				fs.Register(newFocusable(fmt.Sprintf("w%d", i)))
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				fs.FocusNext()
			}
		})

		b.Run(fmt.Sprintf("prev_%d_widgets", n), func(b *testing.B) {
			fs := NewFocusScope()
			for i := 0; i < n; i++ {
				fs.Register(newFocusable(fmt.Sprintf("w%d", i)))
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				fs.FocusPrev()
			}
		})
	}
}

// ---------------------------------------------------------------------------
// BenchmarkFlexRender — measure + layout + render a flex tree.
// ---------------------------------------------------------------------------

func BenchmarkFlexRender(b *testing.B) {
	for _, n := range []int{10, 100} {
		b.Run(fmt.Sprintf("children_%d", n), func(b *testing.B) {
			children := make([]FlexChild, n)
			for i := range children {
				children[i] = Fixed(&renderLeaf{
					preferred: Size{Width: 80, Height: 1},
					text:      fmt.Sprintf("line %04d: some content here", i),
				})
			}
			flex := VBox(children...)
			bounds := Rect{X: 0, Y: 0, Width: 80, Height: n}
			flex.Measure(Loose(80, n))
			flex.Layout(bounds)

			buf := NewBuffer(80, n)
			ctx := RenderContext{Buffer: buf, Bounds: bounds}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				flex.Render(ctx)
			}
		})
	}
}

// renderLeaf is a simple widget that writes a text line during Render.
type renderLeaf struct {
	preferred Size
	bounds    Rect
	text      string
}

func (r *renderLeaf) Measure(c Constraints) Size {
	return c.Constrain(r.preferred)
}

func (r *renderLeaf) Layout(bounds Rect) {
	r.bounds = bounds
}

func (r *renderLeaf) Render(ctx RenderContext) {
	if ctx.Buffer == nil {
		return
	}
	ctx.Buffer.SetString(r.bounds.X, r.bounds.Y, r.text, backend.DefaultStyle())
}

func (r *renderLeaf) HandleMessage(msg Message) HandleResult {
	return Unhandled()
}
