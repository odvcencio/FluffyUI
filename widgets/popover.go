package widgets

import (
	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

// PopoverPlacement controls where the popover appears relative to the anchor.
type PopoverPlacement int

const (
	PopoverAuto PopoverPlacement = iota
	PopoverBelow
	PopoverAbove
)

// Popover positions a child widget relative to an anchor rect.
type Popover struct {
	Base

	Child            runtime.Widget
	Anchor           runtime.Rect
	Placement        PopoverPlacement
	Gap              int
	MatchAnchorWidth bool
	DismissOnOutside bool
	DismissOnEscape  bool

	childBounds runtime.Rect
	onClose     func()
	closed      bool
}

// PopoverOption configures a popover.
type PopoverOption func(*Popover)

// NewPopover creates a popover anchored to the given rect.
func NewPopover(anchor runtime.Rect, child runtime.Widget, opts ...PopoverOption) *Popover {
	p := &Popover{
		Child:     child,
		Anchor:    anchor,
		Placement: PopoverAuto,
	}
	p.Base.Role = accessibility.RoleGroup
	for _, opt := range opts {
		if opt != nil {
			opt(p)
		}
	}
	return p
}

// WithPopoverPlacement sets the placement behavior.
func WithPopoverPlacement(placement PopoverPlacement) PopoverOption {
	return func(p *Popover) {
		if p == nil {
			return
		}
		p.Placement = placement
	}
}

// WithPopoverGap sets the gap between anchor and popover.
func WithPopoverGap(gap int) PopoverOption {
	return func(p *Popover) {
		if p == nil {
			return
		}
		p.Gap = gap
	}
}

// WithPopoverMatchAnchorWidth sets whether the popover should match anchor width.
func WithPopoverMatchAnchorWidth(enabled bool) PopoverOption {
	return func(p *Popover) {
		if p == nil {
			return
		}
		p.MatchAnchorWidth = enabled
	}
}

// WithPopoverDismissOnOutside sets whether to dismiss on outside clicks.
func WithPopoverDismissOnOutside(enabled bool) PopoverOption {
	return func(p *Popover) {
		if p == nil {
			return
		}
		p.DismissOnOutside = enabled
	}
}

// WithPopoverDismissOnEscape sets whether to dismiss on Escape.
func WithPopoverDismissOnEscape(enabled bool) PopoverOption {
	return func(p *Popover) {
		if p == nil {
			return
		}
		p.DismissOnEscape = enabled
	}
}

// WithPopoverOnClose registers a callback invoked when the popover closes.
func WithPopoverOnClose(fn func()) PopoverOption {
	return func(p *Popover) {
		if p == nil {
			return
		}
		p.onClose = fn
	}
}

// Measure returns the full available size.
func (p *Popover) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

// Layout positions the child relative to the anchor.
func (p *Popover) Layout(bounds runtime.Rect) {
	if p == nil {
		return
	}
	p.Base.Layout(bounds)
	if p.Child == nil {
		p.childBounds = runtime.Rect{}
		return
	}

	content := p.ContentBounds()
	if content.Width <= 0 || content.Height <= 0 {
		p.childBounds = runtime.Rect{}
		return
	}

	childSize := p.Child.Measure(runtime.Constraints{
		MinWidth:  0,
		MaxWidth:  content.Width,
		MinHeight: 0,
		MaxHeight: content.Height,
	})

	width := childSize.Width
	height := childSize.Height

	if p.MatchAnchorWidth && width < p.Anchor.Width {
		width = p.Anchor.Width
	}
	if width > content.Width {
		width = content.Width
	}
	if height > content.Height {
		height = content.Height
	}
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	x := p.Anchor.X
	if x < content.X {
		x = content.X
	}
	if x+width > content.X+content.Width {
		x = content.X + content.Width - width
	}

	yBelow := p.Anchor.Y + p.Anchor.Height + p.Gap
	yAbove := p.Anchor.Y - height - p.Gap
	y := yBelow

	switch p.Placement {
	case PopoverAbove:
		y = yAbove
	case PopoverBelow:
		y = yBelow
	default:
		bottom := content.Y + content.Height
		if yBelow+height > bottom && yAbove >= content.Y {
			y = yAbove
		} else if yBelow+height > bottom {
			y = bottom - height
		}
	}

	if y < content.Y {
		y = content.Y
	}
	if y+height > content.Y+content.Height {
		y = content.Y + content.Height - height
	}

	p.childBounds = runtime.Rect{X: x, Y: y, Width: width, Height: height}
	p.Child.Layout(p.childBounds)
}

// Render draws the child widget.
func (p *Popover) Render(ctx runtime.RenderContext) {
	if p == nil || p.Child == nil {
		return
	}
	runtime.RenderChild(ctx, p.Child)
}

// HandleMessage forwards input to the child and optionally dismisses.
func (p *Popover) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if p == nil {
		return runtime.Unhandled()
	}
	if p.Child != nil {
		if result := p.Child.HandleMessage(msg); result.Handled {
			return result
		}
	}

	switch m := msg.(type) {
	case runtime.MouseMsg:
		if p.DismissOnOutside && m.Action == runtime.MousePress {
			if !p.childBounds.Contains(m.X, m.Y) {
				p.close()
				return runtime.WithCommand(runtime.PopOverlay{})
			}
		}
	case runtime.KeyMsg:
		if p.DismissOnEscape && m.Key == terminal.KeyEscape {
			p.close()
			return runtime.WithCommand(runtime.PopOverlay{})
		}
	}
	return runtime.Unhandled()
}

// Mount marks the popover as active.
func (p *Popover) Mount() {}

// Unmount closes the popover.
func (p *Popover) Unmount() {
	if p == nil {
		return
	}
	p.close()
}

// ChildWidgets returns the popover content.
func (p *Popover) ChildWidgets() []runtime.Widget {
	if p == nil || p.Child == nil {
		return nil
	}
	return []runtime.Widget{p.Child}
}

func (p *Popover) close() {
	if p == nil || p.closed {
		return
	}
	p.closed = true
	if p.onClose != nil {
		p.onClose()
	}
}

var _ runtime.Widget = (*Popover)(nil)
var _ runtime.ChildProvider = (*Popover)(nil)
var _ runtime.Lifecycle = (*Popover)(nil)
