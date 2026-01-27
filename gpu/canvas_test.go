package gpu

import (
	"image/color"
	"testing"
)

func TestGPUCanvasFillRect(t *testing.T) {
	canvas, err := NewGPUCanvas(8, 8)
	if err != nil {
		t.Fatalf("new canvas: %v", err)
	}
	defer canvas.Dispose()
	canvas.Clear(color.RGBA{})
	canvas.SetFillColor(color.RGBA{R: 255, G: 0, B: 0, A: 255})
	canvas.FillRect(0, 0, 8, 8)
	pixels := canvas.End()
	if len(pixels) == 0 {
		t.Fatalf("expected pixels")
	}
	idx := (4*8 + 4) * 4
	if pixels[idx] != 255 || pixels[idx+1] != 0 || pixels[idx+2] != 0 {
		t.Fatalf("expected red pixel, got %v %v %v", pixels[idx], pixels[idx+1], pixels[idx+2])
	}
}

func TestGPUCanvasBlurEffect(t *testing.T) {
	canvas, err := NewGPUCanvas(5, 5)
	if err != nil {
		t.Fatalf("new canvas: %v", err)
	}
	defer canvas.Dispose()
	canvas.Clear(color.RGBA{})
	canvas.SetFillColor(color.RGBA{R: 255, G: 255, B: 255, A: 255})
	canvas.FillRect(2, 2, 1, 1)
	canvas.ApplyEffect(BlurEffect{Radius: 1})
	pixels := canvas.End()
	idx := (2*5 + 3) * 4
	if pixels[idx+3] == 0 {
		t.Fatalf("expected blurred neighbor to have alpha")
	}
}
