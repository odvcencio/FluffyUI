// Package widgets provides reusable widgets for terminal UIs.
package widgets

import (
	"strings"

	"github.com/mattn/go-runewidth"

	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
)

// Base provides common functionality for widgets.
// Embed this in widget structs to get default implementations.
type Base struct {
	accessibility.Base
	bounds      runtime.Rect
	focused     bool
	needsRender bool
	id          string
}

// Layout stores the assigned bounds.
func (b *Base) Layout(bounds runtime.Rect) {
	if b == nil {
		return
	}
	if b.bounds != bounds {
		b.bounds = bounds
		b.needsRender = true
	}
}

// Bounds returns the widget's assigned bounds.
func (b *Base) Bounds() runtime.Rect {
	if b == nil {
		return runtime.Rect{}
	}
	return b.bounds
}

// ID returns the optional explicit widget identifier.
func (b *Base) ID() string {
	if b == nil {
		return ""
	}
	return b.id
}

// SetID assigns an explicit widget identifier.
func (b *Base) SetID(id string) {
	if b == nil {
		return
	}
	b.id = strings.TrimSpace(id)
}

// HandleMessage returns Unhandled by default.
func (b *Base) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}

// CanFocus returns false by default.
func (b *Base) CanFocus() bool {
	return false
}

// Focus marks the widget as focused.
func (b *Base) Focus() {
	if b == nil {
		return
	}
	b.focused = true
}

// Blur marks the widget as unfocused.
func (b *Base) Blur() {
	if b == nil {
		return
	}
	b.focused = false
}

// IsFocused returns whether the widget is focused.
func (b *Base) IsFocused() bool {
	if b == nil {
		return false
	}
	return b.focused
}

// Invalidate marks the widget as needing a render pass.
func (b *Base) Invalidate() {
	if b == nil {
		return
	}
	b.needsRender = true
}

// NeedsRender reports whether the widget needs to re-render.
func (b *Base) NeedsRender() bool {
	if b == nil {
		return false
	}
	return b.needsRender
}

// ClearInvalidation clears the render-needed flag.
func (b *Base) ClearInvalidation() {
	if b == nil {
		return
	}
	b.needsRender = false
}

// FocusableBase extends Base for focusable widgets.
type FocusableBase struct {
	Base
}

// CanFocus returns true for focusable widgets.
func (f *FocusableBase) CanFocus() bool {
	return true
}

// drawText is a helper to draw text with word wrapping.
func drawText(buf *runtime.Buffer, bounds runtime.Rect, text string, style backend.Style) {
	x := bounds.X
	y := bounds.Y
	maxX := bounds.X + bounds.Width
	maxY := bounds.Y + bounds.Height

	for _, r := range text {
		if r == '\n' {
			x = bounds.X
			y++
			if y >= maxY {
				break
			}
			continue
		}

		if x >= maxX {
			x = bounds.X
			y++
			if y >= maxY {
				break
			}
		}

		buf.Set(x, y, r, style)
		x++
	}
}

// fillRect fills a rectangle with a character.
func fillRect(buf *runtime.Buffer, bounds runtime.Rect, ch rune, style backend.Style) {
	buf.Fill(bounds, ch, style)
}

// truncateString truncates a string to fit within maxWidth.
// Adds "..." if truncated.
func truncateString(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if runewidth.StringWidth(s) <= maxWidth {
		return s
	}
	if maxWidth <= 3 {
		return runewidth.Truncate(s, maxWidth, "")
	}
	return runewidth.Truncate(s, maxWidth, "...")
}

// padRight pads a string with spaces to reach the given width.
func padRight(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if runewidth.StringWidth(s) >= width {
		return runewidth.Truncate(s, width, "")
	}
	padding := width - runewidth.StringWidth(s)
	return s + strings.Repeat(" ", padding)
}

func writePadded(buf *runtime.Buffer, x, y, width int, text string, style backend.Style) {
	if buf == nil || width <= 0 {
		return
	}
	if x < 0 {
		buf.SetString(x, y, padRight(text, width), style)
		return
	}
	text = runewidth.Truncate(text, width, "")
	buf.SetString(x, y, text, style)
	if pad := width - runewidth.StringWidth(text); pad > 0 {
		buf.Fill(runtime.Rect{X: x + runewidth.StringWidth(text), Y: y, Width: pad, Height: 1}, ' ', style)
	}
}

// centerString centers a string within the given width.
func centerString(s string, width int) string {
	if len(s) >= width {
		return s
	}
	padding := (width - len(s)) / 2
	result := make([]byte, width)
	for i := range result {
		result[i] = ' '
	}
	copy(result[padding:], s)
	return string(result)
}
