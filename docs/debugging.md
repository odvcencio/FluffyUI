# Debugging

FluffyUI includes tooling to help inspect layout and diagnose widget errors.

## ErrorReporter

`runtime.ErrorReporter` captures widget panics with context. It includes the
widget name, path, and (optionally) a tree or stack trace.

```go
app := runtime.NewApp(runtime.AppConfig{
    ErrorReporter: &runtime.ErrorReporter{
        ShowWidgetTree: true,
        ShowStackTrace: true,
    },
})
```

When a widget panics, the report includes a `Widget Path` like:

```
Grid[0,0] > Panel > Button
```

## DebugOverlay

Wrap your root widget with `DebugOverlay` to visualize bounds and layout:

```go
root := widgets.NewDebugOverlay(appRoot,
    widgets.WithDebugLabels(true),
)
app.SetRoot(root)
```

Use this in development to quickly spot layout and clipping issues.
