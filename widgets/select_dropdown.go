package widgets

import (
	"strings"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

type selectDropdown struct {
	FocusableBase

	options       []SelectOption
	selected      int
	offset        int
	label         string
	style         backend.Style
	selectedStyle backend.Style
	disabledStyle backend.Style
	styleSet      bool
	selectedSet   bool
	disabledSet   bool
	onSelect      func(index int)
	onClose       func()
}

func newSelectDropdown(parent *Select) *selectDropdown {
	drop := &selectDropdown{
		options:       parent.options,
		selected:      parent.selected,
		label:         parent.label,
		style:         parent.style,
		selectedStyle: parent.focusStyle,
		disabledStyle: backend.DefaultStyle().Dim(true),
		styleSet:      parent.styleSet,
		selectedSet:   parent.focusSet,
	}
	drop.Base.Role = accessibility.RoleList
	drop.Base.Label = strings.TrimSpace(parent.label)
	drop.Base.Description = "Select options"
	drop.ensureSelectable()
	drop.onSelect = func(index int) {
		parent.SetSelected(index)
	}
	drop.onClose = func() {
		parent.dropdownOpen = false
	}
	return drop
}

// Measure returns the size needed for the dropdown list.
func (d *selectDropdown) Measure(constraints runtime.Constraints) runtime.Size {
	return d.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		width := 0
		for _, option := range d.options {
			w := textWidth(option.Label) + 2
			if w > width {
				width = w
			}
		}
		if width < 4 {
			width = 4
		}
		height := len(d.options)
		if height < 1 {
			height = 1
		}
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: height})
	})
}

// Render draws the dropdown list.
func (d *selectDropdown) Render(ctx runtime.RenderContext) {
	if d == nil {
		return
	}
	d.syncA11y()
	outer := d.bounds
	content := d.ContentBounds()
	if outer.Width <= 0 || outer.Height <= 0 {
		return
	}

	baseStyle := resolveBaseStyle(ctx, d, d.style, d.styleSet)
	ctx.Buffer.Fill(outer, ' ', baseStyle)
	if content.Width <= 0 || content.Height <= 0 {
		return
	}

	d.ensureSelectable()
	d.ensureVisible(content.Height)

	selectedStyle := d.selectedStyle
	if d.selectedSet {
		selectedStyle = mergeBackendStyles(baseStyle, d.selectedStyle)
	} else {
		selectedStyle = baseStyle.Reverse(true)
	}
	if !d.disabledSet && d.disabledStyle == (backend.Style{}) {
		d.disabledStyle = baseStyle.Dim(true)
	}

	for i := 0; i < content.Height; i++ {
		index := d.offset + i
		if index < 0 || index >= len(d.options) {
			break
		}
		option := d.options[index]
		line := " " + option.Label
		if textWidth(line) > content.Width {
			line = truncateString(line, content.Width)
		}
		style := baseStyle
		if index == d.selected {
			style = selectedStyle
		}
		if option.Disabled {
			if d.disabledSet {
				style = mergeBackendStyles(style, d.disabledStyle)
			} else {
				style = style.Dim(true)
			}
		}
		writePadded(ctx.Buffer, content.X, content.Y+i, content.Width, line, style)
	}
}

// HandleMessage processes navigation and selection.
func (d *selectDropdown) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if d == nil || !d.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if ok {
		switch key.Key {
		case terminal.KeyUp:
			if d.moveSelection(-1) {
				return runtime.Handled()
			}
			return runtime.Handled()
		case terminal.KeyDown:
			if d.moveSelection(1) {
				return runtime.Handled()
			}
			return runtime.Handled()
		case terminal.KeyEnter:
			if d.selected >= 0 && d.selected < len(d.options) && !d.options[d.selected].Disabled {
				if d.onSelect != nil {
					d.onSelect(d.selected)
				}
			}
			d.close()
			return runtime.WithCommand(runtime.PopOverlay{})
		case terminal.KeyEscape:
			d.close()
			return runtime.WithCommand(runtime.PopOverlay{})
		}
	}

	if mouse, ok := msg.(runtime.MouseMsg); ok {
		if mouse.Action == runtime.MousePress && mouse.Button == runtime.MouseLeft {
			content := d.ContentBounds()
			if content.Contains(mouse.X, mouse.Y) {
				index := d.offset + (mouse.Y - content.Y)
				if index >= 0 && index < len(d.options) {
					if !d.options[index].Disabled {
						d.selected = index
						if d.onSelect != nil {
							d.onSelect(index)
						}
					}
					d.close()
					return runtime.WithCommand(runtime.PopOverlay{})
				}
				return runtime.Handled()
			}
		}
	}

	return runtime.Unhandled()
}

func (d *selectDropdown) moveSelection(delta int) bool {
	if len(d.options) == 0 {
		return false
	}
	index := d.selected
	if index < 0 || index >= len(d.options) {
		index = 0
	}
	for i := 0; i < len(d.options); i++ {
		index += delta
		if index < 0 {
			index = len(d.options) - 1
		} else if index >= len(d.options) {
			index = 0
		}
		if !d.options[index].Disabled {
			d.selected = index
			d.syncA11y()
			return true
		}
	}
	return false
}

func (d *selectDropdown) ensureSelectable() {
	if len(d.options) == 0 {
		d.selected = -1
		return
	}
	if d.selected < 0 || d.selected >= len(d.options) {
		d.selected = 0
	}
	if d.options[d.selected].Disabled {
		if !d.moveSelection(1) {
			d.selected = -1
		}
	}
}

func (d *selectDropdown) ensureVisible(height int) {
	if d.selected < 0 {
		d.offset = 0
		return
	}
	if d.selected < d.offset {
		d.offset = d.selected
	} else if d.selected >= d.offset+height {
		d.offset = d.selected - height + 1
	}
	if d.offset < 0 {
		d.offset = 0
	}
}

func (d *selectDropdown) syncA11y() {
	if d == nil {
		return
	}
	if d.Base.Role == "" {
		d.Base.Role = accessibility.RoleList
	}
	label := strings.TrimSpace(d.label)
	if label == "" {
		label = "Select"
	}
	d.Base.Label = label
	if d.selected >= 0 && d.selected < len(d.options) {
		d.Base.Value = &accessibility.ValueInfo{Text: d.options[d.selected].Label}
	} else {
		d.Base.Value = nil
	}
}

func (d *selectDropdown) close() {
	if d.onClose != nil {
		d.onClose()
	}
}

// StyleType returns the selector type name.
func (d *selectDropdown) StyleType() string {
	return "Select"
}

var _ runtime.Widget = (*selectDropdown)(nil)
var _ runtime.Focusable = (*selectDropdown)(nil)
