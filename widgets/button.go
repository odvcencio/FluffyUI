package widgets

import (
	"strings"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	uistyle "github.com/odvcencio/fluffyui/style"
	"github.com/odvcencio/fluffyui/terminal"
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

	label    *state.Signal[string]
	variant  Variant
	disabled *state.Signal[bool]
	loading  *state.Signal[bool]
	onClick  func()
	services runtime.Services
	subs     state.Subscriptions

	style       backend.Style
	focusStyle  backend.Style
	disabledSty backend.Style

	styleSet         bool
	focusStyleSet    bool
	disabledStyleSet bool
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
	btn.syncA11y()
	return btn
}

// Bind attaches app services.
func (b *Button) Bind(services runtime.Services) {
	if b == nil {
		return
	}
	b.services = services
	b.subs.SetScheduler(services.Scheduler())
	b.subs.Observe(b.label, func() {
		b.services.Invalidate()
	})
	b.subs.Observe(b.loading, func() {
		b.services.Invalidate()
	})
	b.subs.Observe(b.disabled, func() {
		b.services.Relayout()
	})
}

// Unbind releases app services.
func (b *Button) Unbind() {
	if b == nil {
		return
	}
	b.subs.Clear()
	b.services = runtime.Services{}
}

// WithVariant sets the button variant.
func WithVariant(v Variant) ButtonOption {
	return func(b *Button) {
		if b != nil {
			b.variant = v
		}
	}
}

// WithClass adds a style class.
func WithClass(class string) ButtonOption {
	return func(b *Button) {
		if b != nil {
			b.AddClass(class)
		}
	}
}

// WithClasses adds multiple style classes.
func WithClasses(classes ...string) ButtonOption {
	return func(b *Button) {
		if b != nil {
			b.AddClasses(classes...)
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

// SetVariant updates the button variant.
func (b *Button) SetVariant(variant Variant) {
	if b == nil {
		return
	}
	b.variant = variant
}

// Primary applies the primary variant and returns the button for chaining.
func (b *Button) Primary() *Button {
	if b != nil {
		b.variant = VariantPrimary
	}
	return b
}

// Secondary applies the secondary variant and returns the button for chaining.
func (b *Button) Secondary() *Button {
	if b != nil {
		b.variant = VariantSecondary
	}
	return b
}

// Danger applies the danger variant and returns the button for chaining.
func (b *Button) Danger() *Button {
	if b != nil {
		b.variant = VariantDanger
	}
	return b
}

// Disabled sets the disabled signal and returns the button for chaining.
func (b *Button) Disabled(disabled *state.Signal[bool]) *Button {
	if b != nil && disabled != nil {
		b.disabled = disabled
	}
	return b
}

// Loading sets the loading signal and returns the button for chaining.
func (b *Button) Loading(loading *state.Signal[bool]) *Button {
	if b != nil && loading != nil {
		b.loading = loading
	}
	return b
}

// OnClick sets the click handler and returns the button for chaining.
func (b *Button) OnClick(fn func()) *Button {
	if b != nil {
		b.onClick = fn
	}
	return b
}

// Class adds a style class and returns the button for chaining.
func (b *Button) Class(class string) *Button {
	if b != nil {
		b.AddClass(class)
	}
	return b
}

// Classes adds style classes and returns the button for chaining.
func (b *Button) Classes(classes ...string) *Button {
	if b != nil {
		b.AddClasses(classes...)
	}
	return b
}

// SetLabel updates the button label.
func (b *Button) SetLabel(label string) {
	if b == nil || b.label == nil {
		return
	}
	b.label.Set(label)
	b.Base.Label = label
	b.syncA11y()
}

// SetStyle updates the button style.
func (b *Button) SetStyle(style backend.Style) {
	if b == nil {
		return
	}
	b.style = style
	b.styleSet = true
}

// SetFocusStyle updates the focus style.
func (b *Button) SetFocusStyle(style backend.Style) {
	if b == nil {
		return
	}
	b.focusStyle = style
	b.focusStyleSet = true
}

// SetDisabledStyle updates the disabled style.
func (b *Button) SetDisabledStyle(style backend.Style) {
	if b == nil {
		return
	}
	b.disabledSty = style
	b.disabledStyleSet = true
}

// StyleType returns the selector type name.
func (b *Button) StyleType() string {
	return "Button"
}

// StyleClasses returns selector classes including the variant.
func (b *Button) StyleClasses() []string {
	classes := b.Base.StyleClasses()
	if b == nil || b.variant == "" {
		return classes
	}
	variant := string(b.variant)
	for _, cls := range classes {
		if cls == variant {
			return classes
		}
	}
	return append(classes, variant)
}

// Measure returns the size needed by the button.
func (b *Button) Measure(constraints runtime.Constraints) runtime.Size {
	return b.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		label := ""
		if b.label != nil {
			label = b.label.Get()
		}
		width := textWidth(label) + 4
		if width < 4 {
			width = 4
		}
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: 1})
	})
}

// Render draws the button.
func (b *Button) Render(ctx runtime.RenderContext) {
	if b == nil {
		return
	}
	outer := b.bounds
	content := b.ContentBounds()
	if outer.Width <= 0 || outer.Height <= 0 {
		return
	}
	label := ""
	if b.label != nil {
		label = b.label.Get()
	}
	loading := b.loading != nil && b.loading.Get()
	disabled := b.disabled != nil && b.disabled.Get()
	b.syncA11yWith(label, disabled, loading)
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
	resolved := ctx.ResolveStyle(b)
	if !resolved.IsZero() {
		final := resolved
		if b.styleSet {
			final = final.Merge(uistyle.FromBackend(b.style))
		}
		if b.focused && b.focusStyleSet {
			final = final.Merge(uistyle.FromBackend(b.focusStyle))
		}
		if disabled && b.disabledStyleSet {
			final = final.Merge(uistyle.FromBackend(b.disabledSty))
		}
		style = final.ToBackend()
	} else {
		if b.focused {
			style = b.focusStyle
		}
		if disabled {
			style = b.disabledSty
		}
	}

	available := max(0, content.Width-2)
	ctx.Buffer.Fill(outer, ' ', style)
	if content.Width <= 0 || content.Height <= 0 {
		return
	}
	text := "[" + truncateString(label, available) + "]"
	writePadded(ctx.Buffer, content.X, content.Y, content.Width, text, style)
}

func (b *Button) syncA11y() {
	label := ""
	if b != nil && b.label != nil {
		label = b.label.Get()
	}
	disabled := b != nil && b.disabled != nil && b.disabled.Get()
	loading := b != nil && b.loading != nil && b.loading.Get()
	b.syncA11yWith(label, disabled, loading)
}

func (b *Button) syncA11yWith(label string, disabled bool, loading bool) {
	if b == nil {
		return
	}
	if b.Base.Role == "" {
		b.Base.Role = accessibility.RoleButton
	}
	b.Base.Label = label
	b.Base.State.Disabled = disabled
	if loading {
		b.Base.Description = "loading"
	} else if b.Base.Description == "loading" {
		b.Base.Description = ""
	}
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
