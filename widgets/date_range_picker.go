package widgets

import (
	"strings"
	"time"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
)

const dateRangePickerGap = 1

// DateRangePicker combines two inputs with a range-select calendar.
type DateRangePicker struct {
	Base
	calendar   *Calendar
	startInput *Input
	endInput   *Input
	format     string
	label      string
	updating   bool
	services   runtime.Services
	subs       state.Subscriptions

	onRangeSelect func(start, end time.Time)
}

// NewDateRangePicker creates a date range picker.
func NewDateRangePicker() *DateRangePicker {
	picker := &DateRangePicker{
		format: "2006-01-02",
		label:  "Date Range",
	}
	picker.calendar = NewCalendar(WithSelectionMode(CalendarSelectionRange))
	picker.startInput = NewInput()
	picker.endInput = NewInput()
	picker.startInput.SetPlaceholder(picker.format)
	picker.endInput.SetPlaceholder(picker.format)
	picker.startInput.SetOnChange(func(text string) { picker.handleInputChange(true, text) })
	picker.endInput.SetOnChange(func(text string) { picker.handleInputChange(false, text) })
	picker.calendar.OnRangeSelect(func(start, end time.Time) {
		picker.syncInputs()
		if picker.onRangeSelect != nil {
			picker.onRangeSelect(start, end)
		}
	})
	picker.Base.Role = accessibility.RoleGroup
	picker.syncA11y()
	picker.syncInputs()
	return picker
}

// Bind attaches app services.
func (d *DateRangePicker) Bind(services runtime.Services) {
	if d == nil {
		return
	}
	d.services = services
	d.subs.Clear()
	d.subs.SetScheduler(services.Scheduler())
	runtime.BindTree(d.startInput, services)
	runtime.BindTree(d.endInput, services)
	runtime.BindTree(d.calendar, services)

	if sig := d.calendar.RangeStartSignal(); sig != nil {
		d.subs.Observe(sig, func() {
			d.syncInputs()
			d.services.Invalidate()
		})
	}
	if sig := d.calendar.RangeEndSignal(); sig != nil {
		d.subs.Observe(sig, func() {
			d.syncInputs()
			d.services.Invalidate()
		})
	}
}

// Unbind releases app services.
func (d *DateRangePicker) Unbind() {
	if d == nil {
		return
	}
	d.subs.Clear()
	runtime.UnbindTree(d.startInput)
	runtime.UnbindTree(d.endInput)
	runtime.UnbindTree(d.calendar)
	d.services = runtime.Services{}
}

// Calendar returns the underlying calendar.
func (d *DateRangePicker) Calendar() *Calendar {
	if d == nil {
		return nil
	}
	return d.calendar
}

// StartInput returns the start input widget.
func (d *DateRangePicker) StartInput() *Input {
	if d == nil {
		return nil
	}
	return d.startInput
}

// EndInput returns the end input widget.
func (d *DateRangePicker) EndInput() *Input {
	if d == nil {
		return nil
	}
	return d.endInput
}

// SetFormat updates the date format.
func (d *DateRangePicker) SetFormat(format string) {
	if d == nil {
		return
	}
	format = strings.TrimSpace(format)
	if format == "" {
		format = "2006-01-02"
	}
	d.format = format
	d.startInput.SetPlaceholder(format)
	d.endInput.SetPlaceholder(format)
	d.syncInputs()
}

// SetLabel updates the accessibility label.
func (d *DateRangePicker) SetLabel(label string) {
	if d == nil {
		return
	}
	d.label = label
	d.syncA11y()
	if d.startInput != nil {
		d.startInput.SetLabel(label + " start")
	}
	if d.endInput != nil {
		d.endInput.SetLabel(label + " end")
	}
}

// SetRange updates the selected date range.
func (d *DateRangePicker) SetRange(start, end *time.Time) {
	if d == nil || d.calendar == nil {
		return
	}
	d.calendar.SetRange(start, end)
	d.syncInputs()
}

// SelectedRange returns the selected range and a flag if both ends are set.
func (d *DateRangePicker) SelectedRange() (time.Time, time.Time, bool) {
	if d == nil || d.calendar == nil {
		return time.Time{}, time.Time{}, false
	}
	start := d.calendar.RangeStart()
	end := d.calendar.RangeEnd()
	if start == nil || end == nil {
		return time.Time{}, time.Time{}, false
	}
	return *start, *end, true
}

// OnRangeSelect registers a selection callback.
func (d *DateRangePicker) OnRangeSelect(fn func(start, end time.Time)) {
	if d == nil {
		return
	}
	d.onRangeSelect = fn
}

// StyleType returns the selector type name.
func (d *DateRangePicker) StyleType() string {
	return "DateRangePicker"
}

// Measure returns the desired size.
func (d *DateRangePicker) Measure(constraints runtime.Constraints) runtime.Size {
	return d.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		startSize := runtime.Size{}
		endSize := runtime.Size{}
		calSize := runtime.Size{}
		if d.startInput != nil {
			startSize = d.startInput.Measure(contentConstraints)
		}
		if d.endInput != nil {
			endSize = d.endInput.Measure(contentConstraints)
		}
		if d.calendar != nil {
			calSize = d.calendar.Measure(contentConstraints)
		}
		rowWidth := startSize.Width + endSize.Width + 3
		width := max(rowWidth, calSize.Width)
		height := startSize.Height + dateRangePickerGap + calSize.Height
		if startSize.Height < 1 {
			height = 1 + dateRangePickerGap + calSize.Height
		}
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: height})
	})
}

// Layout positions the inputs and calendar.
func (d *DateRangePicker) Layout(bounds runtime.Rect) {
	d.Base.Layout(bounds)
	content := d.ContentBounds()
	if content.Width <= 0 || content.Height <= 0 {
		return
	}
	rowHeight := 1
	inputY := content.Y
	sep := " → "
	sepWidth := textWidth(sep)
	leftWidth := (content.Width - sepWidth) / 2
	if leftWidth < 6 {
		leftWidth = max(1, content.Width)
		if d.startInput != nil {
			d.startInput.Layout(runtime.Rect{X: content.X, Y: inputY, Width: content.Width, Height: rowHeight})
		}
		if d.endInput != nil {
			d.endInput.Layout(runtime.Rect{X: content.X, Y: inputY + rowHeight, Width: content.Width, Height: rowHeight})
		}
		inputY += rowHeight * 2
	} else {
		rightWidth := content.Width - leftWidth - sepWidth
		if d.startInput != nil {
			d.startInput.Layout(runtime.Rect{X: content.X, Y: inputY, Width: leftWidth, Height: rowHeight})
		}
		if d.endInput != nil {
			d.endInput.Layout(runtime.Rect{X: content.X + leftWidth + sepWidth, Y: inputY, Width: rightWidth, Height: rowHeight})
		}
		inputY += rowHeight
	}
	calY := inputY + dateRangePickerGap
	if calY < content.Y+content.Height {
		if d.calendar != nil {
			d.calendar.Layout(runtime.Rect{X: content.X, Y: calY, Width: content.Width, Height: content.Height - (calY - content.Y)})
		}
	}
}

// Render draws inputs, separator, and calendar.
func (d *DateRangePicker) Render(ctx runtime.RenderContext) {
	if d == nil {
		return
	}
	d.syncA11y()
	baseStyle := resolveBaseStyle(ctx, d, backend.DefaultStyle(), false)
	if d.startInput != nil {
		d.startInput.Render(ctx)
	}
	if d.endInput != nil {
		d.endInput.Render(ctx)
	}
	content := d.ContentBounds()
	sep := " → "
	sepWidth := textWidth(sep)
	leftWidth := (content.Width - sepWidth) / 2
	if leftWidth >= 6 {
		ctx.Buffer.SetString(content.X+leftWidth, content.Y, sep, baseStyle)
	}
	if d.calendar != nil {
		d.calendar.Render(ctx)
	}
}

// HandleMessage forwards messages to child widgets.
func (d *DateRangePicker) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if d == nil {
		return runtime.Unhandled()
	}
	if d.startInput != nil {
		if result := d.startInput.HandleMessage(msg); result.Handled {
			return result
		}
	}
	if d.endInput != nil {
		if result := d.endInput.HandleMessage(msg); result.Handled {
			return result
		}
	}
	if d.calendar != nil {
		return d.calendar.HandleMessage(msg)
	}
	return runtime.Unhandled()
}

// ChildWidgets returns child widgets for traversal.
func (d *DateRangePicker) ChildWidgets() []runtime.Widget {
	if d == nil {
		return nil
	}
	children := []runtime.Widget{}
	if d.startInput != nil {
		children = append(children, d.startInput)
	}
	if d.endInput != nil {
		children = append(children, d.endInput)
	}
	if d.calendar != nil {
		children = append(children, d.calendar)
	}
	return children
}

func (d *DateRangePicker) handleInputChange(isStart bool, text string) {
	if d == nil || d.updating {
		return
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	parsed, err := time.Parse(d.format, text)
	if err != nil {
		return
	}
	parsed = normalizeDate(parsed)
	start := d.calendar.RangeStart()
	end := d.calendar.RangeEnd()
	if isStart {
		start = &parsed
	} else {
		end = &parsed
	}
	d.calendar.SetRange(start, end)
}

func (d *DateRangePicker) syncInputs() {
	if d == nil {
		return
	}
	start := d.calendar.RangeStart()
	end := d.calendar.RangeEnd()
	d.updating = true
	defer func() { d.updating = false }()
	if d.startInput != nil {
		if start != nil {
			d.startInput.SetText(start.Format(d.format))
		} else {
			d.startInput.SetText("")
		}
	}
	if d.endInput != nil {
		if end != nil {
			d.endInput.SetText(end.Format(d.format))
		} else {
			d.endInput.SetText("")
		}
	}
}

func (d *DateRangePicker) syncA11y() {
	if d == nil {
		return
	}
	if d.Base.Role == "" {
		d.Base.Role = accessibility.RoleGroup
	}
	label := strings.TrimSpace(d.label)
	if label == "" {
		label = "Date Range"
	}
	d.Base.Label = label
}

var _ runtime.Widget = (*DateRangePicker)(nil)
var _ runtime.ChildProvider = (*DateRangePicker)(nil)
var _ runtime.Bindable = (*DateRangePicker)(nil)
var _ runtime.Unbindable = (*DateRangePicker)(nil)
