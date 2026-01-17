package widgets_test

import "github.com/odvcencio/fluffy-ui/widgets"

func ExampleButton() {
	button := widgets.NewButton("Save", widgets.WithVariant(widgets.VariantPrimary))
	button.SetLabel("Save changes")
	_ = button
}

func ExampleCheckbox() {
	checked := true
	checkbox := widgets.NewCheckbox("Accept terms")
	checkbox.SetChecked(&checked)
	_ = checkbox
}

func ExampleRadio() {
	group := widgets.NewRadioGroup()
	first := widgets.NewRadio("First", group)
	second := widgets.NewRadio("Second", group)
	group.SetSelected(1)
	_ = first
	_ = second
}

func ExampleSelect() {
	selector := widgets.NewSelect(
		widgets.SelectOption{Label: "Small", Value: "S"},
		widgets.SelectOption{Label: "Medium", Value: "M"},
		widgets.SelectOption{Label: "Large", Value: "L"},
	)
	selector.SetSelected(1)
	_ = selector
}

func ExampleInput() {
	input := widgets.NewInput()
	input.SetPlaceholder("Search...")
	input.SetText("fluffy")
	input.OnSubmit(func(text string) {})
	_ = input
}

func ExampleTextArea() {
	area := widgets.NewTextArea()
	area.SetText("First line\nSecond line")
	area.OnChange(func(text string) {})
	_ = area
}
