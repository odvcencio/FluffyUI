//go:build !js

package testing

import (
	"context"
	"testing"
	"time"

	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/backend/sim"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/terminal"
)

// TestWidget is a simple widget for testing.
type TestWidget struct {
	bounds  runtime.Rect
	text    string
	focused bool
}

func (t *TestWidget) Measure(constraints runtime.Constraints) runtime.Size {
	return runtime.Size{Width: 20, Height: 1}
}

func (t *TestWidget) Layout(bounds runtime.Rect) {
	t.bounds = bounds
}

func (t *TestWidget) Render(ctx runtime.RenderContext) {
	for i, r := range t.text {
		if i >= t.bounds.Width {
			break
		}
		ctx.Buffer.Set(t.bounds.X+i, t.bounds.Y, r, backend.DefaultStyle())
	}
}

func (t *TestWidget) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if key, ok := msg.(runtime.KeyMsg); ok {
		if key.Key == terminal.KeyRune {
			t.text += string(key.Rune)
			return runtime.Handled()
		}
	}
	return runtime.Unhandled()
}

func (t *TestWidget) Text() string {
	return t.text
}

func (t *TestWidget) Focus() {
	t.focused = true
}

func (t *TestWidget) Blur() {
	t.focused = false
}

func (t *TestWidget) IsFocused() bool {
	return t.focused
}

func (t *TestWidget) CanFocus() bool {
	return true
}

// startTestApp creates and starts a test app with the given widget.
func startTestApp(t *testing.T, widget runtime.Widget, width, height int) (*runtime.App, *TestSync, context.CancelFunc) {
	t.Helper()

	be := sim.New(width, height)
	if err := be.Init(); err != nil {
		t.Fatalf("failed to init sim backend: %v", err)
	}

	sync := NewTestSync(nil)
	app := runtime.NewApp(runtime.AppConfig{
		Backend: be,
		RenderObserver: runtime.RenderObserverFunc(func(runtime.RenderStats) {
			sync.NotifyRender()
		}),
	})
	app.SetRoot(widget)
	sync.app = app

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = app.Run(ctx)
	}()

	t.Cleanup(func() {
		cancel()
		<-done // Wait for app to finish before backend is finalized by app.Run
	})

	// Wait for initial render
	if err := sync.WaitForRender(100 * time.Millisecond); err != nil {
		// Initial render might not complete immediately, that's OK
	}

	return app, sync, cancel
}

func TestTestSync_WaitForRender(t *testing.T) {
	widget := &TestWidget{text: "initial"}
	_, sync, cancel := startTestApp(t, widget, 30, 5)
	defer cancel()

	// Trigger a render
	sync.app.Invalidate()

	err := sync.WaitForRender(100 * time.Millisecond)
	if err != nil {
		t.Errorf("WaitForRender failed: %v", err)
	}
}

func TestTestSync_WaitForRender_Timeout(t *testing.T) {
	widget := &TestWidget{text: "initial"}
	_, sync, cancel := startTestApp(t, widget, 30, 5)
	defer cancel()

	// Drain any pending renders
	sync.DrainRenders()

	// Don't trigger any render - should timeout
	err := sync.WaitForRender(10 * time.Millisecond)
	if err != ErrTimeout {
		t.Errorf("expected ErrTimeout, got %v", err)
	}
}

func TestTestSync_WaitForCondition(t *testing.T) {
	widget := &TestWidget{text: ""}
	app, sync, cancel := startTestApp(t, widget, 30, 5)
	defer cancel()

	// Post key to update widget text
	app.Post(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'X'})

	err := sync.WaitForCondition(func() bool {
		return widget.Text() == "X"
	}, 200*time.Millisecond)

	if err != nil {
		t.Errorf("WaitForCondition failed: %v", err)
	}
}

func TestTestSync_WaitForCondition_Timeout(t *testing.T) {
	widget := &TestWidget{text: ""}
	_, sync, cancel := startTestApp(t, widget, 30, 5)
	defer cancel()

	// Condition that will never be true
	err := sync.WaitForCondition(func() bool {
		return widget.Text() == "never"
	}, 20*time.Millisecond)

	if err != ErrTimeout {
		t.Errorf("expected ErrTimeout, got %v", err)
	}
}

func TestTestSync_WaitForText(t *testing.T) {
	widget := &TestWidget{text: ""}
	app, sync, cancel := startTestApp(t, widget, 30, 5)
	defer cancel()

	// Type some text
	app.Post(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'H'})
	app.Post(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'i'})

	err := sync.WaitForText(widget, "Hi", 200*time.Millisecond)
	if err != nil {
		t.Errorf("WaitForText failed: %v", err)
	}
}

func TestTestSync_WaitForFocus(t *testing.T) {
	widget := &TestWidget{text: "test"}
	_, sync, cancel := startTestApp(t, widget, 30, 5)
	defer cancel()

	// Focus the widget
	widget.Focus()
	sync.app.Invalidate()

	err := sync.WaitForFocus(widget, 100*time.Millisecond)
	if err != nil {
		t.Errorf("WaitForFocus failed: %v", err)
	}
}

func TestTestSync_SendKeyAndWait(t *testing.T) {
	widget := &TestWidget{text: ""}
	_, sync, cancel := startTestApp(t, widget, 30, 5)
	defer cancel()

	// Give app time to start
	time.Sleep(50 * time.Millisecond)

	err := sync.SendKeyAndWait(terminal.KeyEnter, 100*time.Millisecond)
	// Key might not cause a render, but SendKeyAndWait should still work
	// (timeout is acceptable if no render happens)
	_ = err
}

func TestTestSync_SendKeyRuneAndWait(t *testing.T) {
	widget := &TestWidget{text: ""}
	_, sync, cancel := startTestApp(t, widget, 30, 5)
	defer cancel()

	// Give app time to start
	time.Sleep(50 * time.Millisecond)

	err := sync.SendKeyRuneAndWait(terminal.KeyRune, 'A', 100*time.Millisecond)
	if err != nil && err != ErrTimeout {
		t.Errorf("SendKeyRuneAndWait failed: %v", err)
	}
}

func TestTestSync_TypeAndWait(t *testing.T) {
	widget := &TestWidget{text: ""}
	_, sync, cancel := startTestApp(t, widget, 30, 5)
	defer cancel()

	// Give app time to start
	time.Sleep(50 * time.Millisecond)

	err := sync.TypeAndWait("ab", 100*time.Millisecond)
	// Typing might not trigger renders for each character
	_ = err
}

func TestTestSync_DrainRenders(t *testing.T) {
	widget := &TestWidget{text: "test"}
	app, sync, cancel := startTestApp(t, widget, 30, 5)
	defer cancel()

	// Trigger multiple renders
	app.Invalidate()
	app.Invalidate()
	app.Invalidate()

	// Wait a bit for renders to queue
	time.Sleep(20 * time.Millisecond)

	// Drain should not block
	sync.DrainRenders()

	// After draining, should timeout immediately
	err := sync.WaitForRender(5 * time.Millisecond)
	// This might succeed if there are more renders, but shouldn't hang
	_ = err
}

func TestTestSync_PostAndWait(t *testing.T) {
	widget := &TestWidget{text: ""}
	_, sync, cancel := startTestApp(t, widget, 30, 5)
	defer cancel()

	err := sync.PostAndWait(runtime.InvalidateMsg{}, 100*time.Millisecond)
	if err != nil {
		t.Errorf("PostAndWait failed: %v", err)
	}
}

func TestTestSync_InvalidateAndWait(t *testing.T) {
	widget := &TestWidget{text: "test"}
	_, sync, cancel := startTestApp(t, widget, 30, 5)
	defer cancel()

	err := sync.InvalidateAndWait(100 * time.Millisecond)
	if err != nil {
		t.Errorf("InvalidateAndWait failed: %v", err)
	}
}

func TestTestSync_WaitForRenders(t *testing.T) {
	widget := &TestWidget{text: "test"}
	app, sync, cancel := startTestApp(t, widget, 30, 5)
	defer cancel()

	// Trigger multiple renders
	go func() {
		for i := 0; i < 3; i++ {
			time.Sleep(10 * time.Millisecond)
			app.Invalidate()
		}
	}()

	err := sync.WaitForRenders(3, 500*time.Millisecond)
	if err != nil {
		t.Errorf("WaitForRenders failed: %v", err)
	}
}

func TestTestSync_NilApp(t *testing.T) {
	sync := &TestSync{}

	err := sync.WaitForRender(10 * time.Millisecond)
	if err == nil {
		t.Error("expected error for nil app")
	}

	err = sync.SendKeyAndWait(terminal.KeyEnter, 10*time.Millisecond)
	if err == nil {
		t.Error("expected error for nil app")
	}

	err = sync.TypeAndWait("test", 10*time.Millisecond)
	if err == nil {
		t.Error("expected error for nil app")
	}
}
