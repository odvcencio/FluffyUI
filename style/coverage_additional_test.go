package style

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatchFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "theme.fss")
	ch := make(chan error, 4)

	stop := WatchFile(path, 5*time.Millisecond, func(sheet *Stylesheet, err error) {
		ch <- err
	})
	defer stop()

	select {
	case err := <-ch:
		if err == nil {
			t.Fatalf("expected error for missing file")
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("timeout waiting for missing file error")
	}

	if err := os.WriteFile(path, []byte("Button { padding: 1; }"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	deadline := time.After(500 * time.Millisecond)
	for {
		select {
		case err := <-ch:
			if err == nil {
				return
			}
		case <-deadline:
			t.Fatalf("timeout waiting for parse success")
		}
	}
}
