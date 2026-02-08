package main

import (
	"context"
	"sync"
	"testing"
	"time"
)

type recordingSpeaker struct {
	mu      sync.Mutex
	spoken  []string
	stopped int
}

func (r *recordingSpeaker) Speak(_ context.Context, text string) error {
	r.mu.Lock()
	r.spoken = append(r.spoken, text)
	r.mu.Unlock()
	time.Sleep(10 * time.Millisecond)
	return nil
}

func (r *recordingSpeaker) Stop() error {
	r.mu.Lock()
	r.stopped++
	r.mu.Unlock()
	return nil
}

func (r *recordingSpeaker) Close() error { return nil }

func (r *recordingSpeaker) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.spoken)
}

func (r *recordingSpeaker) texts() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]string, len(r.spoken))
	copy(out, r.spoken)
	return out
}

func TestSpeechQueueAssertive(t *testing.T) {
	rec := &recordingSpeaker{}
	queue := NewSpeechQueue(rec, 10*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	go queue.Run(ctx)

	queue.Enqueue(Utterance{Text: "hello", Priority: "assertive"})
	time.Sleep(100 * time.Millisecond)

	if rec.count() == 0 {
		t.Error("expected at least one spoken utterance")
	}
	texts := rec.texts()
	if texts[0] != "hello" {
		t.Errorf("expected 'hello', got %q", texts[0])
	}
}

func TestSpeechQueuePolite(t *testing.T) {
	rec := &recordingSpeaker{}
	queue := NewSpeechQueue(rec, 20*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	go queue.Run(ctx)

	queue.Enqueue(Utterance{Text: "polite message", Priority: "polite"})
	time.Sleep(200 * time.Millisecond)

	if rec.count() == 0 {
		t.Error("expected polite message to be spoken after debounce")
	}
}

func TestSpeechQueueContextCancel(t *testing.T) {
	rec := &recordingSpeaker{}
	queue := NewSpeechQueue(rec, 10*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		queue.Run(ctx)
		close(done)
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("queue.Run did not exit after cancel")
	}
}
