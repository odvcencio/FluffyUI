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
	view := NewInputView()
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

type InputView struct {
	widgets.Component
	header    *widgets.Label
	input     *widgets.Input
	textarea  *widgets.TextArea
	autoComp  *widgets.AutoComplete
	multiSel  *widgets.MultiSelect
	dateRange *widgets.DateRangePicker
	timePick  *widgets.TimePicker
	checkbox  *widgets.Checkbox
	selecter  *widgets.Select
	radioFast *widgets.Radio
	radioSlow *widgets.Radio
	status    *widgets.Label
	buttons   *demo.HBox
}

func NewInputView() *InputView {
	view := &InputView{}
	view.header = widgets.NewLabel("Input Widgets").WithStyle(backend.DefaultStyle().Bold(true))
	view.input = widgets.NewInput()
	view.input.SetPlaceholder("Type and press Enter")
	view.textarea = widgets.NewTextArea()
	view.textarea.SetLabel("Notes")
	view.textarea.SetText("Multi-line input\nwith scrolling")
	view.autoComp = widgets.NewAutoComplete()
	view.autoComp.SetOptions([]string{"Alpha", "Beta", "Gamma", "Delta"})
	view.multiSel = widgets.NewMultiSelect(
		widgets.MultiSelectOption{Label: "One"},
		widgets.MultiSelectOption{Label: "Two"},
		widgets.MultiSelectOption{Label: "Three"},
	)
	view.dateRange = widgets.NewDateRangePicker()
	view.timePick = widgets.NewTimePicker()
	view.timePick.SetShowSeconds(true)
	view.checkbox = widgets.NewCheckbox("Enable feature")
	view.selecter = widgets.NewSelect(
		widgets.SelectOption{Label: "Low"},
		widgets.SelectOption{Label: "Medium"},
		widgets.SelectOption{Label: "High"},
	)
	group := widgets.NewRadioGroup()
	view.radioFast = widgets.NewRadio("Fast", group)
	view.radioSlow = widgets.NewRadio("Slow", group)
	view.status = widgets.NewLabel("Ready")

	primary := widgets.NewButton("Save", widgets.WithVariant(widgets.VariantPrimary), widgets.WithOnClick(func() {
		view.status.SetText("Saved")
		view.Invalidate()
	}))
	secondary := widgets.NewButton("Reset", widgets.WithVariant(widgets.VariantSecondary), widgets.WithOnClick(func() {
		view.status.SetText("Reset")
		view.Invalidate()
	}))
	view.buttons = demo.NewHBox(primary, secondary)
	view.buttons.Gap = 2

	view.input.OnSubmit(func(text string) {
		view.status.SetText("Submitted: " + truncateText(text, 20))
		view.Invalidate()
	})

	view.checkbox.SetOnChange(func(value *bool) {
		state := "off"
		if value != nil && *value {
			state = "on"
		}
		view.status.SetText("Checkbox: " + state)
		view.Invalidate()
	})

	view.autoComp.SetOnSelect(func(value string) {
		view.status.SetText("AutoComplete: " + value)
		view.Invalidate()
	})

	view.multiSel.SetOnChange(func(selected []widgets.MultiSelectOption) {
		view.status.SetText(fmt.Sprintf("MultiSelect: %d selected", len(selected)))
		view.Invalidate()
	})

	view.dateRange.OnRangeSelect(func(start, end time.Time) {
		view.status.SetText(fmt.Sprintf("Range: %s - %s", start.Format("Jan 2"), end.Format("Jan 2")))
		view.Invalidate()
	})

	view.timePick.SetOnChange(func(value time.Time) {
		view.status.SetText("Time: " + value.Format("15:04:05"))
		view.Invalidate()
	})

	group.OnChange(func(index int) {
		label := "-"
		if index == 0 {
			label = "Fast"
		} else if index == 1 {
			label = "Slow"
		}
		view.status.SetText("Radio: " + label)
		view.Invalidate()
	})

	return view
}

func (i *InputView) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (i *InputView) Layout(bounds runtime.Rect) {
	i.Component.Layout(bounds)
	y := bounds.Y

	line := func(w runtime.Widget, height int) {
		if w == nil {
			return
		}
		w.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: height})
		y += height
	}
	measure := func(w runtime.Widget) int {
		if w == nil {
			return 0
		}
		remaining := bounds.Height - (y - bounds.Y)
		if remaining < 1 {
			remaining = 1
		}
		size := w.Measure(runtime.Constraints{MinWidth: bounds.Width, MaxWidth: bounds.Width, MinHeight: 0, MaxHeight: remaining})
		if size.Height < 1 {
			size.Height = 1
		}
		return size.Height
	}

	line(i.header, 1)
	line(i.input, 1)
	line(i.textarea, 4)
	line(i.autoComp, measure(i.autoComp))
	line(i.multiSel, measure(i.multiSel))
	line(i.timePick, measure(i.timePick))
	line(i.dateRange, measure(i.dateRange))
	line(i.checkbox, 1)
	line(i.selecter, 1)
	line(i.radioFast, 1)
	line(i.radioSlow, 1)
	if i.buttons != nil {
		i.buttons.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: 1})
		y++
	}
	line(i.status, 1)
}

func (i *InputView) Render(ctx runtime.RenderContext) {
	for _, child := range i.ChildWidgets() {
		if child != nil {
			child.Render(ctx)
		}
	}
}

func (i *InputView) HandleMessage(msg runtime.Message) runtime.HandleResult {
	for _, child := range i.ChildWidgets() {
		if child == nil {
			continue
		}
		if result := child.HandleMessage(msg); result.Handled {
			return result
		}
	}
	return runtime.Unhandled()
}

func (i *InputView) ChildWidgets() []runtime.Widget {
	children := []runtime.Widget{}
	if i.header != nil {
		children = append(children, i.header)
	}
	if i.input != nil {
		children = append(children, i.input)
	}
	if i.textarea != nil {
		children = append(children, i.textarea)
	}
	if i.autoComp != nil {
		children = append(children, i.autoComp)
	}
	if i.multiSel != nil {
		children = append(children, i.multiSel)
	}
	if i.timePick != nil {
		children = append(children, i.timePick)
	}
	if i.dateRange != nil {
		children = append(children, i.dateRange)
	}
	if i.checkbox != nil {
		children = append(children, i.checkbox)
	}
	if i.selecter != nil {
		children = append(children, i.selecter)
	}
	if i.radioFast != nil {
		children = append(children, i.radioFast)
	}
	if i.radioSlow != nil {
		children = append(children, i.radioSlow)
	}
	if i.buttons != nil {
		children = append(children, i.buttons)
	}
	if i.status != nil {
		children = append(children, i.status)
	}
	return children
}

func truncateText(text string, max int) string {
	if len(text) <= max {
		return text
	}
	if max <= 3 {
		return text[:max]
	}
	return text[:max-3] + "..."
}
