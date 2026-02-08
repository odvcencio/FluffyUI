package runtime

import (
	"testing"
	"time"
)

func TestMeasureCache_HitSameConstraintsCurrentGen(t *testing.T) {
	cache := NewMeasureCache()
	w := newTestWidget(50, 30)
	constraints := Loose(100, 100)
	gen := cache.Generation()

	cache.Set(w, constraints, Size{Width: 50, Height: 30}, gen)

	got, ok := cache.Get(w, constraints, gen)
	if !ok {
		t.Fatal("expected cache hit for same constraints and current generation")
	}
	if got != (Size{Width: 50, Height: 30}) {
		t.Errorf("cached size = %v, want {50, 30}", got)
	}
}

func TestMeasureCache_HitPreviousGeneration(t *testing.T) {
	cache := NewMeasureCache()
	w := newTestWidget(50, 30)
	constraints := Loose(100, 100)
	gen := cache.Generation()

	cache.Set(w, constraints, Size{Width: 50, Height: 30}, gen)

	// Bump to next generation; entry from gen should still be valid at gen+1
	newGen := cache.BumpGeneration()
	got, ok := cache.Get(w, constraints, newGen)
	if !ok {
		t.Fatal("expected cache hit for previous generation (gen-1)")
	}
	if got != (Size{Width: 50, Height: 30}) {
		t.Errorf("cached size = %v, want {50, 30}", got)
	}
}

func TestMeasureCache_MissConstraintsChanged(t *testing.T) {
	cache := NewMeasureCache()
	w := newTestWidget(50, 30)
	gen := cache.Generation()

	cache.Set(w, Loose(100, 100), Size{Width: 50, Height: 30}, gen)

	// Different constraints should miss
	_, ok := cache.Get(w, Loose(200, 200), gen)
	if ok {
		t.Fatal("expected cache miss when constraints changed")
	}
}

func TestMeasureCache_MissGenerationTooOld(t *testing.T) {
	cache := NewMeasureCache()
	w := newTestWidget(50, 30)
	gen := cache.Generation()

	cache.Set(w, Loose(100, 100), Size{Width: 50, Height: 30}, gen)

	// Bump twice: entry at gen is now older than gen+2 - 1
	cache.BumpGeneration()
	twoAhead := cache.BumpGeneration()

	_, ok := cache.Get(w, Loose(100, 100), twoAhead)
	if ok {
		t.Fatal("expected cache miss when generation is too old (> gen-1)")
	}
}

func TestMeasureCache_MissUnknownWidget(t *testing.T) {
	cache := NewMeasureCache()
	w := newTestWidget(50, 30)
	gen := cache.Generation()

	_, ok := cache.Get(w, Loose(100, 100), gen)
	if ok {
		t.Fatal("expected cache miss for widget never stored")
	}
}

func TestMeasureCache_InvalidateRemovesOnlyTarget(t *testing.T) {
	cache := NewMeasureCache()
	w1 := newTestWidget(50, 30)
	w2 := newTestWidget(60, 40)
	constraints := Loose(100, 100)
	gen := cache.Generation()

	cache.Set(w1, constraints, Size{Width: 50, Height: 30}, gen)
	cache.Set(w2, constraints, Size{Width: 60, Height: 40}, gen)

	cache.Invalidate(w1)

	_, ok1 := cache.Get(w1, constraints, gen)
	if ok1 {
		t.Fatal("expected cache miss for invalidated widget w1")
	}

	got, ok2 := cache.Get(w2, constraints, gen)
	if !ok2 {
		t.Fatal("expected cache hit for w2 after invalidating w1")
	}
	if got != (Size{Width: 60, Height: 40}) {
		t.Errorf("w2 cached size = %v, want {60, 40}", got)
	}
}

func TestMeasureCache_Clear(t *testing.T) {
	cache := NewMeasureCache()
	w1 := newTestWidget(50, 30)
	w2 := newTestWidget(60, 40)
	constraints := Loose(100, 100)
	gen := cache.Generation()

	cache.Set(w1, constraints, Size{Width: 50, Height: 30}, gen)
	cache.Set(w2, constraints, Size{Width: 60, Height: 40}, gen)

	cache.Clear()

	if cache.Len() != 0 {
		t.Errorf("cache length after Clear = %d, want 0", cache.Len())
	}

	_, ok1 := cache.Get(w1, constraints, gen)
	_, ok2 := cache.Get(w2, constraints, gen)
	if ok1 || ok2 {
		t.Fatal("expected cache miss for all widgets after Clear")
	}
}

func TestMeasureCache_BumpGenerationIncrements(t *testing.T) {
	cache := NewMeasureCache()
	initial := cache.Generation()

	g1 := cache.BumpGeneration()
	if g1 != initial+1 {
		t.Errorf("first bump = %d, want %d", g1, initial+1)
	}

	g2 := cache.BumpGeneration()
	if g2 != initial+2 {
		t.Errorf("second bump = %d, want %d", g2, initial+2)
	}

	g3 := cache.BumpGeneration()
	if g3 != initial+3 {
		t.Errorf("third bump = %d, want %d", g3, initial+3)
	}
}

func TestMeasureCache_Len(t *testing.T) {
	cache := NewMeasureCache()
	if cache.Len() != 0 {
		t.Errorf("new cache Len = %d, want 0", cache.Len())
	}

	w1 := newTestWidget(50, 30)
	w2 := newTestWidget(60, 40)
	gen := cache.Generation()

	cache.Set(w1, Loose(100, 100), Size{Width: 50, Height: 30}, gen)
	if cache.Len() != 1 {
		t.Errorf("Len after 1 Set = %d, want 1", cache.Len())
	}

	cache.Set(w2, Loose(100, 100), Size{Width: 60, Height: 40}, gen)
	if cache.Len() != 2 {
		t.Errorf("Len after 2 Sets = %d, want 2", cache.Len())
	}

	cache.Invalidate(w1)
	if cache.Len() != 1 {
		t.Errorf("Len after Invalidate = %d, want 1", cache.Len())
	}
}

func TestMeasureCache_OverwriteEntry(t *testing.T) {
	cache := NewMeasureCache()
	w := newTestWidget(50, 30)
	gen := cache.Generation()

	cache.Set(w, Loose(100, 100), Size{Width: 50, Height: 30}, gen)
	cache.Set(w, Loose(200, 200), Size{Width: 100, Height: 60}, gen)

	// Old constraints should miss
	_, ok := cache.Get(w, Loose(100, 100), gen)
	if ok {
		t.Fatal("expected cache miss for old constraints after overwrite")
	}

	// New constraints should hit
	got, ok := cache.Get(w, Loose(200, 200), gen)
	if !ok {
		t.Fatal("expected cache hit for new constraints after overwrite")
	}
	if got != (Size{Width: 100, Height: 60}) {
		t.Errorf("cached size = %v, want {100, 60}", got)
	}
}

func TestMeasureCache_NilReceiver(t *testing.T) {
	var cache *MeasureCache
	w := newTestWidget(50, 30)

	// None of these should panic
	_, ok := cache.Get(w, Loose(100, 100), 1)
	if ok {
		t.Fatal("nil cache Get should return false")
	}

	cache.Set(w, Loose(100, 100), Size{Width: 50, Height: 30}, 1)
	cache.Invalidate(w)
	cache.Clear()

	gen := cache.BumpGeneration()
	if gen != 0 {
		t.Errorf("nil cache BumpGeneration = %d, want 0", gen)
	}

	if cache.Generation() != 0 {
		t.Errorf("nil cache Generation = %d, want 0", cache.Generation())
	}

	if cache.Len() != 0 {
		t.Errorf("nil cache Len = %d, want 0", cache.Len())
	}
}

func TestMeasureCache_TightConstraintsHit(t *testing.T) {
	cache := NewMeasureCache()
	w := newTestWidget(50, 30)
	constraints := Tight(50, 30)
	gen := cache.Generation()

	cache.Set(w, constraints, Size{Width: 50, Height: 30}, gen)

	got, ok := cache.Get(w, constraints, gen)
	if !ok {
		t.Fatal("expected cache hit for tight constraints")
	}
	if got != (Size{Width: 50, Height: 30}) {
		t.Errorf("cached size = %v, want {50, 30}", got)
	}
}

// slowWidget sleeps during Measure to simulate an expensive measurement.
type slowWidget struct {
	testWidget
	delay time.Duration
}

func (s *slowWidget) Measure(constraints Constraints) Size {
	time.Sleep(s.delay)
	return constraints.Constrain(s.preferredSize)
}

func BenchmarkMeasureCache_Lookup(b *testing.B) {
	cache := NewMeasureCache()
	w := &slowWidget{
		testWidget: testWidget{preferredSize: Size{Width: 50, Height: 30}},
		delay:      100 * time.Microsecond,
	}
	constraints := Loose(100, 100)
	gen := cache.Generation()
	result := w.Measure(constraints)
	cache.Set(w, constraints, result, gen)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got, ok := cache.Get(w, constraints, gen)
		if !ok {
			b.Fatal("expected cache hit")
		}
		_ = got
	}
}

func BenchmarkMeasureCache_Uncached(b *testing.B) {
	w := &slowWidget{
		testWidget: testWidget{preferredSize: Size{Width: 50, Height: 30}},
		delay:      100 * time.Microsecond,
	}
	constraints := Loose(100, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got := w.Measure(constraints)
		_ = got
	}
}

func BenchmarkMeasureCache_SetAndGet(b *testing.B) {
	cache := NewMeasureCache()
	w := newTestWidget(50, 30)
	constraints := Loose(100, 100)
	gen := cache.Generation()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(w, constraints, Size{Width: 50, Height: 30}, gen)
		got, ok := cache.Get(w, constraints, gen)
		if !ok {
			b.Fatal("expected cache hit")
		}
		_ = got
	}
}
