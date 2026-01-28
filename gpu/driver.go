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
	BackendWebGL
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
	case BackendWebGL:
		return newWebGLDriver()
	case BackendSoftware:
		return newSoftwareDriver(), nil
	case BackendAuto:
		return newAutoDriver()
	default:
		return newSoftwareDriver(), nil
	}
}

// newAutoDriver selects the best available backend for the platform.
func newAutoDriver() (Driver, error) {
	if runtime.GOOS == "darwin" {
		if drv, err := newMetalDriver(); err == nil {
			return drv, nil
		}
	}
	// Try WebGL first on JS/WASM platforms
	if runtime.GOOS == "js" {
		if drv, err := newWebGLDriver(); err == nil {
			return drv, nil
		}
		return newSoftwareDriver(), nil
	}
	if drv, err := newOpenGLDriver(); err == nil {
		return drv, nil
	}
	return newSoftwareDriver(), nil
}
