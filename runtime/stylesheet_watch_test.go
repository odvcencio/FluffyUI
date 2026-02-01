package runtime

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/backend/sim"
)

func TestWatchStylesheetFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.fss")
	if err := os.WriteFile(path, []byte("* { --accent: #111111; }"), 0o644); err != nil {
		t.Fatalf("write initial stylesheet: %v", err)
	}

	app := NewApp(AppConfig{
		Backend: sim.New(10, 4),
	})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx)
	}()
	t.Cleanup(func() {
		cancel()
		app.ExecuteCommand(Quit{})
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatalf("timeout waiting for app shutdown")
		}
	})

	stop := WatchStylesheetFile(app, path, 5*time.Millisecond)
	t.Cleanup(stop)

	time.Sleep(20 * time.Millisecond)
	if err := os.WriteFile(path, []byte("* { --accent: #222222; }"), 0o644); err != nil {
		t.Fatalf("write updated stylesheet: %v", err)
	}
	timestamp := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(path, timestamp, timestamp); err != nil {
		t.Fatalf("touch updated stylesheet: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if sheet := app.Stylesheet(); sheet != nil {
			if value, ok := sheet.GetVariable("accent"); ok && value == "#222222" {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("expected stylesheet update to apply")
}
