package state

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestDebounce(t *testing.T) {
	var count atomic.Int32
	fn := Debounce(50*time.Millisecond, func() {
		count.Add(1)
	})

	// Call rapidly — timer should keep resetting.
	for i := 0; i < 10; i++ {
		fn()
	}

	// Should not have fired yet.
	if count.Load() != 0 {
		t.Errorf("expected 0 calls immediately after rapid fire, got %d", count.Load())
	}

	// Wait for the debounce period to elapse.
	time.Sleep(100 * time.Millisecond)

	if count.Load() != 1 {
		t.Errorf("expected 1 call after debounce, got %d", count.Load())
	}
}

func TestDebounce_MultipleBursts(t *testing.T) {
	var count atomic.Int32
	fn := Debounce(50*time.Millisecond, func() {
		count.Add(1)
	})

	// First burst.
	fn()
	fn()
	fn()
	time.Sleep(100 * time.Millisecond)

	if count.Load() != 1 {
		t.Errorf("expected 1 call after first burst, got %d", count.Load())
	}

	// Second burst.
	fn()
	fn()
	time.Sleep(100 * time.Millisecond)

	if count.Load() != 2 {
		t.Errorf("expected 2 calls after second burst, got %d", count.Load())
	}
}

func TestThrottle(t *testing.T) {
	var count atomic.Int32
	fn := Throttle(50*time.Millisecond, func() {
		count.Add(1)
	})

	// First call should execute immediately.
	fn()
	if count.Load() != 1 {
		t.Errorf("expected 1 call after first invocation, got %d", count.Load())
	}

	// Rapid calls within the window should be dropped.
	fn()
	fn()
	fn()
	if count.Load() != 1 {
		t.Errorf("expected still 1 call during throttle window, got %d", count.Load())
	}

	// After the duration, the next call should go through.
	time.Sleep(60 * time.Millisecond)
	fn()
	if count.Load() != 2 {
		t.Errorf("expected 2 calls after throttle window expired, got %d", count.Load())
	}
}

func TestThrottle_ZeroDuration(t *testing.T) {
	var count atomic.Int32
	fn := Throttle(0, func() {
		count.Add(1)
	})

	// Every call should execute with zero duration.
	fn()
	fn()
	fn()

	if count.Load() != 3 {
		t.Errorf("expected 3 calls with zero throttle, got %d", count.Load())
	}
}

func TestDebouncedSignal(t *testing.T) {
	source := NewSignal("hello")
	debounced := DebouncedSignal(source, 50*time.Millisecond)

	// Initial value should match.
	if debounced.Get() != "hello" {
		t.Errorf("expected initial value 'hello', got %q", debounced.Get())
	}

	// Rapid updates should not propagate immediately.
	source.Set("a")
	source.Set("b")
	source.Set("c")

	// The debounced signal should still hold the old value.
	time.Sleep(10 * time.Millisecond)
	if debounced.Get() == "c" {
		t.Error("debounced signal updated too quickly")
	}

	// After the delay, it should settle on the final value.
	time.Sleep(100 * time.Millisecond)
	if debounced.Get() != "c" {
		t.Errorf("expected 'c' after debounce delay, got %q", debounced.Get())
	}
}

func TestDebouncedSignal_SingleUpdate(t *testing.T) {
	source := NewSignal(0)
	debounced := DebouncedSignal(source, 50*time.Millisecond)

	source.Set(42)
	time.Sleep(100 * time.Millisecond)

	if debounced.Get() != 42 {
		t.Errorf("expected 42 after single update, got %d", debounced.Get())
	}
}
