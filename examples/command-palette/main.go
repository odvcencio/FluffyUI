package main

import (
	"context"
	"fmt"
	"os"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/examples/internal/demo"
	"github.com/odvcencio/fluffyui/keybind"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/widgets"
)

func main() {
	view := NewPaletteView()
	var palette *widgets.EnhancedPalette
	var registry *keybind.CommandRegistry
	var app *runtime.App

	handler := func(cmd runtime.Command) bool {
		sel, ok := cmd.(runtime.PaletteSelected)
		if !ok || registry == nil || app == nil {
			return false
		}
		ctx := keybind.Context{App: app}
		if screen := app.Screen(); screen != nil {
			if scope := screen.FocusScope(); scope != nil {
				focused := scope.Current()
				ctx.Focused = focused
				if accessible, ok := focused.(accessibility.Accessible); ok {
					ctx.FocusedWidget = accessible
				}
			}
		}
		registry.Execute(sel.ID, ctx)
		if palette != nil {
			palette.Record(sel.ID)
		}
		return true
	}

	bundle, err := demo.NewApp(view, demo.Options{CommandHandler: handler})
	if err != nil {
		fmt.Fprintf(os.Stderr, "app init failed: %v\n", err)
		os.Exit(1)
	}
	registry = bundle.Registry
	app = bundle.App

	registerDemoCommands(registry, view)
	palette = widgets.NewEnhancedPalette(registry)
	palette.SetKeymapStack(bundle.Keymaps)

	registry.Register(keybind.Command{
		ID:          "palette.open",
		Title:       "Open Command Palette",
		Description: "Show the command palette",
		Category:    "Palette",
		Handler: func(ctx keybind.Context) {
			if ctx.App != nil && palette != nil {
				ctx.App.ExecuteCommand(runtime.PushOverlay{Widget: palette.Widget, Modal: true})
			}
		},
	})

	bundle.Keymaps.Push(&keybind.Keymap{
		Name: "palette",
		Bindings: []keybind.Binding{
			{Key: keybind.MustParseKeySequence("ctrl+p"), Command: "palette.open"},
		},
	})

	if err := bundle.App.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
		os.Exit(1)
	}
}

type PaletteView struct {
	widgets.Component
	lastAction string
}

func NewPaletteView() *PaletteView {
	return &PaletteView{lastAction: "none"}
}

func (p *PaletteView) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (p *PaletteView) Layout(bounds runtime.Rect) {
	p.Component.Layout(bounds)
}

func (p *PaletteView) Render(ctx runtime.RenderContext) {
	bounds := p.Bounds()
	lines := []string{
		"Command Palette Demo",
		"",
		"Press Ctrl+P to open the palette.",
		"Use arrows to select, Enter to run.",
		"",
		"Last action: " + p.lastAction,
	}
	for i, line := range lines {
		if i >= bounds.Height {
			break
		}
		if len(line) > bounds.Width {
			line = line[:bounds.Width]
		}
		ctx.Buffer.SetString(bounds.X, bounds.Y+i, line, backend.DefaultStyle())
	}
}

func (p *PaletteView) SetLastAction(action string) {
	p.lastAction = action
	p.Invalidate()
}

func registerDemoCommands(registry *keybind.CommandRegistry, view *PaletteView) {
	if registry == nil {
		return
	}
	registry.RegisterAll(
		keybind.Command{
			ID:          "demo.new",
			Title:       "New Session",
			Description: "Create a fresh session",
			Category:    "Demo",
			Handler: func(ctx keybind.Context) {
				if view != nil {
					view.SetLastAction("new session")
				}
			},
		},
		keybind.Command{
			ID:          "demo.open",
			Title:       "Open Workspace",
			Description: "Open the default workspace",
			Category:    "Demo",
			Handler: func(ctx keybind.Context) {
				if view != nil {
					view.SetLastAction("open workspace")
				}
			},
		},
		keybind.Command{
			ID:          "demo.sync",
			Title:       "Sync Data",
			Description: "Synchronize remote data",
			Category:    "Demo",
			Handler: func(ctx keybind.Context) {
				if view != nil {
					view.SetLastAction("sync data")
				}
			},
		},
	)
}
