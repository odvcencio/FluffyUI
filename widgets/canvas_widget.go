package widgets

import (
	"github.com/odvcencio/fluffyui/graphics"
	"github.com/odvcencio/fluffyui/runtime"
)

// CanvasWidget is a widget that draws using a Canvas.
type CanvasWidget struct {
	Component

	canvas     *graphics.Canvas
	blitter    graphics.Blitter
	draw       func(canvas *graphics.Canvas)
	cellWidth  int
	cellHeight int
}

// CanvasOption configures a CanvasWidget.
type CanvasOption = Option[CanvasWidget]

// WithCanvasBlitter sets the blitter used to render pixels to cells.
func WithCanvasBlitter(blitter graphics.Blitter) CanvasOption {
	return func(w *CanvasWidget) {
		if w == nil || blitter == nil {
			return
		}
		w.blitter = blitter
		w.canvas = nil
	}
}

// NewCanvasWidget creates a CanvasWidget with the draw callback.
func NewCanvasWidget(draw func(canvas *graphics.Canvas), opts ...CanvasOption) *CanvasWidget {
	widget := &CanvasWidget{
		blitter: &graphics.SextantBlitter{},
		draw:    draw,
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(widget)
	}
	return widget
}

// SetBlitter sets the blitter used to render pixels to cells.
func (w *CanvasWidget) SetBlitter(blitter graphics.Blitter) {
	if w == nil || blitter == nil {
		return
	}
	w.blitter = blitter
	w.canvas = nil
}

// Deprecated: prefer WithCanvasBlitter during construction or SetBlitter for mutation.
func (w *CanvasWidget) WithBlitter(blitter graphics.Blitter) *CanvasWidget {
	w.SetBlitter(blitter)
	return w
}

// Measure returns the desired size for the canvas widget.
func (w *CanvasWidget) Measure(constraints runtime.Constraints) runtime.Size {
	return w.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		width := contentConstraints.MaxWidth
		height := contentConstraints.MaxHeight
		if width == 0 {
			width = contentConstraints.MinWidth
		}
		if height == 0 {
			height = contentConstraints.MinHeight
		}
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: height})
	})
}

// Layout updates layout bounds and canvas size.
func (w *CanvasWidget) Layout(bounds runtime.Rect) {
	w.Component.Layout(bounds)
	content := w.ContentBounds()
	if content.Width <= 0 || content.Height <= 0 {
		w.canvas = nil
		w.cellWidth = 0
		w.cellHeight = 0
		return
	}
	if w.canvas == nil || content.Width != w.cellWidth || content.Height != w.cellHeight {
		w.canvas = graphics.NewCanvasWithBlitter(content.Width, content.Height, w.blitter)
		w.cellWidth = content.Width
		w.cellHeight = content.Height
	}
}

// Render draws the canvas.
func (w *CanvasWidget) Render(ctx runtime.RenderContext) {
	if w == nil || w.canvas == nil || w.draw == nil {
		return
	}
	content := w.ContentBounds()
	if content.Width <= 0 || content.Height <= 0 {
		return
	}
	w.canvas.Clear()
	w.draw(w.canvas)
	w.canvas.Render(ctx.Buffer, content.X, content.Y)
}

var _ runtime.Widget = (*CanvasWidget)(nil)
