package widgets

import (
	"fmt"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

// Link is a focusable text widget that renders as an OSC-8 hyperlink.
// In terminals that support OSC-8, the text becomes clickable.
type Link struct {
	FocusableBase

	url        string
	label      string
	style      backend.Style
	focusStyle backend.Style
	onActivate func(url string)
	services   runtime.Services
}

var (
	_ runtime.Widget     = (*Link)(nil)
	_ runtime.Focusable  = (*Link)(nil)
	_ runtime.Bindable   = (*Link)(nil)
	_ runtime.Unbindable = (*Link)(nil)
)

// NewLink creates a new hyperlink widget.
func NewLink(label, url string) *Link {
	l := &Link{
		url:        url,
		label:      label,
		style:      backend.DefaultStyle().Underline(true),
		focusStyle: backend.DefaultStyle().Underline(true).Reverse(true),
	}
	l.Base.Role = accessibility.RoleLink
	l.Base.Label = label
	l.Base.Description = url
	return l
}

// Bind attaches app services.
func (l *Link) Bind(services runtime.Services) {
	if l == nil {
		return
	}
	l.services = services
}

// Unbind releases app services.
func (l *Link) Unbind() {
	if l == nil {
		return
	}
	l.services = runtime.Services{}
}

// SetOnActivate sets the callback when the link is activated (Enter key).
func (l *Link) SetOnActivate(fn func(url string)) {
	l.onActivate = fn
}

// URL returns the link URL.
func (l *Link) URL() string {
	return l.url
}

// SetURL updates the link URL.
func (l *Link) SetURL(url string) {
	l.url = url
	l.Base.Description = url
}

// SetLabel updates the link text.
func (l *Link) SetLabel(label string) {
	l.label = label
	l.Base.Label = label
}

// SetStyle sets the normal style.
func (l *Link) SetStyle(s backend.Style) {
	l.style = s
}

// SetFocusStyle sets the focused style.
func (l *Link) SetFocusStyle(s backend.Style) {
	l.focusStyle = s
}

// Measure returns the size needed to display the link text.
func (l *Link) Measure(constraints runtime.Constraints) runtime.Size {
	return l.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		w := textWidth(l.label)
		if w < 1 {
			w = 1
		}
		return contentConstraints.Constrain(runtime.Size{Width: w, Height: 1})
	})
}

// Render draws the link text with underline styling.
func (l *Link) Render(ctx runtime.RenderContext) {
	if l == nil {
		return
	}
	bounds := l.ContentBounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}

	style := l.style
	if l.focused {
		style = l.focusStyle
	}

	text := clipString(l.label, bounds.Width)
	writePadded(ctx.Buffer, bounds.X, bounds.Y, bounds.Width, text, style)
}

// HandleMessage processes input events.
func (l *Link) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if l == nil || !l.focused {
		return runtime.Unhandled()
	}
	switch m := msg.(type) {
	case runtime.KeyMsg:
		if m.Key == terminal.KeyEnter {
			if l.onActivate != nil {
				l.onActivate(l.url)
			}
			if announcer := l.services.Announcer(); announcer != nil {
				announcer.Announce(fmt.Sprintf("%s activated", l.label), accessibility.PriorityPolite)
			}
			return runtime.Handled()
		}
	}
	return runtime.Unhandled()
}
