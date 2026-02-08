package clipboard

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/odvcencio/fluffyui/terminal"
)

// OSC52Clipboard writes to the system clipboard using the OSC-52 escape
// sequence, which is supported by most modern terminal emulators and works
// transparently through SSH sessions, tmux, and screen.
//
// Write sends: \033]52;c;<base64-data>\a
//
// Read is not supported in most terminal implementations because it requires
// the terminal to respond with the clipboard contents, which few terminals
// allow for security reasons.
type OSC52Clipboard struct {
	mu   sync.Mutex
	w    io.Writer
	caps *terminal.TerminalCaps
}

// NewOSC52Clipboard creates a clipboard that uses OSC-52 escape sequences
// written to w (typically os.Stdout). If caps is nil, terminal capabilities
// are detected automatically.
func NewOSC52Clipboard(w io.Writer, caps *terminal.TerminalCaps) *OSC52Clipboard {
	if caps == nil {
		caps = terminal.DetectCaps()
	}
	return &OSC52Clipboard{w: w, caps: caps}
}

// Read returns an error because OSC-52 read is not reliably supported.
// Most terminals either ignore the read request or have it disabled by default
// for security reasons (clipboard content could be exfiltrated).
func (c *OSC52Clipboard) Read() (string, error) {
	return "", errors.New("clipboard: OSC-52 read is not supported; most terminals disable it for security")
}

// Write sends the text to the system clipboard via the OSC-52 escape sequence.
// The text is base64-encoded and wrapped in the escape: \033]52;c;BASE64\a
func (c *OSC52Clipboard) Write(text string) error {
	if c == nil || c.w == nil {
		return errors.New("clipboard: OSC-52 writer is nil")
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	seq := fmt.Sprintf("\033]52;c;%s\a", encoded)

	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := io.WriteString(c.w, seq)
	return err
}

// Available reports whether the terminal is known to support OSC-52.
// Returns true if terminal capabilities indicate OSC-52 support, or if
// capabilities could not be determined (optimistic default for modern terminals).
func (c *OSC52Clipboard) Available() bool {
	if c == nil {
		return false
	}
	if c.caps != nil {
		return c.caps.OSC52Clipboard
	}
	return true
}
