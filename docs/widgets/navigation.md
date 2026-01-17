# Navigation Widgets

Navigation widgets help users move through workflows. A complete demo is in
`examples/widgets/navigation`, and the command palette demo is in
`examples/command-palette`.

## Tabs

API notes:
- `NewTabs(tabs...)` creates a tab container.
- Tab content is another widget.
- GoDoc example: `ExampleTabs`.

Example:

```go
tabs := widgets.NewTabs(
    widgets.Tab{Title: "Overview", Content: overview},
    widgets.Tab{Title: "Details", Content: details},
)
```

## Menu

API notes:
- `NewMenu(items...)` creates a vertical menu.
- `MenuItem` supports nesting and callbacks.
- GoDoc example: `ExampleMenu`.

Example:

```go
menu := widgets.NewMenu(
    &widgets.MenuItem{Title: "Open"},
    &widgets.MenuItem{Title: "Save"},
)
```

## Breadcrumb

API notes:
- `NewBreadcrumb(items...)` creates a path display.
- GoDoc example: `ExampleBreadcrumb`.

Example:

```go
crumbs := widgets.NewBreadcrumb(
    widgets.BreadcrumbItem{Label: "Home"},
    widgets.BreadcrumbItem{Label: "Projects"},
)
```

## Stepper

API notes:
- `NewStepper(steps...)` builds a step list.
- Each step has a `State`.
- GoDoc example: `ExampleStepper`.

Example:

```go
stepper := widgets.NewStepper(
    widgets.Step{Title: "Plan", State: widgets.StepCompleted},
    widgets.Step{Title: "Ship", State: widgets.StepActive},
)
```

## PaletteWidget and EnhancedPalette

API notes:
- `NewPaletteWidget(title)` creates a fuzzy search palette.
- `NewEnhancedPalette(registry)` wires to the keybind registry.
- GoDoc example: `ExamplePaletteWidget`, `ExampleEnhancedPalette`.

Example:

```go
palette := widgets.NewPaletteWidget("Quick Actions")
palette.SetItems(items)
```
