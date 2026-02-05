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

## Layout diagnostics

Enable layout diagnostics to warn about suspicious measurements (for example
zero-size measurements under non-zero constraints):

```bash
FLUFFYUI_LAYOUT_DEBUG=1 go run ./examples/quickstart
```

Or toggle at runtime:

```go
runtime.SetLayoutDebug(true)
```

## `fluffy doctor`

Use the CLI diagnostic to validate terminal capabilities before debugging app
render issues:

```bash
go run ./cmd/fluffy doctor
```

It reports terminal baseline support, true color, mouse assumptions, Unicode
width checks, Kitty/Sixel detection, and accessibility bridge readiness.

Inline rendering tip:

Use `fluffy.WithInlineMode(true)` and optionally `fluffy.WithInlineHeight(n)`
to run a bounded in-terminal UI without alternate-screen takeover.
For environment-specific caveats, see `docs/inline-mode.md`.
