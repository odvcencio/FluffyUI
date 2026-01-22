package graphics

const pathSampleSteps = 16

type pathOpKind uint8

const (
	pathMove pathOpKind = iota
	pathLine
	pathQuad
	pathCubic
	pathClose
)

type pathOp struct {
	kind     pathOpKind
	p1, p2, p3 Point
}

func pathToPoints(ops []pathOp) []Point {
	var points []Point
	var current Point
	var start Point
	hasCurrent := false
	for _, op := range ops {
		switch op.kind {
		case pathMove:
			current = op.p1
			start = op.p1
			hasCurrent = true
			points = appendPoint(points, current)
		case pathLine:
			if !hasCurrent {
				current = op.p1
				start = op.p1
				hasCurrent = true
				points = appendPoint(points, current)
				continue
			}
			current = op.p1
			points = appendPoint(points, current)
		case pathQuad:
			if !hasCurrent {
				current = op.p2
				start = op.p2
				hasCurrent = true
				points = appendPoint(points, current)
				continue
			}
			points = appendQuadratic(points, current, op.p1, op.p2)
			current = op.p2
		case pathCubic:
			if !hasCurrent {
				current = op.p3
				start = op.p3
				hasCurrent = true
				points = appendPoint(points, current)
				continue
			}
			points = appendCubic(points, current, op.p1, op.p2, op.p3)
			current = op.p3
		case pathClose:
			if !hasCurrent {
				continue
			}
			current = start
			points = appendPoint(points, current)
		}
	}
	return points
}

func appendPoint(points []Point, p Point) []Point {
	if len(points) > 0 {
		last := points[len(points)-1]
		if last.X == p.X && last.Y == p.Y {
			return points
		}
	}
	return append(points, p)
}

func appendQuadratic(points []Point, p0, p1, p2 Point) []Point {
	for step := 1; step <= pathSampleSteps; step++ {
		t := float64(step) / float64(pathSampleSteps)
		inv := 1 - t
		x := inv*inv*float64(p0.X) + 2*inv*t*float64(p1.X) + t*t*float64(p2.X)
		y := inv*inv*float64(p0.Y) + 2*inv*t*float64(p1.Y) + t*t*float64(p2.Y)
		points = appendPoint(points, Point{X: round(x), Y: round(y)})
	}
	return points
}

func appendCubic(points []Point, p0, p1, p2, p3 Point) []Point {
	for step := 1; step <= pathSampleSteps; step++ {
		t := float64(step) / float64(pathSampleSteps)
		inv := 1 - t
		inv2 := inv * inv
		t2 := t * t
		x := inv2*inv*float64(p0.X) + 3*inv2*t*float64(p1.X) + 3*inv*t2*float64(p2.X) + t2*t*float64(p3.X)
		y := inv2*inv*float64(p0.Y) + 3*inv2*t*float64(p1.Y) + 3*inv*t2*float64(p2.Y) + t2*t*float64(p3.Y)
		points = appendPoint(points, Point{X: round(x), Y: round(y)})
	}
	return points
}
