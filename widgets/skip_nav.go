package widgets

import (
	"fmt"

	"m31labs.dev/fluffyui/accessibility"
	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/terminal"
)

// SkipNav is a skip-navigation landmark widget that lets keyboard users
// jump past navigation elements directly to main content. This is an
// essential accessibility pattern for screen reader and keyboard-only users.
//
// SkipNav is only visible when focused (zero height when unfocused).
// When the user presses Enter, focus transfers to the target widget.
type SkipNav struct {
	FocusableBase
	target   runtime.Widget // widget to focus when activated
	label    string
	services runtime.Services
	style    backend.Style
}

var (
	_ runtime.Widget              = (*SkipNav)(nil)
	_ runtime.Focusable           = (*SkipNav)(nil)
	_ runtime.Bindable            = (*SkipNav)(nil)
	_ runtime.Unbindable          = (*SkipNav)(nil)
	_ runtime.FocusLayoutAffecting = (*SkipNav)(nil)
)

// NewSkipNav creates a skip-navigation widget with the given label and target.
// The label is displayed when focused (e.g., "Skip to main content").
// The target is the widget that receives focus when the skip nav is activated.
func NewSkipNav(label string, target runtime.Widget) *SkipNav {
	s := &SkipNav{
		target: target,
		label:  label,
		style:  backend.DefaultStyle().Reverse(true),
	}
	s.Base.Role = accessibility.RoleLink
	s.Base.Label = label
	return s
}

// Bind attaches app services.
func (s *SkipNav) Bind(services runtime.Services) {
	if s == nil {
		return
	}
	s.services = services
}

// Unbind releases app services.
func (s *SkipNav) Unbind() {
	if s == nil {
		return
	}
	s.services = runtime.Services{}
}

// SetLabel updates the skip nav label text.
func (s *SkipNav) SetLabel(label string) {
	if s == nil {
		return
	}
	s.label = label
	s.Base.Label = label
}

// Label returns the skip nav label text.
func (s *SkipNav) Label() string {
	if s == nil {
		return ""
	}
	return s.label
}

// SetTarget updates the focus target widget.
func (s *SkipNav) SetTarget(target runtime.Widget) {
	if s == nil {
		return
	}
	s.target = target
}

// Target returns the focus target widget.
func (s *SkipNav) Target() runtime.Widget {
	if s == nil {
		return nil
	}
	return s.target
}

// SetStyle sets the display style used when the skip nav is visible.
func (s *SkipNav) SetStyle(style backend.Style) {
	if s == nil {
		return
	}
	s.style = style
}

// StyleType returns the selector type name.
func (s *SkipNav) StyleType() string {
	return "SkipNav"
}

// FocusAffectsLayout returns true because the skip nav appears/disappears
// based on focus state, which changes the layout.
func (s *SkipNav) FocusAffectsLayout() bool {
	return true
}

// Measure returns the desired size. When unfocused, the widget has zero height
// so it takes no visual space. When focused, it renders as a single line.
func (s *SkipNav) Measure(constraints runtime.Constraints) runtime.Size {
	if !s.focused {
		return constraints.Constrain(runtime.Size{Width: 0, Height: 0})
	}
	return s.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		w := textWidth(s.label)
		if w < 1 {
			w = 1
		}
		return contentConstraints.Constrain(runtime.Size{Width: w, Height: 1})
	})
}

// Render draws the skip nav label when focused.
func (s *SkipNav) Render(ctx runtime.RenderContext) {
	if s == nil || !s.focused {
		return
	}
	bounds := s.ContentBounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}

	text := clipString(s.label, bounds.Width)
	writePadded(ctx.Buffer, bounds.X, bounds.Y, bounds.Width, text, s.style)
}

// HandleMessage processes keyboard input. Enter activates the skip nav
// and transfers focus to the target widget.
func (s *SkipNav) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if s == nil || !s.focused {
		return runtime.Unhandled()
	}

	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}

	if key.Key == terminal.KeyEnter {
		s.activateTarget()
		return runtime.Handled()
	}

	return runtime.Unhandled()
}

// activateTarget transfers focus to the target widget.
func (s *SkipNav) activateTarget() {
	if s == nil || s.target == nil {
		return
	}

	// Announce the navigation
	if announcer := s.services.Announcer(); announcer != nil {
		announcer.Announce(
			fmt.Sprintf("Navigated to %s", s.label),
			accessibility.PriorityPolite,
		)
	}

	// Focus the target if it is focusable
	if focusable, ok := s.target.(runtime.Focusable); ok && focusable.CanFocus() {
		// Blur ourselves
		s.Blur()
		// Focus the target directly
		focusable.Focus()
	}
}
