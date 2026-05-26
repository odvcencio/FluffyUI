//go:build !js

// Package widgettest provides test harness utilities for FluffyUI widgets.
package widgettest

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"m31labs.dev/fluffyui/backend/sim"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/terminal"
)

// DefaultTimeout is the default timeout for wait operations.
const DefaultTimeout = 100 * time.Millisecond

// ErrTimeout is returned when a wait operation times out.
var ErrTimeout = errors.New("timeout waiting for condition")

// Harness runs a widget inside a live app with a simulation backend.
// It provides deterministic synchronization for testing widget behavior.
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
	rendered := make(chan struct{}, 10)
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
	if !h.WaitForRender(DefaultTimeout) {
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

// Wait is deprecated. Use WaitForRender or WaitForCondition instead.
// Kept for backwards compatibility.
func (h *Harness) Wait(d time.Duration) {
	time.Sleep(d)
}

// DrainRenders clears any pending render notifications.
func (h *Harness) DrainRenders() {
	if h == nil || h.rendered == nil {
		return
	}
	for {
		select {
		case <-h.rendered:
		default:
			return
		}
	}
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

// WaitForRenders waits for at least n renders to complete.
func (h *Harness) WaitForRenders(n int, timeout time.Duration) bool {
	if h == nil || h.rendered == nil {
		return false
	}
	deadline := time.Now().Add(timeout)
	for i := 0; i < n; i++ {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return false
		}
		select {
		case <-h.rendered:
		case <-time.After(remaining):
			return false
		}
	}
	return true
}

// WaitForCondition waits for a condition function to return true.
func (h *Harness) WaitForCondition(fn func() bool, timeout time.Duration) error {
	if h == nil {
		return errors.New("harness not initialized")
	}
	if fn == nil {
		return errors.New("condition function is nil")
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if fn() {
			return nil
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}
		waitTime := remaining
		if waitTime > 10*time.Millisecond {
			waitTime = 10 * time.Millisecond
		}
		select {
		case <-h.rendered:
		case <-time.After(waitTime):
		}
	}
	if fn() {
		return nil
	}
	return ErrTimeout
}

// WaitForFocus waits for a widget to receive focus.
func (h *Harness) WaitForFocus(widget runtime.Widget, timeout time.Duration) error {
	if h == nil {
		return errors.New("harness not initialized")
	}
	focusable, ok := widget.(runtime.Focusable)
	if !ok {
		return errors.New("widget is not focusable")
	}
	return h.WaitForCondition(focusable.IsFocused, timeout)
}

// WaitForText waits for a widget to contain specific text.
func (h *Harness) WaitForText(widget runtime.Widget, text string, timeout time.Duration) error {
	if h == nil {
		return errors.New("harness not initialized")
	}
	textProvider, ok := widget.(interface{ Text() string })
	if !ok {
		return errors.New("widget does not provide Text() method")
	}
	return h.WaitForCondition(func() bool {
		return textProvider.Text() == text
	}, timeout)
}

// InjectKey injects a key event into the backend.
func (h *Harness) InjectKey(key terminal.Key, r rune) {
	if h == nil || h.Backend == nil {
		return
	}
	h.Backend.InjectKey(key, r)
}

// InjectKeyAndWait injects a key and waits for the resulting render.
func (h *Harness) InjectKeyAndWait(key terminal.Key, r rune, timeout time.Duration) bool {
	if h == nil || h.Backend == nil {
		return false
	}
	h.DrainRenders()
	h.Backend.InjectKey(key, r)
	return h.WaitForRender(timeout)
}

// SendKey posts a key message through the app and waits for render.
func (h *Harness) SendKey(key terminal.Key, r rune) bool {
	if h == nil || h.App == nil {
		return false
	}
	h.DrainRenders()
	h.App.Post(runtime.KeyMsg{Key: key, Rune: r})
	return h.WaitForRender(DefaultTimeout)
}

// SendKeyWithTimeout posts a key message and waits with custom timeout.
func (h *Harness) SendKeyWithTimeout(key terminal.Key, r rune, timeout time.Duration) bool {
	if h == nil || h.App == nil {
		return false
	}
	h.DrainRenders()
	h.App.Post(runtime.KeyMsg{Key: key, Rune: r})
	return h.WaitForRender(timeout)
}

// InjectKeyString injects a string as key events.
func (h *Harness) InjectKeyString(str string) {
	if h == nil || h.Backend == nil {
		return
	}
	h.Backend.InjectKeyString(str)
}

// TypeAndWait types a string and waits for renders after each character.
func (h *Harness) TypeAndWait(text string, timeout time.Duration) bool {
	if h == nil || h.App == nil {
		return false
	}
	for _, ch := range text {
		h.DrainRenders()
		h.App.Post(runtime.KeyMsg{Key: terminal.KeyRune, Rune: ch})
		if !h.WaitForRender(timeout) {
			return false
		}
	}
	return true
}

// InjectMouse injects a mouse event into the backend.
func (h *Harness) InjectMouse(x, y int, button terminal.MouseButton) {
	if h == nil || h.Backend == nil {
		return
	}
	h.Backend.InjectMouse(x, y, button)
}

// InjectMouseAndWait injects a mouse event and waits for render.
func (h *Harness) InjectMouseAndWait(x, y int, button terminal.MouseButton, timeout time.Duration) bool {
	if h == nil || h.Backend == nil {
		return false
	}
	h.DrainRenders()
	h.Backend.InjectMouse(x, y, button)
	return h.WaitForRender(timeout)
}

// Resize updates the backend size and emits a resize event.
func (h *Harness) Resize(width, height int) {
	if h == nil || h.Backend == nil {
		return
	}
	h.Backend.InjectResize(width, height)
}

// ResizeAndWait resizes and waits for the resulting render.
func (h *Harness) ResizeAndWait(width, height int, timeout time.Duration) bool {
	if h == nil || h.Backend == nil {
		return false
	}
	h.DrainRenders()
	h.Backend.InjectResize(width, height)
	return h.WaitForRender(timeout)
}

// Capture returns the full screen buffer as a string.
func (h *Harness) Capture() string {
	if h == nil || h.Backend == nil {
		return ""
	}
	return h.Backend.Capture()
}

// CaptureRegion returns a region of the screen buffer.
func (h *Harness) CaptureRegion(x, y, width, height int) string {
	if h == nil || h.Backend == nil {
		return ""
	}
	return h.Backend.CaptureRegion(x, y, width, height)
}

// InvalidateAndWait triggers an invalidation and waits for render.
func (h *Harness) InvalidateAndWait(timeout time.Duration) bool {
	if h == nil || h.App == nil {
		return false
	}
	h.DrainRenders()
	h.App.Invalidate()
	return h.WaitForRender(timeout)
}

// ContainsText checks if the screen contains the given text.
func (h *Harness) ContainsText(text string) bool {
	if h == nil || h.Backend == nil {
		return false
	}
	return h.Backend.ContainsText(text)
}

// AssertContains fails the test if screen doesn't contain the text.
func (h *Harness) AssertContains(text string) {
	h.t.Helper()
	if !h.ContainsText(text) {
		h.t.Errorf("expected screen to contain %q\nScreen:\n%s", text, h.Capture())
	}
}

// AssertNotContains fails the test if screen contains the text.
func (h *Harness) AssertNotContains(text string) {
	h.t.Helper()
	if h.ContainsText(text) {
		h.t.Errorf("expected screen to NOT contain %q\nScreen:\n%s", text, h.Capture())
	}
}
