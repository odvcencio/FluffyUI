package graphics

import (
	"testing"

	"github.com/odvcencio/fluffy-ui/backend"
)

func TestBrailleBlitterSingleDot(t *testing.T) {
	buf := NewPixelBuffer(1, 1, &BrailleBlitter{})
	buf.SetPixel(0, 0, backend.ColorRed)
	blitter := &BrailleBlitter{}
	r, _ := blitter.BlitCell(buf, 0, 0)
	if r != rune(0x2801) {
		t.Fatalf("rune = %U, want %U", r, rune(0x2801))
	}
}

func TestSextantBlitterSpecialPatterns(t *testing.T) {
	buf := NewPixelBuffer(1, 1, &SextantBlitter{})
	blitter := &SextantBlitter{}

	// Left half (column 0)
	buf.Clear()
	buf.SetPixel(0, 0, backend.ColorRed)
	buf.SetPixel(0, 1, backend.ColorRed)
	buf.SetPixel(0, 2, backend.ColorRed)
	left, _ := blitter.BlitCell(buf, 0, 0)
	if left != '\u258C' {
		t.Fatalf("left half = %U, want %U", left, '\u258C')
	}

	// Right half (column 1)
	buf.Clear()
	buf.SetPixel(1, 0, backend.ColorRed)
	buf.SetPixel(1, 1, backend.ColorRed)
	buf.SetPixel(1, 2, backend.ColorRed)
	right, _ := blitter.BlitCell(buf, 0, 0)
	if right != '\u2590' {
		t.Fatalf("right half = %U, want %U", right, '\u2590')
	}

	// Full block
	buf.Clear()
	for y := 0; y < 3; y++ {
		for x := 0; x < 2; x++ {
			buf.SetPixel(x, y, backend.ColorRed)
		}
	}
	full, _ := blitter.BlitCell(buf, 0, 0)
	if full != '\u2588' {
		t.Fatalf("full = %U, want %U", full, '\u2588')
	}
}

func TestBestBlitterSelection(t *testing.T) {
	tests := []struct {
		name string
		caps Capabilities
		want string
	}{
		{name: "kitty", caps: Capabilities{Kitty: true, Unicode: true}, want: "kitty"},
		{name: "sixel", caps: Capabilities{Sixel: true, Unicode: true}, want: "sixel"},
		{name: "unicode", caps: Capabilities{Unicode: true}, want: "sextant"},
		{name: "ascii", caps: Capabilities{Unicode: false}, want: "ascii"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blitter := BestBlitter(&tt.caps)
			if blitter == nil {
				t.Fatalf("BestBlitter returned nil")
			}
			if got := blitter.Name(); got != tt.want {
				t.Fatalf("BestBlitter name = %q, want %q", got, tt.want)
			}
		})
	}
}
