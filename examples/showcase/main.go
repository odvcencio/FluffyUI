package main

import (
	"context"
	"fmt"
	"os"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/fluffy"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	"github.com/odvcencio/fluffyui/widgets"
)

func main() {
	app, err := fluffy.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "app init failed: %v\n", err)
		os.Exit(1)
	}
	app.SetRoot(buildShowcase())

	if err := app.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
		os.Exit(1)
	}
}

func buildShowcase() runtime.Widget {
	overview := widgets.VBox(
		widgets.FlexFixed(widgets.NewLabel("FluffyUI Showcase", widgets.WithLabelStyle(backend.DefaultStyle().Bold(true)))),
		widgets.FlexFixed(widgets.NewAlert("All systems nominal", widgets.AlertSuccess)),
		widgets.FlexFixed(progressWidget()),
		widgets.FlexFixed(widgets.NewSpinner()),
	)
	overview.Gap = 1

	inputs := widgets.VBox(
		widgets.FlexFixed(widgets.NewLabel("Inputs", widgets.WithLabelStyle(backend.DefaultStyle().Bold(true)))),
		widgets.FlexFixed(widgets.NewInput()),
		widgets.FlexFixed(widgets.NewCheckbox("Enable alerts")),
		widgets.FlexFixed(widgets.NewSelect(
			widgets.SelectOption{Label: "Low", Value: "low"},
			widgets.SelectOption{Label: "Medium", Value: "med"},
			widgets.SelectOption{Label: "High", Value: "high"},
		)),
		widgets.FlexFixed(widgets.NewSlider(state.NewSignal(0.5))),
	)
	inputs.Gap = 1

	data := widgets.VBox(
		widgets.FlexFixed(widgets.NewLabel("Data", widgets.WithLabelStyle(backend.DefaultStyle().Bold(true)))),
		widgets.FlexFixed(dataTable()),
		widgets.FlexFixed(sparklineWidget()),
	)
	data.Gap = 1

	tabs := widgets.NewTabs(
		widgets.Tab{Title: "Overview", Content: overview},
		widgets.Tab{Title: "Inputs", Content: inputs},
		widgets.Tab{Title: "Data", Content: data},
	)
	return tabs
}

func progressWidget() runtime.Widget {
	progress := widgets.NewProgress()
	progress.Label = "Capacity"
	progress.Value = 70
	return progress
}

func dataTable() runtime.Widget {
	table := widgets.NewTable(
		widgets.TableColumn{Title: "Service"},
		widgets.TableColumn{Title: "Latency"},
	)
	table.SetRows([][]string{
		{"Auth", "32ms"},
		{"Billing", "45ms"},
		{"Search", "57ms"},
	})
	panel := widgets.NewPanel(table, widgets.WithPanelBorder(backend.DefaultStyle()))
	panel.SetTitle("Services")
	return panel
}

func sparklineWidget() runtime.Widget {
	data := state.NewSignal([]float64{12, 18, 14, 22, 16, 24, 19})
	spark := widgets.NewSparkline(data)
	panel := widgets.NewPanel(spark, widgets.WithPanelBorder(backend.DefaultStyle()))
	panel.SetTitle("Requests")
	return panel
}

