package fur

import "fmt"

// PanelOpts configures panel rendering.
type PanelOpts struct {
	Style   Style
	Padding int
	Title   string
}

// Panel wraps content in a bordered panel.
func Panel(content any) Renderable {
	return PanelWith(toRenderable(content), PanelOpts{Padding: 1, Style: Dim})
}

// PanelWith wraps content in a bordered panel with options.
func PanelWith(content Renderable, opts PanelOpts) Renderable {
	if opts.Padding < 0 {
		opts.Padding = 0
	}
	return panelRenderable{content: content, opts: opts}
}

type panelRenderable struct {
	content Renderable
	opts    PanelOpts
}

func (p panelRenderable) Render(width int) []Line {
	if width <= 0 {
		width = 80
	}
	padding := p.opts.Padding
	innerWidth := width - 2 - padding*2
	if innerWidth < 1 {
		innerWidth = 1
	}
	lines := renderContent(p.content, innerWidth)
	lines = padContent(lines, innerWidth, padding)
	return renderBox(p.opts.Title, lines, innerWidth+padding*2+2, p.opts.Style)
}

func renderContent(content Renderable, width int) []Line {
	if content == nil {
		return []Line{{}}
	}
	return content.Render(width)
}

func padContent(lines []Line, innerWidth, padding int) []Line {
	if padding <= 0 {
		for i, line := range lines {
			lines[i] = padLine(line, innerWidth)
		}
		return lines
	}
	pad := Span{Text: repeatSpaces(padding), Style: DefaultStyle()}
	for i, line := range lines {
		line = padLine(line, innerWidth)
		var out Line
		out = append(out, pad)
		out = append(out, line...)
		out = append(out, pad)
		lines[i] = out
	}
	return lines
}

func toRenderable(content any) Renderable {
	switch v := content.(type) {
	case Renderable:
		return v
	case string:
		return Text(v)
	default:
		return Text(fmt.Sprint(v))
	}
}
