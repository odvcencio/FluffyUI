//go:build linux

package gl

import "github.com/ebitengine/purego"

func libGLName() string {
	return "libGL.so.1"
}

func openLibrary(name string) (uintptr, error) {
	if name == "" {
		name = libGLName()
	}
	return purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_LOCAL)
}

func loadSymbol(handle uintptr, name string) (uintptr, error) {
	return purego.Dlsym(handle, name)
}

func initProcAddress(handle uintptr) error {
	procAddr = func(name string) uintptr {
		if handle == 0 {
			return 0
		}
		if sym, err := loadSymbol(handle, name); err == nil {
			return sym
		}
		return 0
	}
	return nil
}
