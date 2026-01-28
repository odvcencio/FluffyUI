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
}

// NewGPUCanvasWidget creates a GPUCanvasWidget with the draw callback.
func NewGPUCanvasWidget(draw func(canvas *gpu.GPUCanvas)) *GPUCanvasWidget {
	return &GPUCanvasWidget{draw: draw, backend: gpu.BackendAuto}
}

// WithEncoder sets a specific terminal encoder.
func (w *GPUCanvasWidget) WithEncoder(encoder graphics.TerminalEncoder) *GPUCanvasWidget {
	if w == nil {
		return w
	}
	w.encoder = encoder
	w.blitter = nil
	return w
}

// WithBackend sets the GPU backend used by the canvas.
func (w *GPUCanvasWidget) WithBackend(backend gpu.Backend) *GPUCanvasWidget {
	if w == nil {
		return w
	}
	w.backend = backend
	w.driver = nil
	w.canvas = nil
	return w
}

// WithDriver sets a specific driver instance.
func (w *GPUCanvasWidget) WithDriver(driver gpu.Driver) *GPUCanvasWidget {
	if w == nil {
		return w
	}
	w.driver = driver
	w.canvas = nil
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
		caps := terminal.DetectCapabilities()
		if caps.Kitty {
			encoder = graphics.KittyEncoder{}
		} else if caps.Sixel {
			encoder = graphics.SixelEncoder{}
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
	fallback := graphics.NewCanvasWithBlitter(bounds.Width, bounds.Height, blitter)
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
