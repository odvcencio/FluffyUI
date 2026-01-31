//go:build !js

package widgettest

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/backend/sim"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

// Harness runs a widget inside a live app with a simulation backend.
type Harness struct {
	t       *testing.T
	App     *runtime.App
	Backend *sim.Backend

	rendered  chan struct{}
	cancel    context.CancelFunc
	done      chan error
	closeOnce sync.Once
}

// New starts a background app with the provided root widget.
func New(t *testing.T, root runtime.Widget, width, height int) *Harness {
	t.Helper()
	be := sim.New(width, height)
	if err := be.Init(); err != nil {
		t.Fatalf("failed to init sim backend: %v", err)
	}
	rendered := make(chan struct{}, 1)
	app := runtime.NewApp(runtime.AppConfig{
		Backend: be,
		RenderObserver: runtime.RenderObserverFunc(func(runtime.RenderStats) {
			select {
			case rendered <- struct{}{}:
			default:
			}
		}),
	})
	app.SetRoot(root)

	ctx, cancel := context.WithCancel(context.Background())
	h := &Harness{
		t:        t,
		App:      app,
		Backend:  be,
		rendered: rendered,
		cancel:   cancel,
		done:     make(chan error, 1),
	}
	go func() {
		h.done <- app.Run(ctx)
	}()

	t.Cleanup(h.Close)
	app.Invalidate()
	if !h.WaitForRender(100 * time.Millisecond) {
		t.Fatalf("initial render did not complete")
	}
	return h
}

// Close stops the app and releases backend resources.
func (h *Harness) Close() {
	if h == nil {
		return
	}
	h.closeOnce.Do(func() {
		if h.cancel != nil {
			h.cancel()
		}
		if h.done != nil {
			err := <-h.done
			if err != nil && err != context.Canceled {
				h.t.Errorf("app run failed: %v", err)
			}
		}
	})
}

// Wait lets the app process events and render.
func (h *Harness) Wait(d time.Duration) {
	time.Sleep(d)
}

// WaitForRender waits for a render to complete or times out.
func (h *Harness) WaitForRender(timeout time.Duration) bool {
	if h == nil || h.rendered == nil {
		return false
	}
	select {
	case <-h.rendered:
		return true
	case <-time.After(timeout):
		return false
	}
}

// InjectKey injects a key event into the backend.
func (h *Harness) InjectKey(key terminal.Key, r rune) {
	if h == nil || h.Backend == nil {
		return
	}
	h.Backend.InjectKey(key, r)
}

// InjectKeyString injects a string as key events.
func (h *Harness) InjectKeyString(str string) {
	if h == nil || h.Backend == nil {
		return
	}
	h.Backend.InjectKeyString(str)
}

// InjectMouse injects a mouse event into the backend.
func (h *Harness) InjectMouse(x, y int, button terminal.MouseButton) {
	if h == nil || h.Backend == nil {
		return
	}
	h.Backend.InjectMouse(x, y, button)
}

// Resize updates the backend size and emits a resize event.
func (h *Harness) Resize(width, height int) {
	if h == nil || h.Backend == nil {
		return
	}
	h.Backend.InjectResize(width, height)
}

// Capture returns the full screen buffer as a string.
func (h *Harness) Capture() string {
	if h == nil || h.Backend == nil {
		return ""
	}
	return h.Backend.Capture()
}
