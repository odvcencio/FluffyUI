# Theming

FluffyUI uses `backend.Style` to control colors and attributes. Widgets expose
style setters so you can customize per component, or you can build your own
theme helpers.

## Styles

```go
style := backend.DefaultStyle().Foreground(backend.ColorGreen).Bold(true)
label := widgets.NewLabel("Status")
label.SetStyle(style)
```

## Widget-level styling

Many widgets provide setters for normal and focused styles:

- `Button.SetStyle`, `Button.SetFocusStyle`
- `Input.SetStyle`, `Input.SetFocusStyle`
- `Checkbox.SetOnChange` for updating labels or state

## Theme helpers

The `theme` package includes helpers for consistent styling. Use it as a
starting point or replace it with your own theme system.
