package widgets

import (
	"fmt"
	"sort"
	"strings"

	"m31labs.dev/fluffyui/accessibility"
	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/scroll"
	"m31labs.dev/fluffyui/terminal"
)

// ColumnPin determines whether a column is pinned to the left or right edge.
type ColumnPin int

const (
	// PinNone means the column scrolls normally with the middle area.
	PinNone ColumnPin = iota
	// PinLeft pins the column to the left edge.
	PinLeft
	// PinRight pins the column to the right edge.
	PinRight
)

// DataGridColumn defines a column in a DataGrid. Unlike TableColumn, it
// supports resizing constraints, pinning, sortability, and per-cell renderers.
type DataGridColumn struct {
	Header    string                             // column header text
	Width     int                                // current width in cells
	MinWidth  int                                // minimum resize width (default 3)
	MaxWidth  int                                // maximum resize width (0 = unlimited)
	Resizable bool                               // can be resized via Ctrl+Left/Right
	Pinned    ColumnPin                          // pinned to left/right edge
	Sortable  bool                               // can be sorted
	Renderer  func(row int, value string) string // custom cell renderer (optional)
}

// DataGrid is an advanced data grid widget with column resizing, cell
// renderers, column pinning, sorting, and virtual scrolling. It provides
// per-cell selection, inline editing, and ARIA grid semantics.
type DataGrid struct {
	FocusableBase

	columns     []DataGridColumn
	rows        [][]string
	selectedRow int
	selectedCol int
	offset      int
	label       string

	// Sorting
	sortCol   int
	sortDir   SortDirection
	sortedIdx []int // display order indices

	// Inline editing
	editing    bool
	editRow    int
	editCol    int
	editBuf    string
	editCursor int
	onCellEdit func(row, col int, oldValue, newValue string)

	// Virtual scroll
	virtualScroll bool
	virtualList   *scroll.VirtualList
	overscan      int
	lastVisStart  int
	lastVisEnd    int

	// Styles
	style         backend.Style
	headerStyle   backend.Style
	selectedStyle backend.Style
	styleSet      bool

	services runtime.Services
}

// NewDataGrid creates a DataGrid with the given column definitions.
// MinWidth defaults to 3 for all columns. Width is clamped to [MinWidth, MaxWidth].
func NewDataGrid(columns []DataGridColumn) *DataGrid {
	cols := make([]DataGridColumn, len(columns))
	copy(cols, columns)
	for i := range cols {
		if cols[i].MinWidth <= 0 {
			cols[i].MinWidth = 3
		}
		if cols[i].Width < cols[i].MinWidth {
			cols[i].Width = cols[i].MinWidth
		}
		if cols[i].MaxWidth > 0 && cols[i].Width > cols[i].MaxWidth {
			cols[i].Width = cols[i].MaxWidth
		}
	}
	dg := &DataGrid{
		columns:       cols,
		sortCol:       -1,
		sortDir:       SortNone,
		label:         "Data Grid",
		style:         backend.DefaultStyle(),
		headerStyle:   backend.DefaultStyle().Bold(true),
		selectedStyle: backend.DefaultStyle().Reverse(true),
	}
	dg.Base.Role = accessibility.RoleGrid
	dg.syncA11y()
	return dg
}

// ---------------------------------------------------------------------------
// Data management
// ---------------------------------------------------------------------------

// SetRows replaces all row data.
func (dg *DataGrid) SetRows(rows [][]string) {
	if dg == nil {
		return
	}
	dg.rows = rows
	dg.rebuildSortedIndices()
	dg.clampSelection()
	dg.syncA11y()
}

// AddRow appends a single row.
func (dg *DataGrid) AddRow(row []string) {
	if dg == nil {
		return
	}
	dg.rows = append(dg.rows, row)
	dg.rebuildSortedIndices()
	dg.syncA11y()
}

// RemoveRow removes the row at the given display index.
func (dg *DataGrid) RemoveRow(index int) {
	if dg == nil {
		return
	}
	origIdx := dg.mapRowIndex(index)
	if origIdx < 0 || origIdx >= len(dg.rows) {
		return
	}
	dg.rows = append(dg.rows[:origIdx], dg.rows[origIdx+1:]...)
	dg.rebuildSortedIndices()
	dg.clampSelection()
	dg.syncA11y()
}

// Cell returns the cell value at the given display row and column.
func (dg *DataGrid) Cell(row, col int) string {
	if dg == nil || row < 0 || row >= dg.rowCount() {
		return ""
	}
	origRow := dg.mapRowIndex(row)
	if origRow < 0 || origRow >= len(dg.rows) {
		return ""
	}
	if col < 0 || col >= len(dg.rows[origRow]) {
		return ""
	}
	return dg.rows[origRow][col]
}

// SetCell updates a single cell value at the given display row and column.
func (dg *DataGrid) SetCell(row, col int, value string) {
	if dg == nil || row < 0 || row >= dg.rowCount() || col < 0 {
		return
	}
	origRow := dg.mapRowIndex(row)
	if origRow < 0 || origRow >= len(dg.rows) {
		return
	}
	for len(dg.rows[origRow]) <= col {
		dg.rows[origRow] = append(dg.rows[origRow], "")
	}
	dg.rows[origRow][col] = value
}

// RowCount returns the number of visible rows.
func (dg *DataGrid) RowCount() int {
	if dg == nil {
		return 0
	}
	return dg.rowCount()
}

// ColumnCount returns the number of columns.
func (dg *DataGrid) ColumnCount() int {
	if dg == nil {
		return 0
	}
	return len(dg.columns)
}

// Columns returns a copy of the current column definitions.
func (dg *DataGrid) Columns() []DataGridColumn {
	if dg == nil {
		return nil
	}
	out := make([]DataGridColumn, len(dg.columns))
	copy(out, dg.columns)
	return out
}

// ---------------------------------------------------------------------------
// Selection
// ---------------------------------------------------------------------------

// SelectedCell returns the currently selected row and column.
func (dg *DataGrid) SelectedCell() (row, col int) {
	if dg == nil {
		return 0, 0
	}
	return dg.selectedRow, dg.selectedCol
}

// SetSelectedCell updates the selected cell position.
func (dg *DataGrid) SetSelectedCell(row, col int) {
	if dg == nil {
		return
	}
	dg.setSelectedRow(row)
	dg.setSelectedCol(col)
}

// ---------------------------------------------------------------------------
// Column resizing
// ---------------------------------------------------------------------------

// SetColumnWidth sets the width of a column, clamped to [MinWidth, MaxWidth].
func (dg *DataGrid) SetColumnWidth(col, width int) {
	if dg == nil || col < 0 || col >= len(dg.columns) {
		return
	}
	c := &dg.columns[col]
	minW := c.MinWidth
	if minW <= 0 {
		minW = 3
	}
	if width < minW {
		width = minW
	}
	if c.MaxWidth > 0 && width > c.MaxWidth {
		width = c.MaxWidth
	}
	c.Width = width
}

// ColumnWidth returns the current width of a column.
func (dg *DataGrid) ColumnWidth(col int) int {
	if dg == nil || col < 0 || col >= len(dg.columns) {
		return 0
	}
	return dg.columns[col].Width
}

// ---------------------------------------------------------------------------
// Column pinning
// ---------------------------------------------------------------------------

// SetColumnPin sets the pin state for a column.
func (dg *DataGrid) SetColumnPin(col int, pin ColumnPin) {
	if dg == nil || col < 0 || col >= len(dg.columns) {
		return
	}
	dg.columns[col].Pinned = pin
}

// ---------------------------------------------------------------------------
// Sorting
// ---------------------------------------------------------------------------

// Sort sets the sort column and direction, then rebuilds the display order.
func (dg *DataGrid) Sort(col int, dir SortDirection) {
	if dg == nil {
		return
	}
	dg.sortCol = col
	dg.sortDir = dir
	dg.rebuildSortedIndices()
	dg.clampSelection()
	dg.syncA11y()
}

// SortState returns the current sort column and direction.
func (dg *DataGrid) SortState() (col int, dir SortDirection) {
	if dg == nil {
		return -1, SortNone
	}
	return dg.sortCol, dg.sortDir
}

func (dg *DataGrid) rebuildSortedIndices() {
	if dg == nil {
		return
	}
	if dg.sortDir == SortNone || dg.sortCol < 0 {
		dg.sortedIdx = nil
		return
	}
	indices := make([]int, len(dg.rows))
	for i := range indices {
		indices[i] = i
	}
	col := dg.sortCol
	dir := dg.sortDir
	sort.Slice(indices, func(a, b int) bool {
		rowA := dg.rows[indices[a]]
		rowB := dg.rows[indices[b]]
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
	dg.sortedIdx = indices
}

func (dg *DataGrid) mapRowIndex(displayIndex int) int {
	if dg.sortedIdx != nil && displayIndex >= 0 && displayIndex < len(dg.sortedIdx) {
		return dg.sortedIdx[displayIndex]
	}
	return displayIndex
}

func (dg *DataGrid) rowCount() int {
	if dg == nil {
		return 0
	}
	if dg.sortedIdx != nil {
		return len(dg.sortedIdx)
	}
	return len(dg.rows)
}

// ---------------------------------------------------------------------------
// Virtual scroll
// ---------------------------------------------------------------------------

// SetVirtualScroll enables or disables virtual scrolling for large datasets.
func (dg *DataGrid) SetVirtualScroll(enabled bool) {
	if dg == nil {
		return
	}
	dg.virtualScroll = enabled
	if enabled {
		dg.ensureVirtualList()
	}
}

// VirtualScroll reports whether virtual scrolling is enabled.
func (dg *DataGrid) VirtualScroll() bool {
	if dg == nil {
		return false
	}
	return dg.virtualScroll
}

// SetOverscan sets the number of extra rows rendered beyond the viewport.
func (dg *DataGrid) SetOverscan(count int) {
	if dg == nil {
		return
	}
	if count < 0 {
		count = 0
	}
	dg.overscan = count
	if dg.virtualList != nil {
		dg.virtualList.SetOverscan(count)
	}
}

// VisibleRange returns the [start, end) row range currently rendered.
func (dg *DataGrid) VisibleRange() (start, end int) {
	if dg == nil {
		return 0, 0
	}
	return dg.lastVisStart, dg.lastVisEnd
}

func (dg *DataGrid) ensureVirtualList() {
	if dg.virtualList != nil {
		return
	}
	dg.virtualList = scroll.NewVirtualList(dg.rowCount(), 1, nil)
	overscan := dg.overscan
	if overscan <= 0 {
		overscan = 2
	}
	dg.virtualList.SetOverscan(overscan)
}

func (dg *DataGrid) syncVirtualList(rowArea int) {
	if dg.virtualList == nil {
		return
	}
	dg.virtualList.SetItemCount(dg.rowCount())
	dg.virtualList.SetViewportHeight(rowArea)
	dg.virtualList.SetSelected(dg.selectedRow)
	dg.virtualList.EnsureVisible(dg.selectedRow)
}

// ---------------------------------------------------------------------------
// Inline editing
// ---------------------------------------------------------------------------

// SetOnCellEdit sets the callback invoked when a cell edit is committed.
func (dg *DataGrid) SetOnCellEdit(fn func(row, col int, oldValue, newValue string)) {
	if dg == nil {
		return
	}
	dg.onCellEdit = fn
}

// Editing reports whether the grid is in cell edit mode.
func (dg *DataGrid) Editing() bool {
	if dg == nil {
		return false
	}
	return dg.editing
}

// StartEdit enters edit mode on the selected cell.
func (dg *DataGrid) StartEdit() {
	if dg == nil || dg.editing {
		return
	}
	if dg.rowCount() == 0 || len(dg.columns) == 0 {
		return
	}
	dg.editing = true
	dg.editRow = dg.selectedRow
	dg.editCol = dg.selectedCol
	dg.editBuf = dg.Cell(dg.editRow, dg.editCol)
	dg.editCursor = len([]rune(dg.editBuf))
	if announcer := dg.services.Announcer(); announcer != nil {
		announcer.Announce(
			fmt.Sprintf("Editing cell row %d column %d", dg.editRow+1, dg.editCol+1),
			accessibility.PriorityPolite,
		)
	}
}

// CancelEdit discards edits and exits edit mode.
func (dg *DataGrid) CancelEdit() {
	if dg == nil || !dg.editing {
		return
	}
	dg.editing = false
	dg.editBuf = ""
	dg.editCursor = 0
}

// CommitEdit saves edits and exits edit mode.
func (dg *DataGrid) CommitEdit() {
	if dg == nil || !dg.editing {
		return
	}
	oldValue := dg.Cell(dg.editRow, dg.editCol)
	newValue := dg.editBuf
	dg.SetCell(dg.editRow, dg.editCol, newValue)
	dg.editing = false
	dg.editBuf = ""
	dg.editCursor = 0
	if dg.onCellEdit != nil {
		dg.onCellEdit(dg.editRow, dg.editCol, oldValue, newValue)
	}
}

// ---------------------------------------------------------------------------
// Styles
// ---------------------------------------------------------------------------

// SetStyle updates the base style.
func (dg *DataGrid) SetStyle(s backend.Style) {
	if dg == nil {
		return
	}
	dg.style = s
	dg.styleSet = true
}

// SetHeaderStyle updates the header row style.
func (dg *DataGrid) SetHeaderStyle(s backend.Style) {
	if dg == nil {
		return
	}
	dg.headerStyle = s
}

// SetSelectedStyle updates the selected cell/row style.
func (dg *DataGrid) SetSelectedStyle(s backend.Style) {
	if dg == nil {
		return
	}
	dg.selectedStyle = s
}

// SetLabel updates the accessibility label.
func (dg *DataGrid) SetLabel(label string) {
	if dg == nil {
		return
	}
	dg.label = label
	dg.syncA11y()
}

// StyleType returns the selector type name.
func (dg *DataGrid) StyleType() string {
	return "DataGrid"
}

// ---------------------------------------------------------------------------
// Selection helpers
// ---------------------------------------------------------------------------

func (dg *DataGrid) clampSelection() {
	if dg == nil {
		return
	}
	rc := dg.rowCount()
	if rc == 0 {
		dg.selectedRow = 0
		return
	}
	if dg.selectedRow >= rc {
		dg.selectedRow = rc - 1
	}
	if dg.selectedRow < 0 {
		dg.selectedRow = 0
	}
	if len(dg.columns) == 0 {
		dg.selectedCol = 0
		return
	}
	if dg.selectedCol >= len(dg.columns) {
		dg.selectedCol = len(dg.columns) - 1
	}
	if dg.selectedCol < 0 {
		dg.selectedCol = 0
	}
}

func (dg *DataGrid) setSelectedRow(row int) {
	if dg == nil {
		return
	}
	rc := dg.rowCount()
	if rc == 0 {
		dg.selectedRow = 0
		return
	}
	if row < 0 {
		row = 0
	}
	if row >= rc {
		row = rc - 1
	}
	prev := dg.selectedRow
	dg.selectedRow = row
	dg.syncA11y()
	if prev != row {
		if announcer := dg.services.Announcer(); announcer != nil {
			announcer.Announce(fmt.Sprintf("row %d column %d", row+1, dg.selectedCol+1), accessibility.PriorityPolite)
		}
	}
}

func (dg *DataGrid) setSelectedCol(col int) {
	if dg == nil || len(dg.columns) == 0 {
		return
	}
	if col < 0 {
		col = 0
	}
	if col >= len(dg.columns) {
		col = len(dg.columns) - 1
	}
	dg.selectedCol = col
}

// ---------------------------------------------------------------------------
// Column layout
// ---------------------------------------------------------------------------

// columnOrder returns three slices of column indices: pinned-left, unpinned, pinned-right.
func (dg *DataGrid) columnOrder() (left, middle, right []int) {
	for i, c := range dg.columns {
		switch c.Pinned {
		case PinLeft:
			left = append(left, i)
		case PinRight:
			right = append(right, i)
		default:
			middle = append(middle, i)
		}
	}
	return
}

// computeColumnWidths returns the effective width for each column. Columns
// with Width > 0 use that value; others share the remaining space evenly.
func (dg *DataGrid) computeColumnWidths(availableWidth int) []int {
	if len(dg.columns) == 0 {
		return nil
	}
	widths := make([]int, len(dg.columns))
	totalFixed := 0
	flexCount := 0
	for i, c := range dg.columns {
		if c.Width > 0 {
			widths[i] = c.Width
			totalFixed += c.Width
		} else {
			flexCount++
		}
	}
	gaps := len(dg.columns) - 1
	remaining := availableWidth - totalFixed - gaps
	if remaining < 0 {
		remaining = 0
	}
	flexWidth := 0
	if flexCount > 0 {
		flexWidth = remaining / flexCount
		if flexWidth < 1 {
			flexWidth = 1
		}
	}
	for i, c := range dg.columns {
		if c.Width <= 0 {
			widths[i] = flexWidth
		}
	}
	return widths
}

// pinnedWidth returns the total width consumed by pinned columns (including gaps).
func (dg *DataGrid) pinnedWidth(colIndices []int, widths []int) int {
	total := 0
	for i, idx := range colIndices {
		total += widths[idx]
		if i < len(colIndices)-1 {
			total++ // gap
		}
	}
	return total
}

// ---------------------------------------------------------------------------
// Widget interface
// ---------------------------------------------------------------------------

// Measure returns the desired size.
func (dg *DataGrid) Measure(constraints runtime.Constraints) runtime.Size {
	return dg.measureWithStyle(constraints, func(cc runtime.Constraints) runtime.Size {
		height := min(dg.rowCount()+1, cc.MaxHeight) // +1 for header
		if height <= 0 {
			height = cc.MinHeight
		}
		return cc.Constrain(runtime.Size{Width: cc.MaxWidth, Height: height})
	})
}

// Render draws the data grid into the render context.
func (dg *DataGrid) Render(ctx runtime.RenderContext) {
	if dg == nil {
		return
	}
	dg.syncA11y()
	outer := dg.bounds
	content := dg.ContentBounds()
	if outer.Width <= 0 || outer.Height <= 0 {
		return
	}
	baseStyle := resolveBaseStyle(ctx, dg, backend.DefaultStyle(), false)
	if dg.styleSet {
		baseStyle = mergeBackendStyles(baseStyle, dg.style)
	}
	ctx.Buffer.Fill(outer, ' ', baseStyle)
	if content.Width <= 0 || content.Height <= 0 {
		return
	}

	widths := dg.computeColumnWidths(content.Width)
	if len(widths) == 0 {
		return
	}

	leftCols, middleCols, rightCols := dg.columnOrder()
	pinnedLeftW := dg.pinnedWidth(leftCols, widths)
	pinnedRightW := dg.pinnedWidth(rightCols, widths)

	// Header row
	headerStyle := mergeBackendStyles(baseStyle, dg.headerStyle)
	dg.renderHeaderRow(ctx, content, widths, leftCols, middleCols, rightCols, pinnedLeftW, pinnedRightW, headerStyle)

	// Data rows
	rowArea := content.Height - 1
	if rowArea <= 0 {
		return
	}
	rowCount := dg.rowCount()
	dg.clampSelection()

	if dg.virtualScroll {
		dg.renderVirtualRows(ctx, content, widths, baseStyle, rowArea, rowCount, leftCols, middleCols, rightCols, pinnedLeftW, pinnedRightW)
		return
	}

	// Standard scroll
	if dg.selectedRow < dg.offset {
		dg.offset = dg.selectedRow
	}
	if dg.selectedRow >= dg.offset+rowArea {
		dg.offset = dg.selectedRow - rowArea + 1
	}
	dg.lastVisStart = dg.offset
	dg.lastVisEnd = dg.offset + rowArea
	if dg.lastVisEnd > rowCount {
		dg.lastVisEnd = rowCount
	}

	for r := 0; r < rowArea; r++ {
		rowIdx := dg.offset + r
		if rowIdx < 0 || rowIdx >= rowCount {
			break
		}
		dg.renderDataRow(ctx, content, widths, baseStyle, rowIdx, content.Y+1+r, leftCols, middleCols, rightCols, pinnedLeftW, pinnedRightW)
	}
}

func (dg *DataGrid) renderHeaderRow(ctx runtime.RenderContext, content runtime.Rect, widths []int, left, middle, right []int, pinnedLeftW, pinnedRightW int, headerStyle backend.Style) {
	y := content.Y

	// Pinned left
	x := content.X
	for _, idx := range left {
		w := widths[idx]
		title := truncateString(dg.columns[idx].Header, w)
		writePadded(ctx.Buffer, x, y, w, title, headerStyle)
		x += w + 1
	}

	// Middle columns
	middleStart := content.X
	if pinnedLeftW > 0 {
		middleStart = content.X + pinnedLeftW + 1
	}
	mx := middleStart
	middleEnd := content.X + content.Width
	if pinnedRightW > 0 {
		middleEnd = content.X + content.Width - pinnedRightW - 1
	}
	for _, idx := range middle {
		w := widths[idx]
		if mx+w > middleEnd {
			w = middleEnd - mx
		}
		if w <= 0 {
			break
		}
		title := truncateString(dg.columns[idx].Header, w)
		writePadded(ctx.Buffer, mx, y, w, title, headerStyle)
		mx += w + 1
	}

	// Pinned right
	rx := content.X + content.Width - pinnedRightW
	for _, idx := range right {
		w := widths[idx]
		title := truncateString(dg.columns[idx].Header, w)
		writePadded(ctx.Buffer, rx, y, w, title, headerStyle)
		rx += w + 1
	}
}

func (dg *DataGrid) renderDataRow(ctx runtime.RenderContext, content runtime.Rect, widths []int, baseStyle backend.Style, rowIdx, screenY int, left, middle, right []int, pinnedLeftW, pinnedRightW int) {
	rowStyle := baseStyle
	if rowIdx == dg.selectedRow {
		rowStyle = mergeBackendStyles(baseStyle, dg.selectedStyle)
	}

	// Pinned left
	x := content.X
	for _, colIdx := range left {
		w := widths[colIdx]
		cellStyle := rowStyle
		if rowIdx == dg.selectedRow && colIdx == dg.selectedCol {
			cellStyle = mergeBackendStyles(baseStyle, dg.selectedStyle)
		}
		dg.renderCell(ctx, x, screenY, w, rowIdx, colIdx, cellStyle, baseStyle)
		x += w + 1
	}

	// Middle columns
	middleStart := content.X
	if pinnedLeftW > 0 {
		middleStart = content.X + pinnedLeftW + 1
	}
	mx := middleStart
	middleEnd := content.X + content.Width
	if pinnedRightW > 0 {
		middleEnd = content.X + content.Width - pinnedRightW - 1
	}
	for _, colIdx := range middle {
		w := widths[colIdx]
		if mx+w > middleEnd {
			w = middleEnd - mx
		}
		if w <= 0 {
			break
		}
		cellStyle := rowStyle
		if rowIdx == dg.selectedRow && colIdx == dg.selectedCol {
			cellStyle = mergeBackendStyles(baseStyle, dg.selectedStyle)
		}
		dg.renderCell(ctx, mx, screenY, w, rowIdx, colIdx, cellStyle, baseStyle)
		mx += w + 1
	}

	// Pinned right
	rx := content.X + content.Width - pinnedRightW
	for _, colIdx := range right {
		w := widths[colIdx]
		cellStyle := rowStyle
		if rowIdx == dg.selectedRow && colIdx == dg.selectedCol {
			cellStyle = mergeBackendStyles(baseStyle, dg.selectedStyle)
		}
		dg.renderCell(ctx, rx, screenY, w, rowIdx, colIdx, cellStyle, baseStyle)
		rx += w + 1
	}
}

func (dg *DataGrid) renderCell(ctx runtime.RenderContext, x, y, width, rowIdx, colIdx int, cellStyle, baseStyle backend.Style) {
	if dg.editing && rowIdx == dg.editRow && colIdx == dg.editCol {
		dg.renderEditCell(ctx.Buffer, x, y, width, baseStyle)
		return
	}
	cell := dg.Cell(rowIdx, colIdx)
	// Apply custom renderer if present.
	if colIdx >= 0 && colIdx < len(dg.columns) && dg.columns[colIdx].Renderer != nil {
		cell = dg.columns[colIdx].Renderer(rowIdx, cell)
	}
	cell = truncateString(cell, width)
	writePadded(ctx.Buffer, x, y, width, cell, cellStyle)
}

func (dg *DataGrid) renderEditCell(buf *runtime.Buffer, x, y, width int, baseStyle backend.Style) {
	if buf == nil || width <= 0 {
		return
	}
	editStyle := baseStyle.Reverse(true)
	runes := []rune(dg.editBuf)
	for i := 0; i < width; i++ {
		buf.Set(x+i, y, ' ', editStyle)
	}
	for i := 0; i < width && i < len(runes); i++ {
		s := editStyle
		if i == dg.editCursor {
			s = baseStyle.Bold(true)
		}
		buf.Set(x+i, y, runes[i], s)
	}
	if dg.editCursor >= len(runes) && dg.editCursor < width {
		buf.Set(x+dg.editCursor, y, ' ', baseStyle.Bold(true))
	}
}

func (dg *DataGrid) renderVirtualRows(ctx runtime.RenderContext, content runtime.Rect, widths []int, baseStyle backend.Style, rowArea, rowCount int, left, middle, right []int, pinnedLeftW, pinnedRightW int) {
	dg.ensureVirtualList()
	dg.syncVirtualList(rowArea)

	start, end := dg.virtualList.GetVisibleRange()
	dg.offset = dg.virtualList.Offset()
	dg.lastVisStart = start
	dg.lastVisEnd = end

	for i := start; i < end; i++ {
		if i < 0 || i >= rowCount {
			continue
		}
		screenRow := i - dg.offset
		if screenRow < 0 || screenRow >= rowArea {
			continue
		}
		dg.renderDataRow(ctx, content, widths, baseStyle, i, content.Y+1+screenRow, left, middle, right, pinnedLeftW, pinnedRightW)
	}
}

// ---------------------------------------------------------------------------
// HandleMessage
// ---------------------------------------------------------------------------

// HandleMessage processes key events for cell navigation, inline editing, and
// column resizing (Ctrl+Left/Right shrinks/grows the selected column).
func (dg *DataGrid) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if dg == nil || !dg.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}

	// Edit mode
	if dg.editing {
		return dg.handleEditKey(key)
	}

	pageSize := dg.ContentBounds().Height - 1
	if pageSize < 1 {
		pageSize = dg.bounds.Height
	}
	if pageSize < 1 {
		pageSize = 1
	}

	switch key.Key {
	case terminal.KeyUp:
		dg.setSelectedRow(dg.selectedRow - 1)
		return runtime.Handled()
	case terminal.KeyDown:
		dg.setSelectedRow(dg.selectedRow + 1)
		return runtime.Handled()
	case terminal.KeyLeft:
		if key.Ctrl {
			dg.resizeSelectedColumn(-1)
			return runtime.Handled()
		}
		dg.setSelectedCol(dg.selectedCol - 1)
		return runtime.Handled()
	case terminal.KeyRight:
		if key.Ctrl {
			dg.resizeSelectedColumn(1)
			return runtime.Handled()
		}
		dg.setSelectedCol(dg.selectedCol + 1)
		return runtime.Handled()
	case terminal.KeyPageUp:
		dg.setSelectedRow(dg.selectedRow - pageSize)
		return runtime.Handled()
	case terminal.KeyPageDown:
		dg.setSelectedRow(dg.selectedRow + pageSize)
		return runtime.Handled()
	case terminal.KeyHome:
		dg.setSelectedRow(0)
		return runtime.Handled()
	case terminal.KeyEnd:
		dg.setSelectedRow(dg.rowCount() - 1)
		return runtime.Handled()
	case terminal.KeyEnter, terminal.KeyF2:
		dg.StartEdit()
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func (dg *DataGrid) handleEditKey(key runtime.KeyMsg) runtime.HandleResult {
	runes := []rune(dg.editBuf)
	switch key.Key {
	case terminal.KeyEnter:
		dg.CommitEdit()
		return runtime.Handled()
	case terminal.KeyEscape:
		dg.CancelEdit()
		return runtime.Handled()
	case terminal.KeyBackspace:
		if dg.editCursor > 0 {
			runes = append(runes[:dg.editCursor-1], runes[dg.editCursor:]...)
			dg.editBuf = string(runes)
			dg.editCursor--
		}
		return runtime.Handled()
	case terminal.KeyDelete:
		if dg.editCursor < len(runes) {
			runes = append(runes[:dg.editCursor], runes[dg.editCursor+1:]...)
			dg.editBuf = string(runes)
		}
		return runtime.Handled()
	case terminal.KeyLeft:
		if dg.editCursor > 0 {
			dg.editCursor--
		}
		return runtime.Handled()
	case terminal.KeyRight:
		if dg.editCursor < len(runes) {
			dg.editCursor++
		}
		return runtime.Handled()
	case terminal.KeyHome:
		dg.editCursor = 0
		return runtime.Handled()
	case terminal.KeyEnd:
		dg.editCursor = len(runes)
		return runtime.Handled()
	case terminal.KeyRune:
		runes = append(runes[:dg.editCursor], append([]rune{key.Rune}, runes[dg.editCursor:]...)...)
		dg.editBuf = string(runes)
		dg.editCursor++
		return runtime.Handled()
	}
	return runtime.Handled() // Swallow unrecognized keys while editing
}

func (dg *DataGrid) resizeSelectedColumn(delta int) {
	col := dg.selectedCol
	if col < 0 || col >= len(dg.columns) {
		return
	}
	if !dg.columns[col].Resizable {
		return
	}
	dg.SetColumnWidth(col, dg.columns[col].Width+delta)
}

// ---------------------------------------------------------------------------
// scroll.Controller
// ---------------------------------------------------------------------------

// ScrollBy scrolls selection by delta.
func (dg *DataGrid) ScrollBy(dx, dy int) {
	if dg == nil || dg.rowCount() == 0 {
		return
	}
	if dx != 0 {
		dg.setSelectedCol(dg.selectedCol + dx)
	}
	if dy != 0 {
		dg.setSelectedRow(dg.selectedRow + dy)
	}
	dg.Invalidate()
}

// ScrollTo scrolls to an absolute position.
func (dg *DataGrid) ScrollTo(x, y int) {
	if dg == nil || dg.rowCount() == 0 {
		return
	}
	dg.setSelectedRow(y)
	dg.Invalidate()
}

// PageBy scrolls by a number of pages.
func (dg *DataGrid) PageBy(pages int) {
	if dg == nil || dg.rowCount() == 0 {
		return
	}
	pageSize := dg.bounds.Height - 1
	if pageSize < 1 {
		pageSize = 1
	}
	dg.setSelectedRow(dg.selectedRow + pages*pageSize)
	dg.Invalidate()
}

// ScrollToStart scrolls to the first row.
func (dg *DataGrid) ScrollToStart() {
	if dg == nil || dg.rowCount() == 0 {
		return
	}
	dg.setSelectedRow(0)
	dg.Invalidate()
}

// ScrollToEnd scrolls to the last row.
func (dg *DataGrid) ScrollToEnd() {
	if dg == nil || dg.rowCount() == 0 {
		return
	}
	dg.setSelectedRow(dg.rowCount() - 1)
	dg.Invalidate()
}

var _ scroll.Controller = (*DataGrid)(nil)

// ---------------------------------------------------------------------------
// Bind / Unbind
// ---------------------------------------------------------------------------

// Bind attaches app services.
func (dg *DataGrid) Bind(services runtime.Services) {
	if dg == nil {
		return
	}
	dg.services = services
}

// Unbind releases app services.
func (dg *DataGrid) Unbind() {
	if dg == nil {
		return
	}
	dg.services = runtime.Services{}
}

// ---------------------------------------------------------------------------
// Accessibility
// ---------------------------------------------------------------------------

func (dg *DataGrid) syncA11y() {
	if dg == nil {
		return
	}
	if dg.Base.Role == "" {
		dg.Base.Role = accessibility.RoleGrid
	}
	label := strings.TrimSpace(dg.label)
	if label == "" {
		label = "Data Grid"
	}
	dg.Base.Label = label

	// Build description with column headers
	colHeaders := make([]string, 0, len(dg.columns))
	for _, c := range dg.columns {
		h := strings.TrimSpace(c.Header)
		if h != "" {
			colHeaders = append(colHeaders, h)
		}
	}
	if len(colHeaders) > 0 {
		dg.Base.Description = fmt.Sprintf("%d rows, %d columns: %s", dg.rowCount(), len(dg.columns), strings.Join(colHeaders, ", "))
	} else {
		dg.Base.Description = fmt.Sprintf("%d rows, %d columns", dg.rowCount(), len(dg.columns))
	}

	// Sort
	if dg.sortDir != SortNone && dg.sortCol >= 0 && dg.sortCol < len(dg.columns) {
		switch dg.sortDir {
		case SortAsc:
			dg.Base.Sort = "ascending"
		case SortDesc:
			dg.Base.Sort = "descending"
		}
	} else {
		dg.Base.Sort = ""
	}

	// Value: selected cell summary
	if dg.selectedRow >= 0 && dg.selectedRow < dg.rowCount() {
		dg.Base.Value = &accessibility.ValueInfo{Text: dg.selectedCellSummary()}
	} else {
		dg.Base.Value = nil
	}
}

func (dg *DataGrid) selectedCellSummary() string {
	if dg == nil || dg.selectedRow < 0 || dg.selectedRow >= dg.rowCount() {
		return ""
	}
	origRow := dg.mapRowIndex(dg.selectedRow)
	if origRow < 0 || origRow >= len(dg.rows) {
		return ""
	}
	row := dg.rows[origRow]
	cells := make([]string, 0, len(row))
	for _, c := range row {
		c = strings.TrimSpace(c)
		if c != "" {
			cells = append(cells, c)
		}
	}
	return strings.Join(cells, " | ")
}

// ---------------------------------------------------------------------------
// Interface assertions
// ---------------------------------------------------------------------------

var _ runtime.Widget = (*DataGrid)(nil)
var _ runtime.Focusable = (*DataGrid)(nil)
var _ runtime.Bindable = (*DataGrid)(nil)
var _ runtime.Unbindable = (*DataGrid)(nil)
