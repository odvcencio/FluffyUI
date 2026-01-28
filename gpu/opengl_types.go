//go:build !js

package gpu

import (
	"image"
	"unsafe"

	"github.com/odvcencio/fluffyui/gpu/opengl/gl"
)

type openglTexture struct {
	id     uint32
	width  int
	height int
	driver *openglDriver
}

func (t *openglTexture) ID() uint32 {
	if t == nil {
		return 0
	}
	return t.id
}

func (t *openglTexture) Size() (int, int) {
	if t == nil {
		return 0, 0
	}
	return t.width, t.height
}

func (t *openglTexture) Upload(pixels []byte, region image.Rectangle) {
	if t == nil || t.id == 0 || len(pixels) == 0 || t.driver == nil || t.driver.ctx == nil {
		return
	}
	t.driver.run(func() {
		t.upload(pixels, region)
	})
}

func (t *openglTexture) upload(pixels []byte, region image.Rectangle) {
	if t == nil || t.id == 0 || len(pixels) == 0 {
		return
	}
	w := t.width
	h := t.height
	if region.Empty() {
		region = image.Rect(0, 0, w, h)
	}
	if region.Min.X < 0 {
		region.Min.X = 0
	}
	if region.Min.Y < 0 {
		region.Min.Y = 0
	}
	if region.Max.X > w {
		region.Max.X = w
	}
	if region.Max.Y > h {
		region.Max.Y = h
	}
	gl.BindTexture(gl.TEXTURE_2D, t.id)
	ptr := unsafe.Pointer(&pixels[0])
	gl.TexSubImage2D(gl.TEXTURE_2D, 0, int32(region.Min.X), int32(region.Min.Y), int32(region.Dx()), int32(region.Dy()), gl.RGBA, gl.UNSIGNED_BYTE, ptr)
}

func (t *openglTexture) Dispose() {
	if t == nil || t.id == 0 {
		return
	}
	if t.driver == nil || t.driver.ctx == nil {
		return
	}
	t.driver.run(func() {
		t.dispose()
	})
}

func (t *openglTexture) dispose() {
	if t == nil || t.id == 0 {
		return
	}
	id := t.id
	gl.DeleteTextures(1, &id)
	t.id = 0
}

type openglFramebuffer struct {
	id     uint32
	width  int
	height int
	tex    *openglTexture
	driver *openglDriver
}

func (f *openglFramebuffer) ID() uint32 {
	if f == nil {
		return 0
	}
	return f.id
}

func (f *openglFramebuffer) Size() (int, int) {
	if f == nil {
		return 0, 0
	}
	return f.width, f.height
}

func (f *openglFramebuffer) Bind() {
	if f == nil {
		return
	}
	if f.driver == nil || f.driver.ctx == nil {
		return
	}
	f.driver.run(func() {
		f.bind()
	})
}

func (f *openglFramebuffer) bind() {
	if f == nil {
		return
	}
	gl.BindFramebuffer(gl.FRAMEBUFFER, f.id)
	gl.Viewport(0, 0, int32(f.width), int32(f.height))
}

func (f *openglFramebuffer) Texture() Texture {
	if f == nil {
		return nil
	}
	return f.tex
}

func (f *openglFramebuffer) Dispose() {
	if f == nil {
		return
	}
	if f.driver == nil || f.driver.ctx == nil {
		return
	}
	f.driver.run(func() {
		f.dispose()
	})
}

func (f *openglFramebuffer) dispose() {
	if f == nil {
		return
	}
	if f.id != 0 {
		id := f.id
		gl.DeleteFramebuffers(1, &id)
		f.id = 0
	}
	if f.tex != nil {
		f.tex.dispose()
		f.tex = nil
	}
}

type openglShader struct {
	id        uint32
	uniforms  map[string]uniformValue
	locations map[string]int32
	driver    *openglDriver
}

func (s *openglShader) ID() uint32 {
	if s == nil {
		return 0
	}
	return s.id
}

func (s *openglShader) SetUniform(name string, value any) {
	if s == nil {
		return
	}
	s.setUniform(name, value)
}

func (s *openglShader) Dispose() {
	if s == nil || s.id == 0 {
		return
	}
	if s.driver == nil || s.driver.ctx == nil {
		return
	}
	s.driver.run(func() {
		gl.DeleteProgram(s.id)
		s.id = 0
	})
}
