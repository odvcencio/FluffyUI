package widgets

import (
	"strings"
	"time"

	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/state"
)

const datePickerGap = 1

// DatePicker combines a text input and calendar.
type DatePicker struct {
	Base
	calendar *Calendar
	input    *Input
	format   string
	label    string
	updating bool
	services runtime.Services
	subs     state.Subscriptions
}

// NewDatePicker creates a date picker.
func NewDatePicker() *DatePicker {
	picker := &DatePicker{
		format: "2006-01-02",
		label:  "Date Picker",
	}
	picker.calendar = NewCalendar()
	picker.input = NewInput()
	picker.input.SetPlaceholder(picker.format)
	picker.input.OnChange(picker.handleInputChange)
	picker.calendar.OnSelect(func(date time.Time) {
		picker.syncInput()
	})
	picker.Base.Role = accessibility.RoleGroup
	picker.syncA11y()
	picker.syncInput()
	return picker
}

// Bind attaches app services.
func (d *DatePicker) Bind(services runtime.Services) {
	if d == nil {
		return
	}
	d.services = services
	d.subs.Clear()
	d.subs.SetScheduler(services.Scheduler())
	if sig := d.calendar.SelectedDateSignal(); sig != nil {
		d.subs.Observe(sig, func() {
			d.syncInput()
			d.services.Invalidate()
		})
	}
}

// Unbind releases app services.
func (d *DatePicker) Unbind() {
	if d == nil {
		return
	}
	d.subs.Clear()
	d.services = runtime.Services{}
}

// Calendar returns the underlying calendar.
func (d *DatePicker) Calendar() *Calendar {
	if d == nil {
		return nil
	}
	return d.calendar
}

// Input returns the underlying input widget.
func (d *DatePicker) Input() *Input {
	if d == nil {
		return nil
	}
	return d.input
}

// SetFormat updates the date format.
func (d *DatePicker) SetFormat(format string) {
	if d == nil {
		return
	}
	format = strings.TrimSpace(format)
	if format == "" {
		format = "2006-01-02"
	}
	d.format = format
	d.input.SetPlaceholder(format)
	d.syncInput()
}

// SetLabel updates the accessibility label.
func (d *DatePicker) SetLabel(label string) {
	if d == nil {
		return
	}
	d.label = label
	d.syncA11y()
	if d.input != nil {
		d.input.SetLabel(label)
	}
}

// SetSelectedDate updates the selected date.
func (d *DatePicker) SetSelectedDate(date time.Time) {
	if d == nil || d.calendar == nil {
		return
	}
	d.calendar.SetSelectedDate(date)
	d.syncInput()
}

// SelectedDate returns the selected date.
func (d *DatePicker) SelectedDate() time.Time {
	if d == nil || d.calendar == nil {
		return time.Time{}
	}
	return d.calendar.SelectedDate()
}

// SetMinDate sets the minimum selectable date.
func (d *DatePicker) SetMinDate(date *time.Time) {
	if d == nil || d.calendar == nil {
		return
	}
	d.calendar.SetMinDate(date)
}

// SetMaxDate sets the maximum selectable date.
func (d *DatePicker) SetMaxDate(date *time.Time) {
	if d == nil || d.calendar == nil {
		return
	}
	d.calendar.SetMaxDate(date)
}

// SetHighlightDates updates highlighted dates.
func (d *DatePicker) SetHighlightDates(dates []time.Time) {
	if d == nil || d.calendar == nil {
		return
	}
	d.calendar.SetHighlightDates(dates)
}

// SetWeekStart updates the calendar week start.
func (d *DatePicker) SetWeekStart(start time.Weekday) {
	if d == nil || d.calendar == nil {
		return
	}
	d.calendar.SetWeekStart(start)
}

// SetShowWeekNumbers toggles week numbers.
func (d *DatePicker) SetShowWeekNumbers(show bool) {
	if d == nil || d.calendar == nil {
		return
	}
	d.calendar.SetShowWeekNumbers(show)
}

// SetSelectionMode updates selection mode.
func (d *DatePicker) SetSelectionMode(mode CalendarSelectionMode) {
	if d == nil || d.calendar == nil {
		return
	}
	d.calendar.SetSelectionMode(mode)
}

// StyleType returns the selector type name.
func (d *DatePicker) StyleType() string {
	return "DatePicker"
}

// Measure returns the desired size.
func (d *DatePicker) Measure(constraints runtime.Constraints) runtime.Size {
	return d.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		inputSize := runtime.Size{}
		calSize := runtime.Size{}
		if d.input != nil {
			inputSize = d.input.Measure(contentConstraints)
		}
		if d.calendar != nil {
			calSize = d.calendar.Measure(contentConstraints)
		}
		width := max(inputSize.Width, calSize.Width)
		height := inputSize.Height + datePickerGap + calSize.Height
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: height})
	})
}

// Layout positions the input and calendar.
func (d *DatePicker) Layout(bounds runtime.Rect) {
	d.Base.Layout(bounds)
	content := d.ContentBounds()
	if d.input == nil || d.calendar == nil {
		return
	}
	inputHeight := 1
	if inputSize := d.input.Measure(runtime.Constraints{MaxWidth: content.Width, MaxHeight: content.Height}); inputSize.Height > 0 {
		inputHeight = inputSize.Height
	}
	inputBounds := runtime.Rect{X: content.X, Y: content.Y, Width: content.Width, Height: inputHeight}
	d.input.Layout(inputBounds)
	calY := content.Y + inputHeight + datePickerGap
	if calY > content.Y+content.Height {
		calY = content.Y + content.Height
	}
	calBounds := runtime.Rect{X: content.X, Y: calY, Width: content.Width, Height: max(0, content.Height-(calY-content.Y))}
	d.calendar.Layout(calBounds)
}

// Render draws the date picker.
func (d *DatePicker) Render(ctx runtime.RenderContext) {
	if d == nil {
		return
	}
	d.syncA11y()
	if d.input != nil {
		d.input.Render(ctx)
	}
	if d.calendar != nil {
		d.calendar.Render(ctx)
	}
}

// HandleMessage forwards messages to children.
func (d *DatePicker) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if d == nil {
		return runtime.Unhandled()
	}
	if d.input != nil {
		if result := d.input.HandleMessage(msg); result.Handled {
			return result
		}
	}
	if d.calendar != nil {
		if result := d.calendar.HandleMessage(msg); result.Handled {
			return result
		}
	}
	return runtime.Unhandled()
}

// ChildWidgets returns child widgets.
func (d *DatePicker) ChildWidgets() []runtime.Widget {
	if d == nil {
		return nil
	}
	children := make([]runtime.Widget, 0, 2)
	if d.input != nil {
		children = append(children, d.input)
	}
	if d.calendar != nil {
		children = append(children, d.calendar)
	}
	return children
}

func (d *DatePicker) handleInputChange(text string) {
	if d == nil || d.calendar == nil {
		return
	}
	if d.updating {
		return
	}
	date, ok := d.parseDate(text)
	if !ok {
		return
	}
	d.calendar.SetSelectedDate(date)
}

func (d *DatePicker) parseDate(text string) (time.Time, bool) {
	text = strings.TrimSpace(text)
	if text == "" {
		return time.Time{}, false
	}
	loc := time.Local
	if d.calendar != nil {
		if selected := d.calendar.SelectedDate(); !selected.IsZero() {
			loc = selected.Location()
		}
	}
	date, err := time.ParseInLocation(d.format, text, loc)
	if err != nil {
		return time.Time{}, false
	}
	return normalizeDate(date), true
}

func (d *DatePicker) syncInput() {
	if d == nil || d.calendar == nil || d.input == nil {
		return
	}
	date := d.calendar.SelectedDate()
	if date.IsZero() {
		return
	}
	d.updating = true
	d.input.SetText(date.Format(d.format))
	d.updating = false
}

func (d *DatePicker) syncA11y() {
	if d == nil {
		return
	}
	label := strings.TrimSpace(d.label)
	if label == "" {
		label = "Date Picker"
	}
	d.Base.Label = label
}

var _ runtime.Widget = (*DatePicker)(nil)
var _ runtime.ChildProvider = (*DatePicker)(nil)
var _ runtime.Bindable = (*DatePicker)(nil)
var _ runtime.Unbindable = (*DatePicker)(nil)
