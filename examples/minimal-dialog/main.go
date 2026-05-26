// Minimal modal dialog opened by a button, dismissed with Enter or Escape.
package main

import (
	"log"

	"m31labs.dev/fluffyui/fluffy"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/terminal"
	"m31labs.dev/fluffyui/widgets"
)

func centered(d *widgets.Dialog) runtime.Widget {
	o := widgets.NewSimpleWidget()
	o.MeasureFunc = func(c runtime.Constraints) runtime.Size { return d.Measure(c) }
	o.RenderFunc = func(ctx runtime.RenderContext) { d.Render(ctx) }
	o.LayoutFunc = func(b runtime.Rect) { d.Layout(d.CenteredBounds(b)) }
	o.HandleMessageFunc = func(msg runtime.Message) runtime.HandleResult {
		if k, ok := msg.(runtime.KeyMsg); ok && (k.Key == terminal.KeyEscape || k.Key == terminal.KeyEnter) {
			d.HandleMessage(msg); return runtime.WithCommand(runtime.PopOverlay{})
		}
		return d.HandleMessage(msg)
	}
	return o
}
func main() {
	status, open := fluffy.Signal("Press the button to open a dialog."), false
	app, inner := widgets.NewSimpleWidget(), fluffy.VStack(
		fluffy.ReactiveText(func() string { return status.Get() }, status),
		fluffy.Button("Open Dialog", func() { open = true }))
	app.MeasureFunc = func(c runtime.Constraints) runtime.Size { return inner.Measure(c) }
	app.LayoutFunc = func(b runtime.Rect) { inner.Layout(b) }
	app.RenderFunc = func(ctx runtime.RenderContext) { inner.Render(ctx) }
	app.HandleMessageFunc = func(msg runtime.Message) runtime.HandleResult {
		if open {
			open = false
			d := widgets.NewDialog("Confirm", "Are you sure?",
				widgets.DialogButton{Label: "OK", OnClick: func() { status.Set("OK!") }},
				widgets.DialogButton{Label: "Cancel", OnClick: func() { status.Set("Cancelled.") }})
			d.Focus(); return runtime.WithCommand(runtime.PushOverlay{Widget: centered(d), Modal: true})
		}
		return inner.HandleMessage(msg)
	}
	if err := fluffy.Run(app); err != nil { log.Fatal(err) }
}
