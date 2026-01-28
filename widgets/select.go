package widgets

import (
	"strings"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	uistyle "github.com/odvcencio/fluffyui/style"
	"github.com/odvcencio/fluffyui/terminal"
)

// SelectOption represents a selectable option.
type SelectOption struct {
	Label    string
	Value    any
	Disabled bool
}

// SelectMode controls how the select renders.
type SelectMode int

const (
	SelectInline SelectMode = iota
	SelectDropdown
)

// SelectModeOption configures select behavior.
type SelectModeOption func(*Select)

// Select is a dropdown-like selector (inline).
type Select struct {
	FocusableBase

	options      []SelectOption
	selected     int
	onChange     func(option SelectOption)
	label        string
	style        backend.Style
	focusStyle   backend.Style
	styleSet     bool
	focusSet     bool
	services     runtime.Services
	mode         SelectMode
	dropdownOpen bool
}

// NewSelect creates a select widget.
func NewSelect(options ...SelectOption) *Select {
	s := &Select{
		options:    options,
		selected:   0,
		label:      "Select",
		style:      backend.DefaultStyle(),
		focusStyle: backend.DefaultStyle().Reverse(true),
		mode:       SelectInline,
	}
	s.Base.Role = accessibility.RoleList
	s.syncState()
	return s
}

// Apply configures the select with mode options.
func (s *Select) Apply(opts ...SelectModeOption) *Select {
	for _, opt := range opts {
		if opt != nil {
			opt(s)
		}
	}
	return s
}

// WithDropdownMode enables dropdown overlay rendering.
func WithDropdownMode() SelectModeOption {
	return func(s *Select) {
		if s == nil {
			return
		}
		s.mode = SelectDropdown
	}
}

// WithInlineMode forces inline rendering.
func WithInlineMode() SelectModeOption {
	return func(s *Select) {
		if s == nil {
			return
		}
		s.mode = SelectInline
	}
}

// SetMode updates the select rendering mode.
func (s *Select) SetMode(mode SelectMode) {
	if s == nil {
		return
	}
	s.mode = mode
}

// Bind attaches app services.
func (s *Select) Bind(services runtime.Services) {
	if s == nil {
		return
	}
	s.services = services
}

// Unbind releases app services.
func (s *Select) Unbind() {
	if s == nil {
		return
	}
	s.services = runtime.Services{}
}

// SetOnChange sets the change handler.
func (s *Select) SetOnChange(fn func(option SelectOption)) {
	if s == nil {
		return
	}
	s.onChange = fn
}

// SetLabel updates the accessibility label.
func (s *Select) SetLabel(label string) {
	if s == nil {
		return
	}
	s.label = label
	s.syncState()
}

// SetStyle sets the normal style.
func (s *Select) SetStyle(style backend.Style) {
	if s == nil {
		return
	}
	s.style = style
	s.styleSet = true
}

// SetFocusStyle sets the focused style.
func (s *Select) SetFocusStyle(style backend.Style) {
	if s == nil {
		return
	}
	s.focusStyle = style
	s.focusSet = true
}

// StyleType returns the selector type name.
func (s *Select) StyleType() string {
	return "Select"
}

// Selected returns the current selection index.
func (s *Select) Selected() int {
	if s == nil {
		return 0
	}
	return s.selected
}

// SelectedOption returns the current option.
func (s *Select) SelectedOption() (SelectOption, bool) {
	if s == nil || s.selected < 0 || s.selected >= len(s.options) {
		return SelectOption{}, false
	}
	return s.options[s.selected], true
}

// SetSelected updates the selected index.
func (s *Select) SetSelected(index int) {
	if s == nil || index < 0 || index >= len(s.options) {
		return
	}
	if s.options[index].Disabled {
		return
	}
	s.selected = index
	s.syncState()
	s.relayout()
	if s.onChange != nil {
		s.onChange(s.options[index])
	}
}

// Measure returns the desired size.
func (s *Select) Measure(constraints runtime.Constraints) runtime.Size {
	return s.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		label := s.currentLabel()
		width := textWidth(label) + 4
		if width < 6 {
			width = 6
		}
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: 1})
	})
}

// Render draws the select.
func (s *Select) Render(ctx runtime.RenderContext) {
	if s == nil {
		return
	}
	outer := s.bounds
	content := s.ContentBounds()
	if outer.Width <= 0 || outer.Height <= 0 {
		return
	}
	label := s.currentLabel()
	available := max(0, content.Width-4)
	text := "[" + truncateString(label, available) + " v]"
	style := s.style
	resolved := ctx.ResolveStyle(s)
	if !resolved.IsZero() {
		final := resolved
		if s.styleSet {
			final = final.Merge(uistyle.FromBackend(s.style))
		}
		if s.focused && s.focusSet {
			final = final.Merge(uistyle.FromBackend(s.focusStyle))
		}
		style = final.ToBackend()
	} else if s.focused {
		style = s.focusStyle
	}
	ctx.Buffer.Fill(outer, ' ', style)
	if content.Width <= 0 || content.Height <= 0 {
		return
	}
	writePadded(ctx.Buffer, content.X, content.Y, content.Width, text, style)
}

// HandleMessage changes selection.
func (s *Select) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if s == nil {
		return runtime.Unhandled()
	}
	if mouse, ok := msg.(runtime.MouseMsg); ok {
		if s.mode == SelectDropdown && mouse.Action == runtime.MousePress && mouse.Button == runtime.MouseLeft {
			if s.bounds.Contains(mouse.X, mouse.Y) {
				return s.openDropdown()
			}
		}
	}
	if !s.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	switch key.Key {
	case terminal.KeyEnter:
		if s.mode == SelectDropdown {
			return s.openDropdown()
		}
	case terminal.KeyUp, terminal.KeyLeft:
		if s.moveSelection(-1) {
			return runtime.Handled()
		}
	case terminal.KeyDown, terminal.KeyRight:
		if s.moveSelection(1) {
			return runtime.Handled()
		}
	case terminal.KeyHome:
		s.SetSelected(0)
		return runtime.Handled()
	case terminal.KeyEnd:
		s.SetSelected(len(s.options) - 1)
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func (s *Select) openDropdown() runtime.HandleResult {
	if s == nil {
		return runtime.Unhandled()
	}
	if s.dropdownOpen {
		return runtime.Handled()
	}
	s.dropdownOpen = true
	dropdown := newSelectDropdown(s)
	popover := NewPopover(s.bounds, dropdown,
		WithPopoverMatchAnchorWidth(true),
		WithPopoverDismissOnOutside(true),
		WithPopoverDismissOnEscape(true),
		WithPopoverOnClose(func() {
			s.dropdownOpen = false
		}),
	)
	return runtime.WithCommand(runtime.PushOverlay{Widget: popover, Modal: true})
}

func (s *Select) moveSelection(delta int) bool {
	if s == nil || len(s.options) == 0 {
		return false
	}
	index := s.selected
	for i := 0; i < len(s.options); i++ {
		index += delta
		if index < 0 {
			index = len(s.options) - 1
		} else if index >= len(s.options) {
			index = 0
		}
		if !s.options[index].Disabled {
			s.SetSelected(index)
			return true
		}
	}
	return false
}

func (s *Select) currentLabel() string {
	if opt, ok := s.SelectedOption(); ok {
		return opt.Label
	}
	return ""
}

func (s *Select) syncState() {
	if s == nil {
		return
	}
	label := strings.TrimSpace(s.label)
	if label == "" {
		label = "Select"
	}
	if s.Base.Role == "" {
		s.Base.Role = accessibility.RoleList
	}
	s.Base.Label = label
	if opt, ok := s.SelectedOption(); ok {
		s.Base.Value = &accessibility.ValueInfo{Text: opt.Label}
	} else {
		s.Base.Value = nil
	}
}

func (s *Select) relayout() {
	if s == nil {
		return
	}
	s.Invalidate()
	s.services.Relayout()
}

var _ runtime.Widget = (*Select)(nil)
var _ runtime.Focusable = (*Select)(nil)
var _ runtime.Bindable = (*Select)(nil)
var _ runtime.Unbindable = (*Select)(nil)
