package widgets

import (
	"strings"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/forms"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
	flufftest "github.com/odvcencio/fluffyui/testing"
)

type testVirtualListAdapter struct {
	items []string
}

func (a testVirtualListAdapter) Count() int { return len(a.items) }

func (a testVirtualListAdapter) Item(index int) string { return a.items[index] }

func (a testVirtualListAdapter) Render(item string, index int, selected bool, ctx runtime.RenderContext) {
	ctx.Buffer.SetString(ctx.Bounds.X, ctx.Bounds.Y, item, backend.DefaultStyle())
}

func TestAutoCompleteSuggestions(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetOptions([]string{"Apple", "Apricot", "Banana"})
	ac.updateSuggestions("ap")
	if len(ac.suggestions) != 2 {
		t.Fatalf("expected 2 suggestions, got %d", len(ac.suggestions))
	}
	if ac.suggestions[0] != "Apple" || ac.suggestions[1] != "Apricot" {
		t.Errorf("unexpected suggestions: %v", ac.suggestions)
	}
}

func TestMultiSelectToggle(t *testing.T) {
	ms := NewMultiSelect(
		MultiSelectOption{Label: "One"},
		MultiSelectOption{Label: "Two"},
	)
	ms.Focus()
	ms.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	selected := ms.SelectedOptions()
	if len(selected) != 1 || selected[0].Label != "One" {
		t.Fatalf("expected first option selected, got %v", selected)
	}
}

func TestDateRangePickerSetRange(t *testing.T) {
	picker := NewDateRangePicker()
	start := time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
	picker.SetRange(&start, &end)
	if picker.StartInput().Text() != start.Format("2006-01-02") {
		t.Errorf("expected start input %s, got %s", start.Format("2006-01-02"), picker.StartInput().Text())
	}
	if picker.EndInput().Text() != end.Format("2006-01-02") {
		t.Errorf("expected end input %s, got %s", end.Format("2006-01-02"), picker.EndInput().Text())
	}
}

func TestTimePickerFormat(t *testing.T) {
	picker := NewTimePicker()
	picker.SetShowSeconds(true)
	when := time.Date(2026, time.January, 1, 13, 5, 9, 0, time.UTC)
	picker.SetTime(when)
	if picker.Time().Format(picker.format()) != "13:05:09" {
		t.Errorf("expected formatted time 13:05:09, got %s", picker.Time().Format(picker.format()))
	}
	if picker.Time().Hour() != 13 || picker.Time().Minute() != 5 || picker.Time().Second() != 9 {
		t.Errorf("unexpected time from picker: %v", picker.Time())
	}
}

func TestRichTextRender(t *testing.T) {
	view := NewRichText("Hello **World**")
	output := flufftest.RenderToString(view, 20, 3)
	if !strings.Contains(output, "Hello") || !strings.Contains(output, "World") {
		t.Errorf("unexpected rich text output:\n%s", output)
	}
}

func TestDataGridEditingCommit(t *testing.T) {
	grid := NewDataGrid(
		TableColumn{Title: "Name"},
		TableColumn{Title: "Value"},
	)
	grid.SetRows([][]string{{"Alpha", "1"}})
	grid.SetSelected(0, 1)
	grid.StartEditing()
	if !grid.editing {
		t.Fatalf("expected editing to be true")
	}
	grid.editor.SetText("42")
	grid.CommitEditing()
	if grid.GetCell(0, 1) != "42" {
		t.Errorf("expected cell updated to 42, got %s", grid.GetCell(0, 1))
	}
}

func TestInputValidation(t *testing.T) {
	input := NewInput()
	input.SetValidators(forms.Required(""))
	input.SetText("")
	if input.Valid() {
		t.Errorf("expected input to be invalid when empty")
	}
	input.SetText("ok")
	if !input.Valid() {
		t.Errorf("expected input to be valid when filled")
	}
}

func TestVirtualListLazyLoad(t *testing.T) {
	list := NewVirtualList(testVirtualListAdapter{items: []string{"A", "B", "C", "D", "E", "F"}})
	calls := 0
	list.SetLazyLoad(func(start, end, total int) { calls++ })
	_ = flufftest.RenderToString(list, 10, 3)
	if calls == 0 {
		t.Errorf("expected lazy load callback to be called")
	}
}
