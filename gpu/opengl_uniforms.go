package gpu

import (
	"image/color"
	"unsafe"

	"github.com/odvcencio/fluffy-ui/gpu/opengl/gl"
)

type uniformValue struct {
	kind uniformKind
	f    [4]float32
	m    [9]float32
	i    int32
	ptr  []float32
}

type uniformKind uint8

const (
	uniformInt uniformKind = iota
	uniformFloat
	uniformVec2
	uniformVec4
	uniformMat3
)

func (s *openglShader) setUniform(name string, value any) {
	if s == nil {
		return
	}
	if s.uniforms == nil {
		s.uniforms = make(map[string]uniformValue)
	}
	if u, ok := toUniformValue(value); ok {
		s.uniforms[name] = u
	}
}

func (s *openglShader) applyUniforms() {
	if s == nil || len(s.uniforms) == 0 {
		return
	}
	if s.locations == nil {
		s.locations = make(map[string]int32)
	}
	for name, val := range s.uniforms {
		loc, ok := s.locations[name]
		if !ok {
			cname := append([]byte(name), 0)
			loc = gl.GetUniformLocation(s.id, (*uint8)(unsafe.Pointer(&cname[0])))
			s.locations[name] = loc
		}
		if loc < 0 {
			continue
		}
		switch val.kind {
		case uniformInt:
			gl.Uniform1i(loc, val.i)
		case uniformFloat:
			gl.Uniform1f(loc, val.f[0])
		case uniformVec2:
			gl.Uniform2f(loc, val.f[0], val.f[1])
		case uniformVec4:
			gl.Uniform4f(loc, val.f[0], val.f[1], val.f[2], val.f[3])
		case uniformMat3:
			gl.UniformMatrix3fv(loc, 1, gl.TRUE, &val.m[0])
		}
	}
}

func toUniformValue(value any) (uniformValue, bool) {
	switch v := value.(type) {
	case int:
		return uniformValue{kind: uniformInt, i: int32(v)}, true
	case int32:
		return uniformValue{kind: uniformInt, i: v}, true
	case uint32:
		return uniformValue{kind: uniformInt, i: int32(v)}, true
	case float32:
		return uniformValue{kind: uniformFloat, f: [4]float32{v}}, true
	case float64:
		return uniformValue{kind: uniformFloat, f: [4]float32{float32(v)}}, true
	case [2]float32:
		return uniformValue{kind: uniformVec2, f: [4]float32{v[0], v[1]}}, true
	case [2]float64:
		return uniformValue{kind: uniformVec2, f: [4]float32{float32(v[0]), float32(v[1])}}, true
	case []float32:
		return sliceUniform(v)
	case [4]float32:
		return uniformValue{kind: uniformVec4, f: v}, true
	case color.RGBA:
		return uniformValue{kind: uniformVec4, f: [4]float32{float32(v.R) / 255, float32(v.G) / 255, float32(v.B) / 255, float32(v.A) / 255}}, true
	case Matrix3:
		return uniformValue{kind: uniformMat3, m: v.m}, true
	case [9]float32:
		return uniformValue{kind: uniformMat3, m: v}, true
	default:
		return uniformValue{}, false
	}
}

func sliceUniform(values []float32) (uniformValue, bool) {
	switch len(values) {
	case 2:
		return uniformValue{kind: uniformVec2, f: [4]float32{values[0], values[1]}}, true
	case 4:
		return uniformValue{kind: uniformVec4, f: [4]float32{values[0], values[1], values[2], values[3]}}, true
	case 9:
		var m [9]float32
		copy(m[:], values)
		return uniformValue{kind: uniformMat3, m: m}, true
	default:
		return uniformValue{}, false
	}
}
