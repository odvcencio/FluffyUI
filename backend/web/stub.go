//go:build js

package web

import "fmt"

// Backend is not available in WebAssembly builds.
type Backend struct{}

// New returns an error in WebAssembly builds.
func New(config Config) *Backend {
	return nil
}

// NewWithAddr returns an error in WebAssembly builds.
func NewWithAddr(addr string) *Backend {
	return nil
}

// Init returns an error in WebAssembly builds.
func (b *Backend) Init() error {
	return fmt.Errorf("web backend not available in WebAssembly")
}

// Fini is a no-op in WebAssembly builds.
func (b *Backend) Fini() {}

// Size returns zero dimensions in WebAssembly builds.
func (b *Backend) Size() (width, height int) {
	return 0, 0
}

// SetContent is a no-op in WebAssembly builds.
func (b *Backend) SetContent(x, y int, mainc rune, comb []rune, style any) {}

// SetRow is a no-op in WebAssembly builds.
func (b *Backend) SetRow(y int, startX int, cells []any) {}

// SetRect is a no-op in WebAssembly builds.
func (b *Backend) SetRect(x, y, width, height int, cells []any) {}

// Show is a no-op in WebAssembly builds.
func (b *Backend) Show() {}

// Clear is a no-op in WebAssembly builds.
func (b *Backend) Clear() {}

// HideCursor is a no-op in WebAssembly builds.
func (b *Backend) HideCursor() {}

// ShowCursor is a no-op in WebAssembly builds.
func (b *Backend) ShowCursor() {}

// SetCursorPos is a no-op in WebAssembly builds.
func (b *Backend) SetCursorPos(x, y int) {}

// PollEvent returns nil in WebAssembly builds.
func (b *Backend) PollEvent() any {
	return nil
}

// PostEvent returns an error in WebAssembly builds.
func (b *Backend) PostEvent(ev any) error {
	return fmt.Errorf("web backend not available in WebAssembly")
}

// Beep is a no-op in WebAssembly builds.
func (b *Backend) Beep() {}

// Sync is a no-op in WebAssembly builds.
func (b *Backend) Sync() {}

// Addr returns nil in WebAssembly builds.
func (b *Backend) Addr() any {
	return nil
}

// URL returns empty string in WebAssembly builds.
func (b *Backend) URL() string {
	return ""
}
