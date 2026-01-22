package widgets

import (
	"fmt"
	"strings"

	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/scroll"
	"github.com/odvcencio/fluffy-ui/terminal"
)

// TableColumn defines a column in a table.
type TableColumn struct {
	Title string
	Width int
}

// Table is a simple data grid widget.
type Table struct {
	FocusableBase
	Columns       []TableColumn
	Rows          [][]string
	selected      int
	offset        int
	label         string
	style         backend.Style
	headerStyle   backend.Style
	selectedStyle backend.Style
	cachedWidths  []int
	cachedTotal   int
	cachedSig     uint32
}

// NewTable creates a table with columns.
func NewTable(columns ...TableColumn) *Table {
	table := &Table{
		Columns:       columns,
		label:         "Table",
		style:         backend.DefaultStyle(),
		headerStyle:   backend.DefaultStyle().Bold(true),
		selectedStyle: backend.DefaultStyle().Reverse(true),
	}
	table.Base.Role = accessibility.RoleTable
	table.syncA11y()
	return table
}

// SetStyle updates the base table style.
func (t *Table) SetStyle(style backend.Style) {
	if t == nil {
		return
	}
	t.style = style
}

// SetHeaderStyle updates the header style.
func (t *Table) SetHeaderStyle(style backend.Style) {
	if t == nil {
		return
	}
	t.headerStyle = style
}

// SetSelectedStyle updates the selected row style.
func (t *Table) SetSelectedStyle(style backend.Style) {
	if t == nil {
		return
	}
	t.selectedStyle = style
}

// StyleType returns the selector type name.
func (t *Table) StyleType() string {
	return "Table"
}

// SetRows updates table rows.
func (t *Table) SetRows(rows [][]string) {
	if t == nil {
		return
	}
	t.Rows = rows
	t.syncA11y()
}

// SetLabel updates the accessibility label.
func (t *Table) SetLabel(label string) {
	if t == nil {
		return
	}
	t.label = label
	t.syncA11y()
}

// SelectedIndex returns the currently selected row index.
func (t *Table) SelectedIndex() int {
	if t == nil {
		return 0
	}
	return t.selected
}

// SetSelected updates the selected row index.
func (t *Table) SetSelected(index int) {
	if t == nil {
		return
	}
	t.setSelected(index)
}

// Measure returns the desired size.
func (t *Table) Measure(constraints runtime.Constraints) runtime.Size {
	return t.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		height := min(len(t.Rows)+1, contentConstraints.MaxHeight)
		if height <= 0 {
			height = contentConstraints.MinHeight
		}
		return contentConstraints.Constrain(runtime.Size{Width: contentConstraints.MaxWidth, Height: height})
	})
}

// Render draws the table.
func (t *Table) Render(ctx runtime.RenderContext) {
	if t == nil {
		return
	}
	t.syncA11y()
	outer := t.bounds
	content := t.ContentBounds()
	if outer.Width <= 0 || outer.Height <= 0 {
		return
	}
	baseStyle := mergeBackendStyles(resolveBaseStyle(ctx, t, backend.DefaultStyle(), false), t.style)
	ctx.Buffer.Fill(outer, ' ', baseStyle)
	if content.Width <= 0 || content.Height <= 0 {
		return
	}
	widths := t.columnWidths(content.Width)
	if len(widths) == 0 {
		return
	}
	// Header
	headerStyle := mergeBackendStyles(baseStyle, t.headerStyle)
	x := content.X
	for i, col := range t.Columns {
		if x >= content.X+content.Width {
			break
		}
		width := widths[i]
		title := truncateString(col.Title, width)
		writePadded(ctx.Buffer, x, content.Y, width, title, headerStyle)
		x += width + 1
	}

	// Rows
	rowArea := content.Height - 1
	if rowArea <= 0 {
		return
	}
	if t.selected < 0 {
		t.selected = 0
	}
	if t.selected >= len(t.Rows) {
		t.selected = len(t.Rows) - 1
	}
	if t.selected < t.offset {
		t.offset = t.selected
	}
	if t.selected >= t.offset+rowArea {
		t.offset = t.selected - rowArea + 1
	}
	for row := 0; row < rowArea; row++ {
		rowIndex := t.offset + row
		if rowIndex < 0 || rowIndex >= len(t.Rows) {
			break
		}
		style := baseStyle
		if rowIndex == t.selected {
			style = mergeBackendStyles(baseStyle, t.selectedStyle)
		}
		x = content.X
		for colIndex, width := range widths {
			if x >= content.X+content.Width {
				break
			}
			cell := ""
			if colIndex < len(t.Rows[rowIndex]) {
				cell = t.Rows[rowIndex][colIndex]
			}
			cell = truncateString(cell, width)
			writePadded(ctx.Buffer, x, content.Y+1+row, width, cell, style)
			x += width + 1
		}
	}
}

// HandleMessage handles row navigation.
func (t *Table) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if t == nil || !t.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	switch key.Key {
	case terminal.KeyUp:
		t.setSelected(t.selected - 1)
		return runtime.Handled()
	case terminal.KeyDown:
		t.setSelected(t.selected + 1)
		return runtime.Handled()
	case terminal.KeyPageUp:
		t.setSelected(t.selected - t.bounds.Height)
		return runtime.Handled()
	case terminal.KeyPageDown:
		t.setSelected(t.selected + t.bounds.Height)
		return runtime.Handled()
	case terminal.KeyHome:
		t.setSelected(0)
		return runtime.Handled()
	case terminal.KeyEnd:
		t.setSelected(len(t.Rows) - 1)
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func (t *Table) setSelected(index int) {
	if t == nil {
		return
	}
	if len(t.Rows) == 0 {
		t.selected = 0
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= len(t.Rows) {
		index = len(t.Rows) - 1
	}
	t.selected = index
	t.syncA11y()
}

func (t *Table) syncA11y() {
	if t == nil {
		return
	}
	if t.Base.Role == "" {
		t.Base.Role = accessibility.RoleTable
	}
	label := strings.TrimSpace(t.label)
	if label == "" {
		label = "Table"
	}
	t.Base.Label = label
	t.Base.Description = fmt.Sprintf("%d rows, %d columns", len(t.Rows), len(t.Columns))
	if t.selected >= 0 && t.selected < len(t.Rows) {
		t.Base.Value = &accessibility.ValueInfo{Text: t.selectedRowSummary()}
	} else {
		t.Base.Value = nil
	}
}

func (t *Table) selectedRowSummary() string {
	if t == nil || t.selected < 0 || t.selected >= len(t.Rows) {
		return ""
	}
	row := t.Rows[t.selected]
	if len(row) == 0 {
		return ""
	}
	out := make([]string, 0, len(row))
	for _, cell := range row {
		cell = strings.TrimSpace(cell)
		if cell == "" {
			continue
		}
		out = append(out, cell)
	}
	return strings.Join(out, " | ")
}

func (t *Table) columnWidths(total int) []int {
	if len(t.Columns) == 0 {
		return nil
	}
	if total == t.cachedTotal && len(t.cachedWidths) == len(t.Columns) && t.cachedSig == t.columnsSignature() {
		return t.cachedWidths
	}
	available := total - (len(t.Columns) - 1)
	if available < 0 {
		available = 0
	}
	fixed := 0
	flexCount := 0
	for _, col := range t.Columns {
		if col.Width > 0 {
			fixed += col.Width
		} else {
			flexCount++
		}
	}
	widths := make([]int, len(t.Columns))
	remaining := available - fixed
	if remaining < 0 {
		remaining = 0
	}
	flexWidth := 0
	if flexCount > 0 {
		flexWidth = remaining / flexCount
		if flexWidth <= 0 {
			flexWidth = 1
		}
	}
	for i, col := range t.Columns {
		if col.Width > 0 {
			widths[i] = col.Width
		} else {
			widths[i] = flexWidth
		}
	}
	t.cachedTotal = total
	t.cachedSig = t.columnsSignature()
	t.cachedWidths = widths
	return widths
}

func (t *Table) columnsSignature() uint32 {
	if t == nil {
		return 0
	}
	var sig uint32 = uint32(len(t.Columns))
	for _, col := range t.Columns {
		sig = sig*31 + uint32(col.Width+1)
	}
	return sig
}

// ScrollBy scrolls selection by delta.
func (t *Table) ScrollBy(dx, dy int) {
	if t == nil || len(t.Rows) == 0 || dy == 0 {
		return
	}
	t.setSelected(t.selected + dy)
	t.Invalidate()
}

// ScrollTo scrolls to an absolute row index.
func (t *Table) ScrollTo(x, y int) {
	if t == nil || len(t.Rows) == 0 {
		return
	}
	t.setSelected(y)
	t.Invalidate()
}

// PageBy scrolls by a number of pages.
func (t *Table) PageBy(pages int) {
	if t == nil || len(t.Rows) == 0 {
		return
	}
	pageSize := t.bounds.Height - 1
	if pageSize < 1 {
		pageSize = 1
	}
	t.setSelected(t.selected + pages*pageSize)
	t.Invalidate()
}

// ScrollToStart scrolls to the first row.
func (t *Table) ScrollToStart() {
	if t == nil || len(t.Rows) == 0 {
		return
	}
	t.setSelected(0)
	t.Invalidate()
}

// ScrollToEnd scrolls to the last row.
func (t *Table) ScrollToEnd() {
	if t == nil || len(t.Rows) == 0 {
		return
	}
	t.setSelected(len(t.Rows) - 1)
	t.Invalidate()
}

var _ scroll.Controller = (*Table)(nil)
