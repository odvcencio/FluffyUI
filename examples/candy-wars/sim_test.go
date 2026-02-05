package main

import (
	"context"
	"math/rand"
	"path/filepath"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/backend/sim"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
	fluffytest "github.com/odvcencio/fluffyui/testing"
)

func TestCandyWarsSimRun(t *testing.T) {
	rand.Seed(1)

	be := sim.New(120, 40)
	if err := be.Init(); err != nil {
		t.Fatalf("failed to init sim backend: %v", err)
	}

	t.Setenv(metaPathEnv, filepath.Join(t.TempDir(), "candy-wars-meta.json"))

	sync := fluffytest.NewTestSync(nil)
	view := NewAppView()
	app := runtime.NewApp(runtime.AppConfig{
		Backend: be,
		Root:    view,
		RenderObserver: runtime.RenderObserverFunc(func(runtime.RenderStats) {
			sync.NotifyRender()
		}),
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_ = app.Run(ctx)
	}()
	app.Invalidate()

	if err := sync.WaitForRender(500 * time.Millisecond); err != nil {
		t.Fatalf("initial render timeout: %v", err)
	}

	// Start new game (default difficulty).
	be.InjectKey(terminal.KeyEnter, 0)
	if err := sync.WaitForRender(500 * time.Millisecond); err != nil {
		t.Fatalf("start game render timeout: %v", err)
	}
	if view.showNewGame {
		t.Fatalf("expected new game dialog to close\n\nScreen:\n%s", be.Capture())
	}

	initialDay := view.game.Day.Get()

	// Buy two units of the selected candy.
	be.InjectKeyRune('b')
	if err := sync.WaitForRender(500 * time.Millisecond); err != nil {
		t.Fatalf("buy dialog render timeout: %v", err)
	}
	if !waitForAppCondition(app, 500*time.Millisecond, func() bool {
		return view.gameView.showTrade
	}) {
		t.Fatalf("expected trade dialog to open\n\nScreen:\n%s", be.Capture())
	}
	if err := app.Call(context.Background(), func(*runtime.App) error {
		view.gameView.tradeInput.SetText("2")
		view.gameView.handleTradeInput(runtime.KeyMsg{Key: terminal.KeyEnter})
		return nil
	}); err != nil {
		t.Fatalf("trade confirm failed: %v", err)
	}
	if err := sync.WaitForRender(500 * time.Millisecond); err != nil {
		t.Fatalf("buy confirm render timeout: %v", err)
	}
	if waitForAppCondition(app, 200*time.Millisecond, func() bool {
		return view.gameView.showTrade
	}) {
		t.Fatalf("expected trade dialog to close\n\nScreen:\n%s", be.Capture())
	}

	inv := view.game.Inventory.Get()
	if inv[CandyTypes[0].Name] == 0 {
		t.Fatalf("expected inventory to increase for %s (msg: %s)\n\nScreen:\n%s", CandyTypes[0].Name, view.game.Message.Get(), be.Capture())
	}

	// End the day early.
	if err := app.Call(context.Background(), func(*runtime.App) error {
		view.game.EndDayEarly()
		view.gameView.refresh()
		return nil
	}); err != nil {
		t.Fatalf("end day call failed: %v", err)
	}
	if err := sync.WaitForRender(500 * time.Millisecond); err != nil {
		t.Fatalf("end day render timeout: %v", err)
	}
	if got := view.game.Day.Get(); got != initialDay+1 {
		t.Fatalf("expected day %d, got %d", initialDay+1, got)
	}

	// Switch to Event Log tab.
	sync.DrainRenders()
	if err := app.Call(context.Background(), func(*runtime.App) error {
		view.gameView.tabs.SetSelected(4)
		app.Relayout()
		return nil
	}); err != nil {
		t.Fatalf("log tab call failed: %v", err)
	}
	if err := sync.WaitForRender(500 * time.Millisecond); err != nil {
		t.Fatalf("log tab render timeout: %v", err)
	}
	if view.gameView.tabs.SelectedIndex() != 4 {
		t.Fatalf("expected log tab selected, got %d", view.gameView.tabs.SelectedIndex())
	}
	if bounds := view.gameView.logPanel.Bounds(); bounds.Width <= 0 || bounds.Height <= 0 {
		t.Fatalf("expected log panel bounds to be laid out, got %+v\n\nScreen:\n%s", bounds, be.Capture())
	}
	if view.gameView.eventLog.EntryCount() == 0 {
		t.Fatalf("expected event log to contain entries after gameplay\n\nScreen:\n%s", be.Capture())
	}

	// Switch to Showcase tab.
	sync.DrainRenders()
	if err := app.Call(context.Background(), func(*runtime.App) error {
		view.gameView.tabs.SetSelected(5)
		app.Relayout()
		return nil
	}); err != nil {
		t.Fatalf("showcase tab call failed: %v", err)
	}
	if err := sync.WaitForRender(500 * time.Millisecond); err != nil {
		t.Fatalf("showcase tab render timeout: %v", err)
	}
	if view.gameView.tabs.SelectedIndex() != 5 {
		t.Fatalf("expected showcase tab selected, got %d", view.gameView.tabs.SelectedIndex())
	}
	if bounds := view.gameView.showcaseTab.Bounds(); bounds.Width <= 0 || bounds.Height <= 0 {
		t.Fatalf("expected showcase bounds to be laid out, got %+v\n\nScreen:\n%s", bounds, be.Capture())
	}
}

func waitForAppCondition(app *runtime.App, timeout time.Duration, fn func() bool) bool {
	if app == nil || fn == nil {
		return false
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		matched := false
		if err := app.Call(context.Background(), func(*runtime.App) error {
			matched = fn()
			return nil
		}); err == nil && matched {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}
