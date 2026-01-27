package runtime

import "testing"

func TestWidgetPoolAcquireRelease(t *testing.T) {
	created := 0
	pool := NewWidgetPool(func() *testWidgetPoolItem {
		created++
		return &testWidgetPoolItem{}
	}, func(item *testWidgetPoolItem) {
		item.value = 0
	}, 4)

	item := pool.Acquire()
	if item == nil {
		t.Fatal("expected non-nil item")
	}
	item.value = 99
	pool.Release(item)

	if pool.Size() != 1 {
		t.Fatalf("expected pool size 1, got %d", pool.Size())
	}

	item2 := pool.Acquire()
	if item2 == nil {
		t.Fatal("expected non-nil item on acquire")
	}
	if item2.value != 0 {
		t.Fatalf("expected reset value 0, got %d", item2.value)
	}
	if created < 1 {
		t.Fatalf("expected at least 1 created item, got %d", created)
	}
}

func TestWidgetPoolMaxSize(t *testing.T) {
	pool := NewWidgetPool(func() *testWidgetPoolItem {
		return &testWidgetPoolItem{}
	}, nil, 2)

	pool.Release(&testWidgetPoolItem{})
	pool.Release(&testWidgetPoolItem{})
	pool.Release(&testWidgetPoolItem{})

	if pool.Size() != 2 {
		t.Fatalf("expected pool size 2, got %d", pool.Size())
	}
}

type testWidgetPoolItem struct {
	value int
}
