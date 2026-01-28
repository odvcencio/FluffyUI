package state

import "testing"

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
