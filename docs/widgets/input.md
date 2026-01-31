# Input Widgets

Input widgets capture user actions and text. A complete demo is in
`examples/widgets/input`, and `examples/settings-form` shows form wiring.

## Button

API notes:
- `NewButton(label, options...)` creates a button.
- Use `WithVariant` and `WithOnClick` options.
- GoDoc example: `ExampleButton`.

Example:

```go
btn := widgets.NewButton("Save", widgets.WithOnClick(func() { save() }))
```

## Checkbox

API notes:
- `NewCheckbox(label)` creates a checkbox.
- `SetOnChange` handles toggles.
- GoDoc example: `ExampleCheckbox`.

Example:

```go
cb := widgets.NewCheckbox("Enable")
cb.SetOnChange(func(value *bool) {
    // value can be nil for indeterminate
})
```

## Radio

API notes:
- `NewRadioGroup()` manages selection.
- `NewRadio(label, group)` creates options.
- GoDoc example: `ExampleRadio`.

Example:

```go
group := widgets.NewRadioGroup()
fast := widgets.NewRadio("Fast", group)
slow := widgets.NewRadio("Slow", group)
```

## Select

API notes:
- `NewSelect(options...)` creates a selector.
- `SetOnChange` is invoked on selection changes.
- GoDoc example: `ExampleSelect`.

Example:

```go
selecter := widgets.NewSelect(
    widgets.SelectOption{Label: "Low"},
    widgets.SelectOption{Label: "High"},
)
```

## AutoComplete

`AutoComplete` combines an input with inline suggestions.

API notes:
- `SetOptions` or `SetProvider` populates suggestions.
- `SetOnSelect` fires when a suggestion is chosen.

Example:

```go
ac := widgets.NewAutoComplete()
ac.SetOptions([]string{"Alpha", "Beta", "Gamma"})
```

## Input

`Input` is a single-line text editor with cursor support.

API notes:
- `SetPlaceholder`, `OnSubmit`, and `OnChange` provide hooks.
- GoDoc example: `ExampleInput`.

Example:

```go
input := widgets.NewInput()
input.SetPlaceholder("Search")
input.OnSubmit(func(text string) { fmt.Println(text) })
```

## MultiSelect

`MultiSelect` allows selecting multiple options in a list.

API notes:
- `SetOptions` updates choices.
- `SetOnChange` receives selected options.

Example:

```go
ms := widgets.NewMultiSelect(
    widgets.MultiSelectOption{Label: "One"},
    widgets.MultiSelectOption{Label: "Two"},
)
```

## TextArea

`TextArea` is a multi-line text editor with scrolling.

API notes:
- `SetText` updates content.
- `OnChange` notifies edits.
- GoDoc example: `ExampleTextArea`.

Example:

```go
area := widgets.NewTextArea()
area.SetText("Multi-line\ninput")
```

## DateRangePicker

`DateRangePicker` combines two inputs with a range-select calendar.

API notes:
- `SetRange` updates the current range.
- `OnRangeSelect` fires on calendar selection.

Example:

```go
picker := widgets.NewDateRangePicker()
```

## TimePicker

`TimePicker` provides a time-of-day picker with keyboard input.

API notes:
- `SetShowSeconds` toggles second display.
- `SetTime` updates the selection.

Example:

```go
tp := widgets.NewTimePicker()
tp.SetShowSeconds(true)
```
