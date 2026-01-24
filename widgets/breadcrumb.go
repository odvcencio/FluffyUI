package widgets

import (
	"strings"

	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/terminal"
)

// BreadcrumbItem represents a path segment.
type BreadcrumbItem struct {
	Label   string
	OnClick func()
}

// Breadcrumb renders a path of items.
type Breadcrumb struct {
	FocusableBase
	Items      []BreadcrumbItem
	selected   int // Currently selected/focused item index
	onNavigate func(index int)
	separator  string
}

// NewBreadcrumb creates a breadcrumb.
func NewBreadcrumb(items ...BreadcrumbItem) *Breadcrumb {
	crumb := &Breadcrumb{
		Items:     items,
		separator: " > ",
	}
	crumb.Base.Role = accessibility.RoleList
	crumb.Base.Label = "Breadcrumbs"
	return crumb
}

// SetSeparator sets the separator between items (default " > ").
func (b *Breadcrumb) SetSeparator(sep string) {
	if b != nil {
		b.separator = sep
	}
}

// OnNavigate sets the callback for navigation to a breadcrumb item.
func (b *Breadcrumb) OnNavigate(fn func(index int)) {
	if b != nil {
		b.onNavigate = fn
	}
}

// Selected returns the currently selected item index.
func (b *Breadcrumb) Selected() int {
	if b == nil {
		return 0
	}
	return b.selected
}

// Measure returns desired size.
func (b *Breadcrumb) Measure(constraints runtime.Constraints) runtime.Size {
	return b.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		width := 0
		sep := b.separator
		if sep == "" {
			sep = " > "
		}
		for i, item := range b.Items {
			width += len(item.Label)
			if i < len(b.Items)-1 {
				width += len(sep)
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

	sep := b.separator
	if sep == "" {
		sep = " > "
	}

	x := bounds.X
	normalStyle := backend.DefaultStyle()
	selectedStyle := normalStyle.Reverse(true)
	sepStyle := normalStyle.Dim(true)

	for i, item := range b.Items {
		if i > 0 {
			// Draw separator
			if x+len(sep) <= bounds.X+bounds.Width {
				ctx.Buffer.SetString(x, bounds.Y, sep, sepStyle)
				x += len(sep)
			}
		}

		// Draw item
		style := normalStyle
		if b.focused && i == b.selected {
			style = selectedStyle
		}
		label := item.Label
		available := bounds.X + bounds.Width - x
		if len(label) > available {
			label = label[:available]
		}
		if len(label) > 0 {
			ctx.Buffer.SetString(x, bounds.Y, label, style)
			x += len(label)
		}

		if x >= bounds.X+bounds.Width {
			break
		}
	}

	// Fill remaining space
	for ; x < bounds.X+bounds.Width; x++ {
		ctx.Buffer.Set(x, bounds.Y, ' ', normalStyle)
	}
}

// HandleMessage handles mouse clicks and keyboard navigation.
func (b *Breadcrumb) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if b == nil || len(b.Items) == 0 {
		return runtime.Unhandled()
	}

	switch m := msg.(type) {
	case runtime.MouseMsg:
		if m.Action == runtime.MousePress && m.Button == runtime.MouseLeft {
			index := b.itemAtPosition(m.X, m.Y)
			if index >= 0 && index < len(b.Items) {
				b.selected = index
				b.activateItem(index)
				return runtime.Handled()
			}
		}

	case runtime.KeyMsg:
		if !b.focused {
			return runtime.Unhandled()
		}

		switch m.Key {
		case terminal.KeyLeft:
			if b.selected > 0 {
				b.selected--
				b.Invalidate()
				return runtime.Handled()
			}
		case terminal.KeyRight:
			if b.selected < len(b.Items)-1 {
				b.selected++
				b.Invalidate()
				return runtime.Handled()
			}
		case terminal.KeyHome:
			if b.selected != 0 {
				b.selected = 0
				b.Invalidate()
				return runtime.Handled()
			}
		case terminal.KeyEnd:
			if b.selected != len(b.Items)-1 {
				b.selected = len(b.Items) - 1
				b.Invalidate()
				return runtime.Handled()
			}
		case terminal.KeyEnter:
			b.activateItem(b.selected)
			return runtime.Handled()
		}
	}

	return runtime.Unhandled()
}

// activateItem calls the OnClick handler or onNavigate for the given index.
func (b *Breadcrumb) activateItem(index int) {
	if index < 0 || index >= len(b.Items) {
		return
	}
	item := b.Items[index]
	if item.OnClick != nil {
		item.OnClick()
	} else if b.onNavigate != nil {
		b.onNavigate(index)
	}
}

// itemAtPosition returns the item index at the given screen position.
// Returns -1 if no item is at that position.
func (b *Breadcrumb) itemAtPosition(x, y int) int {
	bounds := b.ContentBounds()
	if y < bounds.Y || y >= bounds.Y+bounds.Height {
		return -1
	}

	sep := b.separator
	if sep == "" {
		sep = " > "
	}

	currentX := bounds.X
	for i, item := range b.Items {
		if i > 0 {
			currentX += len(sep)
		}
		itemEnd := currentX + len(item.Label)
		if x >= currentX && x < itemEnd {
			return i
		}
		currentX = itemEnd
	}
	return -1
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
	sep := b.separator
	if sep == "" {
		sep = " > "
	}
	return strings.Join(parts, sep)
}
