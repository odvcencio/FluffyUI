package widgets

import (
	"fmt"
	"sort"
	"strings"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/scroll"
	"github.com/odvcencio/fluffyui/terminal"
)

// SortDirection indicates the direction of column sorting.
type SortDirection int

const (
	// SortNone means no sorting is applied.
	SortNone SortDirection = iota
	// SortAsc sorts in ascending order.
	SortAsc
	// SortDesc sorts in descending order.
	SortDesc
)

// TableSortState holds the current sort column and direction.
type TableSortState struct {
	Column    int
	Direction SortDirection
}

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
	dataSource    TabularDataSource
	selected      int
	offset        int
	label         string
	style         backend.Style
	headerStyle   backend.Style
	selectedStyle backend.Style
	cachedWidths  []int
	cachedTotal   int
	cachedSig     uint32
	services      runtime.Services
	sortState     TableSortState
	filterFn      func(row []string) bool
	sortedIndices []int
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
	t.dataSource = nil
	t.Rows = rows
	t.rebuildSortedIndices()
	t.syncA11y()
}

// SetDataSource sets a virtualized data source for large datasets.
func (t *Table) SetDataSource(source TabularDataSource) {
	if t == nil {
		return
	}
	t.dataSource = source
	t.syncA11y()
	t.Invalidate()
}

// DataSource returns the active data source.
func (t *Table) DataSource() TabularDataSource {
	if t == nil {
		return nil
	}
	return t.dataSource
}

// SetLabel updates the accessibility label.
func (t *Table) SetLabel(label string) {
	if t == nil {
		return
	}
	t.label = label
	t.syncA11y()
}

// SetSortColumn sets the sort column and direction, then rebuilds the view.
func (t *Table) SetSortColumn(col int, dir SortDirection) {
	if t == nil {
		return
	}
	t.sortState = TableSortState{Column: col, Direction: dir}
	t.rebuildSortedIndices()
	t.syncA11y()
}

// Sort returns the current sort state.
func (t *Table) Sort() TableSortState {
	if t == nil {
		return TableSortState{}
	}
	return t.sortState
}

// SetFilter sets a filter function. Only rows for which fn returns true will
// be displayed. Pass nil to remove the filter. Filtering only applies to
// static Rows data, not to a TabularDataSource.
func (t *Table) SetFilter(fn func(row []string) bool) {
	if t == nil {
		return
	}
	t.filterFn = fn
	t.rebuildSortedIndices()
	t.selected = 0
	t.offset = 0
	t.syncA11y()
}

// ClearFilter removes any active filter.
func (t *Table) ClearFilter() {
	if t == nil {
		return
	}
	t.filterFn = nil
	t.rebuildSortedIndices()
	t.selected = 0
	t.offset = 0
	t.syncA11y()
}

// rebuildSortedIndices recomputes the display-order indices by applying the
// current filter and sort to the static Rows slice. When no filter or sort is
// active (or when a dataSource is set), sortedIndices is set to nil so that
// the original row order is used directly.
func (t *Table) rebuildSortedIndices() {
	if t == nil {
		return
	}
	// Sorting/filtering only applies to static Rows, not dataSource.
	if t.dataSource != nil {
		t.sortedIndices = nil
		return
	}
	hasFilter := t.filterFn != nil
	hasSort := t.sortState.Direction != SortNone

	if !hasFilter && !hasSort {
		t.sortedIndices = nil
		return
	}

	// Build initial index set, applying filter.
	indices := make([]int, 0, len(t.Rows))
	for i, row := range t.Rows {
		if hasFilter && !t.filterFn(row) {
			continue
		}
		indices = append(indices, i)
	}

	// Apply sort.
	if hasSort && t.sortState.Column >= 0 {
		col := t.sortState.Column
		dir := t.sortState.Direction
		sort.Slice(indices, func(a, b int) bool {
			rowA := t.Rows[indices[a]]
			rowB := t.Rows[indices[b]]
			var cellA, cellB string
			if col < len(rowA) {
				cellA = rowA[col]
			}
			if col < len(rowB) {
				cellB = rowB[col]
			}
			if dir == SortAsc {
				return cellA < cellB
			}
			return cellA > cellB
		})
	}

	t.sortedIndices = indices
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

// RowCount returns the number of rows.
func (t *Table) RowCount() int {
	if t == nil {
		return 0
	}
	return t.rowCount()
}

// ColumnCount returns the number of columns.
func (t *Table) ColumnCount() int {
	if t == nil {
		return 0
	}
	return len(t.Columns)
}

// SelectedRow returns the currently selected row data, or nil if no selection.
func (t *Table) SelectedRow() []string {
	if t == nil || t.selected < 0 || t.selected >= t.rowCount() {
		return nil
	}
	if provider, ok := t.dataSource.(TabularRowProvider); ok {
		return provider.Row(t.selected)
	}
	if t.dataSource != nil {
		return nil
	}
	origIndex := t.mapRowIndex(t.selected)
	if origIndex < 0 || origIndex >= len(t.Rows) {
		return nil
	}
	return t.Rows[origIndex]
}

// GetCell returns the cell value at the given row and column.
// When sorting or filtering is active, row refers to the display index.
func (t *Table) GetCell(row, col int) string {
	if t == nil || row < 0 || row >= t.rowCount() {
		return ""
	}
	if t.dataSource != nil {
		return t.dataSource.Cell(row, col)
	}
	origRow := t.mapRowIndex(row)
	if origRow < 0 || origRow >= len(t.Rows) {
		return ""
	}
	if col < 0 || col >= len(t.Rows[origRow]) {
		return ""
	}
	return t.Rows[origRow][col]
}

// SetCell updates a cell value at the given row and column.
// When sorting or filtering is active, row refers to the display index.
func (t *Table) SetCell(row, col int, value string) {
	if t == nil || row < 0 || row >= t.rowCount() {
		return
	}
	if editable, ok := t.dataSource.(TabularEditable); ok {
		editable.SetCell(row, col, value)
		return
	}
	if t.dataSource != nil {
		return
	}
	origRow := t.mapRowIndex(row)
	if origRow < 0 || origRow >= len(t.Rows) {
		return
	}
	// Expand row if needed
	for len(t.Rows[origRow]) <= col {
		t.Rows[origRow] = append(t.Rows[origRow], "")
	}
	t.Rows[origRow][col] = value
}

// Measure returns the desired size.
func (t *Table) Measure(constraints runtime.Constraints) runtime.Size {
	return t.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		height := min(t.rowCount()+1, contentConstraints.MaxHeight)
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
	rowCount := t.rowCount()
	if t.selected < 0 {
		t.selected = 0
	}
	if t.selected >= rowCount {
		t.selected = rowCount - 1
	}
	if t.selected < t.offset {
		t.offset = t.selected
	}
	if t.selected >= t.offset+rowArea {
		t.offset = t.selected - rowArea + 1
	}
	for row := 0; row < rowArea; row++ {
		rowIndex := t.offset + row
		if rowIndex < 0 || rowIndex >= rowCount {
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
			cell := t.GetCell(rowIndex, colIndex)
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
		t.setSelected(t.rowCount() - 1)
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func (t *Table) setSelected(index int) {
	if t == nil {
		return
	}
	rowCount := t.rowCount()
	if rowCount == 0 {
		t.selected = 0
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= rowCount {
		index = rowCount - 1
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
	t.Base.Description = fmt.Sprintf("%d rows, %d columns", t.rowCount(), len(t.Columns))
	if t.selected >= 0 && t.selected < t.rowCount() {
		t.Base.Value = &accessibility.ValueInfo{Text: t.selectedRowSummary()}
	} else {
		t.Base.Value = nil
	}
}

func (t *Table) selectedRowSummary() string {
	if t == nil || t.selected < 0 || t.selected >= t.rowCount() {
		return ""
	}
	if provider, ok := t.dataSource.(TabularRowProvider); ok {
		return summarizeRow(provider.Row(t.selected))
	}
	if t.dataSource != nil {
		cells := make([]string, 0, len(t.Columns))
		for col := range t.Columns {
			cell := strings.TrimSpace(t.dataSource.Cell(t.selected, col))
			if cell == "" {
				continue
			}
			cells = append(cells, cell)
		}
		return strings.Join(cells, " | ")
	}
	origIndex := t.mapRowIndex(t.selected)
	if origIndex < 0 || origIndex >= len(t.Rows) {
		return ""
	}
	return summarizeRow(t.Rows[origIndex])
}

func (t *Table) rowCount() int {
	if t == nil {
		return 0
	}
	if t.dataSource != nil {
		if count := t.dataSource.RowCount(); count > 0 {
			return count
		}
		return 0
	}
	if t.sortedIndices != nil {
		return len(t.sortedIndices)
	}
	return len(t.Rows)
}

// mapRowIndex translates a display row index to an original Rows index.
// When no sortedIndices are active it returns the index unchanged.
func (t *Table) mapRowIndex(displayIndex int) int {
	if t.sortedIndices != nil && displayIndex >= 0 && displayIndex < len(t.sortedIndices) {
		return t.sortedIndices[displayIndex]
	}
	return displayIndex
}

func summarizeRow(row []string) string {
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
	if t == nil || t.rowCount() == 0 || dy == 0 {
		return
	}
	t.setSelected(t.selected + dy)
	t.Invalidate()
}

// ScrollTo scrolls to an absolute row index.
func (t *Table) ScrollTo(x, y int) {
	if t == nil || t.rowCount() == 0 {
		return
	}
	t.setSelected(y)
	t.Invalidate()
}

// PageBy scrolls by a number of pages.
func (t *Table) PageBy(pages int) {
	if t == nil || t.rowCount() == 0 {
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
	if t == nil || t.rowCount() == 0 {
		return
	}
	t.setSelected(0)
	t.Invalidate()
}

// ScrollToEnd scrolls to the last row.
func (t *Table) ScrollToEnd() {
	if t == nil || t.rowCount() == 0 {
		return
	}
	t.setSelected(t.rowCount() - 1)
	t.Invalidate()
}

var _ scroll.Controller = (*Table)(nil)

// Bind attaches app services.
func (t *Table) Bind(services runtime.Services) {
	if t == nil {
		return
	}
	t.services = services
}

// Unbind releases app services.
func (t *Table) Unbind() {
	if t == nil {
		return
	}
	t.services = runtime.Services{}
}

var _ runtime.Widget = (*Table)(nil)
var _ runtime.Focusable = (*Table)(nil)
var _ runtime.Bindable = (*Table)(nil)
var _ runtime.Unbindable = (*Table)(nil)
