package graphics

import (
	"math"
	"testing"
)

func TestPathToPoints(t *testing.T) {
	ops := []pathOp{
		{kind: pathMove, p1: Point{X: 0, Y: 0}},
		{kind: pathLine, p1: Point{X: 2, Y: 0}},
		{kind: pathLine, p1: Point{X: 2, Y: 0}},
		{kind: pathQuad, p1: Point{X: 3, Y: 1}, p2: Point{X: 4, Y: 2}},
		{kind: pathCubic, p1: Point{X: 5, Y: 2}, p2: Point{X: 6, Y: 3}, p3: Point{X: 7, Y: 4}},
		{kind: pathClose},
	}
	pts := pathToPoints(ops)
	if len(pts) == 0 {
		t.Fatalf("expected points")
	}
	if pts[0] != (Point{X: 0, Y: 0}) {
		t.Fatalf("unexpected start point")
	}
	if pts[len(pts)-1] != (Point{X: 0, Y: 0}) {
		t.Fatalf("expected closed path")
	}
}

func TestRasterPrimitives(t *testing.T) {
	linePoints := map[Point]struct{}{}
	bresenhamLine(0, 0, 2, 0, func(x, y int) {
		linePoints[Point{X: x, Y: y}] = struct{}{}
	})
	if len(linePoints) != 3 {
		t.Fatalf("expected 3 points, got %d", len(linePoints))
	}

	alphaCount := 0
	wuLine(0, 0, 2, 0, func(x, y int, alpha float64) {
		if alpha > 0 {
			alphaCount++
		}
	})
	if alphaCount == 0 {
		t.Fatalf("expected wu line samples")
	}

	circleCount := 0
	midpointCircle(0, 0, 2, func(x, y int) { circleCount++ })
	if circleCount == 0 {
		t.Fatalf("expected circle points")
	}

	fillCount := 0
	fillCircle(0, 0, 1, func(x, y int) { fillCount++ })
	if fillCount == 0 {
		t.Fatalf("expected filled circle points")
	}

	ellipseCount := 0
	midpointEllipse(0, 0, 2, 1, func(x, y int) { ellipseCount++ })
	if ellipseCount == 0 {
		t.Fatalf("expected ellipse points")
	}

	fillEllipseCount := 0
	fillEllipse(0, 0, 2, 1, func(x, y int) { fillEllipseCount++ })
	if fillEllipseCount == 0 {
		t.Fatalf("expected filled ellipse points")
	}

	arcCount := 0
	drawArc(0, 0, 3, 0, math.Pi/2, func(x, y int) { arcCount++ })
	if arcCount == 0 {
		t.Fatalf("expected arc points")
	}
}

func TestBezierAndSpline(t *testing.T) {
	bezierCount := 0
	rasterBezier(Point{X: 0, Y: 0}, Point{X: 2, Y: 3}, Point{X: 4, Y: 3}, Point{X: 6, Y: 0}, 0.5, func(x, y int) {
		bezierCount++
	})
	if bezierCount == 0 {
		t.Fatalf("expected bezier points")
	}

	quadCount := 0
	rasterQuadBezier(Point{X: 0, Y: 0}, Point{X: 2, Y: 2}, Point{X: 4, Y: 0}, 0.5, func(x, y int) {
		quadCount++
	})
	if quadCount == 0 {
		t.Fatalf("expected quad bezier points")
	}

	splineCount := 0
	points := []Point{{X: 0, Y: 0}, {X: 2, Y: 2}, {X: 4, Y: 0}}
	drawSpline(points, func(x, y int) { splineCount++ })
	if splineCount == 0 {
		t.Fatalf("expected spline points")
	}
}

func TestPolygonAndHelpers(t *testing.T) {
	polyCount := 0
	fillPolygon([]Point{{X: 0, Y: 0}, {X: 2, Y: 0}, {X: 1, Y: 2}}, func(x, y int) { polyCount++ })
	if polyCount == 0 {
		t.Fatalf("expected polygon fill")
	}

	if abs(-2) != 2 || sign(-2) != -1 || sign(0) != 0 || sign(2) != 1 {
		t.Fatalf("abs/sign helpers failed")
	}
	if round(1.4) != 1 || round(1.5) != 2 || round(-1.5) != -2 {
		t.Fatalf("round helper failed")
	}
	if fpart(1.25) != 0.25 || rfpart(1.25) != 0.75 {
		t.Fatalf("fraction helpers failed")
	}

	if clampRadius(0, 10, 10) != 0 {
		t.Fatalf("expected zero clamp")
	}
	if clampRadius(100, 10, 10) != 5 {
		t.Fatalf("expected clamp to half min size")
	}
	if circleOffset(2, 10) != 0 {
		t.Fatalf("expected circle offset clamp")
	}
}

func TestTransformAndVectorHelpers(t *testing.T) {
	id := IdentityTransform()
	if !id.IsIdentity() {
		t.Fatalf("expected identity")
	}
	x, y := id.Apply(1, 2)
	if x != 1 || y != 2 {
		t.Fatalf("identity apply failed")
	}

	tr := TranslateTransform(2, 3)
	sc := ScaleTransform(2, 2)
	combined := tr.Mul(sc)
	cx, cy := combined.Apply(1, 1)
	if cx == 1 && cy == 1 {
		t.Fatalf("expected transformed point")
	}

	rot := RotateTransform(math.Pi / 2)
	rx, ry := rot.Apply(1, 0)
	if math.Abs(rx) > 1e-6 || math.Abs(ry-1) > 1e-6 {
		t.Fatalf("rotate apply failed")
	}

	vec, ok := normalizeVec(Vector2{X: 0, Y: 0})
	if ok || vec != (Vector2{}) {
		t.Fatalf("expected zero vector normalization to fail")
	}
	vec, ok = normalizeVec(Vector2{X: 3, Y: 4})
	if !ok || math.Abs(vec.X-0.6) > 1e-6 || math.Abs(vec.Y-0.8) > 1e-6 {
		t.Fatalf("normalizeVec failed")
	}
	if dotVec(Vector2{X: 1, Y: 2}, Vector2{X: 3, Y: 4}) != 11 {
		t.Fatalf("dotVec failed")
	}
	if crossVec(Vector2{X: 1, Y: 0}, Vector2{X: 0, Y: 1}) != 1 {
		t.Fatalf("crossVec failed")
	}
	if clampFloat(-1, 0, 1) != 0 || clampFloat(2, 0, 1) != 1 {
		t.Fatalf("clampFloat failed")
	}
}

func TestPixelFontGlyph(t *testing.T) {
	if DefaultFont.Glyph('A') == nil {
		t.Fatalf("expected glyph for A")
	}
	if DefaultFont.Glyph('a') == nil {
		t.Fatalf("expected glyph for lowercase")
	}
	if DefaultFont.Glyph('$') == nil {
		t.Fatalf("expected fallback glyph")
	}
	var f *PixelFont
	if f.Glyph('A') != nil {
		t.Fatalf("expected nil glyph for nil font")
	}
}
