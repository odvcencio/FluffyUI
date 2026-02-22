package widgets

import (
	"fmt"
	"html/template"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

// ChipVariant describes the visual style of a chip.
type ChipVariant int

const (
	// ChipDefault is the default chip variant.
	ChipDefault ChipVariant = iota
	// ChipPrimary is the primary chip variant.
	ChipPrimary
	// ChipSuccess is the success chip variant.
	ChipSuccess
	// ChipWarning is the warning chip variant.
	ChipWarning
	// ChipError is the error chip variant.
	ChipError
)

// Chip is a compact label widget, optionally dismissible.
// When dismissible, it becomes focusable and handles Enter/Space to dismiss.
type Chip struct {
	FocusableBase

	label     string
	variant   ChipVariant
	onDismiss func()
	services  runtime.Services

	style    backend.Style
	styleSet bool
}

// ChipOption configures a Chip widget.
type ChipOption = Option[Chip]

// WithChipVariant sets the chip variant for visual styling.
func WithChipVariant(v ChipVariant) ChipOption {
	return func(c *Chip) {
		if c == nil {
			return
		}
		c.variant = v
	}
}

// WithChipDismiss makes the chip dismissible with the given callback.
func WithChipDismiss(fn func()) ChipOption {
	return func(c *Chip) {
		if c == nil {
			return
		}
		c.onDismiss = fn
	}
}

// NewChip creates a chip widget with the given label.
func NewChip(label string, opts ...ChipOption) *Chip {
	c := &Chip{
		label:   label,
		variant: ChipDefault,
		style:   backend.DefaultStyle(),
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(c)
	}
	c.Base.Role = accessibility.RoleStatus
	c.syncA11y()
	return c
}

// Label returns the chip label text.
func (c *Chip) Label() string {
	if c == nil {
		return ""
	}
	return c.label
}

// SetLabel updates the chip label.
func (c *Chip) SetLabel(label string) {
	if c == nil {
		return
	}
	c.label = label
	c.syncA11y()
}

// Variant returns the chip variant.
func (c *Chip) Variant() ChipVariant {
	if c == nil {
		return ChipDefault
	}
	return c.variant
}

// Dismissible returns true if the chip can be dismissed.
func (c *Chip) Dismissible() bool {
	return c != nil && c.onDismiss != nil
}

// CanFocus returns true only if the chip is dismissible.
func (c *Chip) CanFocus() bool {
	if c == nil {
		return false
	}
	return c.onDismiss != nil
}

// SetStyle sets the widget style.
func (c *Chip) SetStyle(style backend.Style) {
	if c == nil {
		return
	}
	c.style = style
	c.styleSet = true
}

// StyleType returns the selector type name for FSS.
func (c *Chip) StyleType() string {
	return "Chip"
}

// Bind attaches app services.
func (c *Chip) Bind(services runtime.Services) {
	if c == nil {
		return
	}
	c.services = services
}

// Unbind releases app services.
func (c *Chip) Unbind() {
	if c == nil {
		return
	}
	c.services = runtime.Services{}
}

// Measure returns the desired size.
func (c *Chip) Measure(constraints runtime.Constraints) runtime.Size {
	return c.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		text := c.renderText()
		width := textWidth(text)
		if width < 1 {
			width = 1
		}
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: 1})
	})
}

// Render draws the chip.
func (c *Chip) Render(ctx runtime.RenderContext) {
	if c == nil {
		return
	}
	c.syncA11y()
	bounds := c.ContentBounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	style := resolveBaseStyle(ctx, c, c.style, c.styleSet)
	text := c.renderText()
	writePadded(ctx.Buffer, bounds.X, bounds.Y, bounds.Width, truncateString(text, bounds.Width), style)
}

// RenderHTML renders the chip as a static HTML button.
func (c *Chip) RenderHTML(ctx runtime.HTMLContext) runtime.HTML {
	return runtime.HTML(fmt.Sprintf(`<button class="fluffy-Chip">%s</button>`,
		template.HTMLEscapeString(c.label)))
}

func (c *Chip) renderText() string {
	if c.onDismiss != nil {
		return fmt.Sprintf("[%s \u00d7]", c.label)
	}
	return fmt.Sprintf("[%s]", c.label)
}

// HandleMessage processes keyboard input for dismissible chips.
func (c *Chip) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if c == nil || !c.focused || c.onDismiss == nil {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	if key.Key == terminal.KeyEnter || (key.Key == terminal.KeyRune && key.Rune == ' ') {
		if announcer := c.services.Announcer(); announcer != nil {
			announcer.Announce(fmt.Sprintf("%s dismissed", c.label), accessibility.PriorityPolite)
		}
		c.onDismiss()
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func (c *Chip) syncA11y() {
	if c == nil {
		return
	}
	if c.Base.Role == "" {
		c.Base.Role = accessibility.RoleStatus
	}
	c.Base.Label = c.label
	if c.onDismiss != nil {
		c.Base.Description = "dismissible, press Enter to dismiss"
	} else {
		c.Base.Description = ""
	}
}

var _ runtime.Widget = (*Chip)(nil)
var _ runtime.Focusable = (*Chip)(nil)
var _ runtime.Bindable = (*Chip)(nil)
var _ runtime.Unbindable = (*Chip)(nil)
