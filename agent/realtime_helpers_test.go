package agent

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend/sim"
	"github.com/odvcencio/fluffyui/runtime"
)

func TestRealTimeNotifierExtras(t *testing.T) {
	agt := New(Config{Sim: sim.New(10, 4)})
	notifier := NewRealTimeNotifier(agt)

	sub1 := notifier.Subscribe("session-1", AllEventsFilter())
	sub2 := notifier.Subscribe("session-2", AllEventsFilter())

	select {
	case <-sub1.Events:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for initial snapshot")
	}
	select {
	case <-sub2.Events:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for initial snapshot")
	}

	notifier.BroadcastSnapshot()
	select {
	case event := <-sub1.Events:
		if event.Type != EventSnapshot {
			t.Fatalf("expected snapshot event, got %s", event.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for broadcast snapshot")
	}

	notifier.broadcastHeartbeat()
	select {
	case event := <-sub1.Events:
		if event.Type != EventHeartbeat {
			t.Fatalf("expected heartbeat event, got %s", event.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for heartbeat")
	}

	notifier.UnsubscribeSession("session-1")
	notifier.mu.RLock()
	_, exists := notifier.subscribers[sub1.ID]
	notifier.mu.RUnlock()
	if exists {
		t.Fatal("expected session-1 subscriber to be removed")
	}
}

func TestRealTimeHelpers(t *testing.T) {
	if !containsSubstring("hello world", "world") {
		t.Fatal("expected containsSubstring true")
	}
	if containsSubstring("hello", "missing") {
		t.Fatal("expected containsSubstring false")
	}
	if !findSubstring("abcde", "bcd") {
		t.Fatal("expected findSubstring true")
	}

	n := &RealTimeNotifier{}
	old := WidgetInfo{ID: "w1", Label: "Old", Value: "A"}
	new := WidgetInfo{ID: "w1", Label: "New", Value: "B", Focused: true, State: accessibility.StateSet{Disabled: true}}
	change := n.compareWidget(old, new)
	if change == nil || len(change.Changes) == 0 {
		t.Fatal("expected widget change")
	}
	if n.compareWidget(old, old) != nil {
		t.Fatal("expected no changes")
	}

	sub := &RealTimeSubscriber{Filters: EventFilters{}}
	if n.shouldSendToSubscriber(sub, UIEvent{Type: EventSnapshot}) != true {
		t.Fatal("expected snapshot to always send")
	}
	if n.shouldSendToSubscriber(sub, UIEvent{Type: EventWidgetChanged}) {
		t.Fatal("expected widget change filtered out")
	}
	sub.Filters.WidgetChanges = true
	if !n.shouldSendToSubscriber(sub, UIEvent{Type: EventWidgetChanged}) {
		t.Fatal("expected widget change to send")
	}
	sub.Filters.AllEvents = true
	if !n.shouldSendToSubscriber(sub, UIEvent{Type: "custom"}) {
		t.Fatal("expected all events to send")
	}
}

func TestRealTimeServerWaitHelpers(t *testing.T) {
	input := &testInput{label: "Name", value: "Ada"}
	root := runtime.VBox(runtime.Fixed(input)).WithGap(1)

	be := sim.New(40, 10)
	app := runtime.NewApp(runtime.AppConfig{
		Backend:           be,
		Root:              root,
		Update:            runtime.DefaultUpdate,
		FocusRegistration: runtime.FocusRegistrationAuto,
		TickRate:          time.Second / 60,
	})

	agt := New(Config{App: app, Sim: be, IncludeText: true, TickRate: time.Millisecond})
	runAppForTest(t, app)

	if err := agt.Focus("Name"); err != nil {
		t.Fatalf("focus: %v", err)
	}

	info := agt.FindByLabel("Name")
	if info == nil {
		t.Fatal("expected widget info")
	}

	opts := DefaultEnhancedServerOptions()
	opts.Addr = "unix:" + filepath.Join(t.TempDir(), "realtime.sock")
	opts.App = app
	opts.Agent = agt

	server, err := NewRealTimeServer(opts)
	if err != nil {
		t.Fatalf("new realtime server: %v", err)
	}

	if _, err := server.WaitForWidget(context.Background(), "Name", time.Second); err != nil {
		t.Fatalf("wait for widget: %v", err)
	}
	if err := server.WaitForFocus(context.Background(), info.ID, time.Second); err != nil {
		t.Fatalf("wait for focus: %v", err)
	}
	if err := server.WaitForValue(context.Background(), info.ID, "Ada", time.Second); err != nil {
		t.Fatalf("wait for value: %v", err)
	}
	if err := server.WaitForText(context.Background(), "Ada", time.Second); err != nil {
		t.Fatalf("wait for text: %v", err)
	}

	_, err = server.WaitForCondition(context.Background(), func(Snapshot) bool { return false }, 10*time.Millisecond)
	if err == nil {
		t.Fatal("expected wait for condition timeout")
	}
}
