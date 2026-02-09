package compositor

import (
	"math/rand"
	"testing"
)

// BenchmarkStyleToANSI_Default measures ANSI encoding for the default style
// (no colors, no attributes) — the minimal-work case.
func BenchmarkStyleToANSI_Default(b *testing.B) {
	s := DefaultStyle()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = StyleToANSI(s)
	}
}

// BenchmarkStyleToANSI_Full measures ANSI encoding for a style with RGB colors
// and multiple attributes — the maximal-work case.
func BenchmarkStyleToANSI_Full(b *testing.B) {
	s := Style{
		FG:            RGB(200, 100, 50),
		BG:            Color256(235),
		Bold:          true,
		Italic:        true,
		Underline:     true,
		Strikethrough: true,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = StyleToANSI(s)
	}
}

// BenchmarkStyleToANSI_Color16 measures ANSI encoding for a basic 16-color style.
func BenchmarkStyleToANSI_Color16(b *testing.B) {
	s := DefaultStyle().WithFG(ColorRed).WithBold(true)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = StyleToANSI(s)
	}
}

// BenchmarkANSIWriterWriteString measures the ANSIWriter's WriteString
// throughput, which is a hot path in the diff renderer.
func BenchmarkANSIWriterWriteString(b *testing.B) {
	w := NewANSIWriter()
	w.Grow(4096)
	text := "Hello, World!"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.WriteString(text)
		if w.Len() > 4000 {
			w.Reset()
		}
	}
}

// BenchmarkANSIWriterMoveTo measures cursor positioning with various patterns.
func BenchmarkANSIWriterMoveTo(b *testing.B) {
	w := NewANSIWriter()
	w.Grow(4096)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.MoveTo(i%80, i/80%24)
		w.WriteRune('X')
		if w.Len() > 4000 {
			w.Reset()
		}
	}
}

// BenchmarkANSIWriterReset measures the cost of resetting the writer for reuse,
// which happens once per frame.
func BenchmarkANSIWriterReset(b *testing.B) {
	w := NewANSIWriter()
	w.Grow(8192)
	// Fill it with some content to make reset non-trivial.
	for j := 0; j < 100; j++ {
		w.SetStyle(DefaultStyle().WithFG(ColorGreen))
		w.WriteString("content")
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Reset()
	}
}

// BenchmarkDiffRender80x24_NoDirty measures diff rendering for a standard
// 80x24 screen where nothing has changed between frames.
func BenchmarkDiffRender80x24_NoDirty(b *testing.B) {
	benchmarkDiffNoDirty(b, 80, 24)
}

// BenchmarkDiffRender200x50_NoDirty measures diff rendering for a large
// 200x50 screen where nothing has changed.
func BenchmarkDiffRender200x50_NoDirty(b *testing.B) {
	benchmarkDiffNoDirty(b, 200, 50)
}

func benchmarkDiffNoDirty(b *testing.B, width, height int) {
	screen := NewScreen(width, height)
	renderer := NewRenderer(screen)
	renderer.SetSyncOutput(false)
	style := DefaultStyle()

	// Initial full frame.
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			screen.Set(x, y, '.', style)
		}
	}
	_ = renderer.RenderFull()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Write the same content — nothing should differ.
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				screen.Set(x, y, '.', style)
			}
		}
		_ = renderer.Render()
	}
}

// BenchmarkDiffRender80x24_AllDirty measures diff rendering when every cell
// on an 80x24 screen has changed.
func BenchmarkDiffRender80x24_AllDirty(b *testing.B) {
	benchmarkDiffAllDirty(b, 80, 24)
}

// BenchmarkDiffRender200x50_AllDirty measures diff rendering when every cell
// on a 200x50 screen has changed.
func BenchmarkDiffRender200x50_AllDirty(b *testing.B) {
	benchmarkDiffAllDirty(b, 200, 50)
}

func benchmarkDiffAllDirty(b *testing.B, width, height int) {
	screen := NewScreen(width, height)
	renderer := NewRenderer(screen)
	renderer.SetSyncOutput(false)

	styleA := DefaultStyle().WithFG(ColorGreen)
	styleB := DefaultStyle().WithFG(ColorRed).WithBold(true)

	// Initial full frame.
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			screen.Set(x, y, 'A', styleA)
		}
	}
	_ = renderer.RenderFull()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Alternate between two styles to force every cell dirty.
		s := styleA
		ch := rune('A')
		if i%2 == 0 {
			s = styleB
			ch = 'B'
		}
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				screen.Set(x, y, ch, s)
			}
		}
		_ = renderer.Render()
	}
}

// BenchmarkDiffRender80x24_10Pct measures diff rendering when ~10% of
// cells have changed on an 80x24 screen.
func BenchmarkDiffRender80x24_10Pct(b *testing.B) {
	benchmarkDiffPartial(b, 80, 24, 0.10)
}

// BenchmarkDiffRender200x50_10Pct measures diff rendering when ~10% of
// cells have changed on a 200x50 screen.
func BenchmarkDiffRender200x50_10Pct(b *testing.B) {
	benchmarkDiffPartial(b, 200, 50, 0.10)
}

func benchmarkDiffPartial(b *testing.B, width, height int, pct float64) {
	screen := NewScreen(width, height)
	renderer := NewRenderer(screen)
	renderer.SetSyncOutput(false)
	baseStyle := DefaultStyle().WithFG(ColorGreen)
	dirtyStyle := DefaultStyle().WithFG(ColorRed).WithBold(true)

	// Initial full frame.
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			screen.Set(x, y, '.', baseStyle)
		}
	}
	_ = renderer.RenderFull()

	// Pre-generate random dirty positions.
	totalCells := width * height
	changeCount := int(float64(totalCells) * pct)
	rng := rand.New(rand.NewSource(42))
	positions := make([][2]int, changeCount)
	for i := range positions {
		positions[i] = [2]int{rng.Intn(width), rng.Intn(height)}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Repaint base.
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				screen.Set(x, y, '.', baseStyle)
			}
		}
		// Mutate ~10% of cells.
		for _, pos := range positions {
			screen.Set(pos[0], pos[1], 'X', dirtyStyle)
		}
		_ = renderer.Render()
	}
}

// BenchmarkStyleDelta measures the cost of computing the ANSI diff between
// two different styles.
func BenchmarkStyleDelta(b *testing.B) {
	from := DefaultStyle().WithFG(ColorGreen).WithBold(true)
	to := DefaultStyle().WithFG(ColorRed).WithItalic(true)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = StyleDelta(from, to)
	}
}

// BenchmarkStyleEqual measures Style.Equal, which is called for every cell
// during diff rendering.
func BenchmarkStyleEqual(b *testing.B) {
	a := DefaultStyle().WithFG(ColorGreen).WithBold(true)
	c := DefaultStyle().WithFG(ColorGreen).WithBold(true)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = a.Equal(c)
	}
}

// BenchmarkCellEqual measures Cell.Equal, used on every cell in the diff loop.
func BenchmarkCellEqual(b *testing.B) {
	a := Cell{Rune: 'A', Width: 1, Style: DefaultStyle().WithFG(ColorGreen)}
	c := Cell{Rune: 'A', Width: 1, Style: DefaultStyle().WithFG(ColorGreen)}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = a.Equal(c)
	}
}
