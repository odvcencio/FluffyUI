//go:build js

package mtl

import "errors"

// Stub implementation for WASM builds where Metal is not available.

var errNotSupported = errors.New("mtl: Metal not supported on WASM")

// Load returns an error on WASM.
func Load() error {
	return errNotSupported
}

// SystemDefaultDevice returns an error on WASM.
func SystemDefaultDevice() (uintptr, error) {
	return 0, errNotSupported
}
