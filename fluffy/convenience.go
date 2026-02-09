package fluffy

import "github.com/odvcencio/fluffyui/widgets"

// Button creates a button with a label and click handler.
// This is a shorthand for NewButton(label, WithOnClick(onClick)).
func Button(label string, onClick func()) *widgets.Button {
	return widgets.NewButton(label, widgets.WithOnClick(onClick))
}

// Input creates a text input with a placeholder.
// This is a shorthand for creating an Input and setting its placeholder.
func Input(placeholder string) *widgets.Input {
	inp := widgets.NewInput()
	inp.SetPlaceholder(placeholder)
	return inp
}

// Checkbox creates a checkbox with a label and change handler.
// The onChange callback receives the new checked state as a *bool
// (nil indicates indeterminate).
func Checkbox(label string, onChange func(*bool)) *widgets.Checkbox {
	cb := widgets.NewCheckbox(label)
	if onChange != nil {
		cb.SetOnChange(onChange)
	}
	return cb
}

// SelectFromStrings creates a select widget from string option labels and a change handler.
// Each string becomes a SelectOption with the string as both Label and Value.
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
