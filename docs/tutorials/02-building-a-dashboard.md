# Tutorial 02: Building a Dashboard

This tutorial sketches a data dashboard using charts, tables, and signals.

## Layout Structure

```go
table := widgets.NewTable(
    widgets.TableColumn{Title: "Service"},
    widgets.TableColumn{Title: "Status"},
    widgets.TableColumn{Title: "Latency"},
)

data := state.NewSignal([]widgets.BarData{{Label: "Auth", Value: 32}})
chart := widgets.NewBarChart(data)

left := widgets.NewPanel(table).WithBorder(backend.DefaultStyle())
right := widgets.NewPanel(chart).WithBorder(backend.DefaultStyle())

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

## Reference

See the full implementation in `examples/dashboard`.
