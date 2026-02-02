package state

import (
	"sync"
	"testing"
)

func TestBatch_DefersNotifications(t *testing.T) {
	s := NewSignal(0)
	called := false
	s.Subscribe(func() {
		called = true
	})

	Batch(func() {
		s.Set(1)
		if called {
			t.Fatal("expected callback to be deferred during batch")
		}
	})
	if !called {
		t.Fatal("expected callback after batch flush")
	}
}

func TestBatch_Nested(t *testing.T) {
	s := NewSignal(0)
	count := 0
	s.Subscribe(func() {
		count++
	})

	Batch(func() {
		Batch(func() {
			s.Set(1)
		})
		if count != 0 {
			t.Fatalf("expected no callbacks during nested batch, got %d", count)
		}
		s.Set(2)
	})
	if count != 2 {
		t.Fatalf("expected 2 callbacks after batch, got %d", count)
	}
}

// TestBatch_ConcurrentEnqueue tests that concurrent enqueue operations
// do not cause race conditions (fix for 1.2 TOCTOU race).
func TestBatch_ConcurrentEnqueue(t *testing.T) {
	// Start at -1 so all Set(0..9) trigger callbacks (value changes)
	s := NewSignal(-1)
	var callCount int
	var mu sync.Mutex
	s.Subscribe(func() {
		mu.Lock()
		callCount++
		mu.Unlock()
	})

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			Batch(func() {
				s.Set(val)
			})
		}(i)
	}
	wg.Wait()

	mu.Lock()
	count := callCount
	mu.Unlock()
	if count != 10 {
		t.Fatalf("expected 10 callbacks, got %d", count)
	}
}
