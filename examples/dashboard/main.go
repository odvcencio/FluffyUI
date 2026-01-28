package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/examples/internal/demo"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	"github.com/odvcencio/fluffyui/widgets"
)

func main() {
	view := NewDashboardView()
	bundle, err := demo.NewApp(view, demo.Options{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "app init failed: %v\n", err)
		os.Exit(1)
	}

	if err := bundle.App.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
		os.Exit(1)
	}
}

type DashboardView struct {
	widgets.Component
	header *widgets.Label

	requestsLabel *widgets.Label
	errorsLabel   *widgets.Label
	uptimeLabel   *widgets.Label

	progress *widgets.Progress
	alert    *widgets.Alert
	spark    *widgets.Sparkline
	latency  *widgets.BarChart
	table    *widgets.Table

	metricsGrid *widgets.Grid
	splitter    *widgets.Splitter
	leftPanel   *widgets.Panel
	rightPanel  *widgets.Panel

	sparkData   *state.Signal[[]float64]
	latencyData *state.Signal[[]widgets.BarData]

	requests int
	errors   int
	uptime   time.Duration
	lastTick time.Time
}

func NewDashboardView() *DashboardView {
	view := &DashboardView{}
	view.header = widgets.NewLabel("System Dashboard").WithStyle(backend.DefaultStyle().Bold(true))

	view.requestsLabel = widgets.NewLabel("Requests: 0")
	view.errorsLabel = widgets.NewLabel("Errors: 0")
	view.uptimeLabel = widgets.NewLabel("Uptime: 0s")

	metricsRow := demo.NewHBox(
		wrapMetricPanel("Requests", view.requestsLabel),
		wrapMetricPanel("Errors", view.errorsLabel),
		wrapMetricPanel("Uptime", view.uptimeLabel),
	)
	metricsRow.Gap = 2

	view.metricsGrid = widgets.NewGrid(1, 1)
	view.metricsGrid.Add(metricsRow, 0, 0, 1, 1)

	view.progress = widgets.NewProgress()
	view.progress.Label = "Capacity"

	view.alert = widgets.NewAlert("All systems nominal", widgets.AlertSuccess)

	view.sparkData = state.NewSignal([]float64{12, 18, 14, 22, 16, 24, 19})
	view.spark = widgets.NewSparkline(view.sparkData)
	view.latencyData = state.NewSignal([]widgets.BarData{
		{Label: "Auth", Value: 32},
		{Label: "Billing", Value: 45},
		{Label: "Search", Value: 57},
	})
	view.latency = widgets.NewBarChart(view.latencyData)
	view.latency.ShowLabels = true
	view.latency.ShowValues = true

	view.table = widgets.NewTable(
		widgets.TableColumn{Title: "Service"},
		widgets.TableColumn{Title: "Status"},
		widgets.TableColumn{Title: "Latency"},
	)
	view.table.SetRows([][]string{
		{"Auth", "OK", "32ms"},
		{"Billing", "OK", "45ms"},
		{"Search", "OK", "57ms"},
	})

	rightColumn := demo.NewVBox(view.alert, view.progress, view.spark, view.latency)
	rightColumn.Gap = 1

	view.leftPanel = widgets.NewPanel(view.table).WithBorder(backend.DefaultStyle())
	view.leftPanel.SetTitle("Services")
	view.rightPanel = widgets.NewPanel(rightColumn).WithBorder(backend.DefaultStyle())
	view.rightPanel.SetTitle("Signals")

	view.splitter = widgets.NewSplitter(view.leftPanel, view.rightPanel)
	view.splitter.Ratio = 0.6

	view.updateMetrics()
	return view
}

func wrapMetricPanel(title string, value *widgets.Label) *widgets.Panel {
	stack := demo.NewVBox(widgets.NewLabel(title), value)
	stack.Gap = 1
	panel := widgets.NewPanel(stack).WithBorder(backend.DefaultStyle())
	panel.SetTitle(title)
	return panel
}

func (d *DashboardView) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (d *DashboardView) Layout(bounds runtime.Rect) {
	d.Component.Layout(bounds)
	y := bounds.Y
	if d.header != nil {
		d.header.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: 1})
		y++
	}
	metricsHeight := 5
	if d.metricsGrid != nil {
		d.metricsGrid.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: metricsHeight})
		y += metricsHeight
	}
	mainHeight := bounds.Height - (y - bounds.Y)
	if mainHeight < 0 {
		mainHeight = 0
	}
	if d.splitter != nil {
		d.splitter.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: mainHeight})
	}
}

func (d *DashboardView) Render(ctx runtime.RenderContext) {
	if d.header != nil {
		d.header.Render(ctx)
	}
	if d.metricsGrid != nil {
		d.metricsGrid.Render(ctx)
	}
	if d.splitter != nil {
		d.splitter.Render(ctx)
	}
}

func (d *DashboardView) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if tick, ok := msg.(runtime.TickMsg); ok {
		d.onTick(tick.Time)
	}
	if d.splitter != nil {
		if result := d.splitter.HandleMessage(msg); result.Handled {
			return result
		}
	}
	return runtime.Unhandled()
}

func (d *DashboardView) ChildWidgets() []runtime.Widget {
	children := []runtime.Widget{}
	if d.header != nil {
		children = append(children, d.header)
	}
	if d.metricsGrid != nil {
		children = append(children, d.metricsGrid)
	}
	if d.splitter != nil {
		children = append(children, d.splitter)
	}
	return children
}

func (d *DashboardView) onTick(now time.Time) {
	if !d.lastTick.IsZero() && now.Sub(d.lastTick) < 700*time.Millisecond {
		return
	}
	d.lastTick = now
	d.requests += 12
	d.errors = (d.errors + 1) % 7
	d.uptime += 700 * time.Millisecond
	d.progress.Value = float64(d.requests % 100)
	if d.errors >= 4 {
		d.alert.Variant = widgets.AlertWarning
		d.alert.Text = "Elevated error rate"
	} else {
		d.alert.Variant = widgets.AlertSuccess
		d.alert.Text = "All systems nominal"
	}
	d.sparkData.Update(func(values []float64) []float64 {
		if len(values) == 0 {
			return []float64{float64(d.requests % 100)}
		}
		values = append(values[1:], float64(d.requests%100))
		return values
	})
	if d.latencyData != nil {
		d.latencyData.Update(func(values []widgets.BarData) []widgets.BarData {
			if len(values) == 0 {
				return values
			}
			updated := make([]widgets.BarData, len(values))
			for i, entry := range values {
				delta := float64(rand.Intn(9) - 4)
				value := entry.Value + delta
				if value < 5 {
					value = 5
				}
				updated[i] = widgets.BarData{Label: entry.Label, Value: value}
			}
			return updated
		})
	}
	d.updateMetrics()
	d.Invalidate()
}

func (d *DashboardView) updateMetrics() {
	if d.requestsLabel != nil {
		d.requestsLabel.SetText(fmt.Sprintf("Requests: %d", d.requests))
	}
	if d.errorsLabel != nil {
		d.errorsLabel.SetText(fmt.Sprintf("Errors: %d", d.errors))
	}
	if d.uptimeLabel != nil {
		d.uptimeLabel.SetText("Uptime: " + d.uptime.Truncate(time.Second).String())
	}
}
