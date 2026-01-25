package fur

import (
	"strconv"
	"strings"
)

// MarkupParser parses BBCode-style markup tags into styled spans.
type MarkupParser struct {
	EnableEmoji bool
}

// NewMarkupParser creates a new markup parser.
func NewMarkupParser() *MarkupParser {
	return &MarkupParser{}
}

var defaultMarkupParser = NewMarkupParser()

// DefaultMarkupParser returns the shared default parser.
func DefaultMarkupParser() *MarkupParser {
	return defaultMarkupParser
}

// Parse converts markup into lines of styled spans.
func (p *MarkupParser) Parse(text string) []Line {
	if p == nil {
		p = DefaultMarkupParser()
	}
	var lines []Line
	var current Line
	stack := []styleFrame{{style: DefaultStyle(), tag: ""}}
	var buffer strings.Builder

	flushBuffer := func() {
		if buffer.Len() == 0 {
			return
		}
		chunk := buffer.String()
		buffer.Reset()
		if p.EnableEmoji {
			chunk = replaceEmoji(chunk)
		}
		parts := strings.Split(chunk, "\n")
		for i, part := range parts {
			if part != "" {
				appendSpan(&current, Span{Text: part, Style: stack[len(stack)-1].style})
			}
			if i < len(parts)-1 {
				lines = append(lines, current)
				current = nil
			}
		}
	}

	for i := 0; i < len(text); {
		ch := text[i]
		if ch == '\\' && i+1 < len(text) {
			next := text[i+1]
			if next == '[' || next == ']' || next == ':' {
				buffer.WriteByte(next)
				i += 2
				continue
			}
		}
		if ch == '[' {
			closeIdx := strings.IndexByte(text[i+1:], ']')
			if closeIdx >= 0 {
				content := text[i+1 : i+1+closeIdx]
				if tag, ok := parseTag(content); ok {
					flushBuffer()
					if tag.closing {
						if tag.name == "" {
							if len(stack) > 1 {
								stack = stack[:len(stack)-1]
							}
						} else {
							for j := len(stack) - 1; j > 0; j-- {
								if stack[j].tag == tag.name {
									stack = stack[:j]
									break
								}
							}
						}
					} else {
						currentStyle := stack[len(stack)-1].style
						currentStyle = tag.applyTo(currentStyle)
						stack = append(stack, styleFrame{style: currentStyle, tag: tag.name})
					}
					i += closeIdx + 2
					continue
				}
			}
		}
		buffer.WriteByte(ch)
		i++
	}
	flushBuffer()
	if len(current) > 0 || len(lines) == 0 {
		lines = append(lines, current)
	}
	return lines
}

type styleFrame struct {
	style Style
	tag   string
}

type tagInfo struct {
	closing bool
	name    string
	delta   styleDelta
}

type styleDelta struct {
	fgSet         bool
	bgSet         bool
	fg            Color
	bg            Color
	bold          bool
	dim           bool
	italic        bool
	underline     bool
	blink         bool
	reverse       bool
	strikethrough bool
}

func (t tagInfo) applyTo(base Style) Style {
	if t.delta.fgSet {
		base.fg = t.delta.fg
	}
	if t.delta.bgSet {
		base.bg = t.delta.bg
	}
	if t.delta.bold {
		base.bold = true
	}
	if t.delta.dim {
		base.dim = true
	}
	if t.delta.italic {
		base.italic = true
	}
	if t.delta.underline {
		base.underline = true
	}
	if t.delta.blink {
		base.blink = true
	}
	if t.delta.reverse {
		base.reverse = true
	}
	if t.delta.strikethrough {
		base.strikethrough = true
	}
	return base
}

func parseTag(content string) (tagInfo, bool) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return tagInfo{}, false
	}
	if strings.HasPrefix(trimmed, "/") {
		name := strings.TrimSpace(strings.TrimPrefix(trimmed, "/"))
		return tagInfo{closing: true, name: strings.ToLower(name)}, true
	}
	tokens := strings.Fields(trimmed)
	if len(tokens) == 0 {
		return tagInfo{}, false
	}
	info := tagInfo{name: strings.ToLower(tokens[0])}
	var delta styleDelta
	applied := false
	for i := 0; i < len(tokens); i++ {
		tok := normalizeToken(tokens[i])
		if tok == "on" {
			if i+1 >= len(tokens) {
				return tagInfo{}, false
			}
			color, ok := parseColorToken(normalizeToken(tokens[i+1]))
			if !ok {
				return tagInfo{}, false
			}
			delta.bg = color
			delta.bgSet = true
			applied = true
			i++
			continue
		}
		switch tok {
		case "bold":
			delta.bold = true
			applied = true
			continue
		case "dim":
			delta.dim = true
			applied = true
			continue
		case "italic":
			delta.italic = true
			applied = true
			continue
		case "underline":
			delta.underline = true
			applied = true
			continue
		case "blink":
			delta.blink = true
			applied = true
			continue
		case "reverse":
			delta.reverse = true
			applied = true
			continue
		case "strike", "strikethrough":
			delta.strikethrough = true
			applied = true
			continue
		}
		if color, ok := parseColorToken(tok); ok {
			delta.fg = color
			delta.fgSet = true
			applied = true
			continue
		}
		return tagInfo{}, false
	}
	if !applied {
		return tagInfo{}, false
	}
	info.delta = delta
	return info, true
}

func normalizeToken(token string) string {
	return strings.ToLower(strings.ReplaceAll(token, "-", "_"))
}

func parseColorToken(token string) (Color, bool) {
	if token == "" {
		return ColorDefault, false
	}
	if strings.HasPrefix(token, "#") {
		hex := strings.TrimPrefix(token, "#")
		if len(hex) != 6 {
			return ColorDefault, false
		}
		value, err := strconv.ParseUint(hex, 16, 32)
		if err != nil {
			return ColorDefault, false
		}
		return Hex(uint32(value)), true
	}
	if strings.HasPrefix(token, "rgb(") && strings.HasSuffix(token, ")") {
		inner := strings.TrimSuffix(strings.TrimPrefix(token, "rgb("), ")")
		parts := strings.Split(inner, ",")
		if len(parts) != 3 {
			return ColorDefault, false
		}
		vals := [3]uint8{}
		for i, part := range parts {
			value, err := strconv.Atoi(strings.TrimSpace(part))
			if err != nil || value < 0 || value > 255 {
				return ColorDefault, false
			}
			vals[i] = uint8(value)
		}
		return RGB(vals[0], vals[1], vals[2]), true
	}
	if strings.HasPrefix(token, "color(") && strings.HasSuffix(token, ")") {
		inner := strings.TrimSuffix(strings.TrimPrefix(token, "color("), ")")
		value, err := strconv.Atoi(strings.TrimSpace(inner))
		if err != nil || value < 0 || value > 255 {
			return ColorDefault, false
		}
		return Color256(uint8(value)), true
	}
	switch token {
	case "black":
		return ColorBlack, true
	case "red":
		return ColorRed, true
	case "green":
		return ColorGreen, true
	case "yellow":
		return ColorYellow, true
	case "blue":
		return ColorBlue, true
	case "magenta":
		return ColorMagenta, true
	case "cyan":
		return ColorCyan, true
	case "white":
		return ColorWhite, true
	case "bright_black", "brightblack":
		return ColorBrightBlack, true
	case "bright_red", "brightred":
		return ColorBrightRed, true
	case "bright_green", "brightgreen":
		return ColorBrightGreen, true
	case "bright_yellow", "brightyellow":
		return ColorBrightYellow, true
	case "bright_blue", "brightblue":
		return ColorBrightBlue, true
	case "bright_magenta", "brightmagenta":
		return ColorBrightMagenta, true
	case "bright_cyan", "brightcyan":
		return ColorBrightCyan, true
	case "bright_white", "brightwhite":
		return ColorBrightWhite, true
	}
	return ColorDefault, false
}
