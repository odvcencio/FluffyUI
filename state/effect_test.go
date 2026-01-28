package state

import "testing"

func TestEffectRunsAndDisposes(t *testing.T) {
	sig := NewSignal(0)
	called := 0
	eff := NewEffect(func() {
		called++
	}, sig)

	if called != 1 {
		t.Fatalf("expected initial run, got %d", called)
	}

	sig.Set(1)
	if called != 2 {
		t.Fatalf("expected effect after change, got %d", called)
	}

	eff.Dispose()
	sig.Set(2)
	if called != 2 {
		t.Fatalf("expected no runs after dispose, got %d", called)
	}
}

func TestEffectTrigger(t *testing.T) {
	called := 0
	eff := NewEffect(func() {
		called++
	})

	if called != 1 {
		t.Fatalf("expected initial run, got %d", called)
	}

	eff.Trigger()
	if called != 2 {
		t.Fatalf("expected trigger to run effect, got %d", called)
	}
}
