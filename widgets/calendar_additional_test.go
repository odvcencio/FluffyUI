package widgets

import (
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

func TestCalendarRangeSelection(t *testing.T) {
	cal := NewCalendar(WithSelectionMode(CalendarSelectionRange))
	start := time.Date(2026, time.March, 3, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.March, 8, 0, 0, 0, 0, time.UTC)
	cal.SetRange(&start, &end)

	gotStart := cal.RangeStart()
	gotEnd := cal.RangeEnd()
	if gotStart == nil || gotEnd == nil {
		t.Fatalf("expected range start/end to be set")
	}
	if !sameDay(*gotStart, start) || !sameDay(*gotEnd, end) {
		t.Fatalf("unexpected range: %v - %v", gotStart, gotEnd)
	}
}

func TestCalendarKeyNavigation(t *testing.T) {
	cal := NewCalendar()
	cal.SetSelectedDate(time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC))
	cal.Focus()

	cal.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if got := cal.SelectedDate(); !sameDay(got, time.Date(2026, time.January, 17, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected +7 days, got %v", got)
	}

	cal.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if got := cal.SelectedDate(); !sameDay(got, time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected -7 days, got %v", got)
	}
}

func TestCalendarMouseHeaderNavigation(t *testing.T) {
	cal := NewCalendar()
	month := time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC)
	cal.SetDisplayedMonth(month)
	cal.SetSelectedDate(month)
	cal.Focus()
	cal.Layout(runtime.Rect{X: 0, Y: 0, Width: 28, Height: 8})

	cal.HandleMessage(runtime.MouseMsg{X: 0, Y: 0, Button: runtime.MouseLeft, Action: runtime.MousePress})
	prev := cal.DisplayedMonth()
	if prev.Month() != time.April {
		t.Fatalf("expected previous month, got %v", prev)
	}
}
