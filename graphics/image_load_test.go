package graphics

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestScaleImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	img.Set(1, 0, color.RGBA{G: 255, A: 255})
	img.Set(0, 1, color.RGBA{B: 255, A: 255})
	img.Set(1, 1, color.RGBA{R: 255, G: 255, A: 255})

	scaled := ScaleImage(img, 1, 1)
	if scaled == nil {
		t.Fatalf("scaled image is nil")
	}
	if scaled == img {
		t.Fatalf("expected scaled image, got original")
	}
	if got := scaled.At(0, 0); got != (color.RGBA{R: 255, A: 255}) {
		t.Fatalf("pixel(0,0) = %#v, want red", got)
	}

	unscaled := ScaleImage(img, 4, 4)
	if unscaled != img {
		t.Fatalf("expected original image when scale >= 1")
	}
}

func TestLoadImage(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tiny.png")
	out, err := os.Create(path)
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	src := image.NewRGBA(image.Rect(0, 0, 1, 1))
	src.Set(0, 0, color.RGBA{R: 10, G: 20, B: 30, A: 255})
	if err := png.Encode(out, src); err != nil {
		_ = out.Close()
		t.Fatalf("encode png: %v", err)
	}
	if err := out.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}

	img, err := LoadImage(path)
	if err != nil {
		t.Fatalf("LoadImage error: %v", err)
	}
	if img.Bounds().Dx() != 1 || img.Bounds().Dy() != 1 {
		t.Fatalf("size = %dx%d, want 1x1", img.Bounds().Dx(), img.Bounds().Dy())
	}
	if got := img.At(0, 0); got != (color.RGBA{R: 10, G: 20, B: 30, A: 255}) {
		t.Fatalf("pixel(0,0) = %#v, want RGB(10,20,30)", got)
	}
}
