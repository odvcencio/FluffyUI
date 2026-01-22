package widgets

import (
	"strings"

	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/style"
)

// Panel is a container widget with optional border and background.
type Panel struct {
	Base
	child       runtime.Widget
	style       backend.Style
	borderStyle backend.Style
	hasBorder   bool
	title       string
	label       string
	styleSet       bool
	borderStyleSet bool
}

// NewPanel creates a new panel widget.
func NewPanel(child runtime.Widget) *Panel {
	panel := &Panel{
		child:       child,
		style:       backend.DefaultStyle(),
		borderStyle: backend.DefaultStyle(),
		hasBorder:   false,
		label:       "Panel",
	}
	panel.Base.Role = accessibility.RoleGroup
	panel.syncA11y()
	return panel
}

// SetStyle sets the panel background style.
func (p *Panel) SetStyle(style backend.Style) {
	p.style = style
	p.styleSet = true
}

// WithStyle sets the style and returns for chaining.
func (p *Panel) WithStyle(style backend.Style) *Panel {
	p.style = style
	p.styleSet = true
	return p
}

// SetBorder enables or disables the border.
func (p *Panel) SetBorder(enabled bool) {
	p.hasBorder = enabled
}

// WithBorder enables border and returns for chaining.
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
	p.title = title
	p.syncA11y()
}

// WithTitle sets title and returns for chaining.
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
			if len(title) > bounds.Width-4 {
				title = title[:bounds.Width-4]
			}
			x := bounds.X + 2
			ctx.Buffer.SetString(x, bounds.Y, title, drawStyle)
		}
	}

	// Render child
	if p.child != nil {
		p.child.Render(ctx)
	}
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
	child runtime.Widget
	style backend.Style
	label string
	styleSet bool
}

// NewBox creates a new box widget.
func NewBox(child runtime.Widget) *Box {
	box := &Box{
		child: child,
		style: backend.DefaultStyle(),
		label: "Box",
	}
	box.Base.Role = accessibility.RoleGroup
	box.syncA11y()
	return box
}

// SetStyle sets the background style.
func (b *Box) SetStyle(style backend.Style) {
	b.style = style
	b.styleSet = true
}

// WithStyle sets style and returns for chaining.
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
	if b.child != nil {
		b.child.Render(ctx)
	}
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
