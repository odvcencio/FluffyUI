package widgets

import (
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
)

// BreadcrumbItem represents a path segment.
type BreadcrumbItem struct {
	Label   string
	OnClick func()
}

// Breadcrumb renders a path of items.
type Breadcrumb struct {
	Base
	Items []BreadcrumbItem
}

// NewBreadcrumb creates a breadcrumb.
func NewBreadcrumb(items ...BreadcrumbItem) *Breadcrumb {
	return &Breadcrumb{Items: items}
}

// Measure returns desired size.
func (b *Breadcrumb) Measure(constraints runtime.Constraints) runtime.Size {
	width := 0
	for i, item := range b.Items {
		width += len(item.Label)
		if i < len(b.Items)-1 {
			width += 3
		}
	}
	if width < 1 {
		width = 1
	}
	return constraints.Constrain(runtime.Size{Width: width, Height: 1})
}

// Render draws breadcrumb text.
func (b *Breadcrumb) Render(ctx runtime.RenderContext) {
	if b == nil {
		return
	}
	bounds := b.bounds
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	text := ""
	for i, item := range b.Items {
		if i > 0 {
			text += " > "
		}
		text += item.Label
	}
	text = truncateString(text, bounds.Width)
	writePadded(ctx.Buffer, bounds.X, bounds.Y, bounds.Width, text, backend.DefaultStyle())
}

// HandleMessage returns unhandled (click not supported).
func (b *Breadcrumb) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}
