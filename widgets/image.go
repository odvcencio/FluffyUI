package widgets

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"m31labs.dev/fluffyui/accessibility"
	"m31labs.dev/fluffyui/graphics"
	"m31labs.dev/fluffyui/runtime"
)

// ImageFit controls how images are sized within their bounds.
type ImageFit int

const (
	// ImageFitContain fits the image within bounds, preserving aspect ratio.
	ImageFitContain ImageFit = iota
	// ImageFitCover fills the bounds, cropping any excess.
	ImageFitCover
	// ImageFitFill stretches the image to fill the bounds exactly.
	ImageFitFill
	// ImageFitScaleDown behaves like Contain but never upscales.
	ImageFitScaleDown
)

// ImageProtocol selects the terminal image protocol.
type ImageProtocol int

const (
	// ImageProtocolAuto detects the best available protocol.
	ImageProtocolAuto ImageProtocol = iota
	// ImageProtocolKitty uses the Kitty image protocol.
	ImageProtocolKitty
	// ImageProtocolSixel uses the Sixel graphics protocol.
	ImageProtocolSixel
	// ImageProtocolITerm2 uses the iTerm2 inline image protocol.
	ImageProtocolITerm2
	// ImageProtocolHalfBlock uses Unicode half-block character fallback.
	ImageProtocolHalfBlock
)

// Image displays an image in the terminal.
type Image struct {
	Base

	source   image.Image
	fitMode  ImageFit
	protocol ImageProtocol
	label    string
	services runtime.Services

	canvas     *graphics.Canvas
	cellWidth  int
	cellHeight int
}

// NewImage creates an Image widget from the given Go image.
func NewImage(src image.Image) *Image {
	w := &Image{
		source:   src,
		fitMode:  ImageFitContain,
		protocol: ImageProtocolAuto,
		label:    "Image",
	}
	w.syncA11y()
	return w
}

// NewImageFromFile creates an Image widget by loading an image from a file.
func NewImageFromFile(path string) (*Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("image: open %s: %w", path, err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("image: decode %s: %w", path, err)
	}

	w := NewImage(img)
	return w, nil
}

// SetFit sets the image fit mode.
func (w *Image) SetFit(fit ImageFit) {
	if w == nil {
		return
	}
	w.fitMode = fit
}

// Fit returns the current fit mode.
func (w *Image) Fit() ImageFit {
	if w == nil {
		return ImageFitContain
	}
	return w.fitMode
}

// SetProtocol sets the image rendering protocol.
func (w *Image) SetProtocol(proto ImageProtocol) {
	if w == nil {
		return
	}
	w.protocol = proto
}

// Protocol returns the current protocol setting.
func (w *Image) Protocol() ImageProtocol {
	if w == nil {
		return ImageProtocolAuto
	}
	return w.protocol
}

// SetImage replaces the displayed image.
func (w *Image) SetImage(img image.Image) {
	if w == nil {
		return
	}
	w.source = img
	// Reset the canvas so it gets rebuilt on next render.
	w.canvas = nil
}

// SetLabel sets the accessible label for the image.
func (w *Image) SetLabel(label string) {
	if w == nil {
		return
	}
	w.label = label
	w.syncA11y()
}

// Label returns the accessible label.
func (w *Image) Label() string {
	if w == nil {
		return ""
	}
	return w.label
}

// StyleType returns the selector type name.
func (w *Image) StyleType() string {
	return "Image"
}

// Measure returns the desired size based on the image aspect ratio and
// constraints. Each terminal cell is approximately 2:1 (twice as tall as
// wide), so for half-block rendering one cell row represents 2 image pixels
// vertically and 1 pixel horizontally.
func (w *Image) Measure(constraints runtime.Constraints) runtime.Size {
	return w.measureWithStyle(constraints, func(cc runtime.Constraints) runtime.Size {
		if w.source == nil {
			return cc.Constrain(runtime.Size{Width: 1, Height: 1})
		}
		bounds := w.source.Bounds()
		imgW := bounds.Dx()
		imgH := bounds.Dy()
		if imgW <= 0 || imgH <= 0 {
			return cc.Constrain(runtime.Size{Width: 1, Height: 1})
		}

		// Convert image pixels to cell dimensions.
		// Each cell is 1 pixel wide and 2 pixels tall (half-block).
		cellW := imgW
		cellH := (imgH + 1) / 2

		desiredW, desiredH := w.fitDimensions(cellW, cellH, cc.MaxWidth, cc.MaxHeight)
		return cc.Constrain(runtime.Size{Width: desiredW, Height: desiredH})
	})
}

// fitDimensions computes the cell dimensions according to the fit mode.
func (w *Image) fitDimensions(cellW, cellH, maxW, maxH int) (int, int) {
	if maxW <= 0 {
		maxW = cellW
	}
	if maxH <= 0 {
		maxH = cellH
	}

	switch w.fitMode {
	case ImageFitFill:
		return maxW, maxH

	case ImageFitCover:
		// Scale up to cover the entire area while preserving aspect ratio.
		scaleW := float64(maxW) / float64(cellW)
		scaleH := float64(maxH) / float64(cellH)
		scale := scaleW
		if scaleH > scale {
			scale = scaleH
		}
		dw := int(float64(cellW)*scale + 0.5)
		dh := int(float64(cellH)*scale + 0.5)
		if dw < 1 {
			dw = 1
		}
		if dh < 1 {
			dh = 1
		}
		return dw, dh

	case ImageFitScaleDown:
		// Like contain but never upscale.
		if cellW <= maxW && cellH <= maxH {
			return cellW, cellH
		}
		return containFit(cellW, cellH, maxW, maxH)

	default: // ImageFitContain
		return containFit(cellW, cellH, maxW, maxH)
	}
}

// containFit scales the image to fit within maxW x maxH preserving aspect ratio.
func containFit(cellW, cellH, maxW, maxH int) (int, int) {
	if cellW <= maxW && cellH <= maxH {
		return cellW, cellH
	}
	scaleW := float64(maxW) / float64(cellW)
	scaleH := float64(maxH) / float64(cellH)
	scale := scaleW
	if scaleH < scale {
		scale = scaleH
	}
	dw := int(float64(cellW)*scale + 0.5)
	dh := int(float64(cellH)*scale + 0.5)
	if dw < 1 {
		dw = 1
	}
	if dh < 1 {
		dh = 1
	}
	return dw, dh
}

// Layout stores the assigned bounds and prepares the canvas.
func (w *Image) Layout(bounds runtime.Rect) {
	if w == nil {
		return
	}
	w.Base.Layout(bounds)
	content := w.ContentBounds()
	if content.Width <= 0 || content.Height <= 0 {
		w.canvas = nil
		w.cellWidth = 0
		w.cellHeight = 0
		return
	}
	if w.canvas == nil || content.Width != w.cellWidth || content.Height != w.cellHeight {
		w.canvas = graphics.NewCanvasWithBlitter(content.Width, content.Height, &graphics.HalfBlockBlitter{})
		w.cellWidth = content.Width
		w.cellHeight = content.Height
	}
}

// Render draws the image to the terminal using the half-block fallback.
func (w *Image) Render(ctx runtime.RenderContext) {
	if w == nil {
		return
	}
	w.syncA11y()
	content := w.ContentBounds()
	if content.Width <= 0 || content.Height <= 0 {
		return
	}
	if w.source == nil || w.canvas == nil {
		return
	}

	w.canvas.Clear()

	// Get pixel dimensions of the canvas.
	pixW, pixH := w.canvas.Size()
	if pixW <= 0 || pixH <= 0 {
		return
	}

	w.canvas.DrawImageScaled(0, 0, pixW, pixH, w.source)
	w.canvas.Render(ctx.Buffer, content.X, content.Y)
}

// HandleMessage returns unhandled (non-interactive widget).
func (w *Image) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}

// Bind attaches app services.
func (w *Image) Bind(services runtime.Services) {
	if w == nil {
		return
	}
	w.services = services
}

// Unbind releases app services.
func (w *Image) Unbind() {
	if w == nil {
		return
	}
	w.services = runtime.Services{}
	w.canvas = nil
}

func (w *Image) syncA11y() {
	if w == nil {
		return
	}
	if w.Base.Role == "" {
		w.Base.Role = accessibility.RoleImage
	}
	if w.label != "" {
		w.Base.Base.Label = w.label
	} else {
		w.Base.Base.Label = "Image"
	}
}

var _ runtime.Widget = (*Image)(nil)
var _ runtime.Bindable = (*Image)(nil)
var _ runtime.Unbindable = (*Image)(nil)
