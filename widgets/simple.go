package widgets

import "github.com/odvcencio/fluffy-ui/runtime"

// SimpleWidget provides function hooks for quick widgets with Base styling.
type SimpleWidget struct {
	Base

	MeasureFunc       func(runtime.Constraints) runtime.Size
	LayoutFunc        func(runtime.Rect)
	RenderFunc        func(runtime.RenderContext)
	HandleMessageFunc func(runtime.Message) runtime.HandleResult
}

// NewSimpleWidget creates a SimpleWidget.
func NewSimpleWidget() *SimpleWidget {
	return &SimpleWidget{}
}

func (s *SimpleWidget) Measure(constraints runtime.Constraints) runtime.Size {
	return s.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		if s == nil || s.MeasureFunc == nil {
			return contentConstraints.MinSize()
		}
		return s.MeasureFunc(contentConstraints)
	})
}

func (s *SimpleWidget) Layout(bounds runtime.Rect) {
	if s == nil {
		return
	}
	s.Base.Layout(bounds)
	if s.LayoutFunc != nil {
		s.LayoutFunc(s.Bounds())
	}
}

func (s *SimpleWidget) Render(ctx runtime.RenderContext) {
	if s == nil || s.RenderFunc == nil {
		return
	}
	s.RenderFunc(ctx)
}

func (s *SimpleWidget) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if s == nil || s.HandleMessageFunc == nil {
		return runtime.Unhandled()
	}
	return s.HandleMessageFunc(msg)
}

var _ runtime.Widget = (*SimpleWidget)(nil)
