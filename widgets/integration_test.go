package widgets

import (
	"context"
	"testing"
	"time"

	"m31labs.dev/fluffyui/backend/sim"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/terminal"
)

// integrationSync provides synchronization for integration tests.
type integrationSync struct {
	renderCh chan struct{}
	app      *runtime.App
}

func (s *integrationSync) notifyRender() {
	select {
	case s.renderCh <- struct{}{}:
	default:
	}
}

func (s *integrationSync) drainRenders() {
	for {
		select {
		case <-s.renderCh:
		default:
			return
		}
	}
}

func (s *integrationSync) waitForRender(timeout time.Duration) bool {
	select {
	case <-s.renderCh:
		return true
	case <-time.After(timeout):
		return false
	}
}

func (s *integrationSync) waitForCondition(fn func() bool, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if s.checkOnAppLoop(fn) {
			return true
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
		case <-s.renderCh:
		case <-time.After(waitTime):
		}
	}
	return s.checkOnAppLoop(fn)
}

func (s *integrationSync) checkOnAppLoop(fn func() bool) bool {
	if s == nil || fn == nil {
		return false
	}
	if s.app == nil {
		return fn()
	}
	var out bool
	err := s.app.Call(context.Background(), func(*runtime.App) error {
		out = fn()
		return nil
	})
	if err != nil {
		return false
	}
	return out
}

func (s *integrationSync) layerCount() int {
	if s == nil || s.app == nil {
		return 0
	}
	count := 0
	_ = s.app.Call(context.Background(), func(*runtime.App) error {
		if screen := s.app.Screen(); screen != nil {
			count = screen.LayerCount()
		}
		return nil
	})
	return count
}

func (s *integrationSync) injectKeyAndWait(be *sim.Backend, key terminal.Key, r rune, timeout time.Duration) bool {
	s.drainRenders()
	be.InjectKey(key, r)
	return s.waitForRender(timeout)
}

func startTestAppWithSync(t *testing.T, be *sim.Backend, root runtime.Widget) (*runtime.App, *integrationSync) {
	t.Helper()

	sync := &integrationSync{
		renderCh: make(chan struct{}, 10),
	}

	app := runtime.NewApp(runtime.AppConfig{
		Backend:           be,
		Root:              root,
		TickRate:          time.Second / 60,
		FocusRegistration: runtime.FocusRegistrationAuto,
		RenderObserver: runtime.RenderObserverFunc(func(runtime.RenderStats) {
			sync.notifyRender()
		}),
	})
	sync.app = app

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx)
	}()

	// Wait for initial render
	sync.waitForRender(100 * time.Millisecond)

	t.Cleanup(func() {
		cancel()
		<-done
	})

	return app, sync
}

func TestIntegration_TabFocusTraversal(t *testing.T) {
	be := sim.New(40, 6)
	if err := be.Init(); err != nil {
		t.Fatalf("failed to init sim backend: %v", err)
	}

	input1 := NewInput()
	input2 := NewInput()
	root := NewStack(input1, input2)

	_, sync := startTestAppWithSync(t, be, root)

	// Wait for first input to be focused
	if !sync.waitForCondition(input1.IsFocused, 100*time.Millisecond) {
		t.Fatalf("expected first input to be focused")
	}

	// Tab to second input
	sync.injectKeyAndWait(be, terminal.KeyTab, 0, 100*time.Millisecond)
	if !sync.waitForCondition(input2.IsFocused, 100*time.Millisecond) {
		t.Fatalf("expected second input to be focused after Tab")
	}

	// Tab wraps back to first
	sync.injectKeyAndWait(be, terminal.KeyTab, 0, 100*time.Millisecond)
	if !sync.waitForCondition(input1.IsFocused, 100*time.Millisecond) {
		t.Fatalf("expected focus to wrap back to first input after Tab")
	}
}

func TestIntegration_SelectDropdownOverlay(t *testing.T) {
	be := sim.New(40, 8)
	if err := be.Init(); err != nil {
		t.Fatalf("failed to init sim backend: %v", err)
	}

	selectWidget := NewSelect(
		SelectOption{Label: "One", Value: 1},
		SelectOption{Label: "Two", Value: 2},
	).Apply(WithDropdownMode())

	app, sync := startTestAppWithSync(t, be, selectWidget)

	if !sync.waitForCondition(func() bool {
		return app.Screen() != nil && app.Screen().LayerCount() == 1
	}, 100*time.Millisecond) {
		t.Fatalf("expected 1 layer on start, got %d", sync.layerCount())
	}

	// Open dropdown
	sync.injectKeyAndWait(be, terminal.KeyEnter, 0, 100*time.Millisecond)
	if !sync.waitForCondition(func() bool {
		return app.Screen().LayerCount() >= 2
	}, 100*time.Millisecond) {
		t.Fatalf("expected dropdown overlay layer, got %d", sync.layerCount())
	}

	// Close dropdown
	sync.injectKeyAndWait(be, terminal.KeyEscape, 0, 100*time.Millisecond)
	if !sync.waitForCondition(func() bool {
		return app.Screen().LayerCount() == 1
	}, 100*time.Millisecond) {
		t.Fatalf("expected overlay to close, got %d layers", sync.layerCount())
	}
}
