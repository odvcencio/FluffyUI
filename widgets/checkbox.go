package widgets

import (
	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/state"
	uistyle "github.com/odvcencio/fluffy-ui/style"
	"github.com/odvcencio/fluffy-ui/terminal"
)

// Checkbox is a toggle input widget.
type Checkbox struct {
	FocusableBase

	label    *state.Signal[string]
	checked  *state.Signal[*bool]
	onChange func(value *bool)

	style      backend.Style
	focusStyle backend.Style
	styleSet   bool
	focusSet   bool
}

// NewCheckbox creates a checkbox with a label.
func NewCheckbox(label string) *Checkbox {
	initial := false
	value := &initial
	cb := &Checkbox{
		label:      state.NewSignal(label),
		checked:    state.NewSignal(value),
		style:      backend.DefaultStyle(),
		focusStyle: backend.DefaultStyle().Reverse(true),
	}
	cb.Base.Role = accessibility.RoleCheckbox
	cb.Base.Label = label
	cb.syncState()
	return cb
}

// SetChecked updates the checkbox value (nil = indeterminate).
func (c *Checkbox) SetChecked(value *bool) {
	if c == nil || c.checked == nil {
		return
	}
	c.checked.Set(value)
	c.syncState()
	if c.onChange != nil {
		c.onChange(value)
	}
}

// Checked returns the current value.
func (c *Checkbox) Checked() *bool {
	if c == nil || c.checked == nil {
		return nil
	}
	return c.checked.Get()
}

// SetOnChange sets the change handler.
func (c *Checkbox) SetOnChange(fn func(value *bool)) {
	if c == nil {
		return
	}
	c.onChange = fn
}

// SetLabel updates the checkbox label.
func (c *Checkbox) SetLabel(label string) {
	if c == nil || c.label == nil {
		return
	}
	c.label.Set(label)
	c.Base.Label = label
}

// SetStyle sets the normal style.
func (c *Checkbox) SetStyle(style backend.Style) {
	if c == nil {
		return
	}
	c.style = style
	c.styleSet = true
}

// SetFocusStyle sets the focused style.
func (c *Checkbox) SetFocusStyle(style backend.Style) {
	if c == nil {
		return
	}
	c.focusStyle = style
	c.focusSet = true
}

// StyleType returns the selector type name.
func (c *Checkbox) StyleType() string {
	return "Checkbox"
}

// Measure returns the size needed.
func (c *Checkbox) Measure(constraints runtime.Constraints) runtime.Size {
	return c.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		label := ""
		if c.label != nil {
			label = c.label.Get()
		}
		width := 4 + textWidth(label)
		if width < 4 {
			width = 4
		}
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: 1})
	})
}

// Render draws the checkbox.
func (c *Checkbox) Render(ctx runtime.RenderContext) {
	if c == nil {
		return
	}
	outer := c.bounds
	content := c.ContentBounds()
	if outer.Width <= 0 || outer.Height <= 0 {
		return
	}
	c.syncState()
	value := c.Checked()
	marker := "[ ]"
	if value == nil {
		marker = "[-]"
	} else if *value {
		marker = "[x]"
	}
	label := ""
	if c.label != nil {
		label = c.label.Get()
	}
	available := max(0, content.Width-4)
	text := marker + " " + truncateString(label, available)
	style := c.style
	resolved := ctx.ResolveStyle(c)
		if !resolved.IsZero() {
			final := resolved
			if c.styleSet {
				final = final.Merge(uistyle.FromBackend(c.style))
			}
			if c.focused && c.focusSet {
				final = final.Merge(uistyle.FromBackend(c.focusStyle))
			}
			style = final.ToBackend()
	} else if c.focused {
		style = c.focusStyle
	}
	ctx.Buffer.Fill(outer, ' ', style)
	if content.Width <= 0 || content.Height <= 0 {
		return
	}
	writePadded(ctx.Buffer, content.X, content.Y, content.Width, text, style)
}

// HandleMessage toggles the checkbox.
func (c *Checkbox) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if c == nil || !c.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	if key.Key == terminal.KeyEnter || (key.Key == terminal.KeyRune && key.Rune == ' ') {
		next := c.toggleValue()
		c.SetChecked(next)
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func (c *Checkbox) toggleValue() *bool {
	current := c.Checked()
	if current == nil || !*current {
		next := true
		return &next
	}
	next := false
	return &next
}

func (c *Checkbox) syncState() {
	if c == nil {
		return
	}
	c.Base.State.Checked = c.Checked()
	if c.label != nil {
		c.Base.Label = c.label.Get()
	}
	if c.Base.Role == "" {
		c.Base.Role = accessibility.RoleCheckbox
	}
}
