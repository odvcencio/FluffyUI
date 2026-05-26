//go:build ignore

// web demonstrates FluffyUI's web backend feature.
//
// This example shows how to run a TUI application that can be accessed
// via web browser, similar to Textual's textual-serve feature.
//
// Usage:
//
//	go run ./examples/web
//	# Then open http://localhost:8080 in your browser
//
// Or specify a custom port:
//
//	FLUFFYUI_BACKEND=web FLUFFYUI_WEB_ADDR=:3000 go run ./examples/web
//
// Features demonstrated:
//   - Web backend with WebSocket transport
//   - Real-time bidirectional communication
//   - ANSI terminal emulation via xterm.js
//   - Keyboard and mouse input forwarding
//   - Multiple concurrent sessions (optional)
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"m31labs.dev/fluffyui/backend/web"
	"m31labs.dev/fluffyui/fluffy"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/state"
	"m31labs.dev/fluffyui/widgets"
)

func main() {
	// Check if web backend should be used
	useWeb := os.Getenv("FLUFFYUI_BACKEND") == "web"

	var app *runtime.App
	var webBackend *web.Backend
	var err error

	if useWeb {
		// Method 1: Environment variable (FLUFFYUI_BACKEND=web)
		app, err = fluffy.NewApp()
	} else {
		// Method 2: Explicit web backend configuration
		webBackend = web.New(web.Config{
			Addr:   ":8080",
			Title:  "FluffyUI Web Demo",
			Path:   "/",
			WSPath: "/ws",
		})
		app, err = fluffy.NewApp(fluffy.WithBackend(webBackend))
	}

	if err != nil {
		log.Fatalf("Failed to create app: %v", err)
	}

	// Get the web backend to print the URL
	if webBackend == nil {
		fmt.Println("=" + strings.Repeat("=", 50))
		fmt.Println("FluffyUI Web Demo")
		fmt.Println("=" + strings.Repeat("=", 50))
		fmt.Println()
		fmt.Println("Open http://localhost:8080 in your browser")
		fmt.Println()
	} else {
		// Wait a moment for the server to start
		time.Sleep(50 * time.Millisecond)

		fmt.Println("=" + strings.Repeat("=", 50))
		fmt.Println("FluffyUI Web Demo")
		fmt.Println("=" + strings.Repeat("=", 50))
		fmt.Println()
		fmt.Println("Terminal URL:", webBackend.URL())
		fmt.Println()
	}

	// Create the demo UI
	root := NewWebDemo()
	app.SetRoot(root)

	// Run the app
	fmt.Println("Press Ctrl+C to stop the server")
	fmt.Println()

	if err := app.Run(context.Background()); err != nil && err != context.Canceled {
		log.Fatalf("App error: %v", err)
	}
}

// WebDemo is the main demo widget.
type WebDemo struct {
	widgets.Component
	counter  *state.Signal[int]
	messages *state.Signal[[]string]
	grid     *widgets.Grid
}

// NewWebDemo creates the demo widget.
func NewWebDemo() *WebDemo {
	d := &WebDemo{
		counter:  state.NewSignal(0),
		messages: state.NewSignal([]string{"Welcome to FluffyUI Web Demo!"}),
	}

	// Create grid layout
	d.grid = widgets.NewGrid(4, 2)
	d.grid.Gap = 1

	// Title
	title := widgets.NewLabel("FluffyUI Web Demo", widgets.WithLabelStyle(fluffy.DefaultStyle().Bold(true)))
	d.grid.Add(title, 0, 0, 1, 2)

	// Counter section
	counterLabel := widgets.NewLabel("Counter: 0")

	incBtn := widgets.NewButton("+", widgets.WithVariant(widgets.VariantPrimary), widgets.WithOnClick(func() {
		d.counter.Update(func(v int) int { return v + 1 })
	}))
	decBtn := widgets.NewButton("-", widgets.WithVariant(widgets.VariantSecondary), widgets.WithOnClick(func() {
		d.counter.Update(func(v int) int { return v - 1 })
	}))

	buttonStack := widgets.NewStack(incBtn, decBtn)
	d.grid.Add(counterLabel, 1, 0, 1, 1)
	d.grid.Add(buttonStack, 1, 1, 1, 1)

	// Message log
	msgLabel := widgets.NewLabel("Messages:")
	d.grid.Add(msgLabel, 2, 0, 1, 2)

	// Add message button
	addBtn := widgets.NewButton("Add Message", widgets.WithVariant(widgets.VariantPrimary), widgets.WithOnClick(func() {
		d.addMessage()
	}))
	d.grid.Add(addBtn, 3, 0, 1, 2)

	return d
}

func (d *WebDemo) Bind(services runtime.Services) {
	d.Component.Bind(services)

	// Observe counter changes
	d.Observe(d.counter, func() {
		d.Invalidate()
	})
}

func (d *WebDemo) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (d *WebDemo) Layout(bounds runtime.Rect) {
	d.Component.Layout(bounds)
	d.grid.Layout(bounds)
}

func (d *WebDemo) Render(ctx runtime.RenderContext) {
	d.grid.Render(ctx)
}

func (d *WebDemo) HandleMessage(msg runtime.Message) runtime.HandleResult {
	// Forward to grid
	if result := d.grid.HandleMessage(msg); result.Handled {
		return result
	}
	return runtime.Unhandled()
}

func (d *WebDemo) ChildWidgets() []runtime.Widget {
	return []runtime.Widget{d.grid}
}

func (d *WebDemo) addMessage() {
	msg := fmt.Sprintf("Message at %s", time.Now().Format("15:04:05"))

	// Add to messages
	d.messages.Update(func(msgs []string) []string {
		return append(msgs, msg)
	})

	d.Invalidate()
}
