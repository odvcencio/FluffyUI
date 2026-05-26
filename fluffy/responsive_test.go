package fluffy

import (
	"strings"
	"testing"

	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/widgets"
)

// testLabel creates a simple label widget for testing that renders a known string.
func testLabel(text string) runtime.Widget {
	return widgets.NewLabel(text)
}

func renderToString(w runtime.Widget, width, height int) string {
	buf := runtime.NewBuffer(width, height)
	c := runtime.Constraints{MaxWidth: width, MaxHeight: height}
	w.Measure(c)
	w.Layout(runtime.Rect{X: 0, Y: 0, Width: width, Height: height})
	w.Render(runtime.RenderContext{Buffer: buf})

	var sb strings.Builder
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			cell := buf.Get(x, y)
			if cell.Rune == 0 {
				sb.WriteRune(' ')
			} else {
				sb.WriteRune(cell.Rune)
			}
		}
		if y < height-1 {
			sb.WriteRune('\n')
		}
	}
	return sb.String()
}

func TestResponsive_PicksCorrectBreakpoint(t *testing.T) {
	wide := testLabel("WIDE")
	medium := testLabel("MEDIUM")
	narrow := testLabel("NARROW")

	w := Responsive(
		Breakpoint{MinWidth: 100, Widget: wide},
		Breakpoint{MinWidth: 60, Widget: medium},
		Breakpoint{MinWidth: 0, Widget: narrow},
	)

	// Width 120 should pick the 100+ breakpoint (wide)
	output := renderToString(w, 120, 1)
	if !strings.Contains(output, "WIDE") {
		t.Errorf("at width 120, expected WIDE, got %q", output)
	}

	// Width 80 should pick the 60+ breakpoint (medium)
	output = renderToString(w, 80, 1)
	if !strings.Contains(output, "MEDIUM") {
		t.Errorf("at width 80, expected MEDIUM, got %q", output)
	}

	// Width 40 should pick the 0+ breakpoint (narrow)
	output = renderToString(w, 40, 1)
	if !strings.Contains(output, "NARROW") {
		t.Errorf("at width 40, expected NARROW, got %q", output)
	}
}

func TestResponsive_FallbackToSmallest(t *testing.T) {
	big := testLabel("BIG")
	small := testLabel("SMALL")

	w := Responsive(
		Breakpoint{MinWidth: 200, Widget: big},
		Breakpoint{MinWidth: 100, Widget: small},
	)

	// Width 50 is less than the smallest breakpoint (100), should fall back to
	// the last (smallest MinWidth) breakpoint.
	output := renderToString(w, 50, 1)
	if !strings.Contains(output, "SMALL") {
		t.Errorf("at width 50, expected SMALL fallback, got %q", output)
	}
}

func TestResponsive_SingleBreakpoint(t *testing.T) {
	only := testLabel("ONLY")

	w := Responsive(
		Breakpoint{MinWidth: 0, Widget: only},
	)

	output := renderToString(w, 80, 1)
	if !strings.Contains(output, "ONLY") {
		t.Errorf("single breakpoint should always match, got %q", output)
	}

	output = renderToString(w, 10, 1)
	if !strings.Contains(output, "ONLY") {
		t.Errorf("single breakpoint should match at small width, got %q", output)
	}
}

func TestResponsive_EmptyBreakpoints(t *testing.T) {
	w := Responsive()
	// Should not panic; returns a valid zero-size widget.
	size := w.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 24})
	if size.Width != 0 && size.Height != 0 {
		// SimpleWidget with no MeasureFunc returns MinSize (0,0).
		t.Logf("empty responsive widget size: %dx%d", size.Width, size.Height)
	}
}

func TestResponsive_MeasureDelegatesToCorrectChild(t *testing.T) {
	// Create widgets with different natural sizes.
	wideWidget := widgets.NewSimpleWidget()
	wideWidget.MeasureFunc = func(c runtime.Constraints) runtime.Size {
		return c.Constrain(runtime.Size{Width: 100, Height: 5})
	}

	narrowWidget := widgets.NewSimpleWidget()
	narrowWidget.MeasureFunc = func(c runtime.Constraints) runtime.Size {
		return c.Constrain(runtime.Size{Width: 30, Height: 10})
	}

	w := Responsive(
		Breakpoint{MinWidth: 80, Widget: wideWidget},
		Breakpoint{MinWidth: 0, Widget: narrowWidget},
	)

	// With a wide constraint, should delegate to wideWidget.
	size := w.Measure(runtime.Constraints{MaxWidth: 120, MaxHeight: 24})
	if size.Width != 100 || size.Height != 5 {
		t.Errorf("wide measure = %dx%d, want 100x5", size.Width, size.Height)
	}

	// With a narrow constraint, should delegate to narrowWidget.
	size = w.Measure(runtime.Constraints{MaxWidth: 40, MaxHeight: 24})
	if size.Width != 30 || size.Height != 10 {
		t.Errorf("narrow measure = %dx%d, want 30x10", size.Width, size.Height)
	}
}

func TestResponsive_BreakpointOrderDoesNotMatter(t *testing.T) {
	wide := testLabel("WIDE")
	narrow := testLabel("NARROW")

	// Pass breakpoints in ascending order (not the typical descending).
	w := Responsive(
		Breakpoint{MinWidth: 0, Widget: narrow},
		Breakpoint{MinWidth: 80, Widget: wide},
	)

	output := renderToString(w, 120, 1)
	if !strings.Contains(output, "WIDE") {
		t.Errorf("at width 120, expected WIDE regardless of input order, got %q", output)
	}

	output = renderToString(w, 40, 1)
	if !strings.Contains(output, "NARROW") {
		t.Errorf("at width 40, expected NARROW regardless of input order, got %q", output)
	}
}

func TestIfWide_SwitchesCorrectly(t *testing.T) {
	wide := testLabel("HORIZONTAL")
	narrow := testLabel("VERTICAL")

	w := IfWide(80, wide, narrow)

	// Wide
	output := renderToString(w, 100, 1)
	if !strings.Contains(output, "HORIZONTAL") {
		t.Errorf("at width 100 >= 80, expected HORIZONTAL, got %q", output)
	}

	// Narrow
	output = renderToString(w, 60, 1)
	if !strings.Contains(output, "VERTICAL") {
		t.Errorf("at width 60 < 80, expected VERTICAL, got %q", output)
	}

	// Exactly at threshold
	output = renderToString(w, 80, 1)
	if !strings.Contains(output, "HORIZONTAL") {
		t.Errorf("at width 80 == threshold, expected HORIZONTAL (>=), got %q", output)
	}
}

func TestAdaptiveStack_HorizontalVerticalSwitching(t *testing.T) {
	a := testLabel("A")
	b := testLabel("B")

	w := AdaptiveStack(40, a, b)

	// Wide: should use HStack (horizontal layout).
	// At width 80, both labels should appear on the same line.
	output := renderToString(w, 80, 3)
	lines := strings.Split(output, "\n")
	firstLine := lines[0]
	if !strings.Contains(firstLine, "A") || !strings.Contains(firstLine, "B") {
		t.Errorf("at width 80, expected both A and B on first line (HStack), got %q", firstLine)
	}

	// Narrow: should use VStack (vertical layout).
	// At width 20, labels should appear on separate lines.
	output = renderToString(w, 20, 3)
	lines = strings.Split(output, "\n")
	if len(lines) < 2 {
		t.Fatalf("at width 20, expected at least 2 lines (VStack), got %d lines", len(lines))
	}
	if !strings.Contains(lines[0], "A") {
		t.Errorf("at width 20, expected A on first line, got %q", lines[0])
	}
	if !strings.Contains(lines[1], "B") {
		t.Errorf("at width 20, expected B on second line, got %q", lines[1])
	}
}

func TestResponsive_HandleMessageDelegates(t *testing.T) {
	handled := false
	child := widgets.NewSimpleWidget()
	child.MeasureFunc = func(c runtime.Constraints) runtime.Size {
		return runtime.Size{Width: 10, Height: 1}
	}
	child.HandleMessageFunc = func(msg runtime.Message) runtime.HandleResult {
		handled = true
		return runtime.Handled()
	}

	w := Responsive(
		Breakpoint{MinWidth: 0, Widget: child},
	)

	// Measure first to set the active breakpoint.
	w.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 24})

	result := w.HandleMessage(runtime.CustomMsg{})
	if !result.Handled {
		t.Error("expected message to be handled")
	}
	if !handled {
		t.Error("expected child HandleMessage to be called")
	}
}

func TestResponsive_NilWidget(t *testing.T) {
	w := Responsive(
		Breakpoint{MinWidth: 0, Widget: nil},
	)

	// Should not panic with a nil widget breakpoint.
	size := w.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 24})
	if size.Width != 0 || size.Height != 0 {
		t.Logf("nil widget breakpoint size: %dx%d", size.Width, size.Height)
	}

	w.Layout(runtime.Rect{X: 0, Y: 0, Width: 80, Height: 24})
	w.Render(runtime.RenderContext{Buffer: runtime.NewBuffer(80, 24)})
	result := w.HandleMessage(runtime.CustomMsg{})
	if result.Handled {
		t.Error("nil widget should not handle messages")
	}
}

func TestResponsive_ExactBoundary(t *testing.T) {
	at80 := testLabel("AT80")
	below := testLabel("BELOW")

	w := Responsive(
		Breakpoint{MinWidth: 80, Widget: at80},
		Breakpoint{MinWidth: 0, Widget: below},
	)

	// Exactly at 80
	output := renderToString(w, 80, 1)
	if !strings.Contains(output, "AT80") {
		t.Errorf("at width 80, expected AT80, got %q", output)
	}

	// One below
	output = renderToString(w, 79, 1)
	if !strings.Contains(output, "BELOW") {
		t.Errorf("at width 79, expected BELOW, got %q", output)
	}
}
