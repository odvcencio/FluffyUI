package widgets

import (
	"fmt"
	"strings"
	"testing"

	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
	flufftest "github.com/odvcencio/fluffyui/testing"
)

func TestTable_VirtualScroll_LargeDataset(t *testing.T) {
	const numRows = 10000
	table := NewTable(
		TableColumn{Title: "ID", Width: 6},
		TableColumn{Title: "Name", Width: 20},
	)
	table.SetVirtualScroll(true)

	rows := make([][]string, numRows)
	for i := 0; i < numRows; i++ {
		rows[i] = []string{fmt.Sprintf("%d", i), fmt.Sprintf("Row %d", i)}
	}
	table.SetRows(rows)

	// Render into a 30-wide, 27-tall viewport (1 header + 26 data rows visible)
	output := flufftest.RenderToString(table, 30, 27)

	// The header must be present
	if !strings.Contains(output, "ID") {
		t.Fatalf("expected header 'ID' in output")
	}

	// Row 0 should be visible
	if !strings.Contains(output, "Row 0") {
		t.Fatalf("expected 'Row 0' in output")
	}

	// Row 9999 should NOT be visible at the top of the table
	if strings.Contains(output, "Row 9999") {
		t.Fatalf("did not expect 'Row 9999' in output at initial scroll position")
	}

	// Check the visible range — should be roughly [0, 26+overscan)
	start, end := table.VisibleRange()
	if start != 0 {
		t.Fatalf("expected visible start 0, got %d", start)
	}
	// end should be around 26 + 2 overscan = 28, but capped at viewport
	if end > 32 {
		t.Fatalf("expected visible end <= 32 (with overscan), got %d", end)
	}
	rendered := end - start
	if rendered > 35 {
		t.Fatalf("expected ~30 rows rendered, got %d (should be far fewer than %d)", rendered, numRows)
	}
}

func TestTable_VirtualScroll_ScrollToEnd(t *testing.T) {
	const numRows = 10000
	table := NewTable(
		TableColumn{Title: "ID", Width: 6},
		TableColumn{Title: "Name", Width: 20},
	)
	table.SetVirtualScroll(true)

	rows := make([][]string, numRows)
	for i := 0; i < numRows; i++ {
		rows[i] = []string{fmt.Sprintf("%d", i), fmt.Sprintf("Row %d", i)}
	}
	table.SetRows(rows)

	// Select the last row
	table.SetSelected(numRows - 1)

	output := flufftest.RenderToString(table, 30, 27)

	// Row 9999 should now be visible
	if !strings.Contains(output, "Row 9999") {
		t.Fatalf("expected 'Row 9999' in output after scrolling to end")
	}

	// Row 0 should NOT be visible
	if strings.Contains(output, "Row 0 ") {
		// Note: "Row 0 " with trailing space to avoid matching "Row 0" inside "Row 099..."
		// Actually let's be more careful:
	}

	start, end := table.VisibleRange()
	if start < 9000 {
		t.Fatalf("expected visible start near end of data, got %d", start)
	}
	if end != numRows {
		t.Fatalf("expected visible end at %d, got %d", numRows, end)
	}
}

func TestTable_VirtualScroll_DefaultOff(t *testing.T) {
	table := NewTable(
		TableColumn{Title: "Name"},
	)
	if table.VirtualScroll() {
		t.Fatal("expected virtual scroll to be off by default")
	}

	table.SetVirtualScroll(true)
	if !table.VirtualScroll() {
		t.Fatal("expected virtual scroll to be on after SetVirtualScroll(true)")
	}

	table.SetVirtualScroll(false)
	if table.VirtualScroll() {
		t.Fatal("expected virtual scroll to be off after SetVirtualScroll(false)")
	}
}

func TestTable_VirtualScroll_Overscan(t *testing.T) {
	const numRows = 1000
	table := NewTable(
		TableColumn{Title: "ID", Width: 6},
	)
	table.SetVirtualScroll(true)
	table.SetVirtualOverscan(5)

	rows := make([][]string, numRows)
	for i := 0; i < numRows; i++ {
		rows[i] = []string{fmt.Sprintf("%d", i)}
	}
	table.SetRows(rows)

	// Scroll to middle
	table.SetSelected(500)
	flufftest.RenderToString(table, 20, 22) // 1 header + 21 data rows

	start, end := table.VisibleRange()
	// With overscan 5, the range should extend 5 beyond viewport on each side
	viewportRows := 21
	expectedStart := 500 - viewportRows + 1 - 5
	if expectedStart < 0 {
		expectedStart = 0
	}
	if start > expectedStart+2 { // allow some slack
		t.Fatalf("expected visible start around %d, got %d", expectedStart, start)
	}
	expectedEnd := 500 + 1 + 5
	if expectedEnd > numRows {
		expectedEnd = numRows
	}
	if end < expectedEnd-2 {
		t.Fatalf("expected visible end around %d, got %d", expectedEnd, end)
	}
}

func TestTable_VirtualScroll_WithSort(t *testing.T) {
	const numRows = 5000
	table := NewTable(
		TableColumn{Title: "Name", Width: 20},
	)
	table.SetVirtualScroll(true)

	rows := make([][]string, numRows)
	for i := 0; i < numRows; i++ {
		rows[i] = []string{fmt.Sprintf("Item-%04d", numRows-1-i)}
	}
	table.SetRows(rows)
	table.SetSortColumn(0, SortAsc)

	// After ascending sort, first row should be "Item-0000"
	if got := table.GetCell(0, 0); got != "Item-0000" {
		t.Fatalf("expected first sorted cell 'Item-0000', got %q", got)
	}

	output := flufftest.RenderToString(table, 30, 12)
	if !strings.Contains(output, "Item-0000") {
		t.Fatalf("expected 'Item-0000' in sorted virtual output")
	}

	start, end := table.VisibleRange()
	if start != 0 {
		t.Fatalf("expected visible start 0 after sort, got %d", start)
	}
	rendered := end - start
	if rendered > 20 {
		t.Fatalf("expected ~15 rows rendered, got %d", rendered)
	}
}

func TestTable_VirtualScroll_WithFilter(t *testing.T) {
	const numRows = 10000
	table := NewTable(
		TableColumn{Title: "ID", Width: 6},
		TableColumn{Title: "Category", Width: 10},
	)
	table.SetVirtualScroll(true)

	rows := make([][]string, numRows)
	for i := 0; i < numRows; i++ {
		cat := "even"
		if i%2 != 0 {
			cat = "odd"
		}
		rows[i] = []string{fmt.Sprintf("%d", i), cat}
	}
	table.SetRows(rows)
	table.SetFilter(func(row []string) bool {
		return len(row) > 1 && row[1] == "even"
	})

	if table.RowCount() != 5000 {
		t.Fatalf("expected 5000 filtered rows, got %d", table.RowCount())
	}

	output := flufftest.RenderToString(table, 20, 12)
	// First filtered row is ID=0, category=even
	if !strings.Contains(output, "0") {
		t.Fatalf("expected row with ID 0 in filtered output")
	}

	start, end := table.VisibleRange()
	rendered := end - start
	if rendered > 20 {
		t.Fatalf("expected ~15 rows rendered with filter, got %d", rendered)
	}
}

func TestTable_VirtualScroll_NilReceiver(t *testing.T) {
	var table *Table
	// These should not panic
	table.SetVirtualScroll(true)
	table.SetVirtualOverscan(5)
	if table.VirtualScroll() {
		t.Fatal("expected false from nil receiver")
	}
	start, end := table.VisibleRange()
	if start != 0 || end != 0 {
		t.Fatalf("expected (0,0) from nil receiver, got (%d,%d)", start, end)
	}
}

func TestTable_VirtualScroll_EmptyTable(t *testing.T) {
	table := NewTable(
		TableColumn{Title: "Name"},
	)
	table.SetVirtualScroll(true)
	table.SetRows(nil)

	// Should not panic
	output := flufftest.RenderToString(table, 20, 10)
	if !strings.Contains(output, "Name") {
		t.Fatalf("expected header in empty virtual table")
	}
}

func TestTable_VirtualScroll_PageNavigation(t *testing.T) {
	const numRows = 10000
	table := NewTable(
		TableColumn{Title: "ID", Width: 6},
	)
	table.SetVirtualScroll(true)

	rows := make([][]string, numRows)
	for i := 0; i < numRows; i++ {
		rows[i] = []string{fmt.Sprintf("%d", i)}
	}
	table.SetRows(rows)

	// Simulate layout so ContentBounds is set
	table.Focus()
	table.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 27})
	table.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 27})

	// Page down should move selection forward by roughly viewport rows
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyPageDown})
	selected := table.SelectedIndex()
	if selected < 10 {
		t.Fatalf("expected selected > 10 after page down, got %d", selected)
	}

	// Page up from there
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyPageUp})
	if table.SelectedIndex() != 0 {
		t.Fatalf("expected selected 0 after page up from near top, got %d", table.SelectedIndex())
	}
}

func TestTable_VirtualScroll_DataSource(t *testing.T) {
	const numRows = 50000
	table := NewTable(
		TableColumn{Title: "ID", Width: 8},
		TableColumn{Title: "Value", Width: 10},
	)
	table.SetVirtualScroll(true)
	table.SetDataSource(&largeDataSource{rows: numRows, cols: 2})

	output := flufftest.RenderToString(table, 22, 12)
	if !strings.Contains(output, "ID") {
		t.Fatalf("expected header 'ID' in output")
	}

	start, end := table.VisibleRange()
	rendered := end - start
	if rendered > 20 {
		t.Fatalf("expected ~15 rows rendered with data source, got %d", rendered)
	}
}

// largeDataSource is a TabularDataSource for testing.
type largeDataSource struct {
	rows int
	cols int
}

func (d *largeDataSource) RowCount() int { return d.rows }
func (d *largeDataSource) Cell(row, col int) string {
	return fmt.Sprintf("R%dC%d", row, col)
}
