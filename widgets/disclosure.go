package widgets

import (
	"fmt"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	uistyle "github.com/odvcencio/fluffyui/style"
	"github.com/odvcencio/fluffyui/terminal"
)

// Disclosure is a single expandable section with a toggle header.
type Disclosure struct {
	FocusableBase
	title    string
	content  runtime.Widget
	expanded bool
	onChange func(expanded bool)
	style      backend.Style
	focusStyle backend.Style
	styleSet   bool
	focusSet   bool
	services   runtime.Services
}

// NewDisclosure creates a disclosure widget with the given title and content.
func NewDisclosure(title string, content runtime.Widget) *Disclosure {
	d := &Disclosure{
		title:      title,
		content:    content,
		style:      backend.DefaultStyle(),
		focusStyle: backend.DefaultStyle().Reverse(true),
	}
	d.Base.Role = accessibility.RoleButton
	d.syncA11y()
	return d
}

// Expanded reports whether the disclosure is expanded.
func (d *Disclosure) Expanded() bool {
	if d == nil { return false }
	return d.expanded
}

// SetExpanded sets the expanded state.
func (d *Disclosure) SetExpanded(expanded bool) {
	if d == nil || d.expanded == expanded { return }
	d.expanded = expanded
	d.syncA11y()
	if announcer := d.services.Announcer(); announcer != nil {
		if expanded {
			announcer.Announce(fmt.Sprintf("%s expanded", d.title), accessibility.PriorityPolite)
		} else {
			announcer.Announce(fmt.Sprintf("%s collapsed", d.title), accessibility.PriorityPolite)
		}
	}
	d.relayout()
	if d.onChange != nil { d.onChange(expanded) }
}

// Toggle toggles the expanded state.
func (d *Disclosure) Toggle() {
	if d == nil { return }
	d.SetExpanded(!d.expanded)
}

// SetOnChange sets the change handler.
func (d *Disclosure) SetOnChange(fn func(expanded bool)) {
	if d == nil { return }
	d.onChange = fn
}

// SetStyle sets the normal style.
func (d *Disclosure) SetStyle(style backend.Style) {
	if d == nil { return }
	d.style = style
	d.styleSet = true
}

// SetFocusStyle sets the focused style.
func (d *Disclosure) SetFocusStyle(style backend.Style) {
	if d == nil { return }
	d.focusStyle = style
	d.focusSet = true
}

// StyleType returns the selector type name.
func (d *Disclosure) StyleType() string { return "Disclosure" }

// Bind attaches app services.
func (d *Disclosure) Bind(services runtime.Services) {
	if d == nil { return }
	d.services = services
}

// Unbind releases app services.
func (d *Disclosure) Unbind() {
	if d == nil { return }
	d.services = runtime.Services{}
}

// Measure returns the desired size.
func (d *Disclosure) Measure(constraints runtime.Constraints) runtime.Size {
	return d.measureWithStyle(constraints, func(cc runtime.Constraints) runtime.Size {
		width := textWidth(d.title) + 4
		if width < 6 { width = 6 }
		height := 1
		if d.expanded && d.content != nil {
			childC := runtime.Constraints{
				MinWidth: cc.MinWidth, MaxWidth: cc.MaxWidth,
				MinHeight: 0, MaxHeight: max(0, cc.MaxHeight-1),
			}
			size := d.content.Measure(childC)
			height += size.Height
			if size.Width > width { width = size.Width }
		}
		return cc.Constrain(runtime.Size{Width: width, Height: height})
	})
}

// Layout positions the header and content.
func (d *Disclosure) Layout(bounds runtime.Rect) {
	d.FocusableBase.Layout(bounds)
	if d.expanded && d.content != nil {
		content := d.ContentBounds()
		if content.Height > 1 {
			d.content.Layout(runtime.Rect{
				X: content.X, Y: content.Y + 1,
				Width: content.Width, Height: content.Height - 1,
			})
		}
	}
}

// Render draws the disclosure header and optional content.
func (d *Disclosure) Render(ctx runtime.RenderContext) {
	if d == nil { return }
	outer := d.bounds
	content := d.ContentBounds()
	if outer.Width <= 0 || outer.Height <= 0 { return }
	style := d.style
	resolved := ctx.ResolveStyle(d)
	if !resolved.IsZero() {
		final := resolved
		if d.styleSet { final = final.Merge(uistyle.FromBackend(d.style)) }
		if d.focused && d.focusSet { final = final.Merge(uistyle.FromBackend(d.focusStyle)) }
		style = final.ToBackend()
	} else if d.focused {
		style = d.focusStyle
	}
	ctx.Buffer.Fill(outer, ' ', style)
	if content.Width <= 0 || content.Height <= 0 { return }
	icon := "\xe2\x96\xb8" // U+25B8 right triangle
	if d.expanded { icon = "\xe2\x96\xbe" } // U+25BE down triangle
	label := icon + " " + truncateString(d.title, max(0, content.Width-3))
	writePadded(ctx.Buffer, content.X, content.Y, content.Width, label, style)
	if d.expanded && d.content != nil && content.Height > 1 {
		runtime.RenderChild(ctx, d.content)
	}
}

// HandleMessage handles keyboard toggling.
func (d *Disclosure) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if d == nil || !d.focused { return runtime.Unhandled() }
	key, ok := msg.(runtime.KeyMsg)
	if !ok { return runtime.Unhandled() }
	switch key.Key {
	case terminal.KeyEnter:
		d.Toggle(); return runtime.Handled()
	case terminal.KeyRune:
		if key.Rune == ' ' { d.Toggle(); return runtime.Handled() }
	}
	return runtime.Unhandled()
}

// ChildWidgets returns the content widget when expanded.
func (d *Disclosure) ChildWidgets() []runtime.Widget {
	if d == nil || d.content == nil || !d.expanded { return nil }
	return []runtime.Widget{d.content}
}

func (d *Disclosure) syncA11y() {
	if d == nil { return }
	if d.Base.Role == "" { d.Base.Role = accessibility.RoleButton }
	d.Base.Label = d.title
	d.Base.State.Expanded = accessibility.BoolPtr(d.expanded)
}

func (d *Disclosure) relayout() {
	if d == nil { return }
	d.Invalidate()
	d.services.Relayout()
}

var _ runtime.Widget = (*Disclosure)(nil)
var _ runtime.Focusable = (*Disclosure)(nil)
var _ runtime.ChildProvider = (*Disclosure)(nil)
var _ runtime.Bindable = (*Disclosure)(nil)
var _ runtime.Unbindable = (*Disclosure)(nil)
