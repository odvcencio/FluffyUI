//go:build !linux && !darwin && !windows

package ghostty

import (
	"errors"

	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/terminal"
)

// Backend is a no-op placeholder for unsupported platforms.
type Backend struct{}

// New reports that ghostty is unsupported on this platform.
func New() (*Backend, error) {
	return nil, errors.New("ghostty backend unsupported on this platform")
}

// Init returns an error since ghostty is unsupported.
func (b *Backend) Init() error {
	return errors.New("ghostty backend unsupported on this platform")
}

// Fini is a no-op.
func (b *Backend) Fini() {}

// Size returns zeros on unsupported platforms.
func (b *Backend) Size() (width, height int) { return 0, 0 }

// SetContent is a no-op.
func (b *Backend) SetContent(x, y int, mainc rune, comb []rune, style backend.Style) {}

// Show is a no-op.
func (b *Backend) Show() {}

// Clear is a no-op.
func (b *Backend) Clear() {}

// HideCursor is a no-op.
func (b *Backend) HideCursor() {}

// ShowCursor is a no-op.
func (b *Backend) ShowCursor() {}

// SetCursorPos is a no-op.
func (b *Backend) SetCursorPos(x, y int) {}

// PollEvent returns nil on unsupported platforms.
func (b *Backend) PollEvent() terminal.Event { return nil }

// PostEvent returns an error on unsupported platforms.
func (b *Backend) PostEvent(ev terminal.Event) error {
	return errors.New("ghostty backend unsupported on this platform")
}

// Beep is a no-op.
func (b *Backend) Beep() {}

// Sync is a no-op.
func (b *Backend) Sync() {}

// Ensure Backend implements backend.Backend.
var _ backend.Backend = (*Backend)(nil)
