# Data Widgets

Use these widgets to render structured data. A complete demo is in
`examples/widgets/data`.

## List

`List` renders a vertical list using a data adapter.

API notes:
- `NewList(adapter)` constructs the list.
- `NewSliceAdapter` and `NewSignalAdapter` wrap data sources.
- `OnSelect` notifies selection changes.
- `SetSelected` and `SelectedItem` allow external control.
- GoDoc example: `ExampleList`.

Example:

```go
items := []string{"Alpha", "Beta"}
adapter := widgets.NewSliceAdapter(items, func(item string, index int, selected bool, ctx runtime.RenderContext) {
    line := item
    if selected {
        line = "> " + line
    }
    ctx.Buffer.SetString(ctx.Bounds.X, ctx.Bounds.Y, line, backend.DefaultStyle())
})
list := widgets.NewList(adapter)
```

## Table

`Table` renders rows and columns with a header.

API notes:
- `NewTable(columns...)` defines columns.
- `SetRows(rows)` updates data.
- GoDoc example: `ExampleTable`.

Example:

```go
table := widgets.NewTable(
    widgets.TableColumn{Title: "Name"},
    widgets.TableColumn{Title: "Value"},
)
table.SetRows([][]string{{"A", "1"}, {"B", "2"}})
```

## Tree

`Tree` renders hierarchical data with expand/collapse state.

API notes:
- `TreeNode` defines the tree structure.
- `NewTree(root)` builds the widget.
- GoDoc example: `ExampleTree`.

Example:

```go
root := &widgets.TreeNode{Label: "Root", Expanded: true}
root.Children = []*widgets.TreeNode{{Label: "Child"}}

view := widgets.NewTree(root)
```

## SearchWidget

`SearchWidget` provides a search bar overlay, useful for filtering data.

API notes:
- `NewSearchWidget()` creates the widget.
- `SetOnSearch` wires filtering logic.
- `SetMatchInfo` shows results count.
- GoDoc example: `ExampleSearchWidget`.

Example:

```go
search := widgets.NewSearchWidget()
search.SetOnSearch(func(query string) {
    // filter data
})
```
