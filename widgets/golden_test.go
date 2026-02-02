package widgets

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	testing2 "github.com/odvcencio/fluffyui/testing"
)

// goldenPath returns the path to a golden file.
func goldenPath(name string) string {
	return filepath.Join("testdata", name+".golden")
}

// assertGolden compares rendered output against a golden file.
func assertGolden(t *testing.T, widget runtime.Widget, width, height int, name string) {
	t.Helper()
	testing2.AssertWidgetGolden(t, widget, width, height, goldenPath(name))
}

// ============================================================================
// Input Widgets
// ============================================================================

func TestGolden_Button(t *testing.T) {
	tests := []struct {
		name   string
		widget *Button
	}{
		{"default", NewButton("Click Me")},
		{"primary", NewButton("Primary", WithVariant(VariantPrimary))},
		{"secondary", NewButton("Secondary", WithVariant(VariantSecondary))},
		{"danger", NewButton("Delete", WithVariant(VariantDanger))},
		{"long_label", NewButton("This is a button with a very long label")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGolden(t, tt.widget, 50, 3, "button_"+tt.name)
		})
	}
}

func TestGolden_Button_Focused(t *testing.T) {
	btn := NewButton("Focused Button")
	btn.Focus()
	assertGolden(t, btn, 20, 1, "button_focused")
}

func TestGolden_Button_Disabled(t *testing.T) {
	disabled := state.NewSignal(true)
	btn := NewButton("Disabled", WithDisabled(disabled))
	assertGolden(t, btn, 15, 1, "button_disabled")
}

func TestGolden_Button_Loading(t *testing.T) {
	loading := state.NewSignal(true)
	btn := NewButton("Loading", WithLoading(loading))
	assertGolden(t, btn, 15, 1, "button_loading")
}

func TestGolden_Input(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *Input
		width  int
		height int
	}{
		{
			"empty",
			func() *Input { return NewInput() },
			20, 1,
		},
		{
			"with_placeholder",
			func() *Input {
				i := NewInput()
				i.SetPlaceholder("Enter text...")
				return i
			},
			25, 1,
		},
		{
			"with_text",
			func() *Input {
				i := NewInput()
				i.SetText("Hello, World!")
				return i
			},
			20, 1,
		},
		{
			"focused",
			func() *Input {
				i := NewInput()
				i.SetText("Focused")
				i.Focus()
				return i
			},
			20, 1,
		},
		{
			"long_text",
			func() *Input {
				i := NewInput()
				i.SetText("This is a very long text that should be scrolled")
				i.Focus()
				return i
			},
			30, 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGolden(t, tt.setup(), tt.width, tt.height, "input_"+tt.name)
		})
	}
}

func TestGolden_TextArea(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *TextArea
		width  int
		height int
	}{
		{
			"empty",
			func() *TextArea { return NewTextArea() },
			30, 5,
		},
		{
			"with_text",
			func() *TextArea {
				ta := NewTextArea()
				ta.SetText("Line 1\nLine 2\nLine 3")
				return ta
			},
			30, 5,
		},
		{
			"focused",
			func() *TextArea {
				ta := NewTextArea()
				ta.SetText("Focused area")
				ta.Focus()
				return ta
			},
			30, 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGolden(t, tt.setup(), tt.width, tt.height, "textarea_"+tt.name)
		})
	}
}

func TestGolden_Checkbox(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *Checkbox
	}{
		{
			"unchecked",
			func() *Checkbox { return NewCheckbox("Unchecked") },
		},
		{
			"checked",
			func() *Checkbox {
				cb := NewCheckbox("Checked")
				checked := true
				cb.SetChecked(&checked)
				return cb
			},
		},
		{
			"indeterminate",
			func() *Checkbox {
				cb := NewCheckbox("Indeterminate")
				cb.SetChecked(nil)
				return cb
			},
		},
		{
			"focused",
			func() *Checkbox {
				cb := NewCheckbox("Focused")
				cb.Focus()
				return cb
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGolden(t, tt.setup(), 20, 1, "checkbox_"+tt.name)
		})
	}
}

func TestGolden_Radio(t *testing.T) {
	group := NewRadioGroup()
	tests := []struct {
		name  string
		setup func() *Radio
	}{
		{
			"unselected",
			func() *Radio {
				g := NewRadioGroup()
				return NewRadio("Option A", g)
			},
		},
		{
			"selected",
			func() *Radio {
				g := NewRadioGroup()
				r := NewRadio("Selected", g)
				g.SetSelected(0)
				return r
			},
		},
		{
			"focused",
			func() *Radio {
				r := NewRadio("Focused", group)
				r.Focus()
				return r
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGolden(t, tt.setup(), 20, 1, "radio_"+tt.name)
		})
	}
}

func TestGolden_Select(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *Select
		width  int
		height int
	}{
		{
			"default",
			func() *Select {
				return NewSelect(
					SelectOption{Label: "Option 1"},
					SelectOption{Label: "Option 2"},
					SelectOption{Label: "Option 3"},
				)
			},
			30, 3,
		},
		{
			"focused",
			func() *Select {
				s := NewSelect(
					SelectOption{Label: "Apple"},
					SelectOption{Label: "Banana"},
					SelectOption{Label: "Cherry"},
				)
				s.Focus()
				return s
			},
			30, 3,
		},
		{
			"with_disabled",
			func() *Select {
				return NewSelect(
					SelectOption{Label: "Enabled"},
					SelectOption{Label: "Disabled", Disabled: true},
					SelectOption{Label: "Also Enabled"},
				)
			},
			30, 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGolden(t, tt.setup(), tt.width, tt.height, "select_"+tt.name)
		})
	}
}

func TestGolden_Slider(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *Slider
		width  int
		height int
	}{
		{
			"default",
			func() *Slider {
				v := state.NewSignal(50.0)
				return NewSlider(v)
			},
			30, 1,
		},
		{
			"min",
			func() *Slider {
				v := state.NewSignal(0.0)
				return NewSlider(v)
			},
			30, 1,
		},
		{
			"max",
			func() *Slider {
				v := state.NewSignal(100.0)
				return NewSlider(v)
			},
			30, 1,
		},
		{
			"focused",
			func() *Slider {
				v := state.NewSignal(75.0)
				s := NewSlider(v)
				s.Focus()
				return s
			},
			30, 1,
		},
		{
			"with_value_display",
			func() *Slider {
				v := state.NewSignal(42.0)
				return NewSlider(v, WithSliderShowValue(true))
			},
			35, 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGolden(t, tt.setup(), tt.width, tt.height, "slider_"+tt.name)
		})
	}
}

// ============================================================================
// Data Widgets
// ============================================================================

func TestGolden_Table(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *Table
		width  int
		height int
	}{
		{
			"empty",
			func() *Table {
				return NewTable(
					TableColumn{Title: "Name"},
					TableColumn{Title: "Value"},
				)
			},
			30, 5,
		},
		{
			"with_data",
			func() *Table {
				tbl := NewTable(
					TableColumn{Title: "ID"},
					TableColumn{Title: "Name"},
					TableColumn{Title: "Status"},
				)
				tbl.SetRows([][]string{
					{"1", "Alice", "Active"},
					{"2", "Bob", "Inactive"},
					{"3", "Charlie", "Active"},
				})
				return tbl
			},
			40, 6,
		},
		{
			"focused",
			func() *Table {
				tbl := NewTable(
					TableColumn{Title: "Item"},
				)
				tbl.SetRows([][]string{{"First"}, {"Second"}, {"Third"}})
				tbl.Focus()
				return tbl
			},
			20, 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGolden(t, tt.setup(), tt.width, tt.height, "table_"+tt.name)
		})
	}
}

func TestGolden_DataGrid(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *DataGrid
		width  int
		height int
	}{
		{
			"default",
			func() *DataGrid {
				g := NewDataGrid(
					TableColumn{Title: "Col A"},
					TableColumn{Title: "Col B"},
				)
				g.SetRows([][]string{{"A1", "B1"}, {"A2", "B2"}})
				return g
			},
			30, 5,
		},
		{
			"selected_cell",
			func() *DataGrid {
				g := NewDataGrid(
					TableColumn{Title: "X"},
					TableColumn{Title: "Y"},
				)
				g.SetRows([][]string{{"1", "2"}, {"3", "4"}})
				g.SetSelected(1, 1)
				g.Focus()
				return g
			},
			20, 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGolden(t, tt.setup(), tt.width, tt.height, "datagrid_"+tt.name)
		})
	}
}

func TestGolden_Tree(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *Tree
		width  int
		height int
	}{
		{
			"simple",
			func() *Tree {
				root := &TreeNode{
					Label: "Root",
					Children: []*TreeNode{
						{Label: "Child 1"},
						{Label: "Child 2"},
					},
				}
				return NewTree(root)
			},
			25, 5,
		},
		{
			"expanded",
			func() *Tree {
				root := &TreeNode{
					Label:    "Root",
					Expanded: true,
					Children: []*TreeNode{
						{Label: "Child A", Children: []*TreeNode{{Label: "Grandchild"}}},
						{Label: "Child B"},
					},
				}
				return NewTree(root)
			},
			25, 6,
		},
		{
			"focused",
			func() *Tree {
				root := &TreeNode{Label: "Focused", Children: []*TreeNode{{Label: "Item"}}}
				tr := NewTree(root)
				tr.Focus()
				return tr
			},
			20, 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGolden(t, tt.setup(), tt.width, tt.height, "tree_"+tt.name)
		})
	}
}

// ============================================================================
// Picker Widgets
// ============================================================================

func TestGolden_Calendar(t *testing.T) {
	fixedNow := time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		setup  func() *Calendar
		width  int
		height int
	}{
		{
			"default",
			func() *Calendar {
				cal := NewCalendar()
				cal.now = func() time.Time { return fixedNow }
				cal.SetDisplayedMonth(fixedNow)
				return cal
			},
			28, 10,
		},
		{
			"with_selection",
			func() *Calendar {
				cal := NewCalendar()
				cal.now = func() time.Time { return fixedNow }
				cal.SetDisplayedMonth(fixedNow)
				cal.SetSelectedDate(fixedNow.AddDate(0, 0, 5))
				return cal
			},
			28, 10,
		},
		{
			"focused",
			func() *Calendar {
				cal := NewCalendar()
				cal.now = func() time.Time { return fixedNow }
				cal.SetDisplayedMonth(fixedNow)
				cal.Focus()
				return cal
			},
			28, 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGolden(t, tt.setup(), tt.width, tt.height, "calendar_"+tt.name)
		})
	}
}

func TestGolden_DatePicker(t *testing.T) {
	fixedNow := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		setup  func() *DatePicker
		width  int
		height int
	}{
		{
			"default",
			func() *DatePicker {
				dp := NewDatePicker()
				dp.calendar.now = func() time.Time { return fixedNow }
				dp.calendar.SetDisplayedMonth(fixedNow)
				return dp
			},
			35, 12,
		},
		{
			"with_date",
			func() *DatePicker {
				dp := NewDatePicker()
				dp.calendar.now = func() time.Time { return fixedNow }
				dp.calendar.SetSelectedDate(fixedNow)
				dp.calendar.SetDisplayedMonth(fixedNow)
				return dp
			},
			35, 12,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGolden(t, tt.setup(), tt.width, tt.height, "datepicker_"+tt.name)
		})
	}
}

func TestGolden_DateRangePicker(t *testing.T) {
	fixedNow := time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		setup  func() *DateRangePicker
		width  int
		height int
	}{
		{
			"default",
			func() *DateRangePicker {
				drp := NewDateRangePicker()
				drp.calendar.now = func() time.Time { return fixedNow }
				return drp
			},
			34, 10,
		},
		{
			"with_range",
			func() *DateRangePicker {
				drp := NewDateRangePicker()
				drp.calendar.now = func() time.Time { return fixedNow }
				start := time.Date(2026, time.January, 5, 0, 0, 0, 0, time.UTC)
				end := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
				drp.SetRange(&start, &end)
				return drp
			},
			34, 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGolden(t, tt.setup(), tt.width, tt.height, "daterangepicker_"+tt.name)
		})
	}
}

func TestGolden_TimePicker(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *TimePicker
		width  int
		height int
	}{
		{
			"default",
			func() *TimePicker {
				tp := NewTimePicker()
				tp.SetTime(time.Date(2026, 1, 1, 14, 30, 0, 0, time.UTC))
				return tp
			},
			10, 1,
		},
		{
			"with_seconds",
			func() *TimePicker {
				tp := NewTimePicker()
				tp.SetShowSeconds(true)
				tp.SetTime(time.Date(2026, 1, 1, 9, 15, 45, 0, time.UTC))
				return tp
			},
			12, 1,
		},
		{
			"focused",
			func() *TimePicker {
				tp := NewTimePicker()
				tp.SetTime(time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC))
				tp.Focus()
				return tp
			},
			10, 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGolden(t, tt.setup(), tt.width, tt.height, "timepicker_"+tt.name)
		})
	}
}

// ============================================================================
// Navigation Widgets
// ============================================================================

func TestGolden_Tabs(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *Tabs
		width  int
		height int
	}{
		{
			"default",
			func() *Tabs {
				return NewTabs(
					Tab{Title: "Tab 1", Content: NewLabel("Content 1")},
					Tab{Title: "Tab 2", Content: NewLabel("Content 2")},
					Tab{Title: "Tab 3", Content: NewLabel("Content 3")},
				)
			},
			40, 5,
		},
		{
			"focused",
			func() *Tabs {
				tabs := NewTabs(
					Tab{Title: "Home"},
					Tab{Title: "Settings"},
				)
				tabs.Focus()
				return tabs
			},
			30, 3,
		},
		{
			"second_selected",
			func() *Tabs {
				tabs := NewTabs(
					Tab{Title: "First"},
					Tab{Title: "Second"},
					Tab{Title: "Third"},
				)
				tabs.SetSelected(1)
				return tabs
			},
			35, 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGolden(t, tt.setup(), tt.width, tt.height, "tabs_"+tt.name)
		})
	}
}

func TestGolden_Menu(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *Menu
		width  int
		height int
	}{
		{
			"simple",
			func() *Menu {
				return NewMenu(
					&MenuItem{ID: "1", Title: "File"},
					&MenuItem{ID: "2", Title: "Edit"},
					&MenuItem{ID: "3", Title: "View"},
				)
			},
			20, 5,
		},
		{
			"with_shortcuts",
			func() *Menu {
				return NewMenu(
					&MenuItem{ID: "1", Title: "New", Shortcut: "Ctrl+N"},
					&MenuItem{ID: "2", Title: "Open", Shortcut: "Ctrl+O"},
					&MenuItem{ID: "3", Title: "Save", Shortcut: "Ctrl+S"},
				)
			},
			25, 5,
		},
		{
			"focused",
			func() *Menu {
				m := NewMenu(
					&MenuItem{ID: "1", Title: "Option A"},
					&MenuItem{ID: "2", Title: "Option B"},
				)
				m.Focus()
				return m
			},
			20, 4,
		},
		{
			"with_children",
			func() *Menu {
				return NewMenu(
					&MenuItem{
						ID:       "1",
						Title:    "Parent",
						Expanded: true,
						Children: []*MenuItem{
							{ID: "1.1", Title: "Child 1"},
							{ID: "1.2", Title: "Child 2"},
						},
					},
				)
			},
			25, 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGolden(t, tt.setup(), tt.width, tt.height, "menu_"+tt.name)
		})
	}
}

func TestGolden_Breadcrumb(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *Breadcrumb
		width  int
		height int
	}{
		{
			"simple",
			func() *Breadcrumb {
				return NewBreadcrumb(
					BreadcrumbItem{Label: "Home"},
					BreadcrumbItem{Label: "Products"},
					BreadcrumbItem{Label: "Category"},
				)
			},
			35, 1,
		},
		{
			"focused",
			func() *Breadcrumb {
				b := NewBreadcrumb(
					BreadcrumbItem{Label: "Root"},
					BreadcrumbItem{Label: "Path"},
				)
				b.Focus()
				return b
			},
			25, 1,
		},
		{
			"custom_separator",
			func() *Breadcrumb {
				b := NewBreadcrumb(
					BreadcrumbItem{Label: "A"},
					BreadcrumbItem{Label: "B"},
					BreadcrumbItem{Label: "C"},
				)
				b.SetSeparator(" / ")
				return b
			},
			20, 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGolden(t, tt.setup(), tt.width, tt.height, "breadcrumb_"+tt.name)
		})
	}
}

func TestGolden_Stepper(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *Stepper
		width  int
		height int
	}{
		{
			"all_pending",
			func() *Stepper {
				return NewStepper(
					Step{Title: "Step 1", State: StepPending},
					Step{Title: "Step 2", State: StepPending},
					Step{Title: "Step 3", State: StepPending},
				)
			},
			50, 1,
		},
		{
			"in_progress",
			func() *Stepper {
				return NewStepper(
					Step{Title: "Done", State: StepCompleted},
					Step{Title: "Current", State: StepActive},
					Step{Title: "Next", State: StepPending},
				)
			},
			50, 1,
		},
		{
			"with_error",
			func() *Stepper {
				return NewStepper(
					Step{Title: "OK", State: StepCompleted},
					Step{Title: "Failed", State: StepError},
					Step{Title: "Pending", State: StepPending},
				)
			},
			50, 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGolden(t, tt.setup(), tt.width, tt.height, "stepper_"+tt.name)
		})
	}
}

// ============================================================================
// Container Widgets
// ============================================================================

func TestGolden_Dialog(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *Dialog
		width  int
		height int
	}{
		{
			"simple",
			func() *Dialog {
				return NewDialog(
					"Confirm",
					"Are you sure?",
					DialogButton{Label: "Yes"},
					DialogButton{Label: "No"},
				)
			},
			30, 8,
		},
		{
			"with_long_body",
			func() *Dialog {
				return NewDialog(
					"Notice",
					"This is a longer message that explains what's happening in more detail.",
				)
			},
			40, 10,
		},
		{
			"focused",
			func() *Dialog {
				d := NewDialog("Alert", "Message",
					DialogButton{Label: "OK"},
				)
				d.Focus()
				return d
			},
			25, 7,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGolden(t, tt.setup(), tt.width, tt.height, "dialog_"+tt.name)
		})
	}
}

func TestGolden_Panel(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *Panel
		width  int
		height int
	}{
		{
			"simple",
			func() *Panel {
				return NewPanel(NewLabel("Panel content"))
			},
			25, 5,
		},
		{
			"with_title",
			func() *Panel {
				return NewPanel(
					NewLabel("Content here"),
					WithPanelTitle("My Panel"),
				)
			},
			25, 5,
		},
		{
			"with_border",
			func() *Panel {
				return NewPanel(
					NewLabel("Bordered"),
					WithPanelBorder(backend.DefaultStyle()),
				)
			},
			20, 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGolden(t, tt.setup(), tt.width, tt.height, "panel_"+tt.name)
		})
	}
}

func TestGolden_Accordion(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *Accordion
		width  int
		height int
	}{
		{
			"collapsed",
			func() *Accordion {
				return NewAccordion(
					NewAccordionSection("Section 1", NewLabel("Content 1")),
					NewAccordionSection("Section 2", NewLabel("Content 2")),
				)
			},
			30, 4,
		},
		{
			"one_expanded",
			func() *Accordion {
				sec1 := NewAccordionSection("Expanded", NewLabel("Visible content"))
				sec1.SetExpanded(true)
				return NewAccordion(
					sec1,
					NewAccordionSection("Collapsed", NewLabel("Hidden")),
				)
			},
			30, 6,
		},
		{
			"focused",
			func() *Accordion {
				acc := NewAccordion(
					NewAccordionSection("First", NewLabel("A")),
					NewAccordionSection("Second", NewLabel("B")),
				)
				acc.Focus()
				return acc
			},
			25, 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGolden(t, tt.setup(), tt.width, tt.height, "accordion_"+tt.name)
		})
	}
}

func TestGolden_Splitter(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *Splitter
		width  int
		height int
	}{
		{
			"horizontal",
			func() *Splitter {
				s := NewSplitter(NewLabel("Left"), NewLabel("Right"))
				s.Orientation = SplitHorizontal
				s.Ratio = 0.5
				return s
			},
			30, 5,
		},
		{
			"vertical",
			func() *Splitter {
				s := NewSplitter(NewLabel("Top"), NewLabel("Bottom"))
				s.Orientation = SplitVertical
				s.Ratio = 0.5
				return s
			},
			20, 10,
		},
		{
			"uneven_ratio",
			func() *Splitter {
				s := NewSplitter(NewLabel("Small"), NewLabel("Large"))
				s.Ratio = 0.25
				return s
			},
			40, 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGolden(t, tt.setup(), tt.width, tt.height, "splitter_"+tt.name)
		})
	}
}

func TestGolden_ScrollView(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *ScrollView
		width  int
		height int
	}{
		{
			"fits_content",
			func() *ScrollView {
				return NewScrollView(NewLabel("Small content"))
			},
			30, 5,
		},
		{
			"overflows",
			func() *ScrollView {
				content := NewStack(
					NewLabel("Line 1"),
					NewLabel("Line 2"),
					NewLabel("Line 3"),
					NewLabel("Line 4"),
					NewLabel("Line 5"),
					NewLabel("Line 6"),
				)
				return NewScrollView(content)
			},
			20, 4,
		},
		{
			"focused",
			func() *ScrollView {
				sv := NewScrollView(NewLabel("Scrollable"))
				sv.Focus()
				return sv
			},
			20, 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGolden(t, tt.setup(), tt.width, tt.height, "scrollview_"+tt.name)
		})
	}
}

// ============================================================================
// Feedback Widgets
// ============================================================================

func TestGolden_Alert(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *Alert
		width  int
		height int
	}{
		{"info", func() *Alert { return NewAlert("Information message", AlertInfo) }, 30, 1},
		{"success", func() *Alert { return NewAlert("Operation completed", AlertSuccess) }, 30, 1},
		{"warning", func() *Alert { return NewAlert("Please review", AlertWarning) }, 25, 1},
		{"error", func() *Alert { return NewAlert("Something went wrong", AlertError) }, 30, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGolden(t, tt.setup(), tt.width, tt.height, "alert_"+tt.name)
		})
	}
}

func TestGolden_Progress(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *Progress
		width  int
		height int
	}{
		{
			"empty",
			func() *Progress {
				p := NewProgress()
				p.Value = 0
				return p
			},
			30, 1,
		},
		{
			"half",
			func() *Progress {
				p := NewProgress()
				p.Value = 50
				return p
			},
			30, 1,
		},
		{
			"full",
			func() *Progress {
				p := NewProgress()
				p.Value = 100
				return p
			},
			30, 1,
		},
		{
			"with_label",
			func() *Progress {
				p := NewProgress()
				p.Value = 75
				p.Label = "Loading"
				return p
			},
			35, 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGolden(t, tt.setup(), tt.width, tt.height, "progress_"+tt.name)
		})
	}
}

func TestGolden_Spinner(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *Spinner
	}{
		{
			"frame_0",
			func() *Spinner {
				s := NewSpinner()
				return s
			},
		},
		{
			"frame_1",
			func() *Spinner {
				s := NewSpinner()
				s.Advance()
				return s
			},
		},
		{
			"frame_2",
			func() *Spinner {
				s := NewSpinner()
				s.Advance()
				s.Advance()
				return s
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGolden(t, tt.setup(), 3, 1, "spinner_"+tt.name)
		})
	}
}

// ============================================================================
// Additional Widgets
// ============================================================================

func TestGolden_Label(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *Label
		width  int
		height int
	}{
		{
			"simple",
			func() *Label { return NewLabel("Hello, World!") },
			20, 1,
		},
		{
			"long_truncated",
			func() *Label { return NewLabel("This is a very long label that should be truncated") },
			30, 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGolden(t, tt.setup(), tt.width, tt.height, "label_"+tt.name)
		})
	}
}

func TestGolden_Text(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *Text
		width  int
		height int
	}{
		{
			"single_line",
			func() *Text { return NewText("Simple text") },
			20, 1,
		},
		{
			"multiline",
			func() *Text { return NewText("Line one\nLine two\nLine three") },
			20, 5,
		},
		{
			"wrapped",
			func() *Text { return NewText("This is a long piece of text that should wrap to multiple lines when rendered") },
			25, 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGolden(t, tt.setup(), tt.width, tt.height, "text_"+tt.name)
		})
	}
}
