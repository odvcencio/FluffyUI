package widgets

import (
	"fmt"
	"strings"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

// DataGridOption configures a DataGrid widget.
type DataGridOption = Option[DataGrid]

// WithDataGridLabel sets the accessibility label.
func WithDataGridLabel(label string) DataGridOption {
	return func(g *DataGrid) {
		if g == nil {
			return
		}
		g.label = label
	}
}

// WithDataGridStyle sets the base style.
func WithDataGridStyle(style backend.Style) DataGridOption {
	return func(g *DataGrid) {
		if g == nil {
			return
		}
		g.style = style
		g.styleSet = true
	}
}

// DataGrid is a table with per-cell selection and inline editing.
type DataGrid struct {
	FocusableBase

	Columns        []TableColumn
	Rows           [][]string
	dataSource     TabularDataSource
	selectedRow    int
	selectedCol    int
	offset         int
	label          string
	style          backend.Style
	headerStyle    backend.Style
	selectedStyle  backend.Style
	editingStyle   backend.Style
	styleSet       bool
	editor         *Input
	editing        bool
	editOriginal   string
	onEdit         func(row, col int, value string)
	onSelect       func(row, col int)
	cachedWidths   []int
	cachedTotal    int
	cachedSig      uint32
	editorHasFocus bool
}

// NewDataGrid creates a new data grid widget.
func NewDataGrid(columns ...TableColumn) *DataGrid {
	grid := &DataGrid{
		Columns:       columns,
		label:         "Data Grid",
		style:         backend.DefaultStyle(),
		headerStyle:   backend.DefaultStyle().Bold(true),
		selectedStyle: backend.DefaultStyle().Reverse(true),
		editingStyle:  backend.DefaultStyle().Reverse(true),
		editor:        NewInput(),
	}
	grid.Base.Role = accessibility.RoleTable
	grid.syncA11y()
	return grid
}

// Bind attaches app services.
func (g *DataGrid) Bind(services runtime.Services) {
	if g == nil {
		return
	}
	if g.editor != nil {
		g.editor.Bind(services)
	}
}

// Unbind releases app services.
func (g *DataGrid) Unbind() {
	if g == nil {
		return
	}
	if g.editor != nil {
		g.editor.Unbind()
	}
}

// SetColumns updates the column definitions.
func (g *DataGrid) SetColumns(columns []TableColumn) {
	if g == nil {
		return
	}
	g.Columns = columns
	g.cachedSig = 0
	g.cachedWidths = nil
	g.syncA11y()
	g.Invalidate()
}

// SetRows updates the data rows.
func (g *DataGrid) SetRows(rows [][]string) {
	if g == nil {
		return
	}
	g.dataSource = nil
	g.Rows = rows
	g.ensureSelectionInRange()
	g.syncA11y()
	g.Invalidate()
}

// SetDataSource sets a virtualized data source for large datasets.
func (g *DataGrid) SetDataSource(source TabularDataSource) {
	if g == nil {
		return
	}
	g.dataSource = source
	g.ensureSelectionInRange()
	g.syncA11y()
	g.Invalidate()
}

// DataSource returns the active data source.
func (g *DataGrid) DataSource() TabularDataSource {
	if g == nil {
		return nil
	}
	return g.dataSource
}

// SetLabel updates the accessibility label.
func (g *DataGrid) SetLabel(label string) {
	if g == nil {
		return
	}
	g.label = label
	g.syncA11y()
}

// SetStyle updates the base style.
func (g *DataGrid) SetStyle(style backend.Style) {
	if g == nil {
		return
	}
	g.style = style
	g.styleSet = true
}

// SetHeaderStyle updates the header style.
func (g *DataGrid) SetHeaderStyle(style backend.Style) {
	if g == nil {
		return
	}
	g.headerStyle = style
}

// SetSelectedStyle updates the selected cell style.
func (g *DataGrid) SetSelectedStyle(style backend.Style) {
	if g == nil {
		return
	}
	g.selectedStyle = style
}

// SetEditingStyle updates the editing cell style.
func (g *DataGrid) SetEditingStyle(style backend.Style) {
	if g == nil {
		return
	}
	g.editingStyle = style
}

// SetOnEdit registers a commit handler for edits.
func (g *DataGrid) SetOnEdit(fn func(row, col int, value string)) {
	if g == nil {
		return
	}
	g.onEdit = fn
}

// SetOnSelect registers a selection handler.
func (g *DataGrid) SetOnSelect(fn func(row, col int)) {
	if g == nil {
		return
	}
	g.onSelect = fn
}

// StyleType returns the selector type name.
func (g *DataGrid) StyleType() string {
	return "DataGrid"
}

// SelectedPosition returns the selected row and column.
func (g *DataGrid) SelectedPosition() (int, int) {
	if g == nil {
		return 0, 0
	}
	return g.selectedRow, g.selectedCol
}

// SetSelected updates the selected row/column.
func (g *DataGrid) SetSelected(row, col int) {
	if g == nil {
		return
	}
	g.selectedRow = row
	g.selectedCol = col
	g.ensureSelectionInRange()
	g.ensureSelectionVisible()
	g.syncA11y()
	g.Invalidate()
	if g.onSelect != nil {
		g.onSelect(g.selectedRow, g.selectedCol)
	}
}

// IsEditing reports whether the grid is in edit mode.
func (g *DataGrid) IsEditing() bool {
	if g == nil {
		return false
	}
	return g.editing
}

// StartEditing begins editing the selected cell.
func (g *DataGrid) StartEditing() {
	if g == nil || g.editor == nil {
		return
	}
	if g.RowCount() == 0 || len(g.Columns) == 0 {
		return
	}
	g.editing = true
	g.editOriginal = g.GetCell(g.selectedRow, g.selectedCol)
	g.editor.SetText(g.editOriginal)
	g.editor.Focus()
	g.editorHasFocus = true
	g.Invalidate()
}

// CancelEditing discards edits.
func (g *DataGrid) CancelEditing() {
	if g == nil {
		return
	}
	g.editing = false
	g.editOriginal = ""
	if g.editor != nil {
		g.editor.Blur()
	}
	g.editorHasFocus = false
	g.Invalidate()
}

// CommitEditing saves edits to the data grid.
func (g *DataGrid) CommitEditing() {
	if g == nil || !g.editing {
		return
	}
	value := ""
	if g.editor != nil {
		value = g.editor.Text()
	}
	g.SetCell(g.selectedRow, g.selectedCol, value)
	if g.onEdit != nil {
		g.onEdit(g.selectedRow, g.selectedCol, value)
	}
	g.editing = false
	g.editOriginal = ""
	if g.editor != nil {
		g.editor.Blur()
	}
	g.editorHasFocus = false
	g.syncA11y()
	g.Invalidate()
}

// RowCount returns the number of rows.
func (g *DataGrid) RowCount() int {
	if g == nil {
		return 0
	}
	if g.dataSource != nil {
		if count := g.dataSource.RowCount(); count > 0 {
			return count
		}
		return 0
	}
	return len(g.Rows)
}

// ColumnCount returns the number of columns.
func (g *DataGrid) ColumnCount() int {
	if g == nil {
		return 0
	}
	return len(g.Columns)
}

// GetCell returns the cell value at row/column.
func (g *DataGrid) GetCell(row, col int) string {
	if g == nil || row < 0 || row >= g.RowCount() {
		return ""
	}
	if g.dataSource != nil {
		return g.dataSource.Cell(row, col)
	}
	if col < 0 || col >= len(g.Rows[row]) {
		return ""
	}
	return g.Rows[row][col]
}

// SetCell updates a cell value at row/column.
func (g *DataGrid) SetCell(row, col int, value string) {
	if g == nil || row < 0 || row >= g.RowCount() {
		return
	}
	if editable, ok := g.dataSource.(TabularEditable); ok {
		editable.SetCell(row, col, value)
		return
	}
	if g.dataSource != nil {
		return
	}
	for len(g.Rows[row]) <= col {
		g.Rows[row] = append(g.Rows[row], "")
	}
	g.Rows[row][col] = value
}

// Measure returns the desired size.
func (g *DataGrid) Measure(constraints runtime.Constraints) runtime.Size {
	return g.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		height := min(g.RowCount()+1, contentConstraints.MaxHeight)
		if height <= 0 {
			height = contentConstraints.MinHeight
		}
		return contentConstraints.Constrain(runtime.Size{Width: contentConstraints.MaxWidth, Height: height})
	})
}

// Layout stores bounds and positions the editor.
func (g *DataGrid) Layout(bounds runtime.Rect) {
	g.FocusableBase.Layout(bounds)
	if g.editor != nil && g.editing {
		g.layoutEditor(g.ContentBounds())
	}
}

// Render draws the grid.
func (g *DataGrid) Render(ctx runtime.RenderContext) {
	if g == nil {
		return
	}
	g.syncA11y()
	outer := g.bounds
	content := g.ContentBounds()
	if outer.Width <= 0 || outer.Height <= 0 {
		return
	}
	baseStyle := resolveBaseStyle(ctx, g, backend.DefaultStyle(), false)
	if g.styleSet {
		baseStyle = mergeBackendStyles(baseStyle, g.style)
	}
	ctx.Buffer.Fill(outer, ' ', baseStyle)
	if content.Width <= 0 || content.Height <= 0 {
		return
	}
	widths := g.columnWidths(content.Width)
	if len(widths) == 0 {
		return
	}
	// Header row
	headerStyle := mergeBackendStyles(baseStyle, g.headerStyle)
	x := content.X
	for i, col := range g.Columns {
		if x >= content.X+content.Width {
			break
		}
		width := widths[i]
		title := truncateString(col.Title, width)
		writePadded(ctx.Buffer, x, content.Y, width, title, headerStyle)
		x += width + 1
	}

	rowArea := content.Height - 1
	if rowArea <= 0 {
		return
	}
	g.ensureSelectionInRange()
	g.ensureSelectionVisibleWithHeight(rowArea)
	rowCount := g.RowCount()
	for row := 0; row < rowArea; row++ {
		rowIndex := g.offset + row
		if rowIndex < 0 || rowIndex >= rowCount {
			break
		}
		x = content.X
		for colIndex, width := range widths {
			if x >= content.X+content.Width {
				break
			}
			cell := g.GetCell(rowIndex, colIndex)
			style := baseStyle
			if rowIndex == g.selectedRow && colIndex == g.selectedCol {
				style = mergeBackendStyles(baseStyle, g.selectedStyle)
				if g.editing {
					style = mergeBackendStyles(baseStyle, g.editingStyle)
				}
			}
			cell = truncateString(cell, width)
			writePadded(ctx.Buffer, x, content.Y+1+row, width, cell, style)
			x += width + 1
		}
	}

	if g.editing && g.editor != nil {
		editStyle := mergeBackendStyles(baseStyle, g.editingStyle)
		g.editor.SetStyle(editStyle)
		g.editor.SetFocusStyle(editStyle.Bold(true))
		g.layoutEditor(content)
		g.editor.Render(ctx)
	}
}

// HandleMessage processes navigation and editing.
func (g *DataGrid) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if g == nil || !g.focused {
		return runtime.Unhandled()
	}
	if g.editing {
		if key, ok := msg.(runtime.KeyMsg); ok {
			switch key.Key {
			case terminal.KeyEnter:
				g.CommitEditing()
				return runtime.Handled()
			case terminal.KeyEscape:
				g.CancelEditing()
				return runtime.Handled()
			}
		}
		if g.editor != nil {
			if !g.editorHasFocus {
				g.editor.Focus()
				g.editorHasFocus = true
			}
			if result := g.editor.HandleMessage(msg); result.Handled {
				return result
			}
		}
		return runtime.Handled()
	}

	switch ev := msg.(type) {
	case runtime.KeyMsg:
		switch ev.Key {
		case terminal.KeyUp:
			g.SetSelected(g.selectedRow-1, g.selectedCol)
			return runtime.Handled()
		case terminal.KeyDown:
			g.SetSelected(g.selectedRow+1, g.selectedCol)
			return runtime.Handled()
		case terminal.KeyLeft:
			g.SetSelected(g.selectedRow, g.selectedCol-1)
			return runtime.Handled()
		case terminal.KeyRight:
			g.SetSelected(g.selectedRow, g.selectedCol+1)
			return runtime.Handled()
		case terminal.KeyHome:
			g.SetSelected(0, g.selectedCol)
			return runtime.Handled()
	case terminal.KeyEnd:
		g.SetSelected(g.RowCount()-1, g.selectedCol)
		return runtime.Handled()
		case terminal.KeyPageUp:
			g.SetSelected(g.selectedRow-(g.bounds.Height-1), g.selectedCol)
			return runtime.Handled()
		case terminal.KeyPageDown:
			g.SetSelected(g.selectedRow+(g.bounds.Height-1), g.selectedCol)
			return runtime.Handled()
		case terminal.KeyTab:
			g.SetSelected(g.selectedRow, g.selectedCol+1)
			return runtime.Handled()
		case terminal.KeyEnter, terminal.KeyF2:
			g.StartEditing()
			return runtime.Handled()
		}
	case runtime.MouseMsg:
		if ev.Button == runtime.MouseWheelUp {
			g.SetSelected(g.selectedRow-1, g.selectedCol)
			return runtime.Handled()
		}
		if ev.Button == runtime.MouseWheelDown {
			g.SetSelected(g.selectedRow+1, g.selectedCol)
			return runtime.Handled()
		}
		if ev.Button == runtime.MouseLeft && ev.Action == runtime.MousePress {
			row, col, ok := g.cellAt(ev.X, ev.Y)
			if ok {
				g.SetSelected(row, col)
				return runtime.Handled()
			}
		}
	}
	return runtime.Unhandled()
}

func (g *DataGrid) ensureSelectionInRange() {
	if g == nil {
		return
	}
	rowCount := g.RowCount()
	if rowCount == 0 {
		g.selectedRow = 0
		g.selectedCol = 0
		return
	}
	if g.selectedRow < 0 {
		g.selectedRow = 0
	}
	if g.selectedRow >= rowCount {
		g.selectedRow = rowCount - 1
	}
	if g.selectedCol < 0 {
		g.selectedCol = 0
	}
	if len(g.Columns) > 0 && g.selectedCol >= len(g.Columns) {
		g.selectedCol = len(g.Columns) - 1
	}
	if len(g.Columns) == 0 {
		g.selectedCol = 0
	}
}

func (g *DataGrid) ensureSelectionVisible() {
	g.ensureSelectionVisibleWithHeight(g.ContentBounds().Height - 1)
}

func (g *DataGrid) ensureSelectionVisibleWithHeight(rowArea int) {
	if g == nil || rowArea <= 0 {
		return
	}
	if g.selectedRow < g.offset {
		g.offset = g.selectedRow
	}
	if g.selectedRow >= g.offset+rowArea {
		g.offset = g.selectedRow - rowArea + 1
	}
	if g.offset < 0 {
		g.offset = 0
	}
	maxOffset := g.RowCount() - rowArea
	if maxOffset < 0 {
		maxOffset = 0
	}
	if g.offset > maxOffset {
		g.offset = maxOffset
	}
}

func (g *DataGrid) cellAt(x, y int) (int, int, bool) {
	if g == nil {
		return 0, 0, false
	}
	content := g.ContentBounds()
	if x < content.X || x >= content.X+content.Width || y < content.Y+1 || y >= content.Y+content.Height {
		return 0, 0, false
	}
	widths := g.columnWidths(content.Width)
	if len(widths) == 0 {
		return 0, 0, false
	}
	rowIndex := g.offset + (y - (content.Y + 1))
	if rowIndex < 0 || rowIndex >= g.RowCount() {
		return 0, 0, false
	}
	colIndex := -1
	cursor := content.X
	for i, width := range widths {
		if x >= cursor && x < cursor+width {
			colIndex = i
			break
		}
		cursor += width + 1
	}
	if colIndex < 0 {
		return 0, 0, false
	}
	return rowIndex, colIndex, true
}

func (g *DataGrid) layoutEditor(content runtime.Rect) {
	if g == nil || g.editor == nil || !g.editing {
		return
	}
	if content.Width <= 0 || content.Height <= 0 {
		return
	}
	widths := g.columnWidths(content.Width)
	if len(widths) == 0 {
		return
	}
	rowArea := content.Height - 1
	if rowArea <= 0 {
		return
	}
	row := g.selectedRow - g.offset
	if row < 0 || row >= rowArea {
		return
	}
	x := content.X
	for colIndex, width := range widths {
		if colIndex == g.selectedCol {
			editorBounds := runtime.Rect{
				X:      x,
				Y:      content.Y + 1 + row,
				Width:  width,
				Height: 1,
			}
			g.editor.Layout(editorBounds)
			return
		}
		x += width + 1
	}
}

func (g *DataGrid) syncA11y() {
	if g == nil {
		return
	}
	if g.Base.Role == "" {
		g.Base.Role = accessibility.RoleTable
	}
	label := strings.TrimSpace(g.label)
	if label == "" {
		label = "Data Grid"
	}
	g.Base.Label = label
	rowCount := g.RowCount()
	g.Base.Description = fmt.Sprintf("%d rows, %d columns", rowCount, len(g.Columns))
	if rowCount > 0 && g.selectedRow >= 0 && g.selectedRow < rowCount {
		value := g.GetCell(g.selectedRow, g.selectedCol)
		g.Base.Value = &accessibility.ValueInfo{Text: value}
	} else {
		g.Base.Value = nil
	}
}

func (g *DataGrid) columnWidths(total int) []int {
	if len(g.Columns) == 0 {
		return nil
	}
	if total == g.cachedTotal && len(g.cachedWidths) == len(g.Columns) && g.cachedSig == g.columnsSignature() {
		return g.cachedWidths
	}
	available := total - (len(g.Columns) - 1)
	if available < 0 {
		available = 0
	}
	fixed := 0
	flexCount := 0
	for _, col := range g.Columns {
		if col.Width > 0 {
			fixed += col.Width
		} else {
			flexCount++
		}
	}
	widths := make([]int, len(g.Columns))
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
	for i, col := range g.Columns {
		if col.Width > 0 {
			widths[i] = col.Width
		} else {
			widths[i] = flexWidth
		}
	}
	g.cachedTotal = total
	g.cachedSig = g.columnsSignature()
	g.cachedWidths = widths
	return widths
}

func (g *DataGrid) columnsSignature() uint32 {
	if g == nil {
		return 0
	}
	var sig uint32 = uint32(len(g.Columns))
	for _, col := range g.Columns {
		sig = sig*31 + uint32(col.Width+1)
	}
	return sig
}

var _ runtime.Widget = (*DataGrid)(nil)
var _ runtime.Focusable = (*DataGrid)(nil)
