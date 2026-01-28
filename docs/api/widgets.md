# Widgets API

FluffyUI ships with a catalog of composable widgets. Most widgets embed
`widgets.Base` for layout and accessibility, and can be styled via FSS.

## Basics

```go
label := widgets.NewLabel("Hello")
button := widgets.NewButton("Submit").OnClick(func() { /* ... */ })
```

## Layout & Containers

- `Panel`, `Box`, `Section` (framed containers)
- `Grid`, `Stack`, `Splitter` (layout primitives)
- `ScrollView` (scrollable child)
- `Dialog` (modal overlays)
- `CanvasWidget` (custom canvas renderer)

## Inputs & Forms

- `Input`, `MultilineInput`
- `Checkbox`, `Radio`, `Select`, `Stepper`
- `Search` (filterable list/search entry)

## Data & Charts

- `Table`, `List`, `Tree`
- `Sparkline`, `BarChart`, `LineChart`
- `Gauge`, `Progress`, `ProgressBar`

## Navigation

- `Tabs`, `Breadcrumb`, `Menu`
- `Palette`, `EnhancedPalette` (command palettes)

## Feedback & Status

- `Alert`, `Toasts`, `Spinner`, `SignalLabel`
- `AnimatedGauge`, `AnimatedWidget` (motion wrappers)

## Graphics & Media

- `CanvasWidget` (draw with `graphics.Canvas`)
- `VideoPlayer` (render decoded frames)

## Text & Display

- `Text`, `Label`

## Examples

Browse interactive examples in `examples/widgets/*` and larger demos in
`examples/dashboard`, `examples/candy-wars`, and `examples/video-player`.
