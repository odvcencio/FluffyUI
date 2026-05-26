package widgets

import (
	"strings"
	"testing"

	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/terminal"
	flufftest "m31labs.dev/fluffyui/testing"
)

func newEditableTable() *Table {
	table := NewTable(
		TableColumn{Title: "Name", Width: 10},
		TableColumn{Title: "Value", Width: 10},
	)
	table.SetRows([][]string{
		{"Alice", "100"},
		{"Bob", "200"},
		{"Charlie", "300"},
	})
	table.Focus()
	return table
}

func TestTable_StartEdit_PopulatesBuffer(t *testing.T) {
	table := newEditableTable()
	table.SetSelected(1)
	table.SetSelectedCol(0)

	table.StartEdit()

	if !table.Editing() {
		t.Fatal("expected table to be in editing mode")
	}
	if table.EditRow() != 1 {
		t.Fatalf("expected edit row 1, got %d", table.EditRow())
	}
	if table.EditCol() != 0 {
		t.Fatalf("expected edit col 0, got %d", table.EditCol())
	}
	if table.EditBuffer() != "Bob" {
		t.Fatalf("expected edit buffer 'Bob', got %q", table.EditBuffer())
	}
	if table.EditCursorPos() != 3 {
		t.Fatalf("expected cursor at 3, got %d", table.EditCursorPos())
	}
}

func TestTable_StartEdit_SecondColumn(t *testing.T) {
	table := newEditableTable()
	table.SetSelected(0)
	table.SetSelectedCol(1)

	table.StartEdit()

	if table.EditBuffer() != "100" {
		t.Fatalf("expected edit buffer '100', got %q", table.EditBuffer())
	}
	if table.EditCol() != 1 {
		t.Fatalf("expected edit col 1, got %d", table.EditCol())
	}
}

func TestTable_TypeCharacters(t *testing.T) {
	table := newEditableTable()
	table.SetSelected(0)
	table.SetSelectedCol(0)
	table.StartEdit()

	// Move cursor to end (already there), type "X"
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'X'})

	if table.EditBuffer() != "AliceX" {
		t.Fatalf("expected 'AliceX', got %q", table.EditBuffer())
	}
	if table.EditCursorPos() != 6 {
		t.Fatalf("expected cursor at 6, got %d", table.EditCursorPos())
	}
}

func TestTable_TypeAtMiddle(t *testing.T) {
	table := newEditableTable()
	table.SetSelected(0)
	table.SetSelectedCol(0)
	table.StartEdit()

	// Move cursor to position 2
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyHome})
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})

	// Insert 'Z' at position 2
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'Z'})

	if table.EditBuffer() != "AlZice" {
		t.Fatalf("expected 'AlZice', got %q", table.EditBuffer())
	}
	if table.EditCursorPos() != 3 {
		t.Fatalf("expected cursor at 3, got %d", table.EditCursorPos())
	}
}

func TestTable_Backspace(t *testing.T) {
	table := newEditableTable()
	table.SetSelected(0)
	table.SetSelectedCol(0)
	table.StartEdit()

	// Backspace at end removes last char
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyBackspace})

	if table.EditBuffer() != "Alic" {
		t.Fatalf("expected 'Alic', got %q", table.EditBuffer())
	}
	if table.EditCursorPos() != 4 {
		t.Fatalf("expected cursor at 4, got %d", table.EditCursorPos())
	}
}

func TestTable_BackspaceAtStart(t *testing.T) {
	table := newEditableTable()
	table.SetSelected(0)
	table.SetSelectedCol(0)
	table.StartEdit()

	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyHome})
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyBackspace})

	// Should be no-op
	if table.EditBuffer() != "Alice" {
		t.Fatalf("expected 'Alice', got %q", table.EditBuffer())
	}
	if table.EditCursorPos() != 0 {
		t.Fatalf("expected cursor at 0, got %d", table.EditCursorPos())
	}
}

func TestTable_Delete(t *testing.T) {
	table := newEditableTable()
	table.SetSelected(0)
	table.SetSelectedCol(0)
	table.StartEdit()

	// Move to start, delete first char
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyHome})
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDelete})

	if table.EditBuffer() != "lice" {
		t.Fatalf("expected 'lice', got %q", table.EditBuffer())
	}
	if table.EditCursorPos() != 0 {
		t.Fatalf("expected cursor at 0, got %d", table.EditCursorPos())
	}
}

func TestTable_DeleteAtEnd(t *testing.T) {
	table := newEditableTable()
	table.SetSelected(0)
	table.SetSelectedCol(0)
	table.StartEdit()

	// Delete at end should be no-op
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDelete})

	if table.EditBuffer() != "Alice" {
		t.Fatalf("expected 'Alice', got %q", table.EditBuffer())
	}
}

func TestTable_CursorMovement(t *testing.T) {
	table := newEditableTable()
	table.SetSelected(0)
	table.SetSelectedCol(0)
	table.StartEdit() // cursor at 5 (end of "Alice")

	// Left
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})
	if table.EditCursorPos() != 4 {
		t.Fatalf("expected cursor at 4 after Left, got %d", table.EditCursorPos())
	}

	// Right
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if table.EditCursorPos() != 5 {
		t.Fatalf("expected cursor at 5 after Right, got %d", table.EditCursorPos())
	}

	// Right again at end should stay
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if table.EditCursorPos() != 5 {
		t.Fatalf("expected cursor at 5 (clamped), got %d", table.EditCursorPos())
	}

	// Home
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyHome})
	if table.EditCursorPos() != 0 {
		t.Fatalf("expected cursor at 0 after Home, got %d", table.EditCursorPos())
	}

	// Left at start should stay
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})
	if table.EditCursorPos() != 0 {
		t.Fatalf("expected cursor at 0 (clamped), got %d", table.EditCursorPos())
	}

	// End
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnd})
	if table.EditCursorPos() != 5 {
		t.Fatalf("expected cursor at 5 after End, got %d", table.EditCursorPos())
	}
}

func TestTable_CommitEdit_CallsCallback(t *testing.T) {
	table := newEditableTable()
	table.SetSelected(1)
	table.SetSelectedCol(1)

	var cbRow, cbCol int
	var cbOld, cbNew string
	table.SetOnCellEdit(func(row, col int, old, new string) {
		cbRow = row
		cbCol = col
		cbOld = old
		cbNew = new
	})

	table.StartEdit() // edit "200"
	// Clear and type "250"
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyHome})
	// Select all by deleting 3 chars
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDelete})
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDelete})
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDelete})
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: '2'})
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: '5'})
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: '0'})

	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})

	if table.Editing() {
		t.Fatal("expected editing to be false after commit")
	}
	if cbRow != 1 || cbCol != 1 {
		t.Fatalf("expected callback row=1, col=1, got row=%d, col=%d", cbRow, cbCol)
	}
	if cbOld != "200" {
		t.Fatalf("expected old value '200', got %q", cbOld)
	}
	if cbNew != "250" {
		t.Fatalf("expected new value '250', got %q", cbNew)
	}
	// Verify cell was actually updated
	if got := table.GetCell(1, 1); got != "250" {
		t.Fatalf("expected cell to be '250', got %q", got)
	}
}

func TestTable_CancelEdit_RestoresValue(t *testing.T) {
	table := newEditableTable()
	table.SetSelected(0)
	table.SetSelectedCol(0)

	table.StartEdit() // edit "Alice"
	// Type some chars
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'X'})
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'Y'})

	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})

	if table.Editing() {
		t.Fatal("expected editing to be false after cancel")
	}
	// Cell should retain original value since cancel doesn't write
	if got := table.GetCell(0, 0); got != "Alice" {
		t.Fatalf("expected cell to remain 'Alice', got %q", got)
	}
	if table.EditBuffer() != "" {
		t.Fatalf("expected edit buffer to be cleared, got %q", table.EditBuffer())
	}
}

func TestTable_EnterKey_StartsEdit(t *testing.T) {
	table := newEditableTable()
	table.SetSelected(0)
	table.SetSelectedCol(0)

	result := table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if !result.Handled {
		t.Fatal("expected Enter to be handled")
	}
	if !table.Editing() {
		t.Fatal("expected Enter to start editing")
	}
}

func TestTable_F2Key_StartsEdit(t *testing.T) {
	table := newEditableTable()
	table.SetSelected(0)
	table.SetSelectedCol(0)

	result := table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyF2})
	if !result.Handled {
		t.Fatal("expected F2 to be handled")
	}
	if !table.Editing() {
		t.Fatal("expected F2 to start editing")
	}
}

func TestTable_EscapeKey_CancelsEdit(t *testing.T) {
	table := newEditableTable()
	table.StartEdit()

	result := table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})
	if !result.Handled {
		t.Fatal("expected Escape to be handled")
	}
	if table.Editing() {
		t.Fatal("expected Escape to cancel editing")
	}
}

func TestTable_EnterKey_CommitsEdit(t *testing.T) {
	table := newEditableTable()
	table.SetSelected(0)
	table.SetSelectedCol(0)
	table.StartEdit()

	// Type something
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: '!'})

	result := table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if !result.Handled {
		t.Fatal("expected Enter to be handled in edit mode")
	}
	if table.Editing() {
		t.Fatal("expected Enter to commit edit")
	}
	if got := table.GetCell(0, 0); got != "Alice!" {
		t.Fatalf("expected 'Alice!', got %q", got)
	}
}

func TestTable_NavigationWhileNotEditing(t *testing.T) {
	table := newEditableTable()

	// Left/Right should change selected column (not start editing)
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if table.SelectedCol() != 1 {
		t.Fatalf("expected selected col 1, got %d", table.SelectedCol())
	}

	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})
	if table.SelectedCol() != 0 {
		t.Fatalf("expected selected col 0, got %d", table.SelectedCol())
	}

	// Should clamp at 0
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})
	if table.SelectedCol() != 0 {
		t.Fatalf("expected selected col clamped at 0, got %d", table.SelectedCol())
	}
}

func TestTable_SelectedCol_Clamping(t *testing.T) {
	table := newEditableTable()

	table.SetSelectedCol(-5)
	if table.SelectedCol() != 0 {
		t.Fatalf("expected col clamped to 0, got %d", table.SelectedCol())
	}

	table.SetSelectedCol(999)
	if table.SelectedCol() != 1 {
		t.Fatalf("expected col clamped to 1, got %d", table.SelectedCol())
	}
}

func TestTable_EditMode_SwallowsUnrecognizedKeys(t *testing.T) {
	table := newEditableTable()
	table.StartEdit()

	// PageUp should be swallowed (not navigate) while editing
	result := table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyPageUp})
	if !result.Handled {
		t.Fatal("expected unrecognized key to be handled (swallowed) in edit mode")
	}
}

func TestTable_StartEdit_NoRowsNoop(t *testing.T) {
	table := NewTable(TableColumn{Title: "Name"})
	table.Focus()
	table.StartEdit()

	if table.Editing() {
		t.Fatal("expected StartEdit to be no-op with no rows")
	}
}

func TestTable_StartEdit_NoColumnsNoop(t *testing.T) {
	table := NewTable()
	table.SetRows([][]string{{"a"}})
	table.Focus()
	table.StartEdit()

	if table.Editing() {
		t.Fatal("expected StartEdit to be no-op with no columns")
	}
}

func TestTable_DoubleStartEdit_Noop(t *testing.T) {
	table := newEditableTable()
	table.StartEdit()

	// Starting edit again should be a no-op
	table.SetSelected(2)
	table.StartEdit()

	// Should still be editing row 0
	if table.EditRow() != 0 {
		t.Fatalf("expected edit row 0 (first start), got %d", table.EditRow())
	}
}

func TestTable_CommitEdit_WithDataSource(t *testing.T) {
	table := NewTable(
		TableColumn{Title: "Name", Width: 10},
		TableColumn{Title: "Value", Width: 10},
	)
	ds := &testEditableSource{
		rows: [][]string{
			{"Alpha", "1"},
			{"Bravo", "2"},
		},
	}
	table.SetDataSource(ds)
	table.Focus()
	table.SetSelected(0)
	table.SetSelectedCol(1)

	table.StartEdit()
	// Clear and type new value
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyHome})
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDelete})
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: '9'})

	table.CommitEdit()

	if ds.lastSetRow != 0 || ds.lastSetCol != 1 {
		t.Fatalf("expected SetCell(0, 1), got (%d, %d)", ds.lastSetRow, ds.lastSetCol)
	}
	if ds.lastSetVal != "9" {
		t.Fatalf("expected SetCell value '9', got %q", ds.lastSetVal)
	}
}

func TestTable_EditRender(t *testing.T) {
	table := newEditableTable()
	table.SetSelected(0)
	table.SetSelectedCol(0)
	table.StartEdit()

	output := flufftest.RenderToString(table, 25, 5)

	// The edit buffer should be rendered (it contains "Alice")
	// At minimum, we should NOT see the original cell as plain text
	// The rendering will show the edit buffer contents
	if !strings.Contains(output, "Alice") {
		t.Fatalf("expected rendered output to contain edit buffer 'Alice', got:\n%s", output)
	}
}

func TestTable_CommitEdit_WithSortedTable(t *testing.T) {
	table := NewTable(
		TableColumn{Title: "Name", Width: 10},
		TableColumn{Title: "Value", Width: 10},
	)
	table.SetRows([][]string{
		{"Charlie", "3"},
		{"Alpha", "1"},
		{"Bravo", "2"},
	})
	table.SetSortColumn(0, SortAsc)
	table.Focus()

	// After sort: display[0]=Alpha, display[1]=Bravo, display[2]=Charlie
	// Select display row 0 (Alpha), col 1 (value "1")
	table.SetSelected(0)
	table.SetSelectedCol(1)

	var cbOld, cbNew string
	table.SetOnCellEdit(func(row, col int, old, new string) {
		cbOld = old
		cbNew = new
	})

	table.StartEdit()
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyHome})
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDelete})
	table.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: '9'})
	table.CommitEdit()

	if cbOld != "1" {
		t.Fatalf("expected old value '1', got %q", cbOld)
	}
	if cbNew != "9" {
		t.Fatalf("expected new value '9', got %q", cbNew)
	}
	// Verify the underlying cell was updated
	if got := table.GetCell(0, 1); got != "9" {
		t.Fatalf("expected cell '9', got %q", got)
	}
}

func TestTable_NilReceiver_EditMethods(t *testing.T) {
	var table *Table

	// These should all be safe on nil receiver
	table.SetOnCellEdit(func(row, col int, old, new string) {})
	table.StartEdit()
	table.CancelEdit()
	table.CommitEdit()
	table.SetSelectedCol(0)

	if table.Editing() {
		t.Fatal("expected Editing() to return false on nil receiver")
	}
	if table.EditRow() != 0 {
		t.Fatal("expected EditRow() to return 0 on nil receiver")
	}
	if table.EditCol() != 0 {
		t.Fatal("expected EditCol() to return 0 on nil receiver")
	}
	if table.EditBuffer() != "" {
		t.Fatal("expected EditBuffer() to return empty on nil receiver")
	}
	if table.EditCursorPos() != 0 {
		t.Fatal("expected EditCursorPos() to return 0 on nil receiver")
	}
	if table.SelectedCol() != 0 {
		t.Fatal("expected SelectedCol() to return 0 on nil receiver")
	}
}

// testEditableSource implements TabularDataSource and TabularEditable.
type testEditableSource struct {
	rows       [][]string
	lastSetRow int
	lastSetCol int
	lastSetVal string
}

func (s *testEditableSource) RowCount() int { return len(s.rows) }
func (s *testEditableSource) Cell(row, col int) string {
	if row < 0 || row >= len(s.rows) || col < 0 || col >= len(s.rows[row]) {
		return ""
	}
	return s.rows[row][col]
}
func (s *testEditableSource) SetCell(row, col int, value string) {
	s.lastSetRow = row
	s.lastSetCol = col
	s.lastSetVal = value
	if row >= 0 && row < len(s.rows) && col >= 0 && col < len(s.rows[row]) {
		s.rows[row][col] = value
	}
}

var _ TabularEditable = (*testEditableSource)(nil)
