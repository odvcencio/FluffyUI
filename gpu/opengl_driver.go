//go:build !js

package gpu

import (
	"errors"
	"image"
	"runtime"
	"unsafe"

	"github.com/odvcencio/fluffyui/gpu/opengl/gl"
)

type openglDriver struct {
	ctx     *osmesaContext
	maxTex  int
	worker  *glWorker
	vao     uint32
	vbo     uint32
	ebo     uint32
	vboCap  int
	eboCap  int
	readFBO uint32
}

func newOpenGLDriver() (Driver, error) {
	if runtime.GOOS == "darwin" {
		return nil, ErrUnsupported
	}
	driver := &openglDriver{maxTex: 8192}
	driver.worker = newGLWorker()
	var initErr error
	driver.run(func() {
		// Create OSMesa context first to ensure the library is available
		// before initializing GL function pointers. This prevents segfaults
		// when OSMesa is not installed.
		ctx, err := newOSMesaContext(1, 1)
		if err != nil {
			initErr = ErrUnsupported
			return
		}
		driver.ctx = ctx
		if err := gl.Init(); err != nil {
			driver.ctx.Destroy()
			driver.ctx = nil
			initErr = err
			return
		}
		gl.Viewport(0, 0, 1, 1)
		gl.GenVertexArrays(1, &driver.vao)
		gl.GenBuffers(1, &driver.vbo)
		gl.GenBuffers(1, &driver.ebo)
	})
	if initErr != nil {
		if driver.worker != nil {
			driver.worker.stop()
		}
		return nil, initErr
	}
	return driver, nil
}

func (d *openglDriver) Backend() Backend {
	return BackendOpenGL
}

func (d *openglDriver) Init() error {
	if d == nil || d.ctx == nil {
		return ErrUnsupported
	}
	return nil
}

func (d *openglDriver) Dispose() {
	if d == nil {
		return
	}
	if d.worker != nil {
		d.run(func() {
			if d.vbo != 0 {
				gl.DeleteBuffers(1, &d.vbo)
				d.vbo = 0
				d.vboCap = 0
			}
			if d.ebo != 0 {
				gl.DeleteBuffers(1, &d.ebo)
				d.ebo = 0
				d.eboCap = 0
			}
			if d.vao != 0 {
				gl.DeleteVertexArrays(1, &d.vao)
				d.vao = 0
			}
			if d.readFBO != 0 {
				gl.DeleteFramebuffers(1, &d.readFBO)
				d.readFBO = 0
			}
			if d.ctx != nil {
				d.ctx.Destroy()
				d.ctx = nil
			}
		})
		d.worker.stop()
		d.worker = nil
		return
	}
	if d.ctx != nil {
		d.ctx.Destroy()
		d.ctx = nil
	}
}

func (d *openglDriver) run(fn func()) {
	if d == nil {
		if fn != nil {
			fn()
		}
		return
	}
	if d.worker != nil {
		d.worker.run(fn)
		return
	}
	if fn != nil {
		fn()
	}
}

func (d *openglDriver) NewTexture(width, height int) (Texture, error) {
	if d == nil || d.ctx == nil {
		return nil, ErrUnsupported
	}
	if width <= 0 || height <= 0 {
		return nil, ErrUnsupported
	}
	var tex *openglTexture
	var err error
	d.run(func() {
		tex, err = d.newTexture(width, height)
	})
	return tex, err
}

func (d *openglDriver) NewFramebuffer(width, height int) (Framebuffer, error) {
	if d == nil || d.ctx == nil {
		return nil, ErrUnsupported
	}
	if width <= 0 || height <= 0 {
		return nil, ErrUnsupported
	}
	var fb *openglFramebuffer
	var err error
	d.run(func() {
		if d.ctx == nil {
			err = ErrUnsupported
			return
		}
		if resizeErr := d.ctx.Resize(width, height); resizeErr != nil {
			err = resizeErr
			return
		}
		gl.Viewport(0, 0, int32(width), int32(height))
		tex, texErr := d.newTexture(width, height)
		if texErr != nil {
			err = texErr
			return
		}
		var fbo uint32
		gl.GenFramebuffers(1, &fbo)
		gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, tex.id, 0)
		status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER)
		if status != gl.FRAMEBUFFER_COMPLETE {
			gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
			gl.DeleteFramebuffers(1, &fbo)
			if tex != nil {
				tex.dispose()
			}
			err = errors.New("opengl: framebuffer incomplete")
			return
		}
		fb = &openglFramebuffer{id: fbo, width: width, height: height, tex: tex, driver: d}
	})
	if err != nil {
		return nil, err
	}
	return fb, nil
}

func (d *openglDriver) NewShader(src ShaderSource) (Shader, error) {
	if d == nil || d.ctx == nil {
		return nil, ErrUnsupported
	}
	var shader *openglShader
	var err error
	d.run(func() {
		vertex := src.GLSL.Vertex
		fragment := src.GLSL.Fragment
		if vertex == "" || fragment == "" {
			err = errors.New("opengl: missing shader source")
			return
		}
		vs, shaderErr := compileShader(vertex, gl.VERTEX_SHADER)
		if shaderErr != nil {
			err = shaderErr
			return
		}
		fs, shaderErr := compileShader(fragment, gl.FRAGMENT_SHADER)
		if shaderErr != nil {
			gl.DeleteShader(vs)
			err = shaderErr
			return
		}
		program := gl.CreateProgram()
		gl.AttachShader(program, vs)
		gl.AttachShader(program, fs)
		gl.LinkProgram(program)
		if linkErr := checkProgram(program); linkErr != nil {
			gl.DeleteShader(vs)
			gl.DeleteShader(fs)
			gl.DeleteProgram(program)
			err = linkErr
			return
		}
		gl.DeleteShader(vs)
		gl.DeleteShader(fs)
		shader = &openglShader{id: program, driver: d}
	})
	if err != nil {
		return nil, err
	}
	return shader, nil
}

func (d *openglDriver) Clear(r, g, b, a float32) {
	if d == nil || d.ctx == nil {
		return
	}
	d.run(func() {
		gl.ClearColor(r, g, b, a)
		gl.Clear(gl.COLOR_BUFFER_BIT)
	})
}

func (d *openglDriver) Draw(call DrawCall) {
	if d == nil || d.ctx == nil {
		return
	}
	d.run(func() {
		d.draw(call)
	})
}

func (d *openglDriver) ReadPixels(fb Framebuffer, rect image.Rectangle) ([]byte, error) {
	if d == nil || d.ctx == nil {
		return nil, ErrUnsupported
	}
	var pixels []byte
	var err error
	d.run(func() {
		pixels, err = d.readPixels(fb, rect)
	})
	return pixels, err
}

func (d *openglDriver) MaxTextureSize() int {
	if d == nil {
		return 0
	}
	return d.maxTex
}

func (d *openglDriver) newTexture(width, height int) (*openglTexture, error) {
	if d == nil || d.ctx == nil {
		return nil, ErrUnsupported
	}
	if width <= 0 || height <= 0 {
		return nil, ErrUnsupported
	}
	var id uint32
	gl.GenTextures(1, &id)
	gl.BindTexture(gl.TEXTURE_2D, id)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(gl.TEXTURE_2D, 0, int32(gl.RGBA), int32(width), int32(height), 0, gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(uintptr(0)))
	return &openglTexture{id: id, width: width, height: height, driver: d}, nil
}

func (d *openglDriver) draw(call DrawCall) {
	if d == nil || d.ctx == nil {
		return
	}
	if call.Target != nil {
		if fb, ok := call.Target.(*openglFramebuffer); ok {
			fb.bind()
		}
	}
	shader, ok := call.Shader.(*openglShader)
	if !ok || shader == nil {
		return
	}
	gl.UseProgram(shader.id)
	shader.applyUniforms()
	if len(call.Vertices) == 0 || len(call.Indices) == 0 {
		return
	}
	if call.Scissor != nil {
		gl.Enable(gl.SCISSOR_TEST)
		gl.Scissor(int32(call.Scissor.Min.X), int32(call.Scissor.Min.Y), int32(call.Scissor.Dx()), int32(call.Scissor.Dy()))
	}
	switch call.Blend {
	case BlendAlpha:
		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	case BlendAdditive:
		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
	default:
		gl.Disable(gl.BLEND)
	}
	if len(call.Textures) > 0 {
		if tex, ok := call.Textures[0].(*openglTexture); ok {
			gl.ActiveTexture(gl.TEXTURE0)
			gl.BindTexture(gl.TEXTURE_2D, tex.id)
		}
	}
	if d.vao == 0 {
		gl.GenVertexArrays(1, &d.vao)
	}
	gl.BindVertexArray(d.vao)
	vertexBytes := len(call.Vertices) * 4
	indexBytes := len(call.Indices) * 2
	if d.vbo == 0 {
		gl.GenBuffers(1, &d.vbo)
	}
	gl.BindBuffer(gl.ARRAY_BUFFER, d.vbo)
	if vertexBytes > d.vboCap {
		gl.BufferData(gl.ARRAY_BUFFER, uintptr(vertexBytes), gl.Ptr(call.Vertices), gl.DYNAMIC_DRAW)
		d.vboCap = vertexBytes
	} else {
		gl.BufferSubData(gl.ARRAY_BUFFER, 0, uintptr(vertexBytes), gl.Ptr(call.Vertices))
	}
	if d.ebo == 0 {
		gl.GenBuffers(1, &d.ebo)
	}
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, d.ebo)
	if indexBytes > d.eboCap {
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, uintptr(indexBytes), gl.Ptr(call.Indices), gl.DYNAMIC_DRAW)
		d.eboCap = indexBytes
	} else {
		gl.BufferSubData(gl.ELEMENT_ARRAY_BUFFER, 0, uintptr(indexBytes), gl.Ptr(call.Indices))
	}
	if call.Layout != nil && len(call.Layout.Attributes) > 0 {
		applyLayout(call.Layout)
	} else {
		layout := inferLayout(call.Vertices)
		offset := uintptr(0)
		if layout.posSize > 0 {
			gl.EnableVertexAttribArray(0)
			gl.VertexAttribPointer(0, int32(layout.posSize), gl.FLOAT, gl.FALSE, int32(layout.stride*4), unsafe.Pointer(offset))
			offset += uintptr(layout.posSize * 4)
		}
		if layout.uvSize > 0 {
			gl.EnableVertexAttribArray(1)
			gl.VertexAttribPointer(1, int32(layout.uvSize), gl.FLOAT, gl.FALSE, int32(layout.stride*4), unsafe.Pointer(offset))
			offset += uintptr(layout.uvSize * 4)
		}
		if layout.colorSize > 0 {
			gl.EnableVertexAttribArray(2)
			gl.VertexAttribPointer(2, int32(layout.colorSize), gl.FLOAT, gl.FALSE, int32(layout.stride*4), unsafe.Pointer(offset))
		}
	}
	gl.DrawElements(gl.TRIANGLES, int32(len(call.Indices)), gl.UNSIGNED_SHORT, unsafe.Pointer(uintptr(0)))
	if call.Scissor != nil {
		gl.Disable(gl.SCISSOR_TEST)
	}
}

func (d *openglDriver) readPixels(fb Framebuffer, rect image.Rectangle) ([]byte, error) {
	if d == nil || d.ctx == nil {
		return nil, ErrUnsupported
	}
	width := d.ctx.width
	height := d.ctx.height
	if fb != nil {
		glfb, ok := fb.(*openglFramebuffer)
		if !ok || glfb == nil {
			return nil, ErrUnsupported
		}
		width = glfb.width
		height = glfb.height
		glfb.bind()
	}
	if rect.Empty() {
		rect = image.Rect(0, 0, width, height)
	}
	if rect.Min.X < 0 {
		rect.Min.X = 0
	}
	if rect.Min.Y < 0 {
		rect.Min.Y = 0
	}
	if rect.Max.X > width {
		rect.Max.X = width
	}
	if rect.Max.Y > height {
		rect.Max.Y = height
	}
	w := rect.Dx()
	h := rect.Dy()
	if w <= 0 || h <= 0 {
		return nil, nil
	}
	pixels := make([]byte, w*h*4)
	gl.ReadPixels(int32(rect.Min.X), int32(rect.Min.Y), int32(w), int32(h), gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&pixels[0]))
	flipPixelsVertical(pixels, w, h)
	return pixels, nil
}

func (d *openglDriver) ReadTexturePixels(tex Texture, rect image.Rectangle) ([]byte, int, int, error) {
	if d == nil || d.ctx == nil {
		return nil, 0, 0, ErrUnsupported
	}
	gltex, ok := tex.(*openglTexture)
	if !ok || gltex == nil {
		return nil, 0, 0, ErrUnsupported
	}
	var pixels []byte
	var w, h int
	var err error
	d.run(func() {
		pixels, w, h, err = d.readTexturePixels(gltex, rect)
	})
	return pixels, w, h, err
}

func (d *openglDriver) readTexturePixels(tex *openglTexture, rect image.Rectangle) ([]byte, int, int, error) {
	if tex == nil {
		return nil, 0, 0, ErrUnsupported
	}
	width := tex.width
	height := tex.height
	if rect.Empty() {
		rect = image.Rect(0, 0, width, height)
	}
	if rect.Min.X < 0 {
		rect.Min.X = 0
	}
	if rect.Min.Y < 0 {
		rect.Min.Y = 0
	}
	if rect.Max.X > width {
		rect.Max.X = width
	}
	if rect.Max.Y > height {
		rect.Max.Y = height
	}
	w := rect.Dx()
	h := rect.Dy()
	if w <= 0 || h <= 0 {
		return nil, 0, 0, nil
	}
	if d.readFBO == 0 {
		gl.GenFramebuffers(1, &d.readFBO)
	}
	gl.BindFramebuffer(gl.FRAMEBUFFER, d.readFBO)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, tex.id, 0)
	status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER)
	if status != gl.FRAMEBUFFER_COMPLETE {
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
		return nil, 0, 0, errors.New("opengl: readback framebuffer incomplete")
	}
	pixels := make([]byte, w*h*4)
	gl.ReadPixels(int32(rect.Min.X), int32(rect.Min.Y), int32(w), int32(h), gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&pixels[0]))
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	flipPixelsVertical(pixels, w, h)
	return pixels, w, h, nil
}

func compileShader(src string, typ uint32) (uint32, error) {
	shader := gl.CreateShader(typ)
	cstr := append([]byte(src), 0)
	ptr := (*uint8)(unsafe.Pointer(&cstr[0]))
	gl.ShaderSource(shader, 1, &ptr, nil)
	gl.CompileShader(shader)
	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		log, _ := shaderInfoLog(shader)
		gl.DeleteShader(shader)
		return 0, errors.New(log)
	}
	return shader, nil
}

func shaderInfoLog(shader uint32) (string, error) {
	var length int32
	gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &length)
	if length <= 0 {
		return "", nil
	}
	buf := make([]byte, length)
	gl.GetShaderInfoLog(shader, length, nil, &buf[0])
	return string(buf), nil
}

func checkProgram(program uint32) error {
	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		log, _ := programInfoLog(program)
		return errors.New(log)
	}
	return nil
}

func programInfoLog(program uint32) (string, error) {
	var length int32
	gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &length)
	if length <= 0 {
		return "", nil
	}
	buf := make([]byte, length)
	gl.GetProgramInfoLog(program, length, nil, &buf[0])
	return string(buf), nil
}


