package widgets

import (
	"strings"
	"testing"

	flufftest "m31labs.dev/fluffyui/testing"
)

func TestTable_SortAsc(t *testing.T) {
	table := NewTable(
		TableColumn{Title: "Name"},
		TableColumn{Title: "Value"},
	)
	table.SetRows([][]string{
		{"Charlie", "3"},
		{"Alpha", "1"},
		{"Bravo", "2"},
	})

	table.SetSortColumn(0, SortAsc)

	if table.RowCount() != 3 {
		t.Fatalf("expected 3 rows, got %d", table.RowCount())
	}
	// After ascending sort on column 0: Alpha, Bravo, Charlie
	if got := table.GetCell(0, 0); got != "Alpha" {
		t.Fatalf("expected first row 'Alpha', got %q", got)
	}
	if got := table.GetCell(1, 0); got != "Bravo" {
		t.Fatalf("expected second row 'Bravo', got %q", got)
	}
	if got := table.GetCell(2, 0); got != "Charlie" {
		t.Fatalf("expected third row 'Charlie', got %q", got)
	}
	// Verify corresponding values follow
	if got := table.GetCell(0, 1); got != "1" {
		t.Fatalf("expected first row value '1', got %q", got)
	}
	if got := table.GetCell(2, 1); got != "3" {
		t.Fatalf("expected third row value '3', got %q", got)
	}
}

func TestTable_SortDesc(t *testing.T) {
	table := NewTable(
		TableColumn{Title: "Name"},
		TableColumn{Title: "Value"},
	)
	table.SetRows([][]string{
		{"Alpha", "1"},
		{"Charlie", "3"},
		{"Bravo", "2"},
	})

	table.SetSortColumn(0, SortDesc)

	if table.RowCount() != 3 {
		t.Fatalf("expected 3 rows, got %d", table.RowCount())
	}
	// After descending sort on column 0: Charlie, Bravo, Alpha
	if got := table.GetCell(0, 0); got != "Charlie" {
		t.Fatalf("expected first row 'Charlie', got %q", got)
	}
	if got := table.GetCell(1, 0); got != "Bravo" {
		t.Fatalf("expected second row 'Bravo', got %q", got)
	}
	if got := table.GetCell(2, 0); got != "Alpha" {
		t.Fatalf("expected third row 'Alpha', got %q", got)
	}
}

func TestTable_SortNone(t *testing.T) {
	table := NewTable(
		TableColumn{Title: "Name"},
	)
	table.SetRows([][]string{
		{"Charlie"},
		{"Alpha"},
		{"Bravo"},
	})

	// Apply sort then clear it
	table.SetSortColumn(0, SortAsc)
	table.SetSortColumn(0, SortNone)

	if table.RowCount() != 3 {
		t.Fatalf("expected 3 rows, got %d", table.RowCount())
	}
	// Original order restored
	if got := table.GetCell(0, 0); got != "Charlie" {
		t.Fatalf("expected first row 'Charlie', got %q", got)
	}
	if got := table.GetCell(1, 0); got != "Alpha" {
		t.Fatalf("expected second row 'Alpha', got %q", got)
	}
	if got := table.GetCell(2, 0); got != "Bravo" {
		t.Fatalf("expected third row 'Bravo', got %q", got)
	}

	// Sort state should reflect SortNone
	state := table.Sort()
	if state.Direction != SortNone {
		t.Fatalf("expected SortNone direction, got %d", state.Direction)
	}
}

func TestTable_Filter(t *testing.T) {
	table := NewTable(
		TableColumn{Title: "Name"},
		TableColumn{Title: "Value"},
	)
	table.SetRows([][]string{
		{"Apple", "fruit"},
		{"Broccoli", "vegetable"},
		{"Banana", "fruit"},
		{"Carrot", "vegetable"},
	})

	// Filter to only fruits
	table.SetFilter(func(row []string) bool {
		return len(row) > 1 && row[1] == "fruit"
	})

	if table.RowCount() != 2 {
		t.Fatalf("expected 2 filtered rows, got %d", table.RowCount())
	}
	if got := table.GetCell(0, 0); got != "Apple" {
		t.Fatalf("expected first filtered row 'Apple', got %q", got)
	}
	if got := table.GetCell(1, 0); got != "Banana" {
		t.Fatalf("expected second filtered row 'Banana', got %q", got)
	}

	// SelectedRow should return filtered data
	table.SetSelected(1)
	row := table.SelectedRow()
	if row == nil || row[0] != "Banana" {
		t.Fatalf("expected selected row 'Banana', got %v", row)
	}
}

func TestTable_FilterAndSort(t *testing.T) {
	table := NewTable(
		TableColumn{Title: "Name"},
		TableColumn{Title: "Category"},
	)
	table.SetRows([][]string{
		{"Banana", "fruit"},
		{"Broccoli", "vegetable"},
		{"Apple", "fruit"},
		{"Carrot", "vegetable"},
		{"Cherry", "fruit"},
	})

	// Filter to fruits, then sort ascending by name
	table.SetFilter(func(row []string) bool {
		return len(row) > 1 && row[1] == "fruit"
	})
	table.SetSortColumn(0, SortAsc)

	if table.RowCount() != 3 {
		t.Fatalf("expected 3 filtered rows, got %d", table.RowCount())
	}
	// Sorted fruits: Apple, Banana, Cherry
	if got := table.GetCell(0, 0); got != "Apple" {
		t.Fatalf("expected first row 'Apple', got %q", got)
	}
	if got := table.GetCell(1, 0); got != "Banana" {
		t.Fatalf("expected second row 'Banana', got %q", got)
	}
	if got := table.GetCell(2, 0); got != "Cherry" {
		t.Fatalf("expected third row 'Cherry', got %q", got)
	}

	// Render should work with filtered+sorted data
	output := flufftest.RenderToString(table, 30, 6)
	if !strings.Contains(output, "Apple") {
		t.Fatalf("expected render output to contain 'Apple', got:\n%s", output)
	}
	if strings.Contains(output, "Broccoli") {
		t.Fatalf("expected render output to NOT contain 'Broccoli', got:\n%s", output)
	}
}

func TestTable_ClearFilter(t *testing.T) {
	table := NewTable(
		TableColumn{Title: "Name"},
		TableColumn{Title: "Value"},
	)
	table.SetRows([][]string{
		{"Alpha", "1"},
		{"Bravo", "2"},
		{"Charlie", "3"},
	})

	// Apply filter
	table.SetFilter(func(row []string) bool {
		return row[0] == "Bravo"
	})
	if table.RowCount() != 1 {
		t.Fatalf("expected 1 filtered row, got %d", table.RowCount())
	}

	// Clear filter
	table.ClearFilter()
	if table.RowCount() != 3 {
		t.Fatalf("expected 3 rows after clearing filter, got %d", table.RowCount())
	}
	// Original order restored
	if got := table.GetCell(0, 0); got != "Alpha" {
		t.Fatalf("expected first row 'Alpha', got %q", got)
	}
	if got := table.GetCell(2, 0); got != "Charlie" {
		t.Fatalf("expected third row 'Charlie', got %q", got)
	}
}

func TestTable_SetSortColumn(t *testing.T) {
	table := NewTable(
		TableColumn{Title: "Name"},
		TableColumn{Title: "Value"},
	)
	table.SetRows([][]string{
		{"Zebra", "26"},
		{"Apple", "1"},
		{"Mango", "13"},
	})

	// Sort by column 1 ascending (lexicographic: "1", "13", "26")
	table.SetSortColumn(1, SortAsc)

	state := table.Sort()
	if state.Column != 1 {
		t.Fatalf("expected sort column 1, got %d", state.Column)
	}
	if state.Direction != SortAsc {
		t.Fatalf("expected SortAsc, got %d", state.Direction)
	}
	if got := table.GetCell(0, 1); got != "1" {
		t.Fatalf("expected first value '1', got %q", got)
	}
	if got := table.GetCell(1, 1); got != "13" {
		t.Fatalf("expected second value '13', got %q", got)
	}
	if got := table.GetCell(2, 1); got != "26" {
		t.Fatalf("expected third value '26', got %q", got)
	}

	// Switch to descending
	table.SetSortColumn(1, SortDesc)

	state = table.Sort()
	if state.Direction != SortDesc {
		t.Fatalf("expected SortDesc, got %d", state.Direction)
	}
	if got := table.GetCell(0, 1); got != "26" {
		t.Fatalf("expected first value '26' in desc, got %q", got)
	}
	if got := table.GetCell(2, 1); got != "1" {
		t.Fatalf("expected third value '1' in desc, got %q", got)
	}

	// Verify SelectedRow works after sort
	table.SetSelected(0)
	row := table.SelectedRow()
	if row == nil || row[0] != "Zebra" {
		t.Fatalf("expected selected row 'Zebra', got %v", row)
	}

	// Render should succeed
	output := flufftest.RenderToString(table, 30, 5)
	if !strings.Contains(output, "Zebra") {
		t.Fatalf("expected render output to contain 'Zebra', got:\n%s", output)
	}
}

func TestTable_SortNilReceiver(t *testing.T) {
	// Verify nil receiver safety for new methods.
	var table *Table
	table.SetSortColumn(0, SortAsc)
	table.SetFilter(func(row []string) bool { return true })
	table.ClearFilter()
	state := table.Sort()
	if state.Direction != SortNone {
		t.Fatalf("expected SortNone from nil table, got %d", state.Direction)
	}
}

func TestTable_FilterEmptyResult(t *testing.T) {
	table := NewTable(
		TableColumn{Title: "Name"},
	)
	table.SetRows([][]string{
		{"Alpha"},
		{"Bravo"},
	})

	// Filter that matches nothing
	table.SetFilter(func(row []string) bool {
		return false
	})

	if table.RowCount() != 0 {
		t.Fatalf("expected 0 rows, got %d", table.RowCount())
	}
	if got := table.GetCell(0, 0); got != "" {
		t.Fatalf("expected empty cell for out-of-range, got %q", got)
	}
	if row := table.SelectedRow(); row != nil {
		t.Fatalf("expected nil selected row, got %v", row)
	}
}

func TestTable_SetRowsResetsSort(t *testing.T) {
	table := NewTable(
		TableColumn{Title: "Name"},
	)
	table.SetRows([][]string{
		{"Charlie"},
		{"Alpha"},
	})
	table.SetSortColumn(0, SortAsc)

	// Verify sort is applied
	if got := table.GetCell(0, 0); got != "Alpha" {
		t.Fatalf("expected 'Alpha' after sort, got %q", got)
	}

	// SetRows should rebuild indices with existing sort
	table.SetRows([][]string{
		{"Zulu"},
		{"Delta"},
		{"Foxtrot"},
	})

	// Sort state persists, so new data should be sorted
	if table.RowCount() != 3 {
		t.Fatalf("expected 3 rows, got %d", table.RowCount())
	}
	if got := table.GetCell(0, 0); got != "Delta" {
		t.Fatalf("expected 'Delta' first after SetRows with sort, got %q", got)
	}
	if got := table.GetCell(1, 0); got != "Foxtrot" {
		t.Fatalf("expected 'Foxtrot' second after SetRows with sort, got %q", got)
	}
	if got := table.GetCell(2, 0); got != "Zulu" {
		t.Fatalf("expected 'Zulu' third after SetRows with sort, got %q", got)
	}
}
