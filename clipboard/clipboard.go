// Package clipboard provides clipboard abstractions.
package clipboard

import "sync"

// Clipboard provides read/write clipboard access.
type Clipboard interface {
	Read() (string, error)
	Write(text string) error
	Available() bool
}

// Target supports clipboard operations for focused widgets.
type Target interface {
	ClipboardCopy() (string, bool)
	ClipboardCut() (string, bool)
	ClipboardPaste(text string) bool
}

// Command identifiers for clipboard actions.
const (
	CommandCopy  = "clipboard.copy"
	CommandCut   = "clipboard.cut"
	CommandPaste = "clipboard.paste"
)

// MemoryClipboard is an in-memory clipboard implementation.
type MemoryClipboard struct {
	mu    sync.Mutex
	value string
}

// Read returns the current clipboard value.
func (c *MemoryClipboard) Read() (string, error) {
	if c == nil {
		return "", nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value, nil
}

// Write updates the clipboard value.
func (c *MemoryClipboard) Write(text string) error {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	c.value = text
	c.mu.Unlock()
	return nil
}

// Available reports whether the clipboard is available.
func (c *MemoryClipboard) Available() bool {
	return true
}

// UnavailableClipboard is a no-op clipboard implementation.
type UnavailableClipboard struct{}

// Read returns an empty string.
func (c UnavailableClipboard) Read() (string, error) {
	return "", nil
}

// Write is a no-op.
func (c UnavailableClipboard) Write(text string) error {
	return nil
}

// Available always reports false.
func (c UnavailableClipboard) Available() bool {
	return false
}
