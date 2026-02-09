package style

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewFileWatcher(t *testing.T) {
	sheet := NewStylesheet()
	w := NewFileWatcher("/tmp/nonexistent.fss", sheet)
	if w == nil {
		t.Fatal("expected non-nil FileWatcher")
	}
	if w.path != "/tmp/nonexistent.fss" {
		t.Fatalf("expected path /tmp/nonexistent.fss, got %s", w.path)
	}
	if w.stylesheet != sheet {
		t.Fatal("expected stylesheet to be the one provided")
	}
	if w.interval != 500*time.Millisecond {
		t.Fatalf("expected default interval 500ms, got %v", w.interval)
	}
}

func TestFileWatcherReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.fss")
	if err := os.WriteFile(path, []byte(`Button { padding: 2; }`), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	sheet := NewStylesheet()
	w := NewFileWatcher(path, sheet)

	if err := w.Reload(); err != nil {
		t.Fatalf("Reload failed: %v", err)
	}

	// The stylesheet should now have rules from the file.
	sheet.mu.RLock()
	ruleCount := len(sheet.rules)
	sheet.mu.RUnlock()
	if ruleCount == 0 {
		t.Fatal("expected rules after Reload, got 0")
	}
}

func TestFileWatcherReloadInvalidFile(t *testing.T) {
	dir := t.TempDir()

	sheet := NewStylesheet()
	// Pre-populate with a rule so we can verify it's preserved.
	sheet.Add(Select("Label"), Style{Padding: Pad(1)})

	w := NewFileWatcher(filepath.Join(dir, "does_not_exist.fss"), sheet)

	var gotError error
	w.SetOnError(func(err error) {
		gotError = err
	})

	err := w.Reload()
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
	if gotError == nil {
		t.Fatal("expected onError callback to be called")
	}

	// Old stylesheet rules should be preserved.
	sheet.mu.RLock()
	ruleCount := len(sheet.rules)
	sheet.mu.RUnlock()
	if ruleCount != 1 {
		t.Fatalf("expected 1 rule preserved after failed reload, got %d", ruleCount)
	}
}

func TestFileWatcherModificationDetection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "watch.fss")
	if err := os.WriteFile(path, []byte(`* { --color: #aaa; }`), 0o644); err != nil {
		t.Fatalf("write initial file: %v", err)
	}

	sheet := NewStylesheet()
	w := NewFileWatcher(path, sheet)
	w.SetInterval(5 * time.Millisecond)

	var changeCount atomic.Int32
	w.SetOnChange(func(s *Stylesheet) {
		changeCount.Add(1)
	})

	if err := w.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	t.Cleanup(w.Stop)

	// Wait a bit for the watcher to be running.
	time.Sleep(20 * time.Millisecond)

	// Modify the file with a future mod time to guarantee detection.
	if err := os.WriteFile(path, []byte(`* { --color: #bbb; }`), 0o644); err != nil {
		t.Fatalf("write updated file: %v", err)
	}
	future := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(path, future, future); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	// Wait for the change to be detected.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if changeCount.Load() > 0 {
			// Verify the variable was updated.
			if val, ok := sheet.GetVariable("color"); !ok || val != "#bbb" {
				t.Fatalf("expected variable color=#bbb, got %q (ok=%v)", val, ok)
			}
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timeout waiting for file change detection")
}

func TestFileWatcherStop(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "stop.fss")
	if err := os.WriteFile(path, []byte(`Label { margin: 1; }`), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	sheet := NewStylesheet()
	w := NewFileWatcher(path, sheet)
	w.SetInterval(5 * time.Millisecond)

	if err := w.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Stop should not panic.
	w.Stop()

	// Double stop should be safe.
	w.Stop()

	// After stop, further file changes should not trigger reloads.
	var changeCount atomic.Int32
	w.SetOnChange(func(s *Stylesheet) {
		changeCount.Add(1)
	})

	if err := os.WriteFile(path, []byte(`Label { margin: 5; }`), 0o644); err != nil {
		t.Fatalf("write file after stop: %v", err)
	}
	future := time.Now().Add(2 * time.Second)
	_ = os.Chtimes(path, future, future)

	time.Sleep(50 * time.Millisecond)
	if changeCount.Load() != 0 {
		t.Fatal("expected no changes after Stop")
	}
}

func TestFileWatcherOnChangeCallback(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cb.fss")
	if err := os.WriteFile(path, []byte(`Panel { padding: 2; }`), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	sheet := NewStylesheet()
	w := NewFileWatcher(path, sheet)

	var callbackSheet *Stylesheet
	w.SetOnChange(func(s *Stylesheet) {
		callbackSheet = s
	})

	if err := w.Reload(); err != nil {
		t.Fatalf("Reload failed: %v", err)
	}

	if callbackSheet == nil {
		t.Fatal("expected onChange callback to fire on Reload")
	}
	if callbackSheet != sheet {
		t.Fatal("expected onChange to receive the watcher's stylesheet")
	}
}

func TestFileWatcherSetInterval(t *testing.T) {
	sheet := NewStylesheet()
	w := NewFileWatcher("/tmp/test.fss", sheet)

	w.SetInterval(100 * time.Millisecond)
	if w.interval != 100*time.Millisecond {
		t.Fatalf("expected interval 100ms, got %v", w.interval)
	}

	// Negative interval should be ignored.
	w.SetInterval(-1 * time.Millisecond)
	if w.interval != 100*time.Millisecond {
		t.Fatalf("expected interval unchanged at 100ms, got %v", w.interval)
	}

	// Zero should be ignored.
	w.SetInterval(0)
	if w.interval != 100*time.Millisecond {
		t.Fatalf("expected interval unchanged at 100ms, got %v", w.interval)
	}
}

func TestFileWatcherStartTwice(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "twice.fss")
	if err := os.WriteFile(path, []byte(`* { margin: 0; }`), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	sheet := NewStylesheet()
	w := NewFileWatcher(path, sheet)
	w.SetInterval(50 * time.Millisecond)

	if err := w.Start(); err != nil {
		t.Fatalf("first Start failed: %v", err)
	}
	t.Cleanup(w.Stop)

	// Second Start should return an error.
	if err := w.Start(); err == nil {
		t.Fatal("expected error on second Start")
	}
}

func TestFileWatcherNilReceiver(t *testing.T) {
	var w *FileWatcher

	// All nil-receiver methods should not panic.
	w.SetInterval(time.Second)
	w.SetOnChange(nil)
	w.SetOnError(nil)
	w.Stop()

	if err := w.Start(); err == nil {
		t.Fatal("expected error on nil Start")
	}
	if err := w.Reload(); err == nil {
		t.Fatal("expected error on nil Reload")
	}
}

