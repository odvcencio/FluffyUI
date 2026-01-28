package widgets

import (
	"fmt"
	"math"
	"strings"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	"github.com/odvcencio/fluffyui/terminal"
)

// Orientation describes slider orientation.
type Orientation int

const (
	Horizontal Orientation = iota
	Vertical
)

const (
	sliderTrackCharH        = '-'
	sliderTrackCharV        = '|'
	sliderFillCharH         = '='
	sliderFillCharV         = '|'
	sliderThumbChar         = 'O'
	sliderThumbCharInactive = 'o'
)

// SliderOption configures slider behavior.
type SliderOption func(*Slider)

// RangeSliderOption configures range slider behavior.
type RangeSliderOption func(*RangeSlider)

// Slider is a focusable value slider.
type Slider struct {
	FocusableBase

	value       *state.Signal[float64]
	min         float64
	max         float64
	step        float64
	orientation Orientation
	showValue   bool
	valueFormat string

	label      string
	trackStyle backend.Style
	thumbStyle backend.Style
	fillStyle  backend.Style
	style      backend.Style

	dragging bool
	services runtime.Services
	subs     state.Subscriptions
}

// NewSlider creates a new slider.
func NewSlider(value *state.Signal[float64], opts ...SliderOption) *Slider {
	if value == nil {
		value = state.NewSignal(0.0)
	}
	value.SetEqualFunc(state.EqualComparable[float64])
	s := &Slider{
		value:       value,
		min:         0,
		max:         100,
		step:        1,
		orientation: Horizontal,
		showValue:   false,
		valueFormat: "%.0f",
		label:       "Slider",
		trackStyle:  backend.DefaultStyle(),
		thumbStyle:  backend.DefaultStyle().Bold(true),
		fillStyle:   backend.DefaultStyle().Reverse(true),
		style:       backend.DefaultStyle(),
	}
	for _, opt := range opts {
		opt(s)
	}
	s.Base.Role = accessibility.RoleSlider
	s.syncA11y()
	s.SetValue(s.value.Get())
	return s
}

// WithSliderRange configures min, max, and step.
func WithSliderRange(min, max, step float64) SliderOption {
	return func(s *Slider) {
		s.min = min
		s.max = max
		s.step = step
	}
}

// WithSliderOrientation sets orientation.
func WithSliderOrientation(orientation Orientation) SliderOption {
	return func(s *Slider) {
		s.orientation = orientation
	}
}

// WithSliderShowValue toggles value label.
func WithSliderShowValue(show bool) SliderOption {
	return func(s *Slider) {
		s.showValue = show
	}
}

// WithSliderValueFormat sets format string.
func WithSliderValueFormat(format string) SliderOption {
	return func(s *Slider) {
		format = strings.TrimSpace(format)
		if format != "" {
			s.valueFormat = format
		}
	}
}

// WithSliderStyles configures styles.
func WithSliderStyles(track, thumb, fill backend.Style) SliderOption {
	return func(s *Slider) {
		s.trackStyle = track
		s.thumbStyle = thumb
		s.fillStyle = fill
	}
}

// WithRangeSliderRange configures min, max, and step.
func WithRangeSliderRange(min, max, step float64) RangeSliderOption {
	return func(r *RangeSlider) {
		r.min = min
		r.max = max
		r.step = step
	}
}

// WithRangeSliderOrientation sets orientation.
func WithRangeSliderOrientation(orientation Orientation) RangeSliderOption {
	return func(r *RangeSlider) {
		r.orientation = orientation
	}
}

// WithRangeSliderShowValue toggles value label.
func WithRangeSliderShowValue(show bool) RangeSliderOption {
	return func(r *RangeSlider) {
		r.showValue = show
	}
}

// WithRangeSliderValueFormat sets format string.
func WithRangeSliderValueFormat(format string) RangeSliderOption {
	return func(r *RangeSlider) {
		format = strings.TrimSpace(format)
		if format != "" {
			r.valueFormat = format
		}
	}
}

// WithRangeSliderStyles configures styles.
func WithRangeSliderStyles(track, thumb, fill backend.Style) RangeSliderOption {
	return func(r *RangeSlider) {
		r.trackStyle = track
		r.thumbStyle = thumb
		r.fillStyle = fill
	}
}

// SetRange updates min/max/step.
func (s *Slider) SetRange(min, max, step float64) {
	if s == nil {
		return
	}
	s.min = min
	s.max = max
	s.step = step
	s.SetValue(s.Value())
}

// SetValue updates the slider value.
func (s *Slider) SetValue(value float64) {
	if s == nil || s.value == nil {
		return
	}
	value = s.clampValue(value)
	s.value.Set(value)
	s.syncA11y()
	s.services.Invalidate()
}

// Value returns the current value.
func (s *Slider) Value() float64 {
	if s == nil || s.value == nil {
		return 0
	}
	return s.value.Get()
}

// SetLabel updates the accessibility label.
func (s *Slider) SetLabel(label string) {
	if s == nil {
		return
	}
	s.label = label
	s.syncA11y()
}

// SetStyles updates slider styles.
func (s *Slider) SetStyles(base, track, thumb, fill backend.Style) {
	if s == nil {
		return
	}
	s.style = base
	s.trackStyle = track
	s.thumbStyle = thumb
	s.fillStyle = fill
}

// StyleType returns the selector type name.
func (s *Slider) StyleType() string {
	return "Slider"
}

// Bind attaches app services.
func (s *Slider) Bind(services runtime.Services) {
	if s == nil {
		return
	}
	s.services = services
	s.subs.Clear()
	s.subs.SetScheduler(services.Scheduler())
	if s.value != nil {
		s.subs.Observe(s.value, func() {
			s.syncA11y()
			s.services.Invalidate()
		})
	}
}

// Unbind releases app services.
func (s *Slider) Unbind() {
	if s == nil {
		return
	}
	s.subs.Clear()
	s.services = runtime.Services{}
}

// Measure returns desired size.
func (s *Slider) Measure(constraints runtime.Constraints) runtime.Size {
	return s.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		if s.orientation == Vertical {
			height := contentConstraints.MaxHeight
			if height <= 0 {
				height = 5
			}
			return contentConstraints.Constrain(runtime.Size{Width: 1, Height: height})
		}
		width := contentConstraints.MaxWidth
		if width <= 0 {
			width = 10
		}
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: 1})
	})
}

// Render draws the slider.
func (s *Slider) Render(ctx runtime.RenderContext) {
	if s == nil {
		return
	}
	s.syncA11y()
	outer := s.bounds
	content := s.ContentBounds()
	if outer.Width <= 0 || outer.Height <= 0 {
		return
	}
	baseStyle := mergeBackendStyles(resolveBaseStyle(ctx, s, backend.DefaultStyle(), false), s.style)
	trackStyle := mergeBackendStyles(baseStyle, s.trackStyle)
	thumbStyle := mergeBackendStyles(baseStyle, s.thumbStyle)
	fillStyle := mergeBackendStyles(baseStyle, s.fillStyle)
	ctx.Buffer.Fill(outer, ' ', baseStyle)
	if content.Width <= 0 || content.Height <= 0 {
		return
	}

	valueText := ""
	if s.showValue {
		valueText = fmt.Sprintf(s.valueFormat, s.Value())
	}
	trackRect, valueRect := sliderTrackRect(content, s.orientation, valueText)
	if trackRect.Width <= 0 || trackRect.Height <= 0 {
		return
	}

	if s.orientation == Vertical {
		s.renderVertical(ctx, trackRect, valueRect, valueText, trackStyle, thumbStyle, fillStyle)
		return
	}
	s.renderHorizontal(ctx, trackRect, valueRect, valueText, trackStyle, thumbStyle, fillStyle)
}

func (s *Slider) renderHorizontal(ctx runtime.RenderContext, trackRect, valueRect runtime.Rect, valueText string, trackStyle, thumbStyle, fillStyle backend.Style) {
	length := trackRect.Width
	if length <= 0 {
		return
	}
	pos := s.valueToPos(length)
	for i := 0; i < length; i++ {
		ch := sliderTrackCharH
		style := trackStyle
		if i <= pos {
			ch = sliderFillCharH
			style = fillStyle
		}
		ctx.Buffer.Set(trackRect.X+i, trackRect.Y, ch, style)
	}
	ctx.Buffer.Set(trackRect.X+pos, trackRect.Y, sliderThumbChar, thumbStyle)
	if s.showValue && valueText != "" && valueRect.Width > 0 {
		ctx.Buffer.SetString(valueRect.X, valueRect.Y, truncateString(valueText, valueRect.Width), trackStyle)
	}
}

func (s *Slider) renderVertical(ctx runtime.RenderContext, trackRect, valueRect runtime.Rect, valueText string, trackStyle, thumbStyle, fillStyle backend.Style) {
	length := trackRect.Height
	if length <= 0 {
		return
	}
	pos := s.valueToPos(length)
	for i := 0; i < length; i++ {
		y := trackRect.Y + (length - 1 - i)
		ch := sliderTrackCharV
		style := trackStyle
		if i <= pos {
			ch = sliderFillCharV
			style = fillStyle
		}
		ctx.Buffer.Set(trackRect.X, y, ch, style)
	}
	thumbY := trackRect.Y + (length - 1 - pos)
	ctx.Buffer.Set(trackRect.X, thumbY, sliderThumbChar, thumbStyle)
	if s.showValue && valueText != "" && valueRect.Width > 0 {
		ctx.Buffer.SetString(valueRect.X, valueRect.Y, truncateString(valueText, valueRect.Width), trackStyle)
	}
}

// HandleMessage updates value on input.
func (s *Slider) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if s == nil {
		return runtime.Unhandled()
	}
	switch m := msg.(type) {
	case runtime.KeyMsg:
		if !s.focused {
			return runtime.Unhandled()
		}
		return s.handleKey(m)
	case runtime.MouseMsg:
		return s.handleMouse(m)
	}
	return runtime.Unhandled()
}

func (s *Slider) handleKey(key runtime.KeyMsg) runtime.HandleResult {
	step := s.stepSize()
	page := s.pageStep()
	switch key.Key {
	case terminal.KeyLeft, terminal.KeyDown:
		s.SetValue(s.Value() - step)
		return runtime.Handled()
	case terminal.KeyRight, terminal.KeyUp:
		s.SetValue(s.Value() + step)
		return runtime.Handled()
	case terminal.KeyPageDown:
		s.SetValue(s.Value() - page)
		return runtime.Handled()
	case terminal.KeyPageUp:
		s.SetValue(s.Value() + page)
		return runtime.Handled()
	case terminal.KeyHome:
		s.SetValue(s.min)
		return runtime.Handled()
	case terminal.KeyEnd:
		s.SetValue(s.max)
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func (s *Slider) handleMouse(m runtime.MouseMsg) runtime.HandleResult {
	content := s.ContentBounds()
	if !content.Contains(m.X, m.Y) {
		if m.Action == runtime.MouseRelease && m.Button == runtime.MouseLeft {
			s.dragging = false
		}
		return runtime.Unhandled()
	}
	if m.Button != runtime.MouseLeft {
		return runtime.Unhandled()
	}
	valueText := ""
	if s.showValue {
		valueText = fmt.Sprintf(s.valueFormat, s.Value())
	}
	trackRect, _ := sliderTrackRect(content, s.orientation, valueText)
	if trackRect.Width <= 0 || trackRect.Height <= 0 {
		return runtime.Unhandled()
	}
	switch m.Action {
	case runtime.MousePress:
		s.dragging = true
		s.SetValue(s.valueFromPoint(m.X, m.Y, trackRect))
		return runtime.Handled()
	case runtime.MouseMove:
		if s.dragging {
			s.SetValue(s.valueFromPoint(m.X, m.Y, trackRect))
			return runtime.Handled()
		}
	case runtime.MouseRelease:
		if s.dragging {
			s.dragging = false
			s.SetValue(s.valueFromPoint(m.X, m.Y, trackRect))
			return runtime.Handled()
		}
	}
	return runtime.Unhandled()
}

func (s *Slider) valueFromPoint(x, y int, trackRect runtime.Rect) float64 {
	if s.orientation == Vertical {
		length := max(1, trackRect.Height)
		offset := trackRect.Y + trackRect.Height - 1 - y
		if offset < 0 {
			offset = 0
		}
		if offset >= length {
			offset = length - 1
		}
		return s.posToValue(offset, length)
	}
	length := max(1, trackRect.Width)
	offset := x - trackRect.X
	if offset < 0 {
		offset = 0
	}
	if offset >= length {
		offset = length - 1
	}
	return s.posToValue(offset, length)
}

func (s *Slider) valueToPos(length int) int {
	if length <= 1 {
		return 0
	}
	ratio := s.ratio()
	pos := int(math.Round(ratio * float64(length-1)))
	if pos < 0 {
		return 0
	}
	if pos >= length {
		return length - 1
	}
	return pos
}

func (s *Slider) posToValue(pos, length int) float64 {
	if length <= 1 {
		return s.min
	}
	ratio := float64(pos) / float64(length-1)
	value := s.min + ratio*(s.max-s.min)
	return s.clampValue(value)
}

func (s *Slider) ratio() float64 {
	min := s.min
	max := s.max
	if max <= min {
		return 0
	}
	value := s.Value()
	return (value - min) / (max - min)
}

func (s *Slider) clampValue(value float64) float64 {
	if s == nil {
		return value
	}
	if s.max < s.min {
		s.max, s.min = s.min, s.max
	}
	if s.step > 0 {
		steps := math.Round((value - s.min) / s.step)
		value = s.min + steps*s.step
	}
	if value < s.min {
		value = s.min
	}
	if value > s.max {
		value = s.max
	}
	return value
}

func (s *Slider) stepSize() float64 {
	if s.step > 0 {
		return s.step
	}
	return (s.max - s.min) / 100
}

func (s *Slider) pageStep() float64 {
	step := s.stepSize()
	if step <= 0 {
		return 1
	}
	return step * 10
}

func (s *Slider) syncA11y() {
	if s == nil {
		return
	}
	if s.Base.Role == "" {
		s.Base.Role = accessibility.RoleSlider
	}
	label := strings.TrimSpace(s.label)
	if label == "" {
		label = "Slider"
	}
	s.Base.Label = label
	s.Base.Value = &accessibility.ValueInfo{Text: fmt.Sprintf(s.valueFormat, s.Value())}
}

func sliderTrackRect(bounds runtime.Rect, orientation Orientation, valueText string) (runtime.Rect, runtime.Rect) {
	if orientation == Vertical {
		track := bounds
		valueRect := runtime.Rect{}
		if valueText != "" && bounds.Height > 1 {
			track.Height = bounds.Height - 1
			valueRect = runtime.Rect{X: bounds.X, Y: bounds.Y + bounds.Height - 1, Width: bounds.Width, Height: 1}
		}
		return track, valueRect
	}
	track := bounds
	valueRect := runtime.Rect{}
	valueWidth := textWidth(valueText)
	if valueText != "" && bounds.Width > valueWidth+1 {
		track.Width = bounds.Width - valueWidth - 1
		valueRect = runtime.Rect{X: bounds.X + track.Width + 1, Y: bounds.Y, Width: valueWidth, Height: 1}
	}
	return track, valueRect
}

var _ runtime.Widget = (*Slider)(nil)
var _ runtime.Focusable = (*Slider)(nil)
var _ runtime.Bindable = (*Slider)(nil)
var _ runtime.Unbindable = (*Slider)(nil)

type rangeHandle int

const (
	rangeHandleMin rangeHandle = iota
	rangeHandleMax
)

// RangeSlider is a dual-handle slider.
type RangeSlider struct {
	FocusableBase

	minValue    *state.Signal[float64]
	maxValue    *state.Signal[float64]
	min         float64
	max         float64
	step        float64
	orientation Orientation
	showValue   bool
	valueFormat string

	label      string
	trackStyle backend.Style
	thumbStyle backend.Style
	fillStyle  backend.Style
	style      backend.Style

	active   rangeHandle
	dragging bool
	services runtime.Services
	subs     state.Subscriptions
}

// NewRangeSlider creates a range slider.
func NewRangeSlider(minValue, maxValue *state.Signal[float64], opts ...RangeSliderOption) *RangeSlider {
	if minValue == nil {
		minValue = state.NewSignal(0.0)
	}
	if maxValue == nil {
		maxValue = state.NewSignal(0.0)
	}
	minValue.SetEqualFunc(state.EqualComparable[float64])
	maxValue.SetEqualFunc(state.EqualComparable[float64])

	rs := &RangeSlider{
		minValue:    minValue,
		maxValue:    maxValue,
		min:         0,
		max:         100,
		step:        1,
		orientation: Horizontal,
		showValue:   false,
		valueFormat: "%.0f",
		label:       "Range Slider",
		trackStyle:  backend.DefaultStyle(),
		thumbStyle:  backend.DefaultStyle().Bold(true),
		fillStyle:   backend.DefaultStyle().Reverse(true),
		style:       backend.DefaultStyle(),
		active:      rangeHandleMin,
	}
	for _, opt := range opts {
		opt(rs)
	}
	rs.Base.Role = accessibility.RoleSlider
	rs.syncA11y()
	rs.clampValues()
	return rs
}

// SetRange updates min/max/step.
func (r *RangeSlider) SetRange(min, max, step float64) {
	if r == nil {
		return
	}
	r.min = min
	r.max = max
	r.step = step
	r.clampValues()
}

// SetValues updates the min/max values.
func (r *RangeSlider) SetValues(minValue, maxValue float64) {
	if r == nil {
		return
	}
	r.setMinValue(minValue)
	r.setMaxValue(maxValue)
	r.clampValues()
}

// Values returns the current min/max values.
func (r *RangeSlider) Values() (float64, float64) {
	if r == nil {
		return 0, 0
	}
	return r.minValue.Get(), r.maxValue.Get()
}

// SetLabel updates the accessibility label.
func (r *RangeSlider) SetLabel(label string) {
	if r == nil {
		return
	}
	r.label = label
	r.syncA11y()
}

// SetStyles updates slider styles.
func (r *RangeSlider) SetStyles(base, track, thumb, fill backend.Style) {
	if r == nil {
		return
	}
	r.style = base
	r.trackStyle = track
	r.thumbStyle = thumb
	r.fillStyle = fill
}

// StyleType returns the selector type name.
func (r *RangeSlider) StyleType() string {
	return "RangeSlider"
}

// Bind attaches app services.
func (r *RangeSlider) Bind(services runtime.Services) {
	if r == nil {
		return
	}
	r.services = services
	r.subs.Clear()
	r.subs.SetScheduler(services.Scheduler())
	if r.minValue != nil {
		r.subs.Observe(r.minValue, func() {
			r.clampValues()
			r.syncA11y()
			r.services.Invalidate()
		})
	}
	if r.maxValue != nil {
		r.subs.Observe(r.maxValue, func() {
			r.clampValues()
			r.syncA11y()
			r.services.Invalidate()
		})
	}
}

// Unbind releases app services.
func (r *RangeSlider) Unbind() {
	if r == nil {
		return
	}
	r.subs.Clear()
	r.services = runtime.Services{}
}

// Measure returns desired size.
func (r *RangeSlider) Measure(constraints runtime.Constraints) runtime.Size {
	return r.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		if r.orientation == Vertical {
			height := contentConstraints.MaxHeight
			if height <= 0 {
				height = 5
			}
			return contentConstraints.Constrain(runtime.Size{Width: 1, Height: height})
		}
		width := contentConstraints.MaxWidth
		if width <= 0 {
			width = 12
		}
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: 1})
	})
}

// Render draws the range slider.
func (r *RangeSlider) Render(ctx runtime.RenderContext) {
	if r == nil {
		return
	}
	r.syncA11y()
	outer := r.bounds
	content := r.ContentBounds()
	if outer.Width <= 0 || outer.Height <= 0 {
		return
	}
	baseStyle := mergeBackendStyles(resolveBaseStyle(ctx, r, backend.DefaultStyle(), false), r.style)
	trackStyle := mergeBackendStyles(baseStyle, r.trackStyle)
	thumbStyle := mergeBackendStyles(baseStyle, r.thumbStyle)
	fillStyle := mergeBackendStyles(baseStyle, r.fillStyle)
	ctx.Buffer.Fill(outer, ' ', baseStyle)
	if content.Width <= 0 || content.Height <= 0 {
		return
	}

	minValue, maxValue := r.Values()
	valueText := ""
	if r.showValue {
		valueText = fmt.Sprintf(r.valueFormat, minValue) + " - " + fmt.Sprintf(r.valueFormat, maxValue)
	}
	trackRect, valueRect := sliderTrackRect(content, r.orientation, valueText)
	if trackRect.Width <= 0 || trackRect.Height <= 0 {
		return
	}

	if r.orientation == Vertical {
		r.renderVertical(ctx, trackRect, valueRect, valueText, trackStyle, thumbStyle, fillStyle)
		return
	}
	r.renderHorizontal(ctx, trackRect, valueRect, valueText, trackStyle, thumbStyle, fillStyle)
}

func (r *RangeSlider) renderHorizontal(ctx runtime.RenderContext, trackRect, valueRect runtime.Rect, valueText string, trackStyle, thumbStyle, fillStyle backend.Style) {
	length := trackRect.Width
	if length <= 0 {
		return
	}
	minPos := r.valueToPos(r.minValue.Get(), length)
	maxPos := r.valueToPos(r.maxValue.Get(), length)
	if minPos > maxPos {
		minPos, maxPos = maxPos, minPos
	}
	for i := 0; i < length; i++ {
		ch := sliderTrackCharH
		style := trackStyle
		if i >= minPos && i <= maxPos {
			ch = sliderFillCharH
			style = fillStyle
		}
		ctx.Buffer.Set(trackRect.X+i, trackRect.Y, ch, style)
	}
	minChar := sliderThumbCharInactive
	maxChar := sliderThumbCharInactive
	if r.active == rangeHandleMin {
		minChar = sliderThumbChar
	} else {
		maxChar = sliderThumbChar
	}
	ctx.Buffer.Set(trackRect.X+minPos, trackRect.Y, minChar, thumbStyle)
	ctx.Buffer.Set(trackRect.X+maxPos, trackRect.Y, maxChar, thumbStyle)
	if r.showValue && valueText != "" && valueRect.Width > 0 {
		ctx.Buffer.SetString(valueRect.X, valueRect.Y, truncateString(valueText, valueRect.Width), trackStyle)
	}
}

func (r *RangeSlider) renderVertical(ctx runtime.RenderContext, trackRect, valueRect runtime.Rect, valueText string, trackStyle, thumbStyle, fillStyle backend.Style) {
	length := trackRect.Height
	if length <= 0 {
		return
	}
	minPos := r.valueToPos(r.minValue.Get(), length)
	maxPos := r.valueToPos(r.maxValue.Get(), length)
	if minPos > maxPos {
		minPos, maxPos = maxPos, minPos
	}
	for i := 0; i < length; i++ {
		y := trackRect.Y + (length - 1 - i)
		ch := sliderTrackCharV
		style := trackStyle
		if i >= minPos && i <= maxPos {
			ch = sliderFillCharV
			style = fillStyle
		}
		ctx.Buffer.Set(trackRect.X, y, ch, style)
	}
	minChar := sliderThumbCharInactive
	maxChar := sliderThumbCharInactive
	if r.active == rangeHandleMin {
		minChar = sliderThumbChar
	} else {
		maxChar = sliderThumbChar
	}
	minY := trackRect.Y + (length - 1 - minPos)
	maxY := trackRect.Y + (length - 1 - maxPos)
	ctx.Buffer.Set(trackRect.X, minY, minChar, thumbStyle)
	ctx.Buffer.Set(trackRect.X, maxY, maxChar, thumbStyle)
	if r.showValue && valueText != "" && valueRect.Width > 0 {
		ctx.Buffer.SetString(valueRect.X, valueRect.Y, truncateString(valueText, valueRect.Width), trackStyle)
	}
}

// HandleMessage updates value on input.
func (r *RangeSlider) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if r == nil {
		return runtime.Unhandled()
	}
	switch m := msg.(type) {
	case runtime.KeyMsg:
		if !r.focused {
			return runtime.Unhandled()
		}
		return r.handleKey(m)
	case runtime.MouseMsg:
		return r.handleMouse(m)
	}
	return runtime.Unhandled()
}

func (r *RangeSlider) handleKey(key runtime.KeyMsg) runtime.HandleResult {
	step := r.stepSize()
	page := r.pageStep()
	switch key.Key {
	case terminal.KeyLeft, terminal.KeyDown:
		r.adjustActive(-step)
		return runtime.Handled()
	case terminal.KeyRight, terminal.KeyUp:
		r.adjustActive(step)
		return runtime.Handled()
	case terminal.KeyPageDown:
		r.adjustActive(-page)
		return runtime.Handled()
	case terminal.KeyPageUp:
		r.adjustActive(page)
		return runtime.Handled()
	case terminal.KeyHome:
		r.setActiveValue(r.min)
		return runtime.Handled()
	case terminal.KeyEnd:
		r.setActiveValue(r.max)
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func (r *RangeSlider) handleMouse(m runtime.MouseMsg) runtime.HandleResult {
	content := r.ContentBounds()
	if !content.Contains(m.X, m.Y) {
		if m.Action == runtime.MouseRelease && m.Button == runtime.MouseLeft {
			r.dragging = false
		}
		return runtime.Unhandled()
	}
	if m.Button != runtime.MouseLeft {
		return runtime.Unhandled()
	}
	valueText := ""
	if r.showValue {
		minValue, maxValue := r.Values()
		valueText = fmt.Sprintf(r.valueFormat, minValue) + " - " + fmt.Sprintf(r.valueFormat, maxValue)
	}
	trackRect, _ := sliderTrackRect(content, r.orientation, valueText)
	if trackRect.Width <= 0 || trackRect.Height <= 0 {
		return runtime.Unhandled()
	}
	switch m.Action {
	case runtime.MousePress:
		r.dragging = true
		r.setActiveFromPoint(m.X, m.Y, trackRect)
		return runtime.Handled()
	case runtime.MouseMove:
		if r.dragging {
			r.setActiveFromPoint(m.X, m.Y, trackRect)
			return runtime.Handled()
		}
	case runtime.MouseRelease:
		if r.dragging {
			r.dragging = false
			r.setActiveFromPoint(m.X, m.Y, trackRect)
			return runtime.Handled()
		}
	}
	return runtime.Unhandled()
}

func (r *RangeSlider) setActiveFromPoint(x, y int, trackRect runtime.Rect) {
	value := r.valueFromPoint(x, y, trackRect)
	minValue, maxValue := r.Values()
	if math.Abs(value-minValue) <= math.Abs(value-maxValue) {
		r.active = rangeHandleMin
	} else {
		r.active = rangeHandleMax
	}
	r.setActiveValue(value)
}

func (r *RangeSlider) valueFromPoint(x, y int, trackRect runtime.Rect) float64 {
	if r.orientation == Vertical {
		length := max(1, trackRect.Height)
		offset := trackRect.Y + trackRect.Height - 1 - y
		if offset < 0 {
			offset = 0
		}
		if offset >= length {
			offset = length - 1
		}
		return r.posToValue(offset, length)
	}
	length := max(1, trackRect.Width)
	offset := x - trackRect.X
	if offset < 0 {
		offset = 0
	}
	if offset >= length {
		offset = length - 1
	}
	return r.posToValue(offset, length)
}

func (r *RangeSlider) adjustActive(delta float64) {
	if r.active == rangeHandleMax {
		r.setMaxValue(r.maxValue.Get() + delta)
	} else {
		r.setMinValue(r.minValue.Get() + delta)
	}
	r.clampValues()
}

func (r *RangeSlider) setActiveValue(value float64) {
	if r.active == rangeHandleMax {
		r.setMaxValue(value)
	} else {
		r.setMinValue(value)
	}
	r.clampValues()
}

func (r *RangeSlider) setMinValue(value float64) {
	if r.minValue == nil {
		return
	}
	r.minValue.Set(r.clampValue(value))
}

func (r *RangeSlider) setMaxValue(value float64) {
	if r.maxValue == nil {
		return
	}
	r.maxValue.Set(r.clampValue(value))
}

func (r *RangeSlider) clampValues() {
	if r == nil {
		return
	}
	minVal := r.clampValue(r.minValue.Get())
	maxVal := r.clampValue(r.maxValue.Get())
	if minVal > maxVal {
		if r.active == rangeHandleMin {
			minVal = maxVal
		} else {
			maxVal = minVal
		}
	}
	r.minValue.Set(minVal)
	r.maxValue.Set(maxVal)
	r.syncA11y()
	r.services.Invalidate()
}

func (r *RangeSlider) valueToPos(value float64, length int) int {
	if length <= 1 {
		return 0
	}
	if r.max <= r.min {
		return 0
	}
	ratio := (value - r.min) / (r.max - r.min)
	pos := int(math.Round(ratio * float64(length-1)))
	if pos < 0 {
		return 0
	}
	if pos >= length {
		return length - 1
	}
	return pos
}

func (r *RangeSlider) posToValue(pos, length int) float64 {
	if length <= 1 {
		return r.min
	}
	ratio := float64(pos) / float64(length-1)
	value := r.min + ratio*(r.max-r.min)
	return r.clampValue(value)
}

func (r *RangeSlider) clampValue(value float64) float64 {
	if r == nil {
		return value
	}
	if r.max < r.min {
		r.max, r.min = r.min, r.max
	}
	if r.step > 0 {
		steps := math.Round((value - r.min) / r.step)
		value = r.min + steps*r.step
	}
	if value < r.min {
		value = r.min
	}
	if value > r.max {
		value = r.max
	}
	return value
}

func (r *RangeSlider) stepSize() float64 {
	if r.step > 0 {
		return r.step
	}
	return (r.max - r.min) / 100
}

func (r *RangeSlider) pageStep() float64 {
	step := r.stepSize()
	if step <= 0 {
		return 1
	}
	return step * 10
}

func (r *RangeSlider) syncA11y() {
	if r == nil {
		return
	}
	if r.Base.Role == "" {
		r.Base.Role = accessibility.RoleSlider
	}
	label := strings.TrimSpace(r.label)
	if label == "" {
		label = "Range Slider"
	}
	r.Base.Label = label
	minVal, maxVal := r.Values()
	r.Base.Value = &accessibility.ValueInfo{Text: fmt.Sprintf("%s - %s", fmt.Sprintf(r.valueFormat, minVal), fmt.Sprintf(r.valueFormat, maxVal))}
}

var _ runtime.Widget = (*RangeSlider)(nil)
var _ runtime.Focusable = (*RangeSlider)(nil)
var _ runtime.Bindable = (*RangeSlider)(nil)
var _ runtime.Unbindable = (*RangeSlider)(nil)
