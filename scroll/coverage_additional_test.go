package scroll

import (
	"image"
	"testing"

	"github.com/odvcencio/fluffyui/runtime"
)

func TestViewportOffsetAndOnChange(t *testing.T) {
	v := NewViewport(nil)
	v.SetViewSize(runtime.Size{Width: 5, Height: 5})
	v.SetContentSize(runtime.Size{Width: 10, Height: 8})

	calls := 0
	var last image.Point
	v.SetOnChange(func(offset image.Point, content runtime.Size, view runtime.Size) {
		calls++
		last = offset
	})

	v.SetOffset(100, 100)
	if last != (image.Point{X: 5, Y: 3}) {
		t.Fatalf("offset = %+v, want {5 3}", last)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}

	v.SetOffset(100, 100)
	if calls != 1 {
		t.Fatalf("expected no additional call when offset unchanged")
	}

	v.ScrollToStart()
	if last != (image.Point{X: 0, Y: 0}) {
		t.Fatalf("offset = %+v, want {0 0}", last)
	}
}

func TestViewportPaging(t *testing.T) {
	v := NewViewport(nil)
	v.SetViewSize(runtime.Size{Width: 0, Height: 0})
	v.SetContentSize(runtime.Size{Width: 100, Height: 100})

	v.PageBy(2)
	if v.Offset() != (image.Point{X: 0, Y: 2}) {
		t.Fatalf("offset = %+v, want {0 2}", v.Offset())
	}

	v.ScrollToEnd()
	if v.Offset() != v.MaxOffset() {
		t.Fatalf("expected scroll to end")
	}
}
