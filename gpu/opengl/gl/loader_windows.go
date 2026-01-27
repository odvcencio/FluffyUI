//go:build windows

package gl

import (
	"errors"
	"unsafe"

	"github.com/ebitengine/purego"
	"golang.org/x/sys/windows"
)

func libGLName() string {
	return "opengl32.dll"
}

func openLibrary(name string) (uintptr, error) {
	if name == "" {
		name = libGLName()
	}
	h, err := windows.LoadLibrary(name)
	return uintptr(h), err
}

func closeLibrary(handle uintptr) error {
	if handle == 0 {
		return nil
	}
	return windows.FreeLibrary(windows.Handle(handle))
}

func loadSymbol(handle uintptr, name string) (uintptr, error) {
	if handle == 0 {
		return 0, errors.New("library handle missing")
	}
	addr, err := windows.GetProcAddress(windows.Handle(handle), name)
	if err != nil {
		return 0, err
	}
	return addr, nil
}

func initProcAddress(handle uintptr) error {
	var wglGetProcAddress func(*byte) uintptr
	if sym, err := loadSymbol(handle, "wglGetProcAddress"); err == nil && sym != 0 {
		purego.RegisterFunc(&wglGetProcAddress, sym)
	}
	procAddr = func(name string) uintptr {
		if wglGetProcAddress != nil {
			cname := append([]byte(name), 0)
			addr := wglGetProcAddress((*byte)(unsafe.Pointer(&cname[0])))
			if addr != 0 && !isInvalidWGLAddress(addr) {
				return addr
			}
		}
		if sym, err := loadSymbol(handle, name); err == nil {
			return sym
		}
		return 0
	}
	return nil
}

func isInvalidWGLAddress(addr uintptr) bool {
	switch addr {
	case 0, 1, 2, 3, 4, ^uintptr(0):
		return true
	default:
		return false
	}
}
