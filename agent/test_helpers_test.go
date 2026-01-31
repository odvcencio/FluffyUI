package agent

import (
	"context"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/runtime"
)

func runAppForTest(t *testing.T, app *runtime.App) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx)
	}()

	t.Cleanup(func() {
		cancel()
		app.ExecuteCommand(runtime.Quit{})
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatalf("timeout waiting for app shutdown")
		}
	})
}
