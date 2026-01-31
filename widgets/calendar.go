package widgets

import (
	"fmt"
	"strings"
	"time"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	"github.com/odvcencio/fluffyui/terminal"
)

// CalendarSelectionMode controls selection behavior.
type CalendarSelectionMode int

const (
	CalendarSelectionSingle CalendarSelectionMode = iota
	CalendarSelectionRange
)

// CalendarDayState describes rendering flags for a day.
type CalendarDayState struct {
	InMonth     bool
	Selected    bool
	Disabled    bool
	Today       bool
	Highlighted bool
	InRange     bool
	RangeStart  bool
	RangeEnd    bool
}

// DayRenderFunc customizes day rendering.
type DayRenderFunc func(ctx runtime.RenderContext, date time.Time, state CalendarDayState)

// CalendarOption configures a calendar.
type CalendarOption = Option[Calendar]

// Calendar displays a month grid with selectable days.
type Calendar struct {
	FocusableBase

	selectedDate   *state.Signal[time.Time]
	displayedMonth *state.Signal[time.Time]
	minDate        *state.Signal[*time.Time]
	maxDate        *state.Signal[*time.Time]
	highlightDates *state.Signal[[]time.Time]
	rangeStart     *state.Signal[*time.Time]
	rangeEnd       *state.Signal[*time.Time]

	selectionMode CalendarSelectionMode
	weekStart     time.Weekday
	showWeekNums  bool

	label         string
	headerFormat  string
	dayRenderer   DayRenderFunc
	onSelect      func(time.Time)
	onRangeSelect func(start, end time.Time)

	style          backend.Style
	headerStyle    backend.Style
	weekdayStyle   backend.Style
	selectedStyle  backend.Style
	todayStyle     backend.Style
	disabledStyle  backend.Style
	highlightStyle backend.Style
	outsideStyle   backend.Style
	rangeStyle     backend.Style

	now      func() time.Time
	services runtime.Services
	subs     state.Subscriptions
}

// NewCalendar creates a new calendar widget.
func NewCalendar(opts ...CalendarOption) *Calendar {
	now := time.Now()
	today := normalizeDate(now)
	month := monthStart(today)

	selected := state.NewSignal(today)
	selected.SetEqualFunc(equalDay)
	displayed := state.NewSignal(month)
	displayed.SetEqualFunc(equalMonth)
	minDate := state.NewSignal[*time.Time](nil)
	maxDate := state.NewSignal[*time.Time](nil)
	highlights := state.NewSignal([]time.Time(nil))
	rangeStart := state.NewSignal[*time.Time](nil)
	rangeEnd := state.NewSignal[*time.Time](nil)

	cal := &Calendar{
		selectedDate:   selected,
		displayedMonth: displayed,
		minDate:        minDate,
		maxDate:        maxDate,
		highlightDates: highlights,
		rangeStart:     rangeStart,
		rangeEnd:       rangeEnd,
		selectionMode:  CalendarSelectionSingle,
		weekStart:      time.Sunday,
		showWeekNums:   false,
		label:          "Calendar",
		headerFormat:   "January 2006",
		style:          backend.DefaultStyle(),
		headerStyle:    backend.DefaultStyle().Bold(true),
		weekdayStyle:   backend.DefaultStyle().Dim(true),
		selectedStyle:  backend.DefaultStyle().Reverse(true),
		todayStyle:     backend.DefaultStyle().Underline(true),
		disabledStyle:  backend.DefaultStyle().Dim(true),
		highlightStyle: backend.DefaultStyle().Foreground(backend.ColorYellow),
		outsideStyle:   backend.DefaultStyle().Dim(true),
		rangeStyle:     backend.DefaultStyle().Reverse(true),
		now:            func() time.Time { return time.Now() },
	}
	for _, opt := range opts {
		opt(cal)
	}
	cal.Base.Role = accessibility.RoleTable
	cal.syncA11y()
	return cal
}

// WithSelectedDateSignal sets the selected date signal.
func WithSelectedDateSignal(sig *state.Signal[time.Time]) CalendarOption {
	return func(c *Calendar) {
		if sig == nil {
			return
		}
		c.selectedDate = sig
	}
}

// WithDisplayedMonthSignal sets the displayed month signal.
func WithDisplayedMonthSignal(sig *state.Signal[time.Time]) CalendarOption {
	return func(c *Calendar) {
		if sig == nil {
			return
		}
		c.displayedMonth = sig
	}
}

// WithMinDateSignal sets the min date signal.
func WithMinDateSignal(sig *state.Signal[*time.Time]) CalendarOption {
	return func(c *Calendar) {
		if sig == nil {
			return
		}
		c.minDate = sig
	}
}

// WithMaxDateSignal sets the max date signal.
func WithMaxDateSignal(sig *state.Signal[*time.Time]) CalendarOption {
	return func(c *Calendar) {
		if sig == nil {
			return
		}
		c.maxDate = sig
	}
}

// WithHighlightDatesSignal sets the highlight dates signal.
func WithHighlightDatesSignal(sig *state.Signal[[]time.Time]) CalendarOption {
	return func(c *Calendar) {
		if sig == nil {
			return
		}
		c.highlightDates = sig
	}
}

// WithWeekStart sets the week start day.
func WithWeekStart(start time.Weekday) CalendarOption {
	return func(c *Calendar) {
		c.weekStart = start
	}
}

// WithShowWeekNumbers toggles week number display.
func WithShowWeekNumbers(show bool) CalendarOption {
	return func(c *Calendar) {
		c.showWeekNums = show
	}
}

// WithSelectionMode sets the selection mode.
func WithSelectionMode(mode CalendarSelectionMode) CalendarOption {
	return func(c *Calendar) {
		c.selectionMode = mode
	}
}

// WithDayRenderer sets a custom day renderer.
func WithDayRenderer(fn DayRenderFunc) CalendarOption {
	return func(c *Calendar) {
		c.dayRenderer = fn
	}
}

// WithNowFunc overrides the clock (useful for tests).
func WithNowFunc(fn func() time.Time) CalendarOption {
	return func(c *Calendar) {
		if fn == nil {
			return
		}
		c.now = fn
	}
}

// Bind attaches app services.
func (c *Calendar) Bind(services runtime.Services) {
	if c == nil {
		return
	}
	c.services = services
	c.subs.Clear()
	c.subs.SetScheduler(services.Scheduler())
	c.observeSignals()
}

// Unbind releases app services.
func (c *Calendar) Unbind() {
	if c == nil {
		return
	}
	c.subs.Clear()
	c.services = runtime.Services{}
}

// SelectedDateSignal returns the selected date signal.
func (c *Calendar) SelectedDateSignal() *state.Signal[time.Time] {
	if c == nil {
		return nil
	}
	return c.selectedDate
}

// DisplayedMonthSignal returns the displayed month signal.
func (c *Calendar) DisplayedMonthSignal() *state.Signal[time.Time] {
	if c == nil {
		return nil
	}
	return c.displayedMonth
}

// RangeStartSignal returns the range start signal.
func (c *Calendar) RangeStartSignal() *state.Signal[*time.Time] {
	if c == nil {
		return nil
	}
	return c.rangeStart
}

// RangeEndSignal returns the range end signal.
func (c *Calendar) RangeEndSignal() *state.Signal[*time.Time] {
	if c == nil {
		return nil
	}
	return c.rangeEnd
}

// SelectedDate returns the selected date.
func (c *Calendar) SelectedDate() time.Time {
	if c == nil || c.selectedDate == nil {
		return time.Time{}
	}
	return c.selectedDate.Get()
}

// SetSelectedDate updates the selected date.
func (c *Calendar) SetSelectedDate(date time.Time) {
	if c == nil {
		return
	}
	c.selectDate(date)
}

// DisplayedMonth returns the month being displayed.
func (c *Calendar) DisplayedMonth() time.Time {
	if c == nil || c.displayedMonth == nil {
		return time.Time{}
	}
	return c.displayedMonth.Get()
}

// RangeStart returns the current range start date.
func (c *Calendar) RangeStart() *time.Time {
	if c == nil || c.rangeStart == nil {
		return nil
	}
	return c.rangeStart.Get()
}

// RangeEnd returns the current range end date.
func (c *Calendar) RangeEnd() *time.Time {
	if c == nil || c.rangeEnd == nil {
		return nil
	}
	return c.rangeEnd.Get()
}

// SetRange updates the selected date range.
func (c *Calendar) SetRange(start, end *time.Time) {
	if c == nil {
		return
	}
	if start != nil {
		s := normalizeDate(*start)
		c.rangeStart.Set(&s)
	} else {
		c.rangeStart.Set(nil)
	}
	if end != nil {
		e := normalizeDate(*end)
		c.rangeEnd.Set(&e)
	} else {
		c.rangeEnd.Set(nil)
	}
	if start != nil {
		c.displayedMonth.Set(monthStart(*start))
	}
	c.invalidate()
}

// SetDisplayedMonth changes the visible month.
func (c *Calendar) SetDisplayedMonth(date time.Time) {
	if c == nil || c.displayedMonth == nil {
		return
	}
	c.displayedMonth.Set(monthStart(date))
	c.invalidate()
}

// SetMinDate sets the minimum selectable date (inclusive).
func (c *Calendar) SetMinDate(date *time.Time) {
	if c == nil || c.minDate == nil {
		return
	}
	if date != nil {
		d := normalizeDate(*date)
		date = &d
	}
	c.minDate.Set(date)
}

// SetMaxDate sets the maximum selectable date (inclusive).
func (c *Calendar) SetMaxDate(date *time.Time) {
	if c == nil || c.maxDate == nil {
		return
	}
	if date != nil {
		d := normalizeDate(*date)
		date = &d
	}
	c.maxDate.Set(date)
}

// SetHighlightDates sets highlighted dates.
func (c *Calendar) SetHighlightDates(dates []time.Time) {
	if c == nil || c.highlightDates == nil {
		return
	}
	c.highlightDates.Set(dates)
}

// SetWeekStart updates the week start day.
func (c *Calendar) SetWeekStart(start time.Weekday) {
	if c == nil {
		return
	}
	c.weekStart = start
	c.invalidate()
}

// SetShowWeekNumbers toggles week number display.
func (c *Calendar) SetShowWeekNumbers(show bool) {
	if c == nil {
		return
	}
	c.showWeekNums = show
	c.invalidate()
}

// SetSelectionMode updates selection mode.
func (c *Calendar) SetSelectionMode(mode CalendarSelectionMode) {
	if c == nil {
		return
	}
	c.selectionMode = mode
	c.invalidate()
}

// SetDayRenderer updates the day renderer.
func (c *Calendar) SetDayRenderer(fn DayRenderFunc) {
	if c == nil {
		return
	}
	c.dayRenderer = fn
	c.invalidate()
}

// OnSelect registers a single-date selection callback.
func (c *Calendar) OnSelect(fn func(time.Time)) {
	if c == nil {
		return
	}
	c.onSelect = fn
}

// OnRangeSelect registers a range selection callback.
func (c *Calendar) OnRangeSelect(fn func(start, end time.Time)) {
	if c == nil {
		return
	}
	c.onRangeSelect = fn
}

// SetStyles configures calendar styles.
func (c *Calendar) SetStyles(base, header, weekday, selected, today, disabled, highlight, outside, rangeSty backend.Style) {
	if c == nil {
		return
	}
	c.style = base
	c.headerStyle = header
	c.weekdayStyle = weekday
	c.selectedStyle = selected
	c.todayStyle = today
	c.disabledStyle = disabled
	c.highlightStyle = highlight
	c.outsideStyle = outside
	c.rangeStyle = rangeSty
}

// StyleType returns the selector type name.
func (c *Calendar) StyleType() string {
	return "Calendar"
}

// Measure returns the desired size.
func (c *Calendar) Measure(constraints runtime.Constraints) runtime.Size {
	return c.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		height := 8
		if contentConstraints.MaxHeight > 0 {
			height = min(height, contentConstraints.MaxHeight)
		}
		width := contentConstraints.MaxWidth
		if width <= 0 {
			width = 28
		}
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: height})
	})
}

// Render draws the calendar grid.
func (c *Calendar) Render(ctx runtime.RenderContext) {
	if c == nil {
		return
	}
	c.syncA11y()
	outer := c.bounds
	content := c.ContentBounds()
	if outer.Width <= 0 || outer.Height <= 0 {
		return
	}
	baseStyle := mergeBackendStyles(resolveBaseStyle(ctx, c, backend.DefaultStyle(), false), c.style)
	ctx.Buffer.Fill(outer, ' ', baseStyle)
	if content.Width <= 0 || content.Height <= 0 {
		return
	}

	layout := c.layout(content)
	if layout.rows <= 0 || layout.cellWidth <= 0 {
		return
	}

	headerStyle := mergeBackendStyles(baseStyle, c.headerStyle)
	weekdayStyle := mergeBackendStyles(baseStyle, c.weekdayStyle)
	selectedStyle := mergeBackendStyles(baseStyle, c.selectedStyle)
	todayStyle := mergeBackendStyles(baseStyle, c.todayStyle)
	disabledStyle := mergeBackendStyles(baseStyle, c.disabledStyle)
	highlightStyle := mergeBackendStyles(baseStyle, c.highlightStyle)
	outsideStyle := mergeBackendStyles(baseStyle, c.outsideStyle)
	rangeStyle := mergeBackendStyles(baseStyle, c.rangeStyle)

	month := c.DisplayedMonth()
	header := month.Format(c.headerFormat)
	headerX := content.X + max(0, (content.Width-len(header))/2)
	ctx.Buffer.SetString(headerX, layout.headerY, header, headerStyle)
	if content.Width >= 2 {
		ctx.Buffer.Set(content.X, layout.headerY, '<', headerStyle)
		ctx.Buffer.Set(content.X+content.Width-1, layout.headerY, '>', headerStyle)
	}

	weekdayNames := weekdayLabels(c.weekStart)
	weekX := content.X + layout.weekNumWidth
	for i := 0; i < len(weekdayNames) && weekX < content.X+content.Width; i++ {
		label := truncateString(weekdayNames[i], layout.cellWidth)
		ctx.Buffer.SetString(weekX, layout.weekdayY, label, weekdayStyle)
		weekX += layout.cellWidth
	}

	date := layout.start
	for row := 0; row < layout.rows; row++ {
		rowY := layout.gridY + row
		if rowY >= content.Y+content.Height {
			break
		}
		if c.showWeekNums && layout.weekNumWidth > 0 {
			_, week := date.ISOWeek()
			label := fmt.Sprintf("%2d", week)
			ctx.Buffer.SetString(content.X, rowY, label, weekdayStyle)
		}
		x := content.X + layout.weekNumWidth
		for col := 0; col < 7; col++ {
			if x >= content.X+content.Width {
				break
			}
			state := c.dayState(date, month)
			cellBounds := runtime.Rect{X: x, Y: rowY, Width: layout.cellWidth, Height: 1}
			if c.dayRenderer != nil {
				c.dayRenderer(ctx.Sub(cellBounds), date, state)
			} else {
				c.renderDayCell(ctx.Sub(cellBounds), date, state, baseStyle, selectedStyle, todayStyle, disabledStyle, highlightStyle, outsideStyle, rangeStyle)
			}
			date = date.AddDate(0, 0, 1)
			x += layout.cellWidth
		}
	}
}

// HandleMessage handles keyboard navigation and mouse selection.
func (c *Calendar) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if c == nil || !c.focused {
		return runtime.Unhandled()
	}
	switch m := msg.(type) {
	case runtime.KeyMsg:
		return c.handleKey(m)
	case runtime.MouseMsg:
		if m.Action == runtime.MousePress && m.Button == runtime.MouseLeft {
			return c.handleClick(m.X, m.Y)
		}
	}
	return runtime.Unhandled()
}

func (c *Calendar) handleKey(key runtime.KeyMsg) runtime.HandleResult {
	current := c.cursorDate()
	switch key.Key {
	case terminal.KeyLeft:
		c.selectDate(current.AddDate(0, 0, -1))
		return runtime.Handled()
	case terminal.KeyRight:
		c.selectDate(current.AddDate(0, 0, 1))
		return runtime.Handled()
	case terminal.KeyUp:
		c.selectDate(current.AddDate(0, 0, -7))
		return runtime.Handled()
	case terminal.KeyDown:
		c.selectDate(current.AddDate(0, 0, 7))
		return runtime.Handled()
	case terminal.KeyPageUp:
		c.selectDate(addMonthsClamped(current, -1))
		return runtime.Handled()
	case terminal.KeyPageDown:
		c.selectDate(addMonthsClamped(current, 1))
		return runtime.Handled()
	case terminal.KeyHome:
		c.selectDate(monthStart(current))
		return runtime.Handled()
	case terminal.KeyEnd:
		c.selectDate(monthEnd(current))
		return runtime.Handled()
	case terminal.KeyEnter:
		c.selectDate(current)
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func (c *Calendar) handleClick(x, y int) runtime.HandleResult {
	content := c.ContentBounds()
	if !content.Contains(x, y) {
		return runtime.Unhandled()
	}
	layout := c.layout(content)
	if layout.rows <= 0 {
		return runtime.Unhandled()
	}
	if y == layout.headerY {
		if x == content.X {
			c.selectDate(addMonthsClamped(c.cursorDate(), -1))
			return runtime.Handled()
		}
		if x == content.X+content.Width-1 {
			c.selectDate(addMonthsClamped(c.cursorDate(), 1))
			return runtime.Handled()
		}
	}
	if y < layout.gridY {
		return runtime.Unhandled()
	}
	row := y - layout.gridY
	if row < 0 || row >= layout.rows {
		return runtime.Unhandled()
	}
	col := (x - (content.X + layout.weekNumWidth)) / layout.cellWidth
	if col < 0 || col >= 7 {
		return runtime.Unhandled()
	}
	offset := row*7 + col
	date := layout.start.AddDate(0, 0, offset)
	if c.isDisabled(date) {
		return runtime.Unhandled()
	}
	c.selectDate(date)
	return runtime.Handled()
}

func (c *Calendar) renderDayCell(ctx runtime.RenderContext, date time.Time, state CalendarDayState, base, selected, today, disabled, highlight, outside, rangeSty backend.Style) {
	style := base
	if state.Disabled {
		style = disabled
	}
	if !state.InMonth {
		style = outside
	}
	if state.Highlighted {
		style = highlight
	}
	if state.InRange {
		style = rangeSty
	}
	if state.Today {
		style = mergeBackendStyles(style, today)
	}
	if state.Selected {
		style = selected
	}
	day := fmt.Sprintf("%2d", date.Day())
	ctx.Buffer.SetString(ctx.Bounds.X, ctx.Bounds.Y, truncateString(day, ctx.Bounds.Width), style)
}

func (c *Calendar) dayState(date time.Time, month time.Time) CalendarDayState {
	state := CalendarDayState{
		InMonth: month.Year() == date.Year() && month.Month() == date.Month(),
	}
	state.Disabled = c.isDisabled(date)
	state.Today = sameDay(date, normalizeDate(c.now()))
	state.Highlighted = c.isHighlighted(date)

	selected := c.selectedDate.Get()
	state.Selected = sameDay(date, selected)

	if c.selectionMode == CalendarSelectionRange {
		start := c.rangeStart.Get()
		end := c.rangeEnd.Get()
		if start != nil {
			state.RangeStart = sameDay(date, *start)
		}
		if end != nil {
			state.RangeEnd = sameDay(date, *end)
		}
		if start != nil && end != nil {
			s := normalizeDate(*start)
			e := normalizeDate(*end)
			if e.Before(s) {
				s, e = e, s
			}
			state.InRange = !date.Before(s) && !date.After(e)
		}
	}
	return state
}

func (c *Calendar) cursorDate() time.Time {
	if c == nil || c.selectedDate == nil {
		return normalizeDate(c.now())
	}
	current := c.selectedDate.Get()
	if current.IsZero() {
		return normalizeDate(c.now())
	}
	return normalizeDate(current)
}

func (c *Calendar) isHighlighted(date time.Time) bool {
	if c == nil || c.highlightDates == nil {
		return false
	}
	highlights := c.highlightDates.Get()
	for _, d := range highlights {
		if sameDay(d, date) {
			return true
		}
	}
	return false
}

func (c *Calendar) isDisabled(date time.Time) bool {
	if c == nil {
		return false
	}
	date = normalizeDate(date)
	if c.minDate != nil {
		if min := c.minDate.Get(); min != nil && date.Before(*min) {
			return true
		}
	}
	if c.maxDate != nil {
		if max := c.maxDate.Get(); max != nil && date.After(*max) {
			return true
		}
	}
	return false
}

func (c *Calendar) selectDate(date time.Time) {
	if c == nil {
		return
	}
	date = normalizeDate(date)
	date = c.clampDate(date)
	if c.selectedDate != nil {
		c.selectedDate.Set(date)
	}
	if c.selectionMode == CalendarSelectionRange {
		start := c.rangeStart.Get()
		end := c.rangeEnd.Get()
		if start == nil || (start != nil && end != nil) {
			c.rangeStart.Set(&date)
			c.rangeEnd.Set(nil)
		} else {
			s := normalizeDate(*start)
			e := normalizeDate(date)
			if e.Before(s) {
				s, e = e, s
			}
			c.rangeStart.Set(&s)
			c.rangeEnd.Set(&e)
			if c.onRangeSelect != nil {
				c.onRangeSelect(s, e)
			}
		}
	} else if c.onSelect != nil {
		c.onSelect(date)
	}
	if c.displayedMonth != nil {
		c.displayedMonth.Set(monthStart(date))
	}
	c.invalidate()
}

func (c *Calendar) clampDate(date time.Time) time.Time {
	if c == nil {
		return date
	}
	if c.minDate != nil {
		if min := c.minDate.Get(); min != nil && date.Before(*min) {
			return *min
		}
	}
	if c.maxDate != nil {
		if max := c.maxDate.Get(); max != nil && date.After(*max) {
			return *max
		}
	}
	return date
}

func (c *Calendar) invalidate() {
	if c == nil {
		return
	}
	c.Invalidate()
	c.services.Invalidate()
}

func (c *Calendar) observeSignals() {
	if c == nil {
		return
	}
	observe := func(sub state.Subscribable) {
		c.subs.Observe(sub, func() {
			c.invalidate()
			c.syncA11y()
		})
	}
	observe(c.selectedDate)
	observe(c.displayedMonth)
	observe(c.minDate)
	observe(c.maxDate)
	observe(c.highlightDates)
	observe(c.rangeStart)
	observe(c.rangeEnd)
}

func (c *Calendar) syncA11y() {
	if c == nil {
		return
	}
	if c.Base.Role == "" {
		c.Base.Role = accessibility.RoleTable
	}
	label := strings.TrimSpace(c.label)
	if label == "" {
		label = "Calendar"
	}
	c.Base.Label = label
	month := c.DisplayedMonth()
	if !month.IsZero() {
		c.Base.Description = month.Format(c.headerFormat)
	}
	value := c.SelectedDate()
	if c.selectionMode == CalendarSelectionRange {
		start := c.rangeStart.Get()
		end := c.rangeEnd.Get()
		if start != nil && end != nil {
			c.Base.Value = &accessibility.ValueInfo{Text: fmt.Sprintf("%s - %s", start.Format("2006-01-02"), end.Format("2006-01-02"))}
		} else if start != nil {
			c.Base.Value = &accessibility.ValueInfo{Text: start.Format("2006-01-02")}
		} else {
			c.Base.Value = nil
		}
		return
	}
	if !value.IsZero() {
		c.Base.Value = &accessibility.ValueInfo{Text: value.Format("2006-01-02")}
	} else {
		c.Base.Value = nil
	}
}

type calendarLayout struct {
	headerY      int
	weekdayY     int
	gridY        int
	rows         int
	cellWidth    int
	weekNumWidth int
	start        time.Time
}

func (c *Calendar) layout(bounds runtime.Rect) calendarLayout {
	layout := calendarLayout{
		headerY:  bounds.Y,
		weekdayY: bounds.Y + 1,
		gridY:    bounds.Y + 2,
		rows:     max(0, bounds.Height-2),
	}
	if layout.rows > 6 {
		layout.rows = 6
	}
	if layout.rows <= 0 {
		return layout
	}
	if c.showWeekNums {
		layout.weekNumWidth = 3
	}
	available := bounds.Width - layout.weekNumWidth
	if available < 7 {
		layout.cellWidth = max(1, available)
	} else {
		layout.cellWidth = max(2, available/7)
	}
	month := c.DisplayedMonth()
	if month.IsZero() {
		month = monthStart(c.now())
	}
	layout.start = calendarStart(month, c.weekStart)
	return layout
}

func weekdayLabels(start time.Weekday) []string {
	names := []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	if start == time.Sunday {
		return names
	}
	out := make([]string, 7)
	for i := 0; i < 7; i++ {
		out[i] = names[(int(start)+i)%7]
	}
	return out
}

func calendarStart(month time.Time, weekStart time.Weekday) time.Time {
	first := monthStart(month)
	offset := (int(first.Weekday()) - int(weekStart) + 7) % 7
	return first.AddDate(0, 0, -offset)
}

func monthStart(date time.Time) time.Time {
	date = normalizeDate(date)
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
}

func monthEnd(date time.Time) time.Time {
	start := monthStart(date)
	return start.AddDate(0, 1, -1)
}

func addMonthsClamped(date time.Time, delta int) time.Time {
	start := monthStart(date)
	target := start.AddDate(0, delta, 0)
	day := date.Day()
	last := daysInMonth(target.Year(), target.Month())
	if day > last {
		day = last
	}
	return time.Date(target.Year(), target.Month(), day, 0, 0, 0, 0, date.Location())
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()
}

func normalizeDate(date time.Time) time.Time {
	if date.IsZero() {
		return date
	}
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
}

func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

func equalDay(a, b time.Time) bool {
	return sameDay(a, b)
}

func equalMonth(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month()
}

var _ runtime.Widget = (*Calendar)(nil)
var _ runtime.Focusable = (*Calendar)(nil)
var _ runtime.Bindable = (*Calendar)(nil)
var _ runtime.Unbindable = (*Calendar)(nil)
