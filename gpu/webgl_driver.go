//go:build js

package gpu

import (
	"errors"
	"image"
	"syscall/js"
)

// webglDriver implements the Driver interface using WebGL via syscall/js.
type webglDriver struct {
	gl       js.Value
	canvas   js.Value
	maxTex   int
	vaoExt   js.Value

	// Reusable buffers to reduce allocations
	tempBuf  js.Value
	indexBuf js.Value
}

// newWebGLDriver creates a new WebGL driver.
func newWebGLDriver() (Driver, error) {
	// Get the WebGL context from the global document
	doc := js.Global().Get("document")
	if doc.IsUndefined() {
		return nil, ErrUnsupported
	}

	// Try to find an existing canvas or create a dummy one
	canvas := doc.Call("getElementById", "fluffyui-canvas")
	if canvas.IsNull() {
		canvas = doc.Call("createElement", "canvas")
		canvas.Set("width", 1)
		canvas.Set("height", 1)
	}

	// Try WebGL2 first, then fall back to WebGL1
	gl := canvas.Call("getContext", "webgl2")
	if gl.IsNull() {
		gl = canvas.Call("getContext", "webgl")
	}
	if gl.IsNull() {
		gl = canvas.Call("getContext", "experimental-webgl")
	}
	if gl.IsNull() {
		return nil, ErrUnsupported
	}

	d := &webglDriver{
		gl:     gl,
		canvas: canvas,
		maxTex: 4096,
	}

	// Check for VAO extension in WebGL1
	if gl.Get("createVertexArray").IsUndefined() {
		ext := gl.Call("getExtension", "OES_vertex_array_object")
		if !ext.IsNull() {
			d.vaoExt = ext
		}
	}

	// Create reusable buffers
	d.tempBuf = js.Global().Get("ArrayBuffer").New(65536)
	d.indexBuf = js.Global().Get("ArrayBuffer").New(65536)

	return d, nil
}

func (d *webglDriver) Backend() Backend {
	return BackendWebGL
}

func (d *webglDriver) Init() error {
	if d == nil || d.gl.IsUndefined() {
		return ErrUnsupported
	}
	return nil
}

func (d *webglDriver) Dispose() {
	if d == nil {
		return
	}
	d.gl = js.Value{}
	d.canvas = js.Value{}
}

func (d *webglDriver) NewTexture(width, height int) (Texture, error) {
	if d == nil || d.gl.IsUndefined() {
		return nil, ErrUnsupported
	}
	if width <= 0 || height <= 0 {
		return nil, ErrUnsupported
	}

	gl := d.gl
	tex := gl.Call("createTexture")
	if tex.IsNull() {
		return nil, errors.New("webgl: failed to create texture")
	}

	gl.Call("bindTexture", gl.Get("TEXTURE_2D"), tex)
	gl.Call("texParameteri", gl.Get("TEXTURE_2D"), gl.Get("TEXTURE_MIN_FILTER"), gl.Get("LINEAR"))
	gl.Call("texParameteri", gl.Get("TEXTURE_2D"), gl.Get("TEXTURE_MAG_FILTER"), gl.Get("LINEAR"))
	gl.Call("texParameteri", gl.Get("TEXTURE_2D"), gl.Get("TEXTURE_WRAP_S"), gl.Get("CLAMP_TO_EDGE"))
	gl.Call("texParameteri", gl.Get("TEXTURE_2D"), gl.Get("TEXTURE_WRAP_T"), gl.Get("CLAMP_TO_EDGE"))

	// Allocate texture storage
	gl.Call("texImage2D", gl.Get("TEXTURE_2D"), 0, gl.Get("RGBA"), width, height, 0, gl.Get("RGBA"), gl.Get("UNSIGNED_BYTE"), js.Null())

	return &webglTexture{
		id:     tex,
		width:  width,
		height: height,
		driver: d,
	}, nil
}

func (d *webglDriver) NewFramebuffer(width, height int) (Framebuffer, error) {
	if d == nil || d.gl.IsUndefined() {
		return nil, ErrUnsupported
	}
	if width <= 0 || height <= 0 {
		return nil, ErrUnsupported
	}

	// Create texture for the framebuffer
	tex, err := d.NewTexture(width, height)
	if err != nil {
		return nil, err
	}
	webglTex := tex.(*webglTexture)

	gl := d.gl
	fbo := gl.Call("createFramebuffer")
	if fbo.IsNull() {
		return nil, errors.New("webgl: failed to create framebuffer")
	}

	gl.Call("bindFramebuffer", gl.Get("FRAMEBUFFER"), fbo)
	gl.Call("framebufferTexture2D", gl.Get("FRAMEBUFFER"), gl.Get("COLOR_ATTACHMENT0"), gl.Get("TEXTURE_2D"), webglTex.id, 0)

	status := gl.Call("checkFramebufferStatus", gl.Get("FRAMEBUFFER")).Int()
	if status != gl.Get("FRAMEBUFFER_COMPLETE").Int() {
		gl.Call("bindFramebuffer", gl.Get("FRAMEBUFFER"), js.Null())
		gl.Call("deleteFramebuffer", fbo)
		webglTex.Dispose()
		return nil, errors.New("webgl: framebuffer incomplete")
	}

	return &webglFramebuffer{
		id:     fbo,
		width:  width,
		height: height,
		tex:    webglTex,
		driver: d,
	}, nil
}

func (d *webglDriver) NewShader(src ShaderSource) (Shader, error) {
	if d == nil || d.gl.IsUndefined() {
		return nil, ErrUnsupported
	}

	gl := d.gl

	// Compile vertex shader
	vsSrc := src.GLSL.Vertex
	if vsSrc == "" {
		return nil, errors.New("webgl: missing vertex shader source")
	}
	vs := d.compileShader(vsSrc, gl.Get("VERTEX_SHADER"))
	if vs.IsNull() {
		return nil, errors.New("webgl: failed to compile vertex shader")
	}

	// Compile fragment shader
	fsSrc := src.GLSL.Fragment
	if fsSrc == "" {
		gl.Call("deleteShader", vs)
		return nil, errors.New("webgl: missing fragment shader source")
	}
	fs := d.compileShader(fsSrc, gl.Get("FRAGMENT_SHADER"))
	if fs.IsNull() {
		gl.Call("deleteShader", vs)
		return nil, errors.New("webgl: failed to compile fragment shader")
	}

	// Link program
	prog := gl.Call("createProgram")
	gl.Call("attachShader", prog, vs)
	gl.Call("attachShader", prog, fs)
	gl.Call("linkProgram", prog)

	// Check link status
	if !gl.Call("getProgramParameter", prog, gl.Get("LINK_STATUS")).Bool() {
		log := gl.Call("getProgramInfoLog", prog).String()
		gl.Call("deleteShader", vs)
		gl.Call("deleteShader", fs)
		gl.Call("deleteProgram", prog)
		return nil, errors.New("webgl: shader link failed: " + log)
	}

	// Shaders can be deleted after linking
	gl.Call("deleteShader", vs)
	gl.Call("deleteShader", fs)

	return &webglShader{
		id:     prog,
		driver: d,
		uniforms: make(map[string]any),
	}, nil
}

func (d *webglDriver) compileShader(src string, typ js.Value) js.Value {
	gl := d.gl
	shader := gl.Call("createShader", typ)
	gl.Call("shaderSource", shader, src)
	gl.Call("compileShader", shader)

	if !gl.Call("getShaderParameter", shader, gl.Get("COMPILE_STATUS")).Bool() {
		gl.Call("deleteShader", shader)
		return js.Null()
	}
	return shader
}

func (d *webglDriver) Clear(r, g, b, a float32) {
	if d == nil || d.gl.IsUndefined() {
		return
	}
	gl := d.gl
	gl.Call("clearColor", r, g, b, a)
	gl.Call("clear", gl.Get("COLOR_BUFFER_BIT"))
}

func (d *webglDriver) Draw(call DrawCall) {
	if d == nil || d.gl.IsUndefined() {
		return
	}
	gl := d.gl

	// Bind target framebuffer
	if call.Target != nil {
		if fb, ok := call.Target.(*webglFramebuffer); ok {
			gl.Call("bindFramebuffer", gl.Get("FRAMEBUFFER"), fb.id)
		}
	} else {
		gl.Call("bindFramebuffer", gl.Get("FRAMEBUFFER"), js.Null())
	}

	// Use shader program
	shader, ok := call.Shader.(*webglShader)
	if !ok || shader == nil {
		return
	}
	gl.Call("useProgram", shader.id)

	// Apply uniforms
	shader.applyUniforms()

	if len(call.Vertices) == 0 || len(call.Indices) == 0 {
		return
	}

	// Set up blending
	switch call.Blend {
	case BlendAlpha:
		gl.Call("enable", gl.Get("BLEND"))
		gl.Call("blendFunc", gl.Get("SRC_ALPHA"), gl.Get("ONE_MINUS_SRC_ALPHA"))
	case BlendAdditive:
		gl.Call("enable", gl.Get("BLEND"))
		gl.Call("blendFunc", gl.Get("SRC_ALPHA"), gl.Get("ONE"))
	default:
		gl.Call("disable", gl.Get("BLEND"))
	}

	// Bind texture if present
	if len(call.Textures) > 0 {
		if tex, ok := call.Textures[0].(*webglTexture); ok {
			gl.Call("activeTexture", gl.Get("TEXTURE0"))
			gl.Call("bindTexture", gl.Get("TEXTURE_2D"), tex.id)
		}
	}

	// Create and bind vertex buffer
	vbo := gl.Call("createBuffer")
	gl.Call("bindBuffer", gl.Get("ARRAY_BUFFER"), vbo)

	// Convert vertices to Float32Array
	vertArray := js.Global().Get("Float32Array").New(len(call.Vertices))
	for i, v := range call.Vertices {
		vertArray.SetIndex(i, v)
	}
	gl.Call("bufferData", gl.Get("ARRAY_BUFFER"), vertArray, gl.Get("DYNAMIC_DRAW"))

	// Create and bind index buffer
	ebo := gl.Call("createBuffer")
	gl.Call("bindBuffer", gl.Get("ELEMENT_ARRAY_BUFFER"), ebo)

	// Convert indices to Uint16Array
	idxArray := js.Global().Get("Uint16Array").New(len(call.Indices))
	for i, v := range call.Indices {
		idxArray.SetIndex(i, v)
	}
	gl.Call("bufferData", gl.Get("ELEMENT_ARRAY_BUFFER"), idxArray, gl.Get("DYNAMIC_DRAW"))

	// Set up vertex attributes
	layout := inferLayout(call.Vertices)
	stride := layout.stride * 4 // 4 bytes per float32

	offset := 0
	if layout.posSize > 0 {
		loc := gl.Call("getAttribLocation", shader.id, "aPosition")
		if loc.Int() >= 0 {
			gl.Call("enableVertexAttribArray", loc)
			gl.Call("vertexAttribPointer", loc, layout.posSize, gl.Get("FLOAT"), false, stride, offset)
		}
		offset += layout.posSize * 4
	}

	if layout.uvSize > 0 {
		loc := gl.Call("getAttribLocation", shader.id, "aTexCoord")
		if loc.Int() >= 0 {
			gl.Call("enableVertexAttribArray", loc)
			gl.Call("vertexAttribPointer", loc, layout.uvSize, gl.Get("FLOAT"), false, stride, offset)
		}
		offset += layout.uvSize * 4
	}

	if layout.colorSize > 0 {
		loc := gl.Call("getAttribLocation", shader.id, "aColor")
		if loc.Int() >= 0 {
			gl.Call("enableVertexAttribArray", loc)
			gl.Call("vertexAttribPointer", loc, layout.colorSize, gl.Get("FLOAT"), false, stride, offset)
		}
	}

	// Draw
	gl.Call("drawElements", gl.Get("TRIANGLES"), len(call.Indices), gl.Get("UNSIGNED_SHORT"), 0)

	// Clean up
	gl.Call("deleteBuffer", vbo)
	gl.Call("deleteBuffer", ebo)
}

func (d *webglDriver) ReadPixels(fb Framebuffer, rect image.Rectangle) ([]byte, error) {
	if d == nil || d.gl.IsUndefined() {
		return nil, ErrUnsupported
	}
	gl := d.gl

	var width, height int
	if fb != nil {
		webglFb, ok := fb.(*webglFramebuffer)
		if !ok {
			return nil, ErrUnsupported
		}
		gl.Call("bindFramebuffer", gl.Get("FRAMEBUFFER"), webglFb.id)
		width, height = webglFb.width, webglFb.height
	} else {
		width = d.canvas.Get("width").Int()
		height = d.canvas.Get("height").Int()
	}

	if rect.Empty() {
		rect = image.Rect(0, 0, width, height)
	}

	w := rect.Dx()
	h := rect.Dy()
	if w <= 0 || h <= 0 {
		return nil, nil
	}

	// Create a Uint8Array to read pixels into
	pixels := js.Global().Get("Uint8Array").New(w * h * 4)
	gl.Call("readPixels", rect.Min.X, rect.Min.Y, w, h, gl.Get("RGBA"), gl.Get("UNSIGNED_BYTE"), pixels)

	// Copy to Go byte slice
	result := make([]byte, w*h*4)
	js.CopyBytesToGo(result, pixels)

	// Flip vertically (WebGL has origin at bottom-left)
	flipPixelsVertical(result, w, h)

	return result, nil
}

func (d *webglDriver) MaxTextureSize() int {
	if d == nil {
		return 0
	}
	return d.maxTex
}

// webglTexture implements Texture for WebGL.
type webglTexture struct {
	id     js.Value
	width  int
	height int
	driver *webglDriver
}

func (t *webglTexture) ID() uint32 {
	// WebGL objects are not uint32, return a placeholder
	return 1
}

func (t *webglTexture) Size() (int, int) {
	if t == nil {
		return 0, 0
	}
	return t.width, t.height
}

func (t *webglTexture) Upload(pixels []byte, region image.Rectangle) {
	if t == nil || t.driver == nil || len(pixels) == 0 {
		return
	}

	gl := t.driver.gl
	gl.Call("bindTexture", gl.Get("TEXTURE_2D"), t.id)

	if region.Empty() {
		region = image.Rect(0, 0, t.width, t.height)
	}

	// Convert pixels to Uint8Array
	arr := js.Global().Get("Uint8Array").New(len(pixels))
	js.CopyBytesToJS(arr, pixels)

	if region.Min.X == 0 && region.Min.Y == 0 &&
		region.Max.X == t.width && region.Max.Y == t.height {
		// Full upload
		gl.Call("texImage2D", gl.Get("TEXTURE_2D"), 0, gl.Get("RGBA"), t.width, t.height, 0, gl.Get("RGBA"), gl.Get("UNSIGNED_BYTE"), arr)
	} else {
		// Sub-image update
		gl.Call("texSubImage2D", gl.Get("TEXTURE_2D"), 0, region.Min.X, region.Min.Y, region.Dx(), region.Dy(), gl.Get("RGBA"), gl.Get("UNSIGNED_BYTE"), arr)
	}
}

func (t *webglTexture) Dispose() {
	if t == nil || t.driver == nil {
		return
	}
	gl := t.driver.gl
	gl.Call("deleteTexture", t.id)
	t.id = js.Null()
}

// webglFramebuffer implements Framebuffer for WebGL.
type webglFramebuffer struct {
	id     js.Value
	width  int
	height int
	tex    *webglTexture
	driver *webglDriver
}

func (f *webglFramebuffer) ID() uint32 {
	return 1
}

func (f *webglFramebuffer) Size() (int, int) {
	if f == nil {
		return 0, 0
	}
	return f.width, f.height
}

func (f *webglFramebuffer) Bind() {
	if f == nil || f.driver == nil {
		return
	}
	gl := f.driver.gl
	gl.Call("bindFramebuffer", gl.Get("FRAMEBUFFER"), f.id)
}

func (f *webglFramebuffer) Texture() Texture {
	if f == nil {
		return nil
	}
	return f.tex
}

func (f *webglFramebuffer) Dispose() {
	if f == nil || f.driver == nil {
		return
	}
	gl := f.driver.gl
	gl.Call("deleteFramebuffer", f.id)
	if f.tex != nil {
		f.tex.Dispose()
	}
}

// webglShader implements Shader for WebGL.
type webglShader struct {
	id       js.Value
	driver   *webglDriver
	uniforms map[string]any
	locations map[string]js.Value
}

func (s *webglShader) ID() uint32 {
	return 1
}

func (s *webglShader) SetUniform(name string, value any) {
	if s == nil {
		return
	}
	if s.uniforms == nil {
		s.uniforms = make(map[string]any)
	}
	s.uniforms[name] = value
}

func (s *webglShader) applyUniforms() {
	if s == nil || s.driver == nil {
		return
	}
	gl := s.driver.gl

	if s.locations == nil {
		s.locations = make(map[string]js.Value)
	}

	for name, value := range s.uniforms {
		loc, ok := s.locations[name]
		if !ok {
			loc = gl.Call("getUniformLocation", s.id, name)
			s.locations[name] = loc
		}
		if loc.IsNull() {
			continue
		}

		switch v := value.(type) {
		case int:
			gl.Call("uniform1i", loc, v)
		case float32:
			gl.Call("uniform1f", loc, v)
		case Matrix3:
			// Convert to Float32Array
			arr := js.Global().Get("Float32Array").New(9)
			for i := 0; i < 9; i++ {
				arr.SetIndex(i, v.m[i])
			}
			gl.Call("uniformMatrix3fv", loc, false, arr)
		case [9]float32:
			arr := js.Global().Get("Float32Array").New(9)
			for i := 0; i < 9; i++ {
				arr.SetIndex(i, v[i])
			}
			gl.Call("uniformMatrix3fv", loc, false, arr)
		}
	}
}

func (s *webglShader) Dispose() {
	if s == nil || s.driver == nil {
		return
	}
	gl := s.driver.gl
	gl.Call("deleteProgram", s.id)
}
