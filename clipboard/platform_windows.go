//go:build windows

package clipboard

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// NewPlatformClipboard creates a clipboard using PowerShell's
// Set-Clipboard/Get-Clipboard on Windows. Returns nil if PowerShell is not
// found.
func NewPlatformClipboard() *PlatformClipboard {
	ps, err := exec.LookPath("powershell.exe")
	if err != nil {
		return nil
	}
	return &PlatformClipboard{
		readCmd:  []string{ps, "-NoProfile", "-Command", "Get-Clipboard"},
		writeCmd: []string{ps, "-NoProfile", "-Command", "Set-Clipboard", "-Value"},
		tool:     "powershell",
	}
}

// Read returns the current clipboard contents using PowerShell Get-Clipboard.
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

	// PowerShell adds a trailing newline.
	return strings.TrimRight(stdout.String(), "\r\n"), nil
}

// Write sets the clipboard contents using PowerShell Set-Clipboard.
func (c *PlatformClipboard) Write(text string) error {
	if c == nil || len(c.writeCmd) == 0 {
		return fmt.Errorf("clipboard: no platform clipboard tool available")
	}

	// Build the full command: Set-Clipboard -Value 'escaped text'
	escaped := strings.ReplaceAll(text, "'", "''")
	script := fmt.Sprintf("Set-Clipboard -Value '%s'", escaped)

	cmd := exec.Command(c.writeCmd[0], "-NoProfile", "-Command", script)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("clipboard: %s write failed: %w", c.tool, err)
	}
	return nil
}

// Available reports whether PowerShell clipboard access is available.
func (c *PlatformClipboard) Available() bool {
	return c != nil && len(c.writeCmd) > 0
}
