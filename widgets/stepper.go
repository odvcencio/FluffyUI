package widgets

import (
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
)

// StepState describes the state of a step.
type StepState int

const (
	StepPending StepState = iota
	StepActive
	StepCompleted
	StepError
)

// Step describes a step in a stepper.
type Step struct {
	Title string
	State StepState
}

// Stepper renders a sequence of steps.
type Stepper struct {
	Base
	Steps []Step
	style backend.Style
}

// NewStepper creates a stepper.
func NewStepper(steps ...Step) *Stepper {
	return &Stepper{Steps: steps, style: backend.DefaultStyle()}
}

// Measure returns desired size.
func (s *Stepper) Measure(constraints runtime.Constraints) runtime.Size {
	width := 0
	for i, step := range s.Steps {
		width += len(step.Title) + 4
		if i < len(s.Steps)-1 {
			width += 3
		}
	}
	if width < 1 {
		width = 1
	}
	return constraints.Constrain(runtime.Size{Width: width, Height: 1})
}

// Render draws the stepper.
func (s *Stepper) Render(ctx runtime.RenderContext) {
	if s == nil {
		return
	}
	bounds := s.bounds
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	text := ""
	for i, step := range s.Steps {
		prefix := "[ ]"
		switch step.State {
		case StepActive:
			prefix = "[>]"
		case StepCompleted:
			prefix = "[x]"
		case StepError:
			prefix = "[!]"
		}
		if i > 0 {
			text += " -> "
		}
		text += prefix + " " + step.Title
	}
	text = truncateString(text, bounds.Width)
	writePadded(ctx.Buffer, bounds.X, bounds.Y, bounds.Width, text, s.style)
}

// HandleMessage returns unhandled.
func (s *Stepper) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}
