package demo

import (
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/widgets"
)

// VBox lays out children vertically using their measured heights.
type VBox struct {
	widgets.Base
	Children []runtime.Widget
	Gap      int
}

// NewVBox creates a vertical box.
func NewVBox(children ...runtime.Widget) *VBox {
	return &VBox{Children: children}
}

// Measure returns the size needed for the children.
func (v *VBox) Measure(constraints runtime.Constraints) runtime.Size {
	width := 0
	height := 0
	for i, child := range v.Children {
		if child == nil {
			continue
		}
		size := child.Measure(runtime.Constraints{MaxWidth: constraints.MaxWidth, MaxHeight: constraints.MaxHeight})
		if size.Width > width {
			width = size.Width
		}
		height += size.Height
		if i > 0 {
			height += v.Gap
		}
	}
	return constraints.Constrain(runtime.Size{Width: width, Height: height})
}

// Layout assigns bounds to children.
func (v *VBox) Layout(bounds runtime.Rect) {
	v.Base.Layout(bounds)
	y := bounds.Y
	for _, child := range v.Children {
		if child == nil {
			continue
		}
		size := child.Measure(runtime.Constraints{MaxWidth: bounds.Width, MaxHeight: bounds.Height})
		height := size.Height
		if height < 1 {
			height = 1
		}
		if y+height > bounds.Y+bounds.Height {
			height = bounds.Y + bounds.Height - y
		}
		if height <= 0 {
			return
		}
		child.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: height})
		y += height + v.Gap
	}
}

// Render draws children.
func (v *VBox) Render(ctx runtime.RenderContext) {
	for _, child := range v.Children {
		if child != nil {
			child.Render(ctx)
		}
	}
}

// HandleMessage forwards messages to children.
func (v *VBox) HandleMessage(msg runtime.Message) runtime.HandleResult {
	for _, child := range v.Children {
		if child == nil {
			continue
		}
		if result := child.HandleMessage(msg); result.Handled {
			return result
		}
	}
	return runtime.Unhandled()
}

// ChildWidgets returns the children.
func (v *VBox) ChildWidgets() []runtime.Widget {
	if v == nil {
		return nil
	}
	return v.Children
}

// HBox lays out children horizontally using their measured widths.
type HBox struct {
	widgets.Base
	Children []runtime.Widget
	Gap      int
}

// NewHBox creates a horizontal box.
func NewHBox(children ...runtime.Widget) *HBox {
	return &HBox{Children: children}
}

// Measure returns the size needed for the children.
func (h *HBox) Measure(constraints runtime.Constraints) runtime.Size {
	width := 0
	height := 0
	for i, child := range h.Children {
		if child == nil {
			continue
		}
		size := child.Measure(runtime.Constraints{MaxWidth: constraints.MaxWidth, MaxHeight: constraints.MaxHeight})
		width += size.Width
		if i > 0 {
			width += h.Gap
		}
		if size.Height > height {
			height = size.Height
		}
	}
	return constraints.Constrain(runtime.Size{Width: width, Height: height})
}

// Layout assigns bounds to children.
func (h *HBox) Layout(bounds runtime.Rect) {
	h.Base.Layout(bounds)
	x := bounds.X
	for _, child := range h.Children {
		if child == nil {
			continue
		}
		size := child.Measure(runtime.Constraints{MaxWidth: bounds.Width, MaxHeight: bounds.Height})
		width := size.Width
		if width < 1 {
			width = 1
		}
		if x+width > bounds.X+bounds.Width {
			width = bounds.X + bounds.Width - x
		}
		if width <= 0 {
			return
		}
		child.Layout(runtime.Rect{X: x, Y: bounds.Y, Width: width, Height: bounds.Height})
		x += width + h.Gap
	}
}

// Render draws children.
func (h *HBox) Render(ctx runtime.RenderContext) {
	for _, child := range h.Children {
		if child != nil {
			child.Render(ctx)
		}
	}
}

// HandleMessage forwards messages to children.
func (h *HBox) HandleMessage(msg runtime.Message) runtime.HandleResult {
	for _, child := range h.Children {
		if child == nil {
			continue
		}
		if result := child.HandleMessage(msg); result.Handled {
			return result
		}
	}
	return runtime.Unhandled()
}

// ChildWidgets returns the children.
func (h *HBox) ChildWidgets() []runtime.Widget {
	if h == nil {
		return nil
	}
	return h.Children
}
