package widgets

import (
	"fmt"

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
	return &Progress{
		Max:         100,
		ShowPercent: true,
		Style:       GaugeStyle{FillChar: '#', EmptyChar: '-', EmptyStyle: backend.DefaultStyle()},
	}
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
