package gpu

import (
	"image"
	"image/color"
	"testing"
)

func TestFloatToByteAndRGBA(t *testing.T) {
	if floatToByte(-1) != 0 {
		t.Fatalf("expected clamp to 0")
	}
	if floatToByte(2) != 255 {
		t.Fatalf("expected clamp to 255")
	}
	if floatToByte(0.5) != 128 {
		t.Fatalf("expected mid conversion")
	}
	col := rgbaFromFloats(1, 0, 0, 1)
	if col != (color.RGBA{R: 255, A: 255}) {
		t.Fatalf("rgbaFromFloats = %#v", col)
	}
}

func TestPixelHelpers(t *testing.T) {
	pixels := make([]byte, 4*4*4)
	clearPixels(pixels, 4, 4, color.RGBA{R: 1, G: 2, B: 3, A: 4})
	if pixels[0] != 1 || pixels[1] != 2 || pixels[2] != 3 || pixels[3] != 4 {
		t.Fatalf("clearPixels did not set first pixel")
	}

	setPixel(pixels, 4, 4, 1, 1, color.RGBA{R: 10, G: 20, B: 30, A: 40})
	idx := (1*4 + 1) * 4
	if pixels[idx] != 10 || pixels[idx+3] != 40 {
		t.Fatalf("setPixel failed")
	}

	blendPixel(pixels, 4, 4, 1, 1, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	if pixels[idx] == 0 {
		t.Fatalf("blendPixel failed")
	}

	_, _, _, outA := blendRGBA(0, 0, 0, 0, 0, 0, 0, 0)
	if outA != 0 {
		t.Fatalf("expected zero alpha")
	}
}

func TestScaleCropFlipPixels(t *testing.T) {
	src := make([]byte, 2*2*4)
	for i := range src {
		src[i] = byte(i)
	}
	scaled := scalePixels(src, 2, 2, 4, 4)
	if len(scaled) != 4*4*4 {
		t.Fatalf("unexpected scaled size")
	}

	cropped := cropPixels(src, 2, 2, image.Rect(0, 0, 1, 1))
	if len(cropped) != 4 {
		t.Fatalf("unexpected crop size")
	}
	if cropPixels(src, 2, 2, image.Rect(5, 5, 6, 6)) != nil {
		t.Fatalf("expected nil crop for empty rect")
	}

	flipPixelsVertical(src, 2, 2)
}

func TestSoftwareDriverBasics(t *testing.T) {
	d := newSoftwareDriver()
	if d.Backend() != BackendSoftware {
		t.Fatalf("expected software backend")
	}
	if err := d.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}
	tex, err := d.NewTexture(2, 2)
	if err != nil {
		t.Fatalf("new texture: %v", err)
	}
	if _, err := d.NewTexture(0, 0); err == nil {
		t.Fatalf("expected error for invalid texture size")
	}
	fb, err := d.NewFramebuffer(2, 2)
	if err != nil {
		t.Fatalf("new framebuffer: %v", err)
	}
	d.Clear(0.1, 0.2, 0.3, 1)
	pixels, err := d.ReadPixels(fb, image.Rectangle{})
	if err != nil || len(pixels) == 0 {
		t.Fatalf("read pixels failed")
	}

	out, w, h, err := d.ReadTexturePixels(tex, image.Rect(0, 0, 1, 1))
	if err != nil || w != 1 || h != 1 || len(out) != 4 {
		t.Fatalf("read texture pixels failed")
	}
	if d.MaxTextureSize() == 0 {
		t.Fatalf("expected non-zero max texture size")
	}
}

func TestNewDriverDefault(t *testing.T) {
	d, err := NewDriver(Backend(99))
	if err != nil {
		t.Fatalf("NewDriver error: %v", err)
	}
	if d.Backend() != BackendSoftware {
		t.Fatalf("expected fallback to software")
	}
}
