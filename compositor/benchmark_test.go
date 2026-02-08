package compositor

import (
	"math/rand"
	"testing"
)

// BenchmarkRenderFullScreen measures full-screen rendering at standard and
// large terminal sizes. Every cell is populated with styled content.
func BenchmarkRenderFullScreen(b *testing.B) {
	sizes := []struct {
		name   string
		width  int
		height int
	}{
		{"80x24", 80, 24},
		{"200x50", 200, 50},
	}

	for _, sz := range sizes {
		b.Run(sz.name, func(b *testing.B) {
			screen := NewScreen(sz.width, sz.height)
			renderer := NewRenderer(screen)

			styles := []Style{
				DefaultStyle().WithFG(ColorRed).WithBold(true),
				DefaultStyle().WithFG(ColorGreen).WithItalic(true),
				DefaultStyle().WithFG(ColorBlue).WithBG(Color256(235)),
				DefaultStyle().WithFG(RGB(200, 100, 50)).WithUnderline(true),
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Fill every cell with varied content and styles.
				for y := 0; y < sz.height; y++ {
					for x := 0; x < sz.width; x++ {
						ch := rune('A' + (x+y)%26)
						screen.Set(x, y, ch, styles[(x+y)%len(styles)])
					}
				}
				_ = renderer.RenderFull()
			}
		})
	}
}

// BenchmarkRenderDiff measures incremental diff rendering when roughly 10%
// of cells have changed between frames.
func BenchmarkRenderDiff(b *testing.B) {
	sizes := []struct {
		name   string
		width  int
		height int
	}{
		{"80x24", 80, 24},
		{"200x50", 200, 50},
	}

	for _, sz := range sizes {
		b.Run(sz.name, func(b *testing.B) {
			screen := NewScreen(sz.width, sz.height)
			renderer := NewRenderer(screen)
			style := DefaultStyle().WithFG(ColorGreen)

			// Initial full frame.
			for y := 0; y < sz.height; y++ {
				for x := 0; x < sz.width; x++ {
					screen.Set(x, y, '.', style)
				}
			}
			_ = renderer.RenderFull()

			totalCells := sz.width * sz.height
			changePct := 0.10
			changeCount := int(float64(totalCells) * changePct)

			// Pre-generate random positions for changed cells.
			rng := rand.New(rand.NewSource(42))
			positions := make([][2]int, changeCount)
			for i := range positions {
				positions[i] = [2]int{rng.Intn(sz.width), rng.Intn(sz.height)}
			}

			dirtyStyle := DefaultStyle().WithFG(ColorRed).WithBold(true)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Repaint all cells to base state.
				for y := 0; y < sz.height; y++ {
					for x := 0; x < sz.width; x++ {
						screen.Set(x, y, '.', style)
					}
				}
				// Mutate ~10% of cells.
				for _, pos := range positions {
					screen.Set(pos[0], pos[1], 'X', dirtyStyle)
				}
				_ = renderer.Render()
			}
		})
	}
}

// BenchmarkRenderNoChange measures the cost of a diff render when nothing
// has changed. This should be near-zero work.
func BenchmarkRenderNoChange(b *testing.B) {
	sizes := []struct {
		name   string
		width  int
		height int
	}{
		{"80x24", 80, 24},
		{"200x50", 200, 50},
	}

	for _, sz := range sizes {
		b.Run(sz.name, func(b *testing.B) {
			screen := NewScreen(sz.width, sz.height)
			renderer := NewRenderer(screen)
			style := DefaultStyle()

			// Populate and do initial full render.
			for y := 0; y < sz.height; y++ {
				for x := 0; x < sz.width; x++ {
					screen.Set(x, y, ' ', style)
				}
			}
			_ = renderer.RenderFull()

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Write the exact same content (no changes).
				for y := 0; y < sz.height; y++ {
					for x := 0; x < sz.width; x++ {
						screen.Set(x, y, ' ', style)
					}
				}
				output := renderer.Render()
				_ = output
			}
		})
	}
}
