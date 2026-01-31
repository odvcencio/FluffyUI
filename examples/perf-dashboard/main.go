package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/odvcencio/fluffyui/backend"
	backendtcell "github.com/odvcencio/fluffyui/backend/tcell"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/widgets"
)

func main() {
	sampler := runtime.NewRenderSampler(120)

	dashboard := widgets.NewPerformanceDashboard(
		sampler,
		widgets.WithPerformanceRefresh(500*time.Millisecond),
	)
	perfPanel := widgets.NewPanel(dashboard, widgets.WithPanelBorder(backend.DefaultStyle()))
	perfPanel.SetTitle("Performance")

	content := widgets.NewRichText("# Performance Dashboard\n\n" +
		"This demo renders a live performance summary using the render sampler.\n" +
		"The dashboard refreshes every 500ms to surface FPS, render, and flush timing.\n\n" +
		"Try resizing the terminal or interacting with widgets to see updates.")
	content.SetShowScrollbar(true)
	contentPanel := widgets.NewPanel(content, widgets.WithPanelBorder(backend.DefaultStyle()))
	contentPanel.SetTitle("Demo")

	root := widgets.HBox(
		widgets.FlexFlexible(contentPanel, 1),
		widgets.FlexFixed(perfPanel),
	)

	be, err := backendtcell.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "backend init failed: %v\n", err)
		os.Exit(1)
	}

	app := runtime.NewApp(runtime.AppConfig{
		Backend:           be,
		Root:              root,
		TickRate:          time.Second / 30,
		RenderObserver:    sampler,
		FocusRegistration: runtime.FocusRegistrationAuto,
	})

	if err := app.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
		os.Exit(1)
	}
}

