//go:build linux

package gpu

import (
	"errors"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/odvcencio/fluffy-ui/gpu/opengl/gl"
)

const (
	osmesaRGBA       = gl.RGBA
	osmesaFormatType = gl.UNSIGNED_BYTE
)

type osmesaContext struct {
	handle uintptr
	ctx    unsafe.Pointer
	width  int
	height int
	buffer []byte
}

var (
	osmesaCreateContextExt func(format int32, depthBits, stencilBits, accumBits int32, shareCtx unsafe.Pointer) unsafe.Pointer
	osmesaDestroyContext   func(ctx unsafe.Pointer)
	osmesaMakeCurrent      func(ctx unsafe.Pointer, buffer unsafe.Pointer, typ uint32, width, height int32) int32
)

func newOSMesaContext(width, height int) (*osmesaContext, error) {
	h, err := purego.Dlopen("libOSMesa.so.8", purego.RTLD_NOW|purego.RTLD_LOCAL)
	if err != nil {
		h, err = purego.Dlopen("libOSMesa.so.6", purego.RTLD_NOW|purego.RTLD_LOCAL)
		if err != nil {
			h, err = purego.Dlopen("libOSMesa.so", purego.RTLD_NOW|purego.RTLD_LOCAL)
			if err != nil {
				return nil, err
			}
		}
	}
	registerOSMesaFuncs(h)
	ctx := osmesaCreateContextExt(int32(osmesaRGBA), 24, 8, 0, nil)
	if ctx == nil {
		return nil, errors.New("osmesa: create context failed")
	}
	buffer := make([]byte, width*height*4)
	if osmesaMakeCurrent(ctx, unsafe.Pointer(&buffer[0]), osmesaFormatType, int32(width), int32(height)) == 0 {
		osmesaDestroyContext(ctx)
		return nil, errors.New("osmesa: make current failed")
	}
	return &osmesaContext{handle: h, ctx: ctx, width: width, height: height, buffer: buffer}, nil
}

func (c *osmesaContext) Resize(width, height int) error {
	if c == nil || c.ctx == nil {
		return errors.New("osmesa: context missing")
	}
	if width == c.width && height == c.height {
		return nil
	}
	buffer := make([]byte, width*height*4)
	if osmesaMakeCurrent(c.ctx, unsafe.Pointer(&buffer[0]), osmesaFormatType, int32(width), int32(height)) == 0 {
		return errors.New("osmesa: make current failed")
	}
	c.width = width
	c.height = height
	c.buffer = buffer
	return nil
}

func (c *osmesaContext) Buffer() []byte {
	if c == nil {
		return nil
	}
	return c.buffer
}

func (c *osmesaContext) Destroy() {
	if c == nil {
		return
	}
	if c.ctx != nil {
		osmesaDestroyContext(c.ctx)
		c.ctx = nil
	}
	if c.handle != 0 {
		_ = purego.Dlclose(c.handle)
		c.handle = 0
	}
}

func registerOSMesaFuncs(handle uintptr) {
	purego.RegisterLibFunc(&osmesaCreateContextExt, handle, "OSMesaCreateContextExt")
	purego.RegisterLibFunc(&osmesaDestroyContext, handle, "OSMesaDestroyContext")
	purego.RegisterLibFunc(&osmesaMakeCurrent, handle, "OSMesaMakeCurrent")
}
