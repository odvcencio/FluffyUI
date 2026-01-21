package widgets

import (
	"strings"

	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/runtime"
)

// Stack overlays child widgets.
type Stack struct {
	Base
	Children []runtime.Widget
	label    string
}

// NewStack creates a stack container.
func NewStack(children ...runtime.Widget) *Stack {
	stack := &Stack{Children: children, label: "Stack"}
	stack.Base.Role = accessibility.RoleGroup
	stack.syncA11y()
	return stack
}

// SetLabel updates the accessibility label.
func (s *Stack) SetLabel(label string) {
	if s == nil {
		return
	}
	s.label = label
	s.syncA11y()
}

// Measure returns the max size of children.
func (s *Stack) Measure(constraints runtime.Constraints) runtime.Size {
	maxSize := runtime.Size{}
	for _, child := range s.Children {
		if child == nil {
			continue
		}
		size := child.Measure(constraints)
		if size.Width > maxSize.Width {
			maxSize.Width = size.Width
		}
		if size.Height > maxSize.Height {
			maxSize.Height = size.Height
		}
	}
	return constraints.Constrain(maxSize)
}

// Layout assigns bounds to children.
func (s *Stack) Layout(bounds runtime.Rect) {
	s.Base.Layout(bounds)
	for _, child := range s.Children {
		if child != nil {
			child.Layout(bounds)
		}
	}
}

// Render draws children in order.
func (s *Stack) Render(ctx runtime.RenderContext) {
	s.syncA11y()
	for _, child := range s.Children {
		if child != nil {
			child.Render(ctx)
		}
	}
}

// HandleMessage forwards messages to children from top to bottom.
func (s *Stack) HandleMessage(msg runtime.Message) runtime.HandleResult {
	for i := len(s.Children) - 1; i >= 0; i-- {
		child := s.Children[i]
		if child == nil {
			continue
		}
		if result := child.HandleMessage(msg); result.Handled {
			return result
		}
	}
	return runtime.Unhandled()
}

// ChildWidgets returns stacked children.
func (s *Stack) ChildWidgets() []runtime.Widget {
	if s == nil {
		return nil
	}
	return s.Children
}

func (s *Stack) syncA11y() {
	if s == nil {
		return
	}
	if s.Base.Role == "" {
		s.Base.Role = accessibility.RoleGroup
	}
	label := strings.TrimSpace(s.label)
	if label == "" {
		label = "Stack"
	}
	s.Base.Label = label
}
