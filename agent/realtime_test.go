//go:build !js

package agent

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/backend/sim"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/widgets"
)

func TestRealTimeNotifier(t *testing.T) {
	be := sim.New(80, 24)

	label := widgets.NewLabel("Hello")
	app := runtime.NewApp(runtime.AppConfig{Backend: be})
	app.SetRoot(label)

	runAppForTest(t, app)

	// Wait for app to start
	time.Sleep(50 * time.Millisecond)
	app.Post(runtime.InvalidateMsg{})

	agt := New(Config{App: app})
	notifier := NewRealTimeNotifier(agt)
	notifier.Start()
	defer notifier.Stop()

	// Test subscription
	sub := notifier.Subscribe("test-session", DefaultEventFilters())
	if sub == nil {
		t.Fatal("failed to create subscription")
	}
	defer notifier.Unsubscribe(sub.ID)

	// Wait for initial snapshot
	select {
	case event := <-sub.Events:
		if event.Type != EventSnapshot {
			t.Fatalf("expected snapshot event, got %s", event.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for initial snapshot")
	}
}

func TestRealTimeServer(t *testing.T) {
	be := sim.New(80, 24)

	label := widgets.NewLabel("Test")
	app := runtime.NewApp(runtime.AppConfig{Backend: be})
	app.SetRoot(label)

	runAppForTest(t, app)

	// Wait for app to start
	time.Sleep(50 * time.Millisecond)
	app.Post(runtime.InvalidateMsg{})

	opts := DefaultEnhancedServerOptions()
	opts.Addr = "unix:" + filepath.Join(t.TempDir(), "realtime.sock")
	opts.App = app

	server, err := NewRealTimeServer(opts)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer server.Stop()

	// Test subscription
	sub := server.Subscribe("test-session", DefaultEventFilters())
	if sub == nil {
		t.Fatal("failed to create subscription")
	}
	defer server.Unsubscribe(sub.ID)

	// Wait for initial snapshot
	select {
	case event := <-sub.Events:
		if event.Type != EventSnapshot {
			t.Fatalf("expected snapshot event, got %s", event.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for initial snapshot")
	}
}

func TestWaitForCondition(t *testing.T) {
	be := sim.New(80, 24)

	label := widgets.NewLabel("Initial")
	app := runtime.NewApp(runtime.AppConfig{Backend: be})
	app.SetRoot(label)

	runAppForTest(t, app)

	// Wait for app to start
	time.Sleep(50 * time.Millisecond)
	app.Post(runtime.InvalidateMsg{})

	opts := DefaultEnhancedServerOptions()
	opts.Addr = "unix:" + filepath.Join(t.TempDir(), "realtime.sock")
	opts.App = app

	server, err := NewRealTimeServer(opts)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer server.Stop()

	// Test wait for text that already exists
	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second)
	defer cancel2()

	_, err = server.WaitForCondition(ctx2, func(s Snapshot) bool {
		return len(s.Widgets) > 0
	}, time.Second)
	if err != nil {
		t.Fatalf("wait for condition failed: %v", err)
	}
}

func TestEventFilters(t *testing.T) {
	// Test default filters
	filters := DefaultEventFilters()
	if !filters.WidgetChanges {
		t.Error("WidgetChanges should be true by default")
	}
	if !filters.FocusChanges {
		t.Error("FocusChanges should be true by default")
	}
	if filters.TextChanges {
		t.Error("TextChanges should be false by default")
	}

	// Test all events filter
	allFilters := AllEventsFilter()
	if !allFilters.AllEvents {
		t.Error("AllEvents should be true")
	}
}

func TestRealTimeNotifierBroadcast(t *testing.T) {
	be := sim.New(80, 24)

	app := runtime.NewApp(runtime.AppConfig{Backend: be})
	app.SetRoot(widgets.NewLabel("Test"))

	runAppForTest(t, app)

	time.Sleep(50 * time.Millisecond)
	app.Post(runtime.InvalidateMsg{})

	agt := New(Config{App: app})
	notifier := NewRealTimeNotifier(agt)
	notifier.Start()
	defer notifier.Stop()

	// Create subscriber
	sub := notifier.Subscribe("test", DefaultEventFilters())
	if sub == nil {
		t.Fatal("failed to create subscription")
	}
	defer notifier.Unsubscribe(sub.ID)

	// Consume initial snapshot
	<-sub.Events

	// Broadcast a test event
	testEvent := UIEvent{
		Type:      EventFocusChanged,
		Timestamp: time.Now(),
	}
	notifier.Notify(testEvent)

	// Wait for the event
	select {
	case event := <-sub.Events:
		if event.Type != EventFocusChanged {
			t.Fatalf("expected focus_changed event, got %s", event.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}
