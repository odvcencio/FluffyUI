package main

import (
	"log"

	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/fluffy"
	"m31labs.dev/fluffyui/state"
	"m31labs.dev/fluffyui/widgets"
)

func main() {
	if err := fluffy.Run(buildShowcase()); err != nil {
		log.Fatal(err)
	}
}

func buildShowcase() *widgets.Tabs {
	overview := fluffy.VFlex(
		fluffy.Fixed(fluffy.Label("FluffyUI Showcase",
			widgets.WithLabelStyle(backend.DefaultStyle().Bold(true)))),
		fluffy.FixedSpace(1),
		fluffy.Fixed(widgets.NewAlert("All systems nominal", widgets.AlertSuccess)),
		fluffy.FixedSpace(1),
		fluffy.Fixed(progressWidget()),
		fluffy.FixedSpace(1),
		fluffy.Fixed(widgets.NewSpinner()),
	)

	inputs := fluffy.VFlex(
		fluffy.Fixed(fluffy.Label("Inputs",
			widgets.WithLabelStyle(backend.DefaultStyle().Bold(true)))),
		fluffy.FixedSpace(1),
		fluffy.Fixed(widgets.NewInput()),
		fluffy.FixedSpace(1),
		fluffy.Fixed(widgets.NewCheckbox("Enable alerts")),
		fluffy.FixedSpace(1),
		fluffy.Fixed(fluffy.SelectFromStrings(
			[]string{"Low", "Medium", "High"}, nil)),
		fluffy.FixedSpace(1),
		fluffy.Fixed(widgets.NewSlider(state.NewSignal(0.5))),
	)

	data := fluffy.VFlex(
		fluffy.Fixed(fluffy.Label("Data",
			widgets.WithLabelStyle(backend.DefaultStyle().Bold(true)))),
		fluffy.FixedSpace(1),
		fluffy.Fixed(dataTable()),
		fluffy.FixedSpace(1),
		fluffy.Fixed(sparklineWidget()),
	)

	return widgets.NewTabs(
		widgets.Tab{Title: "Overview", Content: overview},
		widgets.Tab{Title: "Inputs", Content: inputs},
		widgets.Tab{Title: "Data", Content: data},
	)
}

func progressWidget() *widgets.Progress {
	progress := widgets.NewProgress()
	progress.Label = "Capacity"
	progress.Value = 70
	return progress
}

func dataTable() *widgets.Panel {
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

func sparklineWidget() *widgets.Panel {
	data := state.NewSignal([]float64{12, 18, 14, 22, 16, 24, 19})
	spark := widgets.NewSparkline(data)
	panel := widgets.NewPanel(spark, widgets.WithPanelBorder(backend.DefaultStyle()))
	panel.SetTitle("Requests")
	return panel
}

