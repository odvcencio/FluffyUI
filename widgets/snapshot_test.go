package widgets

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/state"
	"m31labs.dev/fluffyui/terminal"
)

var updateSnapshots = flag.Bool("update-snapshots", false, "Update golden snapshot files")

// renderToString renders a widget to a string for snapshot comparison.
func renderToString(w runtime.Widget, width, height int) string {
	// Create buffer
	buf := runtime.NewBuffer(width, height)

	// Measure and layout
	constraints := runtime.Constraints{MaxWidth: width, MaxHeight: height}
	w.Measure(constraints)
	w.Layout(runtime.Rect{X: 0, Y: 0, Width: width, Height: height})

	// Render
	ctx := runtime.RenderContext{Buffer: buf}
	w.Render(ctx)

	// Convert to string
	var sb strings.Builder
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			cell := buf.Get(x, y)
			if cell.Rune == 0 {
				sb.WriteRune(' ')
			} else {
				sb.WriteRune(cell.Rune)
			}
		}
		sb.WriteRune('\n')
	}
	return sb.String()
}

// assertSnapshot compares rendered output against a golden file.
func assertSnapshot(t *testing.T, name string, actual string) {
	t.Helper()

	goldenPath := filepath.Join("testdata", name+".golden")

	if *updateSnapshots {
		// Update mode: write the actual output
		if err := os.MkdirAll("testdata", 0755); err != nil {
			t.Fatalf("failed to create testdata dir: %v", err)
		}
		if err := os.WriteFile(goldenPath, []byte(actual), 0644); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
		t.Logf("Updated snapshot: %s", goldenPath)
		return
	}

	// Compare mode: read expected and compare
	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("snapshot file not found: %s\nRun with -update-snapshots to create it.\nActual output:\n%s", goldenPath, actual)
		}
		t.Fatalf("failed to read golden file: %v", err)
	}

	if actual != string(expected) {
		t.Errorf("snapshot mismatch for %s\n\nExpected:\n%s\n\nActual:\n%s\n\nRun with -update-snapshots to update.", name, string(expected), actual)
	}
}

// TestSnapshot_Palette tests the palette widget rendering.
func TestSnapshot_Palette(t *testing.T) {
	p := NewPaletteWidget("Commands")
	p.SetItems([]PaletteItem{
		{ID: "1", Category: "Session", Label: "New Conversation", Shortcut: "/new"},
		{ID: "2", Category: "Session", Label: "Clear Messages", Shortcut: "/clear"},
		{ID: "3", Category: "View", Label: "Toggle Sidebar", Shortcut: "Ctrl+B"},
	})
	p.SetStyles(
		backend.DefaultStyle(),
		backend.DefaultStyle(),
		backend.DefaultStyle().Bold(true),
		backend.DefaultStyle(),
		backend.DefaultStyle(),
		backend.DefaultStyle().Reverse(true),
		backend.DefaultStyle(),
	)
	p.Focus()

	output := renderToString(p, 60, 12)
	assertSnapshot(t, "palette", output)
}

func TestSnapshot_RichText(t *testing.T) {
	view := NewRichText("# Title\n\n- One\n- Two\n\n**Bold** text.")
	output := renderToString(view, 30, 10)
	assertSnapshot(t, "rich_text", output)
}

func TestSnapshot_DataGrid(t *testing.T) {
	grid := NewDataGrid([]DataGridColumn{
		{Header: "Name", Width: 10},
		{Header: "Value", Width: 10},
	})
	grid.SetRows([][]string{{"Alpha", "1"}, {"Beta", "2"}, {"Gamma", "3"}})
	grid.SetSelectedCell(1, 1)
	output := renderToString(grid, 30, 6)
	assertSnapshot(t, "data_grid", output)
}

func TestSnapshot_DateRangePicker(t *testing.T) {
	picker := NewDateRangePicker()
	now := time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC)
	picker.calendar.now = func() time.Time { return now }
	start := time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.January, 5, 0, 0, 0, 0, time.UTC)
	picker.SetRange(&start, &end)
	output := renderToString(picker, 34, 10)
	assertSnapshot(t, "date_range_picker", output)
}

func TestSnapshot_TimePicker(t *testing.T) {
	picker := NewTimePicker()
	picker.SetShowSeconds(true)
	picker.SetTime(time.Date(2026, time.January, 2, 9, 30, 15, 0, time.UTC))
	output := renderToString(picker, 10, 1)
	assertSnapshot(t, "time_picker", output)
}

func TestSnapshot_AutoComplete(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetOptions([]string{"Alpha", "Beta", "Gamma", "Delta"})
	ac.SetQuery("a")
	ac.Focus()
	ac.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	output := renderToString(ac, 20, 6)
	assertSnapshot(t, "autocomplete", output)
}

func TestSnapshot_MultiSelect(t *testing.T) {
	ms := NewMultiSelect(
		MultiSelectOption{Label: "Red"},
		MultiSelectOption{Label: "Blue"},
		MultiSelectOption{Label: "Green"},
	)
	ms.Focus()
	ms.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: ' '})
	ms.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	ms.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: ' '})
	output := renderToString(ms, 20, 4)
	assertSnapshot(t, "multiselect", output)
}

func TestSnapshot_MaskedInput(t *testing.T) {
	mi := NewMaskedInput("(###) ###-####", WithPlaceholder('_'))
	mi.Focus()
	// Type a partial phone number
	for _, d := range "5551234" {
		mi.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: d})
	}
	output := renderToString(mi, 20, 1)
	assertSnapshot(t, "masked_input", output)
}

func TestSnapshot_Log(t *testing.T) {
	log := NewLog(WithShowTime(false), WithAutoScroll(true))
	log.Debug("Debug message")
	log.Info("Info message")
	log.Warn("Warning message")
	log.Error("Error message")
	log.Focus()
	output := renderToString(log, 40, 5)
	assertSnapshot(t, "log", output)
}

func TestSnapshot_HeatMap(t *testing.T) {
	data := state.NewSignal([][]float64{
		{0, 5, 10},
		{15, 20, 25},
		{30, 35, 40},
	})
	hm := NewHeatMap(data)
	hm.RowLabels = []string{"R1", "R2", "R3"}
	hm.ColLabels = []string{"C1", "C2", "C3"}
	output := renderToString(hm, 15, 5)
	assertSnapshot(t, "heatmap", output)
}
