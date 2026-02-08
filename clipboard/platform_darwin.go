//go:build darwin

package clipboard

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// NewPlatformClipboard creates a clipboard using pbcopy/pbpaste on macOS.
// Returns nil if the tools are not found (should never happen on macOS).
func NewPlatformClipboard() *PlatformClipboard {
	pbcopy, err := exec.LookPath("pbcopy")
	if err != nil {
		return nil
	}
	pbpaste, err := exec.LookPath("pbpaste")
	if err != nil {
		return nil
	}
	return &PlatformClipboard{
		readCmd:  []string{pbpaste},
		writeCmd: []string{pbcopy},
		tool:     "pbcopy",
	}
}

// Read returns the current clipboard contents using pbpaste.
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

// Write sets the clipboard contents using pbcopy.
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

// Available reports whether pbcopy/pbpaste are available.
func (c *PlatformClipboard) Available() bool {
	return c != nil && len(c.writeCmd) > 0
}
