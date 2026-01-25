package fur

import "strings"

// Alignment controls horizontal alignment.
type Alignment uint8

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
)

// RuleOpts configures the rule.
type RuleOpts struct {
	Style     Style
	Character rune
	Align     Alignment
}

// Rule creates a horizontal line.
func Rule(title ...string) Renderable {
	text := ""
	if len(title) > 0 {
		text = title[0]
	}
	return ruleRenderable{
		title: text,
		opts:  RuleOpts{Character: '─', Align: AlignCenter},
	}
}

// RuleWith creates a rule with options.
func RuleWith(title string, opts RuleOpts) Renderable {
	if opts.Character == 0 {
		opts.Character = '─'
	}
	return ruleRenderable{title: title, opts: opts}
}

type ruleRenderable struct {
	title string
	opts  RuleOpts
}

func (r ruleRenderable) Render(width int) []Line {
	if width <= 0 {
		width = 80
	}
	lineChar := string(r.opts.Character)
	title := strings.TrimSpace(r.title)
	if title == "" {
		return []Line{{{Text: strings.Repeat(lineChar, width), Style: r.opts.Style}}}
	}
	titleText := " " + title + " "
	titleWidth := stringWidth(titleText)
	if titleWidth >= width {
		truncated := truncateString(titleText, width)
		return []Line{{{Text: truncated, Style: r.opts.Style}}}
	}
	remaining := width - titleWidth
	left := 0
	right := 0
	switch r.opts.Align {
	case AlignLeft:
		left = 0
		right = remaining
	case AlignRight:
		left = remaining
		right = 0
	default:
		left = remaining / 2
		right = remaining - left
	}
	line := strings.Repeat(lineChar, left) + titleText + strings.Repeat(lineChar, right)
	return []Line{{{Text: line, Style: r.opts.Style}}}
}

func truncateString(s string, width int) string {
	if width <= 0 {
		return s
	}
	count := 0
	var out strings.Builder
	for _, r := range s {
		rw := stringWidth(string(r))
		if count+rw > width {
			break
		}
		out.WriteRune(r)
		count += rw
	}
	return out.String()
}
