package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/examples/internal/demo"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	"github.com/odvcencio/fluffyui/terminal"
	"github.com/odvcencio/fluffyui/toast"
	"github.com/odvcencio/fluffyui/widgets"
)

func main() {
	view := NewFeedbackView()
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

type FeedbackView struct {
	widgets.Component
	header     *widgets.Label
	helper     *widgets.Label
	alert      *widgets.Alert
	progress   *widgets.Progress
	spinner    *widgets.Spinner
	spark      *widgets.Sparkline
	sparkData  *state.Signal[[]float64]
	toastMgr   *toast.ToastManager
	toastStack *widgets.ToastStack
	lastTick   time.Time
	lastToast  time.Time
}

func NewFeedbackView() *FeedbackView {
	view := &FeedbackView{}
	view.header = widgets.NewLabel("Feedback Widgets").WithStyle(backend.DefaultStyle().Bold(true))
	view.helper = widgets.NewLabel("Press D for dialog, T for toast")
	view.alert = widgets.NewAlert("All systems nominal", widgets.AlertSuccess)
	view.progress = widgets.NewProgress()
	view.progress.Value = 42
	view.spinner = widgets.NewSpinner()
	view.sparkData = state.NewSignal([]float64{10, 12, 9, 14, 11, 15})
	view.spark = widgets.NewSparkline(view.sparkData)

	view.toastMgr = toast.NewToastManager()
	view.toastStack = widgets.NewToastStack()
	view.toastMgr.SetOnChange(view.toastStack.SetToasts)

	return view
}

func (f *FeedbackView) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (f *FeedbackView) Layout(bounds runtime.Rect) {
	f.Component.Layout(bounds)
	y := bounds.Y

	line := func(w runtime.Widget, height int) {
		if w == nil {
			return
		}
		w.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: height})
		y += height
	}

	line(f.header, 1)
	line(f.helper, 1)
	line(f.alert, 1)
	line(f.progress, 1)
	line(f.spinner, 1)
	line(f.spark, 1)

	if f.toastStack != nil {
		f.toastStack.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: bounds.Height - (y - bounds.Y)})
	}
}

func (f *FeedbackView) Render(ctx runtime.RenderContext) {
	for _, child := range f.ChildWidgets() {
		if child != nil {
			child.Render(ctx)
		}
	}
}

func (f *FeedbackView) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if key, ok := msg.(runtime.KeyMsg); ok {
		switch key.Rune {
		case 'd', 'D':
			return runtime.WithCommand(runtime.PushOverlay{Widget: newDialogOverlay(), Modal: true})
		case 't', 'T':
			if f.toastMgr != nil {
				f.toastMgr.Info("Tip", "Keep pushing forward")
			}
			return runtime.Handled()
		}
	}
	if tick, ok := msg.(runtime.TickMsg); ok {
		f.onTick(tick.Time)
	}
	for _, child := range f.ChildWidgets() {
		if child == nil {
			continue
		}
		if result := child.HandleMessage(msg); result.Handled {
			return result
		}
	}
	return runtime.Unhandled()
}

func (f *FeedbackView) ChildWidgets() []runtime.Widget {
	children := []runtime.Widget{}
	if f.header != nil {
		children = append(children, f.header)
	}
	if f.helper != nil {
		children = append(children, f.helper)
	}
	if f.alert != nil {
		children = append(children, f.alert)
	}
	if f.progress != nil {
		children = append(children, f.progress)
	}
	if f.spinner != nil {
		children = append(children, f.spinner)
	}
	if f.spark != nil {
		children = append(children, f.spark)
	}
	if f.toastStack != nil {
		children = append(children, f.toastStack)
	}
	return children
}

func (f *FeedbackView) onTick(now time.Time) {
	if !f.lastTick.IsZero() && now.Sub(f.lastTick) < 500*time.Millisecond {
		return
	}
	f.lastTick = now
	f.progress.Value += 7
	if f.progress.Value > f.progress.Max {
		f.progress.Value = 0
	}
	f.sparkData.Update(func(values []float64) []float64 {
		if len(values) == 0 {
			return []float64{f.progress.Value}
		}
		return append(values[1:], f.progress.Value)
	})
	if f.toastMgr != nil && now.Sub(f.lastToast) > 4*time.Second {
		f.toastMgr.Success("Build", "Render cycle healthy")
		f.lastToast = now
	}
	f.Invalidate()
}

type dialogOverlay struct {
	dialog *widgets.Dialog
}

func newDialogOverlay() *dialogOverlay {
	d := widgets.NewDialog("Confirm", "Proceed with deployment?",
		widgets.DialogButton{Label: "OK"},
		widgets.DialogButton{Label: "Cancel"},
	)
	d.Focus()
	return &dialogOverlay{dialog: d}
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
