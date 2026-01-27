# Theming

FluffyUI supports two styling layers:

- **Stylesheets** for CSS-like theme rules across widgets.
- **Inline styles** via `backend.Style` for per-widget overrides (still supported).

## Stylesheets

Use the `style` package to define rules and attach them to your app:

```go
sheet := style.NewStylesheet().
    Add(style.Select("*"), style.Style{
        Foreground: style.RGB(240, 238, 232),
        Background: style.RGB(12, 12, 16),
    }).
    Add(style.Select("Button").Class("primary"), style.Style{
        Foreground: style.RGB(12, 12, 16),
        Background: style.RGB(255, 183, 77),
        Bold:       style.Bool(true),
    }).
    Add(style.Select("Button").Pseudo(style.PseudoFocus), style.Style{
        Reverse: style.Bool(true),
    })

app := runtime.NewApp(runtime.AppConfig{
    Root:       root,
    Stylesheet: sheet,
})
```

Widgets can opt into classes and IDs:

```go
btn := widgets.NewButton("Submit", widgets.WithClass("primary"))
btn.SetID("submit-btn")
```

## External stylesheets (.fss)

You can define styles in a `.fss` file and parse them at runtime:

```go
sheet, err := style.ParseFile("app.fss")
if err != nil {
    return err
}

app.SetStylesheet(sheet)
```

### Media queries

FSS supports a focused media query subset:

```fss
@media (min-width: 80) and (orientation: landscape) {
  Button { padding: 2; }
}

@media (prefers-reduced-motion: reduce) {
  Spinner { dim: true; }
}
```

### Hot reload (dev)

Use a file watcher to reload on edits:

```go
stop := style.WatchFile("app.fss", time.Second, func(sheet *style.Stylesheet, err error) {
    if err != nil || sheet == nil {
        return
    }
    app.Services().Scheduler().Schedule(func() {
        app.SetStylesheet(sheet)
    })
})
defer stop()
```

## Layout properties

Styles can also influence layout:

```go
sheet := style.NewStylesheet().
    Add(style.Select("Button"), style.Style{
        Padding: style.PadXY(2, 1),
        Margin:  style.Pad(1),
        Width:   style.Fixed(12),
    })
```

Padding and margin affect measurement and `ContentBounds` for widgets that embed
`widgets.Base`. Width/height rules apply to the widget's border box (padding and
border included). Borders contribute to layout for widgets that render them
(for example `Panel` and `Dialog`).

## Default stylesheet

The `theme` package provides a starter stylesheet based on the default palette:

```go
app := runtime.NewApp(runtime.AppConfig{
    Root:       root,
    Stylesheet: theme.DefaultStylesheet(),
})
```

For full code + style reload during development, use:

```bash
go run ./cmd/fluffy dev -- go run ./examples/quickstart
```

## Inline styles (legacy + overrides)

You can still use `backend.Style` setters for local overrides:

```go
label := widgets.NewLabel("Status")
label.SetStyle(backend.DefaultStyle().Foreground(backend.ColorGreen).Bold(true))
```

Inline styles apply when no stylesheet rule matches, and can also override
stylesheet output when explicitly set.
