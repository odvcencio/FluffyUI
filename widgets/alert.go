package widgets

import (
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
)

// AlertVariant describes alert styling.
type AlertVariant string

const (
	AlertInfo    AlertVariant = "info"
	AlertSuccess AlertVariant = "success"
	AlertWarning AlertVariant = "warning"
	AlertError   AlertVariant = "error"
)

// Alert renders an inline message.
type Alert struct {
	Base
	Variant AlertVariant
	Text    string
	style   backend.Style
}

// NewAlert creates an alert.
func NewAlert(text string, variant AlertVariant) *Alert {
	return &Alert{
		Text:    text,
		Variant: variant,
		style:   backend.DefaultStyle(),
	}
}

// Measure returns desired size.
func (a *Alert) Measure(constraints runtime.Constraints) runtime.Size {
	width := len(a.Text)
	if width < 1 {
		width = 1
	}
	return constraints.Constrain(runtime.Size{Width: width, Height: 1})
}

// Render draws the alert.
func (a *Alert) Render(ctx runtime.RenderContext) {
	if a == nil {
		return
	}
	bounds := a.bounds
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	style := a.style
	switch a.Variant {
	case AlertSuccess:
		style = style.Bold(true)
	case AlertWarning:
		style = style.Bold(true)
	case AlertError:
		style = style.Bold(true).Underline(true)
	}
	text := truncateString(a.Text, bounds.Width)
	writePadded(ctx.Buffer, bounds.X, bounds.Y, bounds.Width, text, style)
}

// HandleMessage returns unhandled.
func (a *Alert) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}
