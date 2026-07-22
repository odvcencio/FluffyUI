package markdown

import (
	"strings"

	"github.com/mattn/go-runewidth"
	extast "github.com/yuin/goldmark/extension/ast"
	"m31labs.dev/fluffyui/compositor"
	"m31labs.dev/fluffyui/theme"
)

// Box drawing characters for table borders
type BoxDrawings struct {
	Horizontal  string
	Vertical    string
	TopLeft     string
	TopRight    string
	BottomLeft  string
	BottomRight string
	LeftT       string
	RightT      string
	TopT        string
	BottomT     string
	Cross       string
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
			cellWidth := runewidth.StringWidth(cell.Text)
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

	minWidth := config.MinColumnWidth
	if minWidth < 1 {
		minWidth = 1
	}
	if available < minWidth*t.Columns {
		minWidth = 1
	}

	// Shrink the widest columns first. This guarantees that the emitted border
	// and row widths agree with TotalWidth instead of merely recording maxWidth
	// after minimum-width clamping has made the table wider than the viewport.
	for totalContent > available {
		widest := -1
		for i, width := range t.ColWidths {
			if width > minWidth && (widest < 0 || width > t.ColWidths[widest]) {
				widest = i
			}
		}
		if widest < 0 {
			break
		}
		t.ColWidths[widest]--
		totalContent--
	}

	t.TotalWidth = (1 + t.Columns + 1) + (t.Columns * config.Padding * 2) + totalContent
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
		lines = append(lines, t.renderRows(row, config)...)

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

func (t *EnhancedTable) renderRows(row TableRow, config TableRendererConfig) []StyledLine {
	wrapped := make([][]string, t.Columns)
	height := 1
	for column := 0; column < t.Columns; column++ {
		text := ""
		if column < len(row.Cells) {
			text = row.Cells[column].Text
		}
		wrapped[column] = wrapTableCell(text, t.ColWidths[column])
		if len(wrapped[column]) > height {
			height = len(wrapped[column])
		}
	}

	lines := make([]StyledLine, 0, height)
	for lineIndex := 0; lineIndex < height; lineIndex++ {
		lines = append(lines, t.renderRowLine(row, wrapped, lineIndex, config))
	}
	return lines
}

func (t *EnhancedTable) renderRowLine(row TableRow, wrapped [][]string, lineIndex int, config TableRendererConfig) StyledLine {
	spans := []StyledSpan{{Text: config.BoxDrawings.Vertical, Style: config.BorderStyle}}
	padding := strings.Repeat(" ", config.Padding)
	for column := 0; column < t.Columns; column++ {
		cell := TableCell{Alignment: getAlignment(t.ColAligns, column)}
		if column < len(row.Cells) {
			cell = row.Cells[column]
		}
		style := config.CellStyle
		if cell.IsHeader || row.IsHeader {
			style = config.HeaderStyle
		}
		content := ""
		if lineIndex < len(wrapped[column]) {
			content = wrapped[column][lineIndex]
		}
		content = t.alignText(content, t.ColWidths[column], cell.Alignment)
		spans = append(spans,
			StyledSpan{Text: padding, Style: config.CellStyle},
			StyledSpan{Text: content, Style: style},
			StyledSpan{Text: padding, Style: config.CellStyle},
			StyledSpan{Text: config.BoxDrawings.Vertical, Style: config.BorderStyle},
		)
	}
	return StyledLine{Spans: spans}
}

func wrapTableCell(text string, width int) []string {
	if width <= 0 || text == "" {
		return []string{""}
	}
	if runewidth.StringWidth(text) <= width {
		return []string{text}
	}

	var lines []string
	var line strings.Builder
	lineWidth := 0
	flush := func() {
		if line.Len() > 0 {
			lines = append(lines, line.String())
			line.Reset()
			lineWidth = 0
		}
	}
	for _, word := range strings.Fields(text) {
		wordWidth := runewidth.StringWidth(word)
		if lineWidth > 0 && lineWidth+1+wordWidth <= width {
			line.WriteByte(' ')
			line.WriteString(word)
			lineWidth += 1 + wordWidth
			continue
		}
		if lineWidth > 0 {
			flush()
		}
		if wordWidth <= width {
			line.WriteString(word)
			lineWidth = wordWidth
			continue
		}
		chunks := splitTableWord(word, width)
		for i, chunk := range chunks {
			if i < len(chunks)-1 || runewidth.StringWidth(chunk) == width {
				lines = append(lines, chunk)
			} else {
				line.WriteString(chunk)
				lineWidth = runewidth.StringWidth(chunk)
			}
		}
	}
	flush()
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}

func splitTableWord(word string, width int) []string {
	var chunks []string
	var chunk strings.Builder
	chunkWidth := 0
	for _, r := range word {
		runeWidth := runewidth.RuneWidth(r)
		if runeWidth < 0 {
			runeWidth = 0
		}
		if chunkWidth > 0 && chunkWidth+runeWidth > width {
			chunks = append(chunks, chunk.String())
			chunk.Reset()
			chunkWidth = 0
		}
		chunk.WriteRune(r)
		chunkWidth += runeWidth
	}
	if chunk.Len() > 0 {
		chunks = append(chunks, chunk.String())
	}
	return chunks
}

func (t *EnhancedTable) alignText(text string, width int, align TableAlignment) string {
	textLen := runewidth.StringWidth(text)
	if textLen >= width {
		return runewidth.Truncate(text, width, "")
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
