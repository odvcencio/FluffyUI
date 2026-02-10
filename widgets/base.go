// Package widgets provides reusable widgets for terminal UIs.
package widgets

import (
	"strings"

	"github.com/mattn/go-runewidth"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/style"
)

// Base provides common functionality for widgets.
// Base should be embedded in widget structs to get default implementations.
type Base struct {
	accessibility.Base
	outerBounds   runtime.Rect
	bounds        runtime.Rect
	layoutStyle   style.Style
	layoutMetrics layoutMetrics
	focused       bool
	hovered       bool
	active        bool
	needsRender   bool
	id            string
	classes       []string
	onMount       func()
	onUnmount     func()
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

// SetKey assigns the stable widget identity (alias for SetID).
func (b *Base) SetKey(key string) {
	if b == nil {
		return
	}
	b.id = strings.TrimSpace(key)
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
		Hovered:  b.hovered,
		Active:   b.active,
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

// SetHovered sets the hover state of the widget.
func (b *Base) SetHovered(hovered bool) {
	if b == nil {
		return
	}
	if b.hovered != hovered {
		b.hovered = hovered
		b.Invalidate()
	}
}

// IsHovered returns true if the widget is currently hovered.
func (b *Base) IsHovered() bool {
	if b == nil {
		return false
	}
	return b.hovered
}

// SetActive sets the active (pressed) state of the widget.
func (b *Base) SetActive(active bool) {
	if b == nil {
		return
	}
	if b.active != active {
		b.active = active
		b.Invalidate()
	}
}

// IsActive returns true if the widget is currently active (pressed).
func (b *Base) IsActive() bool {
	if b == nil {
		return false
	}
	return b.active
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

// OnMount registers a callback that fires when the widget is bound
// (mounted into a screen). If the widget type defines its own Bind method,
// call b.Base.Bind(services) to trigger the hook.
func (b *Base) OnMount(fn func()) {
	if b == nil {
		return
	}
	b.onMount = fn
}

// OnUnmount registers a callback that fires when the widget is unbound
// (removed from a screen). If the widget type defines its own Unbind method,
// call b.Base.Unbind() to trigger the hook.
func (b *Base) OnUnmount(fn func()) {
	if b == nil {
		return
	}
	b.onUnmount = fn
}

// Bind implements runtime.Bindable for Base. Widgets that define their own
// Bind will shadow this; the onMount hook is separately invoked by
// OnMountHook which BindTree calls after Bind.
func (b *Base) Bind(services runtime.Services) {
	if b == nil {
		return
	}
}

// Unbind implements runtime.Unbindable for Base. Widgets that define their
// own Unbind will shadow this; the onUnmount hook is separately invoked by
// OnUnmountHook which UnbindTree calls before Unbind.
func (b *Base) Unbind() {
	if b == nil {
		return
	}
}

// OnMountHook implements runtime.MountHook. It invokes the onMount callback
// if one has been registered via OnMount. BindTree calls this automatically
// after Bind, so hooks fire even if a widget has its own Bind method.
func (b *Base) OnMountHook() {
	if b == nil {
		return
	}
	if b.onMount != nil {
		b.onMount()
	}
}

// OnUnmountHook implements runtime.UnmountHook. It invokes the onUnmount
// callback if one has been registered via OnUnmount. UnbindTree calls this
// automatically before Unbind.
func (b *Base) OnUnmountHook() {
	if b == nil {
		return
	}
	if b.onUnmount != nil {
		b.onUnmount()
	}
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
	onFocus func()
	onBlur  func()
}

// CanFocus returns true for focusable widgets.
func (f *FocusableBase) CanFocus() bool {
	return true
}

// OnFocus registers a callback that fires when the widget gains focus.
func (f *FocusableBase) OnFocus(fn func()) {
	if f == nil {
		return
	}
	f.onFocus = fn
}

// OnBlur registers a callback that fires when the widget loses focus.
func (f *FocusableBase) OnBlur(fn func()) {
	if f == nil {
		return
	}
	f.onBlur = fn
}

// Focus marks the widget as focused and invokes the onFocus hook.
func (f *FocusableBase) Focus() {
	if f == nil {
		return
	}
	f.Base.Focus()
	if f.onFocus != nil {
		f.onFocus()
	}
}

// Blur marks the widget as unfocused and invokes the onBlur hook.
func (f *FocusableBase) Blur() {
	if f == nil {
		return
	}
	f.Base.Blur()
	if f.onBlur != nil {
		f.onBlur()
	}
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
	if textWidth(s) <= maxWidth {
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
	textW := textWidth(s)
	if textW >= width {
		return runewidth.Truncate(s, width, "")
	}
	padding := width - textW
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
	textW := textWidth(text)
	buf.SetString(x, y, text, style)
	if pad := width - textW; pad > 0 {
		buf.Fill(runtime.Rect{X: x + textW, Y: y, Width: pad, Height: 1}, ' ', style)
	}
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
