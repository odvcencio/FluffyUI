package fur

// Renderable can be rendered to a console.
type Renderable interface {
	Render(width int) []Line
}

// Line is a line of styled text.
type Line []Span

// Span is text with a style.
type Span struct {
	Text  string
	Style Style
}

// Text returns a plain text renderable.
func Text(s string) Renderable {
	return textRenderable{text: s}
}

// Markup returns a renderable that parses markup tags.
func Markup(s string) Renderable {
	return markupRenderable{text: s}
}

// Group groups multiple renderables vertically.
func Group(items ...Renderable) Renderable {
	return groupRenderable{items: items}
}

type textRenderable struct {
	text string
}

func (t textRenderable) Render(width int) []Line {
	lines := splitTextLines(t.text, DefaultStyle())
	return wrapLines(lines, width)
}

type markupRenderable struct {
	text string
}

func (m markupRenderable) Render(width int) []Line {
	parser := DefaultMarkupParser()
	lines := parser.Parse(m.text)
	return wrapLines(lines, width)
}

type groupRenderable struct {
	items []Renderable
}

func (g groupRenderable) Render(width int) []Line {
	var out []Line
	for _, item := range g.items {
		if item == nil {
			continue
		}
		lines := item.Render(width)
		if len(lines) == 0 {
			continue
		}
		out = append(out, lines...)
	}
	return out
}
