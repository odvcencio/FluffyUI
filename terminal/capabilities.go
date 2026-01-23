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

	trueColor := strings.Contains(colorterm, "truecolor") || strings.Contains(colorterm, "24bit")
	unicode := strings.Contains(lang, "utf-8") || strings.Contains(lcAll, "utf-8")

	kitty := strings.Contains(term, "kitty") || termProgram == "kitty"
	sixel := strings.Contains(term, "sixel") || strings.Contains(term, "mlterm") || strings.Contains(term, "xterm")

	return Capabilities{
		TrueColor: trueColor,
		Sixel:     sixel,
		Kitty:     kitty,
		Unicode:   unicode,
	}
}
