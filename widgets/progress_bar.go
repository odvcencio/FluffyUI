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

// Measure returns desired size.
func (p *Progress) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.Constrain(runtime.Size{Width: constraints.MaxWidth, Height: 1})
}

// Render draws the progress bar.
func (p *Progress) Render(ctx runtime.RenderContext) {
	if p == nil {
		return
	}
	p.syncA11y()
	bounds := p.bounds
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
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
	DrawGauge(ctx.Buffer, bounds.X, bounds.Y, bounds.Width, ratio, p.Style)
	if p.ShowPercent && bounds.Width >= 4 {
		text := fmt.Sprintf("%3.0f%%", ratio*100)
		ctx.Buffer.SetString(bounds.X+bounds.Width-len(text), bounds.Y, text, backend.DefaultStyle())
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
