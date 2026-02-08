//go:build linux

package clipboard

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// NewPlatformClipboard creates a clipboard using the best available
// platform tool. On Linux it tries xclip, xsel, and wl-copy/wl-paste
// (Wayland) in order. Returns nil if no supported tool is found.
func NewPlatformClipboard() *PlatformClipboard {
	// Wayland: prefer wl-copy/wl-paste if WAYLAND_DISPLAY is set.
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		if wlCopy, err := exec.LookPath("wl-copy"); err == nil {
			if wlPaste, err := exec.LookPath("wl-paste"); err == nil {
				return &PlatformClipboard{
					readCmd:  []string{wlPaste, "--no-newline"},
					writeCmd: []string{wlCopy},
					tool:     "wl-copy",
				}
			}
		}
	}

	// X11: try xclip first.
	if path, err := exec.LookPath("xclip"); err == nil {
		return &PlatformClipboard{
			readCmd:  []string{path, "-selection", "clipboard", "-o"},
			writeCmd: []string{path, "-selection", "clipboard", "-i"},
			tool:     "xclip",
		}
	}

	// X11: fall back to xsel.
	if path, err := exec.LookPath("xsel"); err == nil {
		return &PlatformClipboard{
			readCmd:  []string{path, "--clipboard", "--output"},
			writeCmd: []string{path, "--clipboard", "--input"},
			tool:     "xsel",
		}
	}

	// Wayland fallback: try wl-copy/wl-paste even without WAYLAND_DISPLAY
	// (some setups don't export it).
	if wlCopy, err := exec.LookPath("wl-copy"); err == nil {
		if wlPaste, err := exec.LookPath("wl-paste"); err == nil {
			return &PlatformClipboard{
				readCmd:  []string{wlPaste, "--no-newline"},
				writeCmd: []string{wlCopy},
				tool:     "wl-copy",
			}
		}
	}

	return nil
}

// Read returns the current clipboard contents using the detected tool.
func (c *PlatformClipboard) Read() (string, error) {
	if c == nil || len(c.readCmd) == 0 {
		return "", fmt.Errorf("clipboard: no platform clipboard tool available")
	}

	cmd := exec.Command(c.readCmd[0], c.readCmd[1:]...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("clipboard: %s read failed: %w", c.tool, err)
	}
	return stdout.String(), nil
}

// Write sets the clipboard contents using the detected tool.
func (c *PlatformClipboard) Write(text string) error {
	if c == nil || len(c.writeCmd) == 0 {
		return fmt.Errorf("clipboard: no platform clipboard tool available")
	}

	cmd := exec.Command(c.writeCmd[0], c.writeCmd[1:]...)
	cmd.Stdin = strings.NewReader(text)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("clipboard: %s write failed: %w", c.tool, err)
	}
	return nil
}

// Available reports whether a platform clipboard tool was detected.
func (c *PlatformClipboard) Available() bool {
	return c != nil && len(c.writeCmd) > 0
}
