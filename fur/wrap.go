package fur

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

const defaultTabWidth = 4

func splitTextLines(text string, style Style) []Line {
	parts := strings.Split(text, "\n")
	lines := make([]Line, len(parts))
	for i, part := range parts {
		if part == "" {
			lines[i] = Line{}
			continue
		}
		lines[i] = Line{{Text: part, Style: style}}
	}
	return lines
}

func wrapLines(lines []Line, width int) []Line {
	if width <= 0 {
		return lines
	}
	var out []Line
	for _, line := range lines {
		wrapped := wrapLine(line, width)
		out = append(out, wrapped...)
	}
	return out
}

func wrapLine(line Line, width int) []Line {
	if width <= 0 {
		return []Line{line}
	}
	if len(line) == 0 {
		return []Line{line}
	}
	var out []Line
	var current Line
	currentWidth := 0
	flush := func() {
		out = append(out, current)
		current = nil
		currentWidth = 0
	}
	for _, span := range line {
		tokens := splitTokens(span.Text)
		for _, token := range tokens {
			if token == "" {
				continue
			}
			isSpace := isSpaceToken(token)
			if isSpace && currentWidth == 0 {
				continue
			}
			tokWidth := stringWidth(token)
			if tokWidth == 0 {
				continue
			}
			if currentWidth+tokWidth <= width {
				appendSpan(&current, Span{Text: token, Style: span.Style})
				currentWidth += tokWidth
				continue
			}
			if isSpace {
				flush()
				continue
			}
			if tokWidth > width {
				for _, r := range token {
					if r == '\t' {
						for i := 0; i < defaultTabWidth; i++ {
							if currentWidth+1 > width {
								flush()
							}
							appendSpan(&current, Span{Text: " ", Style: span.Style})
							currentWidth++
						}
						continue
					}
					rw := runewidth.RuneWidth(r)
					if rw == 0 {
						rw = 1
					}
					if currentWidth+rw > width {
						flush()
					}
					appendSpan(&current, Span{Text: string(r), Style: span.Style})
					currentWidth += rw
				}
				continue
			}
			flush()
			appendSpan(&current, Span{Text: token, Style: span.Style})
			currentWidth = tokWidth
		}
	}
	if len(current) > 0 || len(out) == 0 {
		out = append(out, current)
	}
	return out
}

func splitTokens(text string) []string {
	var tokens []string
	var current strings.Builder
	var inSpace bool
	for _, r := range text {
		space := isWhitespace(r)
		if current.Len() == 0 {
			inSpace = space
		}
		if space != inSpace && current.Len() > 0 {
			tokens = append(tokens, current.String())
			current.Reset()
			inSpace = space
		}
		current.WriteRune(r)
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens
}

func isSpaceToken(token string) bool {
	for _, r := range token {
		if !isWhitespace(r) {
			return false
		}
	}
	return true
}

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t'
}

func stringWidth(text string) int {
	width := 0
	for _, r := range text {
		if r == '\t' {
			width += defaultTabWidth
			continue
		}
		rw := runewidth.RuneWidth(r)
		if rw == 0 {
			rw = 1
		}
		width += rw
	}
	return width
}

func lineWidth(line Line) int {
	width := 0
	for _, span := range line {
		width += stringWidth(span.Text)
	}
	return width
}

func appendSpan(line *Line, span Span) {
	if span.Text == "" {
		return
	}
	if line == nil {
		return
	}
	if len(*line) > 0 {
		last := &(*line)[len(*line)-1]
		if last.Style.Equal(span.Style) {
			last.Text += span.Text
			return
		}
	}
	*line = append(*line, span)
}

func truncateLine(line Line, width int) Line {
	if width <= 0 {
		return line
	}
	if lineWidth(line) <= width {
		return line
	}
	var out Line
	currentWidth := 0
	for _, span := range line {
		for _, r := range span.Text {
			if r == '\t' {
				for i := 0; i < defaultTabWidth; i++ {
					if currentWidth+1 > width {
						return out
					}
					appendSpan(&out, Span{Text: " ", Style: span.Style})
					currentWidth++
				}
				continue
			}
			rw := runewidth.RuneWidth(r)
			if rw == 0 {
				rw = 1
			}
			if currentWidth+rw > width {
				return out
			}
			appendSpan(&out, Span{Text: string(r), Style: span.Style})
			currentWidth += rw
		}
	}
	return out
}

func padLine(line Line, width int) Line {
	if width <= 0 {
		return line
	}
	line = truncateLine(line, width)
	pad := width - lineWidth(line)
	if pad <= 0 {
		return line
	}
	appendSpan(&line, Span{Text: strings.Repeat(" ", pad), Style: DefaultStyle()})
	return line
}
