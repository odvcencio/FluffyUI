package widgets

import (
	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	"github.com/odvcencio/fluffyui/terminal"
)

// DrawerSide specifies which edge the drawer slides out from.
type DrawerSide int

const (
	// DrawerLeft opens from the left edge.
	DrawerLeft DrawerSide = iota
	// DrawerRight opens from the right edge.
	DrawerRight
	// DrawerBottom opens from the bottom edge.
	DrawerBottom
)

// Drawer is a slide-out panel that overlays from a screen edge.
// When closed, it takes no space. When open, it renders its child content
// at the specified edge with the specified width or height.
type Drawer struct {
	Base

	child    runtime.Widget
	side     DrawerSide
	open     *state.Signal[bool]
	width    int // for left/right drawers
	height   int // for bottom drawer
	services runtime.Services

	style    backend.Style
	styleSet bool
}

// DrawerOption configures a Drawer widget.
type DrawerOption = Option[Drawer]

// WithDrawerSide sets which edge the drawer opens from.
func WithDrawerSide(side DrawerSide) DrawerOption {
	return func(d *Drawer) {
		if d == nil {
			return
		}
		d.side = side
	}
}

// WithDrawerWidth sets the width for left/right drawers.
func WithDrawerWidth(width int) DrawerOption {
	return func(d *Drawer) {
		if d == nil {
			return
		}
		if width > 0 {
			d.width = width
		}
	}
}

// WithDrawerHeight sets the height for bottom drawers.
func WithDrawerHeight(height int) DrawerOption {
	return func(d *Drawer) {
		if d == nil {
			return
		}
		if height > 0 {
			d.height = height
		}
	}
}

// NewDrawer creates a drawer containing the given child widget.
// The open signal controls whether the drawer is visible.
func NewDrawer(child runtime.Widget, open *state.Signal[bool], opts ...DrawerOption) *Drawer {
	d := &Drawer{
		child:  child,
		open:   open,
		side:   DrawerLeft,
		width:  30,
		height: 10,
		style:  backend.DefaultStyle(),
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(d)
	}
	d.Base.Role = accessibility.RoleComplementary
	d.syncA11y()
	return d
}

// IsOpen returns whether the drawer is currently open.
func (d *Drawer) IsOpen() bool {
	if d == nil || d.open == nil {
		return false
	}
	return d.open.Get()
}

// SetOpen updates the open state and announces the change.
func (d *Drawer) SetOpen(value bool) {
	if d == nil || d.open == nil {
		return
	}
	old := d.open.Get()
	d.open.Set(value)
	d.syncA11y()
	if old != value {
		if announcer := d.services.Announcer(); announcer != nil {
			if value {
				announcer.Announce("drawer opened", accessibility.PriorityPolite)
			} else {
				announcer.Announce("drawer closed", accessibility.PriorityPolite)
			}
		}
	}
}

// SetStyle sets the drawer background style.
func (d *Drawer) SetStyle(style backend.Style) {
	if d == nil {
		return
	}
	d.style = style
	d.styleSet = true
}

// StyleType returns the selector type name for FSS.
func (d *Drawer) StyleType() string {
	return "Drawer"
}

// Bind attaches app services.
func (d *Drawer) Bind(services runtime.Services) {
	if d == nil {
		return
	}
	d.services = services
}

// Unbind releases app services.
func (d *Drawer) Unbind() {
	if d == nil {
		return
	}
	d.services = runtime.Services{}
}

// Measure returns the desired size.
// When closed, returns zero size. When open, returns the drawer dimensions.
func (d *Drawer) Measure(constraints runtime.Constraints) runtime.Size {
	if !d.IsOpen() {
		return runtime.Size{}
	}
	return d.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		switch d.side {
		case DrawerLeft, DrawerRight:
			w := d.width
			if w > contentConstraints.MaxWidth {
				w = contentConstraints.MaxWidth
			}
			h := contentConstraints.MaxHeight
			return contentConstraints.Constrain(runtime.Size{Width: w, Height: h})
		case DrawerBottom:
			h := d.height
			if h > contentConstraints.MaxHeight {
				h = contentConstraints.MaxHeight
			}
			w := contentConstraints.MaxWidth
			return contentConstraints.Constrain(runtime.Size{Width: w, Height: h})
		default:
			return contentConstraints.Constrain(runtime.Size{Width: d.width, Height: d.height})
		}
	})
}

// Layout positions the drawer and its child.
func (d *Drawer) Layout(bounds runtime.Rect) {
	if d == nil {
		return
	}
	d.Base.Layout(bounds)
	if d.child != nil && d.IsOpen() {
		content := d.ContentBounds()
		d.child.Layout(content)
	}
}

// Render draws the drawer and its child content when open.
func (d *Drawer) Render(ctx runtime.RenderContext) {
	if d == nil || !d.IsOpen() {
		return
	}
	d.syncA11y()
	bounds := d.ContentBounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	style := resolveBaseStyle(ctx, d, d.style, d.styleSet)
	ctx.Buffer.Fill(bounds, ' ', style)
	runtime.RenderChild(ctx, d.child)
}

// HandleMessage processes input for the drawer.
// Escape closes the drawer when it is open.
func (d *Drawer) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if d == nil {
		return runtime.Unhandled()
	}
	if !d.IsOpen() {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if ok && key.Key == terminal.KeyEscape {
		d.SetOpen(false)
		return runtime.Handled()
	}
	if d.child != nil {
		return d.child.HandleMessage(msg)
	}
	return runtime.Unhandled()
}

// ChildWidgets returns the child widget when the drawer is open.
func (d *Drawer) ChildWidgets() []runtime.Widget {
	if d == nil || d.child == nil || !d.IsOpen() {
		return nil
	}
	return []runtime.Widget{d.child}
}

func (d *Drawer) syncA11y() {
	if d == nil {
		return
	}
	if d.Base.Role == "" {
		d.Base.Role = accessibility.RoleComplementary
	}
	d.Base.Label = "Drawer"
	isOpen := d.IsOpen()
	d.Base.State.Expanded = &isOpen
	d.Base.State.Hidden = !isOpen
}

var _ runtime.Widget = (*Drawer)(nil)
var _ runtime.ChildProvider = (*Drawer)(nil)
var _ runtime.Bindable = (*Drawer)(nil)
var _ runtime.Unbindable = (*Drawer)(nil)
