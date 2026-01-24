package terminal

import (
	"os"
	"strings"
)

// Capabilities describes terminal rendering features.
type Capabilities struct {
	TrueColor bool
	Sixel     bool
	Kitty     bool
	Unicode   bool
}

// DetectCapabilities inspects environment variables to infer terminal support.
func DetectCapabilities() Capabilities {
	term := strings.ToLower(os.Getenv("TERM"))
	termProgram := strings.ToLower(os.Getenv("TERM_PROGRAM"))
	colorterm := strings.ToLower(os.Getenv("COLORTERM"))
	lang := strings.ToLower(os.Getenv("LANG"))
	lcAll := strings.ToLower(os.Getenv("LC_ALL"))
	kittyWindowID := os.Getenv("KITTY_WINDOW_ID")
	weztermPane := os.Getenv("WEZTERM_PANE")

	trueColor := strings.Contains(colorterm, "truecolor") || strings.Contains(colorterm, "24bit")
	unicode := strings.Contains(lang, "utf-8") || strings.Contains(lcAll, "utf-8")

	// Kitty detection: KITTY_WINDOW_ID is the most reliable indicator
	kitty := kittyWindowID != "" ||
		strings.Contains(term, "kitty") ||
		termProgram == "kitty"

	// WezTerm supports both Kitty graphics and Sixel
	wezterm := weztermPane != "" ||
		strings.Contains(termProgram, "wezterm") ||
		strings.Contains(term, "wezterm")

	if wezterm {
		kitty = true // WezTerm supports Kitty graphics protocol
	}

	// Sixel detection: check known Sixel-capable terminals
	// Note: xterm can support sixel but requires --enable-sixel compile flag,
	// so we don't assume it by default
	sixel := wezterm ||
		strings.Contains(term, "sixel") ||
		strings.Contains(termProgram, "mlterm") ||
		termProgram == "mintty" ||
		termProgram == "foot" ||
		termProgram == "contour" ||
		strings.Contains(term, "mlterm")

	return Capabilities{
		TrueColor: trueColor,
		Sixel:     sixel,
		Kitty:     kitty,
		Unicode:   unicode,
	}
}
