package widgets

import (
	"m31labs.dev/fluffyui/accessibility"
	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/runtime"
)

// Badge is a non-focusable colored label widget.
// It displays a short text like a status indicator or count.
type Badge struct {
	Base

	text     string
	style    backend.Style
	styleSet bool
	services runtime.Services
}

// NewBadge creates a badge with the given text.
func NewBadge(text string) *Badge {
	b := &Badge{
		text:  text,
		style: backend.DefaultStyle(),
	}
	b.Base.Role = accessibility.RoleStatus
	b.Base.Live = accessibility.LivePolite
	b.Base.Label = text
	return b
}

// SetText updates the badge text.
func (b *Badge) SetText(text string) {
	if b == nil {
		return
	}
	b.text = text
	b.Base.Label = text
}

// Text returns the badge text.
func (b *Badge) Text() string {
	if b == nil {
		return ""
	}
	return b.text
}

// SetStyle sets the badge style.
func (b *Badge) SetStyle(style backend.Style) {
	if b == nil {
		return
	}
	b.style = style
	b.styleSet = true
}

// StyleType returns the selector type name.
func (b *Badge) StyleType() string {
	return "Badge"
}

// Measure returns the desired size.
func (b *Badge) Measure(constraints runtime.Constraints) runtime.Size {
	return b.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		width := textWidth(b.text)
		if width < 1 {
			width = 1
		}
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: 1})
	})
}

// Layout stores the assigned bounds.
func (b *Badge) Layout(bounds runtime.Rect) {
	if b == nil {
		return
	}
	b.Base.Layout(bounds)
}

// Render draws the badge.
func (b *Badge) Render(ctx runtime.RenderContext) {
	if b == nil {
		return
	}
	b.syncA11y()
	bounds := b.ContentBounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	style := resolveBaseStyle(ctx, b, b.style, b.styleSet)
	writePadded(ctx.Buffer, bounds.X, bounds.Y, bounds.Width, truncateString(b.text, bounds.Width), style)
}

// HandleMessage returns unhandled (non-interactive widget).
func (b *Badge) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}

func (b *Badge) syncA11y() {
	if b == nil {
		return
	}
	if b.Base.Role == "" {
		b.Base.Role = accessibility.RoleStatus
	}
	b.Base.Label = b.text
	if b.Base.Live == "" {
		b.Base.Live = accessibility.LivePolite
	}
}

// Bind attaches app services.
func (b *Badge) Bind(services runtime.Services) {
	if b == nil {
		return
	}
	b.services = services
}

// Unbind releases app services.
func (b *Badge) Unbind() {
	if b == nil {
		return
	}
	b.services = runtime.Services{}
}

var _ runtime.Widget = (*Badge)(nil)
var _ runtime.Bindable = (*Badge)(nil)
var _ runtime.Unbindable = (*Badge)(nil)
