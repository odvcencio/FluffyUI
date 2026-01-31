package gpu

import (
	"image"
	"image/color"
	"math"
)

// GPUCanvas renders into an offscreen RGBA buffer.
type GPUCanvas struct {
	driver Driver

	useGPU    bool
	pipeline  *gpuPipeline
	gpuFb     Framebuffer
	gpuLayers []Framebuffer

	rasterDriver *softwareDriver
	rasterFb     *softwareFramebuffer
	rasterLayers []*softwareFramebuffer

	width       int
	height      int
	transform   Matrix3
	stack       []Matrix3
	fillColor   color.RGBA
	strokeColor color.RGBA
	strokeWidth float32
	batch       *drawBatch

	path           []pathOp
	pathStart      vec2
	pathCurrent    vec2
	pathHasCurrent bool

	// layers tracked per backend
}

// NewGPUCanvas creates a new GPU canvas.
func NewGPUCanvas(width, height int) (*GPUCanvas, error) {
	driver, err := NewDriver(BackendAuto)
	if err != nil {
		driver = newSoftwareDriver()
	}
	if err := driver.Init(); err != nil {
		driver = newSoftwareDriver()
	}
	return newGPUCanvasWithDriver(width, height, driver)
}

// NewGPUCanvasWithDriver creates a canvas with a provided driver.
func NewGPUCanvasWithDriver(width, height int, driver Driver) (*GPUCanvas, error) {
	if driver == nil {
		return NewGPUCanvas(width, height)
	}
	if err := driver.Init(); err != nil {
		driver = newSoftwareDriver()
	}
	return newGPUCanvasWithDriver(width, height, driver)
}

func newGPUCanvasWithDriver(width, height int, driver Driver) (*GPUCanvas, error) {
	canvas := &GPUCanvas{
		driver:      driver,
		width:       width,
		height:      height,
		transform:   Identity(),
		fillColor:   color.RGBA{R: 255, G: 255, B: 255, A: 255},
		strokeColor: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		strokeWidth: 1,
		batch:       &drawBatch{},
	}
	if driver != nil && driver.Backend() != BackendSoftware {
		pipeline, err := newGPUPipeline(driver, width, height)
		if err != nil {
			return newGPUCanvasWithDriver(width, height, newSoftwareDriver())
		}
		fb, err := driver.NewFramebuffer(width, height)
		if err != nil {
			pipeline.Dispose()
			return newGPUCanvasWithDriver(width, height, newSoftwareDriver())
		}
		canvas.useGPU = true
		canvas.pipeline = pipeline
		canvas.gpuFb = fb
		canvas.gpuLayers = []Framebuffer{fb}
		return canvas, nil
	}
	rasterDriver := newSoftwareDriver()
	if drv, ok := driver.(*softwareDriver); ok {
		rasterDriver = drv
	}
	fbIface, err := rasterDriver.NewFramebuffer(width, height)
	if err != nil {
		return nil, err
	}
	fb := fbIface.(*softwareFramebuffer)
	canvas.rasterDriver = rasterDriver
	canvas.rasterFb = fb
	canvas.rasterLayers = []*softwareFramebuffer{fb}
	return canvas, nil
}

// Size returns pixel dimensions.
func (c *GPUCanvas) Size() (int, int) {
	if c == nil {
		return 0, 0
	}
	return c.width, c.height
}

// Resize reallocates the canvas.
func (c *GPUCanvas) Resize(width, height int) error {
	if c == nil {
		return ErrUnsupported
	}
	if width <= 0 || height <= 0 {
		return ErrUnsupported
	}
	if c.useGPU {
		if c.driver == nil {
			return ErrUnsupported
		}
		fb, err := c.driver.NewFramebuffer(width, height)
		if err != nil {
			return err
		}
		oldLayers := c.gpuLayers
		c.gpuFb = fb
		c.gpuLayers = []Framebuffer{fb}
		c.width = width
		c.height = height
		if c.pipeline != nil {
			c.pipeline.SetProjection(width, height)
		}
		for _, layer := range oldLayers {
			if layer != nil {
				layer.Dispose()
			}
		}
		return nil
	}
	if c.rasterDriver == nil {
		return ErrUnsupported
	}
	fbIface, err := c.rasterDriver.NewFramebuffer(width, height)
	if err != nil {
		return err
	}
	fb := fbIface.(*softwareFramebuffer)
	oldLayers := c.rasterLayers
	c.rasterFb = fb
	c.width = width
	c.height = height
	c.rasterLayers = []*softwareFramebuffer{fb}
	for _, layer := range oldLayers {
		if layer != nil {
			layer.Dispose()
		}
	}
	return nil
}

// Dispose releases resources.
func (c *GPUCanvas) Dispose() {
	if c == nil {
		return
	}
	if c.useGPU {
		for _, layer := range c.gpuLayers {
			if layer != nil {
				layer.Dispose()
			}
		}
		c.gpuLayers = nil
		c.gpuFb = nil
		if c.pipeline != nil {
			c.pipeline.Dispose()
			c.pipeline = nil
		}
	} else {
		for _, layer := range c.rasterLayers {
			if layer != nil {
				layer.Dispose()
			}
		}
		c.rasterLayers = nil
		c.rasterFb = nil
		c.rasterDriver = nil
	}
	if c.driver != nil {
		c.driver.Dispose()
	}
	c.driver = nil
}

// Begin starts a new frame.
func (c *GPUCanvas) Begin() {
	if c == nil {
		return
	}
	c.batch.Reset()
	c.path = c.path[:0]
	c.pathHasCurrent = false
}

// End returns the RGBA pixels for the current frame.
func (c *GPUCanvas) End() []byte {
	if c == nil {
		return nil
	}
	if c.useGPU {
		if c.driver == nil || c.gpuFb == nil {
			return nil
		}
		pixels, _ := c.driver.ReadPixels(c.gpuFb, image.Rectangle{})
		return pixels
	}
	if c.rasterFb == nil || c.rasterFb.tex == nil {
		return nil
	}
	out := make([]byte, len(c.rasterFb.tex.pixels))
	copy(out, c.rasterFb.tex.pixels)
	return out
}

// EndToTexture captures the current frame into a texture.
func (c *GPUCanvas) EndToTexture() Texture {
	if c == nil || c.driver == nil {
		return nil
	}
	if c.useGPU && c.gpuFb != nil {
		return c.gpuFb.Texture()
	}
	pixels := c.End()
	if len(pixels) == 0 {
		return nil
	}
	tex, err := c.driver.NewTexture(c.width, c.height)
	if err != nil {
		return nil
	}
	tex.Upload(pixels, image.Rectangle{})
	return tex
}

// SetFillColor sets the fill color.
func (c *GPUCanvas) SetFillColor(col color.RGBA) {
	if c == nil {
		return
	}
	c.fillColor = col
}

// SetStrokeColor sets the stroke color.
func (c *GPUCanvas) SetStrokeColor(col color.RGBA) {
	if c == nil {
		return
	}
	c.strokeColor = col
}

// SetStrokeWidth sets the stroke width.
func (c *GPUCanvas) SetStrokeWidth(w float32) {
	if c == nil {
		return
	}
	if w <= 0 {
		w = 1
	}
	c.strokeWidth = w
}

// Save stores the current transform.
func (c *GPUCanvas) Save() {
	if c == nil {
		return
	}
	c.stack = append(c.stack, c.transform)
}

// Restore restores the last transform.
func (c *GPUCanvas) Restore() {
	if c == nil || len(c.stack) == 0 {
		return
	}
	c.transform = c.stack[len(c.stack)-1]
	c.stack = c.stack[:len(c.stack)-1]
}

// Translate offsets the current transform.
func (c *GPUCanvas) Translate(x, y float32) {
	if c == nil {
		return
	}
	c.transform = c.transform.Mul(Translate(x, y))
}

// Rotate rotates the current transform.
func (c *GPUCanvas) Rotate(radians float32) {
	if c == nil {
		return
	}
	c.transform = c.transform.Mul(Rotate(radians))
}

// Scale scales the current transform.
func (c *GPUCanvas) Scale(x, y float32) {
	if c == nil {
		return
	}
	c.transform = c.transform.Mul(Scale(x, y))
}

// Clear fills the canvas with a color.
func (c *GPUCanvas) Clear(col color.RGBA) {
	if c == nil {
		return
	}
	if c.usingGPU() {
		c.clearGPU(col)
		return
	}
	fb := c.rasterFb
	if fb == nil || fb.tex == nil {
		return
	}
	clearPixels(fb.tex.pixels, fb.tex.width, fb.tex.height, col)
}

// FillRect fills a rectangle.
func (c *GPUCanvas) FillRect(x, y, w, h float32) {
	if c == nil || w <= 0 || h <= 0 {
		return
	}
	if c.usingGPU() {
		c.fillRectGPU(x, y, w, h)
		return
	}
	points := []vec2{{x: x, y: y}, {x: x + w, y: y}, {x: x + w, y: y + h}, {x: x, y: y + h}}
	c.fillPolygon(points, c.fillColor)
}

// StrokeRect draws a rectangle outline.
func (c *GPUCanvas) StrokeRect(x, y, w, h float32) {
	if c == nil || w <= 0 || h <= 0 {
		return
	}
	if c.usingGPU() {
		c.strokeRectGPU(x, y, w, h)
		return
	}
	points := []vec2{{x: x, y: y}, {x: x + w, y: y}, {x: x + w, y: y + h}, {x: x, y: y + h}}
	c.strokePoints(points, c.strokeColor, c.strokeWidth, true)
}

// FillRoundedRect fills a rounded rectangle.
func (c *GPUCanvas) FillRoundedRect(x, y, w, h, radius float32) {
	if c == nil || w <= 0 || h <= 0 {
		return
	}
	if c.usingGPU() {
		if radius <= 0 {
			c.fillRectGPU(x, y, w, h)
			return
		}
		points := roundedRectPoints(x, y, w, h, radius)
		c.fillPolygonGPU(points, c.fillColor)
		return
	}
	if radius <= 0 {
		c.FillRect(x, y, w, h)
		return
	}
	points := roundedRectPoints(x, y, w, h, radius)
	c.fillPolygon(points, c.fillColor)
}

// FillCircle fills a circle.
func (c *GPUCanvas) FillCircle(cx, cy, r float32) {
	if c == nil || r <= 0 {
		return
	}
	if c.usingGPU() {
		points := circlePoints(cx, cy, r)
		c.fillPolygonGPU(points, c.fillColor)
		return
	}
	points := circlePoints(cx, cy, r)
	c.fillPolygon(points, c.fillColor)
}

// StrokeCircle draws a circle outline.
func (c *GPUCanvas) StrokeCircle(cx, cy, r float32) {
	if c == nil || r <= 0 {
		return
	}
	if c.usingGPU() {
		points := circlePoints(cx, cy, r)
		c.strokePointsGPU(points, c.strokeColor, c.strokeWidth, true)
		return
	}
	points := circlePoints(cx, cy, r)
	c.strokePoints(points, c.strokeColor, c.strokeWidth, true)
}

// DrawLine draws a line.
func (c *GPUCanvas) DrawLine(x1, y1, x2, y2 float32) {
	if c == nil {
		return
	}
	if c.usingGPU() {
		p1 := c.applyTransform(vec2{x: x1, y: y1})
		p2 := c.applyTransform(vec2{x: x2, y: y2})
		c.drawLineGPU(p1, p2, c.strokeColor, c.strokeWidth)
		return
	}
	p1 := c.applyTransform(vec2{x: x1, y: y1})
	p2 := c.applyTransform(vec2{x: x2, y: y2})
	c.drawLine(p1, p2, c.strokeColor, c.strokeWidth)
}

// BeginPath starts a new path.
func (c *GPUCanvas) BeginPath() {
	if c == nil {
		return
	}
	c.path = c.path[:0]
	c.pathHasCurrent = false
}

// MoveTo moves the current point.
func (c *GPUCanvas) MoveTo(x, y float32) {
	if c == nil {
		return
	}
	p := vec2{x: x, y: y}
	c.pathStart = p
	c.pathCurrent = p
	c.pathHasCurrent = true
	c.path = append(c.path, pathOp{kind: pathMove, p1: p})
}

// LineTo adds a line segment.
func (c *GPUCanvas) LineTo(x, y float32) {
	if c == nil || !c.pathHasCurrent {
		return
	}
	p := vec2{x: x, y: y}
	c.pathCurrent = p
	c.path = append(c.path, pathOp{kind: pathLine, p1: p})
}

// QuadraticTo adds a quadratic curve.
func (c *GPUCanvas) QuadraticTo(cx, cy, x, y float32) {
	if c == nil || !c.pathHasCurrent {
		return
	}
	cp := vec2{x: cx, y: cy}
	end := vec2{x: x, y: y}
	c.pathCurrent = end
	c.path = append(c.path, pathOp{kind: pathQuad, p1: cp, p2: end})
}

// BezierTo adds a cubic curve.
func (c *GPUCanvas) BezierTo(c1x, c1y, c2x, c2y, x, y float32) {
	if c == nil || !c.pathHasCurrent {
		return
	}
	cp1 := vec2{x: c1x, y: c1y}
	cp2 := vec2{x: c2x, y: c2y}
	end := vec2{x: x, y: y}
	c.pathCurrent = end
	c.path = append(c.path, pathOp{kind: pathCubic, p1: cp1, p2: cp2, p3: end})
}

// ClosePath closes the current path.
func (c *GPUCanvas) ClosePath() {
	if c == nil || !c.pathHasCurrent {
		return
	}
	c.pathCurrent = c.pathStart
	c.path = append(c.path, pathOp{kind: pathClose})
}

// Fill fills the current path.
func (c *GPUCanvas) Fill() {
	if c == nil || len(c.path) == 0 {
		return
	}
	if c.usingGPU() {
		points := pathToPoints(c.path)
		c.fillPolygonGPU(points, c.fillColor)
		return
	}
	points := pathToPoints(c.path)
	c.fillPolygon(points, c.fillColor)
}

// Stroke strokes the current path.
func (c *GPUCanvas) Stroke() {
	if c == nil || len(c.path) == 0 {
		return
	}
	if c.usingGPU() {
		points := pathToPoints(c.path)
		closed := pathClosed(c.path)
		c.strokePointsGPU(points, c.strokeColor, c.strokeWidth, closed)
		return
	}
	points := pathToPoints(c.path)
	closed := pathClosed(c.path)
	c.strokePoints(points, c.strokeColor, c.strokeWidth, closed)
}

// DrawImage draws a texture at the given position.
func (c *GPUCanvas) DrawImage(img Texture, x, y float32) {
	if c == nil || img == nil {
		return
	}
	if c.usingGPU() {
		w, h := img.Size()
		c.drawTextureGPU(img, x, y, float32(w), float32(h))
		return
	}
	w, h := img.Size()
	c.drawTexture(img, x, y, float32(w), float32(h))
}

// DrawImageScaled draws a texture scaled to the given size.
func (c *GPUCanvas) DrawImageScaled(img Texture, x, y, w, h float32) {
	if c == nil || img == nil {
		return
	}
	if c.usingGPU() {
		c.drawTextureGPU(img, x, y, w, h)
		return
	}
	c.drawTexture(img, x, y, w, h)
}

// DrawText draws text using a pixel font.
func (c *GPUCanvas) DrawText(text string, x, y float32, font *Font) {
	if c == nil || text == "" {
		return
	}
	if c.usingGPU() {
		c.drawTextGPU(text, x, y, font)
		return
	}
	if font == nil {
		font = DefaultFont()
	}
	pxFont := font.Face
	if pxFont == nil {
		return
	}
	cursorX := x
	for _, r := range text {
		glyph := pxFont.Glyph(r)
		for gy, row := range glyph {
			for gx := 0; gx < len(row); gx++ {
				if row[gx] != '#' {
					continue
				}
				p := c.applyTransform(vec2{x: cursorX + float32(gx), y: y + float32(gy)})
				c.setPixelAt(p, c.fillColor)
			}
		}
		cursorX += float32(pxFont.Width + pxFont.Spacing)
	}
}

// ApplyEffect applies a post-processing effect.
func (c *GPUCanvas) ApplyEffect(effect Effect) {
	if c == nil || effect == nil {
		return
	}
	if c.usingGPU() {
		if c.driver == nil || c.gpuFb == nil {
			return
		}
		srcTex := c.gpuFb.Texture()
		dst, err := c.driver.NewFramebuffer(c.width, c.height)
		if err != nil {
			return
		}
		effect.Apply(srcTex, dst, c.driver)
		old := c.gpuFb
		c.gpuFb = dst
		if len(c.gpuLayers) > 0 {
			c.gpuLayers[len(c.gpuLayers)-1] = dst
		}
		if old != nil {
			old.Dispose()
		}
		return
	}
	if c.rasterDriver == nil || c.rasterFb == nil {
		return
	}
	srcTex := c.rasterFb.Texture()
	dstIface, err := c.rasterDriver.NewFramebuffer(c.width, c.height)
	if err != nil {
		return
	}
	dst := dstIface.(*softwareFramebuffer)
	effect.Apply(srcTex, dst, c.rasterDriver)
	old := c.rasterFb
	c.rasterFb = dst
	if len(c.rasterLayers) > 0 {
		c.rasterLayers[len(c.rasterLayers)-1] = dst
	}
	if old != nil {
		old.Dispose()
	}
}

// PushLayer starts a new layer.
func (c *GPUCanvas) PushLayer() {
	if c == nil {
		return
	}
	if c.usingGPU() {
		if c.driver == nil {
			return
		}
		fb, err := c.driver.NewFramebuffer(c.width, c.height)
		if err != nil {
			return
		}
		c.gpuFb = fb
		c.gpuLayers = append(c.gpuLayers, fb)
		return
	}
	if c.rasterDriver == nil {
		return
	}
	fbIface, err := c.rasterDriver.NewFramebuffer(c.width, c.height)
	if err != nil {
		return
	}
	fb := fbIface.(*softwareFramebuffer)
	c.rasterFb = fb
	c.rasterLayers = append(c.rasterLayers, fb)
}

// PopLayer composites the current layer onto the previous one.
func (c *GPUCanvas) PopLayer(effects ...Effect) {
	if c == nil {
		return
	}
	if c.usingGPU() {
		if len(c.gpuLayers) < 2 {
			return
		}
		layer := c.gpuLayers[len(c.gpuLayers)-1]
		c.gpuLayers = c.gpuLayers[:len(c.gpuLayers)-1]
		c.gpuFb = c.gpuLayers[len(c.gpuLayers)-1]
		if layer == nil || c.gpuFb == nil {
			return
		}
		if len(effects) > 0 {
			src := layer.Texture()
			temp := layer
			orig := layer
			for _, effect := range effects {
				dst, err := c.driver.NewFramebuffer(c.width, c.height)
				if err != nil {
					return
				}
				effect.Apply(src, dst, c.driver)
				if temp != orig {
					temp.Dispose()
				}
				src = dst.Texture()
				temp = dst
			}
			if orig != temp {
				orig.Dispose()
			}
			layer = temp
		}
		c.compositeGPU(layer, c.gpuFb)
		if layer != nil {
			layer.Dispose()
		}
		return
	}
	if len(c.rasterLayers) < 2 {
		return
	}
	layer := c.rasterLayers[len(c.rasterLayers)-1]
	c.rasterLayers = c.rasterLayers[:len(c.rasterLayers)-1]
	c.rasterFb = c.rasterLayers[len(c.rasterLayers)-1]
	if layer == nil || c.rasterFb == nil {
		return
	}
	if len(effects) > 0 {
		src := layer.Texture()
		temp := layer
		orig := layer
		for _, effect := range effects {
			dstIface, err := c.rasterDriver.NewFramebuffer(c.width, c.height)
			if err != nil {
				return
			}
			dst := dstIface.(*softwareFramebuffer)
			effect.Apply(src, dst, c.rasterDriver)
			if temp != orig {
				temp.Dispose()
			}
			src = dst.Texture()
			temp = dst
		}
		if orig != temp {
			orig.Dispose()
		}
		layer = temp
	}
	c.composite(layer, c.rasterFb)
	if layer != nil {
		layer.Dispose()
	}
}

func (c *GPUCanvas) currentFramebuffer() (*softwareFramebuffer, bool) {
	if c == nil || c.rasterFb == nil || c.rasterFb.tex == nil {
		return nil, false
	}
	return c.rasterFb, true
}

func (c *GPUCanvas) usingGPU() bool {
	return c != nil && c.useGPU && c.driver != nil && c.gpuFb != nil && c.pipeline != nil
}

func (c *GPUCanvas) setPixelAt(p vec2, col color.RGBA) {
	fb, ok := c.currentFramebuffer()
	if !ok {
		return
	}
	x := int(math.Round(float64(p.x)))
	y := int(math.Round(float64(p.y)))
	if col.A == 255 {
		setPixel(fb.tex.pixels, fb.tex.width, fb.tex.height, x, y, col)
		return
	}
	blendPixel(fb.tex.pixels, fb.tex.width, fb.tex.height, x, y, col)
}

func (c *GPUCanvas) applyTransform(p vec2) vec2 {
	if c == nil {
		return p
	}
	x, y := c.transform.Apply(p.x, p.y)
	return vec2{x: x, y: y}
}

func (c *GPUCanvas) fillPolygon(points []vec2, col color.RGBA) {
	fb, ok := c.currentFramebuffer()
	if !ok || len(points) < 3 {
		return
	}
	transformed := make([]vec2, len(points))
	for i, p := range points {
		transformed[i] = c.applyTransform(p)
	}
	fillPolygon(transformed, func(x, y int) {
		if col.A == 255 {
			setPixel(fb.tex.pixels, fb.tex.width, fb.tex.height, x, y, col)
			return
		}
		blendPixel(fb.tex.pixels, fb.tex.width, fb.tex.height, x, y, col)
	})
}

func (c *GPUCanvas) strokePoints(points []vec2, col color.RGBA, width float32, closed bool) {
	if len(points) < 2 {
		return
	}
	transformed := make([]vec2, len(points))
	for i, p := range points {
		transformed[i] = c.applyTransform(p)
	}
	last := len(transformed) - 1
	for i := 0; i < last; i++ {
		c.drawLine(transformed[i], transformed[i+1], col, width)
	}
	if closed && len(transformed) > 2 {
		c.drawLine(transformed[last], transformed[0], col, width)
	}
}

func (c *GPUCanvas) drawLine(p1, p2 vec2, col color.RGBA, width float32) {
	fb, ok := c.currentFramebuffer()
	if !ok {
		return
	}
	if width <= 1 {
		x1 := int(math.Round(float64(p1.x)))
		y1 := int(math.Round(float64(p1.y)))
		x2 := int(math.Round(float64(p2.x)))
		y2 := int(math.Round(float64(p2.y)))
		bresenhamLine(x1, y1, x2, y2, func(x, y int) {
			if col.A == 255 {
				setPixel(fb.tex.pixels, fb.tex.width, fb.tex.height, x, y, col)
				return
			}
			blendPixel(fb.tex.pixels, fb.tex.width, fb.tex.height, x, y, col)
		})
		return
	}
	radius := width / 2
	x1 := int(math.Round(float64(p1.x)))
	y1 := int(math.Round(float64(p1.y)))
	x2 := int(math.Round(float64(p2.x)))
	y2 := int(math.Round(float64(p2.y)))
	bresenhamLine(x1, y1, x2, y2, func(x, y int) {
		fillCircleAt(fb.tex.pixels, fb.tex.width, fb.tex.height, float32(x), float32(y), radius, col)
	})
}

func (c *GPUCanvas) drawTexture(img Texture, x, y, w, h float32) {
	fb, ok := c.currentFramebuffer()
	if !ok {
		return
	}
	src, srcW, srcH, ok := texturePixels(img, c.driver)
	if !ok {
		return
	}
	if int(w) != srcW || int(h) != srcH {
		src = scalePixels(src, srcW, srcH, int(w), int(h))
		srcW, srcH = int(w), int(h)
	}
	for sy := 0; sy < srcH; sy++ {
		for sx := 0; sx < srcW; sx++ {
			idx := (sy*srcW + sx) * 4
			col := color.RGBA{R: src[idx], G: src[idx+1], B: src[idx+2], A: src[idx+3]}
			p := c.applyTransform(vec2{x: x + float32(sx), y: y + float32(sy)})
			set := col.A == 255
			px := int(math.Round(float64(p.x)))
			py := int(math.Round(float64(p.y)))
			if set {
				setPixel(fb.tex.pixels, fb.tex.width, fb.tex.height, px, py, col)
				continue
			}
			blendPixel(fb.tex.pixels, fb.tex.width, fb.tex.height, px, py, col)
		}
	}
}

func (c *GPUCanvas) composite(src, dst *softwareFramebuffer) {
	if src == nil || dst == nil || src.tex == nil || dst.tex == nil {
		return
	}
	sw, sh := src.tex.width, src.tex.height
	if sw != dst.tex.width || sh != dst.tex.height {
		return
	}
	for y := 0; y < sh; y++ {
		for x := 0; x < sw; x++ {
			idx := (y*sw + x) * 4
			col := color.RGBA{R: src.tex.pixels[idx], G: src.tex.pixels[idx+1], B: src.tex.pixels[idx+2], A: src.tex.pixels[idx+3]}
			if col.A == 0 {
				continue
			}
			blendPixel(dst.tex.pixels, dst.tex.width, dst.tex.height, x, y, col)
		}
	}
}

// Font wraps a pixel font face.
type Font struct {
	Face *PixelFont
}

// DefaultFont returns the default pixel font.
func DefaultFont() *Font {
	return &Font{Face: DefaultPixelFont}
}

func roundedRectPoints(x, y, w, h, radius float32) []vec2 {
	if radius > w/2 {
		radius = w / 2
	}
	if radius > h/2 {
		radius = h / 2
	}
	segments := 8
	points := make([]vec2, 0, segments*4)
	cx := x + radius
	cy := y + radius
	points = append(points, arcPoints(cx, cy, radius, math.Pi, math.Pi*1.5, segments)...)
	cx = x + w - radius
	cy = y + radius
	points = append(points, arcPoints(cx, cy, radius, math.Pi*1.5, math.Pi*2, segments)...)
	cx = x + w - radius
	cy = y + h - radius
	points = append(points, arcPoints(cx, cy, radius, 0, math.Pi*0.5, segments)...)
	cx = x + radius
	cy = y + h - radius
	points = append(points, arcPoints(cx, cy, radius, math.Pi*0.5, math.Pi, segments)...)
	return points
}

func circlePoints(cx, cy, r float32) []vec2 {
	segments := int(math.Max(16, float64(r*2)))
	points := make([]vec2, 0, segments)
	step := float32(2 * math.Pi / float64(segments))
	for i := 0; i < segments; i++ {
		angle := float32(i) * step
		points = append(points, vec2{x: cx + float32(math.Cos(float64(angle)))*r, y: cy + float32(math.Sin(float64(angle)))*r})
	}
	return points
}

func arcPoints(cx, cy, r float32, start, end float64, segments int) []vec2 {
	points := make([]vec2, 0, segments)
	step := (end - start) / float64(segments)
	for i := 0; i <= segments; i++ {
		angle := start + float64(i)*step
		points = append(points, vec2{x: cx + float32(math.Cos(angle))*r, y: cy + float32(math.Sin(angle))*r})
	}
	return points
}

func fillPolygon(points []vec2, plot func(x, y int)) {
	if len(points) < 3 {
		return
	}
	minY := points[0].y
	maxY := points[0].y
	for _, p := range points[1:] {
		if p.y < minY {
			minY = p.y
		}
		if p.y > maxY {
			maxY = p.y
		}
	}
	yStart := int(math.Floor(float64(minY)))
	yEnd := int(math.Ceil(float64(maxY)))
	for y := yStart; y <= yEnd; y++ {
		intersections := make([]float32, 0, len(points))
		for i := 0; i < len(points); i++ {
			p1 := points[i]
			p2 := points[(i+1)%len(points)]
			if (p1.y <= float32(y) && p2.y > float32(y)) || (p2.y <= float32(y) && p1.y > float32(y)) {
				t := (float32(y) - p1.y) / (p2.y - p1.y)
				x := p1.x + t*(p2.x-p1.x)
				intersections = append(intersections, x)
			}
		}
		if len(intersections) < 2 {
			continue
		}
		sortFloat32(intersections)
		for i := 0; i+1 < len(intersections); i += 2 {
			xStart := int(math.Ceil(float64(intersections[i])))
			xEnd := int(math.Floor(float64(intersections[i+1])))
			for x := xStart; x <= xEnd; x++ {
				plot(x, y)
			}
		}
	}
}

func bresenhamLine(x1, y1, x2, y2 int, plot func(x, y int)) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx := -1
	if x1 < x2 {
		sx = 1
	}
	sy := -1
	if y1 < y2 {
		sy = 1
	}
	err := dx - dy
	for {
		plot(x1, y1)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := err * 2
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func fillCircleAt(pixels []byte, w, h int, cx, cy, r float32, col color.RGBA) {
	if r <= 0 {
		return
	}
	minX := int(math.Floor(float64(cx - r)))
	maxX := int(math.Ceil(float64(cx + r)))
	minY := int(math.Floor(float64(cy - r)))
	maxY := int(math.Ceil(float64(cy + r)))
	r2 := r * r
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			dx := float32(x) - cx
			dy := float32(y) - cy
			if dx*dx+dy*dy <= r2 {
				if col.A == 255 {
					setPixel(pixels, w, h, x, y, col)
					continue
				}
				blendPixel(pixels, w, h, x, y, col)
			}
		}
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func sortFloat32(values []float32) {
	for i := 1; i < len(values); i++ {
		j := i
		for j > 0 && values[j-1] > values[j] {
			values[j-1], values[j] = values[j], values[j-1]
			j--
		}
	}
}

func pathClosed(ops []pathOp) bool {
	if len(ops) == 0 {
		return false
	}
	return ops[len(ops)-1].kind == pathClose
}
