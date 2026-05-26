package widgets

import (
	"fmt"
	"strings"

	"m31labs.dev/fluffyui/accessibility"
	"m31labs.dev/fluffyui/runtime"
)

// SplitterOrientation describes the split direction.
type SplitterOrientation int

const (
	SplitHorizontal SplitterOrientation = iota // Left/right
	SplitVertical                              // Top/bottom
)

// Splitter divides space between two panes.
type Splitter struct {
	Base
	First       runtime.Widget
	Second      runtime.Widget
	Orientation SplitterOrientation
	Ratio       float64
	DividerSize int
	label       string
	services    runtime.Services
}

// NewSplitter creates a splitter with two panes.
func NewSplitter(first, second runtime.Widget) *Splitter {
	s := &Splitter{
		First:       first,
		Second:      second,
		Orientation: SplitHorizontal,
		Ratio:       0.5,
		DividerSize: 1,
		label:       "Splitter",
	}
	s.Base.Role = accessibility.RoleGroup
	s.syncA11y()
	return s
}

// SetLabel updates the accessibility label.
func (s *Splitter) SetLabel(label string) {
	if s == nil {
		return
	}
	s.label = label
	s.syncA11y()
}

// Measure returns the max child size.
func (s *Splitter) Measure(constraints runtime.Constraints) runtime.Size {
	return s.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		size := contentConstraints.MinSize()
		if s.First != nil {
			child := s.First.Measure(contentConstraints)
			if child.Width > size.Width {
				size.Width = child.Width
			}
			if child.Height > size.Height {
				size.Height = child.Height
			}
		}
		if s.Second != nil {
			child := s.Second.Measure(contentConstraints)
			if child.Width > size.Width {
				size.Width = child.Width
			}
			if child.Height > size.Height {
				size.Height = child.Height
			}
		}
		return contentConstraints.Constrain(size)
	})
}

// Layout positions the panes.
func (s *Splitter) Layout(bounds runtime.Rect) {
	s.Base.Layout(bounds)
	content := s.ContentBounds()
	if s.Ratio <= 0 {
		s.Ratio = 0.5
	}
	if s.Ratio >= 1 {
		s.Ratio = 0.5
	}
	divider := s.DividerSize
	if divider < 0 {
		divider = 0
	}
	if s.Orientation == SplitHorizontal {
		width := content.Width - divider
		if width < 0 {
			width = 0
		}
		firstWidth := int(float64(width) * s.Ratio)
		secondWidth := width - firstWidth
		if s.First != nil {
			s.First.Layout(runtime.Rect{X: content.X, Y: content.Y, Width: firstWidth, Height: content.Height})
		}
		if s.Second != nil {
			s.Second.Layout(runtime.Rect{
				X:      content.X + firstWidth + divider,
				Y:      content.Y,
				Width:  secondWidth,
				Height: content.Height,
			})
		}
		return
	}
	height := content.Height - divider
	if height < 0 {
		height = 0
	}
	firstHeight := int(float64(height) * s.Ratio)
	secondHeight := height - firstHeight
	if s.First != nil {
		s.First.Layout(runtime.Rect{X: content.X, Y: content.Y, Width: content.Width, Height: firstHeight})
	}
	if s.Second != nil {
		s.Second.Layout(runtime.Rect{
			X:      content.X,
			Y:      content.Y + firstHeight + divider,
			Width:  content.Width,
			Height: secondHeight,
		})
	}
}

// Render draws both panes.
func (s *Splitter) Render(ctx runtime.RenderContext) {
	s.syncA11y()
	runtime.RenderChild(ctx, s.First)
	runtime.RenderChild(ctx, s.Second)
}

// RenderHTML renders the splitter as a static HTML flexbox layout.
func (s *Splitter) RenderHTML(ctx runtime.HTMLContext) runtime.HTML {
	dir := "row"
	if s.Orientation == SplitVertical {
		dir = "column"
	}
	firstPct := int(s.Ratio * 100)

	var b strings.Builder
	fmt.Fprintf(&b, `<div class="fluffy-Splitter" style="display:flex;flex-direction:%s;height:100%%">`, dir)
	childCtx := ctx.Child()

	// First pane
	fmt.Fprintf(&b, `<div class="fluffy-Splitter-first" style="flex:0 0 %d%%;overflow-y:auto">`, firstPct)
	if hr, ok := s.First.(runtime.HTMLRenderer); ok {
		b.WriteString(string(hr.RenderHTML(childCtx)))
	}
	b.WriteString("</div>")

	// Second pane
	b.WriteString(`<div class="fluffy-Splitter-second" style="flex:1 1 auto;overflow-y:auto">`)
	if hr, ok := s.Second.(runtime.HTMLRenderer); ok {
		b.WriteString(string(hr.RenderHTML(childCtx)))
	}
	b.WriteString("</div>")

	b.WriteString("</div>")
	return runtime.HTML(b.String())
}

// HandleMessage forwards messages to child panes.
func (s *Splitter) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if s.First != nil {
		if result := s.First.HandleMessage(msg); result.Handled {
			return result
		}
	}
	if s.Second != nil {
		if result := s.Second.HandleMessage(msg); result.Handled {
			return result
		}
	}
	return runtime.Unhandled()
}

// ChildWidgets returns the panes.
func (s *Splitter) ChildWidgets() []runtime.Widget {
	if s == nil {
		return nil
	}
	children := make([]runtime.Widget, 0, 2)
	if s.First != nil {
		children = append(children, s.First)
	}
	if s.Second != nil {
		children = append(children, s.Second)
	}
	return children
}

// PathSegment returns a debug path segment for the given child.
func (s *Splitter) PathSegment(child runtime.Widget) string {
	if s == nil {
		return "Splitter"
	}
	switch child {
	case s.First:
		return "Splitter[first]"
	case s.Second:
		return "Splitter[second]"
	default:
		return "Splitter"
	}
}

func (s *Splitter) syncA11y() {
	if s == nil {
		return
	}
	if s.Base.Role == "" {
		s.Base.Role = accessibility.RoleGroup
	}
	label := strings.TrimSpace(s.label)
	if label == "" {
		label = "Splitter"
	}
	s.Base.Label = label
	if s.Orientation == SplitVertical {
		s.Base.Description = "vertical split"
		s.Base.Orientation = "vertical"
	} else {
		s.Base.Description = "horizontal split"
		s.Base.Orientation = "horizontal"
	}
}

// Bind attaches app services.
func (s *Splitter) Bind(services runtime.Services) {
	if s == nil {
		return
	}
	s.services = services
}

// Unbind releases app services.
func (s *Splitter) Unbind() {
	if s == nil {
		return
	}
	s.services = runtime.Services{}
}

var _ runtime.Widget = (*Splitter)(nil)
var _ runtime.Bindable = (*Splitter)(nil)
var _ runtime.Unbindable = (*Splitter)(nil)
