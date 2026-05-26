//go:build !js

package widgettest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/widgets"
)

// testWidget renders a fixed string at (0,0) for deterministic snapshot output.
type testWidget struct {
	text   string
	bounds runtime.Rect
}

func (w *testWidget) Measure(c runtime.Constraints) runtime.Size {
	width := len(w.text)
	if width > c.MaxWidth {
		width = c.MaxWidth
	}
	h := 1
	if h > c.MaxHeight {
		h = c.MaxHeight
	}
	return runtime.Size{Width: width, Height: h}
}

func (w *testWidget) Layout(bounds runtime.Rect) {
	w.bounds = bounds
}

func (w *testWidget) Render(ctx runtime.RenderContext) {
	for i, r := range w.text {
		x := w.bounds.X + i
		if x >= w.bounds.X+w.bounds.Width {
			break
		}
		ctx.Buffer.Set(x, w.bounds.Y, r, backend.DefaultStyle())
	}
}

func (w *testWidget) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}

func TestSnapshotString_Label(t *testing.T) {
	label := widgets.NewLabel("Hello")
	got := SnapshotString(label, 10, 1)

	if !strings.Contains(got, "Hello") {
		t.Errorf("expected SnapshotString to contain 'Hello', got %q", got)
	}
	// Should be exactly 10 characters wide (label text + padding spaces).
	if len(got) != 10 {
		t.Errorf("expected SnapshotString width 10, got %d: %q", len(got), got)
	}
}

func TestSnapshotString_MultiRow(t *testing.T) {
	w := &testWidget{text: "AB"}
	got := SnapshotString(w, 5, 3)

	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	// First line should start with "AB" then spaces.
	if !strings.HasPrefix(lines[0], "AB") {
		t.Errorf("expected first line to start with 'AB', got %q", lines[0])
	}
	// Each line should be 5 characters wide.
	for i, line := range lines {
		if len(line) != 5 {
			t.Errorf("line %d: expected width 5, got %d: %q", i, len(line), line)
		}
	}
}

func TestSnapshot_CreatesGoldenFile(t *testing.T) {
	// Use a temp dir as the working directory so we don't pollute the repo.
	dir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	w := &testWidget{text: "XY"}
	Snapshot(t, "create_test", w, 4, 1)

	// Verify the file was created.
	path := filepath.Join(dir, "testdata", "snapshots", "create_test.snap")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("golden file was not created: %v", err)
	}
	if !strings.HasPrefix(string(data), "XY") {
		t.Errorf("golden file content unexpected: %q", string(data))
	}
}

func TestSnapshot_MatchesExisting(t *testing.T) {
	// Pre-create a golden file, then verify Snapshot succeeds when output matches.
	dir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	w := &testWidget{text: "OK"}
	expected := SnapshotString(w, 6, 1)

	snapDir := filepath.Join(dir, "testdata", "snapshots")
	if err := os.MkdirAll(snapDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(snapDir, "match_test.snap"), []byte(expected), 0o644); err != nil {
		t.Fatal(err)
	}

	// Should not fail.
	Snapshot(t, "match_test", w, 6, 1)
}

func TestSnapshot_FailsOnMismatch(t *testing.T) {
	dir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	// Write a golden file with different content.
	snapDir := filepath.Join(dir, "testdata", "snapshots")
	if err := os.MkdirAll(snapDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(snapDir, "mismatch_test.snap"), []byte("WRONG"), 0o644); err != nil {
		t.Fatal(err)
	}

	ft := &fakeTB{}
	w := &testWidget{text: "RIGHT"}
	Snapshot(ft, "mismatch_test", w, 5, 1)

	if !ft.failed {
		t.Error("expected Snapshot to fail when output does not match golden file")
	}
	// Error message should contain diff information.
	found := false
	for _, msg := range ft.messages {
		if strings.Contains(msg, "does not match") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error message with diff, got: %v", ft.messages)
	}
}

func TestSnapshot_UpdateOverwrites(t *testing.T) {
	dir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	// Write an old golden file.
	snapDir := filepath.Join(dir, "testdata", "snapshots")
	if err := os.MkdirAll(snapDir, 0o755); err != nil {
		t.Fatal(err)
	}
	snapPath := filepath.Join(snapDir, "update_test.snap")
	if err := os.WriteFile(snapPath, []byte("OLD"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Enable update mode.
	origUpdate := UpdateSnapshots
	UpdateSnapshots = true
	t.Cleanup(func() { UpdateSnapshots = origUpdate })

	w := &testWidget{text: "NEW"}
	Snapshot(t, "update_test", w, 5, 1)

	// Verify file was overwritten.
	data, err := os.ReadFile(snapPath)
	if err != nil {
		t.Fatalf("golden file not found after update: %v", err)
	}
	if !strings.HasPrefix(string(data), "NEW") {
		t.Errorf("golden file was not updated: got %q", string(data))
	}
}

func TestSnapshotDiff(t *testing.T) {
	want := "line1\nline2\nline3"
	got := "line1\nchanged\nline3"

	diff := snapshotDiff(want, got)

	if !strings.Contains(diff, "--- want") {
		t.Error("diff missing header")
	}
	if !strings.Contains(diff, "+++ got") {
		t.Error("diff missing header")
	}
	if !strings.Contains(diff, "- ") {
		t.Error("diff missing removed line marker")
	}
	if !strings.Contains(diff, "+ ") {
		t.Error("diff missing added line marker")
	}
	if !strings.Contains(diff, "line2") {
		t.Error("diff should contain old line content")
	}
	if !strings.Contains(diff, "changed") {
		t.Error("diff should contain new line content")
	}
}
