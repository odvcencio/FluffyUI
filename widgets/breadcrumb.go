package widgets

import (
	"strings"

	"github.com/odvcencio/fluffy-ui/accessibility"
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
	crumb := &Breadcrumb{Items: items}
	crumb.Base.Role = accessibility.RoleList
	crumb.Base.Label = "Breadcrumbs"
	return crumb
}

// Measure returns desired size.
func (b *Breadcrumb) Measure(constraints runtime.Constraints) runtime.Size {
	return b.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
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
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: 1})
	})
}

// Render draws breadcrumb text.
func (b *Breadcrumb) Render(ctx runtime.RenderContext) {
	if b == nil {
		return
	}
	b.syncA11y()
	bounds := b.ContentBounds()
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

func (b *Breadcrumb) syncA11y() {
	if b == nil {
		return
	}
	if b.Base.Role == "" {
		b.Base.Role = accessibility.RoleList
	}
	b.Base.Label = "Breadcrumbs"
	path := b.pathString()
	if path != "" {
		b.Base.Value = &accessibility.ValueInfo{Text: path}
		b.Base.Description = ""
	} else {
		b.Base.Value = nil
	}
}

func (b *Breadcrumb) pathString() string {
	if b == nil || len(b.Items) == 0 {
		return ""
	}
	parts := make([]string, 0, len(b.Items))
	for _, item := range b.Items {
		if strings.TrimSpace(item.Label) == "" {
			continue
		}
		parts = append(parts, item.Label)
	}
	return strings.Join(parts, " > ")
}
