package widgets

import (
	"fmt"
	"strings"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

// BreadcrumbItem represents a path segment.
type BreadcrumbItem struct {
	Label   string
	OnClick func()
}

// Breadcrumb renders a path of items with optional collapsing.
//
// When the items exceed the available width, the breadcrumb automatically
// collapses middle items into an ellipsis ("..."), keeping the first item
// and the last N visible items (controlled by SetCollapseKeep, default 2).
// The collapse threshold can be configured with SetCollapseThreshold
// (minimum item count before collapsing, default 4).
type Breadcrumb struct {
	FocusableBase
	Items      []BreadcrumbItem
	selected   int // Currently selected/focused item index
	onNavigate func(index int)
	separator  string
	services   runtime.Services

	// Collapse configuration
	collapsible       bool // enable automatic collapsing
	collapseThreshold int  // minimum items before collapsing kicks in (default 4)
	collapseKeep      int  // number of trailing items to keep visible (default 2)
}

// NewBreadcrumb creates a breadcrumb.
func NewBreadcrumb(items ...BreadcrumbItem) *Breadcrumb {
	crumb := &Breadcrumb{
		Items:             items,
		separator:         " > ",
		collapseThreshold: 4,
		collapseKeep:      2,
	}
	crumb.Base.Role = accessibility.RoleNavigation
	crumb.Base.Landmark = accessibility.LandmarkNavigation
	crumb.Base.Label = "Breadcrumbs"
	return crumb
}

// StyleType returns the selector type name for FSS stylesheet targeting.
func (b *Breadcrumb) StyleType() string {
	return "Breadcrumb"
}

// SetSeparator sets the separator between items (default " > ").
func (b *Breadcrumb) SetSeparator(sep string) {
	if b != nil {
		b.separator = sep
	}
}

// SetCollapsible enables or disables automatic collapsing when items overflow.
func (b *Breadcrumb) SetCollapsible(enabled bool) {
	if b != nil {
		b.collapsible = enabled
	}
}

// Collapsible reports whether automatic collapsing is enabled.
func (b *Breadcrumb) Collapsible() bool {
	if b == nil {
		return false
	}
	return b.collapsible
}

// SetCollapseThreshold sets the minimum number of items before collapsing
// activates. Default is 4.
func (b *Breadcrumb) SetCollapseThreshold(n int) {
	if b != nil && n >= 2 {
		b.collapseThreshold = n
	}
}

// CollapseThreshold returns the minimum item count before collapsing.
func (b *Breadcrumb) CollapseThreshold() int {
	if b == nil {
		return 4
	}
	return b.collapseThreshold
}

// SetCollapseKeep sets the number of trailing items kept visible when
// collapsed. Default is 2.
func (b *Breadcrumb) SetCollapseKeep(n int) {
	if b != nil && n >= 1 {
		b.collapseKeep = n
	}
}

// CollapseKeep returns the number of trailing items kept visible.
func (b *Breadcrumb) CollapseKeep() int {
	if b == nil {
		return 2
	}
	return b.collapseKeep
}

// SetOnNavigate sets the callback for navigation to a breadcrumb item.
func (b *Breadcrumb) SetOnNavigate(fn func(index int)) {
	if b != nil {
		b.onNavigate = fn
	}
}

// OnNavigate sets the callback for navigation to a breadcrumb item.
//
// Deprecated: Use SetOnNavigate instead. This method will be removed in v1.0.
func (b *Breadcrumb) OnNavigate(fn func(index int)) {
	b.SetOnNavigate(fn)
}

// Selected returns the currently selected item index.
func (b *Breadcrumb) Selected() int {
	if b == nil {
		return 0
	}
	return b.selected
}

// fullWidth computes the total width needed to render all items.
func (b *Breadcrumb) fullWidth() int {
	sep := b.sep()
	sepW := textWidth(sep)
	w := 0
	for i, item := range b.Items {
		w += textWidth(item.Label)
		if i < len(b.Items)-1 {
			w += sepW
		}
	}
	return w
}

// sep returns the effective separator string.
func (b *Breadcrumb) sep() string {
	if b.separator == "" {
		return " > "
	}
	return b.separator
}

// needsCollapse returns true if the items should be collapsed to fit.
func (b *Breadcrumb) needsCollapse(availableWidth int) bool {
	if !b.collapsible {
		return false
	}
	if len(b.Items) < b.collapseThreshold {
		return false
	}
	return b.fullWidth() > availableWidth
}

// collapsedIndices returns the display indices when collapsed.
// Format: [first, ellipsis_placeholder, last_N...]
// The returned slice contains the original indices of visible items.
// Index -1 is used as a placeholder for the ellipsis.
func (b *Breadcrumb) collapsedIndices() []int {
	n := len(b.Items)
	keep := b.collapseKeep
	if keep >= n {
		keep = n - 1
	}
	if keep < 1 {
		keep = 1
	}
	// first item + ellipsis + last keep items
	indices := make([]int, 0, 2+keep)
	indices = append(indices, 0)  // first item
	indices = append(indices, -1) // ellipsis placeholder
	for i := n - keep; i < n; i++ {
		indices = append(indices, i)
	}
	return indices
}

// Measure returns desired size.
func (b *Breadcrumb) Measure(constraints runtime.Constraints) runtime.Size {
	return b.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		width := b.fullWidth()
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

	sep := b.sep()
	sepWidth := textWidth(sep)

	normalStyle := backend.DefaultStyle()
	selectedStyle := normalStyle.Reverse(true)
	sepStyle := normalStyle.Dim(true)
	// Last item gets underline to indicate "current page" visually
	currentStyle := normalStyle.Bold(true)
	currentSelectedStyle := currentStyle.Reverse(true)

	collapsed := b.needsCollapse(bounds.Width)
	var displayIndices []int
	if collapsed {
		displayIndices = b.collapsedIndices()
	} else {
		displayIndices = make([]int, len(b.Items))
		for i := range b.Items {
			displayIndices[i] = i
		}
	}

	x := bounds.X
	lastItemIdx := len(b.Items) - 1

	for di, origIdx := range displayIndices {
		if di > 0 {
			// Draw separator
			if x+sepWidth <= bounds.X+bounds.Width {
				ctx.Buffer.SetString(x, bounds.Y, sep, sepStyle)
				x += sepWidth
			}
		}

		if origIdx == -1 {
			// Draw ellipsis
			ellipsis := "\u2026"
			ew := textWidth(ellipsis)
			if x+ew <= bounds.X+bounds.Width {
				ctx.Buffer.SetString(x, bounds.Y, ellipsis, normalStyle.Dim(true))
				x += ew
			}
			continue
		}

		item := b.Items[origIdx]
		isLast := origIdx == lastItemIdx
		isFocused := b.focused && origIdx == b.selected

		// Pick the appropriate style
		style := normalStyle
		if isLast && isFocused {
			style = currentSelectedStyle
		} else if isLast {
			style = currentStyle
		} else if isFocused {
			style = selectedStyle
		}

		label := item.Label
		available := bounds.X + bounds.Width - x
		if textWidth(label) > available {
			label = clipString(label, available)
		}
		labelWidth := textWidth(label)
		if labelWidth > 0 {
			ctx.Buffer.SetString(x, bounds.Y, label, style)
			x += labelWidth
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
	if announcer := b.services.Announcer(); announcer != nil {
		announcer.Announce(fmt.Sprintf("navigated to %s", item.Label), accessibility.PriorityPolite)
	}
}

// itemAtPosition returns the item index at the given screen position.
// Returns -1 if no item is at that position.
func (b *Breadcrumb) itemAtPosition(x, y int) int {
	bounds := b.ContentBounds()
	if y < bounds.Y || y >= bounds.Y+bounds.Height {
		return -1
	}

	sep := b.sep()
	sepWidth := textWidth(sep)

	collapsed := b.needsCollapse(bounds.Width)
	var displayIndices []int
	if collapsed {
		displayIndices = b.collapsedIndices()
	} else {
		displayIndices = make([]int, len(b.Items))
		for i := range b.Items {
			displayIndices[i] = i
		}
	}

	currentX := bounds.X
	for di, origIdx := range displayIndices {
		if di > 0 {
			currentX += sepWidth
		}

		if origIdx == -1 {
			// Ellipsis
			currentX += textWidth("\u2026")
			continue
		}

		item := b.Items[origIdx]
		itemEnd := currentX + textWidth(item.Label)
		if x >= currentX && x < itemEnd {
			return origIdx
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
		b.Base.Role = accessibility.RoleNavigation
	}
	if b.Base.Landmark == "" {
		b.Base.Landmark = accessibility.LandmarkNavigation
	}
	b.Base.Label = "Breadcrumbs"
	path := b.pathString()
	if path != "" {
		b.Base.Value = &accessibility.ValueInfo{Text: path}
		b.Base.Description = ""
	} else {
		b.Base.Value = nil
	}
	// Mark the last item as aria-current="page"
	if len(b.Items) > 0 {
		b.Base.Current = "page"
	} else {
		b.Base.Current = ""
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
	sep := b.sep()
	return strings.Join(parts, sep)
}

// Bind attaches app services.
func (b *Breadcrumb) Bind(services runtime.Services) {
	if b == nil {
		return
	}
	b.services = services
}

// Unbind releases app services.
func (b *Breadcrumb) Unbind() {
	if b == nil {
		return
	}
	b.services = runtime.Services{}
}

var _ runtime.Widget = (*Breadcrumb)(nil)
var _ runtime.Focusable = (*Breadcrumb)(nil)
var _ runtime.Bindable = (*Breadcrumb)(nil)
var _ runtime.Unbindable = (*Breadcrumb)(nil)
