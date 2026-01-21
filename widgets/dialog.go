package widgets

import (
	"strings"

	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/terminal"
)

// DialogButton represents an action in a dialog.
type DialogButton struct {
	Label   string
	OnClick func()
}

// Dialog is a modal message container.
type Dialog struct {
	FocusableBase
	Title    string
	Body     string
	Buttons  []DialogButton
	selected int
	style    backend.Style
}

// NewDialog creates a dialog.
func NewDialog(title, body string, buttons ...DialogButton) *Dialog {
	dialog := &Dialog{
		Title:   title,
		Body:    body,
		Buttons: buttons,
		style:   backend.DefaultStyle(),
	}
	dialog.Base.Role = accessibility.RoleDialog
	dialog.Base.Label = title
	dialog.Base.Description = body
	return dialog
}

// Measure returns desired size.
func (d *Dialog) Measure(constraints runtime.Constraints) runtime.Size {
	width := len(d.Title)
	for _, line := range strings.Split(d.Body, "\n") {
		if len(line) > width {
			width = len(line)
		}
	}
	if width < 10 {
		width = 10
	}
	height := 3 + len(strings.Split(d.Body, "\n"))
	if len(d.Buttons) > 0 {
		height++
	}
	return constraints.Constrain(runtime.Size{Width: width + 4, Height: height + 2})
}

// Render draws the dialog.
func (d *Dialog) Render(ctx runtime.RenderContext) {
	if d == nil {
		return
	}
	d.syncA11y()
	bounds := d.bounds
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	ctx.Buffer.DrawBox(bounds, d.style)
	inner := bounds.Inset(1, 1, 1, 1)
	if inner.Width <= 0 || inner.Height <= 0 {
		return
	}
	title := truncateString(d.Title, inner.Width)
	ctx.Buffer.SetString(inner.X, inner.Y, title, d.style.Bold(true))
	bodyLines := strings.Split(d.Body, "\n")
	for i, line := range bodyLines {
		y := inner.Y + 1 + i
		if y >= inner.Y+inner.Height {
			break
		}
		line = truncateString(line, inner.Width)
		ctx.Buffer.SetString(inner.X, y, line, d.style)
	}
	if len(d.Buttons) == 0 {
		return
	}
	buttonY := inner.Y + inner.Height - 1
	x := inner.X
	for i, button := range d.Buttons {
		label := "[" + button.Label + "]"
		if x+len(label) > inner.X+inner.Width {
			break
		}
		style := d.style
		if i == d.selected {
			style = style.Reverse(true)
		}
		ctx.Buffer.SetString(x, buttonY, label, style)
		x += len(label) + 1
	}
}

// HandleMessage handles button selection.
func (d *Dialog) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if d == nil || !d.focused || len(d.Buttons) == 0 {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	switch key.Key {
	case terminal.KeyLeft:
		d.setSelected(d.selected - 1)
		return runtime.Handled()
	case terminal.KeyRight:
		d.setSelected(d.selected + 1)
		return runtime.Handled()
	case terminal.KeyEnter:
		if d.selected >= 0 && d.selected < len(d.Buttons) {
			if d.Buttons[d.selected].OnClick != nil {
				d.Buttons[d.selected].OnClick()
			}
		}
		return runtime.Handled()
	case terminal.KeyEscape:
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func (d *Dialog) setSelected(index int) {
	if len(d.Buttons) == 0 {
		d.selected = 0
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= len(d.Buttons) {
		index = len(d.Buttons) - 1
	}
	d.selected = index
}

func (d *Dialog) syncA11y() {
	if d == nil {
		return
	}
	if d.Base.Role == "" {
		d.Base.Role = accessibility.RoleDialog
	}
	d.Base.Label = d.Title
	d.Base.Description = d.Body
	if d.selected >= 0 && d.selected < len(d.Buttons) {
		d.Base.Value = &accessibility.ValueInfo{Text: d.Buttons[d.selected].Label}
	} else {
		d.Base.Value = nil
	}
}
