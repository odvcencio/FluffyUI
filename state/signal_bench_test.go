package state

import (
	"strconv"
	"testing"
)

// BenchmarkSignalGetString measures reading a string signal, which exercises
// the string-typed deepEqual fast path in recordDependency.
func BenchmarkSignalGetString(b *testing.B) {
	s := NewSignal("hello world")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Get()
	}
}

// BenchmarkSignalSetWithChange measures Set when the value actually changes
// each iteration (no subscribers), exercising the equality check + value store.
func BenchmarkSignalSetWithChange(b *testing.B) {
	s := NewSignal(0)
	s.SetEqualFunc(EqualComparable[int])
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Set(i)
	}
}

// BenchmarkSignalSetString measures Set for string-typed signals,
// exercising the deepEqual string fast path on each comparison.
func BenchmarkSignalSetString(b *testing.B) {
	s := NewSignal("initial")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Set("value-" + strconv.Itoa(i))
	}
}

// BenchmarkSignalSetStringNoChange measures Set when the string value is
// identical, exercising the equality short-circuit for strings.
func BenchmarkSignalSetStringNoChange(b *testing.B) {
	s := NewSignal("constant")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Set("constant")
	}
}

// BenchmarkDeepEqualString measures the fast-path string equality check
// used to suppress redundant signal updates.
func BenchmarkDeepEqualString(b *testing.B) {
	a := "hello world"
	c := "hello world"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = deepEqual(a, c)
	}
}

// BenchmarkDeepEqualInt measures the fast-path int equality check.
func BenchmarkDeepEqualInt(b *testing.B) {
	a := 42
	c := 42
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = deepEqual(a, c)
	}
}

// BenchmarkDeepEqualSlice measures deepEqual for slices, which falls
// back to reflect.DeepEqual (the slow path).
func BenchmarkDeepEqualSlice(b *testing.B) {
	a := []int{1, 2, 3, 4, 5}
	c := []int{1, 2, 3, 4, 5}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = deepEqual(a, c)
	}
}

// BenchmarkComputedAutoTrack measures computed value with automatic
// dependency tracking (no explicit deps), exercising trackDependencies.
func BenchmarkComputedAutoTrack(b *testing.B) {
	s := NewSignal(10)
	s.SetEqualFunc(EqualComparable[int])
	// No explicit deps — uses auto-tracking
	c := NewComputed(func() int { return s.Get() * 2 })
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Set(i)
		_ = c.Get()
	}
}

// BenchmarkHistoryPushUndo measures History push/undo cycles, exercising
// the branching history tree and cachedSize invalidation.
func BenchmarkHistoryPushUndo(b *testing.B) {
	h := NewHistory(0, WithMaxDepth(1000))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Push(i)
		h.Undo()
	}
}
