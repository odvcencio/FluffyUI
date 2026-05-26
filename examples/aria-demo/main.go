package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"m31labs.dev/fluffyui/accessibility"
	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/examples/internal/demo"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/terminal"
	"m31labs.dev/fluffyui/toast"
	"m31labs.dev/fluffyui/widgets"
)

const mcpSocket = "/tmp/fluffyui.sock"

func main() {
	announcer := &accessibility.SimpleAnnouncer{}
	view := NewAriaView(announcer)

	bundle, err := demo.NewApp(view, demo.Options{
		Announcer: announcer,
		TickRate:  500 * time.Millisecond,
		MCP:       mcpSocket,
		TTS:       true,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "app init failed: %v\n", err)
		os.Exit(1)
	}

	if err := bundle.App.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
		os.Exit(1)
	}
}

// ---------------------------------------------------------------------------
// AriaView — top-level split: controls left, announcement log right
// ---------------------------------------------------------------------------

type AriaView struct {
	widgets.Component
	controls  *ariaControls
	log       *widgets.Text
	splitter  *widgets.Splitter
	messages  []string
	announcer *accessibility.SimpleAnnouncer
}

func NewAriaView(announcer *accessibility.SimpleAnnouncer) *AriaView {
	view := &AriaView{announcer: announcer}
	view.controls = newAriaControls(announcer, view)
	view.log = widgets.NewText("Waiting for announcements...")
	view.log.SetLabel("Announcement Log")
	view.splitter = widgets.NewSplitter(view.controls, view.log)
	view.splitter.Ratio = 0.55

	if announcer != nil {
		announcer.SetOnMessage(func(msg accessibility.Announcement) {
			view.appendLog(msg.Message)
		})
	}

	return view
}

func (a *AriaView) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (a *AriaView) Layout(bounds runtime.Rect) {
	a.Component.Layout(bounds)
	if a.splitter != nil {
		a.splitter.Layout(bounds)
	}
}

func (a *AriaView) Render(ctx runtime.RenderContext) {
	if a.splitter != nil {
		a.splitter.Render(ctx)
	}
}

func (a *AriaView) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if a.splitter != nil {
		return a.splitter.HandleMessage(msg)
	}
	return runtime.Unhandled()
}

func (a *AriaView) ChildWidgets() []runtime.Widget {
	if a.splitter == nil {
		return nil
	}
	return []runtime.Widget{a.splitter}
}

func (a *AriaView) appendLog(message string) {
	if strings.TrimSpace(message) == "" {
		return
	}
	const max = 10
	a.messages = append(a.messages, message)
	if len(a.messages) > max {
		a.messages = a.messages[len(a.messages)-max:]
	}
	if a.log != nil {
		a.log.SetText(strings.Join(a.messages, "\n"))
	}
	a.Invalidate()
}

// ---------------------------------------------------------------------------
// ariaControls — left panel with ARIA-annotated widgets
// ---------------------------------------------------------------------------

type ariaControls struct {
	widgets.Base

	header   *widgets.Label
	hint     *widgets.Label
	tabs     *widgets.Tabs
	alert    *widgets.Alert
	progress *widgets.Progress
	search   *widgets.SearchWidget
	btnAlert *widgets.Button
	btnToast *widgets.Button
	logView  *widgets.Log

	// New ARIA compliance demos
	section    *widgets.Section
	toggle     *widgets.Button
	emailInput *widgets.Input
	errorLabel *widgets.Label
	busyBtn    *widgets.Button

	toastMgr   *toast.ToastManager
	toastStack *widgets.ToastStack

	announcer    *accessibility.SimpleAnnouncer
	parent       *AriaView
	alertError   bool
	toggleState  bool
	lastTick     time.Time
}

func newAriaControls(announcer *accessibility.SimpleAnnouncer, parent *AriaView) *ariaControls {
	c := &ariaControls{
		announcer: announcer,
		parent:    parent,
	}

	c.header = widgets.NewLabel("FluffyUI ARIA Demo",
		widgets.WithLabelStyle(backend.DefaultStyle().Bold(true)))

	c.hint = widgets.NewLabel("Tab=next widget  Left/Right=switch tabs  D=dialog  T=toast")

	// Tabs — landmark: navigation (set by default)
	dashContent := widgets.NewLabel("Dashboard: system overview")
	settingsContent := widgets.NewLabel("Settings: preferences panel")
	c.tabs = widgets.NewTabs(
		widgets.Tab{Title: "Dashboard", Content: dashContent},
		widgets.Tab{Title: "Settings", Content: settingsContent},
	)

	// Alert — live: assertive, atomic: true (set by default)
	c.alert = widgets.NewAlert("All systems nominal", widgets.AlertSuccess)

	// Progress — relevant: text, atomic: true (set by default)
	c.progress = widgets.NewProgress()
	c.progress.Value = 0
	c.progress.SetLabel("Build Progress")

	// Search — landmark: search (set by default)
	c.search = widgets.NewSearchWidget()
	c.search.SetLabel("Search widgets")

	// Toggle Alert button — updates alert in-place
	c.btnAlert = widgets.NewButton("Toggle Alert", widgets.WithOnClick(func() {
		c.alertError = !c.alertError
		if c.alertError {
			c.alert.Text = "Network connection lost!"
			c.alert.Variant = widgets.AlertError
		} else {
			c.alert.Text = "All systems nominal"
			c.alert.Variant = widgets.AlertSuccess
		}
		c.parent.Invalidate()
	}))

	// Toast
	c.toastMgr = toast.NewToastManager()
	c.toastStack = widgets.NewToastStack()
	c.toastMgr.SetOnChange(c.toastStack.SetToasts)

	c.btnToast = widgets.NewButton("Show Toast", widgets.WithOnClick(func() {
		c.toastMgr.Success("Deployed", "Build v1.2.3 shipped successfully")
	}))

	// Section — demonstrates Expanded state
	c.section = widgets.NewSection("Relationship Demo")
	c.section.SetExpanded(true)

	// Toggle button — uses Pressed state
	c.toggle = widgets.NewButton("Mute Notifications", widgets.WithOnClick(func() {
		c.toggleState = !c.toggleState
		if c.toggleState {
			c.toggle.SetLabel("Mute Notifications")
			c.toggle.Base.State.Pressed = accessibility.BoolPtr(true)
		} else {
			c.toggle.SetLabel("Mute Notifications")
			c.toggle.Base.State.Pressed = accessibility.BoolPtr(false)
		}
		c.parent.Invalidate()
	}))
	c.toggle.Base.State.Pressed = accessibility.BoolPtr(false)

	// Form validation — Required + Invalid states with DescribedBy
	c.emailInput = widgets.NewInput()
	c.emailInput.SetLabel("Email")
	c.emailInput.SetPlaceholder("user@example.com")
	c.emailInput.Base.State.Required = true
	c.emailInput.Base.State.Invalid = true
	c.emailInput.Base.SetDescribedBy("email-error")

	c.errorLabel = widgets.NewLabel("Please enter a valid email address")
	c.errorLabel.Base.Live = accessibility.LivePolite

	// Busy state demo
	c.busyBtn = widgets.NewButton("Start Build", widgets.WithOnClick(func() {
		c.busyBtn.Base.State.Busy = true
		c.busyBtn.SetLabel("Building...")
		c.parent.Invalidate()
		go func() {
			time.Sleep(2 * time.Second)
			c.busyBtn.Base.State.Busy = false
			c.busyBtn.SetLabel("Start Build")
			c.parent.Invalidate()
		}()
	}))

	// Relationship wiring: button controls the section
	c.btnAlert.Base.SetControls("relationship-section")

	// Log — live: polite, relevant: additions (set by default)
	c.logView = widgets.NewLog(
		widgets.WithMaxLines(50),
		widgets.WithAutoScroll(true),
		widgets.WithShowTime(false),
	)
	c.logView.SetLabel("Activity Log")
	c.logView.Info("ARIA demo started")

	return c
}

func (c *ariaControls) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (c *ariaControls) Layout(bounds runtime.Rect) {
	c.Base.Layout(bounds)
	y := bounds.Y

	line := func(w runtime.Widget, height int) {
		if w == nil {
			return
		}
		w.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: height})
		y += height
	}

	line(c.header, 1)
	line(c.hint, 1)
	line(c.tabs, 3)
	line(c.alert, 1)
	line(c.progress, 1)
	line(c.search, 1)
	line(c.btnAlert, 1)
	line(c.btnToast, 1)
	line(c.section, 1)
	line(c.toggle, 1)
	line(c.emailInput, 1)
	line(c.errorLabel, 1)
	line(c.busyBtn, 1)

	used := y - bounds.Y
	remaining := bounds.Height - used
	if remaining < 1 {
		remaining = 1
	}
	line(c.logView, remaining)

	// Toast stack overlays the log area.
	if c.toastStack != nil {
		c.toastStack.Layout(runtime.Rect{
			X: bounds.X, Y: bounds.Y + used,
			Width: bounds.Width, Height: remaining,
		})
	}
}

func (c *ariaControls) Render(ctx runtime.RenderContext) {
	for _, child := range c.ChildWidgets() {
		if child != nil {
			child.Render(ctx)
		}
	}
}

func (c *ariaControls) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if key, ok := msg.(runtime.KeyMsg); ok {
		switch key.Rune {
		case 'd', 'D':
			return runtime.WithCommand(runtime.PushOverlay{Widget: newDialogOverlay(c.announcer), Modal: true})
		case 't', 'T':
			if c.toastMgr != nil {
				c.toastMgr.Info("Heads up", "New deployment queued")
			}
			return runtime.Handled()
		}
	}

	if tick, ok := msg.(runtime.TickMsg); ok {
		c.onTick(tick.Time)
	}

	for _, child := range c.ChildWidgets() {
		if child == nil {
			continue
		}
		if result := child.HandleMessage(msg); result.Handled {
			return result
		}
	}
	return runtime.Unhandled()
}

func (c *ariaControls) ChildWidgets() []runtime.Widget {
	var children []runtime.Widget
	for _, w := range []runtime.Widget{
		c.header, c.hint, c.tabs, c.alert, c.progress, c.search,
		c.btnAlert, c.btnToast, c.section, c.toggle, c.emailInput,
		c.errorLabel, c.busyBtn, c.logView, c.toastStack,
	} {
		if w != nil {
			children = append(children, w)
		}
	}
	return children
}

func (c *ariaControls) onTick(now time.Time) {
	if !c.lastTick.IsZero() && now.Sub(c.lastTick) < 500*time.Millisecond {
		return
	}
	c.lastTick = now

	c.progress.Value += 3
	if c.progress.Value > c.progress.Max {
		c.progress.Value = 0
		if c.logView != nil {
			c.logView.Info("Build cycle restarted")
		}
	}

	c.parent.Invalidate()
}

// ---------------------------------------------------------------------------
// dialogOverlay
// ---------------------------------------------------------------------------

type dialogOverlay struct {
	dialog    *widgets.Dialog
	announcer *accessibility.SimpleAnnouncer
}

func newDialogOverlay(announcer *accessibility.SimpleAnnouncer) *dialogOverlay {
	d := widgets.NewDialog("Confirm Deployment", "Ship build v1.2.3 to production?",
		widgets.DialogButton{Label: "Deploy"},
		widgets.DialogButton{Label: "Cancel"},
	)
	d.Focus()
	return &dialogOverlay{dialog: d, announcer: announcer}
}

func (d *dialogOverlay) Measure(constraints runtime.Constraints) runtime.Size {
	if d.dialog == nil {
		return constraints.MinSize()
	}
	return d.dialog.Measure(constraints)
}

func (d *dialogOverlay) Layout(bounds runtime.Rect) {
	if d.dialog == nil {
		return
	}
	size := d.dialog.Measure(runtime.Constraints{MaxWidth: bounds.Width, MaxHeight: bounds.Height})
	x := bounds.X + (bounds.Width-size.Width)/2
	y := bounds.Y + (bounds.Height-size.Height)/2
	d.dialog.Layout(runtime.Rect{X: x, Y: y, Width: size.Width, Height: size.Height})
}

func (d *dialogOverlay) Render(ctx runtime.RenderContext) {
	if d.dialog != nil {
		d.dialog.Render(ctx)
	}
}

func (d *dialogOverlay) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if key, ok := msg.(runtime.KeyMsg); ok {
		if key.Key == terminal.KeyEscape || key.Key == terminal.KeyEnter {
			if d.dialog != nil {
				_ = d.dialog.HandleMessage(msg)
			}
			if d.announcer != nil {
				d.announcer.Announce("Dialog dismissed", accessibility.PriorityAssertive)
			}
			return runtime.WithCommand(runtime.PopOverlay{})
		}
	}
	if d.dialog != nil {
		return d.dialog.HandleMessage(msg)
	}
	return runtime.Unhandled()
}

func (d *dialogOverlay) ChildWidgets() []runtime.Widget {
	if d.dialog == nil {
		return nil
	}
	return []runtime.Widget{d.dialog}
}
