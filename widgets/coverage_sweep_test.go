package widgets

import (
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/graphics"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	"github.com/odvcencio/fluffyui/terminal"
	flufftest "github.com/odvcencio/fluffyui/testing"
)

type testGridSource struct {
	rows [][]string
}

func (t *testGridSource) RowCount() int { return len(t.rows) }

func (t *testGridSource) Cell(row, col int) string {
	if row < 0 || row >= len(t.rows) || col < 0 || col >= len(t.rows[row]) {
		return ""
	}
	return t.rows[row][col]
}

func (t *testGridSource) Row(row int) []string {
	if row < 0 || row >= len(t.rows) {
		return nil
	}
	return t.rows[row]
}

func (t *testGridSource) SetCell(row, col int, value string) {
	if row < 0 || row >= len(t.rows) || col < 0 || col >= len(t.rows[row]) {
		return
	}
	t.rows[row][col] = value
}

func TestCalendarOptionsAndSignals(t *testing.T) {
	now := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	min := now.AddDate(0, 0, -3)
	max := now.AddDate(0, 0, 3)

	selected := state.NewSignal(now)
	displayed := state.NewSignal(now)
	minSig := state.NewSignal[*time.Time](&min)
	maxSig := state.NewSignal[*time.Time](&max)
	highlights := state.NewSignal([]time.Time{now})

	cal := NewCalendar(
		WithSelectedDateSignal(selected),
		WithDisplayedMonthSignal(displayed),
		WithMinDateSignal(minSig),
		WithMaxDateSignal(maxSig),
		WithHighlightDatesSignal(highlights),
		WithWeekStart(time.Monday),
		WithShowWeekNumbers(true),
		WithSelectionMode(CalendarSelectionRange),
		WithDayRenderer(func(ctx runtime.RenderContext, date time.Time, state CalendarDayState) {}),
	)

	app := runtime.NewApp(runtime.AppConfig{})
	cal.Bind(app.Services())
	_ = cal.SelectedDateSignal()
	_ = cal.DisplayedMonthSignal()
	_ = cal.RangeStartSignal()
	_ = cal.RangeEndSignal()

	cal.SetSelectedDate(now)
	cal.SetRange(&min, &max)
	cal.SetDisplayedMonth(now)
	cal.SetMinDate(&min)
	cal.SetMaxDate(&max)
	cal.SetHighlightDates([]time.Time{now})
	cal.SetWeekStart(time.Sunday)
	cal.SetShowWeekNumbers(false)
	cal.SetSelectionMode(CalendarSelectionSingle)
	cal.SetDayRenderer(nil)
	cal.SetStyles(
		backend.DefaultStyle(),
		backend.DefaultStyle(),
		backend.DefaultStyle(),
		backend.DefaultStyle(),
		backend.DefaultStyle(),
		backend.DefaultStyle(),
		backend.DefaultStyle(),
		backend.DefaultStyle(),
		backend.DefaultStyle(),
	)
	_ = cal.StyleType()
	_ = monthEnd(now)

	cal.Unbind()
}

func TestDataGridEditingAndSetters(t *testing.T) {
	source := &testGridSource{rows: [][]string{{"1", "2"}, {"3", "4"}}}
	grid := NewDataGrid(TableColumn{Title: "A", Width: 4}, TableColumn{Title: "B", Width: 4})
	WithDataGridLabel("Grid")(grid)
	WithDataGridStyle(backend.DefaultStyle())(grid)
	grid.SetColumns([]TableColumn{{Title: "A", Width: 4}, {Title: "B", Width: 4}})
	grid.SetRows(source.rows)
	grid.SetDataSource(source)
	if grid.DataSource() == nil {
		t.Fatalf("expected data source")
	}
	grid.SetLabel("Grid")
	grid.SetStyle(backend.DefaultStyle())
	grid.SetHeaderStyle(backend.DefaultStyle().Bold(true))
	grid.SetSelectedStyle(backend.DefaultStyle().Reverse(true))
	grid.SetEditingStyle(backend.DefaultStyle().Reverse(true))
	grid.SetOnEdit(func(row, col int, value string) {})
	grid.SetOnSelect(func(row, col int) {})
	_ = grid.StyleType()

	grid.Bind(runtime.Services{})
	grid.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 6})
	grid.Focus()
	grid.SetSelected(0, 1)
	row, col := grid.SelectedPosition()
	if row != 0 || col != 1 {
		t.Fatalf("unexpected selected position")
	}
	grid.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	grid.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: '9'})
	_ = flufftest.RenderToString(grid, 20, 6)
	grid.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	grid.HandleMessage(runtime.MouseMsg{Button: runtime.MouseWheelDown})
	grid.HandleMessage(runtime.MouseMsg{Button: runtime.MouseLeft, Action: runtime.MousePress, X: 1, Y: 2})

	grid.StartEditing()
	grid.CancelEditing()
	grid.CommitEditing()
	_ = grid.IsEditing()
	_ = grid.ColumnCount()
	_ = grid.RowCount()
	_ = grid.DataSource()

	grid.Unbind()
	_ = flufftest.RenderToString(grid, 20, 6)
}

func TestCheckboxAndComponentCoverage(t *testing.T) {
	cb := NewCheckbox("Accept")
	cb.SetLabel("Agree")
	cb.SetOnChange(func(value *bool) {})
	cb.SetStyle(backend.DefaultStyle())
	cb.SetFocusStyle(backend.DefaultStyle().Underline(true))
	value := true
	cb.SetChecked(&value)
	cb.Focus()
	cb.Layout(runtime.Rect{X: 0, Y: 0, Width: 10, Height: 1})
	cb.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	cb.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: ' '})
	_ = cb.StyleType()
	_ = flufftest.RenderToString(cb, 10, 1)

	comp := &Component{}
	sig := state.NewSignal(0)
	comp.Observe(sig, func() {})
}

func TestCanvasWidgetBlitterOptions(t *testing.T) {
	widget := NewCanvasWidget(func(c *graphics.Canvas) {}, WithCanvasBlitter(&graphics.SextantBlitter{}))
	widget.SetBlitter(&graphics.BrailleBlitter{})
	widget.WithBlitter(&graphics.QuadrantBlitter{})
	widget.Layout(runtime.Rect{X: 0, Y: 0, Width: 6, Height: 3})
	_ = flufftest.RenderToString(widget, 6, 3)
}
