package graphics

import "github.com/odvcencio/fluffy-ui/backend"

// Canvas provides high-level drawing operations.
type Canvas struct {
	buffer  *PixelBuffer
	blitter Blitter

	strokeColor Color
	fillColor   Color
	lineWidth   int

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
	c.buffer.SetPixel(x, y, color)
}

// Blend blends a pixel color with alpha.
func (c *Canvas) Blend(x, y int, color Color, alpha float32) {
	if c == nil || c.buffer == nil {
		return
	}
	c.buffer.Blend(x, y, color, alpha)
}

// GetPixel returns the pixel at position.
func (c *Canvas) GetPixel(x, y int) Pixel {
	if c == nil || c.buffer == nil {
		return Pixel{}
	}
	return c.buffer.Get(x, y)
}

// DrawLine draws a line from (x1,y1) to (x2,y2).
func (c *Canvas) DrawLine(x1, y1, x2, y2 int) {
	if c == nil || c.buffer == nil {
		return
	}
	bresenhamLine(x1, y1, x2, y2, func(x, y int) {
		c.buffer.SetPixel(x, y, c.strokeColor)
	})
}

// DrawLineAA draws an anti-aliased line.
func (c *Canvas) DrawLineAA(x1, y1, x2, y2 int) {
	if c == nil || c.buffer == nil {
		return
	}
	wuLine(float64(x1), float64(y1), float64(x2), float64(y2), func(x, y int, alpha float64) {
		c.buffer.Blend(x, y, c.strokeColor, float32(alpha))
	})
}

// DrawRect draws a rectangle outline.
func (c *Canvas) DrawRect(x, y, w, h int) {
	if c == nil || c.buffer == nil || w <= 0 || h <= 0 {
		return
	}
	for dx := 0; dx < w; dx++ {
		c.buffer.SetPixel(x+dx, y, c.strokeColor)
		c.buffer.SetPixel(x+dx, y+h-1, c.strokeColor)
	}
	for dy := 0; dy < h; dy++ {
		c.buffer.SetPixel(x, y+dy, c.strokeColor)
		c.buffer.SetPixel(x+w-1, y+dy, c.strokeColor)
	}
}

// FillRect fills a rectangle.
func (c *Canvas) FillRect(x, y, w, h int) {
	if c == nil || c.buffer == nil || w <= 0 || h <= 0 {
		return
	}
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			c.buffer.SetPixel(x+dx, y+dy, c.fillColor)
		}
	}
}

// DrawEllipse draws an ellipse outline.
func (c *Canvas) DrawEllipse(cx, cy, rx, ry int) {
	if c == nil || c.buffer == nil {
		return
	}
	midpointEllipse(cx, cy, rx, ry, func(x, y int) {
		c.buffer.SetPixel(x, y, c.strokeColor)
	})
}

// FillEllipse fills an ellipse.
func (c *Canvas) FillEllipse(cx, cy, rx, ry int) {
	if c == nil || c.buffer == nil {
		return
	}
	fillEllipse(cx, cy, rx, ry, func(x, y int) {
		c.buffer.SetPixel(x, y, c.fillColor)
	})
}

// DrawCircle draws a circle outline.
func (c *Canvas) DrawCircle(cx, cy, radius int) {
	if c == nil || c.buffer == nil || radius <= 0 {
		return
	}
	midpointCircle(cx, cy, radius, func(x, y int) {
		c.buffer.SetPixel(x, y, c.strokeColor)
	})
}

// FillCircle fills a circle.
func (c *Canvas) FillCircle(cx, cy, radius int) {
	if c == nil || c.buffer == nil || radius <= 0 {
		return
	}
	fillCircle(cx, cy, radius, func(x, y int) {
		c.buffer.SetPixel(x, y, c.fillColor)
	})
}

// DrawBezier draws a cubic bezier curve.
func (c *Canvas) DrawBezier(p0, p1, p2, p3 Point) {
	if c == nil || c.buffer == nil {
		return
	}
	rasterBezier(p0, p1, p2, p3, 0.5, func(x, y int) {
		c.buffer.SetPixel(x, y, c.strokeColor)
	})
}

// DrawQuadBezier draws a quadratic bezier curve.
func (c *Canvas) DrawQuadBezier(p0, p1, p2 Point) {
	if c == nil || c.buffer == nil {
		return
	}
	rasterQuadBezier(p0, p1, p2, 0.5, func(x, y int) {
		c.buffer.SetPixel(x, y, c.strokeColor)
	})
}

// DrawSpline draws a Catmull-Rom spline through points.
func (c *Canvas) DrawSpline(points []Point) {
	if c == nil || c.buffer == nil {
		return
	}
	drawSpline(points, func(x, y int) {
		c.buffer.SetPixel(x, y, c.strokeColor)
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
			c.buffer.SetPixel(x, y, c.strokeColor)
		})
	}
}

// FillPolygon fills a polygon.
func (c *Canvas) FillPolygon(points []Point) {
	if c == nil || c.buffer == nil || len(points) < 3 {
		return
	}
	fillPolygon(points, func(x, y int) {
		c.buffer.SetPixel(x, y, c.fillColor)
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
				c.buffer.SetPixel(x, y, c.strokeColor)
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
				c.buffer.SetPixel(x, y, c.strokeColor)
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
				c.buffer.SetPixel(x, y, c.strokeColor)
			})
			current = op.p3
		case pathClose:
			if !hasCurrent {
				continue
			}
			bresenhamLine(current.X, current.Y, start.X, start.Y, func(x, y int) {
				c.buffer.SetPixel(x, y, c.strokeColor)
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
		c.buffer.SetPixel(x, y, c.fillColor)
	})
}

// Render outputs the canvas to a RenderTarget at position (x, y).
func (c *Canvas) Render(target backend.RenderTarget, x, y int) {
	if c == nil || c.buffer == nil || c.blitter == nil || target == nil {
		return
	}
	cellW, cellH := c.CellSize()
	for cy := 0; cy < cellH; cy++ {
		for cx := 0; cx < cellW; cx++ {
			char, style := c.blitter.BlitCell(c.buffer, cx, cy)
			target.SetContent(x+cx, y+cy, char, nil, style)
		}
	}
}
