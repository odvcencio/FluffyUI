package widgets

import (
	"strings"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/i18n"
	"github.com/odvcencio/fluffyui/runtime"
	uistyle "github.com/odvcencio/fluffyui/style"
)

// Text is a simple text display widget.
type Text struct {
	Base
	text         string
	style        backend.Style
	lines        []string // Cached line splits
	a11yLabel    string
	styleSet     bool
	direction    i18n.Direction
	directionSet bool // true if direction was explicitly set via SetDirection
	services     runtime.Services
}

// TextOption configures a Text widget.
type TextOption = Option[Text]

// WithTextStyle sets the text style.
func WithTextStyle(style backend.Style) TextOption {
	return func(t *Text) {
		if t == nil {
			return
		}
		t.SetStyle(style)
	}
}

// WithTextDirection sets an explicit text direction (DirectionLTR or DirectionRTL).
// When set, auto-detection is bypassed for rendering.
func WithTextDirection(dir i18n.Direction) TextOption {
	return func(t *Text) {
		if t == nil {
			return
		}
		t.SetDirection(dir)
	}
}

// WithTextA11yLabel sets an accessibility label override.
func WithTextA11yLabel(label string) TextOption {
	return func(t *Text) {
		if t == nil {
			return
		}
		t.SetA11yLabel(label)
	}
}

// NewText creates a new text widget.
func NewText(text string, opts ...TextOption) *Text {
	t := &Text{
		text:  text,
		style: backend.DefaultStyle(),
		lines: strings.Split(text, "\n"),
	}
	t.Base.Role = accessibility.RoleText
	t.Base.Label = text
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(t)
	}
	t.syncA11y()
	return t
}

// SetText updates the displayed text.
func (t *Text) SetText(text string) {
	t.text = text
	t.lines = strings.Split(text, "\n")
	t.syncA11y()
}

// SetA11yLabel overrides the accessibility label without changing visible text.
func (t *Text) SetA11yLabel(label string) {
	t.a11yLabel = label
	t.syncA11y()
}

// Text returns the current text.
func (t *Text) Text() string {
	return t.text
}

// SetStyle sets the text style.
func (t *Text) SetStyle(style backend.Style) {
	t.style = style
	t.styleSet = true
}

// SetDirection sets an explicit text direction override.
// Pass DirectionRTL for right-to-left or DirectionLTR for left-to-right.
// When set, auto-detection is bypassed.
func (t *Text) SetDirection(dir i18n.Direction) {
	t.direction = dir
	t.directionSet = true
}

// Direction returns the effective text direction.
// If a direction was explicitly set, it returns that.
// Otherwise, it auto-detects from the text content.
func (t *Text) Direction() i18n.Direction {
	if t.directionSet {
		return t.direction
	}
	return i18n.DetectDirection(t.text)
}

// Deprecated: prefer WithTextStyle during construction or SetStyle for mutation.
func (t *Text) WithStyle(style backend.Style) *Text {
	t.style = style
	t.styleSet = true
	return t
}

// StyleType returns the selector type name.
func (t *Text) StyleType() string {
	return "Text"
}

// Measure returns the size needed to display the text.
func (t *Text) Measure(constraints runtime.Constraints) runtime.Size {
	return t.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		// Calculate width: longest line
		maxWidth := 0
		for _, line := range t.lines {
			lineWidth := textWidth(line)
			if lineWidth > maxWidth {
				maxWidth = lineWidth
			}
		}

		// Height: number of lines
		height := len(t.lines)
		if height == 0 {
			height = 1
		}

		return contentConstraints.Constrain(runtime.Size{
			Width:  maxWidth,
			Height: height,
		})
	})
}

// Render draws the text.
func (t *Text) Render(ctx runtime.RenderContext) {
	if t == nil {
		return
	}
	bounds := t.ContentBounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	t.syncA11y()

	style := t.style
	resolved := ctx.ResolveStyle(t)
	if !resolved.IsZero() {
		final := resolved
		if t.styleSet {
			final = final.Merge(uistyle.FromBackend(t.style))
		}
		style = final.ToBackend()
	}

	dir := t.Direction()
	for i, line := range t.lines {
		if i >= bounds.Height {
			break
		}
		y := bounds.Y + i
		displayLine := line
		// Apply bidi reordering for RTL text
		if dir == i18n.DirectionRTL {
			displayLine = i18n.BidiReorder(displayLine, i18n.DirectionRTL)
		}
		if textWidth(displayLine) > bounds.Width {
			displayLine = clipString(displayLine, bounds.Width)
		}
		// Right-align RTL text
		x := bounds.X
		if dir == i18n.DirectionRTL {
			lineW := textWidth(displayLine)
			if lineW < bounds.Width {
				x = bounds.X + bounds.Width - lineW
			}
		}
		ctx.Buffer.SetString(x, y, displayLine, style)
	}
}

func (t *Text) syncA11y() {
	if t == nil {
		return
	}
	if t.Base.Role == "" {
		t.Base.Role = accessibility.RoleText
	}
	override := strings.TrimSpace(t.a11yLabel)
	if override != "" {
		t.Base.Label = override
		value := strings.TrimSpace(t.text)
		if value != "" {
			t.Base.Value = &accessibility.ValueInfo{Text: value}
		} else {
			t.Base.Value = nil
		}
		return
	}
	label := strings.TrimSpace(t.text)
	if label == "" {
		label = "Text"
	}
	t.Base.Label = label
	t.Base.Value = nil
}

// Label is a single-line text widget often used for headers/labels.
type Label struct {
	Base
	text         string
	style        backend.Style
	alignment    Alignment
	a11yLabel    string
	styleSet     bool
	direction    i18n.Direction
	directionSet bool // true if direction was explicitly set via SetDirection
	services     runtime.Services
}

// LabelOption configures a Label widget.
type LabelOption = Option[Label]

// WithLabelStyle sets the label style.
func WithLabelStyle(style backend.Style) LabelOption {
	return func(l *Label) {
		if l == nil {
			return
		}
		l.SetStyle(style)
	}
}

// WithLabelAlignment sets the label alignment.
func WithLabelAlignment(align Alignment) LabelOption {
	return func(l *Label) {
		if l == nil {
			return
		}
		l.SetAlignment(align)
	}
}

// WithLabelDirection sets an explicit text direction (DirectionLTR or DirectionRTL).
// When set, auto-detection is bypassed for rendering.
func WithLabelDirection(dir i18n.Direction) LabelOption {
	return func(l *Label) {
		if l == nil {
			return
		}
		l.SetDirection(dir)
	}
}

// WithLabelA11yLabel sets an accessibility label override.
func WithLabelA11yLabel(label string) LabelOption {
	return func(l *Label) {
		if l == nil {
			return
		}
		l.SetA11yLabel(label)
	}
}

// Alignment specifies text alignment.
type Alignment int

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
)

// NewLabel creates a new label widget.
func NewLabel(text string, opts ...LabelOption) *Label {
	l := &Label{
		text:      text,
		style:     backend.DefaultStyle(),
		alignment: AlignLeft,
	}
	l.Base.Role = accessibility.RoleText
	l.Base.Label = text
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(l)
	}
	l.syncA11y()
	return l
}

// SetText updates the label text.
func (l *Label) SetText(text string) {
	l.text = text
	l.syncA11y()
}

// SetA11yLabel overrides the accessibility label without changing visible text.
func (l *Label) SetA11yLabel(label string) {
	l.a11yLabel = label
	l.syncA11y()
}

// SetStyle sets the label style.
func (l *Label) SetStyle(style backend.Style) {
	l.style = style
	l.styleSet = true
}

// SetAlignment sets text alignment.
func (l *Label) SetAlignment(align Alignment) {
	l.alignment = align
}

// SetDirection sets an explicit text direction override.
// Pass DirectionRTL for right-to-left or DirectionLTR for left-to-right.
// When set, auto-detection is bypassed.
func (l *Label) SetDirection(dir i18n.Direction) {
	l.direction = dir
	l.directionSet = true
}

// Direction returns the effective text direction.
// If a direction was explicitly set, it returns that.
// Otherwise, it auto-detects from the text content.
func (l *Label) Direction() i18n.Direction {
	if l.directionSet {
		return l.direction
	}
	return i18n.DetectDirection(l.text)
}

// Deprecated: prefer WithLabelStyle during construction or SetStyle for mutation.
func (l *Label) WithStyle(style backend.Style) *Label {
	l.style = style
	l.styleSet = true
	return l
}

// StyleType returns the selector type name.
func (l *Label) StyleType() string {
	return "Label"
}

// Deprecated: prefer WithLabelAlignment during construction or SetAlignment for mutation.
func (l *Label) WithAlignment(align Alignment) *Label {
	l.alignment = align
	return l
}

// Measure returns the size needed for the label.
func (l *Label) Measure(constraints runtime.Constraints) runtime.Size {
	return l.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		return contentConstraints.Constrain(runtime.Size{
			Width:  textWidth(l.text),
			Height: 1,
		})
	})
}

// Render draws the label.
func (l *Label) Render(ctx runtime.RenderContext) {
	if l == nil {
		return
	}
	bounds := l.ContentBounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	l.syncA11y()

	text := l.text
	dir := l.Direction()

	// Apply bidi reordering for RTL text
	if dir == i18n.DirectionRTL {
		text = i18n.BidiReorder(text, i18n.DirectionRTL)
	}
	if textWidth(text) > bounds.Width {
		text = truncateString(text, bounds.Width)
	}

	// Calculate X position based on alignment.
	// RTL text defaults to right-alignment unless explicitly overridden.
	x := bounds.X
	textW := textWidth(text)
	alignment := l.alignment
	if dir == i18n.DirectionRTL && alignment == AlignLeft {
		alignment = AlignRight
	}
	switch alignment {
	case AlignCenter:
		x = bounds.X + (bounds.Width-textW)/2
	case AlignRight:
		x = bounds.X + bounds.Width - textW
	}

	style := l.style
	resolved := ctx.ResolveStyle(l)
	if !resolved.IsZero() {
		final := resolved
		if l.styleSet {
			final = final.Merge(uistyle.FromBackend(l.style))
		}
		style = final.ToBackend()
	}
	ctx.Buffer.SetString(x, bounds.Y, text, style)
}

func (l *Label) syncA11y() {
	if l == nil {
		return
	}
	if l.Base.Role == "" {
		l.Base.Role = accessibility.RoleText
	}
	override := strings.TrimSpace(l.a11yLabel)
	if override != "" {
		l.Base.Label = override
		value := strings.TrimSpace(l.text)
		if value != "" {
			l.Base.Value = &accessibility.ValueInfo{Text: value}
		} else {
			l.Base.Value = nil
		}
		return
	}
	label := strings.TrimSpace(l.text)
	if label == "" {
		label = "Label"
	}
	l.Base.Label = label
	l.Base.Value = nil
}

// Bind attaches app services.
func (t *Text) Bind(services runtime.Services) {
	if t == nil {
		return
	}
	t.services = services
}

// Unbind releases app services.
func (t *Text) Unbind() {
	if t == nil {
		return
	}
	t.services = runtime.Services{}
}

// Bind attaches app services.
func (l *Label) Bind(services runtime.Services) {
	if l == nil {
		return
	}
	l.services = services
}

// Unbind releases app services.
func (l *Label) Unbind() {
	if l == nil {
		return
	}
	l.services = runtime.Services{}
}

var _ runtime.Widget = (*Text)(nil)
var _ runtime.Bindable = (*Text)(nil)
var _ runtime.Unbindable = (*Text)(nil)
var _ runtime.Widget = (*Label)(nil)
var _ runtime.Bindable = (*Label)(nil)
var _ runtime.Unbindable = (*Label)(nil)
