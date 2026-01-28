package widgets

import (
	"math"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/graphics"
	"github.com/odvcencio/fluffyui/runtime"
)

// ChartSeries represents a line chart series.
type ChartSeries struct {
	Data   []float64
	Color  backend.Color
	Smooth bool
	Fill   bool
}

// Axis controls min/max scaling for chart values.
type Axis struct {
	Min  float64
	Max  float64
	Auto bool
}

// LineChart renders one or more series using a CanvasWidget.
type LineChart struct {
	CanvasWidget
	series []ChartSeries
	yAxis  Axis
	label  string
}

// NewLineChart creates an empty line chart.
func NewLineChart() *LineChart {
	chart := &LineChart{
		yAxis: Axis{Auto: true},
		label: "Line Chart",
	}
	chart.CanvasWidget = *NewCanvasWidget(chart.drawChart)
	return chart
}

// StyleType returns the selector type name.
func (c *LineChart) StyleType() string { return "LineChart" }

// SetSeries replaces the chart series.
func (c *LineChart) SetSeries(series []ChartSeries) {
	if c == nil {
		return
	}
	c.series = append([]ChartSeries(nil), series...)
	c.Invalidate()
}

// AddSeries appends a new series.
func (c *LineChart) AddSeries(series ChartSeries) {
	if c == nil {
		return
	}
	c.series = append(c.series, series)
	c.Invalidate()
}

// SetYAxis fixes the Y axis range.
func (c *LineChart) SetYAxis(minValue, maxValue float64) {
	if c == nil {
		return
	}
	c.yAxis = Axis{Min: minValue, Max: maxValue, Auto: false}
	c.Invalidate()
}

// AutoYAxis enables auto-scaling on the Y axis.
func (c *LineChart) AutoYAxis() {
	if c == nil {
		return
	}
	c.yAxis.Auto = true
	c.Invalidate()
}

func (c *LineChart) drawChart(canvas *graphics.Canvas) {
	if c == nil || canvas == nil {
		return
	}
	w, h := canvas.Size()
	if w <= 0 || h <= 0 || len(c.series) == 0 {
		return
	}

	minY, maxY := c.yAxis.Min, c.yAxis.Max
	if c.yAxis.Auto {
		minY, maxY = chartSeriesRange(c.series)
	}
	if maxY == minY {
		maxY = minY + 1
	}

	for _, s := range c.series {
		points := chartSeriesPoints(s.Data, w, h, minY, maxY)
		if len(points) < 2 {
			continue
		}
		canvas.SetStrokeColor(s.Color)
		if s.Smooth {
			canvas.DrawSpline(points)
		} else {
			for i := 1; i < len(points); i++ {
				canvas.DrawLineAA(points[i-1].X, points[i-1].Y, points[i].X, points[i].Y)
			}
		}
		if s.Fill {
			fillPoints := make([]graphics.Point, 0, len(points)+2)
			fillPoints = append(fillPoints, points...)
			fillPoints = append(fillPoints,
				graphics.Point{X: w - 1, Y: h - 1},
				graphics.Point{X: 0, Y: h - 1},
			)
			canvas.SetFillColor(dimColor(s.Color, 0.3))
			canvas.FillPolygon(fillPoints)
		}
	}
}

func chartSeriesRange(series []ChartSeries) (float64, float64) {
	minY := 0.0
	maxY := 1.0
	initialized := false
	for _, s := range series {
		for _, v := range s.Data {
			if !initialized {
				minY = v
				maxY = v
				initialized = true
				continue
			}
			if v < minY {
				minY = v
			}
			if v > maxY {
				maxY = v
			}
		}
	}
	if !initialized {
		return 0, 1
	}
	return minY, maxY
}

func chartSeriesPoints(data []float64, w, h int, minY, maxY float64) []graphics.Point {
	if len(data) < 2 {
		return nil
	}
	points := make([]graphics.Point, len(data))
	span := maxY - minY
	if span == 0 {
		span = 1
	}
	for i, v := range data {
		x := int(math.Round(float64(i) / float64(len(data)-1) * float64(w-1)))
		y := int(math.Round((1 - (v-minY)/span) * float64(h-1)))
		points[i] = graphics.Point{X: x, Y: y}
	}
	return points
}

func dimColor(color backend.Color, factor float64) backend.Color {
	if factor <= 0 {
		return backend.ColorBlack
	}
	if factor >= 1 {
		return color
	}
	if !color.IsRGB() {
		return color
	}
	r, g, b := color.RGB()
	return backend.ColorRGB(uint8(float64(r)*factor), uint8(float64(g)*factor), uint8(float64(b)*factor))
}

// Measure keeps the chart flexible within constraints.
func (c *LineChart) Measure(constraints runtime.Constraints) runtime.Size {
	return c.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		return contentConstraints.MaxSize()
	})
}

var _ runtime.Widget = (*LineChart)(nil)
