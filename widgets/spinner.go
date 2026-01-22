package widgets

import (
	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
)

// Spinner is an animated loading indicator.
type Spinner struct {
	Base
	Frames []string
	index  int
	style  backend.Style
	styleSet bool
}

// NewSpinner creates a spinner.
func NewSpinner() *Spinner {
	spinner := &Spinner{
		Frames: []string{"-", "\\", "|", "/"},
		style:  backend.DefaultStyle(),
	}
	spinner.Base.Role = accessibility.RoleStatus
	spinner.Base.Label = "Loading"
	return spinner
}

// SetStyle updates the spinner style.
func (s *Spinner) SetStyle(style backend.Style) {
	if s == nil {
		return
	}
	s.style = style
	s.styleSet = true
}

// StyleType returns the selector type name.
func (s *Spinner) StyleType() string {
	return "Spinner"
}

// Advance moves to the next frame.
func (s *Spinner) Advance() {
	if s == nil || len(s.Frames) == 0 {
		return
	}
	s.index = (s.index + 1) % len(s.Frames)
}

// Measure returns desired size.
func (s *Spinner) Measure(constraints runtime.Constraints) runtime.Size {
	return s.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		return contentConstraints.Constrain(runtime.Size{Width: 1, Height: 1})
	})
}

// Render draws the spinner frame.
func (s *Spinner) Render(ctx runtime.RenderContext) {
	if s == nil || len(s.Frames) == 0 {
		return
	}
	s.syncA11y()
	bounds := s.ContentBounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	frame := s.Frames[s.index%len(s.Frames)]
	frame = truncateString(frame, bounds.Width)
	style := resolveBaseStyle(ctx, s, s.style, s.styleSet)
	ctx.Buffer.SetString(bounds.X, bounds.Y, frame, style)
}

// HandleMessage advances on ticks.
func (s *Spinner) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if s == nil {
		return runtime.Unhandled()
	}
	if _, ok := msg.(runtime.TickMsg); ok {
		s.Advance()
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func (s *Spinner) syncA11y() {
	if s == nil {
		return
	}
	if s.Base.Role == "" {
		s.Base.Role = accessibility.RoleStatus
	}
	if s.Base.Label == "" {
		s.Base.Label = "Loading"
	}
}
