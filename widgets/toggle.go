package widgets

import (
	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	"github.com/odvcencio/fluffyui/terminal"
)

// Toggle is a focusable on/off switch widget.
// It is visually distinct from a checkbox, rendering as a sliding toggle.
type Toggle struct {
	FocusableBase

	on       *state.Signal[bool]
	label    *state.Signal[string]
	onChange func(bool)
	services runtime.Services

	style    backend.Style
	styleSet bool
}

// ToggleOption configures a Toggle widget.
type ToggleOption = Option[Toggle]

// WithToggleLabel sets the label displayed next to the toggle.
func WithToggleLabel(label *state.Signal[string]) ToggleOption {
	return func(t *Toggle) {
		if t == nil {
			return
		}
		t.label = label
	}
}

// WithToggleOnChange sets the callback invoked when the toggle state changes.
func WithToggleOnChange(fn func(bool)) ToggleOption {
	return func(t *Toggle) {
		if t == nil {
			return
		}
		t.onChange = fn
	}
}

// NewToggle creates a toggle switch bound to the given signal.
func NewToggle(on *state.Signal[bool], opts ...ToggleOption) *Toggle {
	t := &Toggle{
		on:    on,
		style: backend.DefaultStyle(),
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(t)
	}
	t.Base.Role = accessibility.RoleSwitch
	t.syncA11y()
	return t
}

// On returns the current toggle state.
func (t *Toggle) On() bool {
	if t == nil || t.on == nil {
		return false
	}
	return t.on.Get()
}

// SetOn updates the toggle state.
func (t *Toggle) SetOn(value bool) {
	if t == nil || t.on == nil {
		return
	}
	old := t.on.Get()
	t.on.Set(value)
	t.syncA11y()
	if old != value {
		if t.onChange != nil {
			t.onChange(value)
		}
		if announcer := t.services.Announcer(); announcer != nil {
			if value {
				announcer.Announce("on", accessibility.PriorityPolite)
			} else {
				announcer.Announce("off", accessibility.PriorityPolite)
			}
		}
	}
}

// Label returns the current label text.
func (t *Toggle) Label() string {
	if t == nil || t.label == nil {
		return ""
	}
	return t.label.Get()
}

// SetStyle sets the widget style.
func (t *Toggle) SetStyle(style backend.Style) {
	if t == nil {
		return
	}
	t.style = style
	t.styleSet = true
}

// StyleType returns the selector type name for FSS.
func (t *Toggle) StyleType() string {
	return "Toggle"
}

// Bind attaches app services.
func (t *Toggle) Bind(services runtime.Services) {
	if t == nil {
		return
	}
	t.services = services
}

// Unbind releases app services.
func (t *Toggle) Unbind() {
	if t == nil {
		return
	}
	t.services = runtime.Services{}
}

// Measure returns the desired size.
func (t *Toggle) Measure(constraints runtime.Constraints) runtime.Size {
	return t.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		// Toggle track: 6 chars wide (e.g. "[*---]" or "[---*]")
		// plus space + label
		width := 6
		label := t.Label()
		if label != "" {
			width += 1 + textWidth(label)
		}
		if width < 6 {
			width = 6
		}
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: 1})
	})
}

// Render draws the toggle switch.
func (t *Toggle) Render(ctx runtime.RenderContext) {
	if t == nil {
		return
	}
	t.syncA11y()
	bounds := t.ContentBounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	style := resolveBaseStyle(ctx, t, t.style, t.styleSet)
	text := t.renderText()
	writePadded(ctx.Buffer, bounds.X, bounds.Y, bounds.Width, truncateString(text, bounds.Width), style)
}

func (t *Toggle) renderText() string {
	on := t.On()
	label := t.Label()
	var track string
	if on {
		track = "[\u2501\u2501\u2501\u25cf]" // [===*]
	} else {
		track = "[\u25cf\u2501\u2501\u2501]" // [*===]
	}
	suffix := ""
	if on {
		suffix = " On"
	} else {
		suffix = " Off"
	}
	if label != "" {
		suffix = " " + label
	}
	return track + suffix
}

// HandleMessage processes keyboard input for the toggle.
func (t *Toggle) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if t == nil || !t.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	if key.Key == terminal.KeyEnter || (key.Key == terminal.KeyRune && key.Rune == ' ') {
		t.SetOn(!t.On())
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func (t *Toggle) syncA11y() {
	if t == nil {
		return
	}
	if t.Base.Role == "" {
		t.Base.Role = accessibility.RoleSwitch
	}
	label := t.Label()
	if label == "" {
		label = "Toggle"
	}
	t.Base.Label = label
	on := t.On()
	t.Base.State.Checked = &on
	t.Base.RoleDescription = "toggle switch"
}

var _ runtime.Widget = (*Toggle)(nil)
var _ runtime.Focusable = (*Toggle)(nil)
var _ runtime.Bindable = (*Toggle)(nil)
var _ runtime.Unbindable = (*Toggle)(nil)
