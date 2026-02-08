package state

import (
	"fmt"
	"testing"
)

// ---------------------------------------------------------------------------
// BenchmarkSignalGet — read a signal value.
// ---------------------------------------------------------------------------

func BenchmarkSignalGet(b *testing.B) {
	sig := NewSignal(42)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sig.Get()
	}
}

// ---------------------------------------------------------------------------
// BenchmarkSignalSet — set a signal value with no subscribers (raw write path).
// ---------------------------------------------------------------------------

func BenchmarkSignalSet(b *testing.B) {
	sig := NewSignal(0)
	// Ensure equality check triggers a change each time.
	sig.SetEqualFunc(EqualComparable[int])

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sig.Set(i)
	}
}

// ---------------------------------------------------------------------------
// BenchmarkSignalSetNoChange — set to the same value (equality short-circuit).
// ---------------------------------------------------------------------------

func BenchmarkSignalSetNoChange(b *testing.B) {
	sig := NewSignal(42)
	sig.SetEqualFunc(EqualComparable[int])

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sig.Set(42)
	}
}

// ---------------------------------------------------------------------------
// BenchmarkSignalNotify — set with 1, 10, and 100 subscribers.
// ---------------------------------------------------------------------------

func BenchmarkSignalNotify(b *testing.B) {
	for _, n := range []int{1, 10, 100} {
		b.Run(fmt.Sprintf("subscribers_%d", n), func(b *testing.B) {
			sig := NewSignal(0)
			sig.SetEqualFunc(EqualComparable[int])

			// Attach n subscribers; keep unsub funcs alive to prevent GC.
			unsubs := make([]func(), n)
			for j := 0; j < n; j++ {
				unsubs[j] = sig.Subscribe(func() {
					// no-op subscriber
				})
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				sig.Set(i)
			}
			b.StopTimer()

			for _, unsub := range unsubs {
				unsub()
			}
		})
	}
}

// ---------------------------------------------------------------------------
// BenchmarkComputedGet — read a cached computed value (no recomputation).
// ---------------------------------------------------------------------------

func BenchmarkComputedGet(b *testing.B) {
	a := NewSignal(10)
	a.SetEqualFunc(EqualComparable[int])

	comp := NewComputed(func() int {
		return a.Get() * 2
	}, a)
	comp.SetEqualFunc(EqualComparable[int])

	// Warm the cached value.
	_ = comp.Get()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = comp.Get()
	}
}

// ---------------------------------------------------------------------------
// BenchmarkComputedRecompute — read after dependency change (forced recompute).
// ---------------------------------------------------------------------------

func BenchmarkComputedRecompute(b *testing.B) {
	a := NewSignal(0)
	a.SetEqualFunc(EqualComparable[int])

	comp := NewComputed(func() int {
		return a.Get() * 2
	}, a)
	comp.SetEqualFunc(EqualComparable[int])

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Set(i)       // Triggers recompute.
		_ = comp.Get() // Read the new value.
	}
}

// ---------------------------------------------------------------------------
// BenchmarkComputedChain — chain of computed values (depth 5).
// ---------------------------------------------------------------------------

func BenchmarkComputedChain(b *testing.B) {
	root := NewSignal(0)
	root.SetEqualFunc(EqualComparable[int])

	prev := root
	var last *Computed[int]
	for depth := 0; depth < 5; depth++ {
		dep := prev // capture for closure
		c := NewComputed(func() int {
			return dep.Get() + 1
		}, dep)
		c.SetEqualFunc(EqualComparable[int])
		last = c
		prev = c.signal
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		root.Set(i)
		_ = last.Get()
	}
}

// ---------------------------------------------------------------------------
// BenchmarkBatchUpdate — batch 100 signal updates, single notification flush.
// ---------------------------------------------------------------------------

func BenchmarkBatchUpdate(b *testing.B) {
	const count = 100
	signals := make([]*Signal[int], count)
	unsubs := make([]func(), count)
	for i := range signals {
		signals[i] = NewSignal(0)
		signals[i].SetEqualFunc(EqualComparable[int])
		unsubs[i] = signals[i].Subscribe(func() {
			// no-op
		})
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Batch(func() {
			for j := 0; j < count; j++ {
				signals[j].Set(i + j)
			}
		})
	}
	b.StopTimer()

	for _, unsub := range unsubs {
		unsub()
	}
}

// ---------------------------------------------------------------------------
// BenchmarkSubscribeUnsubscribe — measure subscription churn.
// ---------------------------------------------------------------------------

func BenchmarkSubscribeUnsubscribe(b *testing.B) {
	sig := NewSignal(0)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		unsub := sig.Subscribe(func() {})
		unsub()
	}
}
