package graphics

import (
	"testing"

	"github.com/odvcencio/fluffy-ui/backend"
)

func TestPixelBufferSetGet(t *testing.T) {
	buf := NewPixelBuffer(1, 1, &BrailleBlitter{})
	buf.SetPixel(0, 0, backend.ColorRed)
	p := buf.Get(0, 0)
	if !p.Set {
		t.Fatalf("pixel not set")
	}
	if p.Color != backend.ColorRed {
		t.Fatalf("color = %v, want %v", p.Color, backend.ColorRed)
	}
	if got := buf.Get(-1, 0); got.Set {
		t.Fatalf("expected out-of-bounds to be unset")
	}
}

func TestPixelBufferBlend(t *testing.T) {
	buf := NewPixelBuffer(1, 1, &BrailleBlitter{})
	buf.SetPixel(0, 0, backend.ColorRGB(0, 0, 0))
	buf.Blend(0, 0, backend.ColorRGB(255, 0, 0), 0.5)
	p := buf.Get(0, 0)
	if !p.Color.IsRGB() {
		t.Fatalf("expected RGB color")
	}
	r, g, b := p.Color.RGB()
	if r != 128 || g != 0 || b != 0 {
		t.Fatalf("blend RGB = %d,%d,%d, want 128,0,0", r, g, b)
	}
}
