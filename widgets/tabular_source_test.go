package widgets

import (
	"testing"

	"github.com/odvcencio/fluffyui/runtime"
	flufftest "github.com/odvcencio/fluffyui/testing"
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

func TestDataGridDataSourceEditable(t *testing.T) {
	source := &testTabularSource{rows: [][]string{{"Alpha", "1"}}}
	grid := NewDataGrid(TableColumn{Title: "Name"}, TableColumn{Title: "Value"})
	grid.SetDataSource(source)
	grid.SetSelected(0, 1)

	grid.StartEditing()
	grid.editor.SetText("42")
	grid.CommitEditing()

	if source.rows[0][1] != "42" {
		t.Fatalf("expected data source updated, got %q", source.rows[0][1])
	}
	if source.setCalls == 0 {
		t.Fatalf("expected SetCell to be called")
	}
}

func TestDataGridLargeDataSourceRender(t *testing.T) {
	source := &testTabularSource{rows: make([][]string, 10000)}
	for i := range source.rows {
		source.rows[i] = []string{"Row", "Value"}
	}
	grid := NewDataGrid(TableColumn{Title: "Col1"}, TableColumn{Title: "Col2"})
	grid.SetDataSource(source)

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

