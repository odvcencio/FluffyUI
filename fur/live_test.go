package fur

import (
	"bytes"
	"context"
	"testing"
	"time"
)

func TestLiveBasic(t *testing.T) {
	progress := NewProgress(100)
	live := NewLive(progress)

	if live == nil {
		t.Fatal("NewLive returned nil")
	}
}

func TestLiveWithRate(t *testing.T) {
	progress := NewProgress(100)
	live := NewLive(progress).WithRate(50 * time.Millisecond)

	if live.rate != 50*time.Millisecond {
		t.Errorf("got rate %v, want 50ms", live.rate)
	}
}

func TestLiveWithTransient(t *testing.T) {
	progress := NewProgress(100)
	live := NewLive(progress).WithTransient(true)

	if !live.transient {
		t.Error("expected transient=true")
	}
}

func TestLiveUpdate(t *testing.T) {
	progress := NewProgress(100)
	live := NewLive(progress)

	newProgress := NewProgress(200)
	live.Update(newProgress)

	if live.currentRenderable() != newProgress {
		t.Error("Update did not change renderable")
	}
}

func TestLiveStop(t *testing.T) {
	var buf bytes.Buffer
	c := New(WithOutput(&buf), WithNoColor(), WithWidth(40))

	progress := NewProgress(100)
	live := NewLive(progress).WithConsole(c).WithRate(10 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		_ = live.Start(ctx)
		close(done)
	}()

	time.Sleep(30 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// success
	case <-time.After(500 * time.Millisecond):
		t.Error("Live.Start did not stop after context cancel")
	}
}

func TestLiveExplicitStop(t *testing.T) {
	var buf bytes.Buffer
	c := New(WithOutput(&buf), WithNoColor(), WithWidth(40))

	progress := NewProgress(100)
	live := NewLive(progress).WithConsole(c).WithRate(10 * time.Millisecond)

	done := make(chan struct{})

	go func() {
		_ = live.Start(context.Background())
		close(done)
	}()

	time.Sleep(30 * time.Millisecond)
	live.Stop()

	select {
	case <-done:
		// success
	case <-time.After(500 * time.Millisecond):
		t.Error("Live.Start did not stop after Stop()")
	}
}
