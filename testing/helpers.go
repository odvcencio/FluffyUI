// Package testing provides test utilities for FluffyUI applications.
//
// This package bridges the simulation backend with the runtime package,
// providing convenient helpers for widget testing.
package testing

import (
	"strings"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/backend/sim"
	"github.com/odvcencio/fluffyui/runtime"
)

// RenderToString renders a widget to a string without requiring a backend.
// This is useful for snapshot testing and simple output verification.
func RenderToString(w runtime.Widget, width, height int) string {
	buf := runtime.NewBuffer(width, height)

	constraints := runtime.Constraints{MaxWidth: width, MaxHeight: height}
	w.Measure(constraints)
	w.Layout(runtime.Rect{X: 0, Y: 0, Width: width, Height: height})

	ctx := runtime.RenderContext{Buffer: buf}
	w.Render(ctx)

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

// RenderTo renders a widget to a simulation backend.
// The widget is measured, laid out, and rendered to the backend's buffer.
func RenderTo(be *sim.Backend, w runtime.Widget, width, height int) {
	buf := runtime.NewBuffer(width, height)

	constraints := runtime.Constraints{MaxWidth: width, MaxHeight: height}
	w.Measure(constraints)
	w.Layout(runtime.Rect{X: 0, Y: 0, Width: width, Height: height})

	ctx := runtime.RenderContext{Buffer: buf}
	w.Render(ctx)

	// Copy buffer to backend
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			cell := buf.Get(x, y)
			r := cell.Rune
			if r == 0 {
				r = ' '
			}
			be.SetContent(x, y, r, nil, cell.Style)
		}
	}
	be.Show()
}

// RenderWidget renders a widget to a newly created simulation backend.
// Returns the backend for further inspection.
func RenderWidget(w runtime.Widget, width, height int) *sim.Backend {
	be := sim.New(width, height)
	if err := be.Init(); err != nil {
		panic("failed to init sim backend: " + err.Error())
	}
	RenderTo(be, w, width, height)
	return be
}

// AssertContains fails the test if the backend screen does not contain the text.
func AssertContains(t *testing.T, be *sim.Backend, text string) {
	t.Helper()
	if !be.ContainsText(text) {
		t.Errorf("expected screen to contain %q\nScreen:\n%s", text, be.Capture())
	}
}

// AssertNotContains fails the test if the backend screen contains the text.
func AssertNotContains(t *testing.T, be *sim.Backend, text string) {
	t.Helper()
	if be.ContainsText(text) {
		t.Errorf("expected screen to NOT contain %q\nScreen:\n%s", text, be.Capture())
	}
}

// AssertTextAt fails the test if the text is not at the given position.
func AssertTextAt(t *testing.T, be *sim.Backend, x, y int, text string) {
	t.Helper()
	region := be.CaptureRegion(x, y, len(text), 1)
	if region != text {
		t.Errorf("expected %q at (%d,%d), got %q\nScreen:\n%s", text, x, y, region, be.Capture())
	}
}

// AssertCellStyle fails the test if the cell at (x,y) doesn't have the expected style attributes.
func AssertCellStyle(t *testing.T, be *sim.Backend, x, y int, bold, italic, underline bool) {
	t.Helper()
	_, _, style := be.CaptureCell(x, y)
	attrs := style.Attributes()

	if bold && attrs&backend.AttrBold == 0 {
		t.Errorf("expected cell (%d,%d) to be bold", x, y)
	}
	if italic && attrs&backend.AttrItalic == 0 {
		t.Errorf("expected cell (%d,%d) to be italic", x, y)
	}
	if underline && attrs&backend.AttrUnderline == 0 {
		t.Errorf("expected cell (%d,%d) to be underlined", x, y)
	}
}

// NewAnnouncer returns an in-memory announcer for accessibility tests.
func NewAnnouncer() *accessibility.SimpleAnnouncer {
	return &accessibility.SimpleAnnouncer{}
}

// Announcements returns captured announcements for a SimpleAnnouncer.
func Announcements(announcer accessibility.Announcer) []accessibility.Announcement {
	if announcer == nil {
		return nil
	}
	if simple, ok := announcer.(*accessibility.SimpleAnnouncer); ok {
		return simple.History()
	}
	return nil
}

// AssertAnnounced fails the test if the announcer never emitted a matching message.
func AssertAnnounced(t *testing.T, announcer accessibility.Announcer, contains string) {
	t.Helper()
	history := Announcements(announcer)
	for _, entry := range history {
		if strings.Contains(entry.Message, contains) {
			return
		}
	}
	t.Errorf("expected announcement containing %q, got %v", contains, history)
}

// NewTestBackend creates an initialized simulation backend for testing.
// Returns the backend; callers should defer be.Fini().
func NewTestBackend(t *testing.T, width, height int) *sim.Backend {
	t.Helper()
	be := sim.New(width, height)
	if err := be.Init(); err != nil {
		t.Fatalf("failed to init sim backend: %v", err)
	}
	t.Cleanup(be.Fini)
	return be
}

// WaitForRender waits for the app to process messages and render.
// This is useful after injecting input events.
func WaitForRender(timeout time.Duration) {
	time.Sleep(timeout)
}

// SnapshotMismatch represents a difference between expected and actual output.
type SnapshotMismatch struct {
	Expected string
	Actual   string
}

// AssertSnapshot compares the rendered output against expected content.
// Returns a SnapshotMismatch if they differ, nil otherwise.
func AssertSnapshot(t *testing.T, w runtime.Widget, width, height int, expected string) {
	t.Helper()
	actual := RenderToString(w, width, height)
	if actual != expected {
		t.Errorf("snapshot mismatch:\n--- Expected ---\n%s\n--- Actual ---\n%s", expected, actual)
	}
}

// CaptureWidget captures a widget's rendered output as a string.
// Shorthand for RenderToString.
func CaptureWidget(w runtime.Widget, width, height int) string {
	return RenderToString(w, width, height)
}

// MeasureWidget measures a widget and returns its preferred size.
func MeasureWidget(w runtime.Widget, maxWidth, maxHeight int) runtime.Size {
	return w.Measure(runtime.Constraints{MaxWidth: maxWidth, MaxHeight: maxHeight})
}

// LayoutAndRender performs the full measure-layout-render cycle on a widget.
// Returns the buffer containing the rendered output.
func LayoutAndRender(w runtime.Widget, width, height int) *runtime.Buffer {
	buf := runtime.NewBuffer(width, height)

	constraints := runtime.Constraints{MaxWidth: width, MaxHeight: height}
	w.Measure(constraints)
	w.Layout(runtime.Rect{X: 0, Y: 0, Width: width, Height: height})
	w.Render(runtime.RenderContext{Buffer: buf})

	return buf
}

// GetCell returns the cell content at the given position after rendering.
func GetCell(w runtime.Widget, width, height, x, y int) runtime.Cell {
	buf := LayoutAndRender(w, width, height)
	return buf.Get(x, y)
}
