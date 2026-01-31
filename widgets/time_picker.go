package widgets

import (
	"fmt"
	"strings"
	"time"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

// TimePicker allows selecting a time of day.
type TimePicker struct {
	FocusableBase

	hour        int
	minute      int
	second      int
	showSeconds bool
	selected    int
	label       string
	style       backend.Style
	selectedSty backend.Style
	baseDate    time.Time
	now         func() time.Time

	digitBuf  string
	lastDigit time.Time

	onChange func(time.Time)
	onSubmit func(time.Time)
}

// NewTimePicker creates a new time picker.
func NewTimePicker() *TimePicker {
	now := time.Now()
	base := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	p := &TimePicker{
		hour:        now.Hour(),
		minute:      now.Minute(),
		second:      now.Second(),
		showSeconds: false,
		selected:    0,
		label:       "Time Picker",
		style:       backend.DefaultStyle(),
		selectedSty: backend.DefaultStyle().Reverse(true),
		baseDate:    base,
		now:         time.Now,
	}
	p.Base.Role = accessibility.RoleTextbox
	p.syncA11y()
	return p
}

// SetShowSeconds toggles second display.
func (t *TimePicker) SetShowSeconds(show bool) {
	if t == nil {
		return
	}
	t.showSeconds = show
	if !show && t.selected > 1 {
		t.selected = 1
	}
}

// SetLabel updates the accessibility label.
func (t *TimePicker) SetLabel(label string) {
	if t == nil {
		return
	}
	t.label = label
	t.syncA11y()
}

// SetOnChange registers a change callback.
func (t *TimePicker) SetOnChange(fn func(time.Time)) {
	if t == nil {
		return
	}
	t.onChange = fn
}

// SetOnSubmit registers a submit callback.
func (t *TimePicker) SetOnSubmit(fn func(time.Time)) {
	if t == nil {
		return
	}
	t.onSubmit = fn
}

// SetTime updates the selected time.
func (t *TimePicker) SetTime(value time.Time) {
	if t == nil {
		return
	}
	t.baseDate = time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
	t.hour = value.Hour()
	t.minute = value.Minute()
	t.second = value.Second()
	t.syncA11y()
}

// Time returns the current time selection.
func (t *TimePicker) Time() time.Time {
	if t == nil {
		return time.Time{}
	}
	return time.Date(t.baseDate.Year(), t.baseDate.Month(), t.baseDate.Day(), t.hour, t.minute, t.second, 0, t.baseDate.Location())
}

// StyleType returns the selector type name.
func (t *TimePicker) StyleType() string {
	return "TimePicker"
}

// Measure returns desired size.
func (t *TimePicker) Measure(constraints runtime.Constraints) runtime.Size {
	width := 5
	if t.showSeconds {
		width = 8
	}
	return constraints.Constrain(runtime.Size{Width: width, Height: 1})
}

// Render draws the time.
func (t *TimePicker) Render(ctx runtime.RenderContext) {
	if t == nil {
		return
	}
	t.syncA11y()
	outer := t.bounds
	if outer.Width <= 0 || outer.Height <= 0 {
		return
	}
	style := resolveBaseStyle(ctx, t, t.style, true)
	selectedStyle := t.selectedSty
	ctx.Buffer.Fill(outer, ' ', style)

	x := outer.X
	fields := []string{fmt.Sprintf("%02d", t.hour), fmt.Sprintf("%02d", t.minute)}
	if t.showSeconds {
		fields = append(fields, fmt.Sprintf("%02d", t.second))
	}
	for idx, field := range fields {
		fieldStyle := style
		if idx == t.selected {
			fieldStyle = selectedStyle
		}
		ctx.Buffer.SetString(x, outer.Y, field, fieldStyle)
		x += textWidth(field)
		if idx < len(fields)-1 {
			ctx.Buffer.SetString(x, outer.Y, ":", style)
			x += 1
		}
	}
}

// HandleMessage processes keyboard input.
func (t *TimePicker) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if t == nil || !t.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	switch key.Key {
	case terminal.KeyLeft:
		if t.selected > 0 {
			t.selected--
		}
		return runtime.Handled()
	case terminal.KeyRight:
		if t.selected < t.maxFieldIndex() {
			t.selected++
		}
		return runtime.Handled()
	case terminal.KeyUp:
		t.incrementField(1)
		return runtime.Handled()
	case terminal.KeyDown:
		t.incrementField(-1)
		return runtime.Handled()
	case terminal.KeyEnter:
		if t.onSubmit != nil {
			t.onSubmit(t.Time())
		}
		return runtime.Handled()
	case terminal.KeyRune:
		if key.Rune >= '0' && key.Rune <= '9' {
			t.handleDigit(int(key.Rune - '0'))
			return runtime.Handled()
		}
	}
	return runtime.Unhandled()
}

func (t *TimePicker) maxFieldIndex() int {
	if t.showSeconds {
		return 2
	}
	return 1
}

func (t *TimePicker) incrementField(delta int) {
	if t == nil {
		return
	}
	apply := func(value, max int) int {
		value += delta
		for value < 0 {
			value += max
		}
		return value % max
	}
	switch t.selected {
	case 0:
		t.hour = apply(t.hour, 24)
	case 1:
		t.minute = apply(t.minute, 60)
	case 2:
		t.second = apply(t.second, 60)
	}
	t.notifyChange()
}

func (t *TimePicker) handleDigit(digit int) {
	if t == nil {
		return
	}
	now := t.now()
	if now.Sub(t.lastDigit) > time.Second {
		t.digitBuf = ""
	}
	t.lastDigit = now
	t.digitBuf += fmt.Sprintf("%d", digit)
	if len(t.digitBuf) > 2 {
		t.digitBuf = t.digitBuf[len(t.digitBuf)-2:]
	}
	value := 0
	if len(t.digitBuf) == 1 {
		value = digit
	} else {
		_, _ = fmt.Sscanf(t.digitBuf, "%02d", &value)
	}
	t.setFieldValue(value)
	if len(t.digitBuf) == 2 {
		t.digitBuf = ""
		if t.selected < t.maxFieldIndex() {
			t.selected++
		}
	}
}

func (t *TimePicker) setFieldValue(value int) {
	switch t.selected {
	case 0:
		t.hour = clampInt(value, 0, 23)
	case 1:
		t.minute = clampInt(value, 0, 59)
	case 2:
		t.second = clampInt(value, 0, 59)
	}
	t.notifyChange()
}

func (t *TimePicker) notifyChange() {
	if t.onChange != nil {
		t.onChange(t.Time())
	}
	t.syncA11y()
}

func (t *TimePicker) syncA11y() {
	if t == nil {
		return
	}
	if t.Base.Role == "" {
		t.Base.Role = accessibility.RoleTextbox
	}
	label := strings.TrimSpace(t.label)
	if label == "" {
		label = "Time Picker"
	}
	t.Base.Label = label
	t.Base.Value = &accessibility.ValueInfo{Text: t.Time().Format(t.format())}
}

func (t *TimePicker) format() string {
	if t.showSeconds {
		return "15:04:05"
	}
	return "15:04"
}

var _ runtime.Widget = (*TimePicker)(nil)
var _ runtime.Focusable = (*TimePicker)(nil)
