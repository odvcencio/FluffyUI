package widgets

import (
	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/state"
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

// Measure returns the size needed.
func (c *Checkbox) Measure(constraints runtime.Constraints) runtime.Size {
	label := ""
	if c.label != nil {
		label = c.label.Get()
	}
	width := 4 + len(label)
	if width < 4 {
		width = 4
	}
	return constraints.Constrain(runtime.Size{Width: width, Height: 1})
}

// Render draws the checkbox.
func (c *Checkbox) Render(ctx runtime.RenderContext) {
	if c == nil {
		return
	}
	bounds := c.bounds
	if bounds.Width <= 0 || bounds.Height <= 0 {
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
	text := marker + " " + truncateString(label, bounds.Width-4)
	style := c.style
	if c.focused {
		style = c.focusStyle
	}
	writePadded(ctx.Buffer, bounds.X, bounds.Y, bounds.Width, text, style)
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
