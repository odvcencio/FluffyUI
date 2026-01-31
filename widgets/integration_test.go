package widgets

import (
	"context"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/backend/sim"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

func startTestApp(t *testing.T, be *sim.Backend, root runtime.Widget) *runtime.App {
	t.Helper()
	app := runtime.NewApp(runtime.AppConfig{
		Backend:           be,
		Root:              root,
		TickRate:          time.Second / 60,
		FocusRegistration: runtime.FocusRegistrationAuto,
	})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx)
	}()
	// Allow the app to initialize.
	time.Sleep(20 * time.Millisecond)
	t.Cleanup(func() {
		cancel()
		<-done
	})
	return app
}

func TestIntegration_TabFocusTraversal(t *testing.T) {
	be := sim.New(40, 6)
	if err := be.Init(); err != nil {
		t.Fatalf("failed to init sim backend: %v", err)
	}

	input1 := NewInput()
	input2 := NewInput()
	root := NewStack(input1, input2)

	_ = startTestApp(t, be, root)

	if !input1.IsFocused() {
		t.Fatalf("expected first input to be focused")
	}

	be.InjectKey(terminal.KeyTab, 0)
	time.Sleep(20 * time.Millisecond)
	if !input2.IsFocused() {
		t.Fatalf("expected second input to be focused after Tab")
	}

	be.InjectKey(terminal.KeyTab, 0)
	time.Sleep(20 * time.Millisecond)
	if !input1.IsFocused() {
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

	app := startTestApp(t, be, selectWidget)

	if app.Screen().LayerCount() != 1 {
		t.Fatalf("expected 1 layer on start, got %d", app.Screen().LayerCount())
	}

	be.InjectKey(terminal.KeyEnter, 0)
	time.Sleep(30 * time.Millisecond)
	if app.Screen().LayerCount() < 2 {
		t.Fatalf("expected dropdown overlay layer, got %d", app.Screen().LayerCount())
	}

	be.InjectKey(terminal.KeyEscape, 0)
	time.Sleep(30 * time.Millisecond)
	if app.Screen().LayerCount() != 1 {
		t.Fatalf("expected overlay to close, got %d layers", app.Screen().LayerCount())
	}
}
