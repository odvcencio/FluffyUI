package widgets

import (
	"strings"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

// MultiSelectOption represents an option in a multi-select list.
type MultiSelectOption struct {
	Label    string
	Value    any
	Disabled bool
}

// MultiSelect renders a list of options with multiple selection.
type MultiSelect struct {
	FocusableBase
	options       []MultiSelectOption
	selected      int
	offset        int
	checked       map[int]bool
	label         string
	style         backend.Style
	selectedStyle backend.Style
	checkedStyle  backend.Style
	disabledStyle backend.Style
	onChange      func(selected []MultiSelectOption)
}

// NewMultiSelect creates a new multi-select list.
func NewMultiSelect(options ...MultiSelectOption) *MultiSelect {
	m := &MultiSelect{
		options:       options,
		selected:      0,
		label:         "Multi Select",
		style:         backend.DefaultStyle(),
		selectedStyle: backend.DefaultStyle().Reverse(true),
		checkedStyle:  backend.DefaultStyle().Foreground(backend.ColorGreen),
		disabledStyle: backend.DefaultStyle().Dim(true),
		checked:       map[int]bool{},
	}
	m.Base.Role = accessibility.RoleList
	m.syncA11y()
	return m
}

// SetOptions updates the list of options.
func (m *MultiSelect) SetOptions(options []MultiSelectOption) {
	if m == nil {
		return
	}
	m.options = options
	m.selected = 0
	m.offset = 0
	m.checked = map[int]bool{}
	m.syncA11y()
}

// SetOnChange registers a change callback.
func (m *MultiSelect) SetOnChange(fn func(selected []MultiSelectOption)) {
	if m == nil {
		return
	}
	m.onChange = fn
}

// SetLabel updates the accessibility label.
func (m *MultiSelect) SetLabel(label string) {
	if m == nil {
		return
	}
	m.label = label
	m.syncA11y()
}

// SelectedOptions returns the currently selected options.
func (m *MultiSelect) SelectedOptions() []MultiSelectOption {
	if m == nil {
		return nil
	}
	out := make([]MultiSelectOption, 0, len(m.checked))
	for idx, ok := range m.checked {
		if ok && idx >= 0 && idx < len(m.options) {
			out = append(out, m.options[idx])
		}
	}
	return out
}

// StyleType returns the selector type name.
func (m *MultiSelect) StyleType() string {
	return "MultiSelect"
}

// Measure returns desired size.
func (m *MultiSelect) Measure(constraints runtime.Constraints) runtime.Size {
	return m.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		width := 0
		for _, opt := range m.options {
			w := textWidth(opt.Label) + 4
			if w > width {
				width = w
			}
		}
		if width < 6 {
			width = 6
		}
		height := len(m.options)
		if height < 1 {
			height = 1
		}
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: height})
	})
}

// Render draws the options list.
func (m *MultiSelect) Render(ctx runtime.RenderContext) {
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
	m.ensureVisible(content.Height)
	for i := 0; i < content.Height; i++ {
		idx := m.offset + i
		if idx < 0 || idx >= len(m.options) {
			break
		}
		opt := m.options[idx]
		checked := m.checked[idx]
		prefix := "[ ] "
		style := baseStyle
		if checked {
			prefix = "[x] "
			style = mergeBackendStyles(style, m.checkedStyle)
		}
		if idx == m.selected {
			style = mergeBackendStyles(style, m.selectedStyle)
		}
		if opt.Disabled {
			style = mergeBackendStyles(style, m.disabledStyle)
		}
		line := prefix + opt.Label
		line = truncateString(line, content.Width)
		writePadded(ctx.Buffer, content.X, content.Y+i, content.Width, line, style)
	}
}

// HandleMessage processes navigation and toggling.
func (m *MultiSelect) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if m == nil || !m.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	switch key.Key {
	case terminal.KeyUp:
		m.moveSelection(-1)
		return runtime.Handled()
	case terminal.KeyDown:
		m.moveSelection(1)
		return runtime.Handled()
	case terminal.KeyHome:
		m.selected = 0
		return runtime.Handled()
	case terminal.KeyEnd:
		if len(m.options) > 0 {
			m.selected = len(m.options) - 1
		}
		return runtime.Handled()
	case terminal.KeyEnter, terminal.KeyRune:
		if key.Key == terminal.KeyRune && key.Rune != ' ' {
			break
		}
		m.toggleSelected()
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func (m *MultiSelect) moveSelection(delta int) {
	if m == nil || len(m.options) == 0 {
		return
	}
	next := m.selected + delta
	if next < 0 {
		next = 0
	}
	if next >= len(m.options) {
		next = len(m.options) - 1
	}
	m.selected = next
}

func (m *MultiSelect) toggleSelected() {
	if m == nil || m.selected < 0 || m.selected >= len(m.options) {
		return
	}
	if m.options[m.selected].Disabled {
		return
	}
	m.checked[m.selected] = !m.checked[m.selected]
	m.syncA11y()
	if m.onChange != nil {
		m.onChange(m.SelectedOptions())
	}
}

func (m *MultiSelect) ensureVisible(height int) {
	if height <= 0 {
		return
	}
	if m.selected < m.offset {
		m.offset = m.selected
	}
	if m.selected >= m.offset+height {
		m.offset = m.selected - height + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
	maxOffset := max(0, len(m.options)-height)
	if m.offset > maxOffset {
		m.offset = maxOffset
	}
}

func (m *MultiSelect) syncA11y() {
	if m == nil {
		return
	}
	if m.Base.Role == "" {
		m.Base.Role = accessibility.RoleList
	}
	label := strings.TrimSpace(m.label)
	if label == "" {
		label = "Multi Select"
	}
	m.Base.Label = label
	m.Base.Description = "multi-select list"
}

var _ runtime.Widget = (*MultiSelect)(nil)
var _ runtime.Focusable = (*MultiSelect)(nil)
