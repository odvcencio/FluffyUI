package widgets

import (
	"fmt"
	"strings"

	"m31labs.dev/fluffyui/accessibility"
	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/runtime"
	uistyle "m31labs.dev/fluffyui/style"
	"m31labs.dev/fluffyui/terminal"
)

// SegmentedControl is a horizontal row of mutually exclusive options.
type SegmentedControl struct {
	FocusableBase
	options       []string
	selected      int
	onChange      func(index int, label string)
	style         backend.Style
	selectedStyle backend.Style
	styleSet      bool
	selectedSet   bool
	services      runtime.Services
}

// NewSegmentedControl creates a segmented control with the given options.
func NewSegmentedControl(options ...string) *SegmentedControl {
	s := &SegmentedControl{
		options:       options,
		style:         backend.DefaultStyle(),
		selectedStyle: backend.DefaultStyle().Reverse(true),
	}
	s.Base.Role = accessibility.RoleRadioGroup
	s.Base.Orientation = "horizontal"
	s.syncA11y()
	return s
}

// Selected returns the currently selected index.
func (s *SegmentedControl) Selected() int {
	if s == nil { return 0 }
	return s.selected
}

// SetSelected updates the selected index.
func (s *SegmentedControl) SetSelected(index int) {
	if s == nil || index < 0 || index >= len(s.options) || s.selected == index { return }
	s.selected = index
	s.syncA11y()
	if announcer := s.services.Announcer(); announcer != nil {
		announcer.AnnounceChange(s)
	}
	s.relayout()
	if s.onChange != nil { s.onChange(index, s.options[index]) }
}

// SetOnChange sets the change handler.
func (s *SegmentedControl) SetOnChange(fn func(index int, label string)) {
	if s == nil { return }
	s.onChange = fn
}

// SetStyle sets the normal style.
func (s *SegmentedControl) SetStyle(style backend.Style) {
	if s == nil { return }
	s.style = style
	s.styleSet = true
}

// SetSelectedStyle sets the selected segment style.
func (s *SegmentedControl) SetSelectedStyle(style backend.Style) {
	if s == nil { return }
	s.selectedStyle = style
	s.selectedSet = true
}

// StyleType returns the selector type name.
func (s *SegmentedControl) StyleType() string { return "SegmentedControl" }

// Bind attaches app services.
func (s *SegmentedControl) Bind(services runtime.Services) {
	if s == nil { return }
	s.services = services
}

// Unbind releases app services.
func (s *SegmentedControl) Unbind() {
	if s == nil { return }
	s.services = runtime.Services{}
}

// Measure returns the desired size.
func (s *SegmentedControl) Measure(constraints runtime.Constraints) runtime.Size {
	return s.measureWithStyle(constraints, func(cc runtime.Constraints) runtime.Size {
		width := 2 // [ and ]
		for i, opt := range s.options {
			width += textWidth(opt) + 2
			if i < len(s.options)-1 { width++ }
		}
		if width < 4 { width = 4 }
		return cc.Constrain(runtime.Size{Width: width, Height: 1})
	})
}

// Render draws the segmented control as [Option A | Option B | Option C].
func (s *SegmentedControl) Render(ctx runtime.RenderContext) {
	if s == nil { return }
	outer := s.bounds
	content := s.ContentBounds()
	if outer.Width <= 0 || outer.Height <= 0 { return }
	baseStyle := s.style
	resolved := ctx.ResolveStyle(s)
	if !resolved.IsZero() {
		final := resolved
		if s.styleSet { final = final.Merge(uistyle.FromBackend(s.style)) }
		baseStyle = final.ToBackend()
	}
	ctx.Buffer.Fill(outer, ' ', baseStyle)
	if content.Width <= 0 || content.Height <= 0 { return }

	// Build segments and render with per-segment highlighting
	var sb strings.Builder
	sb.WriteByte('[')
	for i, opt := range s.options {
		sb.WriteString(" " + opt + " ")
		if i < len(s.options)-1 { sb.WriteByte('|') }
	}
	sb.WriteByte(']')
	line := sb.String()

	x := content.X
	segIdx := -1
	inSegment := false
	for _, r := range line {
		if x >= content.X+content.Width { break }
		if r == '[' { segIdx = 0; inSegment = true }
		st := baseStyle
		if inSegment && segIdx == s.selected { st = s.selectedStyle }
		ctx.Buffer.SetString(x, content.Y, string(r), st)
		x += textWidth(string(r))
		if r == '|' { segIdx++ }
	}
}

// HandleMessage handles keyboard navigation.
func (s *SegmentedControl) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if s == nil || !s.focused { return runtime.Unhandled() }
	key, ok := msg.(runtime.KeyMsg)
	if !ok { return runtime.Unhandled() }
	switch key.Key {
	case terminal.KeyLeft:
		if s.selected > 0 { s.SetSelected(s.selected - 1); return runtime.Handled() }
	case terminal.KeyRight:
		if s.selected < len(s.options)-1 { s.SetSelected(s.selected + 1); return runtime.Handled() }
	case terminal.KeyHome:
		s.SetSelected(0); return runtime.Handled()
	case terminal.KeyEnd:
		s.SetSelected(len(s.options) - 1); return runtime.Handled()
	}
	return runtime.Unhandled()
}

func (s *SegmentedControl) syncA11y() {
	if s == nil { return }
	if s.Base.Role == "" { s.Base.Role = accessibility.RoleRadioGroup }
	s.Base.Orientation = "horizontal"
	if s.selected >= 0 && s.selected < len(s.options) {
		s.Base.Value = &accessibility.ValueInfo{
			Text: s.options[s.selected], Current: float64(s.selected + 1),
			Min: 1, Max: float64(len(s.options)),
		}
		s.Base.PosInSet = s.selected + 1
		s.Base.SetSize = len(s.options)
		s.Base.State.Selected = true
	} else {
		s.Base.Value = nil
		s.Base.PosInSet = 0
		s.Base.SetSize = len(s.options)
		s.Base.State.Selected = false
	}
	if len(s.options) > 0 {
		s.Base.Description = fmt.Sprintf("%d of %d, use left right to switch",
			s.selected+1, len(s.options))
	}
}

func (s *SegmentedControl) relayout() {
	if s == nil { return }
	s.Invalidate()
	s.services.Relayout()
}

var _ runtime.Widget = (*SegmentedControl)(nil)
var _ runtime.Focusable = (*SegmentedControl)(nil)
var _ runtime.Bindable = (*SegmentedControl)(nil)
var _ runtime.Unbindable = (*SegmentedControl)(nil)
