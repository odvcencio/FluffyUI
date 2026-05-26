package main

import (
	"strings"
	"testing"

	"m31labs.dev/fluffyui/runtime"
	fluffytest "m31labs.dev/fluffyui/testing"
	"m31labs.dev/fluffyui/widgets"
)

func TestShowcaseTabRender(t *testing.T) {
	simple := widgets.NewScrollView(widgets.NewLabel("Kitchen Sink"))
	simpleOut := fluffytest.RenderToString(simple, 40, 5)
	if !strings.Contains(simpleOut, "Kitchen Sink") {
		t.Fatalf("expected basic scroll view to render\n\nScroll:\n%s", simpleOut)
	}

	showcase := NewShowcaseTabContent()
	const width, height = 120, 40
	constraints := runtime.Constraints{MaxWidth: width, MaxHeight: height}
	if size := showcase.Measure(constraints); size.Width <= 0 || size.Height <= 0 {
		t.Fatalf("expected showcase measure to be non-zero, got %+v", size)
	}
	showcase.Layout(runtime.Rect{X: 0, Y: 0, Width: width, Height: height})

	if showcase.scroll == nil {
		t.Fatalf("expected showcase scroll view to be initialized")
	}
	if size := showcase.scroll.Measure(constraints); size.Width <= 0 || size.Height <= 0 {
		t.Fatalf("expected scroll measure to be non-zero, got %+v", size)
	}
	if bounds := showcase.scroll.Bounds(); bounds.Width <= 0 || bounds.Height <= 0 {
		t.Fatalf("expected scroll bounds to be laid out, got %+v", bounds)
	}
	if showcase.content == nil {
		t.Fatalf("expected showcase content to be initialized")
	}
	contentSize := showcase.content.Measure(constraints)
	if contentSize.Width <= 0 || contentSize.Height <= 0 {
		t.Fatalf("expected showcase content measure to be non-zero, got %+v", contentSize)
	}
	if scrollSize := showcase.scroll.ContentSize(); scrollSize.Width <= 0 || scrollSize.Height <= 0 {
		t.Fatalf("expected scroll content size to be non-zero, got %+v (content: %+v)", scrollSize, contentSize)
	}

	buf := runtime.NewBuffer(width, height)
	showcase.Render(runtime.RenderContext{Buffer: buf, Bounds: runtime.Rect{X: 0, Y: 0, Width: width, Height: height}})
	nonSpace := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			cell := buf.Get(x, y)
			if cell.Rune != 0 && cell.Rune != ' ' {
				nonSpace++
				break
			}
		}
		if nonSpace > 0 {
			break
		}
	}
	output := fluffytest.RenderToString(showcase, width, height)
	if !strings.Contains(output, "Kitchen Sink") {
		alt := fluffytest.RenderToString(showcase.scroll, width, height)
		t.Fatalf("expected Kitchen Sink showcase to render (non-space cells: %d)\n\nShowcase:\n%s\n\nScroll:\n%s", nonSpace, output, alt)
	}
}
