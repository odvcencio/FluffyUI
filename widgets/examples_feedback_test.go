package widgets_test

import (
	"github.com/odvcencio/fluffy-ui/state"
	"github.com/odvcencio/fluffy-ui/toast"
	"github.com/odvcencio/fluffy-ui/widgets"
)

func ExampleDialog() {
	dialog := widgets.NewDialog("Confirm", "Delete the item?", widgets.DialogButton{Label: "OK"})
	_ = dialog
}

func ExampleSpinner() {
	spinner := widgets.NewSpinner()
	spinner.Advance()
	_ = spinner
}

func ExampleProgress() {
	progress := widgets.NewProgress()
	progress.Value = 65
	_ = progress
}

func ExampleAlert() {
	alert := widgets.NewAlert("All systems nominal", widgets.AlertSuccess)
	_ = alert
}

func ExampleToastStack() {
	manager := toast.NewToastManager()
	stack := widgets.NewToastStack()
	manager.SetOnChange(stack.SetToasts)
	_ = stack
}

func ExampleSparkline() {
	data := state.NewSignal([]float64{1, 2, 3, 2, 4})
	spark := widgets.NewSparkline(data)
	_ = spark
}

func ExampleBarChart() {
	data := state.NewSignal([]widgets.BarData{
		{Label: "alpha", Value: 42},
		{Label: "beta", Value: 27},
	})
	chart := widgets.NewBarChart(data)
	_ = chart
}
