package effects

import (
	"math"
	"testing"

	"github.com/odvcencio/fluffyui/animation"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/graphics"
)

func TestLerpColorRGB(t *testing.T) {
	start := backend.ColorRGB(10, 20, 30)
	end := backend.ColorRGB(110, 120, 130)
	got := lerpColor(start, end, 0.5)
	want := backend.ColorRGB(60, 70, 80)
	if got != want {
		t.Fatalf("unexpected lerp result: got %v want %v", got, want)
	}
}

func TestLerpColorNonRGB(t *testing.T) {
	start := backend.ColorRed
	end := backend.ColorBlue
	if got := lerpColor(start, end, 0.4); got != start {
		t.Fatalf("expected start color for t<0.5, got %v", got)
	}
	if got := lerpColor(start, end, 0.5); got != end {
		t.Fatalf("expected end color for t>=0.5, got %v", got)
	}
}

func TestLinearGradient(t *testing.T) {
	canvas := graphics.NewCanvasWithBlitter(4, 1, &graphics.ASCIIBlitter{})
	start := backend.ColorRGB(10, 20, 30)
	end := backend.ColorRGB(110, 120, 130)

	LinearGradient(canvas, 0, 0, 4, 1, start, end, 0)

	p0 := canvas.GetPixel(0, 0)
	if !p0.Set || p0.Color != start {
		t.Fatalf("expected start color at (0,0), got %+v", p0)
	}
	p3 := canvas.GetPixel(3, 0)
	want := lerpColor(start, end, 0.75)
	if !p3.Set || p3.Color != want {
		t.Fatalf("unexpected gradient color at (3,0): got %+v want %v", p3, want)
	}
}

func TestRadialGradient(t *testing.T) {
	canvas := graphics.NewCanvasWithBlitter(5, 5, &graphics.ASCIIBlitter{})
	center := backend.ColorRGB(1, 2, 3)
	edge := backend.ColorRGB(200, 201, 202)

	RadialGradient(canvas, 2, 2, 2, center, edge)

	pc := canvas.GetPixel(2, 2)
	if !pc.Set || pc.Color != center {
		t.Fatalf("expected center color at (2,2), got %+v", pc)
	}
	pe := canvas.GetPixel(4, 2)
	if !pe.Set || pe.Color != edge {
		t.Fatalf("expected edge color at (4,2), got %+v", pe)
	}
}

func TestGlow(t *testing.T) {
	canvas := graphics.NewCanvasWithBlitter(7, 7, &graphics.ASCIIBlitter{})
	color := backend.ColorRGB(255, 0, 0)

	Glow(canvas, 3, 3, 2, color, 1.0)

	p := canvas.GetPixel(3, 3)
	if !p.Set || p.Color != color {
		t.Fatalf("expected glow at center, got %+v", p)
	}
	if math.Abs(float64(p.Alpha-1.0)) > 0.01 {
		t.Fatalf("unexpected alpha: %v", p.Alpha)
	}
}

func TestRipple(t *testing.T) {
	canvas := graphics.NewCanvasWithBlitter(9, 9, &graphics.ASCIIBlitter{})
	color := backend.ColorRGB(0, 255, 0)

	Ripple(canvas, 4, 4, 2, 4, color, 0.6)

	p := canvas.GetPixel(6, 4)
	if !p.Set || p.Color != color {
		t.Fatalf("expected ripple pixel at (6,4), got %+v", p)
	}
	if math.Abs(float64(p.Alpha-0.5)) > 0.05 {
		t.Fatalf("unexpected ripple alpha: %v", p.Alpha)
	}
}

func TestShadow(t *testing.T) {
	canvas := graphics.NewCanvasWithBlitter(8, 8, &graphics.ASCIIBlitter{})
	color := backend.ColorRGB(20, 30, 40)

	Shadow(canvas, 2, 2, 2, 2, 0, 0, 1, color)

	p := canvas.GetPixel(2, 2)
	if !p.Set || p.Color != color {
		t.Fatalf("expected shadow pixel at (2,2), got %+v", p)
	}
	if math.Abs(float64(p.Alpha-0.3)) > 0.05 {
		t.Fatalf("unexpected shadow alpha: %v", p.Alpha)
	}
}

func TestShimmer(t *testing.T) {
	canvas := graphics.NewCanvasWithBlitter(10, 2, &graphics.ASCIIBlitter{})
	color := backend.ColorRGB(100, 150, 200)

	Shimmer(canvas, 0, 0, 8, 2, 0.2, color)

	p := canvas.GetPixel(1, 0)
	if !p.Set || p.Color != color {
		t.Fatalf("expected shimmer pixel at (1,0), got %+v", p)
	}
	if p.Alpha <= 0 {
		t.Fatalf("expected shimmer alpha > 0")
	}
}

func TestConfettiAndSparkleCounts(t *testing.T) {
	ps := animation.NewParticleSystem(50)

	Confetti(ps, 5, 5, 10)
	if ps.ParticleCount() != 10 {
		t.Fatalf("unexpected confetti count: %d", ps.ParticleCount())
	}

	ps.Clear()
	Sparkle(ps, 0, 0, 10, 10, 0.1)
	if ps.ParticleCount() != 10 {
		t.Fatalf("unexpected sparkle count: %d", ps.ParticleCount())
	}
}
