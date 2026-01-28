package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/examples/internal/demo"
	"github.com/odvcencio/fluffyui/forms"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/widgets"
)

func main() {
	view := NewFormView()
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

type FormView struct {
	widgets.Component
	form       *forms.Form
	title      *widgets.Label
	nameLabel  *widgets.Label
	emailLabel *widgets.Label
	nameInput  *widgets.Input
	emailInput *widgets.Input
	newsletter *widgets.Checkbox
	submitBtn  *widgets.Button
	resetBtn   *widgets.Button
	buttonRow  *demo.HBox
	status     *widgets.Label
	errors     *widgets.Text
}

func NewFormView() *FormView {
	nameField := forms.NewField("name", "", forms.Required("Name is required"), forms.MinLength(2, "Name is too short"))
	emailField := forms.NewField("email", "", forms.Required("Email is required"), forms.Email("Email looks invalid"))
	newsField := forms.NewField("newsletter", false)

	form := forms.NewForm(nameField, emailField, newsField)
	view := &FormView{form: form}

	view.title = widgets.NewLabel("Settings Form").WithStyle(backend.DefaultStyle().Bold(true))
	view.nameLabel = widgets.NewLabel("Name")
	view.emailLabel = widgets.NewLabel("Email")
	view.nameInput = widgets.NewInput()
	view.emailInput = widgets.NewInput()
	view.newsletter = widgets.NewCheckbox("Subscribe to updates")
	view.status = widgets.NewLabel("Ready")
	view.errors = widgets.NewText("")
	view.errors.SetStyle(backend.DefaultStyle().Foreground(backend.ColorRed))

	view.submitBtn = widgets.NewButton("Save", widgets.WithVariant(widgets.VariantPrimary), widgets.WithOnClick(func() {
		view.submit()
	}))
	view.resetBtn = widgets.NewButton("Reset", widgets.WithVariant(widgets.VariantSecondary), widgets.WithOnClick(func() {
		view.reset()
	}))
	view.buttonRow = demo.NewHBox(view.submitBtn, view.resetBtn)
	view.buttonRow.Gap = 2

	view.nameInput.OnChange(func(text string) {
		form.Set("name", text)
		view.validate()
	})
	view.emailInput.OnChange(func(text string) {
		form.Set("email", text)
		view.validate()
	})
	view.newsletter.SetOnChange(func(value *bool) {
		checked := false
		if value != nil {
			checked = *value
		}
		form.Set("newsletter", checked)
		view.validate()
	})

	form.OnSubmit(func(values forms.Values) {
		view.status.SetText("Saved")
		view.Invalidate()
	})
	form.OnCancel(func() {
		view.status.SetText("Canceled")
		view.Invalidate()
	})

	view.validate()
	return view
}

func (f *FormView) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (f *FormView) Layout(bounds runtime.Rect) {
	f.Component.Layout(bounds)
	y := bounds.Y

	layoutLine := func(w runtime.Widget) {
		if w == nil {
			return
		}
		w.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: 1})
		y++
	}

	layoutLine(f.title)
	layoutLine(f.nameLabel)
	layoutLine(f.nameInput)
	layoutLine(f.emailLabel)
	layoutLine(f.emailInput)
	layoutLine(f.newsletter)

	if f.buttonRow != nil {
		f.buttonRow.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: 1})
		y++
	}

	layoutLine(f.status)

	remaining := bounds.Height - (y - bounds.Y)
	if remaining < 1 {
		remaining = 1
	}
	if f.errors != nil {
		f.errors.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: remaining})
	}
}

func (f *FormView) Render(ctx runtime.RenderContext) {
	if f.title != nil {
		f.title.Render(ctx)
	}
	if f.nameLabel != nil {
		f.nameLabel.Render(ctx)
	}
	if f.emailLabel != nil {
		f.emailLabel.Render(ctx)
	}
	if f.nameInput != nil {
		f.nameInput.Render(ctx)
	}
	if f.emailInput != nil {
		f.emailInput.Render(ctx)
	}
	if f.newsletter != nil {
		f.newsletter.Render(ctx)
	}
	if f.buttonRow != nil {
		f.buttonRow.Render(ctx)
	}
	if f.status != nil {
		f.status.Render(ctx)
	}
	if f.errors != nil {
		f.errors.Render(ctx)
	}
}

func (f *FormView) HandleMessage(msg runtime.Message) runtime.HandleResult {
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

func (f *FormView) ChildWidgets() []runtime.Widget {
	children := []runtime.Widget{}
	if f.title != nil {
		children = append(children, f.title)
	}
	if f.nameLabel != nil {
		children = append(children, f.nameLabel)
	}
	if f.nameInput != nil {
		children = append(children, f.nameInput)
	}
	if f.emailLabel != nil {
		children = append(children, f.emailLabel)
	}
	if f.emailInput != nil {
		children = append(children, f.emailInput)
	}
	if f.newsletter != nil {
		children = append(children, f.newsletter)
	}
	if f.buttonRow != nil {
		children = append(children, f.buttonRow)
	}
	if f.status != nil {
		children = append(children, f.status)
	}
	if f.errors != nil {
		children = append(children, f.errors)
	}
	return children
}

func (f *FormView) submit() {
	if f.form == nil {
		return
	}
	f.form.Submit()
	f.validate()
}

func (f *FormView) reset() {
	if f.form == nil {
		return
	}
	f.form.Reset()
	if f.nameInput != nil {
		f.nameInput.SetText(toString(f.form.Get("name")))
	}
	if f.emailInput != nil {
		f.emailInput.SetText(toString(f.form.Get("email")))
	}
	if f.newsletter != nil {
		value, _ := f.form.Get("newsletter").(bool)
		f.newsletter.SetChecked(&value)
	}
	f.status.SetText("Reset")
	f.validate()
	f.Invalidate()
}

func (f *FormView) validate() {
	if f.form == nil || f.errors == nil {
		return
	}
	errors := f.form.Validate()
	if len(errors) == 0 {
		f.errors.SetText("No validation errors.")
		return
	}
	lines := make([]string, 0, len(errors))
	for _, err := range errors {
		lines = append(lines, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	f.errors.SetText(strings.Join(lines, "\n"))
}

func toString(value any) string {
	if value == nil {
		return ""
	}
	if s, ok := value.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", value)
}
