package graphics

import (
	"image"
	"math"

	"github.com/odvcencio/fluffyui/backend"
)

// Canvas provides high-level drawing operations.
type Canvas struct {
	buffer  *PixelBuffer
	blitter Blitter

	strokeColor Color
	fillColor   Color
	lineWidth   int

	transform      Transform
	transformStack []Transform

	path           []pathOp
	pathStart      Point
	pathCurrent    Point
	pathHasCurrent bool
}

// NewCanvas creates a canvas for the given cell dimensions.
func NewCanvas(cellWidth, cellHeight int) *Canvas {
	return NewCanvasWithBlitter(cellWidth, cellHeight, &SextantBlitter{})
}

// NewCanvasWithBlitter creates a canvas with the provided blitter.
func NewCanvasWithBlitter(cellWidth, cellHeight int, blitter Blitter) *Canvas {
	if blitter == nil {
		blitter = &SextantBlitter{}
	}
	return &Canvas{
		buffer:      NewPixelBuffer(cellWidth, cellHeight, blitter),
		blitter:     blitter,
		strokeColor: backend.ColorWhite,
		fillColor:   backend.ColorWhite,
		lineWidth:   1,
		transform:   IdentityTransform(),
	}
}

// Size returns pixel dimensions.
func (c *Canvas) Size() (width, height int) {
	if c == nil || c.buffer == nil {
		return 0, 0
	}
	return c.buffer.Size()
}

// CellSize returns cell dimensions.
func (c *Canvas) CellSize() (width, height int) {
	if c == nil || c.buffer == nil {
		return 0, 0
	}
	pw, ph := 1, 1
	if c.blitter != nil {
		pw, ph = c.blitter.PixelsPerCell()
		if pw <= 0 {
			pw = 1
		}
		if ph <= 0 {
			ph = 1
		}
	}
	w, h := c.buffer.Size()
	return w / pw, h / ph
}

// Clear clears the canvas.
func (c *Canvas) Clear() {
	if c == nil || c.buffer == nil {
		return
	}
	c.buffer.Clear()
	c.path = nil
	c.pathHasCurrent = false
}

// Save stores the current transform on the stack.
func (c *Canvas) Save() {
	if c == nil {
		return
	}
	c.transformStack = append(c.transformStack, c.transform)
}

// Restore restores the most recently saved transform.
func (c *Canvas) Restore() {
	if c == nil {
		return
	}
	if len(c.transformStack) == 0 {
		return
	}
	c.transform = c.transformStack[len(c.transformStack)-1]
	c.transformStack = c.transformStack[:len(c.transformStack)-1]
}

// Translate offsets the current transform.
func (c *Canvas) Translate(dx, dy int) {
	if c == nil {
		return
	}
	c.transform = c.transform.Mul(TranslateTransform(float64(dx), float64(dy)))
}

// Rotate rotates the current transform clockwise by angle (radians).
func (c *Canvas) Rotate(angle float64) {
	if c == nil {
		return
	}
	c.transform = c.transform.Mul(RotateTransform(angle))
}

// Scale scales the current transform.
func (c *Canvas) Scale(sx, sy float64) {
	if c == nil {
		return
	}
	c.transform = c.transform.Mul(ScaleTransform(sx, sy))
}

// SetStrokeColor sets the stroke color.
func (c *Canvas) SetStrokeColor(color Color) { c.strokeColor = color }

// SetFillColor sets the fill color.
func (c *Canvas) SetFillColor(color Color) { c.fillColor = color }

// SetLineWidth sets the line width.
func (c *Canvas) SetLineWidth(width int) { c.lineWidth = width }

// SetPixel sets a pixel color.
func (c *Canvas) SetPixel(x, y int, color Color) {
	if c == nil || c.buffer == nil {
		return
	}
	c.plotPixel(x, y, color)
}

// Blend blends a pixel color with alpha.
func (c *Canvas) Blend(x, y int, color Color, alpha float32) {
	if c == nil || c.buffer == nil {
		return
	}
	c.blendPixel(x, y, color, float64(alpha))
}

// GetPixel returns the pixel at position.
func (c *Canvas) GetPixel(x, y int) Pixel {
	if c == nil || c.buffer == nil {
		return Pixel{}
	}
	tx, ty := c.transformPoint(x, y)
	return c.buffer.Get(tx, ty)
}

func (c *Canvas) transformPoint(x, y int) (int, int) {
	if c == nil || c.transform.IsIdentity() {
		return x, y
	}
	fx, fy := c.transform.Apply(float64(x), float64(y))
	return round(fx), round(fy)
}

func (c *Canvas) plotPixel(x, y int, color Color) {
	if c == nil || c.buffer == nil {
		return
	}
	tx, ty := c.transformPoint(x, y)
	c.buffer.SetPixel(tx, ty, color)
}

func (c *Canvas) blendPixel(x, y int, color Color, alpha float64) {
	if c == nil || c.buffer == nil {
		return
	}
	if alpha <= 0 {
		return
	}
	if alpha >= 1 {
		c.plotPixel(x, y, color)
		return
	}
	tx, ty := c.transformPoint(x, y)
	c.buffer.Blend(tx, ty, color, float32(alpha))
}

// DrawLine draws a line from (x1,y1) to (x2,y2).
func (c *Canvas) DrawLine(x1, y1, x2, y2 int) {
	if c == nil || c.buffer == nil {
		return
	}
	bresenhamLine(x1, y1, x2, y2, func(x, y int) {
		c.plotPixel(x, y, c.strokeColor)
	})
}

// DrawLineAA draws an anti-aliased line.
func (c *Canvas) DrawLineAA(x1, y1, x2, y2 int) {
	if c == nil || c.buffer == nil {
		return
	}
	wuLine(float64(x1), float64(y1), float64(x2), float64(y2), func(x, y int, alpha float64) {
		c.blendPixel(x, y, c.strokeColor, alpha)
	})
}

// DrawRect draws a rectangle outline.
func (c *Canvas) DrawRect(x, y, w, h int) {
	if c == nil || c.buffer == nil || w <= 0 || h <= 0 {
		return
	}
	for dx := 0; dx < w; dx++ {
		c.plotPixel(x+dx, y, c.strokeColor)
		c.plotPixel(x+dx, y+h-1, c.strokeColor)
	}
	for dy := 0; dy < h; dy++ {
		c.plotPixel(x, y+dy, c.strokeColor)
		c.plotPixel(x+w-1, y+dy, c.strokeColor)
	}
}

// FillRect fills a rectangle.
func (c *Canvas) FillRect(x, y, w, h int) {
	if c == nil || c.buffer == nil || w <= 0 || h <= 0 {
		return
	}
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			c.plotPixel(x+dx, y+dy, c.fillColor)
		}
	}
}

// DrawEllipse draws an ellipse outline.
func (c *Canvas) DrawEllipse(cx, cy, rx, ry int) {
	if c == nil || c.buffer == nil {
		return
	}
	midpointEllipse(cx, cy, rx, ry, func(x, y int) {
		c.plotPixel(x, y, c.strokeColor)
	})
}

// FillEllipse fills an ellipse.
func (c *Canvas) FillEllipse(cx, cy, rx, ry int) {
	if c == nil || c.buffer == nil {
		return
	}
	fillEllipse(cx, cy, rx, ry, func(x, y int) {
		c.plotPixel(x, y, c.fillColor)
	})
}

// DrawCircle draws a circle outline.
func (c *Canvas) DrawCircle(cx, cy, radius int) {
	if c == nil || c.buffer == nil || radius <= 0 {
		return
	}
	midpointCircle(cx, cy, radius, func(x, y int) {
		c.plotPixel(x, y, c.strokeColor)
	})
}

// FillCircle fills a circle.
func (c *Canvas) FillCircle(cx, cy, radius int) {
	if c == nil || c.buffer == nil || radius <= 0 {
		return
	}
	fillCircle(cx, cy, radius, func(x, y int) {
		c.plotPixel(x, y, c.fillColor)
	})
}

// DrawArc draws an arc along a circle.
func (c *Canvas) DrawArc(cx, cy, radius int, startAngle, endAngle float64) {
	if c == nil || c.buffer == nil || radius <= 0 {
		return
	}
	drawArc(cx, cy, radius, startAngle, endAngle, func(x, y int) {
		c.plotPixel(x, y, c.strokeColor)
	})
}

// DrawRoundedRect draws a rounded rectangle outline.
func (c *Canvas) DrawRoundedRect(x, y, w, h, radius int) {
	if c == nil || c.buffer == nil || w <= 0 || h <= 0 {
		return
	}
	radius = clampRadius(radius, w, h)
	if radius <= 0 {
		c.DrawRect(x, y, w, h)
		return
	}
	right := x + w - 1
	bottom := y + h - 1

	c.DrawLine(x+radius, y, right-radius, y)
	c.DrawLine(x+radius, bottom, right-radius, bottom)
	c.DrawLine(x, y+radius, x, bottom-radius)
	c.DrawLine(right, y+radius, right, bottom-radius)

	c.DrawArc(x+radius, y+radius, radius, math.Pi, 1.5*math.Pi)
	c.DrawArc(right-radius, y+radius, radius, 1.5*math.Pi, 2*math.Pi)
	c.DrawArc(right-radius, bottom-radius, radius, 0, 0.5*math.Pi)
	c.DrawArc(x+radius, bottom-radius, radius, 0.5*math.Pi, math.Pi)
}

// FillRoundedRect fills a rounded rectangle.
func (c *Canvas) FillRoundedRect(x, y, w, h, radius int) {
	if c == nil || c.buffer == nil || w <= 0 || h <= 0 {
		return
	}
	radius = clampRadius(radius, w, h)
	if radius <= 0 {
		c.FillRect(x, y, w, h)
		return
	}

	centerLeft := float64(x + radius)
	centerRight := float64(x + w - radius - 1)
	centerTop := float64(y + radius)
	centerBottom := float64(y + h - radius - 1)

	for dy := 0; dy < h; dy++ {
		py := y + dy
		left := x
		right := x + w - 1

		if py < y+radius {
			dyFloat := float64(py) - centerTop
			dx := circleOffset(radius, dyFloat)
			left = int(centerLeft) - dx
			right = int(centerRight) + dx
		} else if py > y+h-radius-1 {
			dyFloat := float64(py) - centerBottom
			dx := circleOffset(radius, dyFloat)
			left = int(centerLeft) - dx
			right = int(centerRight) + dx
		}

		if left < x {
			left = x
		}
		if right > x+w-1 {
			right = x + w - 1
		}

		for px := left; px <= right; px++ {
			c.plotPixel(px, py, c.fillColor)
		}
	}
}

// DrawBezier draws a cubic bezier curve.
func (c *Canvas) DrawBezier(p0, p1, p2, p3 Point) {
	if c == nil || c.buffer == nil {
		return
	}
	rasterBezier(p0, p1, p2, p3, 0.5, func(x, y int) {
		c.plotPixel(x, y, c.strokeColor)
	})
}

// DrawQuadBezier draws a quadratic bezier curve.
func (c *Canvas) DrawQuadBezier(p0, p1, p2 Point) {
	if c == nil || c.buffer == nil {
		return
	}
	rasterQuadBezier(p0, p1, p2, 0.5, func(x, y int) {
		c.plotPixel(x, y, c.strokeColor)
	})
}

// DrawSpline draws a Catmull-Rom spline through points.
func (c *Canvas) DrawSpline(points []Point) {
	if c == nil || c.buffer == nil {
		return
	}
	drawSpline(points, func(x, y int) {
		c.plotPixel(x, y, c.strokeColor)
	})
}

// DrawPolygon draws a polygon outline.
func (c *Canvas) DrawPolygon(points []Point) {
	if c == nil || c.buffer == nil || len(points) < 2 {
		return
	}
	for i := 0; i < len(points); i++ {
		a := points[i]
		b := points[(i+1)%len(points)]
		bresenhamLine(a.X, a.Y, b.X, b.Y, func(x, y int) {
			c.plotPixel(x, y, c.strokeColor)
		})
	}
}

// FillPolygon fills a polygon.
func (c *Canvas) FillPolygon(points []Point) {
	if c == nil || c.buffer == nil || len(points) < 3 {
		return
	}
	fillPolygon(points, func(x, y int) {
		c.plotPixel(x, y, c.fillColor)
	})
}

// DrawTriangle draws a triangle outline.
func (c *Canvas) DrawTriangle(p1, p2, p3 Point) {
	c.DrawPolygon([]Point{p1, p2, p3})
}

// FillTriangle fills a triangle.
func (c *Canvas) FillTriangle(p1, p2, p3 Point) {
	c.FillPolygon([]Point{p1, p2, p3})
}

// DrawText draws pixel-font text.
func (c *Canvas) DrawText(x, y int, text string, font *PixelFont) {
	if c == nil || c.buffer == nil {
		return
	}
	if font == nil {
		font = DefaultFont
	}
	if font == nil || font.Width <= 0 || font.Height <= 0 {
		return
	}
	cursorX := x
	cursorY := y
	startX := x
	advance := font.Width + font.Spacing
	lineAdvance := font.Height + font.Spacing

	for _, r := range text {
		if r == '\n' {
			cursorY += lineAdvance
			cursorX = startX
			continue
		}
		glyph := font.Glyph(r)
		if glyph == nil {
			cursorX += advance
			continue
		}
		for row := 0; row < font.Height && row < len(glyph); row++ {
			line := glyph[row]
			for col := 0; col < font.Width && col < len(line); col++ {
				cell := line[col]
				if cell != '.' && cell != ' ' {
					c.plotPixel(cursorX+col, cursorY+row, c.strokeColor)
				}
			}
		}
		cursorX += advance
	}
}

// DrawImage draws an image without scaling.
func (c *Canvas) DrawImage(x, y int, img image.Image) {
	if c == nil || c.buffer == nil || img == nil {
		return
	}
	bounds := img.Bounds()
	if bounds.Empty() {
		return
	}
	for dy := 0; dy < bounds.Dy(); dy++ {
		for dx := 0; dx < bounds.Dx(); dx++ {
			r, g, b, a := img.At(bounds.Min.X+dx, bounds.Min.Y+dy).RGBA()
			alpha := float64(a) / 65535.0
			if alpha <= 0 {
				continue
			}
			color := backend.ColorRGB(uint8(r>>8), uint8(g>>8), uint8(b>>8))
			if alpha >= 1 {
				c.plotPixel(x+dx, y+dy, color)
			} else {
				c.blendPixel(x+dx, y+dy, color, alpha)
			}
		}
	}
}

// DrawImageScaled draws an image scaled to w×h using nearest neighbor sampling.
func (c *Canvas) DrawImageScaled(x, y, w, h int, img image.Image) {
	if c == nil || c.buffer == nil || img == nil || w <= 0 || h <= 0 {
		return
	}
	bounds := img.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()
	if srcW <= 0 || srcH <= 0 {
		return
	}
	for dy := 0; dy < h; dy++ {
		srcY := bounds.Min.Y + dy*srcH/h
		for dx := 0; dx < w; dx++ {
			srcX := bounds.Min.X + dx*srcW/w
			r, g, b, a := img.At(srcX, srcY).RGBA()
			alpha := float64(a) / 65535.0
			if alpha <= 0 {
				continue
			}
			color := backend.ColorRGB(uint8(r>>8), uint8(g>>8), uint8(b>>8))
			if alpha >= 1 {
				c.plotPixel(x+dx, y+dy, color)
			} else {
				c.blendPixel(x+dx, y+dy, color, alpha)
			}
		}
	}
}

// BeginPath starts a new path.
func (c *Canvas) BeginPath() {
	if c == nil {
		return
	}
	c.path = c.path[:0]
	c.pathHasCurrent = false
}

// MoveTo moves the current point.
func (c *Canvas) MoveTo(x, y int) {
	if c == nil {
		return
	}
	p := Point{X: x, Y: y}
	c.pathStart = p
	c.pathCurrent = p
	c.pathHasCurrent = true
	c.path = append(c.path, pathOp{kind: pathMove, p1: p})
}

// LineTo adds a line segment.
func (c *Canvas) LineTo(x, y int) {
	if c == nil {
		return
	}
	if !c.pathHasCurrent {
		c.MoveTo(x, y)
		return
	}
	p := Point{X: x, Y: y}
	c.pathCurrent = p
	c.path = append(c.path, pathOp{kind: pathLine, p1: p})
}

// QuadraticCurveTo adds a quadratic bezier segment.
func (c *Canvas) QuadraticCurveTo(cpx, cpy, x, y int) {
	if c == nil {
		return
	}
	if !c.pathHasCurrent {
		c.MoveTo(x, y)
		return
	}
	end := Point{X: x, Y: y}
	cp := Point{X: cpx, Y: cpy}
	c.pathCurrent = end
	c.path = append(c.path, pathOp{kind: pathQuad, p1: cp, p2: end})
}

// BezierCurveTo adds a cubic bezier segment.
func (c *Canvas) BezierCurveTo(cp1x, cp1y, cp2x, cp2y, x, y int) {
	if c == nil {
		return
	}
	if !c.pathHasCurrent {
		c.MoveTo(x, y)
		return
	}
	end := Point{X: x, Y: y}
	cp1 := Point{X: cp1x, Y: cp1y}
	cp2 := Point{X: cp2x, Y: cp2y}
	c.pathCurrent = end
	c.path = append(c.path, pathOp{kind: pathCubic, p1: cp1, p2: cp2, p3: end})
}

// ArcTo adds a rounded corner between two line segments.
func (c *Canvas) ArcTo(x1, y1, x2, y2, radius int) {
	if c == nil {
		return
	}
	if !c.pathHasCurrent {
		c.MoveTo(x1, y1)
		return
	}
	if radius <= 0 {
		c.LineTo(x1, y1)
		return
	}

	p0 := c.pathCurrent
	p1 := Point{X: x1, Y: y1}
	p2 := Point{X: x2, Y: y2}

	if pointsEqual(p0, p1) || pointsEqual(p1, p2) {
		c.LineTo(x1, y1)
		return
	}

	v1 := Vector2{X: float64(p0.X - p1.X), Y: float64(p0.Y - p1.Y)}
	v2 := Vector2{X: float64(p2.X - p1.X), Y: float64(p2.Y - p1.Y)}
	n1, ok1 := normalizeVec(v1)
	n2, ok2 := normalizeVec(v2)
	if !ok1 || !ok2 {
		c.LineTo(x1, y1)
		return
	}

	dot := clampFloat(dotVec(n1, n2), -1, 1)
	angle := math.Acos(dot)
	if angle == 0 {
		c.LineTo(x1, y1)
		return
	}

	t := float64(radius) / math.Tan(angle/2)
	if t <= 0 || math.IsInf(t, 0) || math.IsNaN(t) {
		c.LineTo(x1, y1)
		return
	}

	p1t := Vector2{X: float64(p1.X) + n1.X*t, Y: float64(p1.Y) + n1.Y*t}
	p2t := Vector2{X: float64(p1.X) + n2.X*t, Y: float64(p1.Y) + n2.Y*t}

	bisector := Vector2{X: n1.X + n2.X, Y: n1.Y + n2.Y}
	bisector, ok := normalizeVec(bisector)
	if !ok {
		c.LineTo(x1, y1)
		return
	}
	centerDist := float64(radius) / math.Sin(angle/2)
	center := Vector2{X: float64(p1.X) + bisector.X*centerDist, Y: float64(p1.Y) + bisector.Y*centerDist}

	c.LineTo(round(p1t.X), round(p1t.Y))

	start := math.Atan2(p1t.Y-center.Y, p1t.X-center.X)
	end := math.Atan2(p2t.Y-center.Y, p2t.X-center.X)
	clockwise := crossVec(n1, n2) > 0
	if clockwise {
		if end < start {
			end += 2 * math.Pi
		}
	} else if end > start {
		end -= 2 * math.Pi
	}

	points := arcSamplePoints(center.X, center.Y, float64(radius), start, end)
	for i := 1; i < len(points); i++ {
		c.LineTo(points[i].X, points[i].Y)
	}
}

// ClosePath closes the current sub-path.
func (c *Canvas) ClosePath() {
	if c == nil || !c.pathHasCurrent {
		return
	}
	c.pathCurrent = c.pathStart
	c.path = append(c.path, pathOp{kind: pathClose})
}

// Stroke renders the current path using the stroke color.
func (c *Canvas) Stroke() {
	if c == nil || c.buffer == nil || len(c.path) == 0 {
		return
	}
	var current Point
	var start Point
	hasCurrent := false
	for _, op := range c.path {
		switch op.kind {
		case pathMove:
			current = op.p1
			start = op.p1
			hasCurrent = true
		case pathLine:
			if !hasCurrent {
				current = op.p1
				start = op.p1
				hasCurrent = true
				continue
			}
			bresenhamLine(current.X, current.Y, op.p1.X, op.p1.Y, func(x, y int) {
				c.plotPixel(x, y, c.strokeColor)
			})
			current = op.p1
		case pathQuad:
			if !hasCurrent {
				current = op.p2
				start = op.p2
				hasCurrent = true
				continue
			}
			rasterQuadBezier(current, op.p1, op.p2, 0.5, func(x, y int) {
				c.plotPixel(x, y, c.strokeColor)
			})
			current = op.p2
		case pathCubic:
			if !hasCurrent {
				current = op.p3
				start = op.p3
				hasCurrent = true
				continue
			}
			rasterBezier(current, op.p1, op.p2, op.p3, 0.5, func(x, y int) {
				c.plotPixel(x, y, c.strokeColor)
			})
			current = op.p3
		case pathClose:
			if !hasCurrent {
				continue
			}
			bresenhamLine(current.X, current.Y, start.X, start.Y, func(x, y int) {
				c.plotPixel(x, y, c.strokeColor)
			})
			current = start
		}
	}
}

// Fill fills the current path using the fill color.
func (c *Canvas) Fill() {
	if c == nil || c.buffer == nil || len(c.path) == 0 {
		return
	}
	points := pathToPoints(c.path)
	if len(points) < 3 {
		return
	}
	fillPolygon(points, func(x, y int) {
		c.plotPixel(x, y, c.fillColor)
	})
}

// Render outputs the canvas to a RenderTarget at position (x, y).
func (c *Canvas) Render(target backend.RenderTarget, x, y int) {
	if c == nil || c.buffer == nil || c.blitter == nil || target == nil {
		return
	}
	if imageBlitter, ok := c.blitter.(ImageBlitter); ok {
		if imageTarget, ok := target.(backend.ImageTarget); ok {
			cellW, cellH := c.CellSize()
			image := imageBlitter.BuildImage(c.buffer, cellW, cellH)
			if image.Width > 0 && image.Height > 0 {
				imageTarget.SetImage(x, y, image)
				return
			}
		}
	}
	cellW, cellH := c.CellSize()
	for cy := 0; cy < cellH; cy++ {
		for cx := 0; cx < cellW; cx++ {
			char, style := c.blitter.BlitCell(c.buffer, cx, cy)
			target.SetContent(x+cx, y+cy, char, nil, style)
		}
	}
}
