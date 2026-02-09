package widgets

import (
	"fmt"
	"sort"
	"strings"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

// StatusBarAlignment controls where an item is placed within the bar.
type StatusBarAlignment int

const (
	// StatusBarLeft packs items to the left edge.
	StatusBarLeft StatusBarAlignment = iota
	// StatusBarCenter places items in the center.
	StatusBarCenter
	// StatusBarRight packs items to the right edge.
	StatusBarRight
)

// StatusBarItem represents a single section in a status bar.
type StatusBarItem struct {
	Key       string              // unique identifier
	Text      string              // display text
	Alignment StatusBarAlignment  // placement within the bar
	Style     backend.Style       // optional custom style
	OnClick   func()              // optional click handler
	Tooltip   string              // tooltip description
	Priority  int                 // higher priority = kept when space is limited
	Visible   bool                // whether this item is shown
}

// StatusBar is a VS Code-style bottom status bar with left, center, and right sections.
type StatusBar struct {
	FocusableBase
	items    []StatusBarItem
	bgStyle  backend.Style
	services runtime.Services
	selected int // index into visible items for keyboard navigation
}

// NewStatusBar creates a status bar with the given items.
func NewStatusBar(items ...StatusBarItem) *StatusBar {
	sb := &StatusBar{
		items:   make([]StatusBarItem, 0, len(items)),
		bgStyle: backend.DefaultStyle(),
	}
	for _, item := range items {
		item.Visible = true
		sb.items = append(sb.items, item)
	}
	sb.Base.Role = accessibility.RoleToolbar
	sb.Base.Label = "Status Bar"
	sb.Base.Orientation = "horizontal"
	return sb
}

// AddItem appends an item to the status bar.
func (sb *StatusBar) AddItem(item StatusBarItem) {
	if sb == nil {
		return
	}
	if !item.Visible {
		item.Visible = true
	}
	sb.items = append(sb.items, item)
}

// UpdateItem updates the text of an item identified by key.
func (sb *StatusBar) UpdateItem(key string, text string) {
	if sb == nil {
		return
	}
	for i := range sb.items {
		if sb.items[i].Key == key {
			sb.items[i].Text = text
			return
		}
	}
}

// RemoveItem removes the item with the given key.
func (sb *StatusBar) RemoveItem(key string) {
	if sb == nil {
		return
	}
	for i := range sb.items {
		if sb.items[i].Key == key {
			sb.items = append(sb.items[:i], sb.items[i+1:]...)
			// Clamp selected index.
			visible := sb.visibleItems()
			if sb.selected >= len(visible) && len(visible) > 0 {
				sb.selected = len(visible) - 1
			}
			if len(visible) == 0 {
				sb.selected = 0
			}
			return
		}
	}
}

// SetItemVisible sets the visibility of an item identified by key.
func (sb *StatusBar) SetItemVisible(key string, visible bool) {
	if sb == nil {
		return
	}
	for i := range sb.items {
		if sb.items[i].Key == key {
			sb.items[i].Visible = visible
			// Clamp selected index.
			vis := sb.visibleItems()
			if sb.selected >= len(vis) && len(vis) > 0 {
				sb.selected = len(vis) - 1
			}
			if len(vis) == 0 {
				sb.selected = 0
			}
			return
		}
	}
}

// SetBackground sets the bar background style.
func (sb *StatusBar) SetBackground(style backend.Style) {
	if sb == nil {
		return
	}
	sb.bgStyle = style
}

// Items returns a copy of all items.
func (sb *StatusBar) Items() []StatusBarItem {
	if sb == nil {
		return nil
	}
	out := make([]StatusBarItem, len(sb.items))
	copy(out, sb.items)
	return out
}

// Selected returns the currently selected visible item index.
func (sb *StatusBar) Selected() int {
	if sb == nil {
		return 0
	}
	return sb.selected
}

// StyleType returns the selector type name.
func (sb *StatusBar) StyleType() string {
	return "StatusBar"
}

// visibleItems returns items where Visible is true.
func (sb *StatusBar) visibleItems() []StatusBarItem {
	if sb == nil {
		return nil
	}
	var vis []StatusBarItem
	for _, item := range sb.items {
		if item.Visible {
			vis = append(vis, item)
		}
	}
	return vis
}

// fittingItems returns the visible items that fit within the given width,
// dropping lower-priority items first when space is limited.
func (sb *StatusBar) fittingItems(width int) []StatusBarItem {
	visible := sb.visibleItems()
	if len(visible) == 0 {
		return nil
	}

	// Check if everything fits.
	if sb.totalWidth(visible) <= width {
		return visible
	}

	// Sort by priority ascending so we can drop lowest first.
	sorted := make([]StatusBarItem, len(visible))
	copy(sorted, visible)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Priority < sorted[j].Priority
	})

	// Progressively drop the lowest-priority item until we fit.
	for len(sorted) > 0 && sb.totalWidth(sorted) > width {
		sorted = sorted[1:] // drop lowest priority
	}

	// Restore original order among the survivors.
	kept := make(map[string]bool, len(sorted))
	for _, item := range sorted {
		kept[item.Key] = true
	}
	var result []StatusBarItem
	for _, item := range visible {
		if kept[item.Key] {
			result = append(result, item)
		}
	}
	return result
}

// totalWidth computes the total width needed for a set of items with separators.
func (sb *StatusBar) totalWidth(items []StatusBarItem) int {
	if len(items) == 0 {
		return 0
	}
	width := 0
	first := true
	for _, item := range items {
		if !first {
			width += 3 // " | "
		}
		width += textWidth(item.Text)
		first = false
	}
	return width
}

// Measure returns the desired size (full width, one line tall).
func (sb *StatusBar) Measure(constraints runtime.Constraints) runtime.Size {
	return sb.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		visible := sb.visibleItems()
		width := sb.totalWidth(visible)
		if width < 1 {
			width = 1
		}
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: 1})
	})
}

// Layout stores the assigned bounds.
func (sb *StatusBar) Layout(bounds runtime.Rect) {
	if sb == nil {
		return
	}
	sb.Base.Layout(bounds)
}

// Render draws the status bar.
func (sb *StatusBar) Render(ctx runtime.RenderContext) {
	if sb == nil {
		return
	}
	sb.syncA11y()
	bounds := sb.ContentBounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}

	bgStyle := sb.bgStyle

	// Fill the entire bar with the background.
	ctx.Buffer.Fill(bounds, ' ', bgStyle)

	fitting := sb.fittingItems(bounds.Width)
	if len(fitting) == 0 {
		return
	}

	// Partition into left, center, right groups.
	var leftItems, centerItems, rightItems []StatusBarItem
	for _, item := range fitting {
		switch item.Alignment {
		case StatusBarCenter:
			centerItems = append(centerItems, item)
		case StatusBarRight:
			rightItems = append(rightItems, item)
		default:
			leftItems = append(leftItems, item)
		}
	}

	// Build the visible item list in render order for selection mapping.
	visibleForSelect := sb.visibleItems()

	sep := " | "
	sepWidth := textWidth(sep)
	sepStyle := bgStyle.Dim(true)

	// Render left items.
	x := bounds.X
	for i, item := range leftItems {
		if i > 0 {
			if x+sepWidth <= bounds.X+bounds.Width {
				ctx.Buffer.SetString(x, bounds.Y, sep, sepStyle)
				x += sepWidth
			}
		}
		style := sb.itemStyle(item, visibleForSelect, bgStyle)
		label := item.Text
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

	// Render right items (right-aligned).
	rightTotal := sb.groupWidth(rightItems, sepWidth)
	rx := bounds.X + bounds.Width - rightTotal
	if rx < x {
		rx = x
	}
	for i, item := range rightItems {
		if i > 0 {
			if rx+sepWidth <= bounds.X+bounds.Width {
				ctx.Buffer.SetString(rx, bounds.Y, sep, sepStyle)
				rx += sepWidth
			}
		}
		style := sb.itemStyle(item, visibleForSelect, bgStyle)
		label := item.Text
		available := bounds.X + bounds.Width - rx
		if textWidth(label) > available {
			label = clipString(label, available)
		}
		labelWidth := textWidth(label)
		if labelWidth > 0 {
			ctx.Buffer.SetString(rx, bounds.Y, label, style)
			rx += labelWidth
		}
		if rx >= bounds.X+bounds.Width {
			break
		}
	}

	// Render center items.
	if len(centerItems) > 0 {
		centerTotal := sb.groupWidth(centerItems, sepWidth)
		// Place center text between left and right sections.
		leftEnd := x
		rightStart := bounds.X + bounds.Width - rightTotal
		if rightStart < leftEnd {
			rightStart = leftEnd
		}
		gap := rightStart - leftEnd
		cx := leftEnd + (gap-centerTotal)/2
		if cx < leftEnd {
			cx = leftEnd
		}
		for i, item := range centerItems {
			if i > 0 {
				if cx+sepWidth <= bounds.X+bounds.Width {
					ctx.Buffer.SetString(cx, bounds.Y, sep, sepStyle)
					cx += sepWidth
				}
			}
			style := sb.itemStyle(item, visibleForSelect, bgStyle)
			label := item.Text
			available := bounds.X + bounds.Width - cx
			if textWidth(label) > available {
				label = clipString(label, available)
			}
			labelWidth := textWidth(label)
			if labelWidth > 0 {
				ctx.Buffer.SetString(cx, bounds.Y, label, style)
				cx += labelWidth
			}
			if cx >= bounds.X+bounds.Width {
				break
			}
		}
	}
}

// groupWidth computes the total width of a group of items with separators.
func (sb *StatusBar) groupWidth(items []StatusBarItem, sepWidth int) int {
	width := 0
	for i, item := range items {
		if i > 0 {
			width += sepWidth
		}
		width += textWidth(item.Text)
	}
	return width
}

// itemStyle returns the style for a given item, highlighting if selected and focused.
func (sb *StatusBar) itemStyle(item StatusBarItem, visibleItems []StatusBarItem, fallback backend.Style) backend.Style {
	visIdx := -1
	for i, v := range visibleItems {
		if v.Key == item.Key {
			visIdx = i
			break
		}
	}

	base := fallback
	if item.Style != (backend.Style{}) {
		base = item.Style
	}

	if sb.focused && visIdx == sb.selected {
		return base.Reverse(true)
	}
	return base
}

// HandleMessage handles keyboard navigation and mouse clicks.
func (sb *StatusBar) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if sb == nil {
		return runtime.Unhandled()
	}

	switch m := msg.(type) {
	case runtime.MouseMsg:
		if m.Action == runtime.MousePress && m.Button == runtime.MouseLeft {
			index := sb.itemAtPosition(m.X, m.Y)
			if index >= 0 {
				sb.selected = index
				sb.activateItem(index)
				return runtime.Handled()
			}
		}

	case runtime.KeyMsg:
		if !sb.focused {
			return runtime.Unhandled()
		}

		visible := sb.visibleItems()
		if len(visible) == 0 {
			return runtime.Unhandled()
		}

		switch m.Key {
		case terminal.KeyLeft:
			if sb.selected > 0 {
				sb.selected--
				sb.Invalidate()
				return runtime.Handled()
			}
		case terminal.KeyRight, terminal.KeyTab:
			if sb.selected < len(visible)-1 {
				sb.selected++
				sb.Invalidate()
				return runtime.Handled()
			}
		case terminal.KeyHome:
			if sb.selected != 0 {
				sb.selected = 0
				sb.Invalidate()
				return runtime.Handled()
			}
		case terminal.KeyEnd:
			last := len(visible) - 1
			if sb.selected != last {
				sb.selected = last
				sb.Invalidate()
				return runtime.Handled()
			}
		case terminal.KeyEnter:
			sb.activateItem(sb.selected)
			return runtime.Handled()
		}
	}

	return runtime.Unhandled()
}

// activateItem invokes the OnClick handler of the visible item at the given index.
func (sb *StatusBar) activateItem(index int) {
	visible := sb.visibleItems()
	if index < 0 || index >= len(visible) {
		return
	}
	item := visible[index]
	if item.OnClick != nil {
		item.OnClick()
	}
	if announcer := sb.services.Announcer(); announcer != nil {
		announcer.Announce(fmt.Sprintf("activated %s", item.Text), accessibility.PriorityPolite)
	}
}

// itemAtPosition returns the visible item index at the given screen position.
// Returns -1 if no item is at that position.
func (sb *StatusBar) itemAtPosition(x, y int) int {
	bounds := sb.ContentBounds()
	if y < bounds.Y || y >= bounds.Y+bounds.Height {
		return -1
	}

	visible := sb.visibleItems()
	sep := " | "
	sepWidth := textWidth(sep)

	currentX := bounds.X
	for i, item := range visible {
		if i > 0 {
			currentX += sepWidth
		}
		itemEnd := currentX + textWidth(item.Text)
		if x >= currentX && x < itemEnd {
			return i
		}
		currentX = itemEnd
	}
	return -1
}

func (sb *StatusBar) syncA11y() {
	if sb == nil {
		return
	}
	if sb.Base.Role == "" {
		sb.Base.Role = accessibility.RoleToolbar
	}
	sb.Base.Label = "Status Bar"
	sb.Base.Orientation = "horizontal"

	visible := sb.visibleItems()
	if len(visible) > 0 {
		var parts []string
		for _, item := range visible {
			parts = append(parts, item.Text)
		}
		sb.Base.Description = strings.Join(parts, " | ")
		sb.Base.SetSize = len(visible)
		if sb.selected >= 0 && sb.selected < len(visible) {
			sb.Base.Value = &accessibility.ValueInfo{Text: visible[sb.selected].Text}
			sb.Base.PosInSet = sb.selected + 1 // 1-based
		}
	} else {
		sb.Base.Description = ""
		sb.Base.Value = nil
		sb.Base.SetSize = 0
		sb.Base.PosInSet = 0
	}
}

// Bind attaches app services.
func (sb *StatusBar) Bind(services runtime.Services) {
	if sb == nil {
		return
	}
	sb.services = services
}

// Unbind releases app services.
func (sb *StatusBar) Unbind() {
	if sb == nil {
		return
	}
	sb.services = runtime.Services{}
}

var _ runtime.Widget = (*StatusBar)(nil)
var _ runtime.Focusable = (*StatusBar)(nil)
var _ runtime.Bindable = (*StatusBar)(nil)
var _ runtime.Unbindable = (*StatusBar)(nil)
