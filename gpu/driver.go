package gpu

import (
	"errors"
	"image"
	"runtime"
)

// Backend identifies a driver backend.
type Backend int

const (
	BackendAuto Backend = iota
	BackendOpenGL
	BackendMetal
	BackendSoftware
)

// ErrUnsupported is returned when a backend is unavailable.
var ErrUnsupported = errors.New("gpu backend unsupported")

// Driver provides low-level rendering operations.
type Driver interface {
	Backend() Backend
	Init() error
	Dispose()

	NewTexture(width, height int) (Texture, error)
	NewFramebuffer(width, height int) (Framebuffer, error)
	NewShader(src ShaderSource) (Shader, error)

	Clear(r, g, b, a float32)
	Draw(call DrawCall)
	ReadPixels(fb Framebuffer, rect image.Rectangle) ([]byte, error)

	MaxTextureSize() int
}

// NewDriver returns a driver for the requested backend.
func NewDriver(backend Backend) (Driver, error) {
	switch backend {
	case BackendOpenGL:
		return newOpenGLDriver()
	case BackendMetal:
		return newMetalDriver()
	case BackendSoftware:
		return newSoftwareDriver(), nil
	case BackendAuto:
		if runtime.GOOS == "darwin" {
			if drv, err := newMetalDriver(); err == nil {
				return drv, nil
			}
		}
		if drv, err := newOpenGLDriver(); err == nil {
			return drv, nil
		}
		return newSoftwareDriver(), nil
	default:
		return newSoftwareDriver(), nil
	}
}
