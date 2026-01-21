package widgets

import (
	"fmt"
	"strings"

	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/terminal"
)

// SectionItem represents a single item in a collapsible section.
type SectionItem struct {
	Icon    rune   // Status icon (✓, →, ○, ⟳)
	Text    string // Item text
	Active  bool   // Whether this item is currently active/in-progress
	SubText string // Optional secondary text (e.g., spinner detail)
}

// Section represents a collapsible section in the sidebar.
type Section struct {
	FocusableBase

	title    string
	items    []SectionItem
	expanded bool
	maxItems int // Max items to show when expanded (0 = all)
	label    string

	// Styles
	headerStyle   backend.Style
	itemStyle     backend.Style
	activeStyle   backend.Style
	iconStyle     backend.Style
	completedIcon backend.Style
	pendingIcon   backend.Style
	activeIcon    backend.Style
}

// NewSection creates a new collapsible section.
func NewSection(title string) *Section {
	section := &Section{
		title:         title,
		expanded:      true,
		maxItems:      0,
		label:         "Section",
		headerStyle:   backend.DefaultStyle().Bold(true),
		itemStyle:     backend.DefaultStyle(),
		activeStyle:   backend.DefaultStyle().Bold(true),
		iconStyle:     backend.DefaultStyle(),
		completedIcon: backend.DefaultStyle().Foreground(backend.ColorGreen),
		pendingIcon:   backend.DefaultStyle().Foreground(backend.ColorDefault),
		activeIcon:    backend.DefaultStyle().Foreground(backend.ColorYellow),
	}
	section.Base.Role = accessibility.RoleGroup
	section.syncA11y()
	return section
}

// SetTitle updates the section title.
func (s *Section) SetTitle(title string) {
	s.title = title
	s.syncA11y()
}

// SetItems updates the section items.
func (s *Section) SetItems(items []SectionItem) {
	s.items = items
	s.syncA11y()
}

// SetExpanded sets the expanded state.
func (s *Section) SetExpanded(expanded bool) {
	s.expanded = expanded
	s.syncA11y()
}

// IsExpanded returns the expanded state.
func (s *Section) IsExpanded() bool {
	return s.expanded
}

// Toggle toggles the expanded state.
func (s *Section) Toggle() {
	s.expanded = !s.expanded
	s.syncA11y()
}

// SetMaxItems sets the maximum items to show when expanded.
func (s *Section) SetMaxItems(max int) {
	s.maxItems = max
}

// SetStyles configures the section appearance.
func (s *Section) SetStyles(header, item, active, completed, pending, activeIcon backend.Style) {
	s.headerStyle = header
	s.itemStyle = item
	s.activeStyle = active
	s.completedIcon = completed
	s.pendingIcon = pending
	s.activeIcon = activeIcon
}

// Measure returns the preferred size.
func (s *Section) Measure(constraints runtime.Constraints) runtime.Size {
	height := 1 // Header always shown

	if s.expanded {
		itemCount := len(s.items)
		if s.maxItems > 0 && itemCount > s.maxItems {
			itemCount = s.maxItems
		}
		height += itemCount
	}

	return runtime.Size{
		Width:  constraints.MaxWidth,
		Height: height,
	}
}

// Layout stores the assigned bounds.
func (s *Section) Layout(bounds runtime.Rect) {
	s.Base.Layout(bounds)
}

// Render draws the section.
func (s *Section) Render(ctx runtime.RenderContext) {
	b := s.bounds
	if b.Width < 5 || b.Height < 1 {
		return
	}
	s.syncA11y()

	y := b.Y

	// Draw header
	expandIcon := '▼'
	if !s.expanded {
		expandIcon = '▶'
	}
	ctx.Buffer.Set(b.X, y, expandIcon, s.headerStyle)
	ctx.Buffer.Set(b.X+1, y, ' ', s.headerStyle)

	title := s.title
	if len(title) > b.Width-3 {
		title = title[:b.Width-3]
	}
	ctx.Buffer.SetString(b.X+2, y, title, s.headerStyle)
	y++

	// Draw items if expanded
	if s.expanded {
		itemCount := len(s.items)
		if s.maxItems > 0 && itemCount > s.maxItems {
			itemCount = s.maxItems
		}

		for i := 0; i < itemCount && y < b.Y+b.Height; i++ {
			item := s.items[i]

			// Indent
			ctx.Buffer.Set(b.X, y, ' ', s.itemStyle)
			ctx.Buffer.Set(b.X+1, y, ' ', s.itemStyle)

			// Icon
			iconStyle := s.iconStyle
			switch item.Icon {
			case '✓':
				iconStyle = s.completedIcon
			case '→', '⟳':
				iconStyle = s.activeIcon
			case '○':
				iconStyle = s.pendingIcon
			}
			ctx.Buffer.Set(b.X+2, y, item.Icon, iconStyle)
			ctx.Buffer.Set(b.X+3, y, ' ', s.itemStyle)

			// Text
			textStyle := s.itemStyle
			if item.Active {
				textStyle = s.activeStyle
			}

			text := item.Text
			maxText := b.Width - 5
			if len(text) > maxText {
				text = text[:maxText-3] + "..."
			}
			ctx.Buffer.SetString(b.X+4, y, text, textStyle)
			y++

			// SubText on next line if present
			if item.SubText != "" && y < b.Y+b.Height {
				ctx.Buffer.SetString(b.X+4, y, "  "+item.SubText, s.itemStyle)
				y++
			}
		}
	}
}

// HandleMessage processes input.
func (s *Section) HandleMessage(msg runtime.Message) runtime.HandleResult {
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}

	// Space or Enter toggles expansion
	if key.Key == terminal.KeyEnter || (key.Key == terminal.KeyRune && key.Rune == ' ') {
		s.expanded = !s.expanded
		s.syncA11y()
		return runtime.Handled()
	}

	return runtime.Unhandled()
}

// ContentHeight returns the height needed for content (items).
func (s *Section) ContentHeight() int {
	if !s.expanded {
		return 0
	}
	itemCount := len(s.items)
	if s.maxItems > 0 && itemCount > s.maxItems {
		itemCount = s.maxItems
	}
	return itemCount
}

func (s *Section) syncA11y() {
	if s == nil {
		return
	}
	if s.Base.Role == "" {
		s.Base.Role = accessibility.RoleGroup
	}
	label := strings.TrimSpace(s.title)
	if label == "" {
		label = strings.TrimSpace(s.label)
	}
	if label == "" {
		label = "Section"
	}
	s.Base.Label = label
	state := "collapsed"
	if s.expanded {
		state = "expanded"
	}
	s.Base.State.Expanded = accessibility.BoolPtr(s.expanded)
	s.Base.Description = fmt.Sprintf("%s, %d items", state, len(s.items))
	if active := s.activeItemLabel(); active != "" {
		s.Base.Value = &accessibility.ValueInfo{Text: active}
	} else {
		s.Base.Value = nil
	}
}

func (s *Section) activeItemLabel() string {
	if s == nil {
		return ""
	}
	for _, item := range s.items {
		if item.Active {
			return item.Text
		}
	}
	return ""
}
