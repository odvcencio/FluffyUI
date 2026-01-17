package widgets

import (
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
)

// Spinner is an animated loading indicator.
type Spinner struct {
	Base
	Frames []string
	index  int
	style  backend.Style
}

// NewSpinner creates a spinner.
func NewSpinner() *Spinner {
	return &Spinner{
		Frames: []string{"-", "\\", "|", "/"},
		style:  backend.DefaultStyle(),
	}
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
	return constraints.Constrain(runtime.Size{Width: 1, Height: 1})
}

// Render draws the spinner frame.
func (s *Spinner) Render(ctx runtime.RenderContext) {
	if s == nil || len(s.Frames) == 0 {
		return
	}
	bounds := s.bounds
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	frame := s.Frames[s.index%len(s.Frames)]
	frame = truncateString(frame, bounds.Width)
	ctx.Buffer.SetString(bounds.X, bounds.Y, frame, s.style)
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
