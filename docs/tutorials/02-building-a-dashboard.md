# Tutorial 02: Building a Dashboard

This tutorial sketches a data dashboard using charts, tables, and signals.

<a href="../demos/sparkline.gif" target="_blank"><img src="../demos/sparkline.gif" alt="Live sparkline data"></a>

## Layout Structure

```go
table := widgets.NewTable(
    widgets.TableColumn{Title: "Service"},
    widgets.TableColumn{Title: "Status"},
    widgets.TableColumn{Title: "Latency"},
)

data := state.NewSignal([]widgets.BarData{{Label: "Auth", Value: 32}})
chart := widgets.NewBarChart(data)

left := widgets.NewPanel(table, widgets.WithPanelBorder(backend.DefaultStyle()))
right := widgets.NewPanel(chart, widgets.WithPanelBorder(backend.DefaultStyle()))

split := widgets.NewSplitter(left, right)
split.Ratio = 0.6
```

## Live Updates

```go
data.Update(func(values []widgets.BarData) []widgets.BarData {
    values[0].Value += 3
    return values
})
```

<a href="../demos/table.gif" target="_blank"><img src="../demos/table.gif" alt="Table widget"></a>

## Reference

See the full implementation in `examples/dashboard`.
