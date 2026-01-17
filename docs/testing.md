# Testing

FluffyUI includes a simulation backend for deterministic tests. This allows
widget rendering and input handling without a real terminal.

## Simulation backend

```go
be := sim.New(40, 10)
if err := be.Init(); err != nil {
    t.Fatalf("init: %v", err)
}
defer be.Fini()

app := runtime.NewApp(runtime.AppConfig{Backend: be})
app.SetRoot(widgets.NewLabel("Hello"))

ctx, cancel := context.WithCancel(context.Background())
cancel()
_ = app.Run(ctx)
```

## Capturing output

The simulation backend can capture rendered output for assertions:

```go
be.Show()
if !be.ContainsText("Hello") {
    t.Fatalf("expected label")
}
```

## Input injection

Inject keys or mouse events directly on the backend:

```go
be.InjectKeyRune('a')
```

See `backend/sim` tests for additional helpers.
