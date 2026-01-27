package widgets

import (
	"strings"
	"testing"
	"time"

	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/terminal"
	fluffytest "github.com/odvcencio/fluffy-ui/testing"
)

func TestCalendarSelectClampsToMinMax(t *testing.T) {
	cal := NewCalendar()
	min := time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC)
	max := time.Date(2026, time.January, 20, 0, 0, 0, 0, time.UTC)
	cal.SetMinDate(&min)
	cal.SetMaxDate(&max)

	cal.SetSelectedDate(time.Date(2026, time.January, 5, 0, 0, 0, 0, time.UTC))
	if got := cal.SelectedDate(); !sameDay(got, min) {
		t.Fatalf("selected date = %v, want %v", got, min)
	}

	cal.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})
	if got := cal.SelectedDate(); !sameDay(got, min) {
		t.Fatalf("selected date after left = %v, want %v", got, min)
	}
}

func TestCalendarRenderHeader(t *testing.T) {
	cal := NewCalendar()
	cal.SetDisplayedMonth(time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC))
	cal.SetSelectedDate(time.Date(2026, time.March, 14, 0, 0, 0, 0, time.UTC))

	out := fluffytest.RenderToString(cal, 32, 8)
	if !strings.Contains(out, "March 2026") {
		t.Fatalf("expected header to contain month, got:\n%s", out)
	}
}
