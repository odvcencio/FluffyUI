package terminal

import (
	"fmt"
	"os"
	"strings"
)

// TerminalCaps describes the terminal's feature support, detected from
// environment variables at startup. This enables progressive enhancement:
// features like synchronized output, cursor shapes, or hyperlinks are only
// used when the terminal is known to support them.
type TerminalCaps struct {
	TrueColor       bool   // 24-bit color support
	SynchronizedOut bool   // CSI ?2026h support
	CursorShapes    bool   // DECSCUSR support
	OSC8Hyperlinks  bool   // OSC-8 hyperlink support
	Sixel           bool   // Sixel graphics
	KittyGraphics   bool   // Kitty image protocol
	BracketedPaste  bool   // Bracketed paste mode
	MouseSGR        bool   // SGR mouse protocol
	UnicodeWidth    bool   // Full Unicode width support
	OSC52Clipboard  bool   // OSC-52 clipboard support
	TermProgram     string // TERM_PROGRAM value
	Term            string // TERM value
}

// DetectCaps probes environment variables and returns a TerminalCaps
// describing the current terminal's feature support. This uses only
// environment variable inspection (no DA1/DA2 escape sequences).
func DetectCaps() *TerminalCaps {
	term := os.Getenv("TERM")
	termProgram := os.Getenv("TERM_PROGRAM")
	colorterm := os.Getenv("COLORTERM")

	termLower := strings.ToLower(term)
	termProgramLower := strings.ToLower(termProgram)
	colortermLower := strings.ToLower(colorterm)

	caps := &TerminalCaps{
		Term:        term,
		TermProgram: termProgram,
	}

	isDumb := termLower == "dumb"

	// Defaults for modern terminals.
	if !isDumb {
		caps.BracketedPaste = true
		caps.MouseSGR = true
		caps.UnicodeWidth = true
	}

	// COLORTERM detection.
	if colortermLower == "truecolor" || colortermLower == "24bit" {
		caps.TrueColor = true
	}

	// TERM_PROGRAM-based detection.
	switch {
	case termProgramLower == "wezterm":
		caps.TrueColor = true
		caps.SynchronizedOut = true
		caps.CursorShapes = true
		caps.OSC8Hyperlinks = true
		caps.KittyGraphics = true
		caps.OSC52Clipboard = true

	case termProgramLower == "iterm.app" || termProgramLower == "iterm2":
		caps.TrueColor = true
		caps.SynchronizedOut = true
		caps.CursorShapes = true
		caps.OSC8Hyperlinks = true
		caps.OSC52Clipboard = true

	case termProgramLower == "kitty":
		caps.TrueColor = true
		caps.SynchronizedOut = true
		caps.CursorShapes = true
		caps.OSC8Hyperlinks = true
		caps.KittyGraphics = true
		caps.OSC52Clipboard = true

	case termProgramLower == "foot":
		caps.TrueColor = true
		caps.SynchronizedOut = true
		caps.CursorShapes = true
		caps.OSC8Hyperlinks = true
		caps.Sixel = true
		caps.OSC52Clipboard = true

	case termProgramLower == "mintty":
		caps.TrueColor = true
		caps.CursorShapes = true
		caps.OSC8Hyperlinks = true
		caps.OSC52Clipboard = true

	case termProgramLower == "contour":
		caps.TrueColor = true
		caps.SynchronizedOut = true
		caps.CursorShapes = true
		caps.OSC8Hyperlinks = true
		caps.Sixel = true
		caps.KittyGraphics = true
		caps.OSC52Clipboard = true

	case termProgramLower == "rio":
		caps.TrueColor = true
		caps.CursorShapes = true

	case termProgramLower == "ghostty":
		caps.TrueColor = true
		caps.SynchronizedOut = true
		caps.CursorShapes = true
		caps.OSC8Hyperlinks = true
		caps.KittyGraphics = true
		caps.OSC52Clipboard = true

	case termProgramLower == "alacritty":
		caps.TrueColor = true
		caps.CursorShapes = true
		caps.OSC52Clipboard = true

	case termProgramLower == "vscode" || strings.HasPrefix(termProgramLower, "vscode"):
		caps.TrueColor = true
		caps.OSC8Hyperlinks = true
		caps.CursorShapes = true
		caps.OSC52Clipboard = true
	}

	// TERM-based detection (supplements TERM_PROGRAM).
	if termLower == "xterm-kitty" {
		caps.TrueColor = true
		caps.SynchronizedOut = true
		caps.CursorShapes = true
		caps.OSC8Hyperlinks = true
		caps.KittyGraphics = true
		caps.OSC52Clipboard = true
	}

	if termLower == "foot" || strings.HasPrefix(termLower, "foot-") {
		caps.TrueColor = true
		caps.SynchronizedOut = true
		caps.CursorShapes = true
		caps.OSC8Hyperlinks = true
		caps.Sixel = true
		caps.OSC52Clipboard = true
	}

	// tmux and screen pass OSC-52 through to the outer terminal.
	if strings.HasPrefix(termLower, "tmux") || strings.HasPrefix(termLower, "screen") {
		caps.OSC52Clipboard = true
	}

	// xterm supports OSC-52 natively.
	if strings.HasPrefix(termLower, "xterm") && !caps.OSC52Clipboard {
		caps.OSC52Clipboard = true
	}

	return caps
}

// String returns a human-readable summary of the terminal capabilities,
// suitable for diagnostic output.
func (c *TerminalCaps) String() string {
	return fmt.Sprintf(
		"TrueColor=%t SyncOut=%t CursorShapes=%t OSC8=%t OSC52=%t Sixel=%t KittyGfx=%t BracketPaste=%t MouseSGR=%t Unicode=%t TERM=%q TERM_PROGRAM=%q",
		c.TrueColor,
		c.SynchronizedOut,
		c.CursorShapes,
		c.OSC8Hyperlinks,
		c.OSC52Clipboard,
		c.Sixel,
		c.KittyGraphics,
		c.BracketedPaste,
		c.MouseSGR,
		c.UnicodeWidth,
		c.Term,
		c.TermProgram,
	)
}
