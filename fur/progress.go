package fur

import (
	"fmt"
	"math"
	"strings"
)

// Progress renders a simple progress bar.
type Progress struct {
	total       int
	current     int
	label       string
	showPercent bool
}

// NewProgress creates a new progress bar with the given total.
func NewProgress(total int) *Progress {
	if total < 0 {
		total = 0
	}
	return &Progress{total: total, showPercent: true}
}

// WithLabel sets the progress label.
func (p *Progress) WithLabel(label string) *Progress {
	if p == nil {
		return p
	}
	p.label = label
	return p
}

// WithShowPercent toggles percent display.
func (p *Progress) WithShowPercent(show bool) *Progress {
	if p == nil {
		return p
	}
	p.showPercent = show
	return p
}

// Set updates the current progress value.
func (p *Progress) Set(value int) {
	if p == nil {
		return
	}
	p.current = value
}

// Render renders the progress bar.
func (p *Progress) Render(width int) []Line {
	if p == nil {
		return nil
	}
	if width <= 0 {
		width = 40
	}
	total := p.total
	if total <= 0 {
		total = 1
	}
	current := p.current
	if current < 0 {
		current = 0
	}
	if current > total {
		current = total
	}
	ratio := float64(current) / float64(total)
	prefix := ""
	if p.label != "" {
		prefix = p.label + " "
	}
	percent := ""
	if p.showPercent {
		percent = fmt.Sprintf(" %3.0f%%", ratio*100)
	}
	barWidth := width - stringWidth(prefix) - stringWidth(percent) - 2
	if barWidth < 4 {
		barWidth = 4
	}
	filled := int(math.Round(ratio * float64(barWidth)))
	if filled > barWidth {
		filled = barWidth
	}
	empty := barWidth - filled
	var line Line
	if prefix != "" {
		appendSpan(&line, Span{Text: prefix, Style: DefaultStyle()})
	}
	appendSpan(&line, Span{Text: "[", Style: DefaultStyle()})
	if filled > 0 {
		appendSpan(&line, Span{Text: strings.Repeat("=", filled), Style: DefaultStyle().Foreground(ColorGreen)})
	}
	if empty > 0 {
		appendSpan(&line, Span{Text: strings.Repeat("-", empty), Style: Dim})
	}
	appendSpan(&line, Span{Text: "]", Style: DefaultStyle()})
	if percent != "" {
		appendSpan(&line, Span{Text: percent, Style: Dim})
	}
	return []Line{line}
}
