package widgets

import (
	"strings"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/style"
)

// Panel is a container widget with optional border and background.
type Panel struct {
	Base
	child          runtime.Widget
	style          backend.Style
	borderStyle    backend.Style
	hasBorder      bool
	title          string
	label          string
	styleSet       bool
	borderStyleSet bool
}

// PanelOption configures a Panel widget.
type PanelOption = Option[Panel]

// WithPanelStyle sets the panel background style.
func WithPanelStyle(style backend.Style) PanelOption {
	return func(p *Panel) {
		if p == nil {
			return
		}
		p.SetStyle(style)
	}
}

// WithPanelBorder enables a border with the given style.
func WithPanelBorder(style backend.Style) PanelOption {
	return func(p *Panel) {
		if p == nil {
			return
		}
		p.hasBorder = true
		p.borderStyle = style
		p.borderStyleSet = true
	}
}

// WithPanelTitle sets the panel title.
func WithPanelTitle(title string) PanelOption {
	return func(p *Panel) {
		if p == nil {
			return
		}
		p.SetTitle(title)
	}
}

// NewPanel creates a new panel widget.
func NewPanel(child runtime.Widget, opts ...PanelOption) *Panel {
	panel := &Panel{
		child:       child,
		style:       backend.DefaultStyle(),
		borderStyle: backend.DefaultStyle(),
		hasBorder:   false,
		label:       "Panel",
	}
	panel.Base.Role = accessibility.RoleGroup
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(panel)
	}
	panel.syncA11y()
	return panel
}

// SetStyle sets the panel background style.
func (p *Panel) SetStyle(style backend.Style) {
	if p == nil {
		return
	}
	p.style = style
	p.styleSet = true
}

// Deprecated: prefer WithPanelStyle during construction or SetStyle for mutation.
func (p *Panel) WithStyle(style backend.Style) *Panel {
	p.style = style
	p.styleSet = true
	return p
}

// SetBorder enables or disables the border.
func (p *Panel) SetBorder(enabled bool) {
	if p == nil {
		return
	}
	p.hasBorder = enabled
}

// SetBorderStyle sets the border style and enables the border.
func (p *Panel) SetBorderStyle(style backend.Style) {
	if p == nil {
		return
	}
	p.hasBorder = true
	p.borderStyle = style
	p.borderStyleSet = true
}

// Deprecated: prefer WithPanelBorder during construction or SetBorderStyle for mutation.
func (p *Panel) WithBorder(style backend.Style) *Panel {
	p.hasBorder = true
	p.borderStyle = style
	p.borderStyleSet = true
	return p
}

// StyleType returns the selector type name.
func (p *Panel) StyleType() string {
	return "Panel"
}

// SetTitle sets the panel title (shown in border).
func (p *Panel) SetTitle(title string) {
	if p == nil {
		return
	}
	p.title = title
	p.syncA11y()
}

// Deprecated: prefer WithPanelTitle during construction or SetTitle for mutation.
func (p *Panel) WithTitle(title string) *Panel {
	p.title = title
	p.syncA11y()
	return p
}

// Measure returns the size needed for the panel.
func (p *Panel) Measure(constraints runtime.Constraints) runtime.Size {
	return p.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		extraBorder := 0
		if p.hasBorder && p.layoutMetrics.border == 0 {
			extraBorder = 1
		}

		childConstraints := shrinkConstraints(contentConstraints, extraBorder, extraBorder, extraBorder, extraBorder)
		if p.child == nil {
			size := runtime.Size{Width: extraBorder * 2, Height: extraBorder * 2}
			return contentConstraints.Constrain(size)
		}

		childSize := p.child.Measure(childConstraints)
		size := runtime.Size{
			Width:  childSize.Width + extraBorder*2,
			Height: childSize.Height + extraBorder*2,
		}
		return contentConstraints.Constrain(size)
	})
}

// Layout positions the panel and its child.
func (p *Panel) Layout(bounds runtime.Rect) {
	p.Base.Layout(bounds)

	if p.child == nil {
		return
	}

	childBounds := p.ContentBounds()
	if p.hasBorder && p.layoutMetrics.border == 0 {
		childBounds = childBounds.Inset(1, 1, 1, 1)
	}
	p.child.Layout(childBounds)
}

// Render draws the panel.
func (p *Panel) Render(ctx runtime.RenderContext) {
	bounds := p.bounds
	if bounds.Width == 0 || bounds.Height == 0 {
		return
	}
	p.syncA11y()

	resolved := ctx.ResolveStyle(p)
	background := p.style
	borderStyle := p.borderStyle
	hasBorder := p.hasBorder
	borderSpec := (*style.Border)(nil)

	if !resolved.IsZero() {
		final := resolved
		if p.styleSet {
			final = final.Merge(style.FromBackend(p.style))
		}
		background = final.ToBackend()
		borderSpec = final.Border
		if p.borderStyleSet {
			borderStyle = p.borderStyle
		}
	}

	// Fill background
	ctx.Buffer.Fill(bounds, ' ', background)

	// Draw border if enabled
	if borderSpec != nil || hasBorder {
		drawStyle := borderStyle
		drawSpec := borderSpec
		if drawSpec == nil {
			drawSpec = &style.Border{Style: style.BorderRounded}
		}
		if drawSpec.Color.Mode != style.ColorNone.Mode {
			drawStyle = style.Style{Foreground: drawSpec.Color}.ToBackend()
		}
		drawn := false
		switch drawSpec.Style {
		case style.BorderDouble:
			ctx.Buffer.DrawDoubleBox(bounds, drawStyle)
			drawn = true
		case style.BorderSingle:
			ctx.Buffer.DrawBox(bounds, drawStyle)
			drawn = true
		case style.BorderRounded:
			ctx.Buffer.DrawRoundedBox(bounds, drawStyle)
			drawn = true
		}

		// Draw title in top border
		if drawn && p.title != "" {
			title := " " + p.title + " "
			if textWidth(title) > bounds.Width-4 {
				title = clipString(title, bounds.Width-4)
			}
			x := bounds.X + 2
			ctx.Buffer.SetString(x, bounds.Y, title, drawStyle)
		}
	}

	// Render child
	runtime.RenderChild(ctx, p.child)
}

// HandleMessage delegates to child.
func (p *Panel) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if p.child != nil {
		return p.child.HandleMessage(msg)
	}
	return runtime.Unhandled()
}

// ChildWidgets returns the panel's child widget.
func (p *Panel) ChildWidgets() []runtime.Widget {
	if p.child == nil {
		return nil
	}
	return []runtime.Widget{p.child}
}

func (p *Panel) syncA11y() {
	if p == nil {
		return
	}
	if p.Base.Role == "" {
		p.Base.Role = accessibility.RoleGroup
	}
	label := strings.TrimSpace(p.title)
	if label == "" {
		label = strings.TrimSpace(p.label)
	}
	if label == "" {
		label = "Panel"
	}
	p.Base.Label = label
}

// Box is a simple container that fills its background.
type Box struct {
	Base
	child    runtime.Widget
	style    backend.Style
	label    string
	styleSet bool
}

// BoxOption configures a Box widget.
type BoxOption = Option[Box]

// WithBoxStyle sets the box background style.
func WithBoxStyle(style backend.Style) BoxOption {
	return func(b *Box) {
		if b == nil {
			return
		}
		b.SetStyle(style)
	}
}

// NewBox creates a new box widget.
func NewBox(child runtime.Widget, opts ...BoxOption) *Box {
	box := &Box{
		child: child,
		style: backend.DefaultStyle(),
		label: "Box",
	}
	box.Base.Role = accessibility.RoleGroup
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(box)
	}
	box.syncA11y()
	return box
}

// SetStyle sets the background style.
func (b *Box) SetStyle(style backend.Style) {
	if b == nil {
		return
	}
	b.style = style
	b.styleSet = true
}

// Deprecated: prefer WithBoxStyle during construction or SetStyle for mutation.
func (b *Box) WithStyle(style backend.Style) *Box {
	b.style = style
	b.styleSet = true
	return b
}

// StyleType returns the selector type name.
func (b *Box) StyleType() string {
	return "Box"
}

// Measure returns the child's size.
func (b *Box) Measure(constraints runtime.Constraints) runtime.Size {
	return b.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		if b.child == nil {
			return contentConstraints.MinSize()
		}
		return b.child.Measure(contentConstraints)
	})
}

// Layout assigns bounds to the box and child.
func (b *Box) Layout(bounds runtime.Rect) {
	b.Base.Layout(bounds)
	if b.child != nil {
		b.child.Layout(b.ContentBounds())
	}
}

// Render draws the background and child.
func (b *Box) Render(ctx runtime.RenderContext) {
	// Fill background
	background := b.style
	resolved := ctx.ResolveStyle(b)
	if !resolved.IsZero() {
		final := resolved
		if b.styleSet {
			final = final.Merge(style.FromBackend(b.style))
		}
		background = final.ToBackend()
	}
	ctx.Buffer.Fill(b.bounds, ' ', background)

	// Render child
	runtime.RenderChild(ctx, b.child)
}

// HandleMessage delegates to child.
func (b *Box) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if b.child != nil {
		return b.child.HandleMessage(msg)
	}
	return runtime.Unhandled()
}

// ChildWidgets returns the box's child widget.
func (b *Box) ChildWidgets() []runtime.Widget {
	if b.child == nil {
		return nil
	}
	return []runtime.Widget{b.child}
}

func (b *Box) syncA11y() {
	if b == nil {
		return
	}
	if b.Base.Role == "" {
		b.Base.Role = accessibility.RoleGroup
	}
	label := strings.TrimSpace(b.label)
	if label == "" {
		label = "Box"
	}
	b.Base.Label = label
}

var _ runtime.Widget = (*Panel)(nil)
var _ runtime.Widget = (*Box)(nil)
