package widgets

import (
	"strings"
	"time"
	"unicode"

	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/terminal"
)

// DialogButton represents an action in a dialog.
type DialogButton struct {
	Label   string
	Key     rune // Keyboard shortcut (e.g., 'Y' for Yes). 0 = no shortcut.
	OnClick func()
}

// Dialog is a modal message container with optional custom content,
// keyboard shortcuts, auto-dismiss timer, and dismiss callbacks.
type Dialog struct {
	FocusableBase
	Title   string
	Body    string         // Text body (used if Content is nil)
	Content runtime.Widget // Custom content widget (takes precedence over Body)
	Buttons []DialogButton

	selected    int
	dismissable bool   // Whether Escape closes the dialog (default true)
	onDismiss   func() // Callback when dismissed via Escape

	// Auto-dismiss timer
	autoDismiss time.Duration // 0 = disabled
	startTime   time.Time
	paused      bool

	style backend.Style
	styleSet bool
}

// NewDialog creates a dialog with title, body text, and optional buttons.
// Use builder methods to add custom content, auto-dismiss, etc.
func NewDialog(title, body string, buttons ...DialogButton) *Dialog {
	dialog := &Dialog{
		Title:       title,
		Body:        body,
		Buttons:     buttons,
		dismissable: true,
		style:       backend.DefaultStyle(),
	}
	dialog.Base.Role = accessibility.RoleDialog
	dialog.Base.Label = title
	dialog.Base.Description = body
	return dialog
}

// StyleType returns the selector type name.
func (d *Dialog) StyleType() string {
	return "Dialog"
}

// SetStyle updates the dialog style.
func (d *Dialog) SetStyle(style backend.Style) {
	if d == nil {
		return
	}
	d.style = style
	d.styleSet = true
}

// WithContent sets a custom widget as dialog body (replaces text Body).
func (d *Dialog) WithContent(content runtime.Widget) *Dialog {
	d.Content = content
	return d
}

// WithDismissable sets whether Escape closes the dialog (default true).
func (d *Dialog) WithDismissable(dismissable bool) *Dialog {
	d.dismissable = dismissable
	return d
}

// OnDismiss sets callback when dialog is dismissed via Escape.
func (d *Dialog) OnDismiss(fn func()) *Dialog {
	d.onDismiss = fn
	return d
}

// WithAutoDismiss enables auto-dismiss after duration (0 = disabled).
// Call ShouldDismiss() periodically to check if time has elapsed.
func (d *Dialog) WithAutoDismiss(duration time.Duration) *Dialog {
	d.autoDismiss = duration
	d.startTime = time.Now()
	return d
}

// CenteredBounds returns bounds to center dialog within parent rect.
func (d *Dialog) CenteredBounds(parent runtime.Rect) runtime.Rect {
	size := d.Measure(runtime.Constraints{
		MaxWidth:  parent.Width,
		MaxHeight: parent.Height,
	})
	x := parent.X + (parent.Width-size.Width)/2
	y := parent.Y + (parent.Height-size.Height)/2
	return runtime.Rect{X: x, Y: y, Width: size.Width, Height: size.Height}
}

// PauseTimer pauses the auto-dismiss timer.
func (d *Dialog) PauseTimer() { d.paused = true }

// ResumeTimer resumes the auto-dismiss timer.
func (d *Dialog) ResumeTimer() { d.paused = false }

// IsPaused returns whether the auto-dismiss timer is paused.
func (d *Dialog) IsPaused() bool { return d.paused }

// TimerProgress returns 0.0-1.0 progress toward auto-dismiss.
func (d *Dialog) TimerProgress(now time.Time) float64 {
	if d.autoDismiss <= 0 {
		return 0
	}
	elapsed := now.Sub(d.startTime)
	progress := float64(elapsed) / float64(d.autoDismiss)
	if progress > 1.0 {
		return 1.0
	}
	if progress < 0 {
		return 0
	}
	return progress
}

// ShouldDismiss returns true if auto-dismiss time has elapsed.
func (d *Dialog) ShouldDismiss(now time.Time) bool {
	if d.autoDismiss <= 0 || d.paused {
		return false
	}
	return now.Sub(d.startTime) >= d.autoDismiss
}

// Measure returns desired size.
func (d *Dialog) Measure(constraints runtime.Constraints) runtime.Size {
	return d.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		width := len(d.Title)

		// Measure body text width
		if d.Content == nil {
			for _, line := range strings.Split(d.Body, "\n") {
				if len(line) > width {
					width = len(line)
				}
			}
		} else {
			// For custom content, use a reasonable default or measure it
			contentSize := d.Content.Measure(contentConstraints)
			if contentSize.Width > width {
				width = contentSize.Width
			}
		}

		if width < 10 {
			width = 10
		}

		// Calculate height
		height := 3 // title + padding
		if d.Content == nil {
			height += len(strings.Split(d.Body, "\n"))
		} else {
			contentSize := d.Content.Measure(contentConstraints)
			height += contentSize.Height
		}
		if len(d.Buttons) > 0 {
			height++
		}
		if d.autoDismiss > 0 {
			height++ // timer bar
		}

		return contentConstraints.Constrain(runtime.Size{Width: width + 4, Height: height + 2})
	})
}

// Layout positions the dialog and its content.
func (d *Dialog) Layout(bounds runtime.Rect) {
	d.FocusableBase.Layout(bounds)

	if d.Content != nil {
		inner := d.ContentBounds().Inset(1, 1, 1, 1)
		contentBounds := runtime.Rect{
			X:      inner.X,
			Y:      inner.Y + 1, // below title
			Width:  inner.Width,
			Height: inner.Height - 1,
		}
		if len(d.Buttons) > 0 {
			contentBounds.Height--
		}
		if d.autoDismiss > 0 {
			contentBounds.Height--
		}
		d.Content.Layout(contentBounds)
	}
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

	baseStyle := resolveBaseStyle(ctx, d, d.style, d.styleSet)

	// Fill background and draw border
	ctx.Buffer.Fill(bounds, ' ', baseStyle)
	ctx.Buffer.DrawBox(bounds, baseStyle)

	inner := d.ContentBounds().Inset(1, 1, 1, 1)
	if inner.Width <= 0 || inner.Height <= 0 {
		return
	}

	// Title
	title := truncateString(d.Title, inner.Width)
	ctx.Buffer.SetString(inner.X, inner.Y, title, baseStyle.Bold(true))

	// Calculate content area
	contentEndY := inner.Y + inner.Height
	if len(d.Buttons) > 0 {
		contentEndY--
	}
	if d.autoDismiss > 0 {
		contentEndY--
	}

	// Render body: custom Content or text Body
	if d.Content != nil {
		d.Content.Render(ctx)
	} else {
		bodyLines := strings.Split(d.Body, "\n")
		for i, line := range bodyLines {
			y := inner.Y + 1 + i
			if y >= contentEndY {
				break
			}
			line = truncateString(line, inner.Width)
			ctx.Buffer.SetString(inner.X, y, line, baseStyle)
		}
	}

	// Timer bar (if auto-dismiss enabled)
	if d.autoDismiss > 0 {
		timerY := contentEndY
		progress := 1.0 - d.TimerProgress(time.Now())
		barWidth := inner.Width
		filled := int(float64(barWidth) * progress)

		for i := 0; i < barWidth; i++ {
			ch := '░'
			if i < filled {
				ch = '█'
			}
			ctx.Buffer.Set(inner.X+i, timerY, ch, baseStyle.Dim(true))
		}
	}

	// Buttons
	if len(d.Buttons) == 0 {
		return
	}
	buttonY := inner.Y + inner.Height - 1
	x := inner.X
	for i, button := range d.Buttons {
		var label string
		if button.Key != 0 {
			label = "[" + string(unicode.ToUpper(button.Key)) + "] " + button.Label
		} else {
			label = "[" + button.Label + "]"
		}
		if x+len(label) > inner.X+inner.Width {
			break
		}
		style := baseStyle
		if i == d.selected {
			style = style.Reverse(true)
		}
		ctx.Buffer.SetString(x, buttonY, label, style)
		x += len(label) + 2
	}
}

// HandleMessage handles button selection and keyboard shortcuts.
func (d *Dialog) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if d == nil || !d.focused {
		return runtime.Unhandled()
	}

	// Mouse: pause/resume timer when hovering
	if mouse, ok := msg.(runtime.MouseMsg); ok && d.autoDismiss > 0 {
		if d.bounds.Contains(mouse.X, mouse.Y) {
			d.PauseTimer()
		} else {
			d.ResumeTimer()
		}
	}

	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		// Pass non-key events to content
		if d.Content != nil {
			return d.Content.HandleMessage(msg)
		}
		return runtime.Unhandled()
	}

	// Escape handling
	if key.Key == terminal.KeyEscape {
		if d.dismissable && d.onDismiss != nil {
			d.onDismiss()
		}
		return runtime.Handled()
	}

	// Check keyboard shortcuts (case-insensitive)
	for _, btn := range d.Buttons {
		if btn.Key != 0 && (key.Rune == btn.Key ||
			key.Rune == unicode.ToLower(btn.Key) ||
			key.Rune == unicode.ToUpper(btn.Key)) {
			if btn.OnClick != nil {
				btn.OnClick()
			}
			return runtime.Handled()
		}
	}

	// Arrow key navigation
	if len(d.Buttons) > 0 {
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
		}
	}

	// Pass to content if set
	if d.Content != nil {
		return d.Content.HandleMessage(msg)
	}

	return runtime.Handled()
}

// ChildWidgets returns the content widget for proper widget tree traversal.
func (d *Dialog) ChildWidgets() []runtime.Widget {
	if d.Content == nil {
		return nil
	}
	return []runtime.Widget{d.Content}
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
	if d.Content == nil {
		d.Base.Description = d.Body
	}
	if d.selected >= 0 && d.selected < len(d.Buttons) {
		d.Base.Value = &accessibility.ValueInfo{Text: d.Buttons[d.selected].Label}
	} else {
		d.Base.Value = nil
	}
}
