package widgets

import (
	"testing"

	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/runtime"
)

func TestLineChart_NewLineChart(t *testing.T) {
	chart := NewLineChart()
	if chart == nil {
		t.Fatal("expected non-nil chart")
	}
	if chart.label != "Line Chart" {
		t.Errorf("expected label 'Line Chart', got %q", chart.label)
	}
	if !chart.yAxis.Auto {
		t.Error("expected auto Y axis by default")
	}
}

func TestLineChart_SetSeries(t *testing.T) {
	chart := NewLineChart()
	series := []ChartSeries{
		{Data: []float64{1, 2, 3}, Color: backend.ColorRed},
		{Data: []float64{4, 5, 6}, Color: backend.ColorBlue},
	}
	chart.SetSeries(series)

	if len(chart.series) != 2 {
		t.Fatalf("expected 2 series, got %d", len(chart.series))
	}
	if chart.series[0].Color != backend.ColorRed {
		t.Error("expected first series to be red")
	}
}

func TestLineChart_AddSeries(t *testing.T) {
	chart := NewLineChart()
	chart.AddSeries(ChartSeries{Data: []float64{1, 2, 3}})
	chart.AddSeries(ChartSeries{Data: []float64{4, 5, 6}})

	if len(chart.series) != 2 {
		t.Fatalf("expected 2 series, got %d", len(chart.series))
	}
}

func TestLineChart_SetYAxis(t *testing.T) {
	chart := NewLineChart()
	chart.SetYAxis(0, 100)

	if chart.yAxis.Auto {
		t.Error("expected auto to be false")
	}
	if chart.yAxis.Min != 0 || chart.yAxis.Max != 100 {
		t.Errorf("expected min=0, max=100, got min=%f, max=%f", chart.yAxis.Min, chart.yAxis.Max)
	}
}

func TestLineChart_AutoYAxis(t *testing.T) {
	chart := NewLineChart()
	chart.SetYAxis(0, 100)
	chart.AutoYAxis()

	if !chart.yAxis.Auto {
		t.Error("expected auto to be true")
	}
}

func TestLineChart_StyleType(t *testing.T) {
	chart := NewLineChart()
	if chart.StyleType() != "LineChart" {
		t.Errorf("expected StyleType 'LineChart', got %q", chart.StyleType())
	}
}

func TestLineChart_Measure(t *testing.T) {
	chart := NewLineChart()
	constraints := runtime.Constraints{
		MinWidth:  10,
		MinHeight: 5,
		MaxWidth:  100,
		MaxHeight: 50,
	}
	size := chart.Measure(constraints)
	if size.Width <= 0 || size.Height <= 0 {
		t.Errorf("expected positive size, got %dx%d", size.Width, size.Height)
	}
}

func TestLineChart_Layout(t *testing.T) {
	chart := NewLineChart()
	bounds := runtime.Rect{X: 0, Y: 0, Width: 80, Height: 24}
	chart.Layout(bounds)

	// Verify bounds were stored
	storedBounds := chart.Bounds()
	if storedBounds.Width <= 0 || storedBounds.Height <= 0 {
		t.Error("expected positive bounds after layout")
	}
}

func TestLineChart_HandleMessage(t *testing.T) {
	chart := NewLineChart()
	result := chart.HandleMessage(runtime.ResizeMsg{})
	if result.Handled {
		t.Error("expected message to be unhandled")
	}
}

func TestLineChart_NilSafe(t *testing.T) {
	var chart *LineChart

	// These should not panic
	chart.SetSeries(nil)
	chart.AddSeries(ChartSeries{})
	chart.SetYAxis(0, 100)
	chart.AutoYAxis()
	chart.Layout(runtime.Rect{})
	chart.Render(runtime.RenderContext{})
	result := chart.HandleMessage(runtime.ResizeMsg{})
	if result.Handled {
		t.Error("expected unhandled for nil chart")
	}
}

func TestChartSeriesRange(t *testing.T) {
	tests := []struct {
		name   string
		series []ChartSeries
		minY   float64
		maxY   float64
	}{
		{
			name:   "empty",
			series: []ChartSeries{},
			minY:   0,
			maxY:   1,
		},
		{
			name: "single point",
			series: []ChartSeries{
				{Data: []float64{5}},
			},
			minY: 5,
			maxY: 5,
		},
		{
			name: "multiple points",
			series: []ChartSeries{
				{Data: []float64{1, 5, 3}},
			},
			minY: 1,
			maxY: 5,
		},
		{
			name: "multiple series",
			series: []ChartSeries{
				{Data: []float64{1, 2, 3}},
				{Data: []float64{0, 10}},
			},
			minY: 0,
			maxY: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minY, maxY := chartSeriesRange(tt.series)
			if minY != tt.minY {
				t.Errorf("expected minY=%f, got %f", tt.minY, minY)
			}
			if maxY != tt.maxY {
				t.Errorf("expected maxY=%f, got %f", tt.maxY, maxY)
			}
		})
	}
}

func TestChartSeriesPoints(t *testing.T) {
	t.Run("single point", func(t *testing.T) {
		points := chartSeriesPoints([]float64{5}, 100, 50, 0, 10)
		if points != nil {
			t.Error("expected nil for single point")
		}
	})

	t.Run("two points", func(t *testing.T) {
		points := chartSeriesPoints([]float64{0, 10}, 100, 50, 0, 10)
		if len(points) != 2 {
			t.Fatalf("expected 2 points, got %d", len(points))
		}
		// First point should be at x=0, y=49 (high value = low y)
		if points[0].X != 0 {
			t.Errorf("expected first X=0, got %d", points[0].X)
		}
		// Last point should be at x=99, y=0 (10 at max)
		if points[1].X != 99 {
			t.Errorf("expected last X=99, got %d", points[1].X)
		}
	})
}

func TestDimColor(t *testing.T) {
	tests := []struct {
		name   string
		color  backend.Color
		factor float64
	}{
		{"zero factor", backend.ColorRed, 0},
		{"full factor", backend.ColorRed, 1},
		{"half factor", backend.ColorRed, 0.5},
		{"negative factor", backend.ColorRed, -1},
		{"over factor", backend.ColorRed, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			_ = dimColor(tt.color, tt.factor)
		})
	}
}
