//go:build !js

package gpu

// newWebGLDriver returns ErrUnsupported on non-WASM builds.
// WebGL is only available in browser environments.
func newWebGLDriver() (Driver, error) {
	return nil, ErrUnsupported
}
