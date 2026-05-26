package widgets

import (
	"fmt"
	"strings"

	"m31labs.dev/fluffyui/accessibility"
	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/scroll"
	"m31labs.dev/fluffyui/terminal"
)

// MenuItem describes a menu entry.
type MenuItem struct {
	ID       string
	Title    string
	Shortcut string
	Children []*MenuItem
	Expanded bool
	Disabled bool
	OnSelect func()
}

// Menu renders a vertical menu.
type Menu struct {
	FocusableBase
	Items         []*MenuItem
	selectedIndex int
	offset        int
	label         string
	labelTrimmed  string
	style         backend.Style
	selectedStyle backend.Style
	indentCache   []string
	flatCache     []menuRow
	flatDirty     bool
	itemsLen      int
	itemsFirst    *MenuItem
	services      runtime.Services
}

// NewMenu creates a new menu.
func NewMenu(items ...*MenuItem) *Menu {
	menu := &Menu{
		Items:         items,
		selectedIndex: 0,
		label:         "Menu",
		style:         backend.DefaultStyle(),
		selectedStyle: backend.DefaultStyle().Reverse(true),
		flatDirty:     true,
		itemsLen:      len(items),
		itemsFirst:    firstItem(items),
	}
	menu.Base.Role = accessibility.RoleMenu
	menu.Base.Landmark = accessibility.LandmarkNavigation
	menu.syncA11y()
	return menu
}

// Bind attaches app services.
func (m *Menu) Bind(services runtime.Services) {
	if m == nil {
		return
	}
	m.services = services
}

// Unbind releases app services.
func (m *Menu) Unbind() {
	if m == nil {
		return
	}
	m.services = runtime.Services{}
}

// SetStyle updates the menu base style.
func (m *Menu) SetStyle(style backend.Style) {
	if m == nil {
		return
	}
	m.style = style
}

// SetSelectedStyle updates the selected row style.
func (m *Menu) SetSelectedStyle(style backend.Style) {
	if m == nil {
		return
	}
	m.selectedStyle = style
}

// StyleType returns the selector type name.
func (m *Menu) StyleType() string {
	return "Menu"
}

// SetItems replaces the menu items and clears cached rows.
func (m *Menu) SetItems(items ...*MenuItem) {
	if m == nil {
		return
	}
	m.Items = items
	m.itemsLen = len(items)
	m.itemsFirst = firstItem(items)
	m.flatDirty = true
	m.syncA11y()
}

// SetLabel updates the accessibility label.
func (m *Menu) SetLabel(label string) {
	if m == nil {
		return
	}
	m.label = label
	m.labelTrimmed = strings.TrimSpace(label)
	m.syncA11y()
}

// Measure returns desired size.
func (m *Menu) Measure(constraints runtime.Constraints) runtime.Size {
	return m.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		count := len(m.flatten())
		height := min(count, contentConstraints.MaxHeight)
		if height <= 0 {
			height = contentConstraints.MinHeight
		}
		return contentConstraints.Constrain(runtime.Size{Width: contentConstraints.MaxWidth, Height: height})
	})
}

// Render draws the menu.
func (m *Menu) Render(ctx runtime.RenderContext) {
	if m == nil {
		return
	}
	m.syncA11y()
	outer := m.bounds
	content := m.ContentBounds()
	if outer.Width <= 0 || outer.Height <= 0 {
		return
	}
	baseStyle := mergeBackendStyles(resolveBaseStyle(ctx, m, backend.DefaultStyle(), false), m.style)
	ctx.Buffer.Fill(outer, ' ', baseStyle)
	if content.Width <= 0 || content.Height <= 0 {
		return
	}
	rows := m.flatten()
	if len(rows) == 0 {
		return
	}
	if m.selectedIndex < 0 {
		m.selectedIndex = 0
	}
	if m.selectedIndex >= len(rows) {
		m.selectedIndex = len(rows) - 1
	}
	if m.selectedIndex < m.offset {
		m.offset = m.selectedIndex
	}
	if m.selectedIndex >= m.offset+content.Height {
		m.offset = m.selectedIndex - content.Height + 1
	}
	for i := 0; i < content.Height; i++ {
		rowIndex := m.offset + i
		if rowIndex < 0 || rowIndex >= len(rows) {
			break
		}
		row := rows[rowIndex]
		style := baseStyle
		if rowIndex == m.selectedIndex {
			style = mergeBackendStyles(baseStyle, m.selectedStyle)
		}
		prefix := "  "
		if len(row.item.Children) > 0 {
			if row.item.Expanded {
				prefix = "- "
			} else {
				prefix = "+ "
			}
		}
		indent := m.indent(row.depth)
		line := indent + prefix + row.item.Title
		if row.item.Shortcut != "" {
			line += " (" + row.item.Shortcut + ")"
		}
		line = truncateString(line, content.Width)
		writePadded(ctx.Buffer, content.X, content.Y+i, content.Width, line, style)
	}
}

// HandleMessage handles navigation and selection.
func (m *Menu) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if m == nil || !m.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	rows := m.flatten()
	switch key.Key {
	case terminal.KeyUp:
		m.setSelected(m.selectedIndex-1, len(rows))
		return runtime.Handled()
	case terminal.KeyDown:
		m.setSelected(m.selectedIndex+1, len(rows))
		return runtime.Handled()
	case terminal.KeyLeft:
		if row := m.selectedRow(rows); row != nil && row.item.Expanded {
			row.item.Expanded = false
			m.flatDirty = true
		}
		return runtime.Handled()
	case terminal.KeyRight:
		if row := m.selectedRow(rows); row != nil && len(row.item.Children) > 0 {
			row.item.Expanded = true
			m.flatDirty = true
		}
		return runtime.Handled()
	case terminal.KeyHome:
		m.setSelected(0, len(rows))
		return runtime.Handled()
	case terminal.KeyEnd:
		m.setSelected(len(rows)-1, len(rows))
		return runtime.Handled()
	case terminal.KeyEnter:
		if row := m.selectedRow(rows); row != nil && !row.item.Disabled {
			if len(row.item.Children) > 0 {
				row.item.Expanded = !row.item.Expanded
				m.flatDirty = true
			}
			if row.item.OnSelect != nil {
				row.item.OnSelect()
			}
		}
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

type menuRow struct {
	item  *MenuItem
	depth int
}

func (m *Menu) flatten() []menuRow {
	currentFirst := firstItem(m.Items)
	if m.itemsLen != len(m.Items) || m.itemsFirst != currentFirst {
		m.itemsLen = len(m.Items)
		m.itemsFirst = currentFirst
		m.flatDirty = true
	}
	if !m.flatDirty {
		return m.flatCache
	}
	rows := m.flatCache[:0]
	var walk func(items []*MenuItem, depth int)
	walk = func(items []*MenuItem, depth int) {
		for _, item := range items {
			if item == nil {
				continue
			}
			rows = append(rows, menuRow{item: item, depth: depth})
			if item.Expanded {
				walk(item.Children, depth+1)
			}
		}
	}
	walk(m.Items, 0)
	m.flatCache = rows
	m.flatDirty = false
	return m.flatCache
}

func (m *Menu) setSelected(index int, count int) {
	if count == 0 {
		m.selectedIndex = 0
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= count {
		index = count - 1
	}
	prev := m.selectedIndex
	m.selectedIndex = index
	m.syncA11y()
	if prev != index {
		if announcer := m.services.Announcer(); announcer != nil {
			announcer.AnnounceChange(m)
		}
	}
}

func (m *Menu) selectedRow(rows []menuRow) *menuRow {
	if m.selectedIndex < 0 || m.selectedIndex >= len(rows) {
		return nil
	}
	return &rows[m.selectedIndex]
}

// ScrollBy scrolls selection by delta.
func (m *Menu) ScrollBy(dx, dy int) {
	if m == nil || dy == 0 {
		return
	}
	rows := m.flatten()
	m.setSelected(m.selectedIndex+dy, len(rows))
	m.Invalidate()
}

func (m *Menu) indent(depth int) string {
	if depth <= 0 {
		return ""
	}
	if len(m.indentCache) == 0 {
		m.indentCache = []string{""}
	}
	for len(m.indentCache) <= depth {
		m.indentCache = append(m.indentCache, m.indentCache[len(m.indentCache)-1]+"  ")
	}
	return m.indentCache[depth]
}

func firstItem(items []*MenuItem) *MenuItem {
	if len(items) == 0 {
		return nil
	}
	return items[0]
}

// ScrollTo scrolls to an absolute row index.
func (m *Menu) ScrollTo(x, y int) {
	if m == nil {
		return
	}
	rows := m.flatten()
	m.setSelected(y, len(rows))
	m.Invalidate()
}

// PageBy scrolls by a number of pages.
func (m *Menu) PageBy(pages int) {
	if m == nil {
		return
	}
	rows := m.flatten()
	pageSize := m.bounds.Height
	if pageSize < 1 {
		pageSize = 1
	}
	m.setSelected(m.selectedIndex+pages*pageSize, len(rows))
	m.Invalidate()
}

// ScrollToStart scrolls to the first row.
func (m *Menu) ScrollToStart() {
	if m == nil {
		return
	}
	rows := m.flatten()
	m.setSelected(0, len(rows))
	m.Invalidate()
}

// ScrollToEnd scrolls to the last row.
func (m *Menu) ScrollToEnd() {
	if m == nil {
		return
	}
	rows := m.flatten()
	m.setSelected(len(rows)-1, len(rows))
	m.Invalidate()
}

func (m *Menu) syncA11y() {
	if m == nil {
		return
	}
	if m.Base.Role == "" {
		m.Base.Role = accessibility.RoleMenu
	}
	m.Base.Orientation = "vertical"
	label := m.labelTrimmed
	if label == "" {
		label = "Menu"
	}
	m.Base.Label = label
	if m.flatDirty {
		m.flatten()
	}
	m.Base.Description = fmt.Sprintf("%d items", len(m.flatCache))
	if row := m.selectedRow(m.flatCache); row != nil && row.item != nil {
		m.Base.Value = &accessibility.ValueInfo{Text: row.item.Title}
		if len(row.item.Children) > 0 {
			m.Base.State.Expanded = accessibility.BoolPtr(row.item.Expanded)
			m.Base.HasPopup = "menu"
		} else {
			m.Base.State.Expanded = nil
			m.Base.HasPopup = ""
		}
		m.Base.Level = row.depth + 1 // aria-level for nested menus
	} else {
		m.Base.Value = nil
		m.Base.State.Expanded = nil
		m.Base.HasPopup = ""
		m.Base.Level = 0
	}
}

var _ scroll.Controller = (*Menu)(nil)

var _ runtime.Widget = (*Menu)(nil)
var _ runtime.Focusable = (*Menu)(nil)
var _ runtime.Bindable = (*Menu)(nil)
var _ runtime.Unbindable = (*Menu)(nil)
