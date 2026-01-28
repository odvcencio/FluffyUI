package graphics

import (
	"testing"

	"github.com/odvcencio/fluffyui/backend"
)

func TestCanvasDrawLine(t *testing.T) {
	canvas := NewCanvasWithBlitter(2, 1, &HalfBlockBlitter{})
	canvas.SetStrokeColor(backend.ColorRed)
	canvas.DrawLine(0, 0, 1, 0)

	if !canvas.GetPixel(0, 0).Set {
		t.Fatalf("expected pixel at (0,0) to be set")
	}
	if !canvas.GetPixel(1, 0).Set {
		t.Fatalf("expected pixel at (1,0) to be set")
	}
}

func TestCanvasFillRect(t *testing.T) {
	canvas := NewCanvasWithBlitter(1, 1, &HalfBlockBlitter{})
	canvas.SetFillColor(backend.ColorGreen)
	canvas.FillRect(0, 0, 1, 2)

	if !canvas.GetPixel(0, 0).Set {
		t.Fatalf("expected pixel at (0,0) to be set")
	}
	if !canvas.GetPixel(0, 1).Set {
		t.Fatalf("expected pixel at (0,1) to be set")
	}
}

func TestCanvasDrawLineAA(t *testing.T) {
	canvas := NewCanvasWithBlitter(2, 2, &HalfBlockBlitter{})
	canvas.SetStrokeColor(backend.ColorRed)
	canvas.DrawLineAA(0, 0, 1, 1)

	if !canvas.GetPixel(0, 0).Set {
		t.Fatalf("expected pixel at (0,0) to be set")
	}
}

func TestCanvasDrawEllipse(t *testing.T) {
	canvas := NewCanvasWithBlitter(6, 4, &HalfBlockBlitter{})
	canvas.SetStrokeColor(backend.ColorBlue)
	canvas.DrawEllipse(3, 3, 2, 1)

	if !canvas.GetPixel(1, 3).Set {
		t.Fatalf("expected ellipse pixel at (1,3) to be set")
	}
	if !canvas.GetPixel(5, 3).Set {
		t.Fatalf("expected ellipse pixel at (5,3) to be set")
	}
}

func TestCanvasDrawBezier(t *testing.T) {
	canvas := NewCanvasWithBlitter(8, 4, &HalfBlockBlitter{})
	canvas.SetStrokeColor(backend.ColorYellow)
	canvas.DrawBezier(
		Point{X: 0, Y: 0},
		Point{X: 2, Y: 3},
		Point{X: 5, Y: 3},
		Point{X: 7, Y: 0},
	)

	if !canvas.GetPixel(0, 0).Set {
		t.Fatalf("expected bezier start pixel to be set")
	}
	if !canvas.GetPixel(7, 0).Set {
		t.Fatalf("expected bezier end pixel to be set")
	}
}

func TestCanvasFillPolygon(t *testing.T) {
	canvas := NewCanvasWithBlitter(5, 5, &HalfBlockBlitter{})
	canvas.SetFillColor(backend.ColorGreen)
	canvas.FillPolygon([]Point{{X: 1, Y: 1}, {X: 3, Y: 1}, {X: 2, Y: 3}})

	if !canvas.GetPixel(2, 2).Set {
		t.Fatalf("expected filled polygon pixel to be set")
	}
}

func TestCanvasPathFill(t *testing.T) {
	canvas := NewCanvasWithBlitter(5, 5, &HalfBlockBlitter{})
	canvas.SetFillColor(backend.ColorCyan)
	canvas.BeginPath()
	canvas.MoveTo(1, 1)
	canvas.LineTo(3, 1)
	canvas.LineTo(3, 3)
	canvas.LineTo(1, 3)
	canvas.ClosePath()
	canvas.Fill()

	if !canvas.GetPixel(2, 2).Set {
		t.Fatalf("expected path fill pixel to be set")
	}
}
