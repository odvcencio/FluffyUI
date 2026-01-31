package widgets

import (
	"strings"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	uistyle "github.com/odvcencio/fluffyui/style"
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
	styleSet bool
}

// NewAlert creates an alert.
func NewAlert(text string, variant AlertVariant) *Alert {
	alert := &Alert{
		Text:    text,
		Variant: variant,
		style:   backend.DefaultStyle(),
	}
	alert.Base.Role = accessibility.RoleAlert
	alert.Base.Label = text
	return alert
}

// SetStyle updates the alert style.
func (a *Alert) SetStyle(style backend.Style) {
	if a == nil {
		return
	}
	a.style = style
	a.styleSet = true
}

// StyleType returns the selector type name.
func (a *Alert) StyleType() string {
	return "Alert"
}

// StyleClasses returns selector classes including the variant.
func (a *Alert) StyleClasses() []string {
	if a == nil {
		return nil
	}
	classes := a.Base.StyleClasses()
	if a.Variant == "" {
		return classes
	}
	variant := string(a.Variant)
	for _, cls := range classes {
		if cls == variant {
			return classes
		}
	}
	return append(classes, variant)
}

// Measure returns desired size.
func (a *Alert) Measure(constraints runtime.Constraints) runtime.Size {
	return a.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		width := textWidth(a.Text)
		if width < 1 {
			width = 1
		}
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: 1})
	})
}

// Render draws the alert.
func (a *Alert) Render(ctx runtime.RenderContext) {
	if a == nil {
		return
	}
	a.syncA11y()
	outer := a.bounds
	content := a.ContentBounds()
	if outer.Width <= 0 || outer.Height <= 0 {
		return
	}
	resolved := ctx.ResolveStyle(a)
	style := a.style
	if !resolved.IsZero() {
		final := resolved
		if a.styleSet {
			final = final.Merge(uistyle.FromBackend(a.style))
		}
		style = final.ToBackend()
	} else {
		switch a.Variant {
		case AlertSuccess:
			style = style.Bold(true)
		case AlertWarning:
			style = style.Bold(true)
		case AlertError:
			style = style.Bold(true).Underline(true)
		}
	}
	ctx.Buffer.Fill(outer, ' ', style)
	if content.Width <= 0 || content.Height <= 0 {
		return
	}
	text := truncateString(a.Text, content.Width)
	writePadded(ctx.Buffer, content.X, content.Y, content.Width, text, style)
}

// HandleMessage returns unhandled.
func (a *Alert) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}

func (a *Alert) syncA11y() {
	if a == nil {
		return
	}
	if a.Base.Role == "" {
		a.Base.Role = accessibility.RoleAlert
	}
	label := strings.TrimSpace(a.Text)
	if label == "" {
		label = "Alert"
	}
	a.Base.Label = label
	if a.Variant != "" {
		a.Base.Description = string(a.Variant)
	}
}

var _ runtime.Widget = (*Alert)(nil)
