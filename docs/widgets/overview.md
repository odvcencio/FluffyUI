# Widgets Overview

FluffyUI ships with a growing catalog of 88+ widgets. The list below is grouped by
category with links to the relevant guides and examples.

GoDoc usage examples live in the `widgets` package (for example `ExampleList`,
`ExampleGrid`, and `ExampleTabs`).

## Layout

See `docs/widgets/layout.md` and `examples/widgets/layout`.

- Grid
- Flex (VStack / HStack)
- Splitter
- Stack
- ScrollView
- Panel and Box
- AspectRatio

## Data

See `docs/widgets/data.md` and `examples/widgets/data`.

<div style="display: grid; grid-template-columns: 1fr 1fr; gap: 1.5rem; margin: 1.5rem 0;">
<div>
<div class="fluffy-demo" data-cast="../demos/table.cast" data-cols="80" data-rows="24"></div>
</div>
<div>
<div class="fluffy-demo" data-cast="../demos/list.cast" data-cols="80" data-rows="24"></div>
</div>
</div>

- List
- Table
- DataGrid
- Tree
- DirectoryTree
- Log
- RichText
- SearchWidget

## Input

See `docs/widgets/input.md` and `examples/widgets/input`.

<div style="display: grid; grid-template-columns: 1fr 1fr; gap: 1.5rem; margin: 1.5rem 0;">
<div>
<div class="fluffy-demo" data-cast="../demos/buttons.cast" data-cols="80" data-rows="24"></div>
</div>
<div>
<div class="fluffy-demo" data-cast="../demos/input.cast" data-cols="80" data-rows="24"></div>
</div>
</div>

- Button
- Checkbox
- Radio
- Select
- AutoComplete
- MultiSelect
- Input
- MaskedInput
- TextArea
- Slider
- RangeSlider
- DatePicker
- DateRangePicker
- TimePicker

## Navigation

See `docs/widgets/navigation.md` and `examples/widgets/navigation`.

<div style="margin: 1.5rem 0;">
<div class="fluffy-demo" data-cast="../demos/tabs.cast" data-cols="80" data-rows="24"></div>
</div>

- Tabs
- Menu
- Breadcrumb
- Stepper
- PaletteWidget and EnhancedPalette
- Accordion
- Section

## Feedback

See `docs/widgets/feedback.md` and `examples/widgets/feedback`.

<div style="display: grid; grid-template-columns: 1fr 1fr; gap: 1.5rem; margin: 1.5rem 0;">
<div>
<div class="fluffy-demo" data-cast="../demos/dialog.cast" data-cols="80" data-rows="24"></div>
</div>
<div>
<div class="fluffy-demo" data-cast="../demos/sparkline.cast" data-cols="80" data-rows="24"></div>
</div>
</div>

- Dialog
- Spinner
- Progress
- Alert
- ToastStack
- Charts (Sparkline, BarChart, LineChart)

## Developer helpers

- SimpleWidget (function-based widget for quick prototypes)
- DebugOverlay (visualize bounds and layout)
- AsyncImage (load images without blocking)

## Building your own

See `docs/widgets/custom.md` for a step-by-step custom widget guide.
