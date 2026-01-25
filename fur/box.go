package fur

import "strings"

func renderBox(title string, content []Line, width int, borderStyle Style) []Line {
	if len(content) == 0 {
		content = []Line{{}}
	}
	title = strings.TrimSpace(title)
	contentWidth := 0
	for _, line := range content {
		if w := lineWidth(line); w > contentWidth {
			contentWidth = w
		}
	}
	titleText := ""
	titleWidth := 0
	if title != "" {
		titleText = " " + title + " "
		titleWidth = stringWidth(titleText)
	}
	if width <= 0 {
		inner := contentWidth
		if titleWidth > inner {
			inner = titleWidth
		}
		width = inner + 2
	}
	innerWidth := width - 2
	if innerWidth < 1 {
		innerWidth = 1
	}
	if titleWidth > innerWidth {
		titleText = truncateString(titleText, innerWidth)
		titleWidth = stringWidth(titleText)
	}

	var out []Line
	if titleText == "" {
		out = append(out, Line{{Text: "╭" + strings.Repeat("─", innerWidth) + "╮", Style: borderStyle}})
	} else {
		remaining := innerWidth - titleWidth
		left := remaining / 2
		right := remaining - left
		line := "╭" + strings.Repeat("─", left) + titleText + strings.Repeat("─", right) + "╮"
		out = append(out, Line{{Text: line, Style: borderStyle}})
	}

	for _, line := range content {
		line = padLine(line, innerWidth)
		var boxed Line
		boxed = append(boxed, Span{Text: "│", Style: borderStyle})
		boxed = append(boxed, line...)
		boxed = append(boxed, Span{Text: "│", Style: borderStyle})
		out = append(out, boxed)
	}

	out = append(out, Line{{Text: "╰" + strings.Repeat("─", innerWidth) + "╯", Style: borderStyle}})
	return out
}
