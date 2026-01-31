package widgets

import (
	"context"
	"fmt"
	"image"
	"sync"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/graphics"
	"github.com/odvcencio/fluffyui/runtime"
)

// AsyncImage loads an image asynchronously and renders it when ready.
type AsyncImage struct {
	Component
	loader      func() (image.Image, error)
	placeholder runtime.Widget
	blitter     graphics.Blitter
	scaleMode   graphics.ScaleMode
	scaleToFit  bool
	center      bool

	canvas     *graphics.Canvas
	cellWidth  int
	cellHeight int

	mu       sync.Mutex
	img      image.Image
	err      error
	loading  bool
	loadOnce sync.Once
}

// AsyncImageOption configures an async image widget.
type AsyncImageOption = Option[AsyncImage]

// NewAsyncImage loads an image from disk asynchronously.
func NewAsyncImage(path string, opts ...AsyncImageOption) *AsyncImage {
	return NewAsyncImageWithLoader(func() (image.Image, error) {
		return graphics.LoadImage(path)
	}, opts...)
}

// NewAsyncImageWithLoader loads an image using a custom loader.
func NewAsyncImageWithLoader(loader func() (image.Image, error), opts ...AsyncImageOption) *AsyncImage {
	w := &AsyncImage{
		loader:     loader,
		blitter:    &graphics.SextantBlitter{},
		scaleMode:  graphics.ScaleNearest,
		scaleToFit: true,
		center:     true,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(w)
		}
	}
	return w
}

// WithAsyncImagePlaceholder sets a placeholder widget for loading/error states.
func WithAsyncImagePlaceholder(widget runtime.Widget) AsyncImageOption {
	return func(w *AsyncImage) {
		if w == nil {
			return
		}
		w.placeholder = widget
	}
}

// WithAsyncImageBlitter overrides the blitter used for rendering.
func WithAsyncImageBlitter(blitter graphics.Blitter) AsyncImageOption {
	return func(w *AsyncImage) {
		if w == nil || blitter == nil {
			return
		}
		w.blitter = blitter
		w.canvas = nil
	}
}

// WithAsyncImageScaleMode sets the scaling interpolation mode.
func WithAsyncImageScaleMode(mode graphics.ScaleMode) AsyncImageOption {
	return func(w *AsyncImage) {
		if w == nil {
			return
		}
		w.scaleMode = mode
	}
}

// WithAsyncImageScaleToFit toggles scaling to fit the widget bounds.
func WithAsyncImageScaleToFit(enabled bool) AsyncImageOption {
	return func(w *AsyncImage) {
		if w == nil {
			return
		}
		w.scaleToFit = enabled
	}
}

// WithAsyncImageCenter toggles centering within the widget bounds.
func WithAsyncImageCenter(enabled bool) AsyncImageOption {
	return func(w *AsyncImage) {
		if w == nil {
			return
		}
		w.center = enabled
	}
}

// Bind attaches services and starts loading.
func (w *AsyncImage) Bind(services runtime.Services) {
	w.Component.Bind(services)
	w.startLoad()
}

// Unbind releases subscriptions.
func (w *AsyncImage) Unbind() {
	w.Component.Unbind()
}

// Measure returns desired size for the image.
func (w *AsyncImage) Measure(constraints runtime.Constraints) runtime.Size {
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

// Layout updates layout bounds and placeholder layout.
func (w *AsyncImage) Layout(bounds runtime.Rect) {
	w.Component.Layout(bounds)
	content := w.ContentBounds()
	if w.placeholder != nil {
		w.placeholder.Layout(content)
	}
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

// Render draws the image or placeholder.
func (w *AsyncImage) Render(ctx runtime.RenderContext) {
	if w == nil {
		return
	}
	content := w.ContentBounds()
	if content.Width <= 0 || content.Height <= 0 {
		return
	}
	img, err, loading := w.imageState()
	if img == nil {
		if w.placeholder != nil {
			runtime.RenderChild(ctx, w.placeholder)
			return
		}
		message := "Loading..."
		if err != nil {
			message = "Image error"
		}
		if loading {
			message = "Loading..."
		}
		ctx.Buffer.SetString(content.X, content.Y, clipString(message, content.Width), backend.DefaultStyle())
		return
	}
	if w.canvas == nil {
		return
	}
	w.canvas.Clear()
	pixelW, pixelH := w.canvas.Size()
	if pixelW <= 0 || pixelH <= 0 {
		return
	}
	drawImg := img
	if w.scaleToFit {
		drawImg = graphics.ScaleImageMode(img, pixelW, pixelH, w.scaleMode)
	}
	offsetX := 0
	offsetY := 0
	if w.center && drawImg != nil {
		bounds := drawImg.Bounds()
		if bounds.Dx() < pixelW {
			offsetX = (pixelW - bounds.Dx()) / 2
		}
		if bounds.Dy() < pixelH {
			offsetY = (pixelH - bounds.Dy()) / 2
		}
	}
	if drawImg != nil {
		w.canvas.DrawImage(offsetX, offsetY, drawImg)
	}
	w.canvas.Render(ctx.Buffer, content.X, content.Y)
}

func (w *AsyncImage) startLoad() {
	if w == nil || w.loader == nil {
		return
	}
	w.loadOnce.Do(func() {
		w.setLoading(true)
		services := w.Services
		effect := runtime.Effect{Run: func(ctx context.Context, post runtime.PostFunc) {
			img, err := w.loader()
			w.setResult(img, err)
			services.Invalidate()
		}}
		if services != (runtime.Services{}) {
			services.Spawn(effect)
			return
		}
		go func() {
			img, err := w.loader()
			w.setResult(img, err)
		}()
	})
}

func (w *AsyncImage) setLoading(loading bool) {
	w.mu.Lock()
	w.loading = loading
	w.mu.Unlock()
}

func (w *AsyncImage) setResult(img image.Image, err error) {
	w.mu.Lock()
	w.img = img
	w.err = err
	w.loading = false
	w.mu.Unlock()
}

func (w *AsyncImage) imageState() (image.Image, error, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.img, w.err, w.loading
}

// Error returns the last load error, if any.
func (w *AsyncImage) Error() error {
	_, err, _ := w.imageState()
	if err == nil {
		return nil
	}
	return fmt.Errorf("async image: %w", err)
}

var _ runtime.Widget = (*AsyncImage)(nil)
var _ runtime.Bindable = (*AsyncImage)(nil)
var _ runtime.Unbindable = (*AsyncImage)(nil)
