//go:build js

package gpu

// newOpenGLDriver returns ErrUnsupported on WASM since OpenGL is not available.
func newOpenGLDriver() (Driver, error) {
	return nil, ErrUnsupported
}
