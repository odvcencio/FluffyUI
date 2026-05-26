package widgets

import (
	"testing"

	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/terminal"
	flufftest "m31labs.dev/fluffyui/testing"
)

type testTabularSource struct {
	rows     [][]string
	setCalls int
}

func (t *testTabularSource) RowCount() int {
	return len(t.rows)
}

func (t *testTabularSource) Cell(row, col int) string {
	if row < 0 || row >= len(t.rows) {
		return ""
	}
	if col < 0 || col >= len(t.rows[row]) {
		return ""
	}
	return t.rows[row][col]
}

func (t *testTabularSource) SetCell(row, col int, value string) {
	if row < 0 || row >= len(t.rows) {
		return
	}
	for len(t.rows[row]) <= col {
		t.rows[row] = append(t.rows[row], "")
	}
	t.rows[row][col] = value
	t.setCalls++
}

func (t *testTabularSource) Row(row int) []string {
	if row < 0 || row >= len(t.rows) {
		return nil
	}
	return t.rows[row]
}

func TestTableDataSource(t *testing.T) {
	source := &testTabularSource{rows: [][]string{{"Alpha", "1"}, {"Beta", "2"}}}
	table := NewTable(TableColumn{Title: "Name"}, TableColumn{Title: "Value"})
	table.SetDataSource(source)

	if table.RowCount() != 2 {
		t.Fatalf("expected row count 2, got %d", table.RowCount())
	}
	if got := table.GetCell(1, 1); got != "2" {
		t.Fatalf("expected cell value 2, got %q", got)
	}
	table.SetSelected(0)
	if table.SelectedRow() == nil {
		t.Fatalf("expected SelectedRow to return data from provider")
	}
}

func TestDataGridCellEditing(t *testing.T) {
	grid := NewDataGrid([]DataGridColumn{{Header: "Name", Width: 10}, {Header: "Value", Width: 10}})
	grid.SetRows([][]string{{"Alpha", "1"}})
	grid.SetSelectedCell(0, 1)
	grid.Focus()

	grid.StartEdit()
	// Clear existing "1" with backspace, then type "42"
	grid.HandleMessage(runtime.KeyMsg{Key: terminal.KeyBackspace})
	grid.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: '4'})
	grid.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: '2'})
	grid.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter}) // commits

	if grid.Cell(0, 1) != "42" {
		t.Fatalf("expected cell updated, got %q", grid.Cell(0, 1))
	}
}

func TestDataGridLargeDataSetRender(t *testing.T) {
	rows := make([][]string, 10000)
	for i := range rows {
		rows[i] = []string{"Row", "Value"}
	}
	grid := NewDataGrid([]DataGridColumn{{Header: "Col1", Width: 10}, {Header: "Col2", Width: 10}})
	grid.SetRows(rows)

	output := flufftest.RenderToString(grid, 20, 6)
	if output == "" {
		t.Fatalf("expected render output")
	}
}

func TestTableLargeDataSourceRender(t *testing.T) {
	source := &testTabularSource{rows: make([][]string, 10000)}
	for i := range source.rows {
		source.rows[i] = []string{"Row", "Value"}
	}
	table := NewTable(TableColumn{Title: "Col1"}, TableColumn{Title: "Col2"})
	table.SetDataSource(source)

	output := flufftest.RenderToString(table, 20, 6)
	if output == "" {
		t.Fatalf("expected render output")
	}
}

var _ TabularEditable = (*testTabularSource)(nil)
var _ TabularRowProvider = (*testTabularSource)(nil)
var _ runtime.Widget = (*Table)(nil)

