package gpu

import (
	"errors"
	"image"
	"testing"
)

type stubShader struct {
	id       uint32
	uniforms map[string]any
}

func (s *stubShader) ID() uint32 { return s.id }
func (s *stubShader) SetUniform(name string, value any) {
	if s.uniforms == nil {
		s.uniforms = make(map[string]any)
	}
	s.uniforms[name] = value
}
func (s *stubShader) Dispose() {}

type stubDriver struct {
	shaders int
}

func (d *stubDriver) Backend() Backend { return BackendSoftware }
func (d *stubDriver) Init() error      { return nil }
func (d *stubDriver) Dispose()         {}
func (d *stubDriver) NewTexture(width, height int) (Texture, error) {
	return nil, ErrUnsupported
}
func (d *stubDriver) NewFramebuffer(width, height int) (Framebuffer, error) {
	return nil, ErrUnsupported
}
func (d *stubDriver) NewShader(src ShaderSource) (Shader, error) {
	d.shaders++
	return &stubShader{id: uint32(d.shaders)}, nil
}
func (d *stubDriver) Clear(r, g, b, a float32) {}
func (d *stubDriver) Draw(call DrawCall)       {}
func (d *stubDriver) ReadPixels(fb Framebuffer, rect image.Rectangle) ([]byte, error) {
	return nil, ErrUnsupported
}
func (d *stubDriver) MaxTextureSize() int { return 0 }

func TestMatrixHelpers(t *testing.T) {
	id := Identity()
	if !id.IsIdentity() {
		t.Fatalf("expected identity")
	}
	tr := Translate(2, 3)
	sx, sy := tr.Apply(1, 1)
	if sx != 3 || sy != 4 {
		t.Fatalf("translate apply failed")
	}
	sc := Scale(2, 3)
	sx, sy = sc.Apply(1, 1)
	if sx != 2 || sy != 3 {
		t.Fatalf("scale apply failed")
	}
	rot := Rotate(0)
	sx, sy = rot.Apply(1, 2)
	if sx != 1 || sy != 2 {
		t.Fatalf("rotate apply failed")
	}
	mul := tr.Mul(sc)
	mx, my := mul.Apply(1, 1)
	if mx == 1 && my == 1 {
		t.Fatalf("expected composed transform")
	}
}

func TestPathToPointsAndBezier(t *testing.T) {
	ops := []pathOp{
		{kind: pathMove, p1: vec2{x: 0, y: 0}},
		{kind: pathLine, p1: vec2{x: 1, y: 0}},
		{kind: pathQuad, p1: vec2{x: 2, y: 1}, p2: vec2{x: 3, y: 0}},
		{kind: pathCubic, p1: vec2{x: 4, y: 1}, p2: vec2{x: 5, y: 1}, p3: vec2{x: 6, y: 0}},
		{kind: pathClose},
	}
	points := pathToPoints(ops)
	if len(points) == 0 {
		t.Fatalf("expected points")
	}
	if quadBezier(vec2{}, vec2{x: 1}, vec2{x: 2}, 0).x != 0 {
		t.Fatalf("expected quad t=0 to return start")
	}
	if cubicBezier(vec2{}, vec2{x: 1}, vec2{x: 2}, vec2{x: 3}, 1).x != 3 {
		t.Fatalf("expected cubic t=1 to return end")
	}
}

func TestShaderHelpers(t *testing.T) {
	if normalizeShaderName("shaders/solid.glsl") != "solid" {
		t.Fatalf("normalizeShaderName failed")
	}
	if normalizeShaderName(" ") != "" {
		t.Fatalf("expected empty name")
	}
	_, _, err := splitGLSL("// Vertex shader\nvoid main() {}\n")
	if err == nil {
		t.Fatalf("expected split error")
	}
	vertex, fragment, err := splitGLSL("// Vertex shader\nvoid main() {}\n// Fragment shader\nvoid main() {}\n")
	if err != nil || vertex == "" || fragment == "" {
		t.Fatalf("expected split success")
	}
	if _, err := LoadShaderSource(" "); err == nil {
		t.Fatalf("expected error for empty shader name")
	}
	if _, err := LoadShaderSource("missing"); err == nil {
		t.Fatalf("expected error for missing shader")
	}
}

func TestGPUPipeline(t *testing.T) {
	if _, err := newGPUPipeline(nil, 1, 1); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("expected unsupported for nil driver")
	}
	d := &stubDriver{}
	pipe, err := newGPUPipeline(d, 100, 50)
	if err != nil {
		t.Fatalf("newGPUPipeline: %v", err)
	}
	if pipe.projection.IsIdentity() {
		t.Fatalf("expected projection set")
	}
	pipe.SetProjection(0, 0)
	if !pipe.projection.IsIdentity() {
		t.Fatalf("expected identity projection")
	}
	pipe.Dispose()
}

func TestProjectionMatrix(t *testing.T) {
	proj := projectionMatrix(0, 0)
	if !proj.IsIdentity() {
		t.Fatalf("expected identity projection")
	}
	proj = projectionMatrix(10, 5)
	x, y := proj.Apply(0, 0)
	if x >= 0 || y <= 0 {
		t.Fatalf("unexpected projection values")
	}
}

func TestDrawBatch(t *testing.T) {
	batch := &drawBatch{}
	batch.Add(DrawCall{Blend: BlendAlpha})
	if len(batch.Calls()) != 1 {
		t.Fatalf("expected 1 call")
	}
	batch.Reset()
	if len(batch.Calls()) != 0 {
		t.Fatalf("expected empty calls")
	}
	batch.Reset()
	batch.Add(DrawCall{Blend: BlendAdditive})
	if batch.Calls()[0].Blend != BlendAdditive {
		t.Fatalf("unexpected call")
	}
}
