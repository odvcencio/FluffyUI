package widgets

import (
	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
)

// Card is a non-focusable bordered container with optional header, body, and footer sections.
type Card struct {
	Base

	header runtime.Widget
	body   runtime.Widget
	footer runtime.Widget
	title  string
	border bool

	style    backend.Style
	styleSet bool
	services runtime.Services
}

// NewCard creates a card with a title and body widget.
func NewCard(title string, body runtime.Widget) *Card {
	c := &Card{
		title:  title,
		body:   body,
		border: true,
		style:  backend.DefaultStyle(),
	}
	c.Base.Role = accessibility.RoleGroup
	c.Base.Label = title
	return c
}

// SetTitle updates the card title.
func (c *Card) SetTitle(title string) {
	if c == nil {
		return
	}
	c.title = title
	c.Base.Label = title
}

// Title returns the card title.
func (c *Card) Title() string {
	if c == nil {
		return ""
	}
	return c.title
}

// SetHeader sets the header widget.
func (c *Card) SetHeader(header runtime.Widget) {
	if c == nil {
		return
	}
	c.header = header
}

// SetBody sets the body widget.
func (c *Card) SetBody(body runtime.Widget) {
	if c == nil {
		return
	}
	c.body = body
}

// SetFooter sets the footer widget.
func (c *Card) SetFooter(footer runtime.Widget) {
	if c == nil {
		return
	}
	c.footer = footer
}

// SetBorder enables or disables the border.
func (c *Card) SetBorder(border bool) {
	if c == nil {
		return
	}
	c.border = border
}

// SetStyle sets the card style.
func (c *Card) SetStyle(style backend.Style) {
	if c == nil {
		return
	}
	c.style = style
	c.styleSet = true
}

// StyleType returns the selector type name.
func (c *Card) StyleType() string {
	return "Card"
}

// ChildWidgets returns child widgets for tree traversal.
func (c *Card) ChildWidgets() []runtime.Widget {
	if c == nil {
		return nil
	}
	var children []runtime.Widget
	if c.header != nil {
		children = append(children, c.header)
	}
	if c.body != nil {
		children = append(children, c.body)
	}
	if c.footer != nil {
		children = append(children, c.footer)
	}
	return children
}

// Measure returns the desired size.
func (c *Card) Measure(constraints runtime.Constraints) runtime.Size {
	return c.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		borderInset := 0
		if c.border {
			borderInset = 2
		}
		innerMaxW := contentConstraints.MaxWidth - borderInset
		if innerMaxW < 0 {
			innerMaxW = 0
		}
		innerMaxH := contentConstraints.MaxHeight - borderInset
		if innerMaxH < 0 {
			innerMaxH = 0
		}
		innerConstraints := runtime.Constraints{
			MaxWidth:  innerMaxW,
			MaxHeight: innerMaxH,
		}

		totalHeight := 0

		// Title row.
		if c.title != "" {
			totalHeight++
		}

		// Header.
		if c.header != nil {
			sz := c.header.Measure(innerConstraints)
			totalHeight += sz.Height
		}

		// Body.
		if c.body != nil {
			bodyConstraints := innerConstraints
			bodyConstraints.MaxHeight = innerConstraints.MaxHeight - totalHeight
			if bodyConstraints.MaxHeight < 0 {
				bodyConstraints.MaxHeight = 0
			}
			sz := c.body.Measure(bodyConstraints)
			totalHeight += sz.Height
		}

		// Footer.
		if c.footer != nil {
			footerConstraints := innerConstraints
			footerConstraints.MaxHeight = innerConstraints.MaxHeight - totalHeight
			if footerConstraints.MaxHeight < 0 {
				footerConstraints.MaxHeight = 0
			}
			sz := c.footer.Measure(footerConstraints)
			totalHeight += sz.Height
		}

		width := contentConstraints.MaxWidth
		height := totalHeight + borderInset
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: height})
	})
}

// Layout positions child widgets within the card.
func (c *Card) Layout(bounds runtime.Rect) {
	if c == nil {
		return
	}
	c.Base.Layout(bounds)
	content := c.ContentBounds()

	borderInset := 0
	if c.border {
		borderInset = 1
	}
	innerX := content.X + borderInset
	innerY := content.Y + borderInset
	innerW := content.Width - borderInset*2
	innerH := content.Height - borderInset*2
	if innerW < 0 {
		innerW = 0
	}
	if innerH < 0 {
		innerH = 0
	}

	y := innerY

	// Title row.
	if c.title != "" {
		y++
	}

	// Header.
	if c.header != nil {
		sz := c.header.Measure(runtime.Constraints{MaxWidth: innerW, MaxHeight: innerH})
		c.header.Layout(runtime.Rect{X: innerX, Y: y, Width: innerW, Height: sz.Height})
		y += sz.Height
	}

	// Body.
	if c.body != nil {
		remaining := innerH - (y - innerY)
		if remaining < 0 {
			remaining = 0
		}
		// Reserve space for footer.
		footerH := 0
		if c.footer != nil {
			fsz := c.footer.Measure(runtime.Constraints{MaxWidth: innerW, MaxHeight: remaining})
			footerH = fsz.Height
		}
		bodyH := remaining - footerH
		if bodyH < 0 {
			bodyH = 0
		}
		c.body.Layout(runtime.Rect{X: innerX, Y: y, Width: innerW, Height: bodyH})
		y += bodyH
	}

	// Footer.
	if c.footer != nil {
		remaining := innerH - (y - innerY)
		if remaining < 0 {
			remaining = 0
		}
		c.footer.Layout(runtime.Rect{X: innerX, Y: y, Width: innerW, Height: remaining})
	}
}

// Render draws the card.
func (c *Card) Render(ctx runtime.RenderContext) {
	if c == nil {
		return
	}
	c.syncA11y()
	content := c.ContentBounds()
	if content.Width <= 0 || content.Height <= 0 {
		return
	}
	style := resolveBaseStyle(ctx, c, c.style, c.styleSet)

	// Fill background.
	ctx.Buffer.Fill(content, ' ', style)

	if c.border {
		c.renderBorder(ctx, content, style)
	}

	// Render title inside the top border if present.
	if c.title != "" {
		titleY := content.Y
		if c.border {
			titleY = content.Y
		}
		borderInset := 0
		if c.border {
			borderInset = 1
		}
		titleX := content.X + borderInset + 1
		titleW := content.Width - borderInset*2 - 2
		if titleW > 0 && titleY < content.Y+content.Height {
			ctx.Buffer.SetString(titleX, titleY, truncateString(c.title, titleW), style.Bold(true))
		}
	}

	// Render children.
	runtime.RenderChild(ctx, c.header)
	runtime.RenderChild(ctx, c.body)
	runtime.RenderChild(ctx, c.footer)
}

func (c *Card) renderBorder(ctx runtime.RenderContext, bounds runtime.Rect, style backend.Style) {
	x, y := bounds.X, bounds.Y
	w, h := bounds.Width, bounds.Height
	if w < 2 || h < 2 {
		return
	}
	// Corners.
	ctx.Buffer.Set(x, y, '+', style)
	ctx.Buffer.Set(x+w-1, y, '+', style)
	ctx.Buffer.Set(x, y+h-1, '+', style)
	ctx.Buffer.Set(x+w-1, y+h-1, '+', style)
	// Top and bottom.
	for i := x + 1; i < x+w-1; i++ {
		ctx.Buffer.Set(i, y, '-', style)
		ctx.Buffer.Set(i, y+h-1, '-', style)
	}
	// Left and right.
	for j := y + 1; j < y+h-1; j++ {
		ctx.Buffer.Set(x, j, '|', style)
		ctx.Buffer.Set(x+w-1, j, '|', style)
	}
}

// HandleMessage forwards messages to children.
func (c *Card) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if c == nil {
		return runtime.Unhandled()
	}
	for _, child := range c.ChildWidgets() {
		if result := child.HandleMessage(msg); result.Handled {
			return result
		}
	}
	return runtime.Unhandled()
}

func (c *Card) syncA11y() {
	if c == nil {
		return
	}
	if c.Base.Role == "" {
		c.Base.Role = accessibility.RoleGroup
	}
	c.Base.Label = c.title
}

// Bind attaches app services.
func (c *Card) Bind(services runtime.Services) {
	if c == nil {
		return
	}
	c.services = services
}

// Unbind releases app services.
func (c *Card) Unbind() {
	if c == nil {
		return
	}
	c.services = runtime.Services{}
}

var _ runtime.Widget = (*Card)(nil)
var _ runtime.ChildProvider = (*Card)(nil)
var _ runtime.Bindable = (*Card)(nil)
var _ runtime.Unbindable = (*Card)(nil)
