package graphics

import (
	"testing"

	"github.com/odvcencio/fluffyui/backend"
)

type testBlitter struct{}

func (testBlitter) Name() string              { return "test" }
func (testBlitter) PixelsPerCell() (int, int) { return 2, 2 }
func (testBlitter) SupportsColor() bool       { return true }
func (testBlitter) SupportsAlpha() bool       { return true }
func (testBlitter) BlitCell(*PixelBuffer, int, int) (rune, backend.Style) {
	return ' ', backend.DefaultStyle()
}

func TestPixelBufferBasics(t *testing.T) {
	buf := NewPixelBuffer(2, 3, testBlitter{})
	w, h := buf.Size()
	if w != 4 || h != 6 {
		t.Fatalf("size = %dx%d, want 4x6", w, h)
	}
	buf.SetPixel(1, 1, backend.ColorRGB(10, 20, 30))
	pix := buf.Get(1, 1)
	if !pix.Set || pix.Alpha != 1 {
		t.Fatalf("unexpected pixel")
	}
	buf.Clear()
	if buf.Get(1, 1).Set {
		t.Fatalf("expected cleared pixel")
	}
}

func TestPixelBufferBlendAlpha(t *testing.T) {
	buf := NewPixelBuffer(1, 1, nil)
	buf.Blend(0, 0, backend.ColorRGB(255, 0, 0), 0)
	if buf.Get(0, 0).Set {
		t.Fatalf("expected no blend when alpha=0")
	}
	buf.Blend(0, 0, backend.ColorRGB(255, 0, 0), 2)
	pix := buf.Get(0, 0)
	if !pix.Set || pix.Color != backend.ColorRGB(255, 0, 0) {
		t.Fatalf("expected blended pixel")
	}

	buf.Blend(0, 0, backend.ColorRGB(0, 0, 255), 0.5)
	pix = buf.Get(0, 0)
	if pix.Alpha <= 0 {
		t.Fatalf("expected alpha to increase")
	}

	base := backend.ColorDefault
	top := backend.ColorRed
	if blendColors(base, top, 0.4) != base {
		t.Fatalf("expected base color for low alpha")
	}
	if blendColors(base, top, 0.6) != top {
		t.Fatalf("expected top color for high alpha")
	}
}
