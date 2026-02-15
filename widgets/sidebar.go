package widgets

import (
	"fmt"
	"strings"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

// SidebarItem represents a single entry in a Sidebar navigation.
type SidebarItem struct {
	Label    string
	Icon     string // optional single-char icon prefix
	OnSelect func()
	Badge    string        // optional badge text drawn right-aligned
	Children []SidebarItem // for nested navigation
	Expanded bool
}

// Sidebar renders a vertical navigation panel with keyboard support.
//
// Items can be nested via Children. Nested items are expanded/collapsed with
// Left/Right or Enter keys. Up/Down moves selection. Enter activates the
// OnSelect callback of the focused item.
type Sidebar struct {
	FocusableBase
	items         []SidebarItem
	selected      int
	width         int // desired fixed width (0 = fill available)
	expanded      bool
	label         string
	style         backend.Style
	selectedStyle backend.Style
	indentCache   []string
	flatCache     []sidebarRow
	flatDirty     bool
	offset        int
	services      runtime.Services
}

// NewSidebar creates a new sidebar navigation widget.
func NewSidebar(items ...SidebarItem) *Sidebar {
	s := &Sidebar{
		items:         items,
		selected:      0,
		expanded:      true,
		label:         "Sidebar",
		style:         backend.DefaultStyle(),
		selectedStyle: backend.DefaultStyle().Reverse(true),
		flatDirty:     true,
	}
	s.Base.Role = accessibility.RoleNavigation
	s.Base.Landmark = accessibility.LandmarkNavigation
	s.syncA11y()
	return s
}

// StyleType returns the selector type name for FSS stylesheet targeting.
func (s *Sidebar) StyleType() string {
	return "Sidebar"
}

// SetItems replaces the sidebar items.
func (s *Sidebar) SetItems(items ...SidebarItem) {
	if s == nil {
		return
	}
	s.items = items
	s.flatDirty = true
	s.syncA11y()
}

// Items returns the current sidebar items.
func (s *Sidebar) GetItems() []SidebarItem {
	if s == nil {
		return nil
	}
	return s.items
}

// SetWidth sets the desired fixed width. 0 means fill available width.
func (s *Sidebar) SetWidth(w int) {
	if s == nil {
		return
	}
	if w < 0 {
		w = 0
	}
	s.width = w
}

// Width returns the desired fixed width.
func (s *Sidebar) Width() int {
	if s == nil {
		return 0
	}
	return s.width
}

// SetExpanded controls whether the sidebar is expanded (visible) or collapsed.
func (s *Sidebar) SetExpanded(expanded bool) {
	if s == nil {
		return
	}
	s.expanded = expanded
}

// Expanded reports whether the sidebar is expanded.
func (s *Sidebar) Expanded() bool {
	if s == nil {
		return false
	}
	return s.expanded
}

// SetLabel updates the accessibility label.
func (s *Sidebar) SetLabel(label string) {
	if s == nil {
		return
	}
	s.label = label
	s.syncA11y()
}

// SetStyle updates the sidebar base style.
func (s *Sidebar) SetStyle(style backend.Style) {
	if s == nil {
		return
	}
	s.style = style
}

// SetSelectedStyle updates the selected row style.
func (s *Sidebar) SetSelectedStyle(style backend.Style) {
	if s == nil {
		return
	}
	s.selectedStyle = style
}

// Selected returns the currently selected flat index.
func (s *Sidebar) Selected() int {
	if s == nil {
		return 0
	}
	return s.selected
}

// SelectedItem returns the currently selected SidebarItem, or nil.
func (s *Sidebar) SelectedItem() *SidebarItem {
	if s == nil {
		return nil
	}
	rows := s.flatten()
	if s.selected < 0 || s.selected >= len(rows) {
		return nil
	}
	return rows[s.selected].item
}

// Measure returns desired size.
func (s *Sidebar) Measure(constraints runtime.Constraints) runtime.Size {
	return s.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		if !s.expanded {
			return contentConstraints.Constrain(runtime.Size{Width: 0, Height: 0})
		}
		rows := s.flatten()
		height := len(rows)
		if height > contentConstraints.MaxHeight {
			height = contentConstraints.MaxHeight
		}
		if height < contentConstraints.MinHeight {
			height = contentConstraints.MinHeight
		}
		w := s.width
		if w <= 0 {
			w = contentConstraints.MaxWidth
		}
		if w > contentConstraints.MaxWidth {
			w = contentConstraints.MaxWidth
		}
		return contentConstraints.Constrain(runtime.Size{Width: w, Height: height})
	})
}

// Render draws the sidebar.
func (s *Sidebar) Render(ctx runtime.RenderContext) {
	if s == nil {
		return
	}
	s.syncA11y()
	if !s.expanded {
		return
	}
	outer := s.bounds
	content := s.ContentBounds()
	if outer.Width <= 0 || outer.Height <= 0 {
		return
	}
	baseStyle := mergeBackendStyles(resolveBaseStyle(ctx, s, backend.DefaultStyle(), false), s.style)
	ctx.Buffer.Fill(outer, ' ', baseStyle)
	if content.Width <= 0 || content.Height <= 0 {
		return
	}
	rows := s.flatten()
	if len(rows) == 0 {
		return
	}
	if s.selected < 0 {
		s.selected = 0
	}
	if s.selected >= len(rows) {
		s.selected = len(rows) - 1
	}

	// Scroll viewport
	if s.selected < s.offset {
		s.offset = s.selected
	}
	if s.selected >= s.offset+content.Height {
		s.offset = s.selected - content.Height + 1
	}

	for i := 0; i < content.Height; i++ {
		rowIndex := s.offset + i
		if rowIndex < 0 || rowIndex >= len(rows) {
			break
		}
		row := rows[rowIndex]
		style := baseStyle
		if rowIndex == s.selected {
			style = mergeBackendStyles(baseStyle, s.selectedStyle)
		}

		indent := s.indent(row.depth)
		prefix := "  "
		if len(row.item.Children) > 0 {
			if row.item.Expanded {
				prefix = "- "
			} else {
				prefix = "+ "
			}
		}

		icon := ""
		if row.item.Icon != "" {
			icon = row.item.Icon + " "
		}

		line := indent + prefix + icon + row.item.Label

		// Draw badge right-aligned if present
		badge := row.item.Badge
		if badge != "" {
			badgeStr := " [" + badge + "]"
			badgeW := textWidth(badgeStr)
			labelW := textWidth(line)
			if labelW+badgeW <= content.Width {
				line = line + strings.Repeat(" ", content.Width-labelW-badgeW) + badgeStr
			} else {
				line = truncateString(line, content.Width-badgeW) + badgeStr
			}
		}

		line = truncateString(line, content.Width)
		writePadded(ctx.Buffer, content.X, content.Y+i, content.Width, line, style)
	}
}

// HandleMessage handles keyboard navigation.
func (s *Sidebar) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if s == nil || !s.focused || !s.expanded {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	rows := s.flatten()
	if len(rows) == 0 {
		return runtime.Unhandled()
	}

	switch key.Key {
	case terminal.KeyUp:
		s.setSelected(s.selected-1, len(rows))
		return runtime.Handled()
	case terminal.KeyDown:
		s.setSelected(s.selected+1, len(rows))
		return runtime.Handled()
	case terminal.KeyHome:
		s.setSelected(0, len(rows))
		return runtime.Handled()
	case terminal.KeyEnd:
		s.setSelected(len(rows)-1, len(rows))
		return runtime.Handled()
	case terminal.KeyLeft:
		if row := s.selectedRow(rows); row != nil {
			if row.item.Expanded && len(row.item.Children) > 0 {
				row.item.Expanded = false
				s.flatDirty = true
				if announcer := s.services.Announcer(); announcer != nil {
					announcer.Announce(fmt.Sprintf("%s collapsed", row.item.Label), accessibility.PriorityPolite)
				}
			}
		}
		return runtime.Handled()
	case terminal.KeyRight:
		if row := s.selectedRow(rows); row != nil && len(row.item.Children) > 0 {
			if !row.item.Expanded {
				row.item.Expanded = true
				s.flatDirty = true
				if announcer := s.services.Announcer(); announcer != nil {
					announcer.Announce(fmt.Sprintf("%s expanded", row.item.Label), accessibility.PriorityPolite)
				}
			}
		}
		return runtime.Handled()
	case terminal.KeyEnter:
		if row := s.selectedRow(rows); row != nil {
			// Toggle expand for parents
			if len(row.item.Children) > 0 {
				row.item.Expanded = !row.item.Expanded
				s.flatDirty = true
				if announcer := s.services.Announcer(); announcer != nil {
					if row.item.Expanded {
						announcer.Announce(fmt.Sprintf("%s expanded", row.item.Label), accessibility.PriorityPolite)
					} else {
						announcer.Announce(fmt.Sprintf("%s collapsed", row.item.Label), accessibility.PriorityPolite)
					}
				}
			}
			// Fire callback
			if row.item.OnSelect != nil {
				row.item.OnSelect()
			}
		}
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

// sidebarRow is a flattened sidebar item with depth info.
type sidebarRow struct {
	item  *SidebarItem
	depth int
}

func (s *Sidebar) flatten() []sidebarRow {
	if s == nil {
		return nil
	}
	if !s.flatDirty && s.flatCache != nil {
		return s.flatCache
	}
	rows := s.flatCache[:0]
	var walk func(items []SidebarItem, depth int)
	walk = func(items []SidebarItem, depth int) {
		for i := range items {
			item := &items[i]
			rows = append(rows, sidebarRow{item: item, depth: depth})
			if item.Expanded && len(item.Children) > 0 {
				walk(item.Children, depth+1)
			}
		}
	}
	walk(s.items, 0)
	s.flatCache = rows
	s.flatDirty = false
	return s.flatCache
}

func (s *Sidebar) setSelected(index int, count int) {
	if count == 0 {
		s.selected = 0
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= count {
		index = count - 1
	}
	prev := s.selected
	s.selected = index
	s.syncA11y()
	if prev != index {
		if announcer := s.services.Announcer(); announcer != nil {
			rows := s.flatten()
			if row := s.selectedRow(rows); row != nil {
				announcer.Announce(row.item.Label, accessibility.PriorityPolite)
			}
		}
	}
}

func (s *Sidebar) selectedRow(rows []sidebarRow) *sidebarRow {
	if s.selected < 0 || s.selected >= len(rows) {
		return nil
	}
	return &rows[s.selected]
}

func (s *Sidebar) indent(depth int) string {
	if depth <= 0 {
		return ""
	}
	if len(s.indentCache) == 0 {
		s.indentCache = []string{""}
	}
	for len(s.indentCache) <= depth {
		s.indentCache = append(s.indentCache, s.indentCache[len(s.indentCache)-1]+"  ")
	}
	return s.indentCache[depth]
}

func (s *Sidebar) syncA11y() {
	if s == nil {
		return
	}
	if s.Base.Role == "" {
		s.Base.Role = accessibility.RoleNavigation
	}
	if s.Base.Landmark == "" {
		s.Base.Landmark = accessibility.LandmarkNavigation
	}
	label := strings.TrimSpace(s.label)
	if label == "" {
		label = "Sidebar"
	}
	s.Base.Label = label
	s.Base.Orientation = "vertical"
	rows := s.flatten()
	s.Base.Description = fmt.Sprintf("%d items", len(rows))
	if row := s.selectedRow(rows); row != nil {
		val := row.item.Label
		if row.item.Badge != "" {
			val += " (" + row.item.Badge + ")"
		}
		s.Base.Value = &accessibility.ValueInfo{Text: val}
		if len(row.item.Children) > 0 {
			s.Base.State.Expanded = accessibility.BoolPtr(row.item.Expanded)
		} else {
			s.Base.State.Expanded = nil
		}
	} else {
		s.Base.Value = nil
		s.Base.State.Expanded = nil
	}
}

// ScrollBy scrolls selection by delta.
func (s *Sidebar) ScrollBy(dx, dy int) {
	if s == nil || dy == 0 {
		return
	}
	rows := s.flatten()
	s.setSelected(s.selected+dy, len(rows))
	s.Invalidate()
}

// ScrollTo scrolls to an absolute row index.
func (s *Sidebar) ScrollTo(x, y int) {
	if s == nil {
		return
	}
	rows := s.flatten()
	s.setSelected(y, len(rows))
	s.Invalidate()
}

// ScrollToStart scrolls to the first row.
func (s *Sidebar) ScrollToStart() {
	if s == nil {
		return
	}
	rows := s.flatten()
	s.setSelected(0, len(rows))
	s.Invalidate()
}

// ScrollToEnd scrolls to the last row.
func (s *Sidebar) ScrollToEnd() {
	if s == nil {
		return
	}
	rows := s.flatten()
	s.setSelected(len(rows)-1, len(rows))
	s.Invalidate()
}

// PageBy scrolls by a number of pages.
func (s *Sidebar) PageBy(pages int) {
	if s == nil {
		return
	}
	rows := s.flatten()
	pageSize := s.bounds.Height
	if pageSize < 1 {
		pageSize = 1
	}
	s.setSelected(s.selected+pages*pageSize, len(rows))
	s.Invalidate()
}

// Bind attaches app services.
func (s *Sidebar) Bind(services runtime.Services) {
	if s == nil {
		return
	}
	s.services = services
}

// Unbind releases app services.
func (s *Sidebar) Unbind() {
	if s == nil {
		return
	}
	s.services = runtime.Services{}
}

var _ runtime.Widget = (*Sidebar)(nil)
var _ runtime.Focusable = (*Sidebar)(nil)
var _ runtime.Bindable = (*Sidebar)(nil)
var _ runtime.Unbindable = (*Sidebar)(nil)
