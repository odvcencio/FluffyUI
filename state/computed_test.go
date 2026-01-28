package state

import "testing"

func TestComputed_Recompute(t *testing.T) {
	a := NewSignal(1)
	b := NewSignal(2)
	a.SetEqualFunc(EqualComparable[int])
	b.SetEqualFunc(EqualComparable[int])

	sum := NewComputed(func() int {
		return a.Get() + b.Get()
	}, a, b)
	sum.SetEqualFunc(EqualComparable[int])

	if got := sum.Get(); got != 3 {
		t.Fatalf("expected initial sum 3, got %d", got)
	}

	calls := 0
	unsub := sum.Subscribe(func() {
		calls++
	})

	if !a.Set(2) {
		t.Fatalf("expected signal change")
	}
	if got := sum.Get(); got != 4 {
		t.Fatalf("expected sum 4 after change, got %d", got)
	}
	if calls != 1 {
		t.Fatalf("expected 1 recompute, got %d", calls)
	}

	if a.Set(2) {
		t.Fatalf("expected no change on equal set")
	}
	if calls != 1 {
		t.Fatalf("expected no extra recompute, got %d", calls)
	}

	b.Set(3)
	if got := sum.Get(); got != 5 {
		t.Fatalf("expected sum 5 after change, got %d", got)
	}
	if calls != 2 {
		t.Fatalf("expected 2 recomputes, got %d", calls)
	}

	unsub()
	b.Set(4)
	if calls != 2 {
		t.Fatalf("expected no recompute after unsubscribe, got %d", calls)
	}
}

func TestComputed_Stop(t *testing.T) {
	a := NewSignal(1)
	a.SetEqualFunc(EqualComparable[int])

	comp := NewComputed(func() int {
		return a.Get()
	}, a)
	comp.SetEqualFunc(EqualComparable[int])

	comp.Stop()
	comp.Stop()

	if !a.Set(2) {
		t.Fatalf("expected signal change")
	}
	if got := comp.Get(); got != 1 {
		t.Fatalf("expected computed to stay at 1 after stop, got %d", got)
	}
}

func TestComputed_Scheduler(t *testing.T) {
	a := NewSignal(1)
	a.SetEqualFunc(EqualComparable[int])

	queue := NewQueue()
	comp := NewComputedWithScheduler(queue, func() int {
		return a.Get()
	}, a)
	comp.SetEqualFunc(EqualComparable[int])

	if !a.Set(2) {
		t.Fatalf("expected signal change")
	}
	if got := comp.Get(); got != 1 {
		t.Fatalf("expected computed to stay at 1 before flush, got %d", got)
	}
	if flushed := queue.Flush(); flushed != 1 {
		t.Fatalf("expected 1 recompute flushed, got %d", flushed)
	}
	if got := comp.Get(); got != 2 {
		t.Fatalf("expected computed to update after flush, got %d", got)
	}
}

func TestComputed_AutoDependencies(t *testing.T) {
	a := NewSignal(1)
	b := NewSignal(2)
	a.SetEqualFunc(EqualComparable[int])
	b.SetEqualFunc(EqualComparable[int])

	sum := NewComputed(func() int {
		return a.Get() + b.Get()
	})
	sum.SetEqualFunc(EqualComparable[int])

	if got := sum.Get(); got != 3 {
		t.Fatalf("expected initial sum 3, got %d", got)
	}

	calls := 0
	sum.Subscribe(func() {
		calls++
	})

	a.Set(4)
	if got := sum.Get(); got != 6 {
		t.Fatalf("expected sum 6 after change, got %d", got)
	}
	if calls != 1 {
		t.Fatalf("expected 1 recompute, got %d", calls)
	}

	b.Set(10)
	if got := sum.Get(); got != 14 {
		t.Fatalf("expected sum 14 after change, got %d", got)
	}
	if calls != 2 {
		t.Fatalf("expected 2 recomputes, got %d", calls)
	}
}

func TestComputed_AutoDependenciesDynamic(t *testing.T) {
	switcher := NewSignal(true)
	a := NewSignal(1)
	b := NewSignal(5)
	switcher.SetEqualFunc(EqualComparable[bool])
	a.SetEqualFunc(EqualComparable[int])
	b.SetEqualFunc(EqualComparable[int])

	comp := NewComputed(func() int {
		if switcher.Get() {
			return a.Get()
		}
		return b.Get()
	})
	comp.SetEqualFunc(EqualComparable[int])

	calls := 0
	comp.Subscribe(func() {
		calls++
	})

	b.Set(6)
	if calls != 0 {
		t.Fatalf("expected no recompute while switcher true, got %d", calls)
	}

	switcher.Set(false)
	if calls != 1 {
		t.Fatalf("expected recompute after switcher flip, got %d", calls)
	}

	a.Set(2)
	if calls != 1 {
		t.Fatalf("expected no recompute after switching off a, got %d", calls)
	}

	b.Set(7)
	if calls != 2 {
		t.Fatalf("expected recompute after b change, got %d", calls)
	}
}
