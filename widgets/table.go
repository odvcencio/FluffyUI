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

	// Virtual scrolling fields
	virtualScroll  bool
	virtualList    *scroll.VirtualList
	virtualOvrscn  int
	lastVisStart   int
	lastVisEnd     int

	// Inline cell editing fields
	editing    bool
	editRow    int
	editCol    int
	editBuf    string
	editCursor int
	selectedCol int
	onCellEdit func(row, col int, oldValue, newValue string)
}

// NewTable creates a data table with the given column definitions. Populate
// rows with SetRows or SetDataSource. Supports keyboard navigation, sorting,
// filtering, and virtual scrolling for large datasets.
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

// SetVirtualScroll enables or disables virtual scrolling for large datasets.
// When enabled, only visible rows (plus overscan) are rendered each frame,
// which dramatically improves performance for tables with thousands of rows.
func (t *Table) SetVirtualScroll(enabled bool) {
	if t == nil {
		return
	}
	t.virtualScroll = enabled
	if enabled {
		t.ensureVirtualList()
	}
}

// VirtualScroll reports whether virtual scrolling is enabled.
func (t *Table) VirtualScroll() bool {
	if t == nil {
		return false
	}
	return t.virtualScroll
}

// SetVirtualOverscan sets the number of extra rows to render above and below
// the visible area when virtual scrolling is enabled. Default is 2.
func (t *Table) SetVirtualOverscan(count int) {
	if t == nil {
		return
	}
	if count < 0 {
		count = 0
	}
	t.virtualOvrscn = count
	if t.virtualList != nil {
		t.virtualList.SetOverscan(count)
	}
}

// VisibleRange returns the [start, end) row indices currently rendered.
// When virtual scrolling is disabled, start is the scroll offset and end
// is offset+viewportRows. When enabled, it returns the virtualized range.
func (t *Table) VisibleRange() (start, end int) {
	if t == nil {
		return 0, 0
	}
	return t.lastVisStart, t.lastVisEnd
}

func (t *Table) ensureVirtualList() {
	if t.virtualList != nil {
		return
	}
	t.virtualList = scroll.NewVirtualList(t.rowCount(), 1, nil)
	overscan := t.virtualOvrscn
	if overscan <= 0 {
		overscan = 2
	}
	t.virtualList.SetOverscan(overscan)
}

func (t *Table) syncVirtualList(rowArea int) {
	if t.virtualList == nil {
		return
	}
	t.virtualList.SetItemCount(t.rowCount())
	t.virtualList.SetViewportHeight(rowArea)
	t.virtualList.SetSelected(t.selected)
	t.virtualList.EnsureVisible(t.selected)
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

	if t.virtualScroll {
		t.renderVirtualRows(ctx, content, widths, baseStyle, rowArea, rowCount)
		return
	}

	if t.selected < t.offset {
		t.offset = t.selected
	}
	if t.selected >= t.offset+rowArea {
		t.offset = t.selected - rowArea + 1
	}
	t.lastVisStart = t.offset
	t.lastVisEnd = t.offset + rowArea
	if t.lastVisEnd > rowCount {
		t.lastVisEnd = rowCount
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
			if t.editing && rowIndex == t.editRow && colIndex == t.editCol {
				t.renderEditCell(ctx.Buffer, x, content.Y+1+row, width, baseStyle)
			} else {
				cell := t.GetCell(rowIndex, colIndex)
				cell = truncateString(cell, width)
				writePadded(ctx.Buffer, x, content.Y+1+row, width, cell, style)
			}
			x += width + 1
		}
	}
}

func (t *Table) renderVirtualRows(ctx runtime.RenderContext, content runtime.Rect, widths []int, baseStyle backend.Style, rowArea, rowCount int) {
	t.ensureVirtualList()
	t.syncVirtualList(rowArea)

	start, end := t.virtualList.GetVisibleRange()
	t.offset = t.virtualList.Offset()
	t.lastVisStart = start
	t.lastVisEnd = end

	for i := start; i < end; i++ {
		if i < 0 || i >= rowCount {
			continue
		}
		screenRow := i - t.offset
		if screenRow < 0 || screenRow >= rowArea {
			continue
		}
		style := baseStyle
		if i == t.selected {
			style = mergeBackendStyles(baseStyle, t.selectedStyle)
		}
		x := content.X
		for colIndex, width := range widths {
			if x >= content.X+content.Width {
				break
			}
			if t.editing && i == t.editRow && colIndex == t.editCol {
				t.renderEditCell(ctx.Buffer, x, content.Y+1+screenRow, width, baseStyle)
				x += width + 1
				continue
			}
			cell := t.GetCell(i, colIndex)
			cell = truncateString(cell, width)
			writePadded(ctx.Buffer, x, content.Y+1+screenRow, width, cell, style)
			x += width + 1
		}
	}
}

// HandleMessage handles row navigation and inline cell editing.
func (t *Table) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if t == nil || !t.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}

	// When in edit mode, handle editing keys first.
	if t.editing {
		return t.handleEditKey(key)
	}

	pageSize := t.ContentBounds().Height - 1
	if pageSize < 1 {
		pageSize = t.bounds.Height
	}
	if pageSize < 1 {
		pageSize = 1
	}
	switch key.Key {
	case terminal.KeyUp:
		t.setSelected(t.selected - 1)
		return runtime.Handled()
	case terminal.KeyDown:
		t.setSelected(t.selected + 1)
		return runtime.Handled()
	case terminal.KeyLeft:
		t.setSelectedCol(t.selectedCol - 1)
		return runtime.Handled()
	case terminal.KeyRight:
		t.setSelectedCol(t.selectedCol + 1)
		return runtime.Handled()
	case terminal.KeyPageUp:
		t.setSelected(t.selected - pageSize)
		return runtime.Handled()
	case terminal.KeyPageDown:
		t.setSelected(t.selected + pageSize)
		return runtime.Handled()
	case terminal.KeyHome:
		t.setSelected(0)
		return runtime.Handled()
	case terminal.KeyEnd:
		t.setSelected(t.rowCount() - 1)
		return runtime.Handled()
	case terminal.KeyEnter, terminal.KeyF2:
		t.StartEdit()
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

// renderEditCell draws a cell being edited, with the edit buffer and cursor.
func (t *Table) renderEditCell(buf *runtime.Buffer, x, y, width int, baseStyle backend.Style) {
	if buf == nil || width <= 0 {
		return
	}
	editStyle := baseStyle.Reverse(true)
	runes := []rune(t.editBuf)
	// Fill the cell with the edit style background
	for i := 0; i < width; i++ {
		buf.Set(x+i, y, ' ', editStyle)
	}
	// Write edit buffer characters
	for i := 0; i < width && i < len(runes); i++ {
		style := editStyle
		if i == t.editCursor {
			// Cursor character: use non-reversed (normal) to make it stand out
			style = baseStyle.Bold(true)
		}
		buf.Set(x+i, y, runes[i], style)
	}
	// If cursor is at or beyond the text, show cursor as a block
	if t.editCursor >= len(runes) && t.editCursor < width {
		buf.Set(x+t.editCursor, y, ' ', baseStyle.Bold(true))
	}
}

// handleEditKey processes key events while in cell edit mode.
func (t *Table) handleEditKey(key runtime.KeyMsg) runtime.HandleResult {
	runes := []rune(t.editBuf)
	switch key.Key {
	case terminal.KeyEnter:
		t.CommitEdit()
		return runtime.Handled()
	case terminal.KeyEscape:
		t.CancelEdit()
		return runtime.Handled()
	case terminal.KeyBackspace:
		if t.editCursor > 0 {
			runes = append(runes[:t.editCursor-1], runes[t.editCursor:]...)
			t.editBuf = string(runes)
			t.editCursor--
		}
		return runtime.Handled()
	case terminal.KeyDelete:
		if t.editCursor < len(runes) {
			runes = append(runes[:t.editCursor], runes[t.editCursor+1:]...)
			t.editBuf = string(runes)
		}
		return runtime.Handled()
	case terminal.KeyLeft:
		if t.editCursor > 0 {
			t.editCursor--
		}
		return runtime.Handled()
	case terminal.KeyRight:
		if t.editCursor < len(runes) {
			t.editCursor++
		}
		return runtime.Handled()
	case terminal.KeyHome:
		t.editCursor = 0
		return runtime.Handled()
	case terminal.KeyEnd:
		t.editCursor = len(runes)
		return runtime.Handled()
	case terminal.KeyRune:
		runes = append(runes[:t.editCursor], append([]rune{key.Rune}, runes[t.editCursor:]...)...)
		t.editBuf = string(runes)
		t.editCursor++
		return runtime.Handled()
	}
	return runtime.Handled() // Swallow unrecognized keys while editing
}

func (t *Table) setSelectedCol(col int) {
	if t == nil || len(t.Columns) == 0 {
		return
	}
	if col < 0 {
		col = 0
	}
	if col >= len(t.Columns) {
		col = len(t.Columns) - 1
	}
	t.selectedCol = col
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
	prev := t.selected
	t.selected = index
	t.syncA11y()
	if prev != index {
		if announcer := t.services.Announcer(); announcer != nil {
			announcer.Announce(fmt.Sprintf("row %d selected", index+1), accessibility.PriorityPolite)
		}
	}
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
	// Build description with column header names
	colHeaders := make([]string, 0, len(t.Columns))
	for _, col := range t.Columns {
		title := strings.TrimSpace(col.Title)
		if title != "" {
			colHeaders = append(colHeaders, title)
		}
	}
	if len(colHeaders) > 0 {
		t.Base.Description = fmt.Sprintf("%d rows, %d columns: %s", t.rowCount(), len(t.Columns), strings.Join(colHeaders, ", "))
	} else {
		t.Base.Description = fmt.Sprintf("%d rows, %d columns", t.rowCount(), len(t.Columns))
	}
	// Set sort direction on the sorted column
	if t.sortState.Direction != SortNone && t.sortState.Column >= 0 && t.sortState.Column < len(t.Columns) {
		switch t.sortState.Direction {
		case SortAsc:
			t.Base.Sort = "ascending"
		case SortDesc:
			t.Base.Sort = "descending"
		}
	} else {
		t.Base.Sort = ""
	}
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

// SetOnCellEdit sets the callback invoked when a cell edit is committed.
// The callback receives the display row, column, old value, and new value.
func (t *Table) SetOnCellEdit(fn func(row, col int, oldValue, newValue string)) {
	if t == nil {
		return
	}
	t.onCellEdit = fn
}

// Editing reports whether the table is currently in cell edit mode.
func (t *Table) Editing() bool {
	if t == nil {
		return false
	}
	return t.editing
}

// EditRow returns the row index being edited (display index).
func (t *Table) EditRow() int {
	if t == nil {
		return 0
	}
	return t.editRow
}

// EditCol returns the column index being edited.
func (t *Table) EditCol() int {
	if t == nil {
		return 0
	}
	return t.editCol
}

// EditBuffer returns the current edit buffer contents.
func (t *Table) EditBuffer() string {
	if t == nil {
		return ""
	}
	return t.editBuf
}

// EditCursorPos returns the cursor position within the edit buffer.
func (t *Table) EditCursorPos() int {
	if t == nil {
		return 0
	}
	return t.editCursor
}

// SelectedCol returns the currently selected column index.
func (t *Table) SelectedCol() int {
	if t == nil {
		return 0
	}
	return t.selectedCol
}

// SetSelectedCol updates the selected column index.
func (t *Table) SetSelectedCol(col int) {
	if t == nil {
		return
	}
	if col < 0 {
		col = 0
	}
	if col >= len(t.Columns) && len(t.Columns) > 0 {
		col = len(t.Columns) - 1
	}
	t.selectedCol = col
}

// StartEdit enters edit mode on the currently selected cell.
// The edit buffer is populated with the current cell value.
func (t *Table) StartEdit() {
	if t == nil || t.editing {
		return
	}
	if t.rowCount() == 0 || len(t.Columns) == 0 {
		return
	}
	t.editing = true
	t.editRow = t.selected
	t.editCol = t.selectedCol
	t.editBuf = t.GetCell(t.editRow, t.editCol)
	t.editCursor = len([]rune(t.editBuf))
	if announcer := t.services.Announcer(); announcer != nil {
		announcer.Announce(
			fmt.Sprintf("Editing cell row %d column %d", t.editRow+1, t.editCol+1),
			accessibility.PriorityPolite,
		)
	}
}

// CancelEdit discards the current edit and exits edit mode.
func (t *Table) CancelEdit() {
	if t == nil || !t.editing {
		return
	}
	t.editing = false
	t.editBuf = ""
	t.editCursor = 0
	if announcer := t.services.Announcer(); announcer != nil {
		announcer.Announce("Edit cancelled", accessibility.PriorityPolite)
	}
}

// CommitEdit saves the current edit and exits edit mode.
// If the data source implements TabularEditable, SetCell is called.
// Otherwise, for static Rows, the cell is updated directly.
// The onCellEdit callback (if set) is invoked with old and new values.
func (t *Table) CommitEdit() {
	if t == nil || !t.editing {
		return
	}
	oldValue := t.GetCell(t.editRow, t.editCol)
	newValue := t.editBuf
	t.SetCell(t.editRow, t.editCol, newValue)
	t.editing = false
	t.editBuf = ""
	t.editCursor = 0
	if t.onCellEdit != nil {
		t.onCellEdit(t.editRow, t.editCol, oldValue, newValue)
	}
	if announcer := t.services.Announcer(); announcer != nil {
		announcer.Announce("Edit complete", accessibility.PriorityPolite)
	}
}

var _ runtime.Widget = (*Table)(nil)
var _ runtime.Focusable = (*Table)(nil)
var _ runtime.Bindable = (*Table)(nil)
var _ runtime.Unbindable = (*Table)(nil)
