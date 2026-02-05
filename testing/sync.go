//go:build !js

// Package testing provides test utilities for FluffyUI applications.
package testing

import (
	"context"
	"errors"
	"time"

	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

// ErrTimeout is returned when a wait operation times out.
var ErrTimeout = errors.New("timeout waiting for condition")

// TestSync provides channel-based synchronization for deterministic testing.
// It replaces time.Sleep() hacks with event-driven waiting that's both faster
// and more reliable.
type TestSync struct {
	renderCh chan struct{}
	eventCh  chan runtime.Message
	app      *runtime.App
}

// NewTestSync creates a TestSync connected to an app.
// The app must have a RenderObserver configured to notify the sync.
func NewTestSync(app *runtime.App) *TestSync {
	return &TestSync{
		renderCh: make(chan struct{}, 10),
		eventCh:  make(chan runtime.Message, 100),
		app:      app,
	}
}

// RenderCh returns the render notification channel.
// Used to integrate with app's RenderObserver.
func (s *TestSync) RenderCh() chan struct{} {
	if s == nil {
		return nil
	}
	return s.renderCh
}

// NotifyRender signals that a render completed.
// Called by the RenderObserver.
func (s *TestSync) NotifyRender() {
	if s == nil || s.renderCh == nil {
		return
	}
	select {
	case s.renderCh <- struct{}{}:
	default:
		// Don't block if channel is full
	}
}

// DrainRenders clears any pending render notifications.
func (s *TestSync) DrainRenders() {
	if s == nil || s.renderCh == nil {
		return
	}
	for {
		select {
		case <-s.renderCh:
		default:
			return
		}
	}
}

// WaitForRender waits for a render to complete.
func (s *TestSync) WaitForRender(timeout time.Duration) error {
	if s == nil || s.renderCh == nil {
		return errors.New("TestSync not initialized")
	}
	select {
	case <-s.renderCh:
		return nil
	case <-time.After(timeout):
		return ErrTimeout
	}
}

// WaitForRenders waits for at least n renders to complete.
func (s *TestSync) WaitForRenders(n int, timeout time.Duration) error {
	if s == nil || s.renderCh == nil {
		return errors.New("TestSync not initialized")
	}
	deadline := time.Now().Add(timeout)
	for i := 0; i < n; i++ {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return ErrTimeout
		}
		select {
		case <-s.renderCh:
		case <-time.After(remaining):
			return ErrTimeout
		}
	}
	return nil
}

// WaitForEvent waits for a message to be processed by the app.
func (s *TestSync) WaitForEvent(timeout time.Duration) (runtime.Message, error) {
	if s == nil || s.eventCh == nil {
		return nil, errors.New("TestSync not initialized")
	}
	select {
	case msg := <-s.eventCh:
		return msg, nil
	case <-time.After(timeout):
		return nil, ErrTimeout
	}
}

// WaitForFocus waits for a widget to receive focus.
func (s *TestSync) WaitForFocus(widget runtime.Widget, timeout time.Duration) error {
	if s == nil || s.app == nil {
		return errors.New("TestSync not initialized")
	}
	focusable, ok := widget.(runtime.Focusable)
	if !ok {
		return errors.New("widget is not focusable")
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if s.checkOnAppLoop(func() bool { return focusable.IsFocused() }) {
			return nil
		}
		// Wait for a render which might update focus
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
	if s.checkOnAppLoop(func() bool { return focusable.IsFocused() }) {
		return nil
	}
	return ErrTimeout
}

// WaitForText waits for a widget to contain specific text.
// The widget must implement runtime.TextProvider or have a Text() method.
func (s *TestSync) WaitForText(widget runtime.Widget, text string, timeout time.Duration) error {
	if s == nil {
		return errors.New("TestSync not initialized")
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if s.checkOnAppLoop(func() bool {
			textProvider, ok := widget.(interface{ Text() string })
			return ok && textProvider.Text() == text
		}) {
			return nil
		}
		// Wait for a render which might update text
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
	if s.checkOnAppLoop(func() bool {
		textProvider, ok := widget.(interface{ Text() string })
		return ok && textProvider.Text() == text
	}) {
		return nil
	}
	return ErrTimeout
}

// WaitForCondition waits for a condition function to return true.
func (s *TestSync) WaitForCondition(fn func() bool, timeout time.Duration) error {
	if s == nil {
		return errors.New("TestSync not initialized")
	}
	if fn == nil {
		return errors.New("condition function is nil")
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if s.checkOnAppLoop(fn) {
			return nil
		}
		// Wait for a render which might change state
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
	if s.checkOnAppLoop(fn) {
		return nil
	}
	return ErrTimeout
}

func (s *TestSync) checkOnAppLoop(fn func() bool) bool {
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

// SendKeyAndWait injects a key event and waits for the resulting render.
func (s *TestSync) SendKeyAndWait(key terminal.Key, timeout time.Duration) error {
	if s == nil || s.app == nil {
		return errors.New("TestSync not initialized")
	}
	// Drain any pending renders first
	s.DrainRenders()
	// Post the key message
	s.app.Post(runtime.KeyMsg{Key: key})
	// Wait for the resulting render
	return s.WaitForRender(timeout)
}

// SendKeyRuneAndWait injects a key event with a rune and waits for the resulting render.
func (s *TestSync) SendKeyRuneAndWait(key terminal.Key, r rune, timeout time.Duration) error {
	if s == nil || s.app == nil {
		return errors.New("TestSync not initialized")
	}
	// Drain any pending renders first
	s.DrainRenders()
	// Post the key message
	s.app.Post(runtime.KeyMsg{Key: key, Rune: r})
	// Wait for the resulting render
	return s.WaitForRender(timeout)
}

// TypeAndWait types a string character by character, waiting for renders after each.
func (s *TestSync) TypeAndWait(text string, timeout time.Duration) error {
	if s == nil || s.app == nil {
		return errors.New("TestSync not initialized")
	}
	for _, ch := range text {
		s.DrainRenders()
		s.app.Post(runtime.KeyMsg{Key: terminal.KeyRune, Rune: ch})
		if err := s.WaitForRender(timeout); err != nil {
			return err
		}
	}
	return nil
}

// SendMouseAndWait injects a mouse event and waits for the resulting render.
func (s *TestSync) SendMouseAndWait(x, y int, button runtime.MouseButton, action runtime.MouseAction, timeout time.Duration) error {
	if s == nil || s.app == nil {
		return errors.New("TestSync not initialized")
	}
	s.DrainRenders()
	s.app.Post(runtime.MouseMsg{X: x, Y: y, Button: button, Action: action})
	return s.WaitForRender(timeout)
}

// InvalidateAndWait triggers an invalidation and waits for the render.
func (s *TestSync) InvalidateAndWait(timeout time.Duration) error {
	if s == nil || s.app == nil {
		return errors.New("TestSync not initialized")
	}
	s.DrainRenders()
	s.app.Invalidate()
	return s.WaitForRender(timeout)
}

// PostAndWait posts a custom message and waits for the render.
func (s *TestSync) PostAndWait(msg runtime.Message, timeout time.Duration) error {
	if s == nil || s.app == nil {
		return errors.New("TestSync not initialized")
	}
	s.DrainRenders()
	s.app.Post(msg)
	return s.WaitForRender(timeout)
}
