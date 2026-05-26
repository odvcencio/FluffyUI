package widgets

import (
	"bytes"
	"fmt"
	"strings"

	"m31labs.dev/fluffyui/accessibility"
	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/markdown"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/scroll"
	"m31labs.dev/fluffyui/terminal"
	"m31labs.dev/fluffyui/theme"
	"github.com/yuin/goldmark"
)

// MarkdownViewerOption configures a MarkdownViewer widget.
type MarkdownViewerOption = Option[MarkdownViewer]

// WithMarkdownTheme sets a custom theme for markdown rendering.
func WithMarkdownTheme(t *theme.Theme) MarkdownViewerOption {
	return func(m *MarkdownViewer) {
		if m == nil {
			return
		}
		m.mdTheme = t
		m.renderer = markdown.NewRenderer(t)
		m.renderContent()
	}
}

// WithMarkdownMaxWidth sets a maximum content width for readability.
// Lines will be wrapped at this width even if more space is available.
// A value of 0 disables the limit (uses full available width).
func WithMarkdownMaxWidth(width int) MarkdownViewerOption {
	return func(m *MarkdownViewer) {
		if m == nil {
			return
		}
		m.maxWidth = width
	}
}

// WithWordWrap enables or disables word wrapping.
func WithWordWrap(enabled bool) MarkdownViewerOption {
	return func(m *MarkdownViewer) {
		if m == nil {
			return
		}
		m.wordWrap = enabled
	}
}

// MarkdownViewer renders markdown content as styled terminal text.
// Supports headings, bold, italic, code blocks, links, lists, tables, and blockquotes.
// Content is scrollable when it exceeds the available height.
type MarkdownViewer struct {
	FocusableBase

	content   string
	lines     []markdown.StyledLine
	wrapped   []richTextLine
	width     int
	offset    int
	style     backend.Style
	styleSet  bool
	showBar   bool
	wordWrap  bool
	maxWidth  int
	scrollbar scroll.Scrollbar
	renderer  *markdown.Renderer
	mdTheme   *theme.Theme
}

// NewMarkdownViewer creates a new MarkdownViewer widget.
func NewMarkdownViewer(content string, opts ...MarkdownViewerOption) *MarkdownViewer {
	m := &MarkdownViewer{
		content:  content,
		style:    backend.DefaultStyle(),
		showBar:  true,
		wordWrap: true,
		scrollbar: scroll.Scrollbar{
			Orientation:  scroll.Vertical,
			Track:        backend.DefaultStyle(),
			Thumb:        backend.DefaultStyle().Reverse(true),
			MinThumbSize: 1,
			Chars:        scroll.DefaultScrollbarChars(),
		},
	}
	m.Base.Role = accessibility.RoleDocument
	for _, opt := range opts {
		opt(m)
	}
	m.renderContent()
	m.syncA11y()
	return m
}

// StyleType returns the selector type name.
func (m *MarkdownViewer) StyleType() string {
	return "MarkdownViewer"
}

// SetContent updates the markdown content.
func (m *MarkdownViewer) SetContent(content string) {
	if m == nil {
		return
	}
	m.content = content
	m.renderContent()
	m.resetLayout()
}

// Content returns the raw markdown content.
func (m *MarkdownViewer) Content() string {
	if m == nil {
		return ""
	}
	return m.content
}

// Measure returns the required size.
func (m *MarkdownViewer) Measure(constraints runtime.Constraints) runtime.Size {
	return m.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		width := contentConstraints.MaxWidth
		if width <= 0 {
			width = contentConstraints.MinWidth
		}
		if width <= 0 {
			width = 1
		}
		if m.maxWidth > 0 && width > m.maxWidth {
			width = m.maxWidth
		}
		m.wrap(width)
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: len(m.wrapped)})
	})
}

// Layout stores bounds and updates wrapping.
func (m *MarkdownViewer) Layout(bounds runtime.Rect) {
	m.FocusableBase.Layout(bounds)
	content := m.ContentBounds()
	if content.Width <= 0 || content.Height <= 0 {
		return
	}
	width := content.Width
	if m.maxWidth > 0 && width > m.maxWidth {
		width = m.maxWidth
	}
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
func (m *MarkdownViewer) Render(ctx runtime.RenderContext) {
	if m == nil {
		return
	}
	m.syncA11y()
	outer := m.bounds
	content := m.ContentBounds()
	if outer.Width <= 0 || outer.Height <= 0 {
		return
	}
	baseStyle := resolveBaseStyle(ctx, m, backend.DefaultStyle(), false)
	if m.styleSet {
		baseStyle = mergeBackendStyles(baseStyle, m.style)
	}
	ctx.Buffer.Fill(outer, ' ', baseStyle)
	if content.Width <= 0 || content.Height <= 0 {
		return
	}
	visibleWidth := content.Width
	if m.maxWidth > 0 && visibleWidth > m.maxWidth {
		visibleWidth = m.maxWidth
	}
	showBar := m.showBar && m.needsScrollbar(content.Height) && visibleWidth > 1
	if showBar {
		visibleWidth--
	}

	for row := 0; row < content.Height; row++ {
		lineIndex := m.offset + row
		if lineIndex < 0 || lineIndex >= len(m.wrapped) {
			continue
		}
		line := m.wrapped[lineIndex]
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
		drawScrollbar(ctx.Buffer, barBounds, m.scrollbar, len(m.wrapped), content.Height, m.offset)
	}
}

// RenderHTML renders the markdown content as static HTML.
func (m *MarkdownViewer) RenderHTML(ctx runtime.HTMLContext) runtime.HTML {
	var buf bytes.Buffer
	md := goldmark.New()
	_ = md.Convert([]byte(m.content), &buf)
	return runtime.HTML(fmt.Sprintf(`<article class="fluffy-MarkdownViewer">%s</article>`, buf.String()))
}

// HandleMessage handles scroll input.
func (m *MarkdownViewer) HandleMessage(msg runtime.Message) runtime.HandleResult {
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
func (m *MarkdownViewer) ScrollBy(dx, dy int) {
	if m == nil || dy == 0 {
		return
	}
	m.offset += dy
	m.clampOffset(m.ContentBounds().Height)
	m.Invalidate()
}

// ScrollTo scrolls to an absolute offset.
func (m *MarkdownViewer) ScrollTo(x, y int) {
	if m == nil {
		return
	}
	m.offset = y
	m.clampOffset(m.ContentBounds().Height)
	m.Invalidate()
}

// PageBy scrolls by pages.
func (m *MarkdownViewer) PageBy(pages int) {
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
func (m *MarkdownViewer) ScrollToStart() {
	m.ScrollTo(0, 0)
}

// ScrollToEnd scrolls to the bottom.
func (m *MarkdownViewer) ScrollToEnd() {
	if m == nil {
		return
	}
	height := m.ContentBounds().Height
	maxOffset := max(0, len(m.wrapped)-height)
	m.ScrollTo(0, maxOffset)
}

func (m *MarkdownViewer) renderContent() {
	if m == nil {
		return
	}
	if m.renderer == nil {
		m.renderer = markdown.NewRenderer(m.mdTheme)
	}
	if strings.TrimSpace(m.content) == "" {
		m.lines = nil
		return
	}
	m.lines = m.renderer.Render("", m.content)
}

func (m *MarkdownViewer) resetLayout() {
	m.wrapped = nil
	if m.width > 0 {
		m.wrap(m.width)
	}
	m.offset = 0
	m.Invalidate()
}

func (m *MarkdownViewer) clampOffset(viewHeight int) {
	maxOffset := max(0, len(m.wrapped)-viewHeight)
	if m.offset < 0 {
		m.offset = 0
	}
	if m.offset > maxOffset {
		m.offset = maxOffset
	}
}

func (m *MarkdownViewer) needsScrollbar(viewHeight int) bool {
	return len(m.wrapped) > viewHeight
}

func (m *MarkdownViewer) wrap(width int) {
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
	if m.wordWrap {
		m.wrapped = wrapRichTextLines(m.lines, width)
	} else {
		// Without word wrap, still use the standard wrapping to convert
		// StyledLine to richTextLine, but use a very large width so
		// lines are not broken.
		m.wrapped = wrapRichTextLines(m.lines, 1<<20)
	}
}

func (m *MarkdownViewer) syncA11y() {
	if m == nil {
		return
	}
	if m.Base.Role == "" {
		m.Base.Role = accessibility.RoleDocument
	}
	m.Base.Label = "Markdown viewer"
}

var _ runtime.Widget = (*MarkdownViewer)(nil)
var _ runtime.Focusable = (*MarkdownViewer)(nil)
var _ scroll.Controller = (*MarkdownViewer)(nil)
