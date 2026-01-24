# Tutorial 01: Getting Started

This tutorial builds a tiny FluffyUI app with a label and a button.

## 1. Create a New App

```go
root := widgets.NewLabel("Hello from FluffyUI")
app := runtime.NewApp(runtime.AppConfig{Backend: backend, Root: root})
```

## 2. Run the App

```go
if err := app.Run(context.Background()); err != nil && err != context.Canceled {
    log.Fatal(err)
}
```

## 3. Next Steps

- Explore widgets in `examples/widgets/*`.
- Try the higher-level demo helper in `examples/internal/demo`.
- Add styles with FSS (`docs/theming.md`).
