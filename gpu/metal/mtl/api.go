//go:build !js

package mtl

import (
	"errors"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

var (
	loadOnce sync.Once
	loadErr  error
	handle   uintptr

	createSystemDefaultDevice func() unsafe.Pointer
)

// Load loads the Metal framework and resolves symbols.
func Load() error {
	loadOnce.Do(func() {
		h, err := purego.Dlopen("/System/Library/Frameworks/Metal.framework/Metal", purego.RTLD_NOW|purego.RTLD_LOCAL)
		if err != nil {
			loadErr = err
			return
		}
		handle = h
		if err := register(); err != nil {
			loadErr = err
		}
	})
	return loadErr
}

// SystemDefaultDevice returns the system default Metal device pointer.
func SystemDefaultDevice() (uintptr, error) {
	if err := Load(); err != nil {
		return 0, err
	}
	if createSystemDefaultDevice == nil {
		return 0, errors.New("mtl: missing MTLCreateSystemDefaultDevice")
	}
	device := createSystemDefaultDevice()
	return uintptr(device), nil
}

func register() error {
	if handle == 0 {
		return errors.New("mtl: Metal not loaded")
	}
	purego.RegisterLibFunc(&createSystemDefaultDevice, handle, "MTLCreateSystemDefaultDevice")
	return nil
}
