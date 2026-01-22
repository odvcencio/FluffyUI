package widgets

import (
	"strings"

	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
)

// Text is a simple text display widget.
type Text struct {
	Base
	text      string
	style     backend.Style
	lines     []string // Cached line splits
	a11yLabel string
}

// NewText creates a new text widget.
func NewText(text string) *Text {
	t := &Text{
		text:  text,
		style: backend.DefaultStyle(),
		lines: strings.Split(text, "\n"),
	}
	t.Base.Role = accessibility.RoleText
	t.Base.Label = text
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
}

// WithStyle sets the style and returns the widget for chaining.
func (t *Text) WithStyle(style backend.Style) *Text {
	t.style = style
	return t
}

// Measure returns the size needed to display the text.
func (t *Text) Measure(constraints runtime.Constraints) runtime.Size {
	// Calculate width: longest line
	maxWidth := 0
	for _, line := range t.lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	// Height: number of lines
	height := len(t.lines)
	if height == 0 {
		height = 1
	}

	return constraints.Constrain(runtime.Size{
		Width:  maxWidth,
		Height: height,
	})
}

// Render draws the text.
func (t *Text) Render(ctx runtime.RenderContext) {
	bounds := t.bounds
	if bounds.Width == 0 || bounds.Height == 0 {
		return
	}
	t.syncA11y()

	style := t.style

	for i, line := range t.lines {
		if i >= bounds.Height {
			break
		}
		y := bounds.Y + i
		displayLine := line
		if len(displayLine) > bounds.Width {
			displayLine = displayLine[:bounds.Width]
		}
		ctx.Buffer.SetString(bounds.X, y, displayLine, style)
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
	text      string
	style     backend.Style
	alignment Alignment
	a11yLabel string
}

// Alignment specifies text alignment.
type Alignment int

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
)

// NewLabel creates a new label widget.
func NewLabel(text string) *Label {
	l := &Label{
		text:      text,
		style:     backend.DefaultStyle(),
		alignment: AlignLeft,
	}
	l.Base.Role = accessibility.RoleText
	l.Base.Label = text
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
}

// SetAlignment sets text alignment.
func (l *Label) SetAlignment(align Alignment) {
	l.alignment = align
}

// WithStyle sets the style and returns for chaining.
func (l *Label) WithStyle(style backend.Style) *Label {
	l.style = style
	return l
}

// WithAlignment sets alignment and returns for chaining.
func (l *Label) WithAlignment(align Alignment) *Label {
	l.alignment = align
	return l
}

// Measure returns the size needed for the label.
func (l *Label) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.Constrain(runtime.Size{
		Width:  len(l.text),
		Height: 1,
	})
}

// Render draws the label.
func (l *Label) Render(ctx runtime.RenderContext) {
	bounds := l.bounds
	if bounds.Width == 0 || bounds.Height == 0 {
		return
	}
	l.syncA11y()

	text := l.text
	if len(text) > bounds.Width {
		text = truncateString(text, bounds.Width)
	}

	// Calculate X position based on alignment
	x := bounds.X
	switch l.alignment {
	case AlignCenter:
		x = bounds.X + (bounds.Width-len(text))/2
	case AlignRight:
		x = bounds.X + bounds.Width - len(text)
	}

	ctx.Buffer.SetString(x, bounds.Y, text, l.style)
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
