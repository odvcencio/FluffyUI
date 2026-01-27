package widgets

import (
	"testing"
	"time"
)

func TestDatePickerSyncFromCalendar(t *testing.T) {
	picker := NewDatePicker()
	date := time.Date(2026, time.February, 14, 0, 0, 0, 0, time.UTC)
	picker.SetSelectedDate(date)

	if got := picker.Input().Text(); got != "2026-02-14" {
		t.Fatalf("input text = %q, want %q", got, "2026-02-14")
	}
}

func TestDatePickerParseInput(t *testing.T) {
	picker := NewDatePicker()
	picker.handleInputChange("2026-03-01")
	selected := picker.SelectedDate()
	if selected.Year() != 2026 || selected.Month() != time.March || selected.Day() != 1 {
		t.Fatalf("selected date = %v, want 2026-03-01", selected)
	}
}
