package main

import (
	"time"
	"unicode"

	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/terminal"
	"github.com/odvcencio/fluffy-ui/widgets"
)

// ModalDialog is a centered dialog with title, content area, and action bar.
type ModalDialog struct {
	widgets.Component
	title       string
	width       int
	height      int
	content     runtime.Widget
	actions     []DialogAction
	selectedAct int
	onDismiss   func()
	dismissable bool

	// Auto-dismiss
	autoDismissDuration time.Duration
	startTime           time.Time
	paused              bool

	style       backend.Style
	titleStyle  backend.Style
	actionStyle backend.Style
}

// DialogAction represents a button in the dialog's action bar.
type DialogAction struct {
	Label    string
	Key      rune
	OnSelect func()
}

// NewModalDialog creates a new modal dialog.
func NewModalDialog(title string, width, height int) *ModalDialog {
	return &ModalDialog{
		title:       title,
		width:       width,
		height:      height,
		dismissable: true,
		style:       backend.DefaultStyle(),
		titleStyle:  backend.DefaultStyle().Bold(true).Reverse(true),
		actionStyle: backend.DefaultStyle().Dim(true),
	}
}

// WithContent sets the dialog's content widget.
func (d *ModalDialog) WithContent(content runtime.Widget) *ModalDialog {
	d.content = content
	return d
}

// WithActions sets the dialog's action buttons.
func (d *ModalDialog) WithActions(actions ...DialogAction) *ModalDialog {
	d.actions = actions
	return d
}

// WithDismissable sets whether Escape closes the dialog.
func (d *ModalDialog) WithDismissable(dismissable bool) *ModalDialog {
	d.dismissable = dismissable
	return d
}

// OnDismiss sets the callback when dialog is dismissed.
func (d *ModalDialog) OnDismiss(fn func()) *ModalDialog {
	d.onDismiss = fn
	return d
}

// WithAutoDismiss enables auto-dismiss after duration.
func (d *ModalDialog) WithAutoDismiss(duration time.Duration) *ModalDialog {
	d.autoDismissDuration = duration
	d.startTime = time.Now()
	return d
}

// TimerProgress returns 0.0-1.0 progress toward auto-dismiss.
func (d *ModalDialog) TimerProgress(now time.Time) float64 {
	if d.autoDismissDuration <= 0 {
		return 0
	}
	elapsed := now.Sub(d.startTime)
	progress := float64(elapsed) / float64(d.autoDismissDuration)
	if progress > 1.0 {
		return 1.0
	}
	if progress < 0 {
		return 0
	}
	return progress
}

// ShouldDismiss returns true if auto-dismiss time has elapsed.
func (d *ModalDialog) ShouldDismiss(now time.Time) bool {
	if d.autoDismissDuration <= 0 || d.paused {
		return false
	}
	return now.Sub(d.startTime) >= d.autoDismissDuration
}

// PauseTimer pauses the auto-dismiss timer.
func (d *ModalDialog) PauseTimer() { d.paused = true }

// ResumeTimer resumes the auto-dismiss timer.
func (d *ModalDialog) ResumeTimer() { d.paused = false }

// IsPaused returns whether timer is paused.
func (d *ModalDialog) IsPaused() bool { return d.paused }

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
	d.Component.Layout(bounds)
	if d.content != nil {
		contentBounds := runtime.Rect{
			X:      bounds.X + 2,
			Y:      bounds.Y + 2,
			Width:  bounds.Width - 4,
			Height: bounds.Height - 4,
		}
		if len(d.actions) > 0 {
			contentBounds.Height -= 1
		}
		if d.autoDismissDuration > 0 {
			contentBounds.Height -= 1 // Room for timer bar
		}
		d.content.Layout(contentBounds)
	}
}

// Render draws the dialog.
func (d *ModalDialog) Render(ctx runtime.RenderContext) {
	bounds := d.Bounds()

	// Fill background and draw border
	ctx.Buffer.Fill(bounds, ' ', d.style)
	ctx.Buffer.DrawBox(bounds, d.style)

	// Draw title
	title := " " + d.title + " "
	if len(title) > bounds.Width-4 {
		title = title[:bounds.Width-4]
	}
	ctx.Buffer.SetString(bounds.X+2, bounds.Y, title, d.titleStyle)

	// Render content
	if d.content != nil {
		d.content.Render(ctx)
	}

	// Render auto-dismiss timer bar
	if d.autoDismissDuration > 0 {
		timerY := bounds.Y + bounds.Height - 3
		if len(d.actions) > 0 {
			timerY = bounds.Y + bounds.Height - 4
		}
		progress := 1.0 - d.TimerProgress(time.Now())
		barWidth := bounds.Width - 4
		filledWidth := int(float64(barWidth) * progress)

		for i := 0; i < barWidth; i++ {
			ch := '░'
			if i < filledWidth {
				ch = '█'
			}
			ctx.Buffer.Set(bounds.X+2+i, timerY, ch, d.style)
		}
	}

	// Render action bar
	if len(d.actions) > 0 {
		actionY := bounds.Y + bounds.Height - 2
		actionLine := ""
		for i, action := range d.actions {
			if i > 0 {
				actionLine += "  "
			}
			actionLine += "[" + string(action.Key) + "] " + action.Label
		}
		ctx.Buffer.SetString(bounds.X+2, actionY, actionLine, d.actionStyle)
	}
}

// HandleMessage handles keyboard input.
func (d *ModalDialog) HandleMessage(msg runtime.Message) runtime.HandleResult {
	// Handle mouse for pause
	if mouse, ok := msg.(runtime.MouseMsg); ok {
		bounds := d.Bounds()
		if bounds.Contains(mouse.X, mouse.Y) {
			d.PauseTimer()
		} else {
			d.ResumeTimer()
		}
	}

	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		if d.content != nil {
			return d.content.HandleMessage(msg)
		}
		return runtime.Unhandled()
	}

	// Handle Escape
	if key.Key == terminal.KeyEscape && d.dismissable {
		if d.onDismiss != nil {
			d.onDismiss()
		}
		return runtime.Handled()
	}

	// Handle action keys
	for _, action := range d.actions {
		if key.Rune == action.Key || key.Rune == unicode.ToLower(action.Key) || key.Rune == unicode.ToUpper(action.Key) {
			if action.OnSelect != nil {
				action.OnSelect()
			}
			return runtime.Handled()
		}
	}

	// Pass to content
	if d.content != nil {
		return d.content.HandleMessage(msg)
	}

	return runtime.Handled()
}

// ChildWidgets returns the content widget.
func (d *ModalDialog) ChildWidgets() []runtime.Widget {
	if d.content == nil {
		return nil
	}
	return []runtime.Widget{d.content}
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
	modal.dismissable = len(choices) == 0
	return modal
}

// Render draws the event modal with choices.
func (e *EventModal) Render(ctx runtime.RenderContext) {
	bounds := e.Bounds()

	// Fill and border
	ctx.Buffer.Fill(bounds, ' ', e.style)
	ctx.Buffer.DrawBox(bounds, e.style)

	// Title
	title := " " + e.title + " "
	ctx.Buffer.SetString(bounds.X+2, bounds.Y, title, e.titleStyle)

	// Message
	lines := splitLines(e.message, bounds.Width-4)
	maxMsgLines := bounds.Height - 4 - len(e.choices)
	for i, line := range lines {
		if i < maxMsgLines {
			ctx.Buffer.SetString(bounds.X+2, bounds.Y+2+i, line, e.style)
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
			style := e.actionStyle
			if i == e.selectedChoice {
				style = style.Reverse(true)
			}
			ctx.Buffer.SetString(bounds.X+2, choiceY+i, line, style)
		}
	}
}

// HandleMessage handles choice selection.
func (e *EventModal) HandleMessage(msg runtime.Message) runtime.HandleResult {
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Handled()
	}

	// No choices = any key dismisses
	if len(e.choices) == 0 {
		if e.onDismiss != nil {
			e.onDismiss()
		}
		return runtime.Handled()
	}

	// Check for choice keys
	for _, choice := range e.choices {
		if key.Rune == choice.Key || key.Rune == unicode.ToUpper(choice.Key) || key.Rune == unicode.ToLower(choice.Key) {
			if choice.OnSelect != nil {
				choice.OnSelect()
			}
			if e.onDismiss != nil {
				e.onDismiss()
			}
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
			if e.onDismiss != nil {
				e.onDismiss()
			}
		}
	}

	return runtime.Handled()
}
