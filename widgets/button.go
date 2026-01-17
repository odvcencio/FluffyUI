package widgets

import (
	"strings"

	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/state"
	"github.com/odvcencio/fluffy-ui/terminal"
)

// Variant controls button styling.
type Variant string

const (
	VariantPrimary   Variant = "primary"
	VariantSecondary Variant = "secondary"
	VariantDanger    Variant = "danger"
)

// Button is a clickable action widget.
type Button struct {
	FocusableBase
	accessibility.Base

	label    *state.Signal[string]
	variant  Variant
	disabled *state.Signal[bool]
	loading  *state.Signal[bool]
	onClick  func()

	style       backend.Style
	focusStyle  backend.Style
	disabledSty backend.Style
}

// ButtonOption configures a button.
type ButtonOption func(*Button)

// NewButton creates a new button.
func NewButton(label string, opts ...ButtonOption) *Button {
	btn := &Button{
		label:       state.NewSignal(label),
		variant:     VariantPrimary,
		disabled:    state.NewSignal(false),
		loading:     state.NewSignal(false),
		style:       backend.DefaultStyle(),
		focusStyle:  backend.DefaultStyle().Reverse(true),
		disabledSty: backend.DefaultStyle().Dim(true),
	}
	btn.Base.Role = accessibility.RoleButton
	btn.Base.Label = label
	for _, opt := range opts {
		opt(btn)
	}
	return btn
}

// WithVariant sets the button variant.
func WithVariant(v Variant) ButtonOption {
	return func(b *Button) {
		if b != nil {
			b.variant = v
		}
	}
}

// WithDisabled sets the disabled signal.
func WithDisabled(disabled *state.Signal[bool]) ButtonOption {
	return func(b *Button) {
		if b != nil && disabled != nil {
			b.disabled = disabled
		}
	}
}

// WithLoading sets the loading signal.
func WithLoading(loading *state.Signal[bool]) ButtonOption {
	return func(b *Button) {
		if b != nil && loading != nil {
			b.loading = loading
		}
	}
}

// WithOnClick sets the click handler.
func WithOnClick(fn func()) ButtonOption {
	return func(b *Button) {
		if b != nil {
			b.onClick = fn
		}
	}
}

// SetLabel updates the button label.
func (b *Button) SetLabel(label string) {
	if b == nil || b.label == nil {
		return
	}
	b.label.Set(label)
	b.Base.Label = label
}

// SetStyle updates the button style.
func (b *Button) SetStyle(style backend.Style) {
	if b == nil {
		return
	}
	b.style = style
}

// SetFocusStyle updates the focus style.
func (b *Button) SetFocusStyle(style backend.Style) {
	if b == nil {
		return
	}
	b.focusStyle = style
}

// Measure returns the size needed by the button.
func (b *Button) Measure(constraints runtime.Constraints) runtime.Size {
	label := ""
	if b.label != nil {
		label = b.label.Get()
	}
	width := len(label) + 4
	if width < 4 {
		width = 4
	}
	return constraints.Constrain(runtime.Size{Width: width, Height: 1})
}

// Render draws the button.
func (b *Button) Render(ctx runtime.RenderContext) {
	if b == nil {
		return
	}
	bounds := b.bounds
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	label := ""
	if b.label != nil {
		label = b.label.Get()
	}
	loading := b.loading != nil && b.loading.Get()
	disabled := b.disabled != nil && b.disabled.Get()
	if loading {
		label = strings.TrimSpace(label) + "..."
	}

	style := b.style
	switch b.variant {
	case VariantPrimary:
		style = style.Bold(true)
	case VariantDanger:
		style = style.Bold(true).Underline(true)
	}
	if b.focused {
		style = b.focusStyle
	}
	if disabled {
		style = b.disabledSty
	}

	text := "[" + truncateString(label, bounds.Width-2) + "]"
	writePadded(ctx.Buffer, bounds.X, bounds.Y, bounds.Width, text, style)
}

// HandleMessage handles button activation.
func (b *Button) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if b == nil || !b.focused {
		return runtime.Unhandled()
	}
	if b.disabled != nil && b.disabled.Get() {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	if key.Key == terminal.KeyEnter || (key.Key == terminal.KeyRune && key.Rune == ' ') {
		if b.onClick != nil {
			b.onClick()
		}
		return runtime.Handled()
	}
	return runtime.Unhandled()
}
