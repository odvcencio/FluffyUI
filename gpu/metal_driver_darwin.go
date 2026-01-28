//go:build darwin

package gpu

import (
	"errors"
	"image"
	"image/color"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"unsafe"

	"github.com/odvcencio/fluffyui/gpu/metal/mtl"
)

type metalDriver struct {
	device  uintptr
	queue   uintptr
	maxTex  int
	current *metalFramebuffer
	worker  *glWorker
}

type metalTexture struct {
	id     uintptr
	width  int
	height int
	driver *metalDriver
}

type metalFramebuffer struct {
	tex    *metalTexture
	width  int
	height int
	driver *metalDriver
}

type metalShader struct {
	pipelines   map[BlendMode]uintptr
	uniformMeta map[string]metalUniformMeta
	uniforms    map[string]*metalUniformValue
	layout      metalVertexLayout
	driver      *metalDriver
}

type metalUniformKind int

const (
	metalUniformFloat metalUniformKind = iota
	metalUniformFloat2
	metalUniformFloat4
	metalUniformFloat4x4
	metalUniformInt
)

type metalUniformMeta struct {
	index int
	kind  metalUniformKind
	size  uintptr
}

type metalUniformValue struct {
	kind metalUniformKind
	f    [16]float32
	i    int32
}

type metalVertexLayout struct {
	stride      int
	attr1Fmt    int
	attr1Offset int
}

func newMetalDriver() (Driver, error) {
	if runtime.GOOS != "darwin" {
		return nil, ErrUnsupported
	}
	if err := objcInit(); err != nil {
		return nil, err
	}
	driver := &metalDriver{maxTex: 8192}
	driver.worker = newGLWorker()
	var initErr error
	driver.run(func() {
		initErr = driver.init()
	})
	if initErr != nil {
		if driver.worker != nil {
			driver.worker.stop()
		}
		return nil, initErr
	}
	return driver, nil
}

func (d *metalDriver) init() error {
	device, err := mtl.SystemDefaultDevice()
	if err != nil || device == 0 {
		return ErrUnsupported
	}
	d.device = device
	if objcRetain != nil {
		objcRetain(d.device)
	}
	queue := objcMsgSend(d.device, sel("newCommandQueue"))
	if queue == 0 {
		return errors.New("metal: failed to create command queue")
	}
	d.queue = queue
	return nil
}

func (d *metalDriver) Backend() Backend {
	return BackendMetal
}

func (d *metalDriver) Init() error {
	if d == nil || d.device == 0 {
		return ErrUnsupported
	}
	return nil
}

func (d *metalDriver) Dispose() {
	if d == nil {
		return
	}
	if d.worker != nil {
		d.run(func() {
			if d.queue != 0 {
				objcReleaseObj(d.queue)
				d.queue = 0
			}
			if d.device != 0 {
				objcReleaseObj(d.device)
				d.device = 0
			}
			d.current = nil
		})
		d.worker.stop()
		d.worker = nil
		return
	}
	if d.queue != 0 {
		objcReleaseObj(d.queue)
		d.queue = 0
	}
	if d.device != 0 {
		objcReleaseObj(d.device)
		d.device = 0
	}
	d.current = nil
}

func (d *metalDriver) run(fn func()) {
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

func (d *metalDriver) NewTexture(width, height int) (Texture, error) {
	if d == nil || d.device == 0 || width <= 0 || height <= 0 {
		return nil, ErrUnsupported
	}
	var tex *metalTexture
	var err error
	d.run(func() {
		tex, err = d.newTexture(width, height)
	})
	return tex, err
}

func (d *metalDriver) newTexture(width, height int) (*metalTexture, error) {
	if d == nil || d.device == 0 {
		return nil, ErrUnsupported
	}
	desc := objcMsgSend4(objcClass("MTLTextureDescriptor"), sel("texture2DDescriptorWithPixelFormat:width:height:mipmapped:"),
		uintptr(mtlPixelFormatRGBA8Unorm), uintptr(width), uintptr(height), uintptr(0))
	if desc == 0 {
		return nil, errors.New("metal: texture descriptor failed")
	}
	usage := mtlTextureUsageShaderRead | mtlTextureUsageRenderTarget
	objcMsgSend1(desc, sel("setUsage:"), uintptr(usage))
	objcMsgSend1(desc, sel("setStorageMode:"), uintptr(mtlStorageModeShared))
	tex := objcMsgSend1(d.device, sel("newTextureWithDescriptor:"), desc)
	objcReleaseObj(desc)
	if tex == 0 {
		return nil, errors.New("metal: newTextureWithDescriptor failed")
	}
	return &metalTexture{id: tex, width: width, height: height, driver: d}, nil
}

func (d *metalDriver) NewFramebuffer(width, height int) (Framebuffer, error) {
	if d == nil || d.device == 0 || width <= 0 || height <= 0 {
		return nil, ErrUnsupported
	}
	var fb *metalFramebuffer
	var err error
	d.run(func() {
		tex, texErr := d.newTexture(width, height)
		if texErr != nil {
			err = texErr
			return
		}
		fb = &metalFramebuffer{tex: tex, width: width, height: height, driver: d}
	})
	return fb, err
}

func (d *metalDriver) NewShader(src ShaderSource) (Shader, error) {
	if d == nil || d.device == 0 {
		return nil, ErrUnsupported
	}
	if src.Metal == "" {
		return nil, errors.New("metal: missing shader source")
	}
	var shader *metalShader
	var err error
	d.run(func() {
		shader, err = d.newShader(src.Metal)
	})
	return shader, err
}

func (d *metalDriver) newShader(source string) (*metalShader, error) {
	if source == "" {
		return nil, errors.New("metal: empty shader source")
	}
	vertexName, fragmentName := parseMetalEntryPoints(source)
	layout := parseMetalVertexLayout(source)
	uniformMeta := parseMetalUniforms(source)
	var shader *metalShader
	withAutoreleasePool(func() {
		lib, err := d.compileLibrary(source)
		if err != nil {
			shader = nil
			return
		}
		defer objcReleaseObj(lib)
		vs := objcMsgSend1(lib, sel("newFunctionWithName:"), nsString(vertexName))
		if vs == 0 {
			shader = nil
			return
		}
		defer objcReleaseObj(vs)
		fs := objcMsgSend1(lib, sel("newFunctionWithName:"), nsString(fragmentName))
		if fs == 0 {
			shader = nil
			return
		}
		defer objcReleaseObj(fs)
		vertexDesc := newMetalVertexDescriptor(layout)
		if vertexDesc == 0 {
			shader = nil
			return
		}
		defer objcReleaseObj(vertexDesc)
		pipelines := make(map[BlendMode]uintptr)
		for _, blend := range []BlendMode{BlendNone, BlendAlpha, BlendAdditive} {
			ps := d.buildPipeline(vs, fs, vertexDesc, blend)
			if ps == 0 {
				shader = nil
				return
			}
			pipelines[blend] = ps
		}
		shader = &metalShader{
			pipelines:   pipelines,
			uniformMeta: uniformMeta,
			uniforms:    make(map[string]*metalUniformValue),
			layout:      layout,
			driver:      d,
		}
	})
	if shader == nil {
		return nil, errors.New("metal: shader compile failed")
	}
	return shader, nil
}

func (d *metalDriver) compileLibrary(source string) (uintptr, error) {
	nsSrc := nsString(source)
	if nsSrc == 0 {
		return 0, errors.New("metal: failed to create source string")
	}
	var errObj uintptr
	lib := objcMsgSend3(d.device, sel("newLibraryWithSource:options:error:"), nsSrc, 0, uintptr(unsafe.Pointer(&errObj)))
	if lib == 0 {
		return 0, errors.New("metal: newLibraryWithSource failed")
	}
	return lib, nil
}

func (d *metalDriver) buildPipeline(vs, fs, vertexDesc uintptr, blend BlendMode) uintptr {
	desc := objcAllocInit("MTLRenderPipelineDescriptor")
	if desc == 0 {
		return 0
	}
	defer objcReleaseObj(desc)
	objcMsgSend1(desc, sel("setVertexFunction:"), vs)
	objcMsgSend1(desc, sel("setFragmentFunction:"), fs)
	objcMsgSend1(desc, sel("setVertexDescriptor:"), vertexDesc)
	attachments := objcMsgSend(desc, sel("colorAttachments"))
	if attachments == 0 {
		return 0
	}
	attachment := objcMsgSend1(attachments, sel("objectAtIndexedSubscript:"), 0)
	if attachment == 0 {
		return 0
	}
	objcMsgSend1(attachment, sel("setPixelFormat:"), uintptr(mtlPixelFormatRGBA8Unorm))
	if blend != BlendNone {
		objcMsgSend1(attachment, sel("setBlendingEnabled:"), 1)
		src := uintptr(mtlBlendFactorSourceAlpha)
		dst := uintptr(mtlBlendFactorOneMinusSourceAlpha)
		if blend == BlendAdditive {
			dst = uintptr(mtlBlendFactorOne)
		}
		objcMsgSend1(attachment, sel("setSourceRGBBlendFactor:"), src)
		objcMsgSend1(attachment, sel("setDestinationRGBBlendFactor:"), dst)
		objcMsgSend1(attachment, sel("setRgbBlendOperation:"), uintptr(mtlBlendOperationAdd))
		objcMsgSend1(attachment, sel("setSourceAlphaBlendFactor:"), src)
		objcMsgSend1(attachment, sel("setDestinationAlphaBlendFactor:"), dst)
		objcMsgSend1(attachment, sel("setAlphaBlendOperation:"), uintptr(mtlBlendOperationAdd))
	}
	var errObj uintptr
	ps := objcMsgSend2(d.device, sel("newRenderPipelineStateWithDescriptor:error:"), desc, uintptr(unsafe.Pointer(&errObj)))
	return ps
}

func (d *metalDriver) Clear(r, g, b, a float32) {
	if d == nil || d.device == 0 {
		return
	}
	d.run(func() {
		d.clear(r, g, b, a)
	})
}

func (d *metalDriver) clear(r, g, b, a float32) {
	withAutoreleasePool(func() {
		target := d.current
		if target == nil || target.tex == nil || target.tex.id == 0 {
			return
		}
		commandBuffer := objcMsgSend(d.queue, sel("commandBuffer"))
		if commandBuffer == 0 {
			return
		}
		pass := objcMsgSend(objcClass("MTLRenderPassDescriptor"), sel("renderPassDescriptor"))
		if pass == 0 {
			return
		}
		attachments := objcMsgSend(pass, sel("colorAttachments"))
		attachment := objcMsgSend1(attachments, sel("objectAtIndexedSubscript:"), 0)
		objcMsgSend1(attachment, sel("setTexture:"), target.tex.id)
		objcMsgSend1(attachment, sel("setLoadAction:"), uintptr(mtlLoadActionClear))
		objcMsgSend1(attachment, sel("setStoreAction:"), uintptr(mtlStoreActionStore))
		objcMsgSendClearColor(attachment, sel("setClearColor:"), mtlClearColor{R: float64(r), G: float64(g), B: float64(b), A: float64(a)})
		encoder := objcMsgSend1(commandBuffer, sel("renderCommandEncoderWithDescriptor:"), pass)
		if encoder != 0 {
			objcMsgSend(encoder, sel("endEncoding"))
		}
		objcMsgSend(commandBuffer, sel("commit"))
		objcMsgSend(commandBuffer, sel("waitUntilCompleted"))
	})
}

func (d *metalDriver) Draw(call DrawCall) {
	if d == nil || d.device == 0 {
		return
	}
	d.run(func() {
		d.draw(call)
	})
}

func (d *metalDriver) draw(call DrawCall) {
	withAutoreleasePool(func() {
		if call.Shader == nil || len(call.Vertices) == 0 || len(call.Indices) == 0 {
			return
		}
		shader, ok := call.Shader.(*metalShader)
		if !ok || shader == nil {
			return
		}
		target := d.current
		if call.Target != nil {
			if fb, ok := call.Target.(*metalFramebuffer); ok {
				target = fb
			}
		}
		if target == nil || target.tex == nil || target.tex.id == 0 {
			return
		}
		pipeline := shader.pipelines[call.Blend]
		if pipeline == 0 {
			pipeline = shader.pipelines[BlendNone]
		}
		commandBuffer := objcMsgSend(d.queue, sel("commandBuffer"))
		if commandBuffer == 0 {
			return
		}
		pass := objcMsgSend(objcClass("MTLRenderPassDescriptor"), sel("renderPassDescriptor"))
		if pass == 0 {
			return
		}
		attachments := objcMsgSend(pass, sel("colorAttachments"))
		attachment := objcMsgSend1(attachments, sel("objectAtIndexedSubscript:"), 0)
		objcMsgSend1(attachment, sel("setTexture:"), target.tex.id)
		objcMsgSend1(attachment, sel("setLoadAction:"), uintptr(mtlLoadActionLoad))
		objcMsgSend1(attachment, sel("setStoreAction:"), uintptr(mtlStoreActionStore))
		encoder := objcMsgSend1(commandBuffer, sel("renderCommandEncoderWithDescriptor:"), pass)
		if encoder == 0 {
			return
		}
		objcMsgSend1(encoder, sel("setRenderPipelineState:"), pipeline)
		viewport := mtlViewport{OriginX: 0, OriginY: 0, Width: float64(target.width), Height: float64(target.height), ZNear: 0, ZFar: 1}
		objcMsgSendViewport(encoder, sel("setViewport:"), viewport)
		if call.Scissor != nil {
			rect := mtlScissorRect{
				X:      uint64(call.Scissor.Min.X),
				Y:      uint64(call.Scissor.Min.Y),
				Width:  uint64(call.Scissor.Dx()),
				Height: uint64(call.Scissor.Dy()),
			}
			objcMsgSendScissor(encoder, sel("setScissorRect:"), rect)
		}
		vertexBytes := len(call.Vertices) * 4
		indexBytes := len(call.Indices) * 2
		if vertexBytes == 0 || indexBytes == 0 {
			objcMsgSend(encoder, sel("endEncoding"))
			objcMsgSend(commandBuffer, sel("commit"))
			objcMsgSend(commandBuffer, sel("waitUntilCompleted"))
			return
		}
		vb := objcMsgSend3(d.device, sel("newBufferWithBytes:length:options:"), uintptr(unsafe.Pointer(&call.Vertices[0])), uintptr(vertexBytes), 0)
		ib := objcMsgSend3(d.device, sel("newBufferWithBytes:length:options:"), uintptr(unsafe.Pointer(&call.Indices[0])), uintptr(indexBytes), 0)
		if vb != 0 {
			objcMsgSend3(encoder, sel("setVertexBuffer:offset:atIndex:"), vb, 0, 0)
		}
		if len(call.Textures) > 0 {
			if tex, ok := call.Textures[0].(*metalTexture); ok && tex != nil && tex.id != 0 {
				objcMsgSend2(encoder, sel("setFragmentTexture:atIndex:"), tex.id, 0)
			}
		}
		shader.applyUniforms(encoder)
		if ib != 0 {
			objcMsgSend5(encoder, sel("drawIndexedPrimitives:indexCount:indexType:indexBuffer:indexBufferOffset:"),
				uintptr(mtlPrimitiveTypeTriangle), uintptr(len(call.Indices)), uintptr(mtlIndexTypeUInt16), ib, 0)
		}
		objcMsgSend(encoder, sel("endEncoding"))
		objcMsgSend(commandBuffer, sel("commit"))
		objcMsgSend(commandBuffer, sel("waitUntilCompleted"))
		if vb != 0 {
			objcReleaseObj(vb)
		}
		if ib != 0 {
			objcReleaseObj(ib)
		}
	})
}

func (d *metalDriver) ReadPixels(fb Framebuffer, rect image.Rectangle) ([]byte, error) {
	if d == nil || d.device == 0 {
		return nil, ErrUnsupported
	}
	var pixels []byte
	var err error
	d.run(func() {
		pixels, err = d.readPixels(fb, rect)
	})
	return pixels, err
}

func (d *metalDriver) readPixels(fb Framebuffer, rect image.Rectangle) ([]byte, error) {
	var tex *metalTexture
	if fb == nil {
		if d.current == nil || d.current.tex == nil {
			return nil, ErrUnsupported
		}
		tex = d.current.tex
	} else {
		mfb, ok := fb.(*metalFramebuffer)
		if !ok || mfb == nil || mfb.tex == nil {
			return nil, ErrUnsupported
		}
		tex = mfb.tex
	}
	return d.readTexture(tex, rect)
}

func (d *metalDriver) ReadTexturePixels(tex Texture, rect image.Rectangle) ([]byte, int, int, error) {
	if d == nil || d.device == 0 {
		return nil, 0, 0, ErrUnsupported
	}
	var pixels []byte
	var w, h int
	var err error
	d.run(func() {
		if mtex, ok := tex.(*metalTexture); ok {
			pixels, err = d.readTexture(mtex, rect)
			if err == nil {
				w = mtex.width
				h = mtex.height
			}
		} else {
			err = ErrUnsupported
		}
	})
	return pixels, w, h, err
}

func (d *metalDriver) readTexture(tex *metalTexture, rect image.Rectangle) ([]byte, error) {
	if tex == nil || tex.id == 0 {
		return nil, ErrUnsupported
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
		return nil, nil
	}
	pixels := make([]byte, w*h*4)
	region := mtlRegion{
		Origin: mtlOrigin{X: uint64(rect.Min.X), Y: uint64(rect.Min.Y), Z: 0},
		Size:   mtlSize{Width: uint64(w), Height: uint64(h), Depth: 1},
	}
	objcMsgSendGetBytes(tex.id, sel("getBytes:bytesPerRow:fromRegion:mipmapLevel:"), unsafe.Pointer(&pixels[0]), uintptr(w*4), region, 0)
	return pixels, nil
}

func (d *metalDriver) MaxTextureSize() int {
	if d == nil {
		return 0
	}
	return d.maxTex
}

func (t *metalTexture) ID() uint32 {
	if t == nil {
		return 0
	}
	return 0
}

func (t *metalTexture) Size() (int, int) {
	if t == nil {
		return 0, 0
	}
	return t.width, t.height
}

func (t *metalTexture) Upload(pixels []byte, region image.Rectangle) {
	if t == nil || t.id == 0 || t.driver == nil || len(pixels) == 0 {
		return
	}
	t.driver.run(func() {
		t.upload(pixels, region)
	})
}

func (t *metalTexture) upload(pixels []byte, region image.Rectangle) {
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
	rw := region.Dx()
	rh := region.Dy()
	if rw <= 0 || rh <= 0 {
		return
	}
	regionStruct := mtlRegion{
		Origin: mtlOrigin{X: uint64(region.Min.X), Y: uint64(region.Min.Y), Z: 0},
		Size:   mtlSize{Width: uint64(rw), Height: uint64(rh), Depth: 1},
	}
	objcMsgSendReplaceRegion(t.id, sel("replaceRegion:mipmapLevel:withBytes:bytesPerRow:"), regionStruct, 0, unsafe.Pointer(&pixels[0]), uintptr(rw*4))
}

func (t *metalTexture) Dispose() {
	if t == nil || t.id == 0 {
		return
	}
	if t.driver != nil {
		t.driver.run(func() {
			objcReleaseObj(t.id)
			t.id = 0
		})
		return
	}
	objcReleaseObj(t.id)
	t.id = 0
}

func (f *metalFramebuffer) ID() uint32 {
	if f == nil {
		return 0
	}
	return 0
}

func (f *metalFramebuffer) Size() (int, int) {
	if f == nil {
		return 0, 0
	}
	return f.width, f.height
}

func (f *metalFramebuffer) Bind() {
	if f == nil || f.driver == nil {
		return
	}
	f.driver.run(func() {
		f.driver.current = f
	})
}

func (f *metalFramebuffer) Texture() Texture {
	if f == nil {
		return nil
	}
	return f.tex
}

func (f *metalFramebuffer) Dispose() {
	if f == nil {
		return
	}
	if f.driver != nil {
		f.driver.run(func() {
			if f.tex != nil {
				f.tex.Dispose()
				f.tex = nil
			}
		})
		return
	}
	if f.tex != nil {
		f.tex.Dispose()
		f.tex = nil
	}
}

func (s *metalShader) ID() uint32 {
	if s == nil {
		return 0
	}
	return 0
}

func (s *metalShader) SetUniform(name string, value any) {
	if s == nil || name == "" {
		return
	}
	meta, ok := s.uniformMeta[name]
	if !ok {
		return
	}
	val := toMetalUniformValue(meta.kind, value)
	if val == nil {
		return
	}
	if s.uniforms == nil {
		s.uniforms = make(map[string]*metalUniformValue)
	}
	s.uniforms[name] = val
}

func (s *metalShader) applyUniforms(encoder uintptr) {
	if s == nil || encoder == 0 || len(s.uniforms) == 0 {
		return
	}
	for name, val := range s.uniforms {
		meta, ok := s.uniformMeta[name]
		if !ok || val == nil {
			continue
		}
		ptr, size := metalUniformBytes(val, meta)
		if ptr == nil || size == 0 {
			continue
		}
		objcMsgSend3(encoder, sel("setVertexBytes:length:atIndex:"), uintptr(ptr), uintptr(size), uintptr(meta.index))
		objcMsgSend3(encoder, sel("setFragmentBytes:length:atIndex:"), uintptr(ptr), uintptr(size), uintptr(meta.index))
	}
}

func (s *metalShader) Dispose() {
	if s == nil {
		return
	}
	if s.driver != nil {
		s.driver.run(func() {
			for _, ps := range s.pipelines {
				if ps != 0 {
					objcReleaseObj(ps)
				}
			}
			s.pipelines = nil
		})
		return
	}
	for _, ps := range s.pipelines {
		if ps != 0 {
			objcReleaseObj(ps)
		}
	}
	s.pipelines = nil
}

func parseMetalEntryPoints(source string) (string, string) {
	vertex := "vertex_main"
	fragment := "fragment_main"
	for _, line := range strings.Split(source, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "vertex ") {
			fields := strings.Fields(trim)
			if len(fields) >= 2 {
				vertex = fields[1]
				if idx := strings.Index(vertex, "("); idx > 0 {
					vertex = vertex[:idx]
				}
			}
		}
		if strings.HasPrefix(trim, "fragment ") {
			fields := strings.Fields(trim)
			if len(fields) >= 2 {
				fragment = fields[1]
				if idx := strings.Index(fragment, "("); idx > 0 {
					fragment = fragment[:idx]
				}
			}
		}
	}
	return vertex, fragment
}

func parseMetalVertexLayout(source string) metalVertexLayout {
	layout := metalVertexLayout{
		stride:      16,
		attr1Fmt:    mtlVertexFormatFloat2,
		attr1Offset: 8,
	}
	re := regexp.MustCompile(`(?m)^\s*(float\w*)\s+\w+\s*\[\[attribute\(1\)\]\]`)
	if match := re.FindStringSubmatch(source); len(match) > 1 {
		if match[1] == "float4" {
			layout.stride = 24
			layout.attr1Fmt = mtlVertexFormatFloat4
			layout.attr1Offset = 8
		}
	}
	return layout
}

func parseMetalUniforms(source string) map[string]metalUniformMeta {
	out := make(map[string]metalUniformMeta)
	re := regexp.MustCompile(`constant\s+([A-Za-z0-9x]+)\s*&?\s*([A-Za-z0-9_]+)\s*\[\[buffer\((\d+)\)\]\]`)
	matches := re.FindAllStringSubmatch(source, -1)
	for _, match := range matches {
		if len(match) < 4 {
			continue
		}
		typ := match[1]
		name := match[2]
		idx, _ := strconv.Atoi(match[3])
		meta := metalUniformMeta{index: idx}
		switch typ {
		case "float":
			meta.kind = metalUniformFloat
			meta.size = 4
		case "float2":
			meta.kind = metalUniformFloat2
			meta.size = 8
		case "float4":
			meta.kind = metalUniformFloat4
			meta.size = 16
		case "float4x4":
			meta.kind = metalUniformFloat4x4
			meta.size = 64
		default:
			continue
		}
		out[name] = meta
	}
	return out
}

func newMetalVertexDescriptor(layout metalVertexLayout) uintptr {
	desc := objcAllocInit("MTLVertexDescriptor")
	if desc == 0 {
		return 0
	}
	attrs := objcMsgSend(desc, sel("attributes"))
	if attrs == 0 {
		objcReleaseObj(desc)
		return 0
	}
	attr0 := objcMsgSend1(attrs, sel("objectAtIndexedSubscript:"), 0)
	attr1 := objcMsgSend1(attrs, sel("objectAtIndexedSubscript:"), 1)
	if attr0 == 0 || attr1 == 0 {
		objcReleaseObj(desc)
		return 0
	}
	objcMsgSend1(attr0, sel("setFormat:"), uintptr(mtlVertexFormatFloat2))
	objcMsgSend1(attr0, sel("setOffset:"), 0)
	objcMsgSend1(attr0, sel("setBufferIndex:"), 0)
	objcMsgSend1(attr1, sel("setFormat:"), uintptr(layout.attr1Fmt))
	objcMsgSend1(attr1, sel("setOffset:"), uintptr(layout.attr1Offset))
	objcMsgSend1(attr1, sel("setBufferIndex:"), 0)
	layouts := objcMsgSend(desc, sel("layouts"))
	if layouts == 0 {
		objcReleaseObj(desc)
		return 0
	}
	layout0 := objcMsgSend1(layouts, sel("objectAtIndexedSubscript:"), 0)
	if layout0 == 0 {
		objcReleaseObj(desc)
		return 0
	}
	objcMsgSend1(layout0, sel("setStride:"), uintptr(layout.stride))
	objcMsgSend1(layout0, sel("setStepFunction:"), uintptr(mtlVertexStepFunctionPerVertex))
	objcMsgSend1(layout0, sel("setStepRate:"), 1)
	return desc
}

func toMetalUniformValue(kind metalUniformKind, value any) *metalUniformValue {
	val := &metalUniformValue{kind: kind}
	switch kind {
	case metalUniformFloat:
		if f, ok := floatFromAny(value); ok {
			val.f[0] = f
			return val
		}
	case metalUniformFloat2:
		if v, ok := vec2FromAny(value); ok {
			val.f[0], val.f[1] = v[0], v[1]
			return val
		}
	case metalUniformFloat4:
		if v, ok := vec4FromAny(value); ok {
			copy(val.f[:4], v[:4])
			return val
		}
	case metalUniformFloat4x4:
		if m, ok := mat4FromAny(value); ok {
			copy(val.f[:16], m[:16])
			return val
		}
	case metalUniformInt:
		if i, ok := intFromAny(value); ok {
			val.i = int32(i)
			return val
		}
	}
	return nil
}

func metalUniformBytes(val *metalUniformValue, meta metalUniformMeta) (unsafe.Pointer, uintptr) {
	if val == nil {
		return nil, 0
	}
	switch meta.kind {
	case metalUniformFloat, metalUniformFloat2, metalUniformFloat4, metalUniformFloat4x4:
		return unsafe.Pointer(&val.f[0]), meta.size
	case metalUniformInt:
		return unsafe.Pointer(&val.i), 4
	default:
		return nil, 0
	}
}

func floatFromAny(value any) (float32, bool) {
	switch v := value.(type) {
	case float32:
		return v, true
	case float64:
		return float32(v), true
	case int:
		return float32(v), true
	case int32:
		return float32(v), true
	default:
		return 0, false
	}
}

func intFromAny(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int32:
		return int(v), true
	case uint32:
		return int(v), true
	default:
		return 0, false
	}
}

func vec2FromAny(value any) ([2]float32, bool) {
	switch v := value.(type) {
	case [2]float32:
		return v, true
	case [2]float64:
		return [2]float32{float32(v[0]), float32(v[1])}, true
	case []float32:
		if len(v) >= 2 {
			return [2]float32{v[0], v[1]}, true
		}
	case []float64:
		if len(v) >= 2 {
			return [2]float32{float32(v[0]), float32(v[1])}, true
		}
	}
	return [2]float32{}, false
}

func vec4FromAny(value any) ([4]float32, bool) {
	switch v := value.(type) {
	case [4]float32:
		return v, true
	case [4]float64:
		return [4]float32{float32(v[0]), float32(v[1]), float32(v[2]), float32(v[3])}, true
	case []float32:
		if len(v) >= 4 {
			return [4]float32{v[0], v[1], v[2], v[3]}, true
		}
	case []float64:
		if len(v) >= 4 {
			return [4]float32{float32(v[0]), float32(v[1]), float32(v[2]), float32(v[3])}, true
		}
	case color.RGBA:
		return [4]float32{float32(v.R) / 255, float32(v.G) / 255, float32(v.B) / 255, float32(v.A) / 255}, true
	}
	return [4]float32{}, false
}

func mat4FromAny(value any) ([16]float32, bool) {
	switch v := value.(type) {
	case [16]float32:
		return v, true
	case []float32:
		if len(v) >= 16 {
			var out [16]float32
			copy(out[:], v[:16])
			return out, true
		}
	case Matrix3:
		return matrix3ToMetal(v), true
	case [9]float32:
		return matrix3ToMetal(Matrix3{m: v}), true
	}
	return [16]float32{}, false
}

func matrix3ToMetal(m Matrix3) [16]float32 {
	a := m.m[0]
	b := m.m[1]
	c := m.m[2]
	d := m.m[3]
	e := m.m[4]
	f := m.m[5]
	return [16]float32{
		a, d, 0, 0,
		b, e, 0, 0,
		0, 0, 1, 0,
		c, f, 0, 1,
	}
}
