package widgets

import (
	"fmt"
	"strings"

	"github.com/odvcencio/fluffy-ui/accessibility"
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
	label string
	styleSet bool
}

// NewStepper creates a stepper.
func NewStepper(steps ...Step) *Stepper {
	stepper := &Stepper{Steps: steps, style: backend.DefaultStyle(), label: "Steps"}
	stepper.Base.Role = accessibility.RoleList
	stepper.syncA11y()
	return stepper
}

// SetStyle updates the stepper style.
func (s *Stepper) SetStyle(style backend.Style) {
	if s == nil {
		return
	}
	s.style = style
	s.styleSet = true
}

// StyleType returns the selector type name.
func (s *Stepper) StyleType() string {
	return "Stepper"
}

// Measure returns desired size.
func (s *Stepper) Measure(constraints runtime.Constraints) runtime.Size {
	return s.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		width := 0
		for i, step := range s.Steps {
			width += textWidth(step.Title) + 4
			if i < len(s.Steps)-1 {
				width += 3
			}
		}
		if width < 1 {
			width = 1
		}
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: 1})
	})
}

// Render draws the stepper.
func (s *Stepper) Render(ctx runtime.RenderContext) {
	if s == nil {
		return
	}
	s.syncA11y()
	outer := s.bounds
	content := s.ContentBounds()
	if outer.Width <= 0 || outer.Height <= 0 {
		return
	}
	style := resolveBaseStyle(ctx, s, s.style, s.styleSet)
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
	ctx.Buffer.Fill(outer, ' ', style)
	if content.Width <= 0 || content.Height <= 0 {
		return
	}
	text = truncateString(text, content.Width)
	writePadded(ctx.Buffer, content.X, content.Y, content.Width, text, style)
}

// HandleMessage returns unhandled.
func (s *Stepper) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}

func (s *Stepper) syncA11y() {
	if s == nil {
		return
	}
	if s.Base.Role == "" {
		s.Base.Role = accessibility.RoleList
	}
	label := strings.TrimSpace(s.label)
	if label == "" {
		label = "Steps"
	}
	s.Base.Label = label
	s.Base.Description = fmt.Sprintf("%d steps", len(s.Steps))
	if active := s.activeStep(); active != "" {
		s.Base.Value = &accessibility.ValueInfo{Text: active}
	} else {
		s.Base.Value = nil
	}
}

func (s *Stepper) activeStep() string {
	for _, step := range s.Steps {
		if step.State == StepActive {
			return step.Title
		}
	}
	return ""
}
