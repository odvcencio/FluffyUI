package widgets

import (
	"strconv"

	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/state"
)

// Sparkline renders a compact single-line chart.
type Sparkline struct {
	Base
	Data  *state.Signal[[]float64]
	Width int
	Style backend.Style
}

// NewSparkline creates a sparkline.
func NewSparkline(data *state.Signal[[]float64]) *Sparkline {
	return &Sparkline{
		Data:  data,
		Style: backend.DefaultStyle(),
	}
}

// Measure returns desired size.
func (s *Sparkline) Measure(constraints runtime.Constraints) runtime.Size {
	width := s.Width
	if width <= 0 {
		width = constraints.MaxWidth
	}
	if width <= 0 {
		width = constraints.MinWidth
	}
	return constraints.Constrain(runtime.Size{Width: width, Height: 1})
}

// Render draws the sparkline.
func (s *Sparkline) Render(ctx runtime.RenderContext) {
	if s == nil || s.Data == nil {
		return
	}
	bounds := s.bounds
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	values := s.Data.Get()
	if len(values) == 0 {
		return
	}
	chars := []rune{' ', '.', ':', '-', '=', '+', '*', '#', '@'}
	min, max := values[0], values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	if max == min {
		max = min + 1
	}
	for i := 0; i < bounds.Width; i++ {
		idx := i * len(values) / bounds.Width
		if idx >= len(values) {
			idx = len(values) - 1
		}
		ratio := (values[idx] - min) / (max - min)
		level := int(ratio * float64(len(chars)-1))
		if level < 0 {
			level = 0
		}
		if level >= len(chars) {
			level = len(chars) - 1
		}
		ctx.Buffer.Set(bounds.X+i, bounds.Y, chars[level], s.Style)
	}
}

// HandleMessage returns unhandled.
func (s *Sparkline) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}

// BarData describes a bar entry.
type BarData struct {
	Label string
	Value float64
}

// BarChart renders horizontal bars.
type BarChart struct {
	Base
	Data       *state.Signal[[]BarData]
	ShowValues bool
	ShowLabels bool
	Style      backend.Style
}

// NewBarChart creates a bar chart.
func NewBarChart(data *state.Signal[[]BarData]) *BarChart {
	return &BarChart{
		Data:       data,
		ShowValues: true,
		ShowLabels: true,
		Style:      backend.DefaultStyle(),
	}
}

// Measure returns desired size.
func (b *BarChart) Measure(constraints runtime.Constraints) runtime.Size {
	height := 0
	if b != nil && b.Data != nil {
		height = len(b.Data.Get())
	}
	if height <= 0 {
		height = constraints.MinHeight
	}
	return constraints.Constrain(runtime.Size{Width: constraints.MaxWidth, Height: height})
}

// Render draws the bars.
func (b *BarChart) Render(ctx runtime.RenderContext) {
	if b == nil || b.Data == nil {
		return
	}
	bounds := b.bounds
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	entries := b.Data.Get()
	if len(entries) == 0 {
		return
	}
	maxVal := entries[0].Value
	for _, entry := range entries {
		if entry.Value > maxVal {
			maxVal = entry.Value
		}
	}
	if maxVal <= 0 {
		maxVal = 1
	}
	for i := 0; i < bounds.Height && i < len(entries); i++ {
		entry := entries[i]
		label := ""
		if b.ShowLabels {
			label = entry.Label + " "
		}
		barWidth := bounds.Width - len(label)
		if barWidth < 1 {
			barWidth = 1
		}
		fill := int((entry.Value / maxVal) * float64(barWidth))
		if fill < 0 {
			fill = 0
		}
		if fill > barWidth {
			fill = barWidth
		}
		bar := ""
		for j := 0; j < barWidth; j++ {
			if j < fill {
				bar += "#"
			} else {
				bar += "-"
			}
		}
		line := label + bar
		if b.ShowValues {
			line += " " + formatFloat(entry.Value)
		}
		line = truncateString(line, bounds.Width)
		writePadded(ctx.Buffer, bounds.X, bounds.Y+i, bounds.Width, line, b.Style)
	}
}

// HandleMessage returns unhandled.
func (b *BarChart) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}
