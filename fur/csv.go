package fur

import (
	"encoding/csv"
	"strings"
)

// CSVTable renders CSV data as a formatted table.
func CSVTable(data string, hasHeader bool) Renderable {
	return csvTableRenderable{
		data:      data,
		hasHeader: hasHeader,
	}
}

// CSVTableFromRecords renders pre-parsed CSV records.
func CSVTableFromRecords(records [][]string, hasHeader bool) Renderable {
	return csvTableRenderable{
		records:   records,
		hasHeader: hasHeader,
	}
}

type csvTableRenderable struct {
	data      string
	records   [][]string
	hasHeader bool
	style     TableStyle
}

// TableStyle configures table appearance.
type TableStyle struct {
	BoxDrawings   BoxDrawings
	HeaderStyle   Style
	CellStyle     Style
	BorderStyle   Style
	Padding       int
	MinColWidth   int
}

// DefaultTableStyle returns the default table style.
func DefaultTableStyle() TableStyle {
	return TableStyle{
		BoxDrawings: RoundedBoxDrawings,
		HeaderStyle: Style{}.Bold(),
		CellStyle:   DefaultStyle(),
		BorderStyle: Style{}.Foreground(ColorBrightBlack),
		Padding:     1,
		MinColWidth: 3,
	}
}

// CompactTableStyle returns a compact table style without outer borders.
func CompactTableStyle() TableStyle {
	return TableStyle{
		BoxDrawings: BoxDrawings{
			Horizontal:       " ",
			Vertical:         "│",
			TopLeft:          "", TopRight: "", BottomLeft: "", BottomRight: "",
			LeftT: "", RightT: "", TopT: "", BottomT: "", Cross: "",
		},
		HeaderStyle: Style{}.Bold().Underline(),
		CellStyle:   DefaultStyle(),
		BorderStyle: Style{}.Foreground(ColorBrightBlack),
		Padding:     1,
		MinColWidth: 1,
	}
}

// WithStyle sets the table style.
func (c csvTableRenderable) WithStyle(style TableStyle) csvTableRenderable {
	c.style = style
	return c
}

func (c csvTableRenderable) Render(width int) []Line {
	style := c.style
	if style.BoxDrawings.Horizontal == "" {
		style = DefaultTableStyle()
	}
	
	records := c.records
	if records == nil {
		var err error
		records, err = c.parseCSV()
		if err != nil {
			return []Line{{{Text: "Error parsing CSV: " + err.Error(), Style: Style{}.Foreground(ColorRed)}}}
		}
	}
	
	if len(records) == 0 {
		return nil
	}
	
	// Calculate column widths
	colCount := 0
	for _, rec := range records {
		if len(rec) > colCount {
			colCount = len(rec)
		}
	}
	
	colWidths := make([]int, colCount)
	for _, rec := range records {
		for i, cell := range rec {
			if i >= colCount {
				break
			}
			cellWidth := stringWidth(cell)
			if cellWidth > colWidths[i] {
				colWidths[i] = cellWidth
			}
		}
	}
	
	// Apply minimum column width
	for i := range colWidths {
		if colWidths[i] < style.MinColWidth {
			colWidths[i] = style.MinColWidth
		}
	}
	
	// Check if we need to constrain widths
	totalWidth := 2 + (colCount-1) + (colCount * style.Padding * 2) // borders
	for _, w := range colWidths {
		totalWidth += w
	}
	
	if width > 0 && totalWidth > width {
		// Redistribute widths proportionally
		available := width - 2 - (colCount-1) - (colCount * style.Padding * 2)
		if available > 0 {
			totalContent := 0
			for _, w := range colWidths {
				totalContent += w
			}
			scale := float64(available) / float64(totalContent)
			for i := range colWidths {
				newWidth := int(float64(colWidths[i]) * scale)
				if newWidth < style.MinColWidth {
					newWidth = style.MinColWidth
				}
				colWidths[i] = newWidth
			}
		}
	}
	
	return c.renderTable(records, colWidths, style)
}

func (c csvTableRenderable) parseCSV() ([][]string, error) {
	reader := csv.NewReader(strings.NewReader(c.data))
	reader.FieldsPerRecord = -1 // Allow variable fields per record
	return reader.ReadAll()
}

func (c csvTableRenderable) renderTable(records [][]string, colWidths []int, style TableStyle) []Line {
	var lines []Line
	box := style.BoxDrawings
	hasOuterBorders := box.TopLeft != ""
	
	// Top border
	if hasOuterBorders {
		lines = append(lines, c.renderBorder(box.TopLeft, box.TopT, box.TopRight, colWidths, style))
	}
	
	// Render rows
	for i, rec := range records {
		isHeader := c.hasHeader && i == 0
		lines = append(lines, c.renderRow(rec, colWidths, style, isHeader))
		
		// Add separator after header
		if isHeader && i < len(records)-1 {
			if hasOuterBorders {
				lines = append(lines, c.renderBorder(box.LeftT, box.Cross, box.RightT, colWidths, style))
			} else {
				// Simple separator for compact style
				lines = append(lines, c.renderSimpleSeparator(colWidths, style))
			}
		}
	}
	
	// Bottom border
	if hasOuterBorders {
		lines = append(lines, c.renderBorder(box.BottomLeft, box.BottomT, box.BottomRight, colWidths, style))
	}
	
	return lines
}

func (c csvTableRenderable) renderBorder(left, middle, right string, colWidths []int, style TableStyle) Line {
	var spans []Span
	borderStyle := style.BorderStyle
	
	spans = append(spans, Span{Text: left, Style: borderStyle})
	
	for i, width := range colWidths {
		padding := strings.Repeat(style.BoxDrawings.Horizontal, style.Padding*2+width)
		spans = append(spans, Span{Text: padding, Style: borderStyle})
		
		if i < len(colWidths)-1 {
			spans = append(spans, Span{Text: middle, Style: borderStyle})
		}
	}
	
	spans = append(spans, Span{Text: right, Style: borderStyle})
	
	return spans
}

func (c csvTableRenderable) renderSimpleSeparator(colWidths []int, style TableStyle) Line {
	var spans []Span
	borderStyle := style.BorderStyle
	
	spans = append(spans, Span{Text: style.BoxDrawings.Vertical, Style: borderStyle})
	
	for i, width := range colWidths {
		sepWidth := style.Padding*2 + width
		// Use dashes or just spaces for compact style
		sep := strings.Repeat("-", sepWidth)
		spans = append(spans, Span{Text: sep, Style: borderStyle})
		
		if i < len(colWidths)-1 {
			spans = append(spans, Span{Text: "+", Style: borderStyle})
		}
	}
	
	spans = append(spans, Span{Text: style.BoxDrawings.Vertical, Style: borderStyle})
	
	return spans
}

func (c csvTableRenderable) renderRow(rec []string, colWidths []int, style TableStyle, isHeader bool) Line {
	var spans []Span
	box := style.BoxDrawings
	
	if box.Vertical != "" {
		spans = append(spans, Span{Text: box.Vertical, Style: style.BorderStyle})
	}
	
	for i, width := range colWidths {
		cell := ""
		if i < len(rec) {
			cell = rec[i]
		}
		
		cellStyle := style.CellStyle
		if isHeader {
			cellStyle = style.HeaderStyle
		}
		
		// Pad content
		cellWidth := stringWidth(cell)
		if cellWidth > width {
			cell = truncateString(cell, width)
			cellWidth = width
		}
		paddingRight := width - cellWidth
		
		// Left padding
		if style.Padding > 0 {
			spans = append(spans, Span{Text: strings.Repeat(" ", style.Padding), Style: style.CellStyle})
		}
		
		// Cell content
		spans = append(spans, Span{Text: cell, Style: cellStyle})
		
		// Right padding
		if style.Padding > 0 || paddingRight > 0 {
			spans = append(spans, Span{Text: strings.Repeat(" ", style.Padding+paddingRight), Style: style.CellStyle})
		}
		
		if box.Vertical != "" {
			spans = append(spans, Span{Text: box.Vertical, Style: style.BorderStyle})
		}
	}
	
	// Fill empty cells
	for i := len(rec); i < len(colWidths); i++ {
		width := colWidths[i]
		padding := strings.Repeat(" ", style.Padding*2+width)
		spans = append(spans, Span{Text: padding, Style: style.CellStyle})
		if box.Vertical != "" {
			spans = append(spans, Span{Text: box.Vertical, Style: style.BorderStyle})
		}
	}
	
	return spans
}

// BoxDrawings defines Unicode box drawing characters.
type BoxDrawings struct {
	Horizontal       string
	Vertical         string
	TopLeft          string
	TopRight         string
	BottomLeft       string
	BottomRight      string
	LeftT            string
	RightT           string
	TopT             string
	BottomT          string
	Cross            string
}

// RoundedBoxDrawings uses rounded corners.
var RoundedBoxDrawings = BoxDrawings{
	Horizontal:  "─",
	Vertical:    "│",
	TopLeft:     "╭",
	TopRight:    "╮",
	BottomLeft:  "╰",
	BottomRight: "╯",
	LeftT:       "├",
	RightT:      "┤",
	TopT:        "┬",
	BottomT:     "┴",
	Cross:       "┼",
}

// SharpBoxDrawings uses sharp corners.
var SharpBoxDrawings = BoxDrawings{
	Horizontal:  "─",
	Vertical:    "│",
	TopLeft:     "┌",
	TopRight:    "┐",
	BottomLeft:  "└",
	BottomRight: "┘",
	LeftT:       "├",
	RightT:      "┤",
	TopT:        "┬",
	BottomT:     "┴",
	Cross:       "┼",
}

// HeavyBoxDrawings uses heavy lines.
var HeavyBoxDrawings = BoxDrawings{
	Horizontal:  "━",
	Vertical:    "┃",
	TopLeft:     "┏",
	TopRight:    "┓",
	BottomLeft:  "┗",
	BottomRight: "┛",
	LeftT:       "┣",
	RightT:      "┫",
	TopT:        "┳",
	BottomT:     "┻",
	Cross:       "╋",
}

// DoubleBoxDrawings uses double lines.
var DoubleBoxDrawings = BoxDrawings{
	Horizontal:  "═",
	Vertical:    "║",
	TopLeft:     "╔",
	TopRight:    "╗",
	BottomLeft:  "╚",
	BottomRight: "╝",
	LeftT:       "╠",
	RightT:      "╣",
	TopT:        "╦",
	BottomT:     "╩",
	Cross:       "╬",
}
