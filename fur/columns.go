package fur

import "strings"

// ColumnsOpts configures column layout.
type ColumnsOpts struct {
	Width   int
	Padding int
	Equal   bool
	Expand  bool
}

// Columns arranges renderables horizontally.
func Columns(items ...Renderable) Renderable {
	return ColumnsWith(items, ColumnsOpts{Padding: 2})
}

// ColumnsWith arranges renderables horizontally with options.
func ColumnsWith(items []Renderable, opts ColumnsOpts) Renderable {
	if opts.Padding <= 0 {
		opts.Padding = 2
	}
	return columnsRenderable{items: items, opts: opts}
}

type columnsRenderable struct {
	items []Renderable
	opts  ColumnsOpts
}

func (c columnsRenderable) Render(width int) []Line {
	if len(c.items) == 0 {
		return nil
	}
	cols := len(c.items)
	padding := c.opts.Padding
	colWidths := make([]int, cols)

	if c.opts.Width > 0 {
		for i := range colWidths {
			colWidths[i] = c.opts.Width
		}
	} else {
		natural := make([]int, cols)
		for i, item := range c.items {
			if item == nil {
				continue
			}
			lines := item.Render(0)
			maxWidth := 0
			for _, line := range lines {
				if w := lineWidth(line); w > maxWidth {
					maxWidth = w
				}
			}
			natural[i] = maxWidth
		}
		if c.opts.Equal {
			maxWidth := 0
			for _, w := range natural {
				if w > maxWidth {
					maxWidth = w
				}
			}
			for i := range colWidths {
				colWidths[i] = maxWidth
			}
		} else {
			copy(colWidths, natural)
		}
		if width > 0 {
			available := width - padding*(cols-1)
			if available < cols {
				available = cols
			}
			total := 0
			for _, w := range colWidths {
				total += w
			}
			if total < available && c.opts.Expand {
				extra := available - total
				for extra > 0 {
					for i := 0; i < cols && extra > 0; i++ {
						colWidths[i]++
						extra--
					}
				}
				total = available
			}
			if total > available {
				shrink := total - available
				for shrink > 0 {
					idx := widestColumn(colWidths)
					if idx < 0 || colWidths[idx] <= 1 {
						break
					}
					colWidths[idx]--
					shrink--
				}
			}
		}
	}

	columnLines := make([][]Line, cols)
	maxLines := 0
	for i, item := range c.items {
		if item == nil {
			continue
		}
		colWidth := colWidths[i]
		columnLines[i] = item.Render(colWidth)
		if len(columnLines[i]) > maxLines {
			maxLines = len(columnLines[i])
		}
	}

	padSpan := Span{Text: repeatSpaces(padding), Style: DefaultStyle()}
	var out []Line
	for row := 0; row < maxLines; row++ {
		var line Line
		for col := 0; col < cols; col++ {
			var cell Line
			if row < len(columnLines[col]) {
				cell = columnLines[col][row]
			}
			cell = padLine(cell, colWidths[col])
			line = append(line, cell...)
			if col < cols-1 && padding > 0 {
				line = append(line, padSpan)
			}
		}
		out = append(out, line)
	}
	return out
}

func widestColumn(widths []int) int {
	idx := -1
	maxWidth := -1
	for i, w := range widths {
		if w > maxWidth {
			maxWidth = w
			idx = i
		}
	}
	return idx
}

func repeatSpaces(count int) string {
	if count <= 0 {
		return ""
	}
	return strings.Repeat(" ", count)
}
