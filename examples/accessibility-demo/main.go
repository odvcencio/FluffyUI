package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/examples/internal/demo"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/widgets"
)

func main() {
	announcer := &accessibility.SimpleAnnouncer{}
	view := NewAccessibilityView(announcer)

	bundle, err := demo.NewApp(view, demo.Options{Announcer: announcer})
	if err != nil {
		fmt.Fprintf(os.Stderr, "app init failed: %v\n", err)
		os.Exit(1)
	}

	if err := bundle.App.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
		os.Exit(1)
	}
}

type AccessibilityView struct {
	widgets.Component
	controls  *accessibilityControls
	log       *widgets.Text
	splitter  *widgets.Splitter
	messages  []string
	announcer *accessibility.SimpleAnnouncer
}

func NewAccessibilityView(announcer *accessibility.SimpleAnnouncer) *AccessibilityView {
	view := &AccessibilityView{announcer: announcer}
	view.controls = newAccessibilityControls(announcer, view)
	view.log = widgets.NewText("Announcements will appear here.")
	view.splitter = widgets.NewSplitter(view.controls, view.log)
	view.splitter.Ratio = 0.5

	if announcer != nil {
		announcer.SetOnMessage(func(msg accessibility.Announcement) {
			view.appendLog(msg.Message)
		})
	}

	return view
}

func (a *AccessibilityView) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (a *AccessibilityView) Layout(bounds runtime.Rect) {
	a.Component.Layout(bounds)
	if a.splitter != nil {
		a.splitter.Layout(bounds)
	}
}

func (a *AccessibilityView) Render(ctx runtime.RenderContext) {
	if a.splitter != nil {
		a.splitter.Render(ctx)
	}
}

func (a *AccessibilityView) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if a.splitter != nil {
		return a.splitter.HandleMessage(msg)
	}
	return runtime.Unhandled()
}

func (a *AccessibilityView) ChildWidgets() []runtime.Widget {
	if a.splitter == nil {
		return nil
	}
	return []runtime.Widget{a.splitter}
}

func (a *AccessibilityView) appendLog(message string) {
	if strings.TrimSpace(message) == "" {
		return
	}
	max := 8
	a.messages = append(a.messages, message)
	if len(a.messages) > max {
		a.messages = a.messages[len(a.messages)-max:]
	}
	if a.log != nil {
		a.log.SetText(strings.Join(a.messages, "\n"))
	}
	a.Invalidate()
}

type accessibilityControls struct {
	widgets.Base
	title     *widgets.Label
	hint      *widgets.Label
	textarea  *widgets.TextArea
	checkbox  *widgets.Checkbox
	announce  *widgets.Button
	announcer *accessibility.SimpleAnnouncer
	parent    *AccessibilityView
}

func newAccessibilityControls(announcer *accessibility.SimpleAnnouncer, parent *AccessibilityView) *accessibilityControls {
	c := &accessibilityControls{
		announcer: announcer,
		parent:    parent,
	}
	c.title = widgets.NewLabel("Accessibility Demo", widgets.WithLabelStyle(backend.DefaultStyle().Bold(true)))
	c.hint = widgets.NewLabel("Tab to move focus. Enter to activate.")
	c.textarea = widgets.NewTextArea()
	c.textarea.SetLabel("Notes")
	c.textarea.SetText("Type here to change the accessibility value.")
	c.checkbox = widgets.NewCheckbox("Enable alerts")
	c.checkbox.SetOnChange(func(value *bool) {
		if c.announcer == nil {
			return
		}
		state := "disabled"
		if value != nil && *value {
			state = "enabled"
		}
		c.announcer.Announce("Alerts "+state, accessibility.PriorityPolite)
	})
	c.announce = widgets.NewButton("Announce status", widgets.WithOnClick(func() {
		if c.announcer != nil {
			c.announcer.Announce("Manual announcement", accessibility.PriorityAssertive)
		}
	}))
	return c
}

func (c *accessibilityControls) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (c *accessibilityControls) Layout(bounds runtime.Rect) {
	c.Base.Layout(bounds)
	y := bounds.Y
	line := func(w runtime.Widget, height int) {
		if w == nil {
			return
		}
		w.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: height})
		y += height
	}
	line(c.title, 1)
	line(c.hint, 1)
	line(c.textarea, 5)
	line(c.checkbox, 1)
	line(c.announce, 1)
}

func (c *accessibilityControls) Render(ctx runtime.RenderContext) {
	if c.title != nil {
		c.title.Render(ctx)
	}
	if c.hint != nil {
		c.hint.Render(ctx)
	}
	if c.textarea != nil {
		c.textarea.Render(ctx)
	}
	if c.checkbox != nil {
		c.checkbox.Render(ctx)
	}
	if c.announce != nil {
		c.announce.Render(ctx)
	}
}

func (c *accessibilityControls) HandleMessage(msg runtime.Message) runtime.HandleResult {
	children := c.ChildWidgets()
	for _, child := range children {
		if child == nil {
			continue
		}
		if result := child.HandleMessage(msg); result.Handled {
			return result
		}
	}
	return runtime.Unhandled()
}

func (c *accessibilityControls) ChildWidgets() []runtime.Widget {
	children := []runtime.Widget{}
	if c.title != nil {
		children = append(children, c.title)
	}
	if c.hint != nil {
		children = append(children, c.hint)
	}
	if c.textarea != nil {
		children = append(children, c.textarea)
	}
	if c.checkbox != nil {
		children = append(children, c.checkbox)
	}
	if c.announce != nil {
		children = append(children, c.announce)
	}
	return children
}
