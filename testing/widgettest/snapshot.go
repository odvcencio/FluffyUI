//go:build !js

package widgettest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/odvcencio/fluffyui/runtime"
)

// UpdateSnapshots controls whether Snapshot overwrites golden files instead of
// comparing against them. Set to true (e.g. via TestMain or an init function
// that reads a flag/env var) to regenerate snapshots.
var UpdateSnapshots = false

// Snapshot captures the rendered output of a widget and compares it against
// a golden file. On first run (or when UpdateSnapshots is true), creates the
// golden file. On subsequent runs, compares against the golden file and fails
// the test if the output differs.
//
// Golden files are stored in testdata/snapshots/<name>.snap relative to the
// caller's test directory. The name should be a simple identifier without
// path separators (e.g. "label_hello", "button_disabled").
func Snapshot(t testing.TB, name string, widget runtime.Widget, width, height int) {
	t.Helper()

	got := SnapshotString(widget, width, height)

	dir := filepath.Join("testdata", "snapshots")
	path := filepath.Join(dir, name+".snap")

	if UpdateSnapshots {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("snapshot: failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("snapshot: failed to write %s: %v", path, err)
		}
		return
	}

	want, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// First run: create the golden file.
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatalf("snapshot: failed to create directory %s: %v", dir, err)
			}
			if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
				t.Fatalf("snapshot: failed to write %s: %v", path, err)
			}
			return
		}
		t.Fatalf("snapshot: failed to read %s: %v", path, err)
	}

	if got != string(want) {
		diff := snapshotDiff(string(want), got)
		t.Errorf("snapshot %q does not match golden file %s\n\n%s", name, path, diff)
	}
}

// SnapshotString captures the rendered output of a widget and returns it as
// a string without comparing against a golden file. Useful for debugging.
//
// The widget goes through a full measure-layout-render cycle at the given
// dimensions. The buffer is then serialized cell-by-cell, row-by-row.
func SnapshotString(widget runtime.Widget, width, height int) string {
	buf := runtime.NewBuffer(width, height)
	buf.Clear()

	constraints := runtime.Constraints{
		MinWidth:  0,
		MaxWidth:  width,
		MinHeight: 0,
		MaxHeight: height,
	}
	widget.Measure(constraints)
	widget.Layout(runtime.Rect{X: 0, Y: 0, Width: width, Height: height})
	widget.Render(runtime.RenderContext{Buffer: buf})

	return bufferToString(buf, width, height)
}

// bufferToString serializes a buffer to a string, one row per line.
// Trailing spaces on each line are preserved to maintain column alignment.
func bufferToString(buf *runtime.Buffer, width, height int) string {
	var sb strings.Builder
	sb.Grow(width*height + height) // cells + newlines
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			cell := buf.Get(x, y)
			r := cell.Rune
			if r == 0 {
				r = ' '
			}
			sb.WriteRune(r)
		}
		if y < height-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

// snapshotDiff produces a human-readable line-by-line diff between the
// expected (want) and actual (got) snapshot strings.
func snapshotDiff(want, got string) string {
	wantLines := strings.Split(want, "\n")
	gotLines := strings.Split(got, "\n")

	maxLines := len(wantLines)
	if len(gotLines) > maxLines {
		maxLines = len(gotLines)
	}

	var sb strings.Builder
	sb.WriteString("--- want\n+++ got\n")

	for i := 0; i < maxLines; i++ {
		var wl, gl string
		if i < len(wantLines) {
			wl = wantLines[i]
		}
		if i < len(gotLines) {
			gl = gotLines[i]
		}
		if wl == gl {
			sb.WriteString(fmt.Sprintf("  %3d: %s\n", i+1, wl))
		} else {
			if i < len(wantLines) {
				sb.WriteString(fmt.Sprintf("- %3d: %s\n", i+1, wl))
			}
			if i < len(gotLines) {
				sb.WriteString(fmt.Sprintf("+ %3d: %s\n", i+1, gl))
			}
		}
	}
	return sb.String()
}
