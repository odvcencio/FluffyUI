package widgets

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
)

// Sparkline renders a compact single-line chart.
type Sparkline struct {
	Base
	Data  *state.Signal[[]float64]
	Width int
	Style backend.Style
	label string
}

// NewSparkline creates a sparkline.
func NewSparkline(data *state.Signal[[]float64]) *Sparkline {
	s := &Sparkline{
		Data:  data,
		Style: backend.DefaultStyle(),
		label: "Sparkline",
	}
	s.Base.Role = accessibility.RoleChart
	s.syncA11y()
	return s
}

// StyleType returns the selector type name.
func (s *Sparkline) StyleType() string {
	return "Sparkline"
}

// Measure returns desired size.
func (s *Sparkline) Measure(constraints runtime.Constraints) runtime.Size {
	return s.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		width := s.Width
		if width <= 0 {
			width = contentConstraints.MaxWidth
		}
		if width <= 0 {
			width = contentConstraints.MinWidth
		}
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: 1})
	})
}

// Render draws the sparkline.
func (s *Sparkline) Render(ctx runtime.RenderContext) {
	if s == nil || s.Data == nil {
		return
	}
	s.syncA11y()
	bounds := s.ContentBounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	values := s.Data.Get()
	if len(values) == 0 {
		return
	}
	style := mergeBackendStyles(resolveBaseStyle(ctx, s, backend.DefaultStyle(), false), s.Style)
	chars := []rune{' ', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
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
		ctx.Buffer.Set(bounds.X+i, bounds.Y, chars[level], style)
	}
}

// HandleMessage returns unhandled.
func (s *Sparkline) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}

func (s *Sparkline) syncA11y() {
	if s == nil {
		return
	}
	if s.Base.Role == "" {
		s.Base.Role = accessibility.RoleChart
	}
	label := strings.TrimSpace(s.label)
	if label == "" {
		label = "Sparkline"
	}
	s.Base.Label = label
	values := []float64(nil)
	if s.Data != nil {
		values = s.Data.Get()
	}
	if len(values) == 0 {
		s.Base.Description = "0 points"
		s.Base.Value = nil
		return
	}
	minVal, maxVal := values[0], values[0]
	for _, v := range values {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}
	last := values[len(values)-1]
	s.Base.Description = fmt.Sprintf("%d points", len(values))
	s.Base.Value = &accessibility.ValueInfo{
		Text: fmt.Sprintf("min %s, max %s, last %s", formatFloat(minVal), formatFloat(maxVal), formatFloat(last)),
	}
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
	label      string
}

// NewBarChart creates a bar chart.
func NewBarChart(data *state.Signal[[]BarData]) *BarChart {
	b := &BarChart{
		Data:       data,
		ShowValues: true,
		ShowLabels: true,
		Style:      backend.DefaultStyle(),
		label:      "Bar Chart",
	}
	b.Base.Role = accessibility.RoleChart
	b.syncA11y()
	return b
}

// StyleType returns the selector type name.
func (b *BarChart) StyleType() string {
	return "BarChart"
}

// Measure returns desired size.
func (b *BarChart) Measure(constraints runtime.Constraints) runtime.Size {
	return b.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		height := 0
		if b != nil && b.Data != nil {
			height = len(b.Data.Get())
		}
		if height <= 0 {
			height = contentConstraints.MinHeight
		}
		return contentConstraints.Constrain(runtime.Size{Width: contentConstraints.MaxWidth, Height: height})
	})
}

// Render draws the bars.
func (b *BarChart) Render(ctx runtime.RenderContext) {
	if b == nil || b.Data == nil {
		return
	}
	b.syncA11y()
	bounds := b.ContentBounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	entries := b.Data.Get()
	if len(entries) == 0 {
		return
	}
	style := mergeBackendStyles(resolveBaseStyle(ctx, b, backend.DefaultStyle(), false), b.Style)
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
		labelWidth := textWidth(label)
		barWidth := bounds.Width - labelWidth
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
		bar := strings.Builder{}
		bar.Grow(barWidth)
		for j := 0; j < barWidth; j++ {
			if j < fill {
				bar.WriteRune('█')
			} else {
				bar.WriteRune('░')
			}
		}
		line := label + bar.String()
		if b.ShowValues {
			line += " " + formatFloat(entry.Value)
		}
		line = truncateString(line, bounds.Width)
		writePadded(ctx.Buffer, bounds.X, bounds.Y+i, bounds.Width, line, style)
	}
}

// HandleMessage returns unhandled.
func (b *BarChart) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}

func (b *BarChart) syncA11y() {
	if b == nil {
		return
	}
	if b.Base.Role == "" {
		b.Base.Role = accessibility.RoleChart
	}
	label := strings.TrimSpace(b.label)
	if label == "" {
		label = "Bar Chart"
	}
	b.Base.Label = label
	entries := []BarData(nil)
	if b.Data != nil {
		entries = b.Data.Get()
	}
	if len(entries) == 0 {
		b.Base.Description = "0 bars"
		b.Base.Value = nil
		return
	}
	maxEntry := entries[0]
	for _, entry := range entries {
		if entry.Value > maxEntry.Value {
			maxEntry = entry
		}
	}
	b.Base.Description = fmt.Sprintf("%d bars", len(entries))
	if strings.TrimSpace(maxEntry.Label) != "" {
		b.Base.Value = &accessibility.ValueInfo{Text: fmt.Sprintf("%s: %s", maxEntry.Label, formatFloat(maxEntry.Value))}
	} else {
		b.Base.Value = &accessibility.ValueInfo{Text: formatFloat(maxEntry.Value)}
	}
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}
