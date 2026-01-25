package fur

import (
	"fmt"
	"html"
	"strings"

	"github.com/odvcencio/fluffy-ui/compositor"
)

// ExportFormat identifies output formats.
type ExportFormat string

const (
	ExportTextFormat ExportFormat = "text"
	ExportHTMLFormat ExportFormat = "html"
	ExportSVGFormat  ExportFormat = "svg"
)

// Export renders a renderable to the requested format.
func Export(r Renderable, width int, format ExportFormat) string {
	switch format {
	case ExportHTMLFormat:
		return ExportHTML(r, width)
	case ExportSVGFormat:
		return ExportSVG(r, width)
	default:
		return ExportText(r, width)
	}
}

// ExportText renders plain text output.
func ExportText(r Renderable, width int) string {
	if r == nil {
		return ""
	}
	lines := r.Render(width)
	var out strings.Builder
	for i, line := range lines {
		for _, span := range line {
			out.WriteString(span.Text)
		}
		if i < len(lines)-1 {
			out.WriteByte('\n')
		}
	}
	return out.String()
}

// ExportHTML renders HTML output with inline styles.
func ExportHTML(r Renderable, width int) string {
	if r == nil {
		return ""
	}
	lines := r.Render(width)
	var out strings.Builder
	out.WriteString("<pre style=\"font-family: ui-monospace, Menlo, Consolas, monospace; font-size: 14px; line-height: 1.4;\">\n")
	for i, line := range lines {
		for _, span := range line {
			style := styleToCSS(span.Style)
			text := html.EscapeString(span.Text)
			if style == "" {
				out.WriteString(text)
				continue
			}
			out.WriteString("<span style=\"")
			out.WriteString(style)
			out.WriteString("\">")
			out.WriteString(text)
			out.WriteString("</span>")
		}
		if i < len(lines)-1 {
			out.WriteByte('\n')
		}
	}
	out.WriteString("\n</pre>")
	return out.String()
}

// ExportSVG renders SVG output with inline styles.
func ExportSVG(r Renderable, width int) string {
	if r == nil {
		return ""
	}
	lines := r.Render(width)
	maxWidth := 0
	for _, line := range lines {
		if w := lineWidth(line); w > maxWidth {
			maxWidth = w
		}
	}
	charWidth := 8
	lineHeight := 16
	svgWidth := maxWidth * charWidth
	svgHeight := len(lines) * lineHeight
	if svgWidth <= 0 {
		svgWidth = charWidth
	}
	if svgHeight <= 0 {
		svgHeight = lineHeight
	}
	var out strings.Builder
	out.WriteString(fmt.Sprintf("<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"%d\" height=\"%d\">", svgWidth, svgHeight))
	out.WriteString("<style>text{font-family: ui-monospace, Menlo, Consolas, monospace; font-size: 14px;}</style>")
	for i, line := range lines {
		y := (i + 1) * lineHeight
		out.WriteString(fmt.Sprintf("<text x=\"0\" y=\"%d\" xml:space=\"preserve\">", y))
		for _, span := range line {
			style := styleToSVG(span.Style)
			text := html.EscapeString(span.Text)
			if style == "" {
				out.WriteString(text)
				continue
			}
			out.WriteString("<tspan ")
			out.WriteString(style)
			out.WriteString(">")
			out.WriteString(text)
			out.WriteString("</tspan>")
		}
		out.WriteString("</text>")
	}
	out.WriteString("</svg>")
	return out.String()
}

func styleToCSS(style Style) string {
	fg := style.fg
	bg := style.bg
	if style.reverse {
		fg, bg = bg, fg
	}
	var parts []string
	if color, ok := colorToHex(fg); ok {
		parts = append(parts, "color: "+color)
	}
	if color, ok := colorToHex(bg); ok {
		parts = append(parts, "background-color: "+color)
	}
	if style.bold {
		parts = append(parts, "font-weight: 700")
	}
	if style.italic {
		parts = append(parts, "font-style: italic")
	}
	if style.underline || style.strikethrough {
		decorations := []string{}
		if style.underline {
			decorations = append(decorations, "underline")
		}
		if style.strikethrough {
			decorations = append(decorations, "line-through")
		}
		parts = append(parts, "text-decoration: "+strings.Join(decorations, " "))
	}
	if style.dim {
		parts = append(parts, "opacity: 0.6")
	}
	return strings.Join(parts, "; ")
}

func styleToSVG(style Style) string {
	fg := style.fg
	if style.reverse {
		fg = style.bg
	}
	var attrs []string
	if color, ok := colorToHex(fg); ok {
		attrs = append(attrs, fmt.Sprintf("fill=\"%s\"", color))
	}
	if style.bold {
		attrs = append(attrs, "font-weight=\"700\"")
	}
	if style.italic {
		attrs = append(attrs, "font-style=\"italic\"")
	}
	if style.underline || style.strikethrough {
		decorations := []string{}
		if style.underline {
			decorations = append(decorations, "underline")
		}
		if style.strikethrough {
			decorations = append(decorations, "line-through")
		}
		attrs = append(attrs, fmt.Sprintf("text-decoration=\"%s\"", strings.Join(decorations, " ")))
	}
	if style.dim {
		attrs = append(attrs, "opacity=\"0.6\"")
	}
	return strings.Join(attrs, " ")
}

func colorToHex(color Color) (string, bool) {
	r, g, b, ok := colorToRGB(color)
	if !ok {
		return "", false
	}
	return fmt.Sprintf("#%02x%02x%02x", r, g, b), true
}

func colorToRGB(color Color) (uint8, uint8, uint8, bool) {
	switch color.Mode {
	case compositor.ColorModeRGB:
		r := uint8((color.Value >> 16) & 0xFF)
		g := uint8((color.Value >> 8) & 0xFF)
		b := uint8(color.Value & 0xFF)
		return r, g, b, true
	case compositor.ColorMode16:
		r, g, b := ansi16ToRGB(uint8(color.Value))
		return r, g, b, true
	case compositor.ColorMode256:
		r, g, b := ansi256ToRGB(uint8(color.Value))
		return r, g, b, true
	default:
		return 0, 0, 0, false
	}
}

func ansi16ToRGB(index uint8) (uint8, uint8, uint8) {
	palette := [16][3]uint8{
		{0, 0, 0},     // black
		{205, 0, 0},   // red
		{0, 205, 0},   // green
		{205, 205, 0}, // yellow
		{0, 0, 238},   // blue
		{205, 0, 205}, // magenta
		{0, 205, 205}, // cyan
		{229, 229, 229},
		{127, 127, 127},
		{255, 0, 0},
		{0, 255, 0},
		{255, 255, 0},
		{92, 92, 255},
		{255, 0, 255},
		{0, 255, 255},
		{255, 255, 255},
	}
	if index > 15 {
		index = 15
	}
	entry := palette[index]
	return entry[0], entry[1], entry[2]
}

func ansi256ToRGB(index uint8) (uint8, uint8, uint8) {
	if index < 16 {
		return ansi16ToRGB(index)
	}
	if index >= 232 {
		value := uint8(8 + (index-232)*10)
		return value, value, value
	}
	index -= 16
	r := index / 36
	g := (index / 6) % 6
	b := index % 6
	levels := []uint8{0, 95, 135, 175, 215, 255}
	return levels[r], levels[g], levels[b]
}
