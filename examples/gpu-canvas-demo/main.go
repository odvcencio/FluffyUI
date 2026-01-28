package main

import (
	"context"
	"fmt"
	"image/color"
	"math"
	"os"
	"strings"
	"time"

	"github.com/odvcencio/fluffyui/examples/internal/demo"
	"github.com/odvcencio/fluffyui/gpu"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/widgets"
)

func main() {
	view := NewGPUCanvasDemo()
	bundle, err := demo.NewApp(view, demo.Options{TickRate: time.Second / 60})
	if err != nil {
		fmt.Fprintf(os.Stderr, "app init failed: %v\n", err)
		os.Exit(1)
	}
	if err := bundle.App.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
		os.Exit(1)
	}
}

type GPUCanvasDemo struct {
	*widgets.GPUCanvasWidget
	phase float32
}

func NewGPUCanvasDemo() *GPUCanvasDemo {
	demo := &GPUCanvasDemo{}
	demo.GPUCanvasWidget = widgets.NewGPUCanvasWidget(demo.draw)
	if backend := strings.TrimSpace(os.Getenv("FLUFFYUI_GPU_BACKEND")); backend != "" {
		switch strings.ToLower(backend) {
		case "opengl", "gl":
			demo.GPUCanvasWidget.WithBackend(gpu.BackendOpenGL)
		case "metal":
			demo.GPUCanvasWidget.WithBackend(gpu.BackendMetal)
		case "webgl":
			demo.GPUCanvasWidget.WithBackend(gpu.BackendWebGL)
		case "software", "cpu":
			demo.GPUCanvasWidget.WithBackend(gpu.BackendSoftware)
		}
	}
	return demo
}

func (d *GPUCanvasDemo) HandleMessage(msg runtime.Message) runtime.HandleResult {
	switch m := msg.(type) {
	case runtime.TickMsg:
		d.phase += 0.02
		if d.phase > math.Pi*2 {
			d.phase -= math.Pi * 2
		}
		d.Invalidate()
		return runtime.Handled()
	case runtime.KeyMsg:
		if m.Rune == 'q' {
			return runtime.WithCommand(runtime.Quit{})
		}
	}
	return runtime.Unhandled()
}

func (d *GPUCanvasDemo) draw(canvas *gpu.GPUCanvas) {
	w, h := canvas.Size()
	if w <= 0 || h <= 0 {
		return
	}
	canvas.SetStrokeWidth(1)
	for y := 0; y < h; y++ {
		t := float32(y) / float32(max(1, h-1))
		col := lerpColor(color.RGBA{R: 10, G: 12, B: 20, A: 255}, color.RGBA{R: 42, G: 32, B: 68, A: 255}, t)
		canvas.SetStrokeColor(col)
		canvas.DrawLine(0, float32(y), float32(w-1), float32(y))
	}

	cx := float32(w)/2 + float32(math.Sin(float64(d.phase)))*float32(w)/6
	cy := float32(h)/2 + float32(math.Cos(float64(d.phase*0.7)))*float32(h)/6

	canvas.PushLayer()
	canvas.SetFillColor(color.RGBA{R: 255, G: 120, B: 90, A: 255})
	canvas.FillCircle(cx, cy, float32(min(w, h))/6)
	canvas.PopLayer(gpu.GlowEffect{Radius: 12, Intensity: 0.8, Color: color.RGBA{R: 255, G: 120, B: 90, A: 200}})

	canvas.SetStrokeColor(color.RGBA{R: 230, G: 220, B: 255, A: 255})
	canvas.SetStrokeWidth(2)
	canvas.StrokeRect(6, 6, float32(w-12), float32(h-12))

	canvas.SetFillColor(color.RGBA{R: 240, G: 240, B: 255, A: 255})
	canvas.FillRoundedRect(12, float32(h)-28, float32(w)-24, 16, 6)

	canvas.SetFillColor(color.RGBA{R: 20, G: 20, B: 30, A: 255})
	canvas.DrawText("GPU CANVAS", 18, float32(h)-24, nil)

	canvas.ApplyEffect(gpu.VignetteEffect{Radius: 0.7, Softness: 0.4})
}

func lerpColor(a, b color.RGBA, t float32) color.RGBA {
	return color.RGBA{
		R: uint8(float32(a.R) + (float32(b.R)-float32(a.R))*t),
		G: uint8(float32(a.G) + (float32(b.G)-float32(a.G))*t),
		B: uint8(float32(a.B) + (float32(b.B)-float32(a.B))*t),
		A: uint8(float32(a.A) + (float32(b.A)-float32(a.A))*t),
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
