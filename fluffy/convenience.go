package fluffy

import (
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	"github.com/odvcencio/fluffyui/widgets"
)

// Button creates a button with a label and click handler.
//
// This is a shorthand for widgets.NewButton(label, WithOnClick(onClick)).
// Use this when you need a simple clickable button with a single handler.
// For buttons that need variants, disabled state, or loading indicators,
// use NewButton with option functions instead.
//
// Example:
//
//	btn := fluffy.Button("Save", func() { save() })
func Button(label string, onClick func()) *widgets.Button {
	return widgets.NewButton(label, widgets.WithOnClick(onClick))
}

// Input creates a text input with a placeholder string.
//
// This is a shorthand for creating a widgets.Input and calling SetPlaceholder.
// Use this for simple text fields where a placeholder is sufficient.
// For inputs that need validators, submit handlers, or custom styling,
// use NewInput and configure it with setter methods instead.
//
// Example:
//
//	email := fluffy.Input("you@example.com")
func Input(placeholder string) *widgets.Input {
	inp := widgets.NewInput()
	inp.SetPlaceholder(placeholder)
	return inp
}

// Checkbox creates a checkbox with a label and change handler.
//
// The onChange callback receives the new checked state as a *bool
// (nil indicates indeterminate). Pass nil for onChange if no handler is needed.
// Use this for simple checkboxes. For checkboxes that need custom styling
// or indeterminate state management, use NewCheckbox and configure it directly.
//
// Example:
//
//	agree := fluffy.Checkbox("I agree", func(checked *bool) {
//	    if checked != nil && *checked {
//	        enableSubmit()
//	    }
//	})
func Checkbox(label string, onChange func(*bool)) *widgets.Checkbox {
	cb := widgets.NewCheckbox(label)
	if onChange != nil {
		cb.SetOnChange(onChange)
	}
	return cb
}

// SelectFromStrings creates a select widget from string option labels and a change handler.
//
// Each string becomes a SelectOption with the string as both Label and Value.
// The onChange callback receives the label of the selected option.
// Use this for simple string-based dropdowns. For selects that need separate
// labels and values, or typed SelectOption structs, use widgets.NewSelect directly.
//
// Example:
//
//	color := fluffy.SelectFromStrings(
//	    []string{"Red", "Green", "Blue"},
//	    func(selected string) { fmt.Println("Picked:", selected) },
//	)
func SelectFromStrings(options []string, onChange func(string)) *widgets.Select {
	selectOpts := make([]widgets.SelectOption, len(options))
	for i, opt := range options {
		selectOpts[i] = widgets.SelectOption{Label: opt, Value: opt}
	}
	s := widgets.NewSelect(selectOpts...)
	if onChange != nil {
		s.SetOnChange(func(opt widgets.SelectOption) {
			onChange(opt.Label)
		})
	}
	return s
}

// ReactiveText creates a text widget that automatically updates when
// any of the given dependencies change. The render function is called
// to produce the displayed text content.
//
// This is the simplest way to display reactive data. For static text,
// use Label instead. For rich formatted text, use widgets.NewSignalLabel
// directly with a custom format function.
//
// Example:
//
//	count := fluffy.Signal(0)
//	label := fluffy.ReactiveText(func() string {
//	    return fmt.Sprintf("Count: %d", count.Get())
//	}, count)
func ReactiveText(render func() string, deps ...state.Subscribable) *widgets.SignalLabel {
	computed := state.NewComputed(render, deps...)
	return widgets.NewSignalLabel(computed, nil)
}

// Form creates a form with labeled fields, validation, and submit handling.
//
// This is a shorthand for widgets.NewForm(fields...).
// Use this when you need a simple form that collects text input from
// labeled fields with built-in validation and Tab navigation.
//
// Example:
//
//	form := fluffy.Form(
//	    fluffy.FormField("Name", fluffy.Input("Enter name")),
//	    fluffy.FormField("Email", fluffy.Input("Enter email")),
//	)
//	form.SetOnSubmit(func(values map[string]string) { ... })
func Form(fields ...widgets.FormFieldDef) *widgets.Form {
	return widgets.NewForm(fields...)
}

// FormField creates a labeled form field definition for use with Form.
//
// This is a shorthand for widgets.FormField(label, widget).
func FormField(label string, widget runtime.Widget) widgets.FormFieldDef {
	return widgets.FormField(label, widget)
}
