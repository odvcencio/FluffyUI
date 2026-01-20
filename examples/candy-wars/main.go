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
	"fmt"
	"os"
	"time"

	"github.com/odvcencio/fluffy-ui/examples/internal/demo"
	"github.com/odvcencio/fluffy-ui/runtime"
)

func main() {
	// rand is auto-seeded in Go 1.20+

	game := NewGame()
	view := NewGameView(game)

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

	// Game tick - price fluctuations and random events
	bundle.App.Every(time.Duration(tickSeconds)*time.Second, func(now time.Time) runtime.Message {
		if game.GameOver.Get() {
			return nil
		}
		game.Tick()
		return nil
	})

	if err := bundle.App.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
		os.Exit(1)
	}
}
