package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/examples/internal/demo"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/widgets"
)

func main() {
	view := NewCalendarView()
	bundle, err := demo.NewApp(view, demo.Options{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "app init failed: %v\n", err)
		os.Exit(1)
	}

	if err := bundle.App.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
		os.Exit(1)
	}
}

type CalendarView struct {
	widgets.Component
	header     *widgets.Label
	calendar   *widgets.Calendar
	datePicker *widgets.DatePicker
	status     *widgets.Label
}

func NewCalendarView() *CalendarView {
	view := &CalendarView{}
	view.header = widgets.NewLabel("Calendar + DatePicker").WithStyle(backend.DefaultStyle().Bold(true))
	view.calendar = widgets.NewCalendar(
		widgets.WithWeekStart(time.Monday),
		widgets.WithShowWeekNumbers(true),
	)
	view.datePicker = widgets.NewDatePicker()
	view.status = widgets.NewLabel("Select a date")

	view.calendar.OnSelect(func(date time.Time) {
		view.status.SetText("Selected: " + date.Format("2006-01-02"))
		view.Invalidate()
	})

	view.datePicker.Calendar().OnSelect(func(date time.Time) {
		view.status.SetText("Picked: " + date.Format("2006-01-02"))
		view.Invalidate()
	})

	highlight := []time.Time{
		time.Now(),
		time.Now().AddDate(0, 0, 3),
	}
	view.calendar.SetHighlightDates(highlight)

	return view
}

func (c *CalendarView) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (c *CalendarView) Layout(bounds runtime.Rect) {
	c.Component.Layout(bounds)
	y := bounds.Y

	line := func(w runtime.Widget, height int) {
		if w == nil {
			return
		}
		w.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: height})
		y += height
	}

	line(c.header, 1)
	line(c.calendar, 8)
	line(c.datePicker, 10)
	line(c.status, 1)
}

func (c *CalendarView) Render(ctx runtime.RenderContext) {
	for _, child := range c.ChildWidgets() {
		if child != nil {
			child.Render(ctx)
		}
	}
}

func (c *CalendarView) HandleMessage(msg runtime.Message) runtime.HandleResult {
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

func (c *CalendarView) ChildWidgets() []runtime.Widget {
	children := []runtime.Widget{}
	if c.header != nil {
		children = append(children, c.header)
	}
	if c.calendar != nil {
		children = append(children, c.calendar)
	}
	if c.datePicker != nil {
		children = append(children, c.datePicker)
	}
	if c.status != nil {
		children = append(children, c.status)
	}
	return children
}
