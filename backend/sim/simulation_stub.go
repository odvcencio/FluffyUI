//go:build js

// Package sim provides a stub simulation backend for WASM builds.
package sim

import (
	"errors"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/terminal"
)

var errNotSupported = errors.New("sim backend not supported on WASM")

// Backend is a stub implementation for WASM.
type Backend struct{}

// New returns a stub backend on WASM.
func New(width, height int) *Backend {
	return &Backend{}
}

// Init returns an error on WASM.
func (b *Backend) Init() error {
	return errNotSupported
}

// Fini is a no-op on WASM.
func (b *Backend) Fini() {}

// Size returns 0 on WASM.
func (b *Backend) Size() (width, height int) {
	return 0, 0
}

// SetContent is a no-op on WASM.
func (b *Backend) SetContent(x, y int, mainc rune, comb []rune, style backend.Style) {}

// SetRow is a no-op on WASM.
func (b *Backend) SetRow(y int, startX int, cells []backend.Cell) {}

// SetRect is a no-op on WASM.
func (b *Backend) SetRect(x, y, width, height int, cells []backend.Cell) {}

// Show is a no-op on WASM.
func (b *Backend) Show() {}

// Clear is a no-op on WASM.
func (b *Backend) Clear() {}

// HideCursor is a no-op on WASM.
func (b *Backend) HideCursor() {}

// ShowCursor is a no-op on WASM.
func (b *Backend) ShowCursor() {}

// SetCursorPos is a no-op on WASM.
func (b *Backend) SetCursorPos(x, y int) {}

// PollEvent returns nil on WASM.
func (b *Backend) PollEvent() terminal.Event {
	return nil
}

// PostEvent returns an error on WASM.
func (b *Backend) PostEvent(ev terminal.Event) error {
	return errNotSupported
}

// Beep is a no-op on WASM.
func (b *Backend) Beep() {}

// Sync is a no-op on WASM.
func (b *Backend) Sync() {}

// Screen returns nil on WASM.
func (b *Backend) Screen() interface{} {
	return nil
}
