package widgets

import (
	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/state"
	uistyle "github.com/odvcencio/fluffy-ui/style"
	"github.com/odvcencio/fluffy-ui/terminal"
)

// RadioGroup manages a set of radio buttons.
type RadioGroup struct {
	selected *state.Signal[int]
	options  []*Radio
	onChange func(index int)
}

// NewRadioGroup creates an empty group.
func NewRadioGroup() *RadioGroup {
	return &RadioGroup{selected: state.NewSignal(-1)}
}

// Selected returns the selected index.
func (g *RadioGroup) Selected() int {
	if g == nil || g.selected == nil {
		return 0
	}
	return g.selected.Get()
}

// SetSelected updates the selected index.
func (g *RadioGroup) SetSelected(index int) {
	if g == nil || g.selected == nil {
		return
	}
	g.selected.Set(index)
	if g.onChange != nil {
		g.onChange(index)
	}
}

// OnChange registers a selection callback.
func (g *RadioGroup) OnChange(fn func(index int)) {
	if g == nil {
		return
	}
	g.onChange = fn
}

// Radio is a single radio option.
type Radio struct {
	FocusableBase

	label      *state.Signal[string]
	group      *RadioGroup
	index      int
	disabled   bool
	style      backend.Style
	focusStyle backend.Style
	styleSet   bool
	focusSet   bool
}

// NewRadio creates a radio option and registers it with the group.
func NewRadio(label string, group *RadioGroup) *Radio {
	r := &Radio{
		label:      state.NewSignal(label),
		group:      group,
		style:      backend.DefaultStyle(),
		focusStyle: backend.DefaultStyle().Reverse(true),
	}
	r.Base.Role = accessibility.RoleRadio
	r.Base.Label = label
	if group != nil {
		r.index = len(group.options)
		group.options = append(group.options, r)
	}
	r.syncState()
	return r
}

// SetLabel updates the radio label.
func (r *Radio) SetLabel(label string) {
	if r == nil || r.label == nil {
		return
	}
	r.label.Set(label)
	r.Base.Label = label
}

// SetDisabled updates disabled state.
func (r *Radio) SetDisabled(disabled bool) {
	if r == nil {
		return
	}
	r.disabled = disabled
	r.Base.State.Disabled = disabled
}

// SetStyle sets the normal style.
func (r *Radio) SetStyle(style backend.Style) {
	if r == nil {
		return
	}
	r.style = style
	r.styleSet = true
}

// SetFocusStyle sets the focused style.
func (r *Radio) SetFocusStyle(style backend.Style) {
	if r == nil {
		return
	}
	r.focusStyle = style
	r.focusSet = true
}

// StyleType returns the selector type name.
func (r *Radio) StyleType() string {
	return "Radio"
}

// Measure returns the desired size.
func (r *Radio) Measure(constraints runtime.Constraints) runtime.Size {
	return r.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		label := ""
		if r.label != nil {
			label = r.label.Get()
		}
		width := 4 + len(label)
		if width < 4 {
			width = 4
		}
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: 1})
	})
}

// Render draws the radio.
func (r *Radio) Render(ctx runtime.RenderContext) {
	if r == nil {
		return
	}
	r.syncState()
	outer := r.bounds
	content := r.ContentBounds()
	if outer.Width <= 0 || outer.Height <= 0 {
		return
	}
	selected := r.isSelected()
	marker := "( )"
	if selected {
		marker = "(*)"
	}
	label := ""
	if r.label != nil {
		label = r.label.Get()
	}
	available := max(0, content.Width-4)
	text := marker + " " + truncateString(label, available)
	style := r.style
	resolved := ctx.ResolveStyle(r)
		if !resolved.IsZero() {
			final := resolved
			if r.styleSet {
				final = final.Merge(uistyle.FromBackend(r.style))
			}
			if r.focused && r.focusSet {
				final = final.Merge(uistyle.FromBackend(r.focusStyle))
			}
			style = final.ToBackend()
	} else {
		if r.focused {
			style = r.focusStyle
		}
		if r.disabled {
			style = style.Dim(true)
		}
	}
	ctx.Buffer.Fill(outer, ' ', style)
	if content.Width <= 0 || content.Height <= 0 {
		return
	}
	writePadded(ctx.Buffer, content.X, content.Y, content.Width, text, style)
}

// HandleMessage handles selection.
func (r *Radio) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if r == nil || !r.focused || r.disabled {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	if key.Key == terminal.KeyEnter || (key.Key == terminal.KeyRune && key.Rune == ' ') {
		if r.group != nil {
			r.group.SetSelected(r.index)
			r.syncState()
			return runtime.Handled()
		}
	}
	return runtime.Unhandled()
}

func (r *Radio) isSelected() bool {
	if r == nil || r.group == nil {
		return false
	}
	return r.group.Selected() == r.index
}

func (r *Radio) syncState() {
	if r == nil {
		return
	}
	selected := r.isSelected()
	r.Base.State.Selected = selected
	if r.label != nil {
		r.Base.Label = r.label.Get()
	}
	if r.Base.Role == "" {
		r.Base.Role = accessibility.RoleRadio
	}
}
