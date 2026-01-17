# Feedback Widgets

Feedback widgets communicate status and progress. A complete demo is in
`examples/widgets/feedback`, and `examples/dashboard` shows a data-heavy layout.

## Dialog

API notes:
- `NewDialog(title, body, buttons...)` creates a modal dialog.
- Use `runtime.PushOverlay` to display it.
- GoDoc example: `ExampleDialog`.

Example:

```go
dialog := widgets.NewDialog("Confirm", "Delete this item?",
    widgets.DialogButton{Label: "OK", OnClick: confirm},
)
```

## Spinner

API notes:
- `NewSpinner()` creates the indicator.
- `HandleMessage` advances on tick messages.
- GoDoc example: `ExampleSpinner`.

Example:

```go
spinner := widgets.NewSpinner()
```

## Progress

API notes:
- `NewProgress()` creates a bar.
- Set `Value` and `Max`.
- GoDoc example: `ExampleProgress`.

Example:

```go
progress := widgets.NewProgress()
progress.Value = 65
```

## Alert

API notes:
- `NewAlert(text, variant)` creates an alert.
- GoDoc example: `ExampleAlert`.

Example:

```go
alert := widgets.NewAlert("All systems nominal", widgets.AlertSuccess)
```

## ToastStack

API notes:
- `ToastManager` manages toasts.
- `ToastStack` renders them.
- GoDoc example: `ExampleToastStack`.

Example:

```go
manager := toast.NewToastManager()
stack := widgets.NewToastStack()
manager.SetOnChange(stack.SetToasts)
```

## Charts

API notes:
- `NewSparkline(signal)` renders compact trends.
- `NewBarChart(signal)` renders horizontal bars.
- GoDoc example: `ExampleSparkline`, `ExampleBarChart`.

Example:

```go
spark := widgets.NewSparkline(state.NewSignal([]float64{1, 2, 3}))
```
