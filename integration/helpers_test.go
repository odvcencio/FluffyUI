package integration

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/runtime"
)

const (
	appTimeout    = 2 * time.Second
	socketTimeout = 2 * time.Second
)

func runApp(t *testing.T, app *runtime.App) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx)
	}()

	waitForScreen(t, app)

	t.Cleanup(func() {
		cancel()
		app.ExecuteCommand(runtime.Quit{})
		select {
		case err := <-done:
			if err != nil && !errors.Is(err, context.Canceled) {
				t.Fatalf("app run: %v", err)
			}
		case <-time.After(appTimeout):
			t.Fatalf("timeout waiting for app shutdown")
		}
	})
}

func waitForScreen(t *testing.T, app *runtime.App) {
	t.Helper()

	deadline := time.Now().Add(appTimeout)
	for time.Now().Before(deadline) {
		if app.Screen() != nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("timeout waiting for app screen")
}

func waitForSocket(t *testing.T, path string) {
	t.Helper()

	deadline := time.Now().Add(socketTimeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("timeout waiting for socket: %s", path)
}
