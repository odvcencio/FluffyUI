//go:build !js

package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/odvcencio/fluffyui/runtime"
)

// RealTimeNotifier handles real-time UI change notifications
type RealTimeNotifier struct {
	mu          sync.RWMutex
	subscribers map[string]*RealTimeSubscriber
	agent       *Agent

	// State tracking for change detection
	lastSnapshot  atomic.Value // Snapshot
	lastText      atomic.Value // string
	lastFocusedID atomic.Value // string

	// Config
	minInterval time.Duration // Minimum time between notifications
	maxInterval time.Duration // Maximum time between notifications (force update)

	// Control
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// RealTimeSubscriber represents a subscriber to real-time updates
type RealTimeSubscriber struct {
	ID        string
	SessionID string
	Filters   EventFilters

	// Channels
	Events chan UIEvent
	done   chan struct{}
}

// EventFilters controls which events a subscriber receives
type EventFilters struct {
	WidgetChanges bool // Widget tree changes
	FocusChanges  bool // Focus changes
	TextChanges   bool // Screen text changes
	ValueChanges  bool // Widget value changes
	StateChanges  bool // Widget state changes (enabled, checked, etc.)
	LayoutChanges bool // Layout/bounds changes
	AllEvents     bool // Receive all events
}

// DefaultEventFilters returns filters that capture common events
func DefaultEventFilters() EventFilters {
	return EventFilters{
		WidgetChanges: true,
		FocusChanges:  true,
		ValueChanges:  true,
		StateChanges:  true,
	}
}

// AllEventsFilter returns filters that capture all events
func AllEventsFilter() EventFilters {
	return EventFilters{AllEvents: true}
}

// UIEvent represents a real-time UI event
type UIEvent struct {
	Type      string          `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	SessionID string          `json:"session_id,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
}

// Event types
const (
	EventWidgetChanged = "widget_changed"
	EventFocusChanged  = "focus_changed"
	EventTextChanged   = "text_changed"
	EventValueChanged  = "value_changed"
	EventStateChanged  = "state_changed"
	EventLayoutChanged = "layout_changed"
	EventSnapshot      = "snapshot"
	EventHeartbeat     = "heartbeat"
)

// NewRealTimeNotifier creates a new real-time notifier
func NewRealTimeNotifier(agent *Agent) *RealTimeNotifier {
	ctx, cancel := context.WithCancel(context.Background())

	n := &RealTimeNotifier{
		subscribers: make(map[string]*RealTimeSubscriber),
		agent:       agent,
		minInterval: 50 * time.Millisecond,  // Debounce rapid changes
		maxInterval: 500 * time.Millisecond, // Force update at least twice per second
		ctx:         ctx,
		cancel:      cancel,
	}

	// Initialize atomic values
	n.lastText.Store("")
	n.lastFocusedID.Store("")

	return n
}

// Start begins the notifier loop
func (n *RealTimeNotifier) Start() {
	if n == nil {
		return
	}
	n.wg.Add(1)
	go n.watchLoop()
}

// Stop stops the notifier
func (n *RealTimeNotifier) Stop() {
	if n == nil {
		return
	}
	n.cancel()
	n.wg.Wait()

	n.mu.Lock()
	for _, sub := range n.subscribers {
		close(sub.done)
	}
	clear(n.subscribers)
	n.mu.Unlock()
}

// Subscribe creates a new subscriber
func (n *RealTimeNotifier) Subscribe(sessionID string, filters EventFilters) *RealTimeSubscriber {
	if n == nil {
		return nil
	}

	sub := &RealTimeSubscriber{
		ID:        generateSubscriberID(),
		SessionID: sessionID,
		Filters:   filters,
		Events:    make(chan UIEvent, 100),
		done:      make(chan struct{}),
	}

	n.mu.Lock()
	n.subscribers[sub.ID] = sub
	n.mu.Unlock()

	// Send initial snapshot
	go n.sendInitialSnapshot(sub)

	return sub
}

// Unsubscribe removes a subscriber
func (n *RealTimeNotifier) Unsubscribe(id string) {
	if n == nil {
		return
	}
	n.mu.Lock()
	if sub, ok := n.subscribers[id]; ok {
		close(sub.done)
		delete(n.subscribers, id)
	}
	n.mu.Unlock()
}

// UnsubscribeSession removes all subscribers for a session
func (n *RealTimeNotifier) UnsubscribeSession(sessionID string) {
	if n == nil || sessionID == "" {
		return
	}
	n.mu.Lock()
	for id, sub := range n.subscribers {
		if sub.SessionID == sessionID {
			close(sub.done)
			delete(n.subscribers, id)
		}
	}
	n.mu.Unlock()
}

// Notify forces a notification to all subscribers
func (n *RealTimeNotifier) Notify(event UIEvent) {
	if n == nil {
		return
	}

	n.mu.RLock()
	subscribers := make([]*RealTimeSubscriber, 0, len(n.subscribers))
	for _, sub := range n.subscribers {
		subscribers = append(subscribers, sub)
	}
	n.mu.RUnlock()

	for _, sub := range subscribers {
		select {
		case <-sub.done:
			continue
		case <-n.ctx.Done():
			return
		default:
		}

		if n.shouldSendToSubscriber(sub, event) {
			select {
			case sub.Events <- event:
			default:
				// Channel full, drop event
			}
		}
	}
}

// BroadcastSnapshot sends a snapshot to all subscribers
func (n *RealTimeNotifier) BroadcastSnapshot() {
	if n == nil || n.agent == nil {
		return
	}

	snap := n.agent.Snapshot()
	data, _ := json.Marshal(snap)

	n.Notify(UIEvent{
		Type:      EventSnapshot,
		Timestamp: time.Now(),
		Data:      data,
	})
}

// watchLoop monitors the UI for changes
func (n *RealTimeNotifier) watchLoop() {
	defer n.wg.Done()

	ticker := time.NewTicker(n.minInterval)
	defer ticker.Stop()

	forceUpdate := time.NewTicker(n.maxInterval)
	defer forceUpdate.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			n.checkForChanges()
		case <-forceUpdate.C:
			n.broadcastHeartbeat()
		}
	}
}

// checkForChanges detects UI changes and notifies subscribers
func (n *RealTimeNotifier) checkForChanges() {
	if n.agent == nil {
		return
	}

	snap := n.agent.Snapshot()
	if snap.Timestamp.IsZero() {
		return
	}

	// Get previous state
	var lastSnap Snapshot
	if ls, ok := n.lastSnapshot.Load().(Snapshot); ok {
		lastSnap = ls
	}

	lastText, _ := n.lastText.Load().(string)
	lastFocusedID, _ := n.lastFocusedID.Load().(string)

	now := time.Now()

	// Check for text changes
	if snap.Text != lastText {
		data, _ := json.Marshal(map[string]any{
			"previous": lastText,
			"current":  snap.Text,
		})
		n.Notify(UIEvent{
			Type:      EventTextChanged,
			Timestamp: now,
			Data:      data,
		})
		n.lastText.Store(snap.Text)
	}

	// Check for focus changes
	if snap.FocusedID != lastFocusedID {
		data, _ := json.Marshal(map[string]any{
			"previous": lastFocusedID,
			"current":  snap.FocusedID,
			"widget":   snap.Focused,
		})
		n.Notify(UIEvent{
			Type:      EventFocusChanged,
			Timestamp: now,
			Data:      data,
		})
		n.lastFocusedID.Store(snap.FocusedID)
	}

	// Check for widget changes
	if n.hasWidgetChanges(lastSnap.Widgets, snap.Widgets) {
		diff := n.diffWidgets(lastSnap.Widgets, snap.Widgets)
		data, _ := json.Marshal(diff)
		n.Notify(UIEvent{
			Type:      EventWidgetChanged,
			Timestamp: now,
			Data:      data,
		})
	}

	n.lastSnapshot.Store(snap)
}

// broadcastHeartbeat sends a heartbeat to all subscribers
func (n *RealTimeNotifier) broadcastHeartbeat() {
	n.Notify(UIEvent{
		Type:      EventHeartbeat,
		Timestamp: time.Now(),
	})
}

// sendInitialSnapshot sends the current snapshot to a new subscriber
func (n *RealTimeNotifier) sendInitialSnapshot(sub *RealTimeSubscriber) {
	if n.agent == nil {
		return
	}

	snap := n.agent.Snapshot()
	data, _ := json.Marshal(snap)

	select {
	case sub.Events <- UIEvent{
		Type:      EventSnapshot,
		Timestamp: time.Now(),
		SessionID: sub.SessionID,
		Data:      data,
	}:
	case <-time.After(time.Second):
		// Timeout
	}
}

// shouldSendToSubscriber checks if an event should be sent to a subscriber
func (n *RealTimeNotifier) shouldSendToSubscriber(sub *RealTimeSubscriber, event UIEvent) bool {
	if sub.Filters.AllEvents {
		return true
	}

	switch event.Type {
	case EventWidgetChanged:
		return sub.Filters.WidgetChanges
	case EventFocusChanged:
		return sub.Filters.FocusChanges
	case EventTextChanged:
		return sub.Filters.TextChanges
	case EventValueChanged:
		return sub.Filters.ValueChanges
	case EventStateChanged:
		return sub.Filters.StateChanges
	case EventLayoutChanged:
		return sub.Filters.LayoutChanges
	case EventSnapshot, EventHeartbeat:
		return true
	default:
		return false
	}
}

// hasWidgetChanges checks if widget tree has changed
func (n *RealTimeNotifier) hasWidgetChanges(old, new []WidgetInfo) bool {
	if len(old) != len(new) {
		return true
	}
	// Simple check - could be more sophisticated
	return false
}

// diffWidgets computes widget tree differences
func (n *RealTimeNotifier) diffWidgets(old, new []WidgetInfo) WidgetDiff {
	return WidgetDiff{
		Added:    n.findAddedWidgets(old, new),
		Removed:  n.findRemovedWidgets(old, new),
		Modified: n.findModifiedWidgets(old, new),
	}
}

func (n *RealTimeNotifier) findAddedWidgets(old, new []WidgetInfo) []WidgetInfo {
	oldIDs := make(map[string]bool)
	for _, w := range old {
		oldIDs[w.ID] = true
	}

	var added []WidgetInfo
	for _, w := range new {
		if !oldIDs[w.ID] {
			added = append(added, w)
		}
	}
	return added
}

func (n *RealTimeNotifier) findRemovedWidgets(old, new []WidgetInfo) []WidgetInfo {
	newIDs := make(map[string]bool)
	for _, w := range new {
		newIDs[w.ID] = true
	}

	var removed []WidgetInfo
	for _, w := range old {
		if !newIDs[w.ID] {
			removed = append(removed, w)
		}
	}
	return removed
}

func (n *RealTimeNotifier) findModifiedWidgets(old, new []WidgetInfo) []WidgetChange {
	oldMap := make(map[string]WidgetInfo)
	for _, w := range old {
		oldMap[w.ID] = w
	}

	var changes []WidgetChange
	for _, newW := range new {
		if oldW, ok := oldMap[newW.ID]; ok {
			if change := n.compareWidget(oldW, newW); change != nil {
				changes = append(changes, *change)
			}
		}
	}
	return changes
}

func (n *RealTimeNotifier) compareWidget(old, new WidgetInfo) *WidgetChange {
	change := &WidgetChange{
		ID:      new.ID,
		Changes: make(map[string]any),
	}

	if old.Value != new.Value {
		change.Changes["value"] = map[string]string{"from": old.Value, "to": new.Value}
	}
	if old.Label != new.Label {
		change.Changes["label"] = map[string]string{"from": old.Label, "to": new.Label}
	}
	if old.Focused != new.Focused {
		change.Changes["focused"] = new.Focused
	}
	if old.State != new.State {
		change.Changes["state"] = map[string]any{"from": old.State, "to": new.State}
	}

	if len(change.Changes) == 0 {
		return nil
	}
	return change
}

// WidgetDiff represents changes to the widget tree
type WidgetDiff struct {
	Added    []WidgetInfo   `json:"added,omitempty"`
	Removed  []WidgetInfo   `json:"removed,omitempty"`
	Modified []WidgetChange `json:"modified,omitempty"`
}

// WidgetChange represents a change to a specific widget
type WidgetChange struct {
	ID      string         `json:"id"`
	Changes map[string]any `json:"changes"`
}

// generateSubscriberID generates a unique subscriber ID
var subscriberIDCounter atomic.Int64

func generateSubscriberID() string {
	return fmt.Sprintf("sub-%d-%d", time.Now().UnixNano(), subscriberIDCounter.Add(1))
}

// RealTimeServer wraps an enhanced server with real-time capabilities
type RealTimeServer struct {
	*EnhancedServer
	notifier *RealTimeNotifier
}

// NewRealTimeServer creates a new real-time capable server
func NewRealTimeServer(opts EnhancedServerOptions) (*RealTimeServer, error) {
	server, err := NewEnhancedServer(opts)
	if err != nil {
		return nil, err
	}

	notifier := NewRealTimeNotifier(server.agent)

	rts := &RealTimeServer{
		EnhancedServer: server,
		notifier:       notifier,
	}

	// Hook into the render loop for change detection
	if opts.App != nil {
		rts.hookIntoRenderLoop(opts.App)
	}

	return rts, nil
}

// Start begins the real-time server
func (rts *RealTimeServer) Start() error {
	if err := rts.EnhancedServer.Start(); err != nil {
		return err
	}
	rts.notifier.Start()
	return nil
}

// Stop stops the real-time server
func (rts *RealTimeServer) Stop() error {
	rts.notifier.Stop()
	return rts.EnhancedServer.Stop()
}

// Subscribe creates a real-time subscription
func (rts *RealTimeServer) Subscribe(sessionID string, filters EventFilters) *RealTimeSubscriber {
	return rts.notifier.Subscribe(sessionID, filters)
}

// Unsubscribe removes a subscription
func (rts *RealTimeServer) Unsubscribe(id string) {
	rts.notifier.Unsubscribe(id)
}

// hookIntoRenderLoop hooks into the app's render loop
func (rts *RealTimeServer) hookIntoRenderLoop(app *runtime.App) {
	// Note: This is a simplified version. In practice, you'd want to
	// use the App's RenderObserver or message loop for tighter integration.
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			if rts.ctx.Err() != nil {
				return
			}
			// The notifier's watchLoop handles change detection
		}
	}()
}

// WaitForCondition waits for a UI condition to be met
func (rts *RealTimeServer) WaitForCondition(ctx context.Context, condition func(Snapshot) bool, timeout time.Duration) (Snapshot, error) {
	if rts == nil || rts.agent == nil {
		return Snapshot{}, errors.New("server not initialized")
	}

	// Check immediately
	snap := rts.agent.Snapshot()
	if condition(snap) {
		return snap, nil
	}

	// Subscribe to events
	sub := rts.Subscribe("", DefaultEventFilters())
	if sub == nil {
		return Snapshot{}, errors.New("failed to subscribe")
	}
	defer rts.Unsubscribe(sub.ID)

	// Set timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Wait for condition
	for {
		select {
		case <-ctx.Done():
			return rts.agent.Snapshot(), ctx.Err()
		case event := <-sub.Events:
			if event.Type == EventSnapshot || event.Type == EventWidgetChanged ||
				event.Type == EventFocusChanged || event.Type == EventValueChanged {
				snap := rts.agent.Snapshot()
				if condition(snap) {
					return snap, nil
				}
			}
		}
	}
}

// WaitForWidget waits for a widget to appear
func (rts *RealTimeServer) WaitForWidget(ctx context.Context, label string, timeout time.Duration) (WidgetInfo, error) {
	snap, err := rts.WaitForCondition(ctx, func(s Snapshot) bool {
		return findByLabelIn(s.Widgets, label) != nil
	}, timeout)

	if err != nil {
		return WidgetInfo{}, err
	}

	widget := findByLabelIn(snap.Widgets, label)
	if widget == nil {
		return WidgetInfo{}, errors.New("widget not found")
	}
	return *widget, nil
}

// WaitForText waits for text to appear on screen
func (rts *RealTimeServer) WaitForText(ctx context.Context, text string, timeout time.Duration) error {
	_, err := rts.WaitForCondition(ctx, func(s Snapshot) bool {
		return s.Text != "" && containsSubstring(s.Text, text)
	}, timeout)
	return err
}

// WaitForFocus waits for a widget to become focused
func (rts *RealTimeServer) WaitForFocus(ctx context.Context, widgetID string, timeout time.Duration) error {
	_, err := rts.WaitForCondition(ctx, func(s Snapshot) bool {
		return s.FocusedID == widgetID
	}, timeout)
	return err
}

// WaitForValue waits for a widget to have a specific value
func (rts *RealTimeServer) WaitForValue(ctx context.Context, widgetID string, value string, timeout time.Duration) error {
	_, err := rts.WaitForCondition(ctx, func(s Snapshot) bool {
		w := findByIDIn(s.Widgets, widgetID)
		return w != nil && w.Value == value
	}, timeout)
	return err
}

// Helper function
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) > 0 &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
