package main

import (
	"strings"
	"time"
	"unicode"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
	"github.com/odvcencio/fluffyui/widgets"
)

// ModalDialog wraps widgets.Dialog with fixed dimensions and game-specific styling.
// This demonstrates how to extend the library dialog for application-specific needs.
type ModalDialog struct {
	*widgets.Dialog
	width       int
	height      int
	titleStyle  backend.Style
	actionStyle backend.Style
}

// DialogAction represents a button in the dialog's action bar.
// Maps to widgets.DialogButton but with game-specific naming.
type DialogAction struct {
	Label    string
	Key      rune
	OnSelect func()
}

// NewModalDialog creates a new modal dialog with fixed dimensions.
func NewModalDialog(title string, width, height int) *ModalDialog {
	dialog := widgets.NewDialog(title, "")
	return &ModalDialog{
		Dialog:      dialog,
		width:       width,
		height:      height,
		titleStyle:  backend.DefaultStyle().Bold(true).Reverse(true),
		actionStyle: backend.DefaultStyle().Dim(true),
	}
}

// WithContent sets the dialog's content widget.
func (d *ModalDialog) WithContent(content runtime.Widget) *ModalDialog {
	d.Dialog.WithContent(content)
	return d
}

// WithActions sets the dialog's action buttons.
func (d *ModalDialog) WithActions(actions ...DialogAction) *ModalDialog {
	buttons := make([]widgets.DialogButton, len(actions))
	for i, a := range actions {
		buttons[i] = widgets.DialogButton{
			Label:   a.Label,
			Key:     a.Key,
			OnClick: a.OnSelect,
		}
	}
	d.Dialog.Buttons = buttons
	return d
}

// WithDismissable sets whether Escape closes the dialog.
func (d *ModalDialog) WithDismissable(dismissable bool) *ModalDialog {
	d.Dialog.WithDismissable(dismissable)
	return d
}

// OnDismiss sets the callback when dialog is dismissed.
func (d *ModalDialog) OnDismiss(fn func()) *ModalDialog {
	d.Dialog.OnDismiss(fn)
	return d
}

// WithAutoDismiss enables auto-dismiss after duration.
func (d *ModalDialog) WithAutoDismiss(duration time.Duration) *ModalDialog {
	d.Dialog.WithAutoDismiss(duration)
	return d
}

// TimerProgress returns 0.0-1.0 progress toward auto-dismiss.
func (d *ModalDialog) TimerProgress(now time.Time) float64 {
	return d.Dialog.TimerProgress(now)
}

// ShouldDismiss returns true if auto-dismiss time has elapsed.
func (d *ModalDialog) ShouldDismiss(now time.Time) bool {
	return d.Dialog.ShouldDismiss(now)
}

// PauseTimer pauses the auto-dismiss timer.
func (d *ModalDialog) PauseTimer() { d.Dialog.PauseTimer() }

// ResumeTimer resumes the auto-dismiss timer.
func (d *ModalDialog) ResumeTimer() { d.Dialog.ResumeTimer() }

// IsPaused returns whether timer is paused.
func (d *ModalDialog) IsPaused() bool { return d.Dialog.IsPaused() }

// Measure returns the dialog's fixed size.
func (d *ModalDialog) Measure(constraints runtime.Constraints) runtime.Size {
	return runtime.Size{Width: d.width, Height: d.height}
}

// CenteredBounds returns bounds centered in parent.
func (d *ModalDialog) CenteredBounds(parent runtime.Rect) runtime.Rect {
	x := parent.X + (parent.Width-d.width)/2
	y := parent.Y + (parent.Height-d.height)/2
	return runtime.Rect{X: x, Y: y, Width: d.width, Height: d.height}
}

// Layout positions the dialog and its content.
func (d *ModalDialog) Layout(bounds runtime.Rect) {
	d.Dialog.Layout(bounds)
}

// Render draws the dialog with game-specific styling.
func (d *ModalDialog) Render(ctx runtime.RenderContext) {
	bounds := d.Dialog.Bounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}

	style := backend.DefaultStyle()

	// Fill background and draw border
	ctx.Buffer.Fill(bounds, ' ', style)
	ctx.Buffer.DrawBox(bounds, style)

	// Draw title with reverse style
	title := " " + d.Dialog.Title + " "
	if len(title) > bounds.Width-4 {
		title = title[:bounds.Width-4]
	}
	ctx.Buffer.SetString(bounds.X+2, bounds.Y, title, d.titleStyle)

	inner := bounds.Inset(1, 1, 1, 1)
	if inner.Width <= 0 || inner.Height <= 0 {
		return
	}

	// Render content
	if d.Dialog.Content != nil {
		d.Dialog.Content.Render(ctx)
	}

	// Render auto-dismiss timer bar
	if d.Dialog.ShouldDismiss(time.Now().Add(time.Hour)) || d.Dialog.TimerProgress(time.Now()) > 0 {
		timerY := bounds.Y + bounds.Height - 3
		if len(d.Dialog.Buttons) > 0 {
			timerY = bounds.Y + bounds.Height - 4
		}
		progress := 1.0 - d.Dialog.TimerProgress(time.Now())
		barWidth := bounds.Width - 4
		filledWidth := int(float64(barWidth) * progress)

		for i := 0; i < barWidth; i++ {
			ch := '░'
			if i < filledWidth {
				ch = '█'
			}
			ctx.Buffer.Set(bounds.X+2+i, timerY, ch, style)
		}
	}

	// Render action bar with game-specific styling
	if len(d.Dialog.Buttons) > 0 {
		actionY := bounds.Y + bounds.Height - 2
		actionLine := ""
		for i, button := range d.Dialog.Buttons {
			if i > 0 {
				actionLine += "  "
			}
			if button.Key != 0 {
				actionLine += "[" + string(button.Key) + "] " + button.Label
			} else {
				actionLine += "[" + button.Label + "]"
			}
		}
		ctx.Buffer.SetString(bounds.X+2, actionY, actionLine, d.actionStyle)
	}
}

// HandleMessage delegates to the underlying Dialog.
func (d *ModalDialog) HandleMessage(msg runtime.Message) runtime.HandleResult {
	d.Dialog.Focus()
	return d.Dialog.HandleMessage(msg)
}

// ChildWidgets returns the content widget.
func (d *ModalDialog) ChildWidgets() []runtime.Widget {
	return d.Dialog.ChildWidgets()
}

// EventModal is a specialized dialog for game events with choices.
type EventModal struct {
	*ModalDialog
	message        string
	choices        []EventChoice
	selectedChoice int
}

// EventChoice represents an option in an event.
type EventChoice struct {
	Key      rune
	Label    string
	OnSelect func()
}

// NewEventModal creates an event dialog.
func NewEventModal(title, message string, choices ...EventChoice) *EventModal {
	lines := splitLines(message, 46)
	height := 6 + len(lines)
	if len(choices) > 0 {
		height += len(choices)
	}
	if height > 16 {
		height = 16
	}

	modal := &EventModal{
		ModalDialog: NewModalDialog(title, 50, height),
		message:     message,
		choices:     choices,
	}
	modal.WithDismissable(len(choices) == 0)
	modal.syncA11y()
	return modal
}

func (e *EventModal) syncA11y() {
	if e == nil || e.Dialog == nil {
		return
	}
	if e.Dialog.Base.Role == "" {
		e.Dialog.Base.Role = accessibility.RoleDialog
	}
	e.Dialog.Base.Label = e.Dialog.Title
	e.Dialog.Base.Description = e.message
	if e.selectedChoice >= 0 && e.selectedChoice < len(e.choices) {
		e.Dialog.Base.Value = &accessibility.ValueInfo{Text: e.choices[e.selectedChoice].Label}
	} else {
		e.Dialog.Base.Value = nil
	}
}

// Render draws the event modal with choices.
func (e *EventModal) Render(ctx runtime.RenderContext) {
	e.syncA11y()
	bounds := e.Dialog.Bounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}

	style := backend.DefaultStyle()

	// Fill and border
	ctx.Buffer.Fill(bounds, ' ', style)
	ctx.Buffer.DrawBox(bounds, style)

	// Title
	title := " " + e.Dialog.Title + " "
	ctx.Buffer.SetString(bounds.X+2, bounds.Y, title, e.titleStyle)

	// Message
	lines := splitLines(e.message, bounds.Width-4)
	maxMsgLines := bounds.Height - 4 - len(e.choices)
	for i, line := range lines {
		if i < maxMsgLines {
			ctx.Buffer.SetString(bounds.X+2, bounds.Y+2+i, line, style)
		}
	}

	// Choices or dismiss hint
	if len(e.choices) == 0 {
		hint := "[Press any key to continue]"
		ctx.Buffer.SetString(bounds.X+2, bounds.Y+bounds.Height-2, hint, e.actionStyle)
	} else {
		choiceY := bounds.Y + bounds.Height - 2 - len(e.choices)
		for i, choice := range e.choices {
			line := "[" + string(choice.Key) + "] " + choice.Label
			choiceStyle := e.actionStyle
			if i == e.selectedChoice {
				choiceStyle = choiceStyle.Reverse(true)
			}
			ctx.Buffer.SetString(bounds.X+2, choiceY+i, line, choiceStyle)
		}
	}
}

// ChildWidgets returns accessibility children describing the message and choices.
func (e *EventModal) ChildWidgets() []runtime.Widget {
	if e == nil {
		return nil
	}
	e.syncA11y()
	lines := splitLines(e.message, 46)
	children := make([]runtime.Widget, 0, len(lines)+len(e.choices)+1)
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		children = append(children, a11yText(line))
	}
	if len(e.choices) == 0 {
		children = append(children, a11yText("[Press any key to continue]"))
		return children
	}
	items := make([]runtime.Widget, 0, len(e.choices))
	for i, choice := range e.choices {
		label := "[" + string(choice.Key) + "] " + choice.Label
		items = append(items, a11yListItem(label, i == e.selectedChoice, false))
	}
	children = append(children, a11yList("Choices", items))
	return children
}

// HandleMessage handles choice selection.
func (e *EventModal) HandleMessage(msg runtime.Message) runtime.HandleResult {
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Handled()
	}

	// No choices = any key dismisses
	if len(e.choices) == 0 {
		e.Dialog.Focus()
		e.Dialog.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})
		return runtime.Handled()
	}

	// Check for choice keys
	for _, choice := range e.choices {
		if key.Rune == choice.Key || key.Rune == unicode.ToUpper(choice.Key) || key.Rune == unicode.ToLower(choice.Key) {
			if choice.OnSelect != nil {
				choice.OnSelect()
			}
			e.Dialog.Focus()
			e.Dialog.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})
			return runtime.Handled()
		}
	}

	// Arrow key navigation
	switch key.Key {
	case terminal.KeyUp:
		if e.selectedChoice > 0 {
			e.selectedChoice--
		}
	case terminal.KeyDown:
		if e.selectedChoice < len(e.choices)-1 {
			e.selectedChoice++
		}
	case terminal.KeyEnter:
		if e.selectedChoice >= 0 && e.selectedChoice < len(e.choices) {
			choice := e.choices[e.selectedChoice]
			if choice.OnSelect != nil {
				choice.OnSelect()
			}
			e.Dialog.Focus()
			e.Dialog.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})
		}
	}

	e.syncA11y()
	return runtime.Handled()
}
