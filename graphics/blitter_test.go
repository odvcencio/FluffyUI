package graphics

import (
	"testing"

	"github.com/odvcencio/fluffyui/backend"
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

func TestKittyBlitter(t *testing.T) {
	blitter := &KittyBlitter{}

	// Test PixelsPerCell returns high resolution (10x20)
	w, h := blitter.PixelsPerCell()
	if w != 10 || h != 20 {
		t.Errorf("PixelsPerCell = (%d, %d), want (10, 20)", w, h)
	}

	// Test capabilities
	if !blitter.SupportsColor() {
		t.Error("SupportsColor should return true")
	}
	if !blitter.SupportsAlpha() {
		t.Error("SupportsAlpha should return true for Kitty")
	}
	if blitter.Name() != "kitty" {
		t.Errorf("Name = %q, want %q", blitter.Name(), "kitty")
	}

	// Test BlitCell returns space (actual rendering uses BuildImage)
	buf := NewPixelBuffer(1, 1, blitter)
	buf.SetPixel(0, 0, backend.ColorRed)
	r, style := blitter.BlitCell(buf, 0, 0)
	if r != ' ' {
		t.Errorf("BlitCell rune = %q, want space", r)
	}
	if style != backend.DefaultStyle() {
		t.Error("BlitCell style should be default")
	}

	// Test implements ImageBlitter
	var _ ImageBlitter = blitter
}

func TestKittyBlitterBuildImage(t *testing.T) {
	blitter := &KittyBlitter{}
	buf := NewPixelBuffer(2, 2, blitter)
	buf.SetPixel(0, 0, backend.ColorRed)
	buf.SetPixel(1, 1, backend.ColorBlue)

	img := blitter.BuildImage(buf, 2, 2)
	if img.Protocol != backend.ImageProtocolKitty {
		t.Errorf("Protocol = %v, want Kitty", img.Protocol)
	}
	if img.Format != backend.ImageFormatRGBA {
		t.Errorf("Format = %v, want RGBA", img.Format)
	}
	w, h := buf.Size()
	if img.Width != w || img.Height != h {
		t.Errorf("Image size = (%d, %d), want (%d, %d)", img.Width, img.Height, w, h)
	}
	if img.CellWidth != 2 || img.CellHeight != 2 {
		t.Errorf("Cell size = (%d, %d), want (2, 2)", img.CellWidth, img.CellHeight)
	}
	if len(img.Pixels) != w*h*4 {
		t.Errorf("Pixels length = %d, want %d", len(img.Pixels), w*h*4)
	}
}

func TestSixelBlitter(t *testing.T) {
	blitter := &SixelBlitter{}

	// Test PixelsPerCell returns sixel resolution (10x12)
	w, h := blitter.PixelsPerCell()
	if w != 10 || h != 12 {
		t.Errorf("PixelsPerCell = (%d, %d), want (10, 12)", w, h)
	}

	// Test capabilities
	if !blitter.SupportsColor() {
		t.Error("SupportsColor should return true")
	}
	if blitter.SupportsAlpha() {
		t.Error("SupportsAlpha should return false for Sixel")
	}
	if blitter.Name() != "sixel" {
		t.Errorf("Name = %q, want %q", blitter.Name(), "sixel")
	}

	// Test BlitCell returns space (actual rendering uses BuildImage)
	buf := NewPixelBuffer(1, 1, blitter)
	buf.SetPixel(0, 0, backend.ColorRed)
	r, style := blitter.BlitCell(buf, 0, 0)
	if r != ' ' {
		t.Errorf("BlitCell rune = %q, want space", r)
	}
	if style != backend.DefaultStyle() {
		t.Error("BlitCell style should be default")
	}

	// Test implements ImageBlitter
	var _ ImageBlitter = blitter
}

func TestSixelBlitterBuildImage(t *testing.T) {
	blitter := &SixelBlitter{}
	buf := NewPixelBuffer(2, 2, blitter)
	buf.SetPixel(0, 0, backend.ColorRed)
	buf.SetPixel(1, 1, backend.ColorGreen)

	img := blitter.BuildImage(buf, 2, 2)
	if img.Protocol != backend.ImageProtocolSixel {
		t.Errorf("Protocol = %v, want Sixel", img.Protocol)
	}
	if img.Format != backend.ImageFormatRGBA {
		t.Errorf("Format = %v, want RGBA", img.Format)
	}
	w, h := buf.Size()
	if img.Width != w || img.Height != h {
		t.Errorf("Image size = (%d, %d), want (%d, %d)", img.Width, img.Height, w, h)
	}
	if len(img.Pixels) != w*h*4 {
		t.Errorf("Pixels length = %d, want %d", len(img.Pixels), w*h*4)
	}
}
