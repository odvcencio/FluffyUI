package widgets

import (
	"strings"

	"github.com/mattn/go-runewidth"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/markdown"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/scroll"
	uistyle "github.com/odvcencio/fluffyui/style"
	"github.com/odvcencio/fluffyui/terminal"
)

// RichTextOption configures a RichText widget.
type RichTextOption = Option[RichText]

// WithRichTextLabel sets the accessibility label.
func WithRichTextLabel(label string) RichTextOption {
	return func(r *RichText) {
		if r == nil {
			return
		}
		r.label = label
	}
}

// WithRichTextStyle sets the base style.
func WithRichTextStyle(style backend.Style) RichTextOption {
	return func(r *RichText) {
		if r == nil {
			return
		}
		r.style = style
		r.styleSet = true
	}
}

// WithRichTextScrollbar toggles the scrollbar.
func WithRichTextScrollbar(show bool) RichTextOption {
	return func(r *RichText) {
		if r == nil {
			return
		}
		r.showBar = show
	}
}

// WithRichTextSource sets the markdown source style key.
func WithRichTextSource(source string) RichTextOption {
	return func(r *RichText) {
		if r == nil {
			return
		}
		r.source = source
	}
}

// WithRichTextRenderer uses a custom markdown renderer.
func WithRichTextRenderer(renderer *markdown.Renderer) RichTextOption {
	return func(r *RichText) {
		if r == nil {
			return
		}
		r.renderer = renderer
	}
}

// RichText renders markdown content with scrolling.
type RichText struct {
	FocusableBase

	content       string
	source        string
	lines         []markdown.StyledLine
	wrapped       []richTextLine
	width         int
	offset        int
	label         string
	style         backend.Style
	styleSet      bool
	showBar       bool
	scrollbar     scroll.Scrollbar
	renderer      *markdown.Renderer
	contentSize   runtime.Size
	anchorOffsets map[string]int
	pendingAnchor string
}

// NewRichText creates a new RichText widget.
func NewRichText(content string, opts ...RichTextOption) *RichText {
	view := &RichText{
		content: content,
		label:   "Rich Text",
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
	for _, opt := range opts {
		opt(view)
	}
	view.renderContent()
	view.syncA11y()
	return view
}

// StyleType returns the selector type name.
func (r *RichText) StyleType() string {
	return "RichText"
}

// SetContent updates the markdown content.
func (r *RichText) SetContent(content string) {
	if r == nil {
		return
	}
	r.content = content
	r.renderContent()
	r.resetLayout()
}

// SetLines sets pre-rendered styled lines.
func (r *RichText) SetLines(lines []markdown.StyledLine) {
	if r == nil {
		return
	}
	r.lines = lines
	r.content = ""
	r.resetLayout()
}

// SetSource sets the markdown source style key.
func (r *RichText) SetSource(source string) {
	if r == nil {
		return
	}
	r.source = strings.TrimSpace(source)
	r.renderContent()
	r.resetLayout()
}

// SetRenderer replaces the markdown renderer.
func (r *RichText) SetRenderer(renderer *markdown.Renderer) {
	if r == nil {
		return
	}
	r.renderer = renderer
	r.renderContent()
	r.resetLayout()
}

// SetLabel updates the accessibility label.
func (r *RichText) SetLabel(label string) {
	if r == nil {
		return
	}
	r.label = label
	r.syncA11y()
}

// SetShowScrollbar toggles the scrollbar.
func (r *RichText) SetShowScrollbar(show bool) {
	if r == nil {
		return
	}
	r.showBar = show
	r.Invalidate()
}

// SetStyle updates the base style.
func (r *RichText) SetStyle(style backend.Style) {
	if r == nil {
		return
	}
	r.style = style
	r.styleSet = true
	r.Invalidate()
}

// Measure returns the required size.
func (r *RichText) Measure(constraints runtime.Constraints) runtime.Size {
	return r.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		width := contentConstraints.MaxWidth
		if width <= 0 {
			width = contentConstraints.MinWidth
		}
		if width <= 0 {
			width = 1
		}
		r.wrap(width)
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: len(r.wrapped)})
	})
}

// Layout stores bounds and updates wrapping.
func (r *RichText) Layout(bounds runtime.Rect) {
	r.FocusableBase.Layout(bounds)
	content := r.ContentBounds()
	if content.Width <= 0 || content.Height <= 0 {
		return
	}
	width := content.Width
	if r.showBar && r.needsScrollbar(content.Height) && width > 1 {
		width--
	}
	if width < 1 {
		width = 1
	}
	if width != r.width {
		r.wrap(width)
	}
	r.clampOffset(content.Height)
}

// Render draws the visible lines.
func (r *RichText) Render(ctx runtime.RenderContext) {
	if r == nil {
		return
	}
	r.syncA11y()
	outer := r.bounds
	content := r.ContentBounds()
	if outer.Width <= 0 || outer.Height <= 0 {
		return
	}
	baseStyle := resolveBaseStyle(ctx, r, backend.DefaultStyle(), false)
	if r.styleSet {
		baseStyle = mergeBackendStyles(baseStyle, r.style)
	}
	ctx.Buffer.Fill(outer, ' ', baseStyle)
	if content.Width <= 0 || content.Height <= 0 {
		return
	}
	visibleWidth := content.Width
	showBar := r.showBar && r.needsScrollbar(content.Height) && content.Width > 1
	if showBar {
		visibleWidth--
	}

	for row := 0; row < content.Height; row++ {
		lineIndex := r.offset + row
		if lineIndex < 0 || lineIndex >= len(r.wrapped) {
			continue
		}
		line := r.wrapped[lineIndex]
		lineBounds := runtime.Rect{X: content.X, Y: content.Y + row, Width: visibleWidth, Height: 1}
		if line.BaseStyle != backend.DefaultStyle() {
			ctx.Buffer.Fill(lineBounds, ' ', line.BaseStyle)
		}
		drawRichTextLine(ctx.Buffer, lineBounds, line)
	}

	if showBar {
		barBounds := runtime.Rect{
			X:      content.X + visibleWidth,
			Y:      content.Y,
			Width:  1,
			Height: content.Height,
		}
		drawScrollbar(ctx.Buffer, barBounds, r.scrollbar, len(r.wrapped), content.Height, r.offset)
	}
}

// HandleMessage handles scroll input.
func (r *RichText) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if r == nil {
		return runtime.Unhandled()
	}
	switch ev := msg.(type) {
	case runtime.KeyMsg:
		if !r.IsFocused() {
			return runtime.Unhandled()
		}
		switch ev.Key {
		case terminal.KeyUp:
			r.ScrollBy(0, -1)
			return runtime.Handled()
		case terminal.KeyDown:
			r.ScrollBy(0, 1)
			return runtime.Handled()
		case terminal.KeyPageUp:
			r.PageBy(-1)
			return runtime.Handled()
		case terminal.KeyPageDown:
			r.PageBy(1)
			return runtime.Handled()
		case terminal.KeyHome:
			r.ScrollToStart()
			return runtime.Handled()
		case terminal.KeyEnd:
			r.ScrollToEnd()
			return runtime.Handled()
		}
	case runtime.MouseMsg:
		if ev.Button == runtime.MouseWheelUp {
			r.ScrollBy(0, -3)
			return runtime.Handled()
		}
		if ev.Button == runtime.MouseWheelDown {
			r.ScrollBy(0, 3)
			return runtime.Handled()
		}
	}
	return runtime.Unhandled()
}

// ScrollBy scrolls the content by delta.
func (r *RichText) ScrollBy(dx, dy int) {
	if r == nil || dy == 0 {
		return
	}
	r.offset += dy
	r.clampOffset(r.ContentBounds().Height)
	r.Invalidate()
}

// ScrollTo scrolls to an absolute offset.
func (r *RichText) ScrollTo(x, y int) {
	if r == nil {
		return
	}
	r.offset = y
	r.clampOffset(r.ContentBounds().Height)
	r.Invalidate()
}

// PageBy scrolls by pages.
func (r *RichText) PageBy(pages int) {
	if r == nil {
		return
	}
	pageSize := r.ContentBounds().Height
	if pageSize < 1 {
		pageSize = 1
	}
	r.ScrollBy(0, pages*pageSize)
}

// ScrollToStart scrolls to the top.
func (r *RichText) ScrollToStart() {
	r.ScrollTo(0, 0)
}

// ScrollToEnd scrolls to the bottom.
func (r *RichText) ScrollToEnd() {
	if r == nil {
		return
	}
	height := r.ContentBounds().Height
	maxOffset := max(0, len(r.wrapped)-height)
	r.ScrollTo(0, maxOffset)
}

// ScrollToAnchor scrolls to a heading anchor if available.
func (r *RichText) ScrollToAnchor(anchor string) {
	if r == nil || anchor == "" {
		return
	}
	r.pendingAnchor = anchor
	if r.anchorOffsets == nil {
		return
	}
	if offset, ok := r.anchorOffsets[anchor]; ok {
		r.pendingAnchor = ""
		r.ScrollTo(0, offset)
	}
}

func (r *RichText) renderContent() {
	if r == nil {
		return
	}
	if r.renderer == nil {
		r.renderer = markdown.NewRenderer(nil)
	}
	if strings.TrimSpace(r.content) == "" {
		r.lines = nil
		return
	}
	r.lines = r.renderer.Render(r.source, r.content)
}

func (r *RichText) resetLayout() {
	r.wrapped = nil
	r.anchorOffsets = nil
	r.pendingAnchor = ""
	if r.width > 0 {
		r.wrap(r.width)
	}
	r.offset = 0
	r.Invalidate()
}

func (r *RichText) clampOffset(viewHeight int) {
	maxOffset := max(0, len(r.wrapped)-viewHeight)
	if r.offset < 0 {
		r.offset = 0
	}
	if r.offset > maxOffset {
		r.offset = maxOffset
	}
}

func (r *RichText) needsScrollbar(viewHeight int) bool {
	return len(r.wrapped) > viewHeight
}

func (r *RichText) wrap(width int) {
	if r == nil {
		return
	}
	if width < 1 {
		width = 1
	}
	if r.width == width && len(r.wrapped) > 0 {
		return
	}
	r.width = width
	r.wrapped = wrapRichTextLines(r.lines, width)
	r.contentSize = runtime.Size{Width: width, Height: len(r.wrapped)}
	r.anchorOffsets = map[string]int{}
	for i, line := range r.wrapped {
		if line.Anchor == "" {
			continue
		}
		if _, ok := r.anchorOffsets[line.Anchor]; !ok {
			r.anchorOffsets[line.Anchor] = i
		}
	}
	if r.pendingAnchor != "" {
		if offset, ok := r.anchorOffsets[r.pendingAnchor]; ok {
			r.pendingAnchor = ""
			r.ScrollTo(0, offset)
		}
	}
}

func (r *RichText) syncA11y() {
	if r == nil {
		return
	}
	if r.Base.Role == "" {
		r.Base.Role = accessibility.RoleText
	}
	label := strings.TrimSpace(r.label)
	if label == "" {
		label = "Rich Text"
	}
	r.Base.Label = label
}

type richTextSpan struct {
	Text  string
	Style backend.Style
}

type richTextLine struct {
	Spans     []richTextSpan
	BlankLine bool
	BaseStyle backend.Style
	Anchor    string
}

func wrapRichTextLines(lines []markdown.StyledLine, width int) []richTextLine {
	if width < 1 {
		return nil
	}
	out := make([]richTextLine, 0, len(lines))
	for _, line := range lines {
		out = append(out, wrapRichTextLine(line, width)...)
	}
	return out
}

func wrapRichTextLine(line markdown.StyledLine, width int) []richTextLine {
	if width < 1 {
		return nil
	}
	if line.BlankLine && len(line.Spans) == 0 && len(line.Prefix) == 0 {
		return []richTextLine{{BlankLine: true}}
	}
	prefix := convertRichTextSpans(line.Prefix)
	prefixWidth := richTextSpanWidth(prefix)
	if prefixWidth > width {
		prefix = truncateRichTextSpans(prefix, width)
		prefixWidth = richTextSpanWidth(prefix)
	}
	var lines []richTextLine
	current := newRichTextLine(prefix, line.Anchor)
	curWidth := prefixWidth
	appendLine := func() {
		lines = append(lines, current)
		current = newRichTextLine(prefix, "")
		curWidth = prefixWidth
	}
	for _, span := range convertRichTextSpans(line.Spans) {
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
			appendRichTextRune(&current, r, span.Style)
			curWidth += rw
		}
	}
	if len(current.Spans) > 0 || prefixWidth > 0 || line.BlankLine {
		lines = append(lines, current)
	}
	if len(lines) == 0 {
		lines = append(lines, richTextLine{BlankLine: true})
	}
	return lines
}

func newRichTextLine(prefix []richTextSpan, anchor string) richTextLine {
	line := richTextLine{Anchor: anchor}
	if len(prefix) > 0 {
		line.Spans = append(line.Spans, prefix...)
	}
	line.BaseStyle = baseStyleForRichTextLine(line.Spans)
	return line
}

func appendRichTextRune(line *richTextLine, r rune, style backend.Style) {
	if line == nil {
		return
	}
	if line.BaseStyle == backend.DefaultStyle() && style.BG() != backend.ColorDefault {
		line.BaseStyle = style
	}
	if len(line.Spans) == 0 {
		line.Spans = append(line.Spans, richTextSpan{Text: string(r), Style: style})
		return
	}
	last := &line.Spans[len(line.Spans)-1]
	if last.Style == style {
		last.Text += string(r)
		return
	}
	line.Spans = append(line.Spans, richTextSpan{Text: string(r), Style: style})
}

func richTextSpanWidth(spans []richTextSpan) int {
	width := 0
	for _, span := range spans {
		width += runewidth.StringWidth(span.Text)
	}
	return width
}

func truncateRichTextSpans(spans []richTextSpan, width int) []richTextSpan {
	if width <= 0 {
		return nil
	}
	var out []richTextSpan
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
		out = append(out, richTextSpan{Text: text, Style: span.Style})
		cur += runewidth.StringWidth(text)
	}
	return out
}

func drawRichTextLine(buf *runtime.Buffer, bounds runtime.Rect, line richTextLine) {
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

func baseStyleForRichTextLine(spans []richTextSpan) backend.Style {
	for _, span := range spans {
		if span.Style.BG() != backend.ColorDefault {
			return span.Style
		}
	}
	return backend.DefaultStyle()
}

func convertRichTextSpans(spans []markdown.StyledSpan) []richTextSpan {
	if len(spans) == 0 {
		return nil
	}
	out := make([]richTextSpan, 0, len(spans))
	for _, span := range spans {
		out = append(out, richTextSpan{Text: span.Text, Style: uistyle.ToBackend(span.Style)})
	}
	return out
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

var _ runtime.Widget = (*RichText)(nil)
var _ runtime.Focusable = (*RichText)(nil)
var _ scroll.Controller = (*RichText)(nil)
