package widgets

import (
	"image/color"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/gpu"
	"github.com/odvcencio/fluffyui/graphics"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

// GPUCanvasWidget draws using a GPU canvas and image protocols.
type GPUCanvasWidget struct {
	Component

	canvas  *gpu.GPUCanvas
	blitter *graphics.GPUBlitter
	encoder graphics.TerminalEncoder
	draw    func(canvas *gpu.GPUCanvas)

	cellWidth   int
	cellHeight  int
	pixelWidth  int
	pixelHeight int
	driver      gpu.Driver
	backend     gpu.Backend

	autoEncoder    graphics.TerminalEncoder
	encoderChecked bool
	fallbackCanvas *graphics.Canvas
	fallbackWidth  int
	fallbackHeight int
}

// GPUCanvasOption configures a GPUCanvasWidget.
type GPUCanvasOption = Option[GPUCanvasWidget]

// WithGPUCanvasEncoder sets a specific terminal encoder.
func WithGPUCanvasEncoder(encoder graphics.TerminalEncoder) GPUCanvasOption {
	return func(w *GPUCanvasWidget) {
		if w == nil {
			return
		}
		w.encoder = encoder
		w.blitter = nil
	}
}

// WithGPUCanvasBackend sets the GPU backend used by the canvas.
func WithGPUCanvasBackend(backend gpu.Backend) GPUCanvasOption {
	return func(w *GPUCanvasWidget) {
		if w == nil {
			return
		}
		w.backend = backend
		w.driver = nil
		w.canvas = nil
	}
}

// WithGPUCanvasDriver sets a specific driver instance.
func WithGPUCanvasDriver(driver gpu.Driver) GPUCanvasOption {
	return func(w *GPUCanvasWidget) {
		if w == nil {
			return
		}
		w.driver = driver
		w.canvas = nil
	}
}

// NewGPUCanvasWidget creates a GPUCanvasWidget with the draw callback.
func NewGPUCanvasWidget(draw func(canvas *gpu.GPUCanvas), opts ...GPUCanvasOption) *GPUCanvasWidget {
	widget := &GPUCanvasWidget{draw: draw, backend: gpu.BackendAuto}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(widget)
	}
	return widget
}

// SetEncoder sets a specific terminal encoder.
func (w *GPUCanvasWidget) SetEncoder(encoder graphics.TerminalEncoder) {
	if w == nil {
		return
	}
	w.encoder = encoder
	w.blitter = nil
	w.autoEncoder = nil
	w.encoderChecked = false
}

// SetBackend sets the GPU backend used by the canvas.
func (w *GPUCanvasWidget) SetBackend(backend gpu.Backend) {
	if w == nil {
		return
	}
	w.backend = backend
	w.driver = nil
	w.canvas = nil
}

// SetDriver sets a specific driver instance.
func (w *GPUCanvasWidget) SetDriver(driver gpu.Driver) {
	if w == nil {
		return
	}
	w.driver = driver
	w.canvas = nil
}

// Deprecated: prefer WithGPUCanvasEncoder during construction or SetEncoder for mutation.
func (w *GPUCanvasWidget) WithEncoder(encoder graphics.TerminalEncoder) *GPUCanvasWidget {
	w.SetEncoder(encoder)
	return w
}

// Deprecated: prefer WithGPUCanvasBackend during construction or SetBackend for mutation.
func (w *GPUCanvasWidget) WithBackend(backend gpu.Backend) *GPUCanvasWidget {
	w.SetBackend(backend)
	return w
}

// Deprecated: prefer WithGPUCanvasDriver during construction or SetDriver for mutation.
func (w *GPUCanvasWidget) WithDriver(driver gpu.Driver) *GPUCanvasWidget {
	w.SetDriver(driver)
	return w
}

// Measure returns the desired size for the widget.
func (w *GPUCanvasWidget) Measure(constraints runtime.Constraints) runtime.Size {
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
func (w *GPUCanvasWidget) Layout(bounds runtime.Rect) {
	w.Component.Layout(bounds)
	content := w.ContentBounds()
	if content.Width <= 0 || content.Height <= 0 {
		w.canvas = nil
		w.cellWidth = 0
		w.cellHeight = 0
		w.pixelWidth = 0
		w.pixelHeight = 0
		return
	}
	w.cellWidth = content.Width
	w.cellHeight = content.Height
}

// Unbind releases GPU resources.
func (w *GPUCanvasWidget) Unbind() {
	w.Component.Unbind()
	if w.canvas != nil {
		w.canvas.Dispose()
		w.canvas = nil
	}
	w.fallbackCanvas = nil
}

// Render draws the GPU canvas.
func (w *GPUCanvasWidget) Render(ctx runtime.RenderContext) {
	if w == nil || w.draw == nil {
		return
	}
	content := w.ContentBounds()
	if content.Width <= 0 || content.Height <= 0 {
		return
	}
	encoder := w.encoder
	if encoder == nil {
		if w.encoderChecked {
			encoder = w.autoEncoder
		} else {
			caps := terminal.DetectCapabilities()
			if caps.Kitty {
				encoder = graphics.KittyEncoder{}
			} else if caps.Sixel {
				encoder = graphics.SixelEncoder{}
			}
			w.autoEncoder = encoder
			w.encoderChecked = true
		}
	}
	if encoder == nil {
		w.renderBrailleFallback(ctx, content)
		return
	}
	cellW, cellH := encoderCellSize(encoder)
	pixelW := content.Width * cellW
	pixelH := content.Height * cellH
	if pixelW <= 0 || pixelH <= 0 {
		return
	}
	if w.canvas == nil {
		var (
			canvas *gpu.GPUCanvas
			err    error
		)
		switch {
		case w.driver != nil:
			canvas, err = gpu.NewGPUCanvasWithDriver(pixelW, pixelH, w.driver)
		case w.backend != gpu.BackendAuto:
			drv, derr := gpu.NewDriver(w.backend)
			if derr == nil {
				canvas, err = gpu.NewGPUCanvasWithDriver(pixelW, pixelH, drv)
			} else {
				err = derr
			}
		default:
			canvas, err = gpu.NewGPUCanvas(pixelW, pixelH)
		}
		if err != nil {
			return
		}
		w.canvas = canvas
	} else if pixelW != w.pixelWidth || pixelH != w.pixelHeight {
		_ = w.canvas.Resize(pixelW, pixelH)
	}
	w.pixelWidth = pixelW
	w.pixelHeight = pixelH
	if w.blitter == nil {
		w.blitter = graphics.NewGPUBlitter(w.canvas, encoder, cellW, cellH)
	} else {
		w.blitter.SetCanvas(w.canvas)
		w.blitter.SetEncoder(encoder)
		w.blitter.SetCellSize(cellW, cellH)
	}
	w.canvas.Begin()
	w.canvas.Clear(color.RGBA{})
	w.draw(w.canvas)
	img := w.blitter.Image()
	if img.Width > 0 && img.Height > 0 {
		if ctx.Buffer != nil {
			ctx.Buffer.SetImage(content.X, content.Y, img)
		}
	}
}

func (w *GPUCanvasWidget) renderBrailleFallback(ctx runtime.RenderContext, bounds runtime.Rect) {
	if w == nil {
		return
	}
	blitter := &graphics.BrailleBlitter{}
	pixelsPerCellW, pixelsPerCellH := blitter.PixelsPerCell()
	pixelW := bounds.Width * pixelsPerCellW
	pixelH := bounds.Height * pixelsPerCellH
	if pixelW <= 0 || pixelH <= 0 {
		return
	}
	if w.canvas == nil {
		var (
			canvas *gpu.GPUCanvas
			err    error
		)
		switch {
		case w.driver != nil:
			canvas, err = gpu.NewGPUCanvasWithDriver(pixelW, pixelH, w.driver)
		case w.backend != gpu.BackendAuto:
			drv, derr := gpu.NewDriver(w.backend)
			if derr == nil {
				canvas, err = gpu.NewGPUCanvasWithDriver(pixelW, pixelH, drv)
			} else {
				err = derr
			}
		default:
			canvas, err = gpu.NewGPUCanvas(pixelW, pixelH)
		}
		if err != nil {
			return
		}
		w.canvas = canvas
	} else if pixelW != w.pixelWidth || pixelH != w.pixelHeight {
		_ = w.canvas.Resize(pixelW, pixelH)
	}
	w.pixelWidth = pixelW
	w.pixelHeight = pixelH
	w.canvas.Begin()
	w.canvas.Clear(color.RGBA{})
	w.draw(w.canvas)
	pixels := w.canvas.End()
	if len(pixels) == 0 {
		return
	}
	if w.fallbackCanvas == nil || w.fallbackWidth != bounds.Width || w.fallbackHeight != bounds.Height {
		w.fallbackCanvas = graphics.NewCanvasWithBlitter(bounds.Width, bounds.Height, blitter)
		w.fallbackWidth = bounds.Width
		w.fallbackHeight = bounds.Height
	} else {
		w.fallbackCanvas.Clear()
	}
	fallback := w.fallbackCanvas
	dstW, dstH := fallback.Size()
	srcW := w.pixelWidth
	srcH := w.pixelHeight
	for y := 0; y < dstH; y++ {
		for x := 0; x < dstW; x++ {
			sx := x * srcW / dstW
			sy := y * srcH / dstH
			idx := (sy*srcW + sx) * 4
			if idx+3 >= len(pixels) {
				continue
			}
			a := pixels[idx+3]
			if a < 16 {
				continue
			}
			col := backend.ColorRGB(pixels[idx], pixels[idx+1], pixels[idx+2])
			fallback.SetPixel(x, y, col)
		}
	}
	fallback.Render(ctx.Buffer, bounds.X, bounds.Y)
}

func encoderCellSize(encoder graphics.TerminalEncoder) (int, int) {
	switch enc := encoder.(type) {
	case graphics.KittyEncoder:
		w := enc.CellWidth
		h := enc.CellHeight
		if w <= 0 {
			w = graphics.DefaultKittyCellWidth
		}
		if h <= 0 {
			h = graphics.DefaultKittyCellHeight
		}
		return w, h
	case graphics.SixelEncoder:
		w := enc.CellWidth
		h := enc.CellHeight
		if w <= 0 {
			w = graphics.DefaultSixelCellWidth
		}
		if h <= 0 {
			h = graphics.DefaultSixelCellHeight
		}
		return w, h
	case *graphics.KittyEncoder:
		if enc == nil {
			return graphics.DefaultKittyCellWidth, graphics.DefaultKittyCellHeight
		}
		w := enc.CellWidth
		h := enc.CellHeight
		if w <= 0 {
			w = graphics.DefaultKittyCellWidth
		}
		if h <= 0 {
			h = graphics.DefaultKittyCellHeight
		}
		return w, h
	case *graphics.SixelEncoder:
		if enc == nil {
			return graphics.DefaultSixelCellWidth, graphics.DefaultSixelCellHeight
		}
		w := enc.CellWidth
		h := enc.CellHeight
		if w <= 0 {
			w = graphics.DefaultSixelCellWidth
		}
		if h <= 0 {
			h = graphics.DefaultSixelCellHeight
		}
		return w, h
	default:
		return graphics.DefaultKittyCellWidth, graphics.DefaultKittyCellHeight
	}
}

var _ runtime.Widget = (*GPUCanvasWidget)(nil)
