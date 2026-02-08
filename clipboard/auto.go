package clipboard

import "io"

// AutoClipboard selects the best available clipboard implementation using
// the following priority:
//
//  1. OSC-52 (works through SSH, tmux, screen)
//  2. Platform-native (xclip, pbcopy, PowerShell, etc.)
//  3. In-memory fallback (always available)
//
// For writes, OSC-52 is preferred when available because it works in remote
// sessions. For reads, the platform clipboard is preferred because OSC-52 read
// is not reliably supported by most terminals.
type AutoClipboard struct {
	primary  Clipboard // preferred clipboard for writes (OSC-52 or platform)
	reader   Clipboard // preferred clipboard for reads (platform or memory)
	fallback Clipboard // always-available in-memory fallback
}

// NewAutoClipboard creates a clipboard that auto-selects the best available
// implementation. The writer w is used for OSC-52 escape sequences (typically
// os.Stdout). Pass nil to skip OSC-52 detection.
func NewAutoClipboard(w io.Writer) *AutoClipboard {
	ac := &AutoClipboard{
		fallback: &MemoryClipboard{},
	}

	// Try platform clipboard (supports both read and write).
	platform := NewPlatformClipboard()

	// Try OSC-52 (write-only in practice).
	var osc *OSC52Clipboard
	if w != nil {
		osc = NewOSC52Clipboard(w, nil)
	}

	// Select primary (for writes): prefer OSC-52 if available, then platform.
	switch {
	case osc != nil && osc.Available():
		ac.primary = osc
	case platform != nil && platform.Available():
		ac.primary = platform
	default:
		ac.primary = ac.fallback
	}

	// Select reader: prefer platform (actual read support), then fallback.
	switch {
	case platform != nil && platform.Available():
		ac.reader = platform
	default:
		ac.reader = ac.fallback
	}

	return ac
}

// Read returns the clipboard contents. Prefers platform-native tools for
// reading since OSC-52 read is not reliably supported.
func (c *AutoClipboard) Read() (string, error) {
	if c == nil {
		return "", nil
	}
	return c.reader.Read()
}

// Write sets the clipboard contents. Prefers OSC-52 when available because
// it works through SSH and multiplexers.
func (c *AutoClipboard) Write(text string) error {
	if c == nil {
		return nil
	}

	// Write to primary clipboard.
	err := c.primary.Write(text)

	// Also write to fallback memory clipboard so Read() works even if
	// platform tools are unavailable.
	if c.fallback != nil && c.primary != c.fallback {
		_ = c.fallback.Write(text)
	}

	return err
}

// Available reports whether any clipboard implementation is available.
// Always returns true for non-nil instances because the in-memory fallback
// is always available.
func (c *AutoClipboard) Available() bool {
	if c == nil {
		return false
	}
	return true
}

// Primary returns the clipboard selected for write operations, useful for
// diagnostics and testing.
func (c *AutoClipboard) Primary() Clipboard {
	if c == nil {
		return nil
	}
	return c.primary
}

// Reader returns the clipboard selected for read operations.
func (c *AutoClipboard) Reader() Clipboard {
	if c == nil {
		return nil
	}
	return c.reader
}
