//go:build !js

package testing

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/runtime"
)

type rendererWidget struct {
	bounds runtime.Rect
	text   string
}

func (w *rendererWidget) Measure(runtime.Constraints) runtime.Size {
	return runtime.Size{Width: len(w.text), Height: 1}
}

func (w *rendererWidget) Layout(bounds runtime.Rect) {
	w.bounds = bounds
}

func (w *rendererWidget) Render(ctx runtime.RenderContext) {
	for i, char := range w.text {
		ctx.Buffer.Set(w.bounds.X+i, w.bounds.Y, char, backend.DefaultStyle())
	}
}

func (w *rendererWidget) HandleMessage(msg runtime.Message) runtime.HandleResult {
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	w.text += string(key.Rune)
	return runtime.Handled()
}

func TestNewTestRenderer_CapturesInitialFrameAndStats(t *testing.T) {
	renderer, err := NewTestRenderer(t.Context(), &rendererWidget{text: "hello"}, 12, 3)
	if err != nil {
		t.Fatalf("NewTestRenderer() error = %v", err)
	}
	t.Cleanup(func() {
		if err := renderer.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	frame := renderer.LatestFrame()
	if frame.Stats.Frame != 1 {
		t.Fatalf("initial frame = %d, want 1", frame.Stats.Frame)
	}
	if !strings.HasPrefix(frame.Screen, "hello") {
		t.Fatalf("initial screen = %q, want hello prefix", frame.Screen)
	}
	if got := renderer.Stats().Frames; got != 1 {
		t.Fatalf("stats frames = %d, want 1", got)
	}
	if char, _, _ := renderer.CaptureCell(0, 0); char != 'h' {
		t.Fatalf("cell rune = %q, want h", char)
	}
}

func TestNewTestRendererWithConfig_PreservesObserver(t *testing.T) {
	observed := make(chan runtime.RenderStats, 1)
	renderer, err := NewTestRendererWithConfig(t.Context(), runtime.AppConfig{
		Root: &rendererWidget{text: "configured"},
		RenderObserver: runtime.RenderObserverFunc(func(stats runtime.RenderStats) {
			observed <- stats
		}),
	}, 12, 3)
	if err != nil {
		t.Fatalf("NewTestRendererWithConfig() error = %v", err)
	}
	t.Cleanup(func() { _ = renderer.Close() })

	select {
	case stats := <-observed:
		if stats.Frame != 1 {
			t.Fatalf("observer frame = %d, want 1", stats.Frame)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("configured observer was not called")
	}
	if !strings.HasPrefix(renderer.Capture(), "configured") {
		t.Fatalf("configured screen = %q, want configured prefix", renderer.Capture())
	}
}

func TestTestRenderer_WaitsForFramesAndVisualIdle(t *testing.T) {
	widget := &rendererWidget{text: "go"}
	renderer, err := NewTestRenderer(t.Context(), widget, 12, 3)
	if err != nil {
		t.Fatalf("NewTestRenderer() error = %v", err)
	}
	t.Cleanup(func() { _ = renderer.Close() })

	baseline := renderer.LatestFrame().Stats.Frame
	renderer.App().Post(runtime.KeyMsg{Rune: '!'})
	frame, err := renderer.WaitForFrameAfter(baseline, 250*time.Millisecond)
	if err != nil {
		t.Fatalf("WaitForFrame() error = %v", err)
	}
	if !strings.HasPrefix(frame.Screen, "go!") {
		t.Fatalf("updated screen = %q, want go! prefix", frame.Screen)
	}

	frame, err = renderer.Flush(250 * time.Millisecond)
	if err != nil {
		t.Fatalf("Flush() error = %v", err)
	}
	if frame.Stats.Frame < 3 {
		t.Fatalf("flushed frame = %d, want at least 3", frame.Stats.Frame)
	}

	started := time.Now()
	if _, err := renderer.WaitForVisualIdle(15*time.Millisecond, 250*time.Millisecond); err != nil {
		t.Fatalf("WaitForVisualIdle() error = %v", err)
	}
	if elapsed := time.Since(started); elapsed < 15*time.Millisecond {
		t.Fatalf("visual idle returned after %v, want at least 15ms", elapsed)
	}
}

func TestTestRenderer_RejectsInvalidDimensionsAndTimesOut(t *testing.T) {
	if _, err := NewTestRenderer(t.Context(), &rendererWidget{}, 0, 3); err == nil {
		t.Fatal("NewTestRenderer() error = nil, want invalid dimensions error")
	}

	renderer, err := NewTestRenderer(t.Context(), &rendererWidget{}, 8, 2)
	if err != nil {
		t.Fatalf("NewTestRenderer() error = %v", err)
	}
	t.Cleanup(func() { _ = renderer.Close() })

	if _, err := renderer.WaitForFrame(5 * time.Millisecond); !errors.Is(err, ErrTimeout) {
		t.Fatalf("WaitForFrame() error = %v, want ErrTimeout", err)
	}
}

func TestTestRenderer_CloseIsIdempotent(t *testing.T) {
	renderer, err := NewTestRenderer(context.Background(), &rendererWidget{}, 8, 2)
	if err != nil {
		t.Fatalf("NewTestRenderer() error = %v", err)
	}
	if err := renderer.Close(); err != nil {
		t.Fatalf("first Close() error = %v", err)
	}
	if err := renderer.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
}
