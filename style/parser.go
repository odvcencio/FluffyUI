package style

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Parse parses a Fluffy Style Sheet (FSS) string.
func Parse(data string) (*Stylesheet, error) {
	parser := fssParser{src: stripComments(data)}
	sheet := NewStylesheet()

	for {
		parser.skipWhitespace()
		if parser.eof() {
			break
		}
		selectors, err := parser.readUntil('{')
		if err != nil {
			return nil, err
		}
		selectorText := strings.TrimSpace(selectors)
		if selectorText == "" {
			return nil, fmt.Errorf("style: empty selector")
		}
		block, err := parser.readUntil('}')
		if err != nil {
			return nil, err
		}
		styleBlock, err := parseStyleBlock(block)
		if err != nil {
			return nil, err
		}
		selectorList := strings.Split(selectorText, ",")
		for _, rawSel := range selectorList {
			selText := strings.TrimSpace(rawSel)
			if selText == "" {
				return nil, fmt.Errorf("style: empty selector entry")
			}
			sel, err := parseSelectorChain(selText)
			if err != nil {
				return nil, err
			}
			sheet.Add(&SelectorBuilder{sel: *sel}, styleBlock)
		}
	}

	return sheet, nil
}

// ParseFile parses a FSS file from disk.
func ParseFile(path string) (*Stylesheet, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(string(data))
}

// MustParse parses a FSS string or panics.
func MustParse(data string) *Stylesheet {
	sheet, err := Parse(data)
	if err != nil {
		panic(err)
	}
	return sheet
}

type fssParser struct {
	src string
	pos int
}

func (p *fssParser) eof() bool {
	return p.pos >= len(p.src)
}

func (p *fssParser) skipWhitespace() {
	for p.pos < len(p.src) {
		switch p.src[p.pos] {
		case ' ', '\t', '\n', '\r':
			p.pos++
		default:
			return
		}
	}
}

func (p *fssParser) readUntil(delim byte) (string, error) {
	start := p.pos
	idx := strings.IndexByte(p.src[p.pos:], delim)
	if idx == -1 {
		return "", fmt.Errorf("style: expected %q", delim)
	}
	p.pos += idx + 1
	return p.src[start : start+idx], nil
}

func parseStyleBlock(block string) (Style, error) {
	var out Style
	entries := strings.Split(block, ";")
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		name, value, ok := strings.Cut(entry, ":")
		if !ok {
			return Style{}, fmt.Errorf("style: expected ':' in declaration %q", entry)
		}
		name = strings.ToLower(strings.TrimSpace(name))
		value = strings.TrimSpace(value)
		if name == "" {
			return Style{}, fmt.Errorf("style: missing property name")
		}
		switch name {
		case "foreground", "color", "fg":
			color, err := parseColor(value)
			if err != nil {
				return Style{}, err
			}
			out.Foreground = color
		case "background", "bg":
			color, err := parseColor(value)
			if err != nil {
				return Style{}, err
			}
			out.Background = color
		case "bold":
			value, err := parseBool(value)
			if err != nil {
				return Style{}, err
			}
			out.Bold = Bool(value)
		case "italic":
			value, err := parseBool(value)
			if err != nil {
				return Style{}, err
			}
			out.Italic = Bool(value)
		case "underline":
			value, err := parseBool(value)
			if err != nil {
				return Style{}, err
			}
			out.Underline = Bool(value)
		case "dim":
			value, err := parseBool(value)
			if err != nil {
				return Style{}, err
			}
			out.Dim = Bool(value)
		case "blink":
			value, err := parseBool(value)
			if err != nil {
				return Style{}, err
			}
			out.Blink = Bool(value)
		case "reverse":
			value, err := parseBool(value)
			if err != nil {
				return Style{}, err
			}
			out.Reverse = Bool(value)
		case "strikethrough":
			value, err := parseBool(value)
			if err != nil {
				return Style{}, err
			}
			out.Strikethrough = Bool(value)
		case "padding":
			spacing, err := parseSpacing(value)
			if err != nil {
				return Style{}, err
			}
			out.Padding = spacing
		case "margin":
			spacing, err := parseSpacing(value)
			if err != nil {
				return Style{}, err
			}
			out.Margin = spacing
		case "width":
			size, err := parseSize(value)
			if err != nil {
				return Style{}, err
			}
			out.Width = size
		case "height":
			size, err := parseSize(value)
			if err != nil {
				return Style{}, err
			}
			out.Height = size
		case "border":
			border, err := parseBorder(value)
			if err != nil {
				return Style{}, err
			}
			out.Border = border
		case "border-radius":
			value, err := parseBool(value)
			if err != nil {
				return Style{}, err
			}
			out.BorderRadius = Bool(value)
		default:
			return Style{}, fmt.Errorf("style: unknown property %q", name)
		}
	}
	return out, nil
}

func parseSelectorChain(text string) (*Selector, error) {
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return nil, fmt.Errorf("style: empty selector")
	}
	var current *Selector
	for _, part := range parts {
		sel, err := parseSelectorPart(part)
		if err != nil {
			return nil, err
		}
		if current != nil {
			sel.Parent = current
		}
		current = sel
	}
	return current, nil
}

func parseSelectorPart(part string) (*Selector, error) {
	if part == "" {
		return nil, fmt.Errorf("style: empty selector part")
	}
	idx := 0
	out := &Selector{Type: "*"}
	if part[idx] == '*' {
		out.Type = "*"
		idx++
	} else if isIdentStart(part[idx]) {
		name, next := readIdent(part, idx)
		out.Type = name
		idx = next
	}
	for idx < len(part) {
		switch part[idx] {
		case '#':
			idx++
			name, next := readIdent(part, idx)
			if name == "" {
				return nil, fmt.Errorf("style: invalid id in selector %q", part)
			}
			out.ID = name
			idx = next
		case '.':
			idx++
			name, next := readIdent(part, idx)
			if name == "" {
				return nil, fmt.Errorf("style: invalid class in selector %q", part)
			}
			out.Classes = append(out.Classes, name)
			idx = next
		case ':':
			idx++
			name, next := readIdent(part, idx)
			if name == "" {
				return nil, fmt.Errorf("style: invalid pseudo-class in selector %q", part)
			}
			pseudo, ok := parsePseudo(name)
			if !ok {
				return nil, fmt.Errorf("style: unknown pseudo-class %q", name)
			}
			out.Pseudo = append(out.Pseudo, pseudo)
			idx = next
		default:
			return nil, fmt.Errorf("style: invalid selector %q", part)
		}
	}
	return out, nil
}

func parsePseudo(name string) (PseudoClass, bool) {
	switch strings.ToLower(name) {
	case string(PseudoFocus):
		return PseudoFocus, true
	case string(PseudoDisabled):
		return PseudoDisabled, true
	case string(PseudoHover):
		return PseudoHover, true
	case string(PseudoActive):
		return PseudoActive, true
	case string(PseudoFirst):
		return PseudoFirst, true
	case string(PseudoLast):
		return PseudoLast, true
	default:
		return "", false
	}
}

func parseBool(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("style: invalid boolean %q", value)
	}
}

func parseSpacing(value string) (*Spacing, error) {
	value = strings.ReplaceAll(value, ",", " ")
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return nil, fmt.Errorf("style: missing spacing")
	}
	values := make([]int, 0, len(fields))
	for _, field := range fields {
		v, err := strconv.Atoi(field)
		if err != nil {
			return nil, fmt.Errorf("style: invalid spacing %q", field)
		}
		values = append(values, v)
	}
	switch len(values) {
	case 1:
		return Pad(values[0]), nil
	case 2:
		return PadXY(values[1], values[0]), nil
	case 4:
		return PadTRBL(values[0], values[1], values[2], values[3]), nil
	default:
		return nil, fmt.Errorf("style: expected 1, 2, or 4 spacing values")
	}
}

func parseSize(value string) (*Size, error) {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return nil, fmt.Errorf("style: missing size")
	}
	switch value {
	case "auto":
		return Auto(), nil
	case "fill":
		return Fill(), nil
	}
	if strings.HasSuffix(value, "%") {
		v := strings.TrimSuffix(value, "%")
		pct, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return nil, fmt.Errorf("style: invalid percent %q", value)
		}
		return Percent(pct), nil
	}
	fixed, err := strconv.Atoi(value)
	if err != nil {
		return nil, fmt.Errorf("style: invalid size %q", value)
	}
	return Fixed(fixed), nil
}

func parseBorder(value string) (*Border, error) {
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return nil, fmt.Errorf("style: missing border")
	}
	var borderStyle BorderStyle
	styleSet := false
	var borderColor Color
	colorSet := false
	for _, field := range fields {
		fieldLower := strings.ToLower(field)
		switch fieldLower {
		case "none":
			borderStyle = BorderNone
			styleSet = true
		case "single":
			borderStyle = BorderSingle
			styleSet = true
		case "double":
			borderStyle = BorderDouble
			styleSet = true
		case "rounded":
			borderStyle = BorderRounded
			styleSet = true
		default:
			color, err := parseColor(field)
			if err != nil {
				return nil, err
			}
			borderColor = color
			colorSet = true
		}
	}
	if !styleSet {
		return nil, fmt.Errorf("style: missing border style")
	}
	border := &Border{Style: borderStyle}
	if colorSet {
		border.Color = borderColor
	}
	return border, nil
}

func parseColor(value string) (Color, error) {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return ColorNone, fmt.Errorf("style: missing color")
	}
	if strings.HasPrefix(value, "#") {
		hex := strings.TrimPrefix(value, "#")
		switch len(hex) {
		case 3:
			hex = fmt.Sprintf("%c%c%c%c%c%c", hex[0], hex[0], hex[1], hex[1], hex[2], hex[2])
		case 6:
			// ok
		default:
			return ColorNone, fmt.Errorf("style: invalid hex color %q", value)
		}
		parsed, err := strconv.ParseUint(hex, 16, 32)
		if err != nil {
			return ColorNone, fmt.Errorf("style: invalid hex color %q", value)
		}
		r := uint8((parsed >> 16) & 0xFF)
		g := uint8((parsed >> 8) & 0xFF)
		b := uint8(parsed & 0xFF)
		return RGB(r, g, b), nil
	}
	if strings.HasPrefix(value, "0x") {
		parsed, err := strconv.ParseUint(strings.TrimPrefix(value, "0x"), 16, 32)
		if err != nil {
			return ColorNone, fmt.Errorf("style: invalid hex color %q", value)
		}
		r := uint8((parsed >> 16) & 0xFF)
		g := uint8((parsed >> 8) & 0xFF)
		b := uint8(parsed & 0xFF)
		return RGB(r, g, b), nil
	}
	if strings.HasPrefix(value, "rgb(") && strings.HasSuffix(value, ")") {
		inner := strings.TrimSuffix(strings.TrimPrefix(value, "rgb("), ")")
		parts := strings.Split(inner, ",")
		if len(parts) != 3 {
			return ColorNone, fmt.Errorf("style: invalid rgb color %q", value)
		}
		r, err := parseColorChannel(parts[0])
		if err != nil {
			return ColorNone, err
		}
		g, err := parseColorChannel(parts[1])
		if err != nil {
			return ColorNone, err
		}
		b, err := parseColorChannel(parts[2])
		if err != nil {
			return ColorNone, err
		}
		return RGB(r, g, b), nil
	}
	if color, ok := namedColors[value]; ok {
		return color, nil
	}
	return ColorNone, fmt.Errorf("style: unknown color %q", value)
}

func parseColorChannel(value string) (uint8, error) {
	v, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("style: invalid rgb channel %q", value)
	}
	if v < 0 || v > 255 {
		return 0, fmt.Errorf("style: rgb channel out of range: %d", v)
	}
	return uint8(v), nil
}

var namedColors = map[string]Color{
	"none":       ColorNone,
	"default":    ColorDefault,
	"black":      ColorBlack,
	"red":        ColorRed,
	"green":      ColorGreen,
	"yellow":     ColorYellow,
	"blue":       ColorBlue,
	"magenta":    ColorMagenta,
	"cyan":       ColorCyan,
	"white":      ColorWhite,
	"brightblack":   ColorBrightBlack,
	"brightred":     ColorBrightRed,
	"brightgreen":   ColorBrightGreen,
	"brightyellow":  ColorBrightYellow,
	"brightblue":    ColorBrightBlue,
	"brightmagenta": ColorBrightMagenta,
	"brightcyan":    ColorBrightCyan,
	"brightwhite":   ColorBrightWhite,
	"bright-black":   ColorBrightBlack,
	"bright-red":     ColorBrightRed,
	"bright-green":   ColorBrightGreen,
	"bright-yellow":  ColorBrightYellow,
	"bright-blue":    ColorBrightBlue,
	"bright-magenta": ColorBrightMagenta,
	"bright-cyan":    ColorBrightCyan,
	"bright-white":   ColorBrightWhite,
}

func stripComments(src string) string {
	if src == "" {
		return src
	}
	out := make([]byte, 0, len(src))
	inLine := false
	inBlock := false
	for i := 0; i < len(src); i++ {
		next := byte(0)
		if i+1 < len(src) {
			next = src[i+1]
		}
		if inLine {
			if src[i] == '\n' {
				inLine = false
				out = append(out, src[i])
			}
			continue
		}
		if inBlock {
			if src[i] == '*' && next == '/' {
				inBlock = false
				i++
			}
			continue
		}
		if src[i] == '/' && next == '/' {
			inLine = true
			i++
			continue
		}
		if src[i] == '/' && next == '*' {
			inBlock = true
			i++
			continue
		}
		out = append(out, src[i])
	}
	return string(out)
}

func isIdentStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isIdentChar(ch byte) bool {
	return isIdentStart(ch) || (ch >= '0' && ch <= '9') || ch == '-'
}

func readIdent(s string, start int) (string, int) {
	idx := start
	for idx < len(s) && isIdentChar(s[idx]) {
		idx++
	}
	return s[start:idx], idx
}
