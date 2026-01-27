package markdown

import (
	"strings"

	"github.com/odvcencio/fluffy-ui/compositor"
	"github.com/odvcencio/fluffy-ui/theme"
	extast "github.com/yuin/goldmark/extension/ast"
)

// Box drawing characters for table borders
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

// RoundedBoxDrawings uses rounded corners for a softer look
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

// SharpBoxDrawings uses sharp corners for a technical look
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

// HeavyBoxDrawings uses heavy lines for emphasis
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

// DoubleBoxDrawings uses double lines
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

// TableAlignment represents cell alignment
type TableAlignment int

const (
	AlignLeft TableAlignment = iota
	AlignCenter
	AlignRight
)

// TableCell represents a single cell in a table
type TableCell struct {
	Text      string
	Alignment TableAlignment
	IsHeader  bool
	Width     int // Computed width
}

// TableRow represents a row in a table
type TableRow struct {
	Cells    []TableCell
	IsHeader bool
}

// EnhancedTable represents a parsed markdown table with layout info
type EnhancedTable struct {
	Rows       []TableRow
	Columns    int
	ColWidths  []int
	ColAligns  []TableAlignment
	TotalWidth int
}

// TableRendererConfig configures the table rendering
type TableRendererConfig struct {
	BoxDrawings    BoxDrawings
	HeaderStyle    compositor.Style
	CellStyle      compositor.Style
	BorderStyle    compositor.Style
	Padding        int
	MinColumnWidth int
}

// DefaultTableRendererConfig returns a default configuration
func DefaultTableRendererConfig(t *theme.Theme) TableRendererConfig {
	if t == nil {
		t = theme.DefaultTheme()
	}
	return TableRendererConfig{
		BoxDrawings:    RoundedBoxDrawings,
		HeaderStyle:    compositor.DefaultStyle().WithFG(t.TextPrimary.FG).WithBold(true),
		CellStyle:      compositor.DefaultStyle().WithFG(t.TextPrimary.FG),
		BorderStyle:    compositor.DefaultStyle().WithFG(t.Border.FG),
		Padding:        1,
		MinColumnWidth: 3,
	}
}

// ParseTable parses a goldmark AST table into an EnhancedTable
func ParseTable(table *extast.Table, source []byte) *EnhancedTable {
	et := &EnhancedTable{
		ColAligns: make([]TableAlignment, 0),
	}

	// Extract alignment from table
	for _, align := range table.Alignments {
		switch align {
		case extast.AlignCenter:
			et.ColAligns = append(et.ColAligns, AlignCenter)
		case extast.AlignRight:
			et.ColAligns = append(et.ColAligns, AlignRight)
		default:
			et.ColAligns = append(et.ColAligns, AlignLeft)
		}
	}

	for row := table.FirstChild(); row != nil; row = row.NextSibling() {
		tr := TableRow{}
		
		switch r := row.(type) {
		case *extast.TableHeader:
			tr.IsHeader = true
			for cell := r.FirstChild(); cell != nil; cell = cell.NextSibling() {
				if c, ok := cell.(*extast.TableCell); ok {
					text := strings.TrimSpace(collectPlainText(c, source))
					cellIdx := len(tr.Cells)
					tr.Cells = append(tr.Cells, TableCell{
						Text:      text,
						IsHeader:  true,
						Alignment: getAlignment(et.ColAligns, cellIdx),
					})
				}
			}
		case *extast.TableRow:
			for cell := r.FirstChild(); cell != nil; cell = cell.NextSibling() {
				if c, ok := cell.(*extast.TableCell); ok {
					text := strings.TrimSpace(collectPlainText(c, source))
					cellIdx := len(tr.Cells)
					tr.Cells = append(tr.Cells, TableCell{
						Text:      text,
						IsHeader:  false,
						Alignment: getAlignment(et.ColAligns, cellIdx),
					})
				}
			}
		}
		
		if len(tr.Cells) > 0 {
			et.Rows = append(et.Rows, tr)
			if len(tr.Cells) > et.Columns {
				et.Columns = len(tr.Cells)
			}
		}
	}

	// Ensure alignment array matches column count
	for len(et.ColAligns) < et.Columns {
		et.ColAligns = append(et.ColAligns, AlignLeft)
	}

	return et
}

func getAlignment(aligns []TableAlignment, idx int) TableAlignment {
	if idx >= 0 && idx < len(aligns) {
		return aligns[idx]
	}
	return AlignLeft
}

// CalculateWidths computes optimal column widths
func (t *EnhancedTable) CalculateWidths(config TableRendererConfig, maxWidth int) {
	if t.Columns == 0 {
		return
	}

	t.ColWidths = make([]int, t.Columns)

	// First pass: find minimum required width for each column
	for _, row := range t.Rows {
		for i, cell := range row.Cells {
			if i >= t.Columns {
				break
			}
			cellWidth := len(cell.Text)
			if cellWidth > t.ColWidths[i] {
				t.ColWidths[i] = cellWidth
			}
		}
	}

	// Apply minimum column width
	for i := range t.ColWidths {
		if t.ColWidths[i] < config.MinColumnWidth {
			t.ColWidths[i] = config.MinColumnWidth
		}
	}

	// Calculate total width including borders and padding
	borderWidth := 1 + t.Columns + 1 // Left border + separators + right border
	paddingWidth := t.Columns * config.Padding * 2
	contentWidth := 0
	for _, w := range t.ColWidths {
		contentWidth += w
	}
	t.TotalWidth = borderWidth + paddingWidth + contentWidth

	// If total exceeds maxWidth, we need to redistribute
	if maxWidth > 0 && t.TotalWidth > maxWidth {
		t.redistributeWidths(maxWidth, config)
	}
}

func (t *EnhancedTable) redistributeWidths(maxWidth int, config TableRendererConfig) {
	available := maxWidth - (1 + t.Columns + 1) - (t.Columns * config.Padding * 2)
	if available <= 0 {
		return
	}

	// Calculate total content width
	totalContent := 0
	for _, w := range t.ColWidths {
		totalContent += w
	}

	if totalContent <= available {
		return
	}

	// Redistribute proportionally
	scale := float64(available) / float64(totalContent)
	newTotal := 0
	for i := range t.ColWidths {
		newWidth := int(float64(t.ColWidths[i]) * scale)
		if newWidth < config.MinColumnWidth {
			newWidth = config.MinColumnWidth
		}
		t.ColWidths[i] = newWidth
		newTotal += newWidth
	}

	// Distribute any remaining space to last column
	if newTotal < available {
		t.ColWidths[t.Columns-1] += available - newTotal
	}

	t.TotalWidth = maxWidth
}

// Render renders the table to styled lines
func (t *EnhancedTable) Render(config TableRendererConfig) []StyledLine {
	if len(t.Rows) == 0 {
		return nil
	}

	var lines []StyledLine
	box := config.BoxDrawings

	// Top border
	lines = append(lines, t.renderBorder(box.TopLeft, box.TopT, box.TopRight, config))

	// Render rows
	for i, row := range t.Rows {
		lines = append(lines, t.renderRow(row, config))
		
		// Add separator after header
		if row.IsHeader && i < len(t.Rows)-1 {
			lines = append(lines, t.renderBorder(box.LeftT, box.Cross, box.RightT, config))
		}
	}

	// Bottom border
	lines = append(lines, t.renderBorder(box.BottomLeft, box.BottomT, box.BottomRight, config))

	return lines
}

func (t *EnhancedTable) renderBorder(left, middle, right string, config TableRendererConfig) StyledLine {
	var spans []StyledSpan
	borderStyle := config.BorderStyle

	spans = append(spans, StyledSpan{Text: left, Style: borderStyle})
	
	for i, width := range t.ColWidths {
		padding := strings.Repeat(config.BoxDrawings.Horizontal, config.Padding*2+width)
		spans = append(spans, StyledSpan{Text: padding, Style: borderStyle})
		
		if i < len(t.ColWidths)-1 {
			spans = append(spans, StyledSpan{Text: middle, Style: borderStyle})
		}
	}
	
	spans = append(spans, StyledSpan{Text: right, Style: borderStyle})
	
	return StyledLine{Spans: spans}
}

func (t *EnhancedTable) renderRow(row TableRow, config TableRendererConfig) StyledLine {
	var spans []StyledSpan
	box := config.BoxDrawings

	spans = append(spans, StyledSpan{Text: box.Vertical, Style: config.BorderStyle})

	for i, cell := range row.Cells {
		if i >= t.Columns {
			break
		}

		width := t.ColWidths[i]
		if i < len(t.ColWidths) {
			width = t.ColWidths[i]
		}

		// Get style based on cell type
		style := config.CellStyle
		if cell.IsHeader || row.IsHeader {
			style = config.HeaderStyle
		}

		// Pad content according to alignment
		content := t.alignText(cell.Text, width, cell.Alignment)
		
		// Add padding spaces
		padding := strings.Repeat(" ", config.Padding)
		spans = append(spans, StyledSpan{Text: padding, Style: config.CellStyle})
		spans = append(spans, StyledSpan{Text: content, Style: style})
		spans = append(spans, StyledSpan{Text: padding, Style: config.CellStyle})
		
		spans = append(spans, StyledSpan{Text: box.Vertical, Style: config.BorderStyle})
	}

	// Fill empty cells if row has fewer cells than columns
	for i := len(row.Cells); i < t.Columns; i++ {
		width := t.ColWidths[i]
		padding := strings.Repeat(" ", config.Padding*2+width)
		spans = append(spans, StyledSpan{Text: padding, Style: config.CellStyle})
		spans = append(spans, StyledSpan{Text: box.Vertical, Style: config.BorderStyle})
	}

	return StyledLine{Spans: spans}
}

func (t *EnhancedTable) alignText(text string, width int, align TableAlignment) string {
	textLen := len(text)
	if textLen >= width {
		return text[:width]
	}

	switch align {
	case AlignCenter:
		leftPad := (width - textLen) / 2
		rightPad := width - textLen - leftPad
		return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
	case AlignRight:
		return strings.Repeat(" ", width-textLen) + text
	default: // AlignLeft
		return text + strings.Repeat(" ", width-textLen)
	}
}

// EnhancedTableRenderer handles rich table rendering
type EnhancedTableRenderer struct {
	config TableRendererConfig
}

// NewEnhancedTableRenderer creates a new table renderer
func NewEnhancedTableRenderer(config TableRendererConfig) *EnhancedTableRenderer {
	return &EnhancedTableRenderer{config: config}
}

// RenderTable renders a markdown table with enhanced formatting
func (r *EnhancedTableRenderer) RenderTable(table *extast.Table, source []byte, maxWidth int) []StyledLine {
	et := ParseTable(table, source)
	if et.Columns == 0 {
		return nil
	}
	
	et.CalculateWidths(r.config, maxWidth)
	return et.Render(r.config)
}

// Alternative styles for different contexts
func TableStyleCompact(t *theme.Theme) TableRendererConfig {
	cfg := DefaultTableRendererConfig(t)
	cfg.Padding = 0
	cfg.BoxDrawings = BoxDrawings{
		Horizontal: " ",
		Vertical:   "│",
		TopLeft:    "", TopRight: "", BottomLeft: "", BottomRight: "",
		LeftT: "", RightT: "", TopT: "", BottomT: "", Cross: "",
	}
	return cfg
}

func TableStyleMinimal(t *theme.Theme) TableRendererConfig {
	cfg := DefaultTableRendererConfig(t)
	cfg.Padding = 1
	cfg.BoxDrawings = BoxDrawings{
		Horizontal: "─",
		Vertical:   "│",
		TopLeft:    "┌", TopRight: "┐",
		BottomLeft: "└", BottomRight: "┘",
		LeftT: "├", RightT: "┤", TopT: "┬", BottomT: "┴", Cross: "┼",
	}
	return cfg
}

func TableStyleHeavy(t *theme.Theme) TableRendererConfig {
	cfg := DefaultTableRendererConfig(t)
	cfg.BoxDrawings = HeavyBoxDrawings
	return cfg
}
