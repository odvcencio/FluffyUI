package style

import (
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/compositor"
)

// Color represents a terminal color.
// This is an alias to compositor.Color for compatibility.
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

// Bool creates a pointer to a bool.
func Bool(value bool) *bool {
	v := value
	return &v
}

// Spacing represents top/right/bottom/left spacing.
type Spacing struct {
	Top, Right, Bottom, Left int
}

// Pad creates uniform spacing.
func Pad(all int) *Spacing {
	return &Spacing{Top: all, Right: all, Bottom: all, Left: all}
}

// PadXY creates horizontal/vertical spacing.
func PadXY(x, y int) *Spacing {
	return &Spacing{Top: y, Right: x, Bottom: y, Left: x}
}

// PadTRBL creates explicit spacing.
func PadTRBL(top, right, bottom, left int) *Spacing {
	return &Spacing{Top: top, Right: right, Bottom: bottom, Left: left}
}

// SizeMode defines sizing behavior.
type SizeMode uint8

const (
	SizeAuto SizeMode = iota
	SizeFixed
	SizePercent
	SizeFill
)

// Size represents a sizing rule.
type Size struct {
	Mode  SizeMode
	Value int
}

// Auto sizes to content.
func Auto() *Size {
	return &Size{Mode: SizeAuto}
}

// Fixed sets a fixed size in cells.
func Fixed(value int) *Size {
	return &Size{Mode: SizeFixed, Value: value}
}

// Percent sets a percentage size.
func Percent(value int) *Size {
	return &Size{Mode: SizePercent, Value: value}
}

// Fill expands to available space.
func Fill() *Size {
	return &Size{Mode: SizeFill}
}

// BorderStyle defines border rendering.
type BorderStyle uint8

const (
	BorderNone BorderStyle = iota
	BorderSingle
	BorderDouble
	BorderRounded
)

// BorderChars defines custom border glyphs.
type BorderChars struct {
	TopLeft, TopRight       rune
	BottomLeft, BottomRight rune
	Horizontal, Vertical    rune
}

// Border defines border styling.
type Border struct {
	Style BorderStyle
	Color Color
	Chars *BorderChars
	// StyleSet/ColorSet distinguish explicit "none" from unset values.
	StyleSet bool
	ColorSet bool
}

// Style defines visual attributes and layout hints.
type Style struct {
	// Colors
	Foreground Color
	Background Color

	// Text attributes
	Bold          *bool
	Italic        *bool
	Underline     *bool
	Dim           *bool
	Blink         *bool
	Reverse       *bool
	Strikethrough *bool

	// Layout
	Padding *Spacing
	Margin  *Spacing
	Width   *Size
	Height  *Size

	// Borders
	Border       *Border
	BorderRadius *bool
}

// IsZero reports whether no style fields are set.
func (s Style) IsZero() bool {
	return s.Foreground.Mode == compositor.ColorModeNone &&
		s.Background.Mode == compositor.ColorModeNone &&
		s.Bold == nil &&
		s.Italic == nil &&
		s.Underline == nil &&
		s.Dim == nil &&
		s.Blink == nil &&
		s.Reverse == nil &&
		s.Strikethrough == nil &&
		s.Padding == nil &&
		s.Margin == nil &&
		s.Width == nil &&
		s.Height == nil &&
		s.Border == nil &&
		s.BorderRadius == nil
}

// AffectsLayout reports whether the style includes layout-affecting fields.
func (s Style) AffectsLayout() bool {
	return s.Padding != nil ||
		s.Margin != nil ||
		s.Width != nil ||
		s.Height != nil ||
		s.Border != nil ||
		s.BorderRadius != nil
}

// Merge overlays the provided style on top of the current one.
func (s Style) Merge(override Style) Style {
	if override.Foreground.Mode != compositor.ColorModeNone {
		s.Foreground = override.Foreground
	}
	if override.Background.Mode != compositor.ColorModeNone {
		s.Background = override.Background
	}
	if override.Bold != nil {
		s.Bold = override.Bold
	}
	if override.Italic != nil {
		s.Italic = override.Italic
	}
	if override.Underline != nil {
		s.Underline = override.Underline
	}
	if override.Dim != nil {
		s.Dim = override.Dim
	}
	if override.Blink != nil {
		s.Blink = override.Blink
	}
	if override.Reverse != nil {
		s.Reverse = override.Reverse
	}
	if override.Strikethrough != nil {
		s.Strikethrough = override.Strikethrough
	}
	if override.Padding != nil {
		s.Padding = override.Padding
	}
	if override.Margin != nil {
		s.Margin = override.Margin
	}
	if override.Width != nil {
		s.Width = override.Width
	}
	if override.Height != nil {
		s.Height = override.Height
	}
	if override.Border != nil {
		s.Border = mergeBorder(s.Border, override.Border)
	}
	if override.BorderRadius != nil {
		s.BorderRadius = override.BorderRadius
	}
	return s
}

// Inherit fills unset, inheritable fields from the parent style.
func (s Style) Inherit(parent Style) Style {
	if s.Foreground.Mode == compositor.ColorModeNone {
		s.Foreground = parent.Foreground
	}
	if s.Background.Mode == compositor.ColorModeNone {
		s.Background = parent.Background
	}
	if s.Bold == nil {
		s.Bold = parent.Bold
	}
	if s.Italic == nil {
		s.Italic = parent.Italic
	}
	if s.Underline == nil {
		s.Underline = parent.Underline
	}
	if s.Dim == nil {
		s.Dim = parent.Dim
	}
	if s.Blink == nil {
		s.Blink = parent.Blink
	}
	if s.Reverse == nil {
		s.Reverse = parent.Reverse
	}
	if s.Strikethrough == nil {
		s.Strikethrough = parent.Strikethrough
	}
	return s
}

func mergeBorder(base *Border, override *Border) *Border {
	if override == nil {
		return base
	}
	if base == nil {
		clone := *override
		return &clone
	}
	merged := *base
	if borderStyleSpecified(*override) {
		merged.Style = override.Style
		merged.StyleSet = override.StyleSet || override.Style != BorderNone
	}
	if borderColorSpecified(*override) {
		merged.Color = override.Color
		merged.ColorSet = override.ColorSet || override.Color.Mode != ColorNone.Mode
	}
	if override.Chars != nil {
		merged.Chars = override.Chars
	}
	return &merged
}

func borderStyleSpecified(border Border) bool {
	return border.StyleSet || border.Style != BorderNone
}

func borderColorSpecified(border Border) bool {
	return border.ColorSet || border.Color.Mode != ColorNone.Mode
}

// ToCompositor converts a Style to a compositor.Style.
func (s Style) ToCompositor() compositor.Style {
	style := compositor.DefaultStyle()
	if s.Foreground.Mode != compositor.ColorModeNone {
		style.FG = s.Foreground
	}
	if s.Background.Mode != compositor.ColorModeNone {
		style.BG = s.Background
	}
	if s.Bold != nil {
		style.Bold = *s.Bold
	}
	if s.Italic != nil {
		style.Italic = *s.Italic
	}
	if s.Underline != nil {
		style.Underline = *s.Underline
	}
	if s.Dim != nil {
		style.Dim = *s.Dim
	}
	if s.Blink != nil {
		style.Blink = *s.Blink
	}
	if s.Reverse != nil {
		style.Reverse = *s.Reverse
	}
	if s.Strikethrough != nil {
		style.Strikethrough = *s.Strikethrough
	}
	return style
}

// ToBackend converts a Style to a backend.Style.
func (s Style) ToBackend() backend.Style {
	return ToBackend(s.ToCompositor())
}

// FromBackend converts a backend.Style to a Style.
func FromBackend(bs backend.Style) Style {
	fg, bg, attrs := bs.Decompose()
	out := Style{}
	if fg != backend.ColorDefault {
		out.Foreground = colorFromBackend(fg)
	}
	if bg != backend.ColorDefault {
		out.Background = colorFromBackend(bg)
	}
	if attrs&backend.AttrBold != 0 {
		out.Bold = Bool(true)
	}
	if attrs&backend.AttrItalic != 0 {
		out.Italic = Bool(true)
	}
	if attrs&backend.AttrUnderline != 0 {
		out.Underline = Bool(true)
	}
	if attrs&backend.AttrDim != 0 {
		out.Dim = Bool(true)
	}
	if attrs&backend.AttrBlink != 0 {
		out.Blink = Bool(true)
	}
	if attrs&backend.AttrReverse != 0 {
		out.Reverse = Bool(true)
	}
	if attrs&backend.AttrStrikeThrough != 0 {
		out.Strikethrough = Bool(true)
	}
	return out
}

func colorFromBackend(c backend.Color) Color {
	if c == backend.ColorDefault || int32(c) < 0 {
		return ColorNone
	}
	if c.IsRGB() {
		r, g, b := c.RGB()
		return RGB(r, g, b)
	}
	value := int(c)
	if value < 0 {
		return ColorNone
	}
	if value <= 15 {
		return Color{Mode: compositor.ColorMode16, Value: uint32(value)}
	}
	if value <= 255 {
		return Color256(uint8(value))
	}
	return ColorNone
}
