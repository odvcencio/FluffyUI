package fur

import (
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/compositor"
	uiStyle "github.com/odvcencio/fluffyui/style"
)

// Color represents a terminal color.
type Color = compositor.Color

// Color helpers.
var (
	ColorNone    = compositor.ColorNone
	ColorDefault = compositor.ColorDefault

	ColorBlack   = compositor.ColorBlack
	ColorRed     = compositor.ColorRed
	ColorGreen   = compositor.ColorGreen
	ColorYellow  = compositor.ColorYellow
	ColorBlue    = compositor.ColorBlue
	ColorMagenta = compositor.ColorMagenta
	ColorCyan    = compositor.ColorCyan
	ColorWhite   = compositor.ColorWhite

	ColorBrightBlack   = compositor.ColorBrightBlack
	ColorBrightRed     = compositor.ColorBrightRed
	ColorBrightGreen   = compositor.ColorBrightGreen
	ColorBrightYellow  = compositor.ColorBrightYellow
	ColorBrightBlue    = compositor.ColorBrightBlue
	ColorBrightMagenta = compositor.ColorBrightMagenta
	ColorBrightCyan    = compositor.ColorBrightCyan
	ColorBrightWhite   = compositor.ColorBrightWhite
)

// Color256 creates a 256-palette color (0-255).
func Color256(index uint8) Color {
	return compositor.Color256(index)
}

// RGB creates a 24-bit true color.
func RGB(r, g, b uint8) Color {
	return compositor.RGB(r, g, b)
}

// Hex creates a color from hex value (0xRRGGBB).
func Hex(hex uint32) Color {
	return compositor.Hex(hex)
}

// Convenience color aliases.
var (
	Black   = ColorBlack
	Red     = ColorRed
	Green   = ColorGreen
	Yellow  = ColorYellow
	Blue    = ColorBlue
	Magenta = ColorMagenta
	Cyan    = ColorCyan
	White   = ColorWhite

	BrightBlack   = ColorBrightBlack
	BrightRed     = ColorBrightRed
	BrightGreen   = ColorBrightGreen
	BrightYellow  = ColorBrightYellow
	BrightBlue    = ColorBrightBlue
	BrightMagenta = ColorBrightMagenta
	BrightCyan    = ColorBrightCyan
	BrightWhite   = ColorBrightWhite
)

// Style defines text styling used by fur renderables.
type Style struct {
	fg            Color
	bg            Color
	bold          bool
	dim           bool
	italic        bool
	underline     bool
	blink         bool
	reverse       bool
	strikethrough bool
}

// DefaultStyle returns a style with default colors and no attributes.
func DefaultStyle() Style {
	return Style{fg: ColorDefault, bg: ColorDefault}
}

// Foreground sets the foreground color.
func (s Style) Foreground(c Color) Style {
	s.fg = c
	return s
}

// Background sets the background color.
func (s Style) Background(c Color) Style {
	s.bg = c
	return s
}

// Bold enables bold.
func (s Style) Bold() Style {
	s.bold = true
	return s
}

// Dim enables dim.
func (s Style) Dim() Style {
	s.dim = true
	return s
}

// Italic enables italic.
func (s Style) Italic() Style {
	s.italic = true
	return s
}

// Underline enables underline.
func (s Style) Underline() Style {
	s.underline = true
	return s
}

// Blink enables blink.
func (s Style) Blink() Style {
	s.blink = true
	return s
}

// Reverse enables reverse video.
func (s Style) Reverse() Style {
	s.reverse = true
	return s
}

// Strikethrough enables strikethrough.
func (s Style) Strikethrough() Style {
	s.strikethrough = true
	return s
}

// Equal reports whether two styles are identical.
func (s Style) Equal(other Style) bool {
	return s.fg == other.fg &&
		s.bg == other.bg &&
		s.bold == other.bold &&
		s.dim == other.dim &&
		s.italic == other.italic &&
		s.underline == other.underline &&
		s.blink == other.blink &&
		s.reverse == other.reverse &&
		s.strikethrough == other.strikethrough
}

// ToCompositor converts a fur style into a compositor style.
func (s Style) ToCompositor() compositor.Style {
	cs := compositor.DefaultStyle()
	cs.FG = s.fg
	cs.BG = s.bg
	cs.Bold = s.bold
	cs.Dim = s.dim
	cs.Italic = s.italic
	cs.Underline = s.underline
	cs.Blink = s.blink
	cs.Reverse = s.reverse
	cs.Strikethrough = s.strikethrough
	return cs
}

// FromCompositor converts a compositor style into a fur style.
func FromCompositor(cs compositor.Style) Style {
	return Style{
		fg:            cs.FG,
		bg:            cs.BG,
		bold:          cs.Bold,
		dim:           cs.Dim,
		italic:        cs.Italic,
		underline:     cs.Underline,
		blink:         cs.Blink,
		reverse:       cs.Reverse,
		strikethrough: cs.Strikethrough,
	}
}

// ToBackend converts a fur style into a backend style.
func (s Style) ToBackend() backend.Style {
	return uiStyle.ToBackend(s.ToCompositor())
}

// Style helpers.
var (
	Bold          = DefaultStyle().Bold()
	Italic        = DefaultStyle().Italic()
	Underline     = DefaultStyle().Underline()
	Dim           = DefaultStyle().Dim()
	Strikethrough = DefaultStyle().Strikethrough()
)
