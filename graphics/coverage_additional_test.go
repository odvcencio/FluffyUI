package graphics

import (
	"testing"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/terminal"
)

func TestBlitterSelectionAndBasics(t *testing.T) {
	caps := terminal.Capabilities{Kitty: true}
	if BestBlitter(&caps).Name() == "" {
		t.Fatalf("expected kitty blitter")
	}
	caps = terminal.Capabilities{Sixel: true, Unicode: true}
	if BestBlitter(&caps).Name() == "" {
		t.Fatalf("expected blitter")
	}
	caps = terminal.Capabilities{Unicode: true}
	if BestBlitter(&caps).Name() == "" {
		t.Fatalf("expected unicode blitter")
	}

	_ = NewBlitter(BlitterAuto)
	_ = NewBlitter(BlitterHalfBlock)
	_ = NewBlitter(BlitterQuadrant)
	_ = NewBlitter(BlitterSextant)
	_ = NewBlitter(BlitterBraille)
	_ = NewBlitter(BlitterKitty)
	_ = NewBlitter(BlitterSixel)
	_ = NewBlitter(BlitterASCII)
	_ = NewBlitter(BlitterType(999))

	blitters := []Blitter{
		&ASCIIBlitter{},
		&HalfBlockBlitter{},
		&QuadrantBlitter{},
		&SextantBlitter{},
		&BrailleBlitter{},
	}

	for _, blitter := range blitters {
		w, h := blitter.PixelsPerCell()
		if w <= 0 || h <= 0 {
			t.Fatalf("invalid pixels per cell")
		}
		_ = blitter.SupportsColor()
		_ = blitter.SupportsAlpha()
		if blitter.Name() == "" {
			t.Fatalf("expected name")
		}
		pixels := NewPixelBuffer(1, 1, blitter)
		pixels.SetPixel(0, 0, backend.ColorRGB(255, 0, 0))
		_, _ = blitter.BlitCell(pixels, 0, 0)
	}
}

func TestCanvasTransformsAndPixels(t *testing.T) {
	canvas := NewCanvas(2, 2)
	canvas.Translate(1, 1)
	canvas.Rotate(0.1)
	canvas.Scale(1.2, 0.8)
	canvas.SetLineWidth(2)
	canvas.SetPixel(0, 0, backend.ColorRGB(0, 255, 0))
	canvas.SetFillColor(backend.ColorRGB(255, 0, 0))
	canvas.FillRect(0, 0, 1, 1)
}

func TestGPUBlitterEncoders(t *testing.T) {
	blitter := NewGPUBlitter(nil, nil, 2, 3)
	blitter.SetCanvas(nil)
	blitter.SetEncoder(nil)
	blitter.SetCellSize(1, 1)
	if img := blitter.Image(); img.Width != 0 {
		t.Fatalf("expected empty image")
	}
	if data := blitter.Encode(); data != nil {
		t.Fatalf("expected nil encode")
	}

	kitty := KittyEncoder{}
	if !kitty.SupportsAlpha() || kitty.Name() == "" {
		t.Fatalf("kitty encoder metadata")
	}
	_ = kitty.Encode([]byte{0xff, 0, 0, 0xff}, 1, 1)

	sixel := SixelEncoder{}
	if sixel.SupportsAlpha() || sixel.Name() == "" {
		t.Fatalf("sixel encoder metadata")
	}
	_ = sixel.Encode([]byte{0xff, 0, 0, 0xff}, 1, 1)
}

func TestDominantColor(t *testing.T) {
	colors := []Color{backend.ColorDefault, backend.ColorRed, backend.ColorRed, backend.ColorBlue}
	if dominantColor(colors) != backend.ColorRed {
		t.Fatalf("expected dominant red")
	}
	if dominantColor(nil) != backend.ColorDefault {
		t.Fatalf("expected default color")
	}
}
