// Candy Wars - A FluffyUI Showcase Game
//
// Navigate the halls of Jefferson Middle School trading candy between
// locations. Buy low, sell high, and avoid getting caught by teachers!
//
// This game demonstrates FluffyUI's capabilities:
// - Reactive state management with Signals
// - Complex widget composition (Tables, Dialogs, Charts, Panels)
// - Keybinding system
// - Dynamic UI updates
// - Game loop with tick-based events
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/odvcencio/fluffy-ui/agent"
	"github.com/odvcencio/fluffy-ui/examples/internal/demo"
	"github.com/odvcencio/fluffy-ui/runtime"
)

func main() {
	// rand is auto-seeded in Go 1.20+

	view := NewAppView()

	bundle, err := demo.NewApp(view, demo.Options{
		CommandHandler: func(cmd runtime.Command) bool {
			if _, ok := cmd.(runtime.Quit); ok {
				return true
			}
			return false
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "app init failed: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := startAgentServer(ctx, bundle.App)
	if server != nil {
		defer server.Close()
	}

	// Game tick - price fluctuations and random events
	bundle.App.Every(time.Duration(tickSeconds)*time.Second, func(now time.Time) runtime.Message {
		if view.showNewGame || view.game == nil || view.game.GameOver.Get() {
			return nil
		}
		view.game.Tick()
		return nil
	})

	if err := bundle.App.Run(ctx); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
		os.Exit(1)
	}
}

func startAgentServer(ctx context.Context, app *runtime.App) *agent.Server {
	addr := strings.TrimSpace(os.Getenv("FLUFFYUI_AGENT"))
	if addr == "" {
		return nil
	}
	opts := agent.ServerOptions{
		Addr:      addr,
		App:       app,
		AllowText: envBool("FLUFFYUI_AGENT_ALLOW_TEXT"),
		TestMode:  envBool("FLUFFYUI_AGENT_TEST_MODE"),
		Token:     os.Getenv("FLUFFYUI_AGENT_TOKEN"),
	}
	server, err := agent.NewServer(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "agent server init failed: %v\n", err)
		return nil
	}
	go func() {
		if err := server.Serve(ctx); err != nil && !errors.Is(err, context.Canceled) {
			fmt.Fprintf(os.Stderr, "agent server error: %v\n", err)
		}
	}()
	return server
}

func envBool(key string) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	switch value {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
