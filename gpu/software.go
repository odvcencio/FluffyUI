package gpu

import (
	"image"
)

type softwareDriver struct {
	nextID  uint32
	current *softwareFramebuffer
}

type softwareTexture struct {
	id     uint32
	width  int
	height int
	pixels []byte
}

type softwareFramebuffer struct {
	id     uint32
	tex    *softwareTexture
	driver *softwareDriver
}

type softwareShader struct {
	id       uint32
	uniforms map[string]any
	apply    func(pixels []byte, width, height int, uniforms map[string]any) []byte
}

func newSoftwareDriver() *softwareDriver {
	return &softwareDriver{nextID: 1}
}

func (d *softwareDriver) Backend() Backend {
	return BackendSoftware
}

func (d *softwareDriver) Init() error {
	return nil
}

func (d *softwareDriver) Dispose() {}

func (d *softwareDriver) NewTexture(width, height int) (Texture, error) {
	if width <= 0 || height <= 0 {
		return nil, ErrUnsupported
	}
	id := d.next()
	tex := &softwareTexture{
		id:     id,
		width:  width,
		height: height,
		pixels: make([]byte, width*height*4),
	}
	return tex, nil
}

func (d *softwareDriver) NewFramebuffer(width, height int) (Framebuffer, error) {
	texIface, err := d.NewTexture(width, height)
	if err != nil {
		return nil, err
	}
	tex := texIface.(*softwareTexture)
	fb := &softwareFramebuffer{
		id:     d.next(),
		tex:    tex,
		driver: d,
	}
	d.current = fb
	return fb, nil
}

func (d *softwareDriver) NewShader(_ ShaderSource) (Shader, error) {
	return &softwareShader{id: d.next(), uniforms: make(map[string]any)}, nil
}

func (d *softwareDriver) Clear(r, g, b, a float32) {
	if d == nil || d.current == nil || d.current.tex == nil {
		return
	}
	col := rgbaFromFloats(r, g, b, a)
	clearPixels(d.current.tex.pixels, d.current.tex.width, d.current.tex.height, col)
}

func (d *softwareDriver) Draw(_ DrawCall) {}

func (d *softwareDriver) ReadPixels(fb Framebuffer, rect image.Rectangle) ([]byte, error) {
	soft, ok := fb.(*softwareFramebuffer)
	if !ok || soft.tex == nil {
		return nil, ErrUnsupported
	}
	w, h := soft.tex.width, soft.tex.height
	if rect.Empty() {
		rect = image.Rect(0, 0, w, h)
	}
	if rect.Min.X < 0 {
		rect.Min.X = 0
	}
	if rect.Min.Y < 0 {
		rect.Min.Y = 0
	}
	if rect.Max.X > w {
		rect.Max.X = w
	}
	if rect.Max.Y > h {
		rect.Max.Y = h
	}
	outW := rect.Dx()
	outH := rect.Dy()
	if outW <= 0 || outH <= 0 {
		return nil, nil
	}
	out := make([]byte, outW*outH*4)
	for y := 0; y < outH; y++ {
		for x := 0; x < outW; x++ {
			srcX := rect.Min.X + x
			srcY := rect.Min.Y + y
			srcIdx := (srcY*w + srcX) * 4
			dstIdx := (y*outW + x) * 4
			copy(out[dstIdx:dstIdx+4], soft.tex.pixels[srcIdx:srcIdx+4])
		}
	}
	return out, nil
}

func (d *softwareDriver) ReadTexturePixels(tex Texture, rect image.Rectangle) ([]byte, int, int, error) {
	sw, ok := tex.(*softwareTexture)
	if !ok || sw == nil {
		return nil, 0, 0, ErrUnsupported
	}
	w, h := sw.width, sw.height
	if rect.Empty() {
		return sw.pixels, w, h, nil
	}
	pixels := cropPixels(sw.pixels, w, h, rect)
	if pixels == nil {
		return nil, 0, 0, nil
	}
	return pixels, rect.Dx(), rect.Dy(), nil
}

func (d *softwareDriver) MaxTextureSize() int {
	return 16384
}

func (d *softwareDriver) next() uint32 {
	d.nextID++
	return d.nextID
}

func (t *softwareTexture) ID() uint32 {
	if t == nil {
		return 0
	}
	return t.id
}

func (t *softwareTexture) Size() (int, int) {
	if t == nil {
		return 0, 0
	}
	return t.width, t.height
}

func (t *softwareTexture) Upload(pixels []byte, region image.Rectangle) {
	if t == nil || len(pixels) == 0 {
		return
	}
	w, h := t.width, t.height
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
	regionW := region.Dx()
	regionH := region.Dy()
	if regionW <= 0 || regionH <= 0 {
		return
	}
	for y := 0; y < regionH; y++ {
		for x := 0; x < regionW; x++ {
			srcIdx := (y*regionW + x) * 4
			dstX := region.Min.X + x
			dstY := region.Min.Y + y
			dstIdx := (dstY*w + dstX) * 4
			if srcIdx+3 >= len(pixels) || dstIdx+3 >= len(t.pixels) {
				continue
			}
			copy(t.pixels[dstIdx:dstIdx+4], pixels[srcIdx:srcIdx+4])
		}
	}
}

func (t *softwareTexture) Dispose() {
	if t == nil {
		return
	}
	t.pixels = nil
}

func (f *softwareFramebuffer) ID() uint32 {
	if f == nil {
		return 0
	}
	return f.id
}

func (f *softwareFramebuffer) Size() (int, int) {
	if f == nil || f.tex == nil {
		return 0, 0
	}
	return f.tex.width, f.tex.height
}

func (f *softwareFramebuffer) Bind() {
	if f == nil || f.driver == nil {
		return
	}
	f.driver.current = f
}

func (f *softwareFramebuffer) Texture() Texture {
	if f == nil {
		return nil
	}
	return f.tex
}

func (f *softwareFramebuffer) Dispose() {
	if f == nil {
		return
	}
	if f.tex != nil {
		f.tex.Dispose()
	}
}

func (s *softwareShader) ID() uint32 {
	if s == nil {
		return 0
	}
	return s.id
}

func (s *softwareShader) SetUniform(name string, value any) {
	if s == nil {
		return
	}
	if s.uniforms == nil {
		s.uniforms = make(map[string]any)
	}
	s.uniforms[name] = value
}

func (s *softwareShader) Dispose() {}
