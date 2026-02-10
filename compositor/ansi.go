package compositor

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ANSI escape sequences.
const (
	ANSIEscape       = "\x1b["
	ANSIClearScreen  = "\x1b[2J"
	ANSIClearLine    = "\x1b[2K"
	ANSICursorHome   = "\x1b[H"
	ANSICursorHide   = "\x1b[?25l"
	ANSICursorShow   = "\x1b[?25h"
	ANSIReset        = "\x1b[0m"
	ANSISaveCursor   = "\x1b[s"
	ANSIRestoreCursor = "\x1b[u"
	ANSIAltScreen    = "\x1b[?1049h"
	ANSIMainScreen   = "\x1b[?1049l"
	ANSISyncStart    = "\x1b[?2026h"
	ANSISyncEnd      = "\x1b[?2026l"
)

// CursorShapeSeq returns the DECSCUSR sequence for the given cursor shape.
// Ps: 0=default, 2=block, 4=underline, 6=beam.
func CursorShapeSeq(ps int) string {
	return fmt.Sprintf("\x1b[%d q", ps)
}

// CursorTo returns ANSI sequence to move cursor to (x, y).
// Coordinates are 0-indexed, but ANSI uses 1-indexed.
func CursorTo(x, y int) string {
	return fmt.Sprintf("\x1b[%d;%dH", y+1, x+1)
}

// CursorUp moves cursor up n lines.
func CursorUp(n int) string {
	if n <= 0 {
		return ""
	}
	return fmt.Sprintf("\x1b[%dA", n)
}

// CursorDown moves cursor down n lines.
func CursorDown(n int) string {
	if n <= 0 {
		return ""
	}
	return fmt.Sprintf("\x1b[%dB", n)
}

// CursorForward moves cursor right n columns.
func CursorForward(n int) string {
	if n <= 0 {
		return ""
	}
	return fmt.Sprintf("\x1b[%dC", n)
}

// CursorBack moves cursor left n columns.
func CursorBack(n int) string {
	if n <= 0 {
		return ""
	}
	return fmt.Sprintf("\x1b[%dD", n)
}

// StyleToANSI converts a Style to ANSI escape sequence.
func StyleToANSI(s Style) string {
	var b strings.Builder
	b.Grow(32) // typical ANSI sequence length

	b.WriteString(ANSIEscape)
	b.WriteByte('0')

	// Attributes
	if s.Bold {
		b.WriteString(";1")
	}
	if s.Dim {
		b.WriteString(";2")
	}
	if s.Italic {
		b.WriteString(";3")
	}
	if s.Underline {
		b.WriteString(";4")
	}
	if s.Blink {
		b.WriteString(";5")
	}
	if s.Reverse {
		b.WriteString(";7")
	}
	if s.Strikethrough {
		b.WriteString(";9")
	}

	// Foreground and background colors
	writeColorToBuilder(&b, s.FG, true)
	writeColorToBuilder(&b, s.BG, false)

	b.WriteByte('m')
	return b.String()
}

// writeColorToBuilder writes a Color to ANSI SGR parameters directly to the builder.
func writeColorToBuilder(b *strings.Builder, c Color, fg bool) {
	switch c.Mode {
	case ColorModeNone, ColorModeDefault:
		// Use default color (39 for FG, 49 for BG)
		if fg {
			b.WriteString(";39")
		} else {
			b.WriteString(";49")
		}

	case ColorMode16:
		// Basic 16 colors: 30-37 for FG (normal), 90-97 for FG (bright)
		// 40-47 for BG (normal), 100-107 for BG (bright)
		idx := c.Value
		if fg {
			if idx < 8 {
				b.WriteByte(';')
				b.WriteString(strconv.Itoa(30 + int(idx)))
			} else {
				b.WriteByte(';')
				b.WriteString(strconv.Itoa(90 + int(idx) - 8))
			}
		} else {
			if idx < 8 {
				b.WriteByte(';')
				b.WriteString(strconv.Itoa(40 + int(idx)))
			} else {
				b.WriteByte(';')
				b.WriteString(strconv.Itoa(100 + int(idx) - 8))
			}
		}

	case ColorMode256:
		// 256-color: 38;5;n for FG, 48;5;n for BG
		if fg {
			b.WriteString(";38;5;")
		} else {
			b.WriteString(";48;5;")
		}
		b.WriteString(strconv.Itoa(int(c.Value)))

	case ColorModeRGB:
		// True color: 38;2;r;g;b for FG, 48;2;r;g;b for BG
		r := (c.Value >> 16) & 0xFF
		g := (c.Value >> 8) & 0xFF
		bl := c.Value & 0xFF
		if fg {
			b.WriteString(";38;2;")
		} else {
			b.WriteString(";48;2;")
		}
		b.WriteString(strconv.Itoa(int(r)))
		b.WriteByte(';')
		b.WriteString(strconv.Itoa(int(g)))
		b.WriteByte(';')
		b.WriteString(strconv.Itoa(int(bl)))
	}
}

// StyleDelta returns ANSI codes to change from 'from' style to 'to' style.
// This is more efficient than always resetting and setting new style.
func StyleDelta(from, to Style) string {
	if from.Equal(to) {
		return ""
	}

	// For simplicity, we always do a full reset and set.
	// A more optimized version could compute minimal changes.
	return StyleToANSI(to)
}

// ANSIWriter helps build ANSI output efficiently.
type ANSIWriter struct {
	buf        strings.Builder
	lastStyle  Style
	styleSet   bool
	lastX      int
	lastY      int
	posSet     bool
	styleCache map[Style]string // Caches StyleToANSI results within a frame.
}

// NewANSIWriter creates a new ANSI writer.
func NewANSIWriter() *ANSIWriter {
	return &ANSIWriter{
		lastX: -1,
		lastY: -1,
	}
}

// MoveTo positions cursor, optimizing for sequential writes.
func (w *ANSIWriter) MoveTo(x, y int) {
	if w.posSet && w.lastY == y && w.lastX == x {
		// Cursor is already at the right position after last write
		return
	}

	if w.posSet && w.lastY == y {
		// Same line, use relative movement
		delta := x - w.lastX
		if delta > 0 && delta < 5 {
			w.buf.WriteString(CursorForward(delta))
			w.lastX = x
			return
		}
	}

	// Full position
	w.buf.WriteString(CursorTo(x, y))
	w.lastX = x
	w.lastY = y
	w.posSet = true
}

// SetStyle changes the current style.
func (w *ANSIWriter) SetStyle(s Style) {
	if w.styleSet && w.lastStyle.Equal(s) {
		return
	}
	// Look up cached ANSI string to avoid repeated StyleToANSI allocations.
	ansi, ok := w.styleCache[s]
	if !ok {
		ansi = StyleToANSI(s)
		if w.styleCache == nil {
			w.styleCache = make(map[Style]string, 16)
		}
		w.styleCache[s] = ansi
	}
	w.buf.WriteString(ansi)
	w.lastStyle = s
	w.styleSet = true
}

// WriteRune writes a single rune.
func (w *ANSIWriter) WriteRune(r rune) {
	w.buf.WriteRune(r)
	w.lastX++ // Advance cursor position
}

// WriteString writes a string.
func (w *ANSIWriter) WriteString(s string) {
	w.buf.WriteString(s)
	w.lastX += utf8.RuneCountInString(s)
}

// Reset resets the writer for reuse, clearing the buffer and state while keeping capacity.
func (w *ANSIWriter) Reset() {
	w.buf.Reset()
	w.lastStyle = Style{}
	w.styleSet = false
	w.lastX = -1
	w.lastY = -1
	w.posSet = false
	// Clear the style cache entries but keep the map allocated.
	for k := range w.styleCache {
		delete(w.styleCache, k)
	}
}

// ResetStyle adds a style reset to the buffer.
func (w *ANSIWriter) ResetStyle() {
	w.buf.WriteString(ANSIReset)
	w.styleSet = false
}

// ShowCursor adds cursor show sequence.
func (w *ANSIWriter) ShowCursor() {
	w.buf.WriteString(ANSICursorShow)
}

// HideCursor adds cursor hide sequence.
func (w *ANSIWriter) HideCursor() {
	w.buf.WriteString(ANSICursorHide)
}

// String returns the accumulated output.
func (w *ANSIWriter) String() string {
	return w.buf.String()
}

// Len returns current buffer length.
func (w *ANSIWriter) Len() int {
	return w.buf.Len()
}

// Grow pre-allocates buffer capacity.
func (w *ANSIWriter) Grow(n int) {
	w.buf.Grow(n)
}

// Hyperlink wraps text in an OSC-8 hyperlink sequence.
// Terminals that don't support OSC-8 will ignore the escape sequences
// and display the text normally.
func Hyperlink(url, text string) string {
	return "\x1b]8;;" + url + "\x1b\\" + text + "\x1b]8;;\x1b\\"
}
