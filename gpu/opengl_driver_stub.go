//go:build js

package gpu

// newOpenGLDriver returns ErrUnsupported on WASM since OpenGL is not available.
// Use newWebGLDriver instead for WebGL support.
func newOpenGLDriver() (Driver, error) {
	return nil, ErrUnsupported
}

// newWebGLDriver returns a WebGL driver on WASM builds.
// The actual implementation is in webgl_driver.go.
