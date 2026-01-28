// Package widgets provides reusable widgets for terminal UIs.
package widgets

import (
	"strings"

	"github.com/mattn/go-runewidth"

	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/style"
)

// Base provides common functionality for widgets.
// Embed this in widget structs to get default implementations.
type Base struct {
	accessibility.Base
	outerBounds  runtime.Rect
	bounds       runtime.Rect
	layoutStyle  style.Style
	layoutMetrics layoutMetrics
	focused      bool
	needsRender  bool
	id           string
	classes      []string
}

// Layout stores the assigned bounds.
func (b *Base) Layout(bounds runtime.Rect) {
	if b == nil {
		return
	}
	b.outerBounds = bounds
	metrics := b.layoutMetrics
	marginTop, marginRight, marginBottom, marginLeft := metrics.marginInsets()
	inner := bounds.Inset(marginTop, marginRight, marginBottom, marginLeft)
	if b.bounds != inner {
		b.bounds = inner
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

// ContentBounds returns the widget's content bounds.
func (b *Base) ContentBounds() runtime.Rect {
	if b == nil {
		return runtime.Rect{}
	}
	metrics := b.layoutMetrics
	top, right, bottom, left := metrics.contentInsets()
	return b.bounds.Inset(top, right, bottom, left)
}

// ApplyStyle stores the resolved style for layout.
func (b *Base) ApplyStyle(s style.Style) {
	if b == nil {
		return
	}
	b.layoutStyle = s
	b.layoutMetrics = layoutMetricsFromStyle(s)
}

// LayoutStyle returns the resolved style used for layout.
func (b *Base) LayoutStyle() style.Style {
	if b == nil {
		return style.Style{}
	}
	return b.layoutStyle
}

// ID returns the optional explicit widget identifier.
func (b *Base) ID() string {
	if b == nil {
		return ""
	}
	return b.id
}

// Key returns the stable widget identity (defaults to ID).
func (b *Base) Key() string {
	if b == nil {
		return ""
	}
	return b.id
}

// StyleID returns the style selector ID.
func (b *Base) StyleID() string {
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

// SetClasses replaces the widget classes.
func (b *Base) SetClasses(classes ...string) {
	if b == nil {
		return
	}
	b.classes = normalizeClasses(classes)
}

// AddClass adds a class if it does not already exist.
func (b *Base) AddClass(class string) {
	if b == nil {
		return
	}
	name := strings.TrimSpace(class)
	if name == "" {
		return
	}
	for _, existing := range b.classes {
		if existing == name {
			return
		}
	}
	b.classes = append(b.classes, name)
}

// AddClasses adds multiple classes.
func (b *Base) AddClasses(classes ...string) {
	if b == nil {
		return
	}
	for _, class := range classes {
		b.AddClass(class)
	}
}

// StyleClasses returns the style selector classes.
func (b *Base) StyleClasses() []string {
	if b == nil {
		return nil
	}
	return b.classes
}

// StyleState returns the default widget style state.
func (b *Base) StyleState() style.WidgetState {
	if b == nil {
		return style.WidgetState{}
	}
	return style.WidgetState{
		Focused:  b.focused,
		Disabled: b.State.Disabled,
	}
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

func normalizeClasses(classes []string) []string {
	if len(classes) == 0 {
		return nil
	}
	out := make([]string, 0, len(classes))
	for _, class := range classes {
		name := strings.TrimSpace(class)
		if name == "" {
			continue
		}
		duplicate := false
		for _, existing := range out {
			if existing == name {
				duplicate = true
				break
			}
		}
		if !duplicate {
			out = append(out, name)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
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

func textWidth(s string) int {
	return runewidth.StringWidth(s)
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

// clipString truncates a string to fit within maxWidth without ellipsis.
func clipString(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	return runewidth.Truncate(s, maxWidth, "")
}

// clipStringRight keeps the rightmost portion of the string within maxWidth.
func clipStringRight(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if textWidth(s) <= maxWidth {
		return s
	}
	runes := []rune(s)
	width := 0
	start := len(runes)
	for start > 0 {
		w := runewidth.RuneWidth(runes[start-1])
		if w < 0 {
			w = 0
		}
		if width+w > maxWidth {
			break
		}
		width += w
		start--
	}
	return string(runes[start:])
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
	if width <= 0 {
		return ""
	}
	if textWidth(s) >= width {
		return runewidth.Truncate(s, width, "")
	}
	pad := width - textWidth(s)
	left := pad / 2
	right := pad - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

func resolveBaseStyle(ctx runtime.RenderContext, widget runtime.Widget, fallback backend.Style, fallbackSet bool) backend.Style {
	resolved := ctx.ResolveStyle(widget)
	if resolved.IsZero() {
		return fallback
	}
	final := resolved
	if fallbackSet {
		final = final.Merge(style.FromBackend(fallback))
	}
	return final.ToBackend()
}

func mergeBackendStyles(base backend.Style, override backend.Style) backend.Style {
	final := style.FromBackend(base).Merge(style.FromBackend(override))
	return final.ToBackend()
}
