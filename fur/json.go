package fur

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// JSON renders JSON data with syntax highlighting.
func JSON(data string) Renderable {
	return jsonRenderable{data: data, indent: 2}
}

// JSONFromValue renders a Go value as formatted JSON.
func JSONFromValue(v any) Renderable {
	return jsonRenderable{value: v, indent: 2}
}

type jsonRenderable struct {
	data   string
	value  any
	indent int
	width  int
}

// WithIndent sets the indentation level.
func (j jsonRenderable) WithIndent(indent int) jsonRenderable {
	j.indent = indent
	return j
}

// WithWidth sets the maximum width for wrapping.
func (j jsonRenderable) WithWidth(width int) jsonRenderable {
	j.width = width
	return j
}

func (j jsonRenderable) Render(width int) []Line {
	if j.width > 0 {
		width = j.width
	}
	
	var data any
	if j.value != nil {
		data = j.value
	} else {
		if err := json.Unmarshal([]byte(j.data), &data); err != nil {
			return []Line{{{Text: "Invalid JSON: " + err.Error(), Style: Style{}.Foreground(ColorRed)}}}
		}
	}
	
	lines := j.formatValue(data, 0)
	return wrapLines(lines, width)
}

func (j jsonRenderable) formatValue(v any, depth int) []Line {
	switch val := v.(type) {
	case map[string]any:
		return j.formatObject(val, depth)
	case []any:
		return j.formatArray(val, depth)
	case string:
		return []Line{{{Text: strconv.Quote(val), Style: Style{}.Foreground(ColorGreen)}}}
	case float64:
		return []Line{{{Text: formatJSONNumber(val), Style: Style{}.Foreground(ColorYellow)}}}
	case bool:
		return []Line{{{Text: strconv.FormatBool(val), Style: Style{}.Foreground(ColorMagenta)}}}
	case nil:
		return []Line{{{Text: "null", Style: Style{}.Foreground(ColorBrightBlack)}}}
	default:
		return []Line{{{Text: fmt.Sprintf("%v", val), Style: DefaultStyle()}}}
	}
}

func (j jsonRenderable) formatObject(obj map[string]any, depth int) []Line {
	if len(obj) == 0 {
		return []Line{{{Text: "{}", Style: DefaultStyle()}}}
	}
	
	var lines []Line
	indent := strings.Repeat(" ", depth*j.indent)
	innerIndent := strings.Repeat(" ", (depth+1)*j.indent)
	
	lines = append(lines, Line{{Text: indent + "{", Style: DefaultStyle()}})
	
	// Sort keys for consistent output
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	for i, key := range keys {
		val := obj[key]
		comma := ","
		if i == len(keys)-1 {
			comma = ""
		}
		
		keySpan := Span{Text: strconv.Quote(key), Style: Style{}.Foreground(ColorCyan)}
		colonSpan := Span{Text: ": ", Style: DefaultStyle()}
		
		valLines := j.formatValue(val, depth+1)
		if len(valLines) == 0 {
			valLines = []Line{{{Text: "null", Style: Style{}.Foreground(ColorBrightBlack)}}}
		}
		
		// First line with key
		firstLine := Line{keySpan, colonSpan}
		firstLine = append(firstLine, valLines[0]...)
		firstLine = append(firstLine, Span{Text: comma, Style: DefaultStyle()})
		lines = append(lines, j.indentLine(firstLine, innerIndent))
		
		// Remaining lines
		for _, vl := range valLines[1:] {
			lines = append(lines, j.indentLine(vl, innerIndent))
		}
	}
	
	lines = append(lines, Line{{Text: indent + "}", Style: DefaultStyle()}})
	return lines
}

func (j jsonRenderable) formatArray(arr []any, depth int) []Line {
	if len(arr) == 0 {
		return []Line{{{Text: "[]", Style: DefaultStyle()}}}
	}
	
	var lines []Line
	indent := strings.Repeat(" ", depth*j.indent)
	innerIndent := strings.Repeat(" ", (depth+1)*j.indent)
	
	lines = append(lines, Line{{Text: indent + "[", Style: DefaultStyle()}})
	
	for i, val := range arr {
		comma := ","
		if i == len(arr)-1 {
			comma = ""
		}
		
		valLines := j.formatValue(val, depth+1)
		if len(valLines) == 0 {
			valLines = []Line{{{Text: "null", Style: Style{}.Foreground(ColorBrightBlack)}}}
		}
		
		// First line
		firstLine := append(Line{}, valLines[0]...)
		firstLine = append(firstLine, Span{Text: comma, Style: DefaultStyle()})
		lines = append(lines, j.indentLine(firstLine, innerIndent))
		
		// Remaining lines
		for _, vl := range valLines[1:] {
			lines = append(lines, j.indentLine(vl, innerIndent))
		}
	}
	
	lines = append(lines, Line{{Text: indent + "]", Style: DefaultStyle()}})
	return lines
}

func (j jsonRenderable) indentLine(line Line, indent string) Line {
	if len(line) == 0 {
		return Line{{Text: indent, Style: DefaultStyle()}}
	}
	// Prepend indent to first span
	return append(Line{{Text: indent, Style: DefaultStyle()}}, line...)
}

func formatJSONNumber(n float64) string {
	if n == float64(int64(n)) {
		return fmt.Sprintf("%.0f", n)
	}
	return strconv.FormatFloat(n, 'f', -1, 64)
}

// JSONCompact renders JSON in a compact single-line format.
func JSONCompact(data string) Renderable {
	return jsonCompactRenderable{data: data}
}

type jsonCompactRenderable struct {
	data string
}

func (j jsonCompactRenderable) Render(width int) []Line {
	var v any
	if err := json.Unmarshal([]byte(j.data), &v); err != nil {
		return []Line{{{Text: "Invalid JSON: " + err.Error(), Style: Style{}.Foreground(ColorRed)}}}
	}
	
	compact, err := json.Marshal(v)
	if err != nil {
		return []Line{{{Text: "Error: " + err.Error(), Style: Style{}.Foreground(ColorRed)}}}
	}
	
	// Apply syntax highlighting to compact JSON
	return JSON(string(compact)).Render(width)
}
