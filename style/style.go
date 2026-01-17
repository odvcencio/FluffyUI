package style

import (
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/compositor"
)

// ToBackend converts a compositor.Style to backend.Style.
func ToBackend(cs compositor.Style) backend.Style {
	style := backend.DefaultStyle()

	if cs.FG.Mode == compositor.ColorModeRGB {
		r := uint8((cs.FG.Value >> 16) & 0xFF)
		g := uint8((cs.FG.Value >> 8) & 0xFF)
		b := uint8(cs.FG.Value & 0xFF)
		style = style.Foreground(backend.ColorRGB(r, g, b))
	} else if cs.FG.Mode != compositor.ColorModeDefault && cs.FG.Mode != compositor.ColorModeNone {
		style = style.Foreground(backend.Color(cs.FG.Value & 0xFF))
	}

	if cs.BG.Mode == compositor.ColorModeRGB {
		r := uint8((cs.BG.Value >> 16) & 0xFF)
		g := uint8((cs.BG.Value >> 8) & 0xFF)
		b := uint8(cs.BG.Value & 0xFF)
		style = style.Background(backend.ColorRGB(r, g, b))
	} else if cs.BG.Mode != compositor.ColorModeDefault && cs.BG.Mode != compositor.ColorModeNone {
		style = style.Background(backend.Color(cs.BG.Value & 0xFF))
	}

	if cs.Bold {
		style = style.Bold(true)
	}
	if cs.Italic {
		style = style.Italic(true)
	}
	if cs.Underline {
		style = style.Underline(true)
	}
	if cs.Dim {
		style = style.Dim(true)
	}
	if cs.Blink {
		style = style.Blink(true)
	}
	if cs.Reverse {
		style = style.Reverse(true)
	}
	if cs.Strikethrough {
		style = style.StrikeThrough(true)
	}

	return style
}

// ToCompositor converts a backend.Style to compositor.Style.
func ToCompositor(bs backend.Style) compositor.Style {
	fg, bg, attrs := bs.Decompose()
	style := compositor.DefaultStyle()
	style.FG = colorToCompositor(fg)
	style.BG = colorToCompositor(bg)
	style.Bold = attrs&backend.AttrBold != 0
	style.Italic = attrs&backend.AttrItalic != 0
	style.Dim = attrs&backend.AttrDim != 0
	style.Underline = attrs&backend.AttrUnderline != 0
	style.Blink = attrs&backend.AttrBlink != 0
	style.Reverse = attrs&backend.AttrReverse != 0
	style.Strikethrough = attrs&backend.AttrStrikeThrough != 0
	return style
}

func colorToCompositor(c backend.Color) compositor.Color {
	if c == backend.ColorDefault || int32(c) < 0 {
		return compositor.ColorDefault
	}
	if c.IsRGB() {
		r, g, b := c.RGB()
		return compositor.RGB(r, g, b)
	}
	value := int(c)
	if value < 0 {
		return compositor.ColorDefault
	}
	if value <= 15 {
		return compositor.Color{Mode: compositor.ColorMode16, Value: uint32(value)}
	}
	if value <= 255 {
		return compositor.Color256(uint8(value))
	}
	return compositor.ColorDefault
}
