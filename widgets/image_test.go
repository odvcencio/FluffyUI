package widgets

import (
	"image"
	"image/color"
	"testing"

	"m31labs.dev/fluffyui/accessibility"
	"m31labs.dev/fluffyui/runtime"
)

// newTestImage creates a small solid-color image for testing.
func newTestImage(w, h int, c color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

func TestImage_Basic(t *testing.T) {
	img := newTestImage(10, 10, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	w := NewImage(img)
	if w == nil {
		t.Fatal("expected non-nil Image")
	}

	if w.StyleType() != "Image" {
		t.Fatalf("expected StyleType 'Image', got %q", w.StyleType())
	}

	// Measure: 10 pixels wide = 10 cells, 10 pixels tall = 5 cell rows (half-block)
	size := w.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 24})
	if size.Width != 10 {
		t.Fatalf("expected width 10, got %d", size.Width)
	}
	if size.Height != 5 {
		t.Fatalf("expected height 5 for 10px tall image, got %d", size.Height)
	}
}

func TestImage_SetProtocol(t *testing.T) {
	w := NewImage(nil)
	if w.Protocol() != ImageProtocolAuto {
		t.Fatalf("expected default protocol Auto, got %d", w.Protocol())
	}

	w.SetProtocol(ImageProtocolKitty)
	if w.Protocol() != ImageProtocolKitty {
		t.Fatalf("expected protocol Kitty, got %d", w.Protocol())
	}

	w.SetProtocol(ImageProtocolSixel)
	if w.Protocol() != ImageProtocolSixel {
		t.Fatalf("expected protocol Sixel, got %d", w.Protocol())
	}

	w.SetProtocol(ImageProtocolITerm2)
	if w.Protocol() != ImageProtocolITerm2 {
		t.Fatalf("expected protocol iTerm2, got %d", w.Protocol())
	}

	w.SetProtocol(ImageProtocolHalfBlock)
	if w.Protocol() != ImageProtocolHalfBlock {
		t.Fatalf("expected protocol HalfBlock, got %d", w.Protocol())
	}
}

func TestImage_SetFit(t *testing.T) {
	w := NewImage(nil)
	if w.Fit() != ImageFitContain {
		t.Fatalf("expected default fit Contain, got %d", w.Fit())
	}

	w.SetFit(ImageFitCover)
	if w.Fit() != ImageFitCover {
		t.Fatalf("expected fit Cover, got %d", w.Fit())
	}

	w.SetFit(ImageFitFill)
	if w.Fit() != ImageFitFill {
		t.Fatalf("expected fit Fill, got %d", w.Fit())
	}

	w.SetFit(ImageFitScaleDown)
	if w.Fit() != ImageFitScaleDown {
		t.Fatalf("expected fit ScaleDown, got %d", w.Fit())
	}
}

func TestImage_FallbackRendering(t *testing.T) {
	img := newTestImage(4, 4, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	w := NewImage(img)

	// Measure, layout, and render.
	constraints := runtime.Constraints{MaxWidth: 10, MaxHeight: 10}
	size := w.Measure(constraints)
	w.Layout(runtime.Rect{X: 0, Y: 0, Width: size.Width, Height: size.Height})

	buf := runtime.NewBuffer(size.Width, size.Height)
	ctx := runtime.RenderContext{Buffer: buf}
	w.Render(ctx)

	// Check that at least one cell was written (non-zero rune or non-default style).
	found := false
	for y := 0; y < size.Height; y++ {
		for x := 0; x < size.Width; x++ {
			cell := buf.Get(x, y)
			if cell.Rune != 0 {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Fatal("expected non-empty output from half-block rendering")
	}
}

func TestImage_NilImage(t *testing.T) {
	w := NewImage(nil)
	if w == nil {
		t.Fatal("expected non-nil Image widget even with nil source")
	}

	// Measure should return a minimal size.
	size := w.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 24})
	if size.Width < 1 || size.Height < 1 {
		t.Fatalf("expected at least 1x1 for nil image, got %dx%d", size.Width, size.Height)
	}

	// Render should not panic.
	w.Layout(runtime.Rect{X: 0, Y: 0, Width: 10, Height: 5})
	buf := runtime.NewBuffer(10, 5)
	ctx := runtime.RenderContext{Buffer: buf}
	w.Render(ctx)
}

func TestImage_Accessibility(t *testing.T) {
	img := newTestImage(4, 4, color.RGBA{R: 0, G: 255, B: 0, A: 255})
	w := NewImage(img)

	if w.AccessibleRole() != accessibility.RoleImage {
		t.Fatalf("expected role %q, got %q", accessibility.RoleImage, w.AccessibleRole())
	}
	if w.AccessibleLabel() != "Image" {
		t.Fatalf("expected label 'Image', got %q", w.AccessibleLabel())
	}

	w.SetLabel("Photo of a cat")
	if w.AccessibleLabel() != "Photo of a cat" {
		t.Fatalf("expected label 'Photo of a cat', got %q", w.AccessibleLabel())
	}
	if w.Label() != "Photo of a cat" {
		t.Fatalf("expected Label() to return 'Photo of a cat', got %q", w.Label())
	}
}

func TestImage_SetImage(t *testing.T) {
	w := NewImage(nil)

	// Replace with a real image.
	img := newTestImage(8, 6, color.RGBA{R: 0, G: 0, B: 255, A: 255})
	w.SetImage(img)

	size := w.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 24})
	if size.Width != 8 {
		t.Fatalf("expected width 8 after SetImage, got %d", size.Width)
	}
	if size.Height != 3 {
		t.Fatalf("expected height 3 for 6px tall image, got %d", size.Height)
	}
}

func TestImage_HandleMessage(t *testing.T) {
	w := NewImage(nil)
	result := w.HandleMessage(runtime.KeyMsg{})
	if result.Handled {
		t.Fatal("expected Image to return unhandled")
	}
}

func TestImage_BindUnbind(t *testing.T) {
	w := NewImage(nil)
	w.Bind(runtime.Services{})
	w.Unbind()
	// Should not panic.
}

func TestImage_MeasureFitModes(t *testing.T) {
	// 20x10 image: 20 cells wide, 5 cells tall.
	img := newTestImage(20, 10, color.White)

	tests := []struct {
		name string
		fit  ImageFit
		maxW int
		maxH int
		wantW int
		wantH int
	}{
		{"contain fits", ImageFitContain, 10, 10, 10, 3},
		{"contain no-op", ImageFitContain, 80, 80, 20, 5},
		{"fill stretches", ImageFitFill, 10, 10, 10, 10},
		{"scale-down small", ImageFitScaleDown, 10, 10, 10, 3},
		{"scale-down no upscale", ImageFitScaleDown, 80, 80, 20, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewImage(img)
			w.SetFit(tt.fit)
			size := w.Measure(runtime.Constraints{MaxWidth: tt.maxW, MaxHeight: tt.maxH})
			if size.Width != tt.wantW {
				t.Errorf("width: got %d, want %d", size.Width, tt.wantW)
			}
			if size.Height != tt.wantH {
				t.Errorf("height: got %d, want %d", size.Height, tt.wantH)
			}
		})
	}
}

func TestImage_NilReceiver(t *testing.T) {
	var w *Image
	// None of these should panic.
	w.SetFit(ImageFitCover)
	w.SetProtocol(ImageProtocolKitty)
	w.SetImage(nil)
	w.SetLabel("test")
	w.Bind(runtime.Services{})
	w.Unbind()
	_ = w.Fit()
	_ = w.Protocol()
	_ = w.Label()
}
