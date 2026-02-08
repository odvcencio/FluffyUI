package accessibility

import (
	"context"
	"sync"
	"testing"
	"time"
)

type testSpeaker struct {
	mu      sync.Mutex
	spoken  []string
	stopped int
}

func (s *testSpeaker) Speak(_ context.Context, text string) error {
	s.mu.Lock()
	s.spoken = append(s.spoken, text)
	s.mu.Unlock()
	return nil
}

func (s *testSpeaker) Stop() error {
	s.mu.Lock()
	s.stopped++
	s.mu.Unlock()
	return nil
}

func (s *testSpeaker) Close() error { return nil }

// slowSpeaker simulates a slow TTS engine (like PowerShell SAPI).
type slowSpeaker struct {
	mu     sync.Mutex
	spoken []string
	delay  time.Duration
}

func (s *slowSpeaker) Speak(_ context.Context, text string) error {
	time.Sleep(s.delay)
	s.mu.Lock()
	s.spoken = append(s.spoken, text)
	s.mu.Unlock()
	return nil
}

func (s *slowSpeaker) Stop() error  { return nil }
func (s *slowSpeaker) Close() error { return nil }

func (s *slowSpeaker) history() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.spoken))
	copy(out, s.spoken)
	return out
}

func (s *testSpeaker) history() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.spoken))
	copy(out, s.spoken)
	return out
}

func (s *testSpeaker) stopCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stopped
}

func TestAssertiveSpeaksImmediately(t *testing.T) {
	speaker := &testSpeaker{}
	ann := &SimpleAnnouncer{}
	ann.SetSpeaker(speaker)
	defer ann.CloseSpeaker()

	ann.Announce("urgent alert", PriorityAssertive)

	// Give goroutine time to run.
	time.Sleep(50 * time.Millisecond)

	history := speaker.history()
	if len(history) == 0 {
		t.Fatal("expected assertive message to be spoken")
	}
	if history[0] != "urgent alert" {
		t.Errorf("expected 'urgent alert', got %q", history[0])
	}
	if speaker.stopCount() == 0 {
		t.Error("expected Stop to be called before assertive speech")
	}
}

func TestPoliteSpeechDebounces(t *testing.T) {
	speaker := &testSpeaker{}
	ann := &SimpleAnnouncer{}
	ann.SetSpeaker(speaker)
	defer ann.CloseSpeaker()

	ann.Announce("first", PriorityPolite)
	ann.Announce("second", PriorityPolite)
	ann.Announce("third", PriorityPolite)

	// Wait for debounce (50ms) plus margin.
	time.Sleep(150 * time.Millisecond)

	history := speaker.history()
	if len(history) != 1 {
		t.Fatalf("expected 1 debounced message, got %d: %v", len(history), history)
	}
	if history[0] != "third" {
		t.Errorf("expected last polite message 'third', got %q", history[0])
	}
}

func TestNilSpeakerNoOp(t *testing.T) {
	ann := &SimpleAnnouncer{}
	// No speaker set — should not panic.
	ann.Announce("hello", PriorityAssertive)
	ann.Announce("world", PriorityPolite)

	history := ann.History()
	if len(history) != 2 {
		t.Fatalf("expected 2 announcements, got %d", len(history))
	}
}

func TestCloseSpeaker(t *testing.T) {
	speaker := &testSpeaker{}
	ann := &SimpleAnnouncer{}
	ann.SetSpeaker(speaker)

	ann.Announce("test", PriorityPolite)
	ann.CloseSpeaker()

	// After close, further announcements should not attempt speech.
	ann.Announce("after close", PriorityAssertive)
	time.Sleep(100 * time.Millisecond)

	// Only "test" might have been queued (debounced), but speaker is closed.
	// The main thing is no panic.
}

func TestHighPrioritySpeaksImmediately(t *testing.T) {
	speaker := &testSpeaker{}
	ann := &SimpleAnnouncer{}
	ann.SetSpeaker(speaker)
	defer ann.CloseSpeaker()

	ann.Announce("high priority", PriorityHigh)
	time.Sleep(50 * time.Millisecond)

	history := speaker.history()
	if len(history) == 0 {
		t.Fatal("expected high priority message to be spoken immediately")
	}
	if history[0] != "high priority" {
		t.Errorf("expected 'high priority', got %q", history[0])
	}
}

func TestPoliteDoesNotCancelAssertive(t *testing.T) {
	// Simulates: assertive fires, then 10ms later polite fires.
	// Polite's debounce timer (50ms) must NOT cancel the assertive speech.
	speaker := &slowSpeaker{delay: 200 * time.Millisecond}
	ann := &SimpleAnnouncer{}
	ann.SetSpeaker(speaker)
	defer ann.CloseSpeaker()

	ann.Announce("dialog opened", PriorityAssertive)
	time.Sleep(10 * time.Millisecond)
	ann.Announce("focus changed", PriorityPolite)

	// Wait for assertive (200ms) + polite debounce (50ms) + polite speech (200ms) + margin.
	time.Sleep(600 * time.Millisecond)

	history := speaker.history()
	if len(history) < 2 {
		t.Fatalf("expected both messages spoken, got %d: %v", len(history), history)
	}
	if history[0] != "dialog opened" {
		t.Errorf("expected assertive first, got %q", history[0])
	}
	if history[1] != "focus changed" {
		t.Errorf("expected polite second, got %q", history[1])
	}
}

func TestUrgentPrioritySpeaksImmediately(t *testing.T) {
	speaker := &testSpeaker{}
	ann := &SimpleAnnouncer{}
	ann.SetSpeaker(speaker)
	defer ann.CloseSpeaker()

	ann.Announce("urgent", PriorityUrgent)
	time.Sleep(50 * time.Millisecond)

	history := speaker.history()
	if len(history) == 0 {
		t.Fatal("expected urgent message to be spoken immediately")
	}
}
