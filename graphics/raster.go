package graphics

import (
	"math"
	"sort"
)

// bresenhamLine draws a line using Bresenham's algorithm.
func bresenhamLine(x1, y1, x2, y2 int, plot func(x, y int)) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx := sign(x2 - x1)
	sy := sign(y2 - y1)
	err := dx - dy

	for {
		plot(x1, y1)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
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

// wuLine draws an anti-aliased line using Xiaolin Wu's algorithm.
func wuLine(x0, y0, x1, y1 float64, plot func(x, y int, alpha float64)) {
	steep := math.Abs(y1-y0) > math.Abs(x1-x0)
	if steep {
		x0, y0 = y0, x0
		x1, y1 = y1, x1
	}
	if x0 > x1 {
		x0, x1 = x1, x0
		y0, y1 = y1, y0
	}

	dx := x1 - x0
	dy := y1 - y0
	gradient := 0.0
	if dx != 0 {
		gradient = dy / dx
	}

	xend := roundFloat(x0)
	yend := y0 + gradient*(xend-x0)
	xgap := rfpart(x0 + 0.5)
	xpxl1 := int(xend)
	ypxl1 := int(math.Floor(yend))

	if steep {
		plot(ypxl1, xpxl1, rfpart(yend)*xgap)
		plot(ypxl1+1, xpxl1, fpart(yend)*xgap)
	} else {
		plot(xpxl1, ypxl1, rfpart(yend)*xgap)
		plot(xpxl1, ypxl1+1, fpart(yend)*xgap)
	}

	intery := yend + gradient

	xend = roundFloat(x1)
	yend = y1 + gradient*(xend-x1)
	xgap = fpart(x1 + 0.5)
	xpxl2 := int(xend)
	ypxl2 := int(math.Floor(yend))

	if steep {
		plot(ypxl2, xpxl2, rfpart(yend)*xgap)
		plot(ypxl2+1, xpxl2, fpart(yend)*xgap)
	} else {
		plot(xpxl2, ypxl2, rfpart(yend)*xgap)
		plot(xpxl2, ypxl2+1, fpart(yend)*xgap)
	}

	if steep {
		for x := xpxl1 + 1; x < xpxl2; x++ {
			plot(int(math.Floor(intery)), x, rfpart(intery))
			plot(int(math.Floor(intery))+1, x, fpart(intery))
			intery += gradient
		}
	} else {
		for x := xpxl1 + 1; x < xpxl2; x++ {
			plot(x, int(math.Floor(intery)), rfpart(intery))
			plot(x, int(math.Floor(intery))+1, fpart(intery))
			intery += gradient
		}
	}
}

// midpointCircle draws a circle using the midpoint algorithm.
func midpointCircle(cx, cy, r int, plot func(x, y int)) {
	x := r
	y := 0
	err := 1 - r

	for x >= y {
		plot(cx+x, cy+y)
		plot(cx-x, cy+y)
		plot(cx+x, cy-y)
		plot(cx-x, cy-y)
		plot(cx+y, cy+x)
		plot(cx-y, cy+x)
		plot(cx+y, cy-x)
		plot(cx-y, cy-x)

		y++
		if err < 0 {
			err += 2*y + 1
		} else {
			x--
			err += 2*(y-x) + 1
		}
	}
}

// fillCircle fills a circle using scanlines.
func fillCircle(cx, cy, r int, plot func(x, y int)) {
	for y := -r; y <= r; y++ {
		dx := int(math.Sqrt(float64(r*r - y*y)))
		for x := -dx; x <= dx; x++ {
			plot(cx+x, cy+y)
		}
	}
}

// midpointEllipse draws an ellipse using the midpoint algorithm.
func midpointEllipse(cx, cy, rx, ry int, plot func(x, y int)) {
	if rx <= 0 || ry <= 0 {
		return
	}
	rx2 := rx * rx
	ry2 := ry * ry
	x := 0
	y := ry
	px := 0
	py := 2 * rx2 * y

	p := float64(ry2) - float64(rx2*ry) + float64(rx2)/4
	for px < py {
		plotEllipsePoints(cx, cy, x, y, plot)
		x++
		px += 2 * ry2
		if p < 0 {
			p += float64(ry2) + float64(px)
		} else {
			y--
			py -= 2 * rx2
			p += float64(ry2) + float64(px) - float64(py)
		}
	}

	p = float64(ry2)*(float64(x)+0.5)*(float64(x)+0.5) +
		float64(rx2)*(float64(y)-1)*(float64(y)-1) -
		float64(rx2*ry2)
	for y >= 0 {
		plotEllipsePoints(cx, cy, x, y, plot)
		y--
		py -= 2 * rx2
		if p > 0 {
			p += float64(rx2) - float64(py)
		} else {
			x++
			px += 2 * ry2
			p += float64(rx2) - float64(py) + float64(px)
		}
	}
}

// fillEllipse fills an ellipse using scanlines.
func fillEllipse(cx, cy, rx, ry int, plot func(x, y int)) {
	if rx <= 0 || ry <= 0 {
		return
	}
	for y := -ry; y <= ry; y++ {
		fy := float64(y)
		x := int(math.Round(float64(rx) * math.Sqrt(1-(fy*fy)/float64(ry*ry))))
		for dx := -x; dx <= x; dx++ {
			plot(cx+dx, cy+y)
		}
	}
}

func drawArc(cx, cy, radius int, startAngle, endAngle float64, plot func(x, y int)) {
	if radius <= 0 {
		return
	}
	points := arcSamplePoints(float64(cx), float64(cy), float64(radius), startAngle, endAngle)
	if len(points) == 0 {
		return
	}
	prev := points[0]
	for i := 1; i < len(points); i++ {
		curr := points[i]
		bresenhamLine(prev.X, prev.Y, curr.X, curr.Y, plot)
		prev = curr
	}
}

func arcSamplePoints(cx, cy, radius float64, startAngle, endAngle float64) []Point {
	delta := endAngle - startAngle
	if delta == 0 {
		return nil
	}
	steps := int(math.Ceil(math.Abs(delta) * radius))
	if steps < 1 {
		steps = 1
	}
	points := make([]Point, 0, steps+1)
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		angle := startAngle + delta*t
		x := round(cx + radius*math.Cos(angle))
		y := round(cy + radius*math.Sin(angle))
		points = appendPoint(points, Point{X: x, Y: y})
	}
	return points
}

func circleOffset(radius int, dy float64) int {
	term := float64(radius*radius) - dy*dy
	if term < 0 {
		term = 0
	}
	return int(math.Round(math.Sqrt(term)))
}

func clampRadius(radius, w, h int) int {
	if radius <= 0 {
		return 0
	}
	maxRadius := min(w, h) / 2
	if radius > maxRadius {
		return maxRadius
	}
	return radius
}

func plotEllipsePoints(cx, cy, x, y int, plot func(x, y int)) {
	plot(cx+x, cy+y)
	plot(cx-x, cy+y)
	plot(cx+x, cy-y)
	plot(cx-x, cy-y)
}

func midpoint(a, b Point) Point {
	return Point{X: (a.X + b.X) / 2, Y: (a.Y + b.Y) / 2}
}

// subdivideBezier splits a cubic bezier at t=0.5.
func subdivideBezier(p0, p1, p2, p3 Point) (q0, q1, q2, q3, r0, r1, r2, r3 Point) {
	m01 := midpoint(p0, p1)
	m12 := midpoint(p1, p2)
	m23 := midpoint(p2, p3)
	m012 := midpoint(m01, m12)
	m123 := midpoint(m12, m23)
	m0123 := midpoint(m012, m123)
	return p0, m01, m012, m0123, m0123, m123, m23, p3
}

// bezierFlatness returns the flatness of a bezier curve.
func bezierFlatness(p0, p1, p2, p3 Point) float64 {
	ux := 3*p1.X - 2*p0.X - p3.X
	uy := 3*p1.Y - 2*p0.Y - p3.Y
	vx := 3*p2.X - 2*p3.X - p0.X
	vy := 3*p2.Y - 2*p3.Y - p0.Y
	u := float64(ux*ux + uy*uy)
	v := float64(vx*vx + vy*vy)
	if u > v {
		return u
	}
	return v
}

const bezierMaxDepth = 16

// rasterBezier draws a bezier curve via recursive subdivision.
func rasterBezier(p0, p1, p2, p3 Point, tolerance float64, plot func(x, y int)) {
	rasterBezierDepth(p0, p1, p2, p3, tolerance, 0, plot)
}

func rasterBezierDepth(p0, p1, p2, p3 Point, tolerance float64, depth int, plot func(x, y int)) {
	if depth >= bezierMaxDepth || bezierFlatness(p0, p1, p2, p3) <= tolerance {
		bresenhamLine(p0.X, p0.Y, p3.X, p3.Y, plot)
		return
	}
	q0, q1, q2, q3, r0, r1, r2, r3 := subdivideBezier(p0, p1, p2, p3)
	if pointsEqual(p0, q0) && pointsEqual(p1, q1) && pointsEqual(p2, q2) && pointsEqual(p3, q3) {
		bresenhamLine(p0.X, p0.Y, p3.X, p3.Y, plot)
		return
	}
	rasterBezierDepth(q0, q1, q2, q3, tolerance, depth+1, plot)
	rasterBezierDepth(r0, r1, r2, r3, tolerance, depth+1, plot)
}

// rasterQuadBezier draws a quadratic bezier curve via cubic conversion.
func rasterQuadBezier(p0, p1, p2 Point, tolerance float64, plot func(x, y int)) {
	c1x := float64(p0.X) + (2.0/3.0)*float64(p1.X-p0.X)
	c1y := float64(p0.Y) + (2.0/3.0)*float64(p1.Y-p0.Y)
	c2x := float64(p2.X) + (2.0/3.0)*float64(p1.X-p2.X)
	c2y := float64(p2.Y) + (2.0/3.0)*float64(p1.Y-p2.Y)
	c1 := Point{X: round(c1x), Y: round(c1y)}
	c2 := Point{X: round(c2x), Y: round(c2y)}
	rasterBezier(p0, c1, c2, p2, tolerance, plot)
}

// drawSpline draws a Catmull-Rom spline through points.
func drawSpline(points []Point, plot func(x, y int)) {
	if len(points) < 2 {
		return
	}
	const steps = 16
	prev := points[0]
	for i := 0; i < len(points)-1; i++ {
		p0 := points[max(0, i-1)]
		p1 := points[i]
		p2 := points[i+1]
		p3 := points[min(len(points)-1, i+2)]
		for step := 1; step <= steps; step++ {
			t := float64(step) / float64(steps)
			x, y := catmullRom(p0, p1, p2, p3, t)
			curr := Point{X: round(x), Y: round(y)}
			bresenhamLine(prev.X, prev.Y, curr.X, curr.Y, plot)
			prev = curr
		}
	}
}

func catmullRom(p0, p1, p2, p3 Point, t float64) (float64, float64) {
	t2 := t * t
	t3 := t2 * t
	x := 0.5 * ((2 * float64(p1.X)) +
		(-float64(p0.X)+float64(p2.X))*t +
		(2*float64(p0.X)-5*float64(p1.X)+4*float64(p2.X)-float64(p3.X))*t2 +
		(-float64(p0.X)+3*float64(p1.X)-3*float64(p2.X)+float64(p3.X))*t3)
	y := 0.5 * ((2 * float64(p1.Y)) +
		(-float64(p0.Y)+float64(p2.Y))*t +
		(2*float64(p0.Y)-5*float64(p1.Y)+4*float64(p2.Y)-float64(p3.Y))*t2 +
		(-float64(p0.Y)+3*float64(p1.Y)-3*float64(p2.Y)+float64(p3.Y))*t3)
	return x, y
}

// fillPolygon fills a polygon using the scanline algorithm.
func fillPolygon(points []Point, plot func(x, y int)) {
	if len(points) < 3 {
		return
	}
	minY, maxY := points[0].Y, points[0].Y
	for _, p := range points[1:] {
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	for y := minY; y <= maxY; y++ {
		var intersections []int
		n := len(points)
		for i := 0; i < n; i++ {
			p1 := points[i]
			p2 := points[(i+1)%n]
			if (p1.Y <= y && p2.Y > y) || (p2.Y <= y && p1.Y > y) {
				x := p1.X + (y-p1.Y)*(p2.X-p1.X)/(p2.Y-p1.Y)
				intersections = append(intersections, x)
			}
		}
		sort.Ints(intersections)
		for i := 0; i+1 < len(intersections); i += 2 {
			for x := intersections[i]; x <= intersections[i+1]; x++ {
				plot(x, y)
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

func sign(v int) int {
	if v < 0 {
		return -1
	}
	if v > 0 {
		return 1
	}
	return 0
}

func round(x float64) int {
	if x < 0 {
		return int(math.Ceil(x - 0.5))
	}
	return int(math.Floor(x + 0.5))
}

func roundFloat(x float64) float64 {
	return float64(round(x))
}

func fpart(x float64) float64 {
	return x - math.Floor(x)
}

func rfpart(x float64) float64 {
	return 1 - fpart(x)
}

func pointsEqual(a, b Point) bool {
	return a.X == b.X && a.Y == b.Y
}