package gpu

const pathSampleSteps = 16

type pathOpKind uint8

const (
	pathMove pathOpKind = iota
	pathLine
	pathQuad
	pathCubic
	pathClose
)

type vec2 struct {
	x float32
	y float32
}

type pathOp struct {
	kind       pathOpKind
	p1, p2, p3 vec2
}

func pathToPoints(ops []pathOp) []vec2 {
	if len(ops) == 0 {
		return nil
	}
	points := make([]vec2, 0, len(ops)*pathSampleSteps)
	var current vec2
	var start vec2
	for _, op := range ops {
		switch op.kind {
		case pathMove:
			current = op.p1
			start = op.p1
			points = append(points, current)
		case pathLine:
			current = op.p1
			points = append(points, current)
		case pathQuad:
			for step := 1; step <= pathSampleSteps; step++ {
				t := float32(step) / float32(pathSampleSteps)
				points = append(points, quadBezier(current, op.p1, op.p2, t))
			}
			current = op.p2
		case pathCubic:
			for step := 1; step <= pathSampleSteps; step++ {
				t := float32(step) / float32(pathSampleSteps)
				points = append(points, cubicBezier(current, op.p1, op.p2, op.p3, t))
			}
			current = op.p3
		case pathClose:
			points = append(points, start)
			current = start
		}
	}
	return points
}

func quadBezier(p0, p1, p2 vec2, t float32) vec2 {
	inv := 1 - t
	return vec2{
		x: inv*inv*p0.x + 2*inv*t*p1.x + t*t*p2.x,
		y: inv*inv*p0.y + 2*inv*t*p1.y + t*t*p2.y,
	}
}

func cubicBezier(p0, p1, p2, p3 vec2, t float32) vec2 {
	inv := 1 - t
	inv2 := inv * inv
	t2 := t * t
	return vec2{
		x: inv2*inv*p0.x + 3*inv2*t*p1.x + 3*inv*t2*p2.x + t2*t*p3.x,
		y: inv2*inv*p0.y + 3*inv2*t*p1.y + 3*inv*t2*p2.y + t2*t*p3.y,
	}
}
