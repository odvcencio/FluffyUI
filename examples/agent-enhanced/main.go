// Example: Real-Time Agent Server
//
// This example demonstrates the real-time agent server which is now
// the default and only mode for agent interaction.
//
// The real-time server provides:
//   - Live UI change notifications
//   - Bidirectional WebSocket communication
//   - Async wait operations for UI conditions
//   - Event subscription system
//
// Usage:
//
//	# Basic TCP server
//	FLUFFYUI_AGENT=tcp::8716 go run .
//
//	# Unix socket with WebSocket
//	FLUFFYUI_AGENT=unix:/tmp/fluffy-agent.sock FLUFFYUI_AGENT_WS=:8765 go run .
//
// Connect with an agent client:
//
//	# JSONL protocol
//	echo '{"type": "hello"}' | nc -U /tmp/fluffy-agent.sock
//
//	# WebSocket protocol
//	ws ws://localhost:8765/agent
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/odvcencio/fluffyui/agent"
	"github.com/odvcencio/fluffyui/backend/tcell"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	"github.com/odvcencio/fluffyui/widgets"
)

func main() {
	// Create a simple counter app
	count := state.NewSignal(0)
	label := widgets.NewLabel("Count: 0")
	
	// Update label when count changes
	count.Subscribe(func() {
		label.SetText(fmt.Sprintf("Count: %d", count.Get()))
	})
	
	button := widgets.NewButton("Increment", widgets.WithOnClick(func() {
		count.Update(func(v int) int { return v + 1 })
	}))

	// Layout using VBox
	root := widgets.VBox(
		widgets.FlexFixed(widgets.NewLabel("Real-Time Agent Server Demo")),
		widgets.FlexFixed(widgets.NewLabel("The agent now operates in real-time mode by default")),
		widgets.FlexSpace(),
		widgets.FlexFixed(label),
		widgets.FlexFixed(button),
	)

	// Create backend
	be, err := tcell.New()
	if err != nil {
		log.Fatalf("Failed to create backend: %v", err)
	}

	// Create app
	app := runtime.NewApp(runtime.AppConfig{
		Backend: be,
		Root:    root,
	})

	// Enable real-time agent server from environment
	// This is the only mode now - no configuration needed
	server, err := agent.EnableFromEnv(app)
	if err != nil {
		log.Fatalf("Failed to start agent server: %v", err)
	}

	if server != nil {
		fmt.Println("Real-time agent server enabled")
		fmt.Printf("Server stats: %+v\n", server.Stats())

		// Demonstrate background task submission
		demoBackgroundTask(server)

		defer server.Stop()
	} else {
		fmt.Println("No agent server configured (set FLUFFYUI_AGENT)")
	}

	// Handle signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Run app
	if err := app.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("App error: %v", err)
	}
}

func demoBackgroundTask(server *agent.RealTimeServer) {
	// Submit a background task that runs independently
	task, err := server.SubmitBackgroundTask(
		"demo-task",
		"A demonstration background task",
		"demo-session",
		func(ctx context.Context, task *agent.BackgroundTask) error {
			fmt.Println("Background task started")
			defer fmt.Println("Background task completed")

			for i := 0; i < 10; i++ {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(1 * time.Second):
					task.SetProgress((i + 1) * 10)
					fmt.Printf("Background task progress: %d%%\n", (i+1)*10)
				}
			}
			return nil
		},
	)

	if err != nil {
		log.Printf("Failed to submit background task: %v", err)
		return
	}

	fmt.Printf("Submitted background task: %s\n", task.ID)

	// Monitor task in background
	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			if task.IsDone() {
				fmt.Printf("Task %s completed with status: %s\n", task.ID, task.Status())
				if err := task.Error(); err != nil {
					fmt.Printf("Task error: %v\n", err)
				}
				return
			}
			fmt.Printf("Task %s progress: %d%%\n", task.ID, task.Progress())
		}
	}()
}
