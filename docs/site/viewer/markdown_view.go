package viewer

import (
	"strings"

	"github.com/mattn/go-runewidth"

	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/compositor"
	"github.com/odvcencio/fluffy-ui/markdown"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/scroll"
	"github.com/odvcencio/fluffy-ui/terminal"
	"github.com/odvcencio/fluffy-ui/widgets"
)

// MarkdownView renders styled markdown lines with internal scrolling.
type MarkdownView struct {
	widgets.FocusableBase
	lines         []markdown.StyledLine
	wrapped       []renderLine
	width         int
	offset        int
	label         string
	style         backend.Style
	showBar       bool
	scrollbar     scroll.Scrollbar
	contentSize   runtime.Size
	anchorOffsets map[string]int
	pendingAnchor string
}

// NewMarkdownView creates a new markdown view.
func NewMarkdownView(lines []markdown.StyledLine) *MarkdownView {
	view := &MarkdownView{
		lines:   lines,
		label:   "Documentation",
		style:   backend.DefaultStyle(),
		showBar: true,
		scrollbar: scroll.Scrollbar{
			Orientation:  scroll.Vertical,
			Track:        backend.DefaultStyle(),
			Thumb:        backend.DefaultStyle().Reverse(true),
			MinThumbSize: 1,
			Chars:        scroll.DefaultScrollbarChars(),
		},
	}
	view.Base.Role = accessibility.RoleText
	view.syncA11y()
	return view
}

// SetLines updates the rendered markdown lines.
func (m *MarkdownView) SetLines(lines []markdown.StyledLine) {
	if m == nil {
		return
	}
	m.lines = lines
	m.wrapped = nil
	m.anchorOffsets = nil
	m.pendingAnchor = ""
	if m.width > 0 {
		m.wrap(m.width)
	}
	m.offset = 0
	m.Invalidate()
}

// SetLabel sets the accessibility label.
func (m *MarkdownView) SetLabel(label string) {
	if m == nil {
		return
	}
	m.label = label
	m.syncA11y()
}

// SetShowScrollbar toggles the vertical scrollbar.
func (m *MarkdownView) SetShowScrollbar(show bool) {
	if m == nil {
		return
	}
	m.showBar = show
}

// SetStyle updates the base style for the view.
func (m *MarkdownView) SetStyle(style backend.Style) {
	if m == nil {
		return
	}
	m.style = style
}

// Measure returns the required size for the content.
func (m *MarkdownView) Measure(constraints runtime.Constraints) runtime.Size {
	width := constraints.MaxWidth
	if width <= 0 {
		width = constraints.MinWidth
	}
	if width <= 0 {
		width = 1
	}
	m.wrap(width)
	return constraints.Constrain(runtime.Size{Width: width, Height: len(m.wrapped)})
}

// Layout stores the bounds and updates wrapping.
func (m *MarkdownView) Layout(bounds runtime.Rect) {
	m.FocusableBase.Layout(bounds)
	content := m.ContentBounds()
	width := content.Width
	if m.showBar && m.needsScrollbar(content.Height) && width > 1 {
		width--
	}
	if width < 1 {
		width = 1
	}
	if width != m.width {
		m.wrap(width)
	}
	m.clampOffset(content.Height)
}

// Render draws the visible lines.
func (m *MarkdownView) Render(ctx runtime.RenderContext) {
	if m == nil {
		return
	}
	m.syncA11y()
	bounds := m.ContentBounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	visibleWidth := bounds.Width
	showBar := m.showBar && m.needsScrollbar(bounds.Height) && bounds.Width > 1
	if showBar {
		visibleWidth--
	}
	baseStyle := m.style
	if baseStyle == (backend.Style{}) {
		baseStyle = backend.DefaultStyle()
	}
	ctx.Buffer.Fill(bounds, ' ', baseStyle)

	for row := 0; row < bounds.Height; row++ {
		lineIndex := m.offset + row
		if lineIndex < 0 || lineIndex >= len(m.wrapped) {
			continue
		}
		line := m.wrapped[lineIndex]
		lineBounds := runtime.Rect{X: bounds.X, Y: bounds.Y + row, Width: visibleWidth, Height: 1}
		if line.BaseStyle != backend.DefaultStyle() {
			ctx.Buffer.Fill(lineBounds, ' ', line.BaseStyle)
		}
		drawRenderLine(ctx.Buffer, lineBounds, line)
	}

	if showBar {
		barBounds := runtime.Rect{
			X:      bounds.X + visibleWidth,
			Y:      bounds.Y,
			Width:  1,
			Height: bounds.Height,
		}
		drawScrollbar(ctx.Buffer, barBounds, m.scrollbar, len(m.wrapped), bounds.Height, m.offset)
	}
}

// HandleMessage handles scroll input.
func (m *MarkdownView) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if m == nil {
		return runtime.Unhandled()
	}
	switch ev := msg.(type) {
	case runtime.KeyMsg:
		if !m.IsFocused() {
			return runtime.Unhandled()
		}
		switch ev.Key {
		case terminal.KeyUp:
			m.ScrollBy(0, -1)
			return runtime.Handled()
		case terminal.KeyDown:
			m.ScrollBy(0, 1)
			return runtime.Handled()
		case terminal.KeyPageUp:
			m.PageBy(-1)
			return runtime.Handled()
		case terminal.KeyPageDown:
			m.PageBy(1)
			return runtime.Handled()
		case terminal.KeyHome:
			m.ScrollToStart()
			return runtime.Handled()
		case terminal.KeyEnd:
			m.ScrollToEnd()
			return runtime.Handled()
		}
	case runtime.MouseMsg:
		if ev.Button == runtime.MouseWheelUp {
			m.ScrollBy(0, -3)
			return runtime.Handled()
		}
		if ev.Button == runtime.MouseWheelDown {
			m.ScrollBy(0, 3)
			return runtime.Handled()
		}
	}
	return runtime.Unhandled()
}

// ScrollBy scrolls the content by delta.
func (m *MarkdownView) ScrollBy(dx, dy int) {
	if m == nil || dy == 0 {
		return
	}
	m.offset += dy
	m.clampOffset(m.ContentBounds().Height)
	m.Invalidate()
}

// ScrollTo scrolls to an absolute offset.
func (m *MarkdownView) ScrollTo(x, y int) {
	if m == nil {
		return
	}
	m.offset = y
	m.clampOffset(m.ContentBounds().Height)
	m.Invalidate()
}

// PageBy scrolls by pages.
func (m *MarkdownView) PageBy(pages int) {
	if m == nil {
		return
	}
	pageSize := m.ContentBounds().Height
	if pageSize < 1 {
		pageSize = 1
	}
	m.ScrollBy(0, pages*pageSize)
}

// ScrollToStart scrolls to the top.
func (m *MarkdownView) ScrollToStart() {
	m.ScrollTo(0, 0)
}

// ScrollToEnd scrolls to the bottom.
func (m *MarkdownView) ScrollToEnd() {
	if m == nil {
		return
	}
	height := m.ContentBounds().Height
	maxOffset := max(0, len(m.wrapped)-height)
	m.ScrollTo(0, maxOffset)
}

// ScrollToAnchor scrolls to a heading anchor if available.
func (m *MarkdownView) ScrollToAnchor(anchor string) {
	if m == nil || anchor == "" {
		return
	}
	m.pendingAnchor = anchor
	if m.anchorOffsets == nil {
		return
	}
	if offset, ok := m.anchorOffsets[anchor]; ok {
		m.pendingAnchor = ""
		m.ScrollTo(0, offset)
	}
}

func (m *MarkdownView) clampOffset(viewHeight int) {
	maxOffset := max(0, len(m.wrapped)-viewHeight)
	if m.offset < 0 {
		m.offset = 0
	}
	if m.offset > maxOffset {
		m.offset = maxOffset
	}
}

func (m *MarkdownView) needsScrollbar(viewHeight int) bool {
	return len(m.wrapped) > viewHeight
}

func (m *MarkdownView) wrap(width int) {
	if m == nil {
		return
	}
	if width < 1 {
		width = 1
	}
	if m.width == width && len(m.wrapped) > 0 {
		return
	}
	m.width = width
	m.wrapped = wrapLines(m.lines, width)
	m.contentSize = runtime.Size{Width: width, Height: len(m.wrapped)}
	m.anchorOffsets = map[string]int{}
	for i, line := range m.wrapped {
		if line.Anchor == "" {
			continue
		}
		if _, ok := m.anchorOffsets[line.Anchor]; !ok {
			m.anchorOffsets[line.Anchor] = i
		}
	}
	if m.pendingAnchor != "" {
		if offset, ok := m.anchorOffsets[m.pendingAnchor]; ok {
			m.pendingAnchor = ""
			m.ScrollTo(0, offset)
		}
	}
}

func (m *MarkdownView) syncA11y() {
	if m == nil {
		return
	}
	if m.Base.Role == "" {
		m.Base.Role = accessibility.RoleText
	}
	label := strings.TrimSpace(m.label)
	if label == "" {
		label = "Documentation"
	}
	m.Base.Label = label
}

type renderSpan struct {
	Text  string
	Style backend.Style
}

type renderLine struct {
	Spans     []renderSpan
	BlankLine bool
	BaseStyle backend.Style
	Anchor    string
}

func wrapLines(lines []markdown.StyledLine, width int) []renderLine {
	if width < 1 {
		return nil
	}
	out := make([]renderLine, 0, len(lines))
	for _, line := range lines {
		out = append(out, wrapLine(line, width)...)
	}
	return out
}

func wrapLine(line markdown.StyledLine, width int) []renderLine {
	if width < 1 {
		return nil
	}
	if line.BlankLine && len(line.Spans) == 0 && len(line.Prefix) == 0 {
		return []renderLine{{BlankLine: true}}
	}
	prefix := convertSpans(line.Prefix)
	prefixWidth := spanWidth(prefix)
	if prefixWidth > width {
		prefix = truncateSpans(prefix, width)
		prefixWidth = spanWidth(prefix)
	}
	var lines []renderLine
	current := newRenderLine(prefix, line.Anchor)
	curWidth := prefixWidth
	appendLine := func() {
		lines = append(lines, current)
		current = newRenderLine(prefix, "")
		curWidth = prefixWidth
	}
	for _, span := range convertSpans(line.Spans) {
		for _, r := range span.Text {
			if r == '\n' {
				appendLine()
				continue
			}
			rw := runewidth.RuneWidth(r)
			if rw <= 0 {
				continue
			}
			if curWidth+rw > width {
				appendLine()
			}
			appendRune(&current, r, span.Style)
			curWidth += rw
		}
	}
	if len(current.Spans) > 0 || prefixWidth > 0 || line.BlankLine {
		lines = append(lines, current)
	}
	if len(lines) == 0 {
		lines = append(lines, renderLine{BlankLine: true})
	}
	return lines
}

func newRenderLine(prefix []renderSpan, anchor string) renderLine {
	line := renderLine{Anchor: anchor}
	if len(prefix) > 0 {
		line.Spans = append(line.Spans, prefix...)
	}
	line.BaseStyle = baseStyleForLine(line.Spans)
	return line
}

func appendRune(line *renderLine, r rune, style backend.Style) {
	if line == nil {
		return
	}
	if line.BaseStyle == backend.DefaultStyle() && style.BG() != backend.ColorDefault {
		line.BaseStyle = style
	}
	if len(line.Spans) == 0 {
		line.Spans = append(line.Spans, renderSpan{Text: string(r), Style: style})
		return
	}
	last := &line.Spans[len(line.Spans)-1]
	if last.Style == style {
		last.Text += string(r)
		return
	}
	line.Spans = append(line.Spans, renderSpan{Text: string(r), Style: style})
}

func spanWidth(spans []renderSpan) int {
	width := 0
	for _, span := range spans {
		width += runewidth.StringWidth(span.Text)
	}
	return width
}

func truncateSpans(spans []renderSpan, width int) []renderSpan {
	if width <= 0 {
		return nil
	}
	var out []renderSpan
	cur := 0
	for _, span := range spans {
		if span.Text == "" {
			continue
		}
		remaining := width - cur
		if remaining <= 0 {
			break
		}
		text := runewidth.Truncate(span.Text, remaining, "")
		if text == "" {
			break
		}
		out = append(out, renderSpan{Text: text, Style: span.Style})
		cur += runewidth.StringWidth(text)
	}
	return out
}

func drawRenderLine(buf *runtime.Buffer, bounds runtime.Rect, line renderLine) {
	if buf == nil || bounds.Width <= 0 {
		return
	}
	x := bounds.X
	maxX := bounds.X + bounds.Width
	y := bounds.Y
	for _, span := range line.Spans {
		for _, r := range span.Text {
			if x >= maxX {
				return
			}
			rw := runewidth.RuneWidth(r)
			if rw <= 0 {
				continue
			}
			if x+rw > maxX {
				return
			}
			buf.Set(x, y, r, span.Style)
			x += rw
		}
	}
}

func baseStyleForLine(spans []renderSpan) backend.Style {
	for _, span := range spans {
		if span.Style.BG() != backend.ColorDefault {
			return span.Style
		}
	}
	return backend.DefaultStyle()
}

func convertSpans(spans []markdown.StyledSpan) []renderSpan {
	if len(spans) == 0 {
		return nil
	}
	out := make([]renderSpan, 0, len(spans))
	for _, span := range spans {
		style := backendStyle(span.Style)
		out = append(out, renderSpan{Text: span.Text, Style: style})
	}
	return out
}

func backendStyle(style compositor.Style) backend.Style {
	out := backend.DefaultStyle()
	out = out.Foreground(colorFromCompositor(style.FG))
	out = out.Background(colorFromCompositor(style.BG))
	out = out.Bold(style.Bold)
	out = out.Dim(style.Dim)
	out = out.Italic(style.Italic)
	out = out.Underline(style.Underline)
	out = out.Blink(style.Blink)
	out = out.Reverse(style.Reverse)
	out = out.StrikeThrough(style.Strikethrough)
	return out
}

func colorFromCompositor(c compositor.Color) backend.Color {
	switch c.Mode {
	case compositor.ColorModeNone, compositor.ColorModeDefault:
		return backend.ColorDefault
	case compositor.ColorMode16, compositor.ColorMode256:
		return backend.Color(c.Value)
	case compositor.ColorModeRGB:
		r := uint8((c.Value >> 16) & 0xFF)
		g := uint8((c.Value >> 8) & 0xFF)
		b := uint8(c.Value & 0xFF)
		return backend.ColorRGB(r, g, b)
	default:
		return backend.ColorDefault
	}
}

func drawScrollbar(buf *runtime.Buffer, bounds runtime.Rect, bar scroll.Scrollbar, total, view, offset int) {
	if buf == nil || bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	if total <= view || view <= 0 {
		for y := 0; y < bounds.Height; y++ {
			buf.Set(bounds.X, bounds.Y+y, bar.Chars.Track, bar.Track)
		}
		return
	}
	maxOffset := total - view
	if maxOffset < 0 {
		maxOffset = 0
	}
	thumbSize := int(float64(view) / float64(total) * float64(view))
	if thumbSize < bar.MinThumbSize {
		thumbSize = bar.MinThumbSize
	}
	if thumbSize > view {
		thumbSize = view
	}
	thumbStart := 0
	if maxOffset > 0 {
		thumbStart = int(float64(offset) / float64(maxOffset) * float64(view-thumbSize))
	}
	for i := 0; i < bounds.Height; i++ {
		style := bar.Track
		ch := bar.Chars.Track
		if i >= thumbStart && i < thumbStart+thumbSize {
			style = bar.Thumb
			ch = bar.Chars.Thumb
		}
		buf.Set(bounds.X, bounds.Y+i, ch, style)
	}
}

var _ runtime.Widget = (*MarkdownView)(nil)
var _ scroll.Controller = (*MarkdownView)(nil)
