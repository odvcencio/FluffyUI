package testing

import (
	"strings"
	"testing"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/widgets"
)

func TestRenderToString(t *testing.T) {
	label := widgets.NewLabel("Hello")
	output := RenderToString(label, 10, 1)

	if !strings.Contains(output, "Hello") {
		t.Errorf("expected output to contain 'Hello', got %q", output)
	}
}

func TestRenderWidget(t *testing.T) {
	label := widgets.NewLabel("Test")
	be := RenderWidget(label, 20, 1)
	defer be.Fini()

	if !be.ContainsText("Test") {
		t.Error("expected backend to contain 'Test'")
	}
}

func TestNewTestBackend(t *testing.T) {
	be := NewTestBackend(t, 80, 24)

	w, h := be.Size()
	if w != 80 {
		t.Errorf("expected width 80, got %d", w)
	}
	if h < 24 || h > 25 {
		t.Errorf("expected height around 24, got %d", h)
	}
	// Cleanup is automatic via t.Cleanup
}

func TestAssertContains(t *testing.T) {
	be := NewTestBackend(t, 20, 1)

	style := backend.DefaultStyle()
	for i, r := range "Hello" {
		be.SetContent(i, 0, r, nil, style)
	}
	be.Show()

	// This should pass
	AssertContains(t, be, "Hello")
}

func TestAssertTextAt(t *testing.T) {
	be := NewTestBackend(t, 20, 1)

	style := backend.DefaultStyle()
	text := "World"
	for i, r := range text {
		be.SetContent(5+i, 0, r, nil, style)
	}
	be.Show()

	// This should pass
	AssertTextAt(t, be, 5, 0, "World")
}

func TestMeasureWidget(t *testing.T) {
	label := widgets.NewLabel("Testing")
	size := MeasureWidget(label, 100, 10)

	if size.Width < 7 {
		t.Errorf("expected width >= 7, got %d", size.Width)
	}
}

func TestLayoutAndRender(t *testing.T) {
	label := widgets.NewLabel("X")
	buf := LayoutAndRender(label, 10, 1)

	cell := buf.Get(0, 0)
	if cell.Rune != 'X' {
		t.Errorf("expected 'X' at (0,0), got %c", cell.Rune)
	}
}

func TestGetCell(t *testing.T) {
	label := widgets.NewLabel("ABC")
	cell := GetCell(label, 10, 1, 1, 0)

	if cell.Rune != 'B' {
		t.Errorf("expected 'B' at position 1, got %c", cell.Rune)
	}
}

// testWidget is a minimal widget for testing
type testWidget struct {
	widgets.Base
	content string
	style   backend.Style
}

func (w *testWidget) Measure(c runtime.Constraints) runtime.Size {
	return runtime.Size{Width: len(w.content), Height: 1}
}

func (w *testWidget) Render(ctx runtime.RenderContext) {
	bounds := w.Bounds()
	for i, r := range w.content {
		if i < bounds.Width {
			ctx.Buffer.Set(bounds.X+i, bounds.Y, r, w.style)
		}
	}
}

func TestRenderToString_CustomWidget(t *testing.T) {
	w := &testWidget{
		content: "Custom",
		style:   backend.DefaultStyle().Bold(true),
	}

	output := RenderToString(w, 20, 1)
	if !strings.Contains(output, "Custom") {
		t.Errorf("expected 'Custom' in output, got %q", output)
	}
}
