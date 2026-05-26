package widgets

import (
	"fmt"
	"strings"

	"m31labs.dev/fluffyui/accessibility"
	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/terminal"
)

const (
	ratingFilledStar = "\u2605" // ★
	ratingEmptyStar  = "\u2606" // ☆
)

// Rating is a focusable star rating widget.
// It renders filled and empty stars based on the current value.
type Rating struct {
	FocusableBase

	value    int
	max      int
	onChange func(int)
	services runtime.Services

	style    backend.Style
	styleSet bool
}

// NewRating creates a rating widget with the given max stars.
func NewRating(maxStars int) *Rating {
	if maxStars < 1 {
		maxStars = 5
	}
	r := &Rating{
		max:   maxStars,
		style: backend.DefaultStyle(),
	}
	r.Base.Role = accessibility.RoleSlider
	r.Base.Label = "Rating"
	r.syncA11y()
	return r
}

// SetOnChange sets the value change handler.
func (r *Rating) SetOnChange(fn func(int)) {
	if r == nil {
		return
	}
	r.onChange = fn
}

// SetValue sets the current rating value.
func (r *Rating) SetValue(value int) {
	if r == nil {
		return
	}
	if value < 0 {
		value = 0
	}
	if value > r.max {
		value = r.max
	}
	old := r.value
	r.value = value
	r.syncA11y()
	if old != value {
		if r.onChange != nil {
			r.onChange(value)
		}
		if announcer := r.services.Announcer(); announcer != nil {
			announcer.Announce(fmt.Sprintf("Rating %d of %d", value, r.max), accessibility.PriorityPolite)
		}
	}
}

// Value returns the current rating value.
func (r *Rating) Value() int {
	if r == nil {
		return 0
	}
	return r.value
}

// Max returns the maximum rating value.
func (r *Rating) Max() int {
	if r == nil {
		return 0
	}
	return r.max
}

// SetStyle sets the widget style.
func (r *Rating) SetStyle(style backend.Style) {
	if r == nil {
		return
	}
	r.style = style
	r.styleSet = true
}

// StyleType returns the selector type name.
func (r *Rating) StyleType() string {
	return "Rating"
}

// Bind attaches app services.
func (r *Rating) Bind(services runtime.Services) {
	if r == nil {
		return
	}
	r.services = services
}

// Unbind releases app services.
func (r *Rating) Unbind() {
	if r == nil {
		return
	}
	r.services = runtime.Services{}
}

// Measure returns the desired size.
func (r *Rating) Measure(constraints runtime.Constraints) runtime.Size {
	return r.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		// Each star is one character wide.
		width := r.max
		if width < 1 {
			width = 1
		}
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: 1})
	})
}

// Render draws the rating widget.
func (r *Rating) Render(ctx runtime.RenderContext) {
	if r == nil {
		return
	}
	r.syncA11y()
	bounds := r.ContentBounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	style := resolveBaseStyle(ctx, r, r.style, r.styleSet)
	text := r.renderText()
	writePadded(ctx.Buffer, bounds.X, bounds.Y, bounds.Width, truncateString(text, bounds.Width), style)
}

func (r *Rating) renderText() string {
	var sb strings.Builder
	for i := 1; i <= r.max; i++ {
		if i <= r.value {
			sb.WriteString(ratingFilledStar)
		} else {
			sb.WriteString(ratingEmptyStar)
		}
	}
	return sb.String()
}

// HandleMessage processes keyboard input.
func (r *Rating) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if r == nil || !r.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	switch key.Key {
	case terminal.KeyRight, terminal.KeyUp:
		if r.value < r.max {
			r.SetValue(r.value + 1)
			return runtime.Handled()
		}
	case terminal.KeyLeft, terminal.KeyDown:
		if r.value > 0 {
			r.SetValue(r.value - 1)
			return runtime.Handled()
		}
	case terminal.KeyHome:
		r.SetValue(0)
		return runtime.Handled()
	case terminal.KeyEnd:
		r.SetValue(r.max)
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func (r *Rating) syncA11y() {
	if r == nil {
		return
	}
	if r.Base.Role == "" {
		r.Base.Role = accessibility.RoleSlider
	}
	r.Base.Label = "Rating"
	r.Base.Value = &accessibility.ValueInfo{
		Min:     0,
		Max:     float64(r.max),
		Current: float64(r.value),
		Text:    fmt.Sprintf("%d of %d stars", r.value, r.max),
	}
	r.Base.Orientation = "horizontal"
}

var _ runtime.Widget = (*Rating)(nil)
var _ runtime.Focusable = (*Rating)(nil)
var _ runtime.Bindable = (*Rating)(nil)
var _ runtime.Unbindable = (*Rating)(nil)
