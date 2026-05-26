package style

import (
	"time"

	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/compositor"
)

// EasingFunc identifies a CSS-like easing function for transitions.
type EasingFunc int

const (
	// EaseLinear applies no acceleration.
	EaseLinear EasingFunc = iota
	// EaseIn starts slow and accelerates (alias for EaseInQuad).
	EaseIn
	// EaseOut starts fast and decelerates (alias for EaseOutQuad).
	EaseOut
	// EaseInOut starts slow, speeds up, then decelerates (alias for EaseInOutQuad).
	EaseInOut

	// Quadratic easing
	EaseInQuad
	EaseOutQuad
	EaseInOutQuad

	// Cubic easing
	EaseInCubic
	EaseOutCubic
	EaseInOutCubic

	// Quartic easing
	EaseInQuart
	EaseOutQuart
	EaseInOutQuart

	// Quintic easing
	EaseInQuint
	EaseOutQuint
	EaseInOutQuint

	// Sinusoidal easing
	EaseInSine
	EaseOutSine
	EaseInOutSine

	// Exponential easing
	EaseInExpo
	EaseOutExpo
	EaseInOutExpo

	// Circular easing
	EaseInCirc
	EaseOutCirc
	EaseInOutCirc

	// Elastic easing (spring-like)
	EaseInElastic
	EaseOutElastic
	EaseInOutElastic

	// Back easing (overshoot)
	EaseInBack
	EaseOutBack
	EaseInOutBack

	// Bounce easing (bouncing ball)
	EaseInBounce
	EaseOutBounce
	EaseInOutBounce
)

// Transition defines a CSS-like property transition.
type Transition struct {
	Property string        // "foreground", "background", "opacity", "all", etc.
	Duration time.Duration // e.g. 200ms
	Easing   EasingFunc    // linear, ease-in, ease-out, ease-in-out
}

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

// TextAlign defines text alignment.
type TextAlign uint8

const (
	TextAlignNone   TextAlign = iota // unset
	TextAlignLeft                    // left-aligned
	TextAlignCenter                  // centered
	TextAlignRight                   // right-aligned
)

// Display defines display behavior.
type Display uint8

const (
	DisplayNone  Display = iota // unset
	DisplayBlock                // block layout
	DisplayFlex                 // flex layout
	DisplayHidden               // hidden (display: none)
)

// Visibility defines element visibility.
type Visibility uint8

const (
	VisibilityNone    Visibility = iota // unset
	VisibilityVisible                   // visible (default)
	VisibilityHidden                    // hidden (reserves space)
)

// Overflow defines overflow behavior.
type Overflow uint8

const (
	OverflowNone    Overflow = iota // unset
	OverflowVisible                 // content overflows
	OverflowHidden                  // content clipped
	OverflowScroll                  // scrollable
)

// CursorStyle defines cursor appearance.
type CursorStyle uint8

const (
	CursorNone    CursorStyle = iota // unset
	CursorDefault                    // default cursor
	CursorPointer                    // pointer (interactive)
	CursorText                       // text selection
)

// TextDecorationStyle defines text decoration.
type TextDecorationStyle uint8

const (
	TextDecorationNone          TextDecorationStyle = iota // unset
	TextDecorationOff                                      // explicitly no decoration
	TextDecorationUnderline                                // underline
	TextDecorationStrikethrough                            // strikethrough
)

// WhiteSpace defines white-space handling.
type WhiteSpace uint8

const (
	WhiteSpaceNone   WhiteSpace = iota // unset
	WhiteSpaceNormal                   // normal wrapping
	WhiteSpaceNowrap                   // no wrapping
	WhiteSpacePre                      // preserve whitespace
)

// ContentAlign defines content alignment within a container.
type ContentAlign uint8

const (
	ContentAlignNone   ContentAlign = iota // unset
	ContentAlignStart                      // start-aligned
	ContentAlignCenter                     // centered
	ContentAlignEnd                        // end-aligned
)

// Dock defines edge pinning for widgets.
type Dock uint8

const (
	DockNone   Dock = iota // unset
	DockTop                // pinned to top
	DockBottom             // pinned to bottom
	DockLeft               // pinned to left
	DockRight              // pinned to right
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

	// Text layout
	TextAlign      TextAlign
	TextDecoration TextDecorationStyle
	WhiteSpace     WhiteSpace

	// Display and visibility
	Display    Display
	Visibility Visibility
	Overflow   Overflow
	Opacity    *float64

	// Cursor
	Cursor CursorStyle

	// Layout constraints
	MinWidth  *int
	MaxWidth  *int
	MinHeight *int
	MaxHeight *int

	// Content alignment
	ContentAlign ContentAlign

	// Edge pinning
	Dock Dock

	// Position offsets
	OffsetX *int
	OffsetY *int

	// Stacking order
	ZIndex *int

	// Transitions
	Transitions []Transition
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
		s.BorderRadius == nil &&
		s.TextAlign == TextAlignNone &&
		s.TextDecoration == TextDecorationNone &&
		s.WhiteSpace == WhiteSpaceNone &&
		s.Display == DisplayNone &&
		s.Visibility == VisibilityNone &&
		s.Overflow == OverflowNone &&
		s.Opacity == nil &&
		s.Cursor == CursorNone &&
		s.MinWidth == nil &&
		s.MaxWidth == nil &&
		s.MinHeight == nil &&
		s.MaxHeight == nil &&
		s.ContentAlign == ContentAlignNone &&
		s.Dock == DockNone &&
		s.OffsetX == nil &&
		s.OffsetY == nil &&
		s.ZIndex == nil &&
		len(s.Transitions) == 0
}

// AffectsLayout reports whether the style includes layout-affecting fields.
func (s Style) AffectsLayout() bool {
	return s.Padding != nil ||
		s.Margin != nil ||
		s.Width != nil ||
		s.Height != nil ||
		s.Border != nil ||
		s.BorderRadius != nil ||
		s.Display != DisplayNone ||
		s.Visibility != VisibilityNone ||
		s.Overflow != OverflowNone ||
		s.WhiteSpace != WhiteSpaceNone ||
		s.TextAlign != TextAlignNone ||
		s.MinWidth != nil ||
		s.MaxWidth != nil ||
		s.MinHeight != nil ||
		s.MaxHeight != nil ||
		s.ContentAlign != ContentAlignNone ||
		s.Dock != DockNone ||
		s.OffsetX != nil ||
		s.OffsetY != nil ||
		s.ZIndex != nil
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
	if override.TextAlign != TextAlignNone {
		s.TextAlign = override.TextAlign
	}
	if override.TextDecoration != TextDecorationNone {
		s.TextDecoration = override.TextDecoration
	}
	if override.WhiteSpace != WhiteSpaceNone {
		s.WhiteSpace = override.WhiteSpace
	}
	if override.Display != DisplayNone {
		s.Display = override.Display
	}
	if override.Visibility != VisibilityNone {
		s.Visibility = override.Visibility
	}
	if override.Overflow != OverflowNone {
		s.Overflow = override.Overflow
	}
	if override.Opacity != nil {
		s.Opacity = override.Opacity
	}
	if override.Cursor != CursorNone {
		s.Cursor = override.Cursor
	}
	if override.MinWidth != nil {
		s.MinWidth = override.MinWidth
	}
	if override.MaxWidth != nil {
		s.MaxWidth = override.MaxWidth
	}
	if override.MinHeight != nil {
		s.MinHeight = override.MinHeight
	}
	if override.MaxHeight != nil {
		s.MaxHeight = override.MaxHeight
	}
	if override.ContentAlign != ContentAlignNone {
		s.ContentAlign = override.ContentAlign
	}
	if override.Dock != DockNone {
		s.Dock = override.Dock
	}
	if override.OffsetX != nil {
		s.OffsetX = override.OffsetX
	}
	if override.OffsetY != nil {
		s.OffsetY = override.OffsetY
	}
	if override.ZIndex != nil {
		s.ZIndex = override.ZIndex
	}
	if len(override.Transitions) > 0 {
		s.Transitions = override.Transitions
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
	// Inheritable layout properties
	if s.TextAlign == TextAlignNone {
		s.TextAlign = parent.TextAlign
	}
	if s.TextDecoration == TextDecorationNone {
		s.TextDecoration = parent.TextDecoration
	}
	if s.WhiteSpace == WhiteSpaceNone {
		s.WhiteSpace = parent.WhiteSpace
	}
	if s.Visibility == VisibilityNone {
		s.Visibility = parent.Visibility
	}
	if s.Cursor == CursorNone {
		s.Cursor = parent.Cursor
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
