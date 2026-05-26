package widgets

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"m31labs.dev/fluffyui/accessibility"
	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/state"
	"m31labs.dev/fluffyui/terminal"
)

// NumberInput is a focusable numeric input with increment/decrement controls.
// It renders as [v] 42 [^] and supports keyboard-driven value changes.
type NumberInput struct {
	FocusableBase

	value    *state.Signal[float64]
	min      float64
	max      float64
	step     float64
	onChange func(float64)
	services runtime.Services

	// editing state: when the user types digits, we accumulate them
	editing    bool
	editBuffer string

	style    backend.Style
	styleSet bool
}

// NumberInputOption configures a NumberInput widget.
type NumberInputOption = Option[NumberInput]

// WithNumberRange sets the minimum and maximum allowed values.
func WithNumberRange(min, max float64) NumberInputOption {
	return func(n *NumberInput) {
		if n == nil {
			return
		}
		n.min = min
		n.max = max
	}
}

// WithNumberStep sets the increment/decrement step size.
func WithNumberStep(step float64) NumberInputOption {
	return func(n *NumberInput) {
		if n == nil {
			return
		}
		if step > 0 {
			n.step = step
		}
	}
}

// WithNumberOnChange sets the value change handler.
func WithNumberOnChange(fn func(float64)) NumberInputOption {
	return func(n *NumberInput) {
		if n == nil {
			return
		}
		n.onChange = fn
	}
}

// NewNumberInput creates a numeric input bound to the given signal.
func NewNumberInput(value *state.Signal[float64], opts ...NumberInputOption) *NumberInput {
	n := &NumberInput{
		value: value,
		min:   0,
		max:   100,
		step:  1,
		style: backend.DefaultStyle(),
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(n)
	}
	n.Base.Role = accessibility.RoleSpinButton
	n.syncA11y()
	return n
}

// Value returns the current numeric value.
func (n *NumberInput) Value() float64 {
	if n == nil || n.value == nil {
		return 0
	}
	return n.value.Get()
}

// SetValue updates the current value, clamping to the min/max range.
func (n *NumberInput) SetValue(v float64) {
	if n == nil || n.value == nil {
		return
	}
	v = math.Max(n.min, math.Min(n.max, v))
	old := n.value.Get()
	n.value.Set(v)
	n.syncA11y()
	if old != v {
		if n.onChange != nil {
			n.onChange(v)
		}
		if announcer := n.services.Announcer(); announcer != nil {
			announcer.Announce(fmt.Sprintf("%.4g", v), accessibility.PriorityPolite)
		}
	}
}

// Min returns the minimum allowed value.
func (n *NumberInput) Min() float64 {
	if n == nil {
		return 0
	}
	return n.min
}

// Max returns the maximum allowed value.
func (n *NumberInput) Max() float64 {
	if n == nil {
		return 0
	}
	return n.max
}

// Step returns the increment/decrement step size.
func (n *NumberInput) Step() float64 {
	if n == nil {
		return 1
	}
	return n.step
}

// SetStyle sets the widget style.
func (n *NumberInput) SetStyle(style backend.Style) {
	if n == nil {
		return
	}
	n.style = style
	n.styleSet = true
}

// StyleType returns the selector type name for FSS.
func (n *NumberInput) StyleType() string {
	return "NumberInput"
}

// Bind attaches app services.
func (n *NumberInput) Bind(services runtime.Services) {
	if n == nil {
		return
	}
	n.services = services
}

// Unbind releases app services.
func (n *NumberInput) Unbind() {
	if n == nil {
		return
	}
	n.services = runtime.Services{}
}

// Measure returns the desired size.
func (n *NumberInput) Measure(constraints runtime.Constraints) runtime.Size {
	return n.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		text := n.renderText()
		width := textWidth(text)
		if width < 5 {
			width = 5
		}
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: 1})
	})
}

// Render draws the number input widget.
func (n *NumberInput) Render(ctx runtime.RenderContext) {
	if n == nil {
		return
	}
	n.syncA11y()
	bounds := n.ContentBounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	style := resolveBaseStyle(ctx, n, n.style, n.styleSet)
	text := n.renderText()
	writePadded(ctx.Buffer, bounds.X, bounds.Y, bounds.Width, truncateString(text, bounds.Width), style)
}

func (n *NumberInput) renderText() string {
	v := n.Value()
	// Format nicely: remove trailing zeros
	text := strconv.FormatFloat(v, 'f', -1, 64)
	return fmt.Sprintf("[v] %s [^]", text)
}

// HandleMessage processes keyboard input for the number input.
func (n *NumberInput) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if n == nil || !n.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	switch key.Key {
	case terminal.KeyUp:
		n.SetValue(n.Value() + n.step)
		return runtime.Handled()
	case terminal.KeyDown:
		n.SetValue(n.Value() - n.step)
		return runtime.Handled()
	case terminal.KeyHome:
		n.SetValue(n.min)
		return runtime.Handled()
	case terminal.KeyEnd:
		n.SetValue(n.max)
		return runtime.Handled()
	case terminal.KeyEnter:
		if n.editing {
			n.commitEdit()
			return runtime.Handled()
		}
		return runtime.Unhandled()
	case terminal.KeyEscape:
		if n.editing {
			n.editing = false
			n.editBuffer = ""
			return runtime.Handled()
		}
		return runtime.Unhandled()
	case terminal.KeyBackspace:
		if n.editing && len(n.editBuffer) > 0 {
			n.editBuffer = n.editBuffer[:len(n.editBuffer)-1]
			if len(n.editBuffer) == 0 {
				n.editing = false
			}
			return runtime.Handled()
		}
		return runtime.Unhandled()
	case terminal.KeyRune:
		r := key.Rune
		if (r >= '0' && r <= '9') || r == '.' || r == '-' {
			if !n.editing {
				n.editing = true
				n.editBuffer = ""
			}
			n.editBuffer += string(r)
			return runtime.Handled()
		}
	}
	return runtime.Unhandled()
}

func (n *NumberInput) commitEdit() {
	if n == nil {
		return
	}
	n.editing = false
	text := strings.TrimSpace(n.editBuffer)
	n.editBuffer = ""
	if text == "" {
		return
	}
	v, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return
	}
	n.SetValue(v)
}

func (n *NumberInput) syncA11y() {
	if n == nil {
		return
	}
	if n.Base.Role == "" {
		n.Base.Role = accessibility.RoleSpinButton
	}
	n.Base.Label = "Number input"
	v := n.Value()
	n.Base.Value = &accessibility.ValueInfo{
		Min:     n.min,
		Max:     n.max,
		Current: v,
		Text:    fmt.Sprintf("%.4g", v),
	}
	n.Base.Orientation = "vertical"
	n.Base.KeyShortcuts = "Up/Down to change value"
}

var _ runtime.Widget = (*NumberInput)(nil)
var _ runtime.Focusable = (*NumberInput)(nil)
var _ runtime.Bindable = (*NumberInput)(nil)
var _ runtime.Unbindable = (*NumberInput)(nil)
