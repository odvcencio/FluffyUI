package widgets

import (
	"fmt"
	"strings"

	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
)

// Progress displays a determinate progress bar.
type Progress struct {
	Base
	Value       float64
	Max         float64
	Label       string
	ShowPercent bool
	Style       GaugeStyle
}

// NewProgress creates a progress widget.
func NewProgress() *Progress {
	p := &Progress{
		Max:         100,
		ShowPercent: true,
		Style:       GaugeStyle{EmptyStyle: backend.DefaultStyle()},
	}
	p.Base.Role = accessibility.RoleProgressBar
	p.syncA11y()
	return p
}

// StyleType returns the selector type name.
func (p *Progress) StyleType() string {
	return "Progress"
}

// Measure returns desired size.
func (p *Progress) Measure(constraints runtime.Constraints) runtime.Size {
	return p.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		return contentConstraints.Constrain(runtime.Size{Width: contentConstraints.MaxWidth, Height: 1})
	})
}

// Render draws the progress bar.
func (p *Progress) Render(ctx runtime.RenderContext) {
	if p == nil {
		return
	}
	p.syncA11y()
	bounds := p.ContentBounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	baseStyle := resolveBaseStyle(ctx, p, backend.DefaultStyle(), false)
	max := p.Max
	if max <= 0 {
		max = 1
	}
	ratio := p.Value / max
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	style := mergeGaugeStyle(baseStyle, p.Style)
	DrawGauge(ctx.Buffer, bounds.X, bounds.Y, bounds.Width, ratio, style)
	if p.ShowPercent && bounds.Width >= 4 {
		text := fmt.Sprintf("%3.0f%%", ratio*100)
		ctx.Buffer.SetString(bounds.X+bounds.Width-textWidth(text), bounds.Y, text, baseStyle)
	}
}

// HandleMessage returns unhandled.
func (p *Progress) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}

func (p *Progress) syncA11y() {
	if p == nil {
		return
	}
	if p.Base.Role == "" {
		p.Base.Role = accessibility.RoleProgressBar
	}
	label := strings.TrimSpace(p.Label)
	if label == "" {
		label = "Progress"
	}
	p.Base.Label = label
	max := p.Max
	if max <= 0 {
		max = 1
	}
	value := p.Value
	if value < 0 {
		value = 0
	}
	if value > max {
		value = max
	}
	ratio := value / max
	p.Base.Value = &accessibility.ValueInfo{
		Min:     0,
		Max:     max,
		Current: value,
		Text:    fmt.Sprintf("%3.0f%%", ratio*100),
	}
}

func mergeGaugeStyle(base backend.Style, style GaugeStyle) GaugeStyle {
	merged := style
	merged.EmptyStyle = mergeBackendStyles(base, style.EmptyStyle)
	if style.EdgeStyle != (backend.Style{}) {
		merged.EdgeStyle = mergeBackendStyles(base, style.EdgeStyle)
	}
	if len(style.Thresholds) > 0 {
		merged.Thresholds = make([]GaugeThreshold, len(style.Thresholds))
		for i, threshold := range style.Thresholds {
			merged.Thresholds[i] = threshold
			merged.Thresholds[i].Style = mergeBackendStyles(base, threshold.Style)
		}
	}
	return merged
}
