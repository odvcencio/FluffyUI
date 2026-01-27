package fur

import (
	"fmt"
	"math"
	"strings"
)

// Sparkline renders a sparkline chart from a series of values.
// Uses Unicode block characters for a crisp display.
func Sparkline(values []float64, width int) Renderable {
	return sparklineRenderable{values: values, width: width}
}

type sparklineRenderable struct {
	values []float64
	width  int
	style  Style
}

// WithStyle sets the sparkline style.
func (s sparklineRenderable) WithStyle(style Style) sparklineRenderable {
	s.style = style
	return s
}

func (s sparklineRenderable) Render(width int) []Line {
	if len(s.values) == 0 {
		return nil
	}
	
	w := s.width
	if w <= 0 {
		w = width
	}
	if w <= 0 {
		w = 40
	}
	
	// Find min/max
	minVal, maxVal := s.values[0], s.values[0]
	for _, v := range s.values {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}
	
	if minVal == maxVal {
		maxVal = minVal + 1
	}
	
	// Sparkline characters (from low to high)
	blocks := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	
	// Sample or aggregate values to fit width
	var displayValues []float64
	if len(s.values) <= w {
		displayValues = s.values
	} else {
		// Aggregate by averaging
		step := float64(len(s.values)) / float64(w)
		for i := 0; i < w; i++ {
			start := int(float64(i) * step)
			end := int(float64(i+1) * step)
			if end > len(s.values) {
				end = len(s.values)
			}
			if start >= len(s.values) {
				break
			}
			sum := 0.0
			count := 0
			for j := start; j < end; j++ {
				sum += s.values[j]
				count++
			}
			if count > 0 {
				displayValues = append(displayValues, sum/float64(count))
			}
		}
	}
	
	// Convert to characters
	style := s.style
	if style == (Style{}) {
		style = DefaultStyle()
	}
	
	var chars []string
	for _, v := range displayValues {
		normalized := (v - minVal) / (maxVal - minVal)
		idx := int(normalized * float64(len(blocks)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(blocks) {
			idx = len(blocks) - 1
		}
		chars = append(chars, blocks[idx])
	}
	
	return []Line{{{Text: strings.Join(chars, ""), Style: style}}}
}

// BarChart renders a horizontal bar chart.
func BarChart(labels []string, values []float64, maxWidth int) Renderable {
	return barChartRenderable{
		labels:   labels,
		values:   values,
		maxWidth: maxWidth,
	}
}

type barChartRenderable struct {
	labels    []string
	values    []float64
	maxWidth  int
	barStyle  Style
	labelStyle Style
}

// WithBarStyle sets the bar style.
func (b barChartRenderable) WithBarStyle(style Style) barChartRenderable {
	b.barStyle = style
	return b
}

// WithLabelStyle sets the label style.
func (b barChartRenderable) WithLabelStyle(style Style) barChartRenderable {
	b.labelStyle = style
	return b
}

func (b barChartRenderable) Render(width int) []Line {
	if len(b.values) == 0 || len(b.labels) != len(b.values) {
		return nil
	}
	
	maxW := b.maxWidth
	if maxW <= 0 {
		maxW = width - 20
	}
	if maxW <= 0 {
		maxW = 40
	}
	
	// Find max value
	maxVal := b.values[0]
	for _, v := range b.values {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal == 0 {
		maxVal = 1
	}
	
	// Find max label width
	maxLabelWidth := 0
	for _, label := range b.labels {
		if w := stringWidth(label); w > maxLabelWidth {
			maxLabelWidth = w
		}
	}
	
	barStyle := b.barStyle
	if barStyle == (Style{}) {
		barStyle = Style{}.Foreground(ColorGreen)
	}
	labelStyle := b.labelStyle
	if labelStyle == (Style{}) {
		labelStyle = DefaultStyle()
	}
	
	// Bar characters
	fullBlock := "█"
	// sevenEighths := "▉"
	// threeQuarters := "▊"
	// fiveEighths := "▋"
	// halfBlock := "▌"
	// threeEighths := "▍"
	// quarterBlock := "▎"
	// oneEighth := "▏"
	
	var lines []Line
	for i, value := range b.values {
		label := b.labels[i]
		if len(b.labels) > i {
			label = b.labels[i]
		}
		
		// Calculate bar length
		ratio := value / maxVal
		barLength := int(ratio * float64(maxW))
		
		// Pad label
		labelPadding := strings.Repeat(" ", maxLabelWidth-stringWidth(label))
		
		// Build bar
		bar := strings.Repeat(fullBlock, barLength)
		
		// Format value
		valueStr := fmt.Sprintf(" %.1f", value)
		
		line := Line{
			{Text: label + labelPadding + " ", Style: labelStyle},
			{Text: bar, Style: barStyle},
			{Text: valueStr, Style: labelStyle},
		}
		lines = append(lines, line)
	}
	
	return lines
}

// PieChart renders a simple ASCII pie chart.
func PieChart(values []float64, labels []string) Renderable {
	return pieChartRenderable{values: values, labels: labels}
}

type pieChartRenderable struct {
	values []float64
	labels []string
}

func (p pieChartRenderable) Render(width int) []Line {
	if len(p.values) == 0 {
		return nil
	}
	
	// Calculate total
	total := 0.0
	for _, v := range p.values {
		total += v
	}
	if total == 0 {
		total = 1
	}
	
	// Calculate percentages
	var lines []Line
	for i, v := range p.values {
		percent := (v / total) * 100
		
		// Create simple bar representation
		filled := int(percent / 5) // 20 chars = 100%
		if filled > 20 {
			filled = 20
		}
		bar := strings.Repeat("█", filled) + strings.Repeat("░", 20-filled)
		
		label := fmt.Sprintf("%.0f%%", percent)
		if i < len(p.labels) {
			label = fmt.Sprintf("%s %.0f%%", p.labels[i], percent)
		}
		
		// Color based on index
		colors := []Color{ColorGreen, ColorBlue, ColorYellow, ColorRed, ColorMagenta, ColorCyan}
		color := colors[i%len(colors)]
		
		lines = append(lines, Line{
			{Text: label + " ", Style: DefaultStyle()},
			{Text: bar, Style: Style{}.Foreground(color)},
		})
	}
	
	return lines
}

// Heatmap renders a heatmap from a 2D array of values.
func Heatmap(data [][]float64, width int) Renderable {
	return heatmapRenderable{data: data, width: width}
}

type heatmapRenderable struct {
	data  [][]float64
	width int
}

func (h heatmapRenderable) Render(width int) []Line {
	if len(h.data) == 0 {
		return nil
	}
	
	// Find min/max
	minVal, maxVal := h.data[0][0], h.data[0][0]
	for _, row := range h.data {
		for _, v := range row {
			if v < minVal {
				minVal = v
			}
			if v > maxVal {
				maxVal = v
			}
		}
	}
	if minVal == maxVal {
		maxVal = minVal + 1
	}
	
	// Heatmap characters (density)
	blocks := []string{"░", "▒", "▓", "█"}
	
	// Colors for heat (blue to red)
	colors := []Color{
		ColorBlue,
		ColorCyan,
		ColorGreen,
		ColorYellow,
		ColorRed,
	}
	
	var lines []Line
	for _, row := range h.data {
		var line Line
		for _, v := range row {
			normalized := (v - minVal) / (maxVal - minVal)
			
			charIdx := int(normalized * float64(len(blocks)-1))
			if charIdx < 0 {
				charIdx = 0
			}
			if charIdx >= len(blocks) {
				charIdx = len(blocks) - 1
			}
			
			colorIdx := int(normalized * float64(len(colors)-1))
			if colorIdx < 0 {
				colorIdx = 0
			}
			if colorIdx >= len(colors) {
				colorIdx = len(colors) - 1
			}
			
			line = append(line, Span{
				Text:  blocks[charIdx],
				Style: Style{}.Foreground(colors[colorIdx]),
			})
		}
		lines = append(lines, line)
	}
	
	return lines
}

// ProgressBar renders a progress bar.
func ProgressBar(current, total float64, width int) Renderable {
	return progressBarRenderable{
		current: current,
		total:   total,
		width:   width,
	}
}

type progressBarRenderable struct {
	current float64
	total   float64
	width   int
	style   Style
}

// WithStyle sets the progress bar style.
func (p progressBarRenderable) WithStyle(style Style) progressBarRenderable {
	p.style = style
	return p
}

func (p progressBarRenderable) Render(width int) []Line {
	w := p.width
	if w <= 0 {
		w = width - 10
	}
	if w <= 0 {
		w = 30
	}
	
	percent := 0.0
	if p.total > 0 {
		percent = p.current / p.total
	}
	if percent < 0 {
		percent = 0
	}
	if percent > 1 {
		percent = 1
	}
	
	filled := int(percent * float64(w))
	empty := w - filled
	
	barChars := []string{"█", "▉", "▊", "▋", "▌", "▍", "▎", "▏"}
	
	// Determine color based on progress
	style := p.style
	if style == (Style{}) {
		if percent < 0.3 {
			style = Style{}.Foreground(ColorRed)
		} else if percent < 0.7 {
			style = Style{}.Foreground(ColorYellow)
		} else {
			style = Style{}.Foreground(ColorGreen)
		}
	}
	
	var bar strings.Builder
	for i := 0; i < filled; i++ {
		bar.WriteString("█")
	}
	
	// Partial block for fractional part
	fraction := (percent * float64(w)) - float64(filled)
	if fraction > 0 && filled < w {
		partialIdx := int(fraction * float64(len(barChars)))
		if partialIdx >= len(barChars) {
			partialIdx = len(barChars) - 1
		}
		bar.WriteString(barChars[partialIdx])
		empty--
	}
	
	for i := 0; i < empty; i++ {
		bar.WriteString("░")
	}
	
	percentStr := fmt.Sprintf(" %3.0f%%", percent*100)
	
	return []Line{{
		{Text: "[", Style: DefaultStyle()},
		{Text: bar.String(), Style: style},
		{Text: "]", Style: DefaultStyle()},
		{Text: percentStr, Style: DefaultStyle()},
	}}
}

// Gauge renders a circular gauge (using Unicode characters).
func Gauge(value float64, min, max float64, width int) Renderable {
	return gaugeRenderable{
		value: value,
		min:   min,
		max:   max,
		width: width,
	}
}

type gaugeRenderable struct {
	value float64
	min   float64
	max   float64
	width int
}

func (g gaugeRenderable) Render(width int) []Line {
	// Clamp value
	v := g.value
	if v < g.min {
		v = g.min
	}
	if v > g.max {
		v = g.max
	}
	
	// Calculate percentage
	percent := 0.0
	if g.max > g.min {
		percent = (v - g.min) / (g.max - g.min)
	}
	
	// Draw a horizontal gauge with ticks
	w := g.width
	if w <= 0 {
		w = 40
	}
	
	position := int(percent * float64(w-1))
	
	var bar strings.Builder
	bar.WriteString("|")
	for i := 0; i < w-2; i++ {
		if i == position {
			bar.WriteString("◆")
		} else if i < position {
			bar.WriteString("─")
		} else {
			bar.WriteString("╌")
		}
	}
	bar.WriteString("|")
	
	valueStr := fmt.Sprintf(" %.1f", v)
	
	return []Line{{
		{Text: bar.String(), Style: DefaultStyle()},
		{Text: valueStr, Style: Style{}.Bold()},
	}}
}

// BulletGraph renders a bullet graph (compact comparison of values).
func BulletGraph(actual, target, max float64, width int) Renderable {
	return bulletGraphRenderable{
		actual: actual,
		target: target,
		max:    max,
		width:  width,
	}
}

type bulletGraphRenderable struct {
	actual float64
	target float64
	max    float64
	width  int
}

func (b bulletGraphRenderable) Render(width int) []Line {
	w := b.width
	if w <= 0 {
		w = 40
	}
	
	// Calculate positions
	actualPos := int((b.actual / b.max) * float64(w-1))
	targetPos := int((b.target / b.max) * float64(w-1))
	
	if actualPos < 0 {
		actualPos = 0
	}
	if actualPos >= w {
		actualPos = w - 1
	}
	if targetPos < 0 {
		targetPos = 0
	}
	if targetPos >= w {
		targetPos = w - 1
	}
	
	// Build the bar
	var bar strings.Builder
	for i := 0; i < w; i++ {
		switch {
		case i == targetPos && i == actualPos:
			bar.WriteString("◆") // Both at same position
		case i == targetPos:
			bar.WriteString("│") // Target marker
		case i == actualPos:
			bar.WriteString("●") // Actual value
		case i < actualPos:
			bar.WriteString("█") // Filled
		default:
			bar.WriteString("░") // Empty
		}
	}
	
	return []Line{{
		{Text: bar.String(), Style: DefaultStyle()},
		{Text: fmt.Sprintf(" %.0f/%.0f", b.actual, b.target), Style: Style{}.Dim()},
	}}
}

// normalize normalizes a slice of values to 0-1 range
func normalize(values []float64) []float64 {
	if len(values) == 0 {
		return nil
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
	
	range_val := maxVal - minVal
	if range_val == 0 {
		range_val = 1
	}
	
	result := make([]float64, len(values))
	for i, v := range values {
		result[i] = (v - minVal) / range_val
	}
	return result
}

// Statistics calculates and displays statistics for a dataset.
func Statistics(values []float64) Renderable {
	if len(values) == 0 {
		return Text("No data")
	}
	
	// Calculate statistics
	min, max := values[0], values[0]
	sum := 0.0
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += v
	}
	mean := sum / float64(len(values))
	
	// Calculate standard deviation
	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	stdDev := math.Sqrt(variance / float64(len(values)))
	
	// Render
	return Group(
		Markup(fmt.Sprintf("[bold]Count:[-]    %d", len(values))),
		Markup(fmt.Sprintf("[bold]Min:[-]      %.2f", min)),
		Markup(fmt.Sprintf("[bold]Max:[-]      %.2f", max)),
		Markup(fmt.Sprintf("[bold]Mean:[-]     %.2f", mean)),
		Markup(fmt.Sprintf("[bold]Std Dev:[-]  %.2f", stdDev)),
	)
}
