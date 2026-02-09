package widgets

import (
	"strings"
	"testing"

	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
	"github.com/odvcencio/fluffyui/theme"
)

func TestNewMarkdownViewer(t *testing.T) {
	m := NewMarkdownViewer("# Hello")
	if m == nil {
		t.Fatal("expected non-nil MarkdownViewer")
	}
	if m.Content() != "# Hello" {
		t.Fatalf("expected content '# Hello', got %q", m.Content())
	}
}

func TestMarkdownViewer_NilSafety(t *testing.T) {
	var m *MarkdownViewer
	// These should not panic.
	m.SetContent("x")
	m.ScrollBy(0, 1)
	m.ScrollTo(0, 0)
	m.PageBy(1)
	m.ScrollToEnd()
	m.Render(runtime.RenderContext{})
	if m.Content() != "" {
		t.Fatal("expected empty content for nil receiver")
	}
	result := m.HandleMessage(runtime.KeyMsg{})
	if result.Handled {
		t.Fatal("nil receiver should return unhandled")
	}
}

func TestMarkdownViewer_SetContent(t *testing.T) {
	m := NewMarkdownViewer("# Old")
	m.SetContent("# New")
	if m.Content() != "# New" {
		t.Fatalf("expected content '# New', got %q", m.Content())
	}
}

func TestMarkdownViewer_SetContentEmpty(t *testing.T) {
	m := NewMarkdownViewer("# Something")
	m.SetContent("")
	if m.Content() != "" {
		t.Fatalf("expected empty content, got %q", m.Content())
	}
	if len(m.lines) != 0 {
		t.Fatalf("expected 0 styled lines for empty content, got %d", len(m.lines))
	}
}

func TestMarkdownViewer_Measure(t *testing.T) {
	m := NewMarkdownViewer("Hello world")
	size := m.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 40})
	if size.Width <= 0 {
		t.Fatalf("expected positive width, got %d", size.Width)
	}
	if size.Height <= 0 {
		t.Fatalf("expected positive height, got %d", size.Height)
	}
}

func TestMarkdownViewer_MeasureMultiline(t *testing.T) {
	content := "# Heading\n\nParagraph one.\n\nParagraph two."
	m := NewMarkdownViewer(content)
	size := m.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 100})
	// The heading, spacer, two paragraphs, and spacer between them should produce multiple lines.
	if size.Height < 3 {
		t.Fatalf("expected at least 3 lines for multi-paragraph content, got %d", size.Height)
	}
}

func TestMarkdownViewer_HandleMessage_Scroll(t *testing.T) {
	// Create content that is taller than the view to enable scrolling.
	lines := make([]string, 50)
	for i := range lines {
		lines[i] = "line"
	}
	content := strings.Join(lines, "\n\n")
	m := NewMarkdownViewer(content)

	// Layout with a small viewport so content overflows.
	constraints := runtime.Constraints{MaxWidth: 40, MaxHeight: 5}
	m.Measure(constraints)
	m.Layout(runtime.Rect{X: 0, Y: 0, Width: 40, Height: 5})

	// Focus the widget so key events are handled.
	m.Focus()

	// Scroll down.
	result := m.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if !result.Handled {
		t.Fatal("expected KeyDown to be handled")
	}
	if m.offset != 1 {
		t.Fatalf("expected offset 1 after KeyDown, got %d", m.offset)
	}

	// Scroll up.
	result = m.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if !result.Handled {
		t.Fatal("expected KeyUp to be handled")
	}
	if m.offset != 0 {
		t.Fatalf("expected offset 0 after KeyUp, got %d", m.offset)
	}

	// Home.
	m.ScrollBy(0, 10)
	result = m.HandleMessage(runtime.KeyMsg{Key: terminal.KeyHome})
	if !result.Handled {
		t.Fatal("expected KeyHome to be handled")
	}
	if m.offset != 0 {
		t.Fatalf("expected offset 0 after Home, got %d", m.offset)
	}

	// End.
	result = m.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnd})
	if !result.Handled {
		t.Fatal("expected KeyEnd to be handled")
	}
	if m.offset <= 0 {
		t.Fatal("expected positive offset after End")
	}

	// Page down from top.
	m.ScrollToStart()
	result = m.HandleMessage(runtime.KeyMsg{Key: terminal.KeyPageDown})
	if !result.Handled {
		t.Fatal("expected KeyPageDown to be handled")
	}
	prevOffset := m.offset

	// Page up.
	result = m.HandleMessage(runtime.KeyMsg{Key: terminal.KeyPageUp})
	if !result.Handled {
		t.Fatal("expected KeyPageUp to be handled")
	}
	if m.offset >= prevOffset {
		t.Fatal("expected offset to decrease after PageUp")
	}
}

func TestMarkdownViewer_HandleMessage_Unfocused(t *testing.T) {
	m := NewMarkdownViewer("# Test")
	// Widget is not focused; key events should be unhandled.
	result := m.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if result.Handled {
		t.Fatal("expected unhandled when not focused")
	}
}

func TestMarkdownViewer_HandleMessage_Mouse(t *testing.T) {
	lines := make([]string, 50)
	for i := range lines {
		lines[i] = "line"
	}
	content := strings.Join(lines, "\n\n")
	m := NewMarkdownViewer(content)
	m.Measure(runtime.Constraints{MaxWidth: 40, MaxHeight: 5})
	m.Layout(runtime.Rect{X: 0, Y: 0, Width: 40, Height: 5})

	result := m.HandleMessage(runtime.MouseMsg{Button: runtime.MouseWheelDown})
	if !result.Handled {
		t.Fatal("expected mouse wheel down to be handled")
	}
	if m.offset <= 0 {
		t.Fatal("expected positive offset after mouse wheel down")
	}

	prev := m.offset
	result = m.HandleMessage(runtime.MouseMsg{Button: runtime.MouseWheelUp})
	if !result.Handled {
		t.Fatal("expected mouse wheel up to be handled")
	}
	if m.offset >= prev {
		t.Fatal("expected offset to decrease after mouse wheel up")
	}
}

func TestMarkdownViewer_AccessibilityRole(t *testing.T) {
	m := NewMarkdownViewer("# Hello")
	if m.AccessibleRole() != "document" {
		t.Fatalf("expected role 'document', got %q", m.AccessibleRole())
	}
}

func TestMarkdownViewer_AccessibilityLabel(t *testing.T) {
	m := NewMarkdownViewer("# Hello")
	if m.AccessibleLabel() != "Markdown viewer" {
		t.Fatalf("expected label 'Markdown viewer', got %q", m.AccessibleLabel())
	}
}

func TestMarkdownViewer_CanFocus(t *testing.T) {
	m := NewMarkdownViewer("# Hello")
	if !m.CanFocus() {
		t.Fatal("expected MarkdownViewer to be focusable")
	}
}

func TestMarkdownViewer_StyleType(t *testing.T) {
	m := NewMarkdownViewer("test")
	if m.StyleType() != "MarkdownViewer" {
		t.Fatalf("expected style type 'MarkdownViewer', got %q", m.StyleType())
	}
}

func TestMarkdownViewer_WithMarkdownTheme(t *testing.T) {
	th := theme.DefaultTheme()
	m := NewMarkdownViewer("# Themed", WithMarkdownTheme(th))
	if m == nil {
		t.Fatal("expected non-nil MarkdownViewer with theme option")
	}
	if m.renderer == nil {
		t.Fatal("expected non-nil renderer after theme option")
	}
}

func TestMarkdownViewer_WithMarkdownMaxWidth(t *testing.T) {
	m := NewMarkdownViewer("Hello world", WithMarkdownMaxWidth(40))
	if m.maxWidth != 40 {
		t.Fatalf("expected maxWidth 40, got %d", m.maxWidth)
	}
	size := m.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 40})
	if size.Width > 40 {
		t.Fatalf("expected width <= 40 with maxWidth=40, got %d", size.Width)
	}
}

func TestMarkdownViewer_WithWordWrap(t *testing.T) {
	m := NewMarkdownViewer("Hello world", WithWordWrap(false))
	if m.wordWrap {
		t.Fatal("expected wordWrap to be false")
	}
}

func TestMarkdownViewer_ScrollInterfaces(t *testing.T) {
	m := NewMarkdownViewer("test")
	// Verify scroll.Controller interface is implemented.
	var _ interface {
		ScrollBy(dx, dy int)
		ScrollTo(x, y int)
		PageBy(pages int)
		ScrollToStart()
		ScrollToEnd()
	} = m
}

func TestMarkdownViewer_Render(t *testing.T) {
	m := NewMarkdownViewer("**bold text**")
	buf := runtime.NewBuffer(40, 5)
	m.Measure(runtime.Constraints{MaxWidth: 40, MaxHeight: 5})
	m.Layout(runtime.Rect{X: 0, Y: 0, Width: 40, Height: 5})
	ctx := runtime.RenderContext{Buffer: buf}
	m.Render(ctx)
	// The rendered output should contain "bold text" somewhere.
	found := false
	cells := buf.Cells()
	w, h := buf.Size()
	for y := 0; y < h; y++ {
		var sb strings.Builder
		for x := 0; x < w; x++ {
			cell := cells[y*w+x]
			if cell.Rune != 0 {
				sb.WriteRune(cell.Rune)
			}
		}
		if strings.Contains(sb.String(), "bold text") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected rendered output to contain 'bold text'")
	}
}

func TestMarkdownViewer_ClampOffset(t *testing.T) {
	m := NewMarkdownViewer("short")
	m.Measure(runtime.Constraints{MaxWidth: 40, MaxHeight: 10})
	m.Layout(runtime.Rect{X: 0, Y: 0, Width: 40, Height: 10})

	// Try to scroll past the end; offset should be clamped.
	m.ScrollBy(0, 100)
	if m.offset < 0 {
		t.Fatalf("expected non-negative offset, got %d", m.offset)
	}

	// Try to scroll before the start; offset should be clamped.
	m.ScrollBy(0, -200)
	if m.offset != 0 {
		t.Fatalf("expected offset 0 after scrolling up from top, got %d", m.offset)
	}
}
