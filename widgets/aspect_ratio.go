package widgets

import (
	"math"
	"strings"

	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/runtime"
)

// AspectRatio constrains a child to a fixed width/height ratio.
type AspectRatio struct {
	Base
	child runtime.Widget
	ratio float64
	label string
}

// NewAspectRatio creates an aspect ratio container.
func NewAspectRatio(child runtime.Widget, ratio float64) *AspectRatio {
	w := &AspectRatio{
		child: child,
		ratio: ratio,
		label: "Aspect Ratio",
	}
	w.Base.Role = accessibility.RoleGroup
	w.syncA11y()
	return w
}

// SetRatio updates the aspect ratio (width / height).
func (a *AspectRatio) SetRatio(ratio float64) {
	if a == nil {
		return
	}
	a.ratio = ratio
}

// SetLabel updates the accessibility label.
func (a *AspectRatio) SetLabel(label string) {
	if a == nil {
		return
	}
	a.label = label
	a.syncA11y()
}

// Measure returns the constrained size preserving the ratio.
func (a *AspectRatio) Measure(constraints runtime.Constraints) runtime.Size {
	return a.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		if a != nil && a.child != nil {
			_ = a.child.Measure(contentConstraints)
		}
		ratio := a.ratio
		if ratio <= 0 {
			ratio = 1
		}
		return fitAspectSize(contentConstraints, ratio)
	})
}

// Layout assigns bounds and centers the child within the ratio box.
func (a *AspectRatio) Layout(bounds runtime.Rect) {
	a.Base.Layout(bounds)
	if a == nil || a.child == nil {
		return
	}
	content := a.ContentBounds()
	if content.Width <= 0 || content.Height <= 0 {
		return
	}
	ratio := a.ratio
	if ratio <= 0 {
		ratio = 1
	}
	childBounds := fitAspectRect(content, ratio)
	a.child.Layout(childBounds)
}

// Render draws the child.
func (a *AspectRatio) Render(ctx runtime.RenderContext) {
	a.syncA11y()
	runtime.RenderChild(ctx, a.child)
}

// HandleMessage forwards messages to the child.
func (a *AspectRatio) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if a == nil || a.child == nil {
		return runtime.Unhandled()
	}
	return a.child.HandleMessage(msg)
}

// ChildWidgets returns the child widget.
func (a *AspectRatio) ChildWidgets() []runtime.Widget {
	if a == nil || a.child == nil {
		return nil
	}
	return []runtime.Widget{a.child}
}

func (a *AspectRatio) syncA11y() {
	if a == nil {
		return
	}
	if a.Base.Role == "" {
		a.Base.Role = accessibility.RoleGroup
	}
	label := strings.TrimSpace(a.label)
	if label == "" {
		label = "Aspect Ratio"
	}
	a.Base.Label = label
}

func fitAspectSize(constraints runtime.Constraints, ratio float64) runtime.Size {
	maxW := constraints.MaxWidth
	maxH := constraints.MaxHeight
	if maxW == layoutMaxInt && maxH == layoutMaxInt {
		return constraints.MinSize()
	}
	if maxW == layoutMaxInt {
		height := maxH
		width := int(math.Round(float64(height) * ratio))
		return constraints.Constrain(runtime.Size{Width: width, Height: height})
	}
	if maxH == layoutMaxInt {
		width := maxW
		height := int(math.Round(float64(width) / ratio))
		return constraints.Constrain(runtime.Size{Width: width, Height: height})
	}
	width := maxW
	height := int(math.Round(float64(width) / ratio))
	if height > maxH {
		height = maxH
		width = int(math.Round(float64(height) * ratio))
	}
	return constraints.Constrain(runtime.Size{Width: width, Height: height})
}

func fitAspectRect(bounds runtime.Rect, ratio float64) runtime.Rect {
	width := bounds.Width
	height := int(math.Round(float64(width) / ratio))
	if height > bounds.Height {
		height = bounds.Height
		width = int(math.Round(float64(height) * ratio))
	}
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	x := bounds.X + (bounds.Width-width)/2
	y := bounds.Y + (bounds.Height-height)/2
	return runtime.Rect{X: x, Y: y, Width: width, Height: height}
}

var _ runtime.Widget = (*AspectRatio)(nil)
var _ runtime.ChildProvider = (*AspectRatio)(nil)
