# Migrating from Bubble Tea

This guide maps common Bubble Tea patterns to FluffyUI equivalents.

Bubble Tea centers around TEA (`Model`, `Msg`, `Cmd`, `View() string`).
FluffyUI centers around a retained widget tree (`Measure`, `Layout`, `Render`)
plus reactive state (`Signal`, `Computed`) and command routing.

## 1) Program bootstrap

Bubble Tea:

```go
p := tea.NewProgram(model{})
if _, err := p.Run(); err != nil {
    log.Fatal(err)
}
```

FluffyUI:

```go
if err := fluffy.Run(fluffy.Label("Hello")); err != nil {
    log.Fatal(err)
}
```

For advanced wiring, use `fluffy.NewApp(...)`, set root widgets, then call
`app.Run(...)`.

## 2) `View() string` -> retained widgets

Bubble Tea:

```go
func (m model) View() string {
    return lipgloss.JoinVertical(
        lipgloss.Left,
        "Title",
        fmt.Sprintf("Count: %d", m.count),
    )
}
```

FluffyUI:

```go
root := fluffy.VStack(
    fluffy.Label("Title"),
    fluffy.Text("Count: 0"),
)
```

## 3) `tea.KeyMsg` handling

Bubble Tea:

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    if key, ok := msg.(tea.KeyMsg); ok && key.String() == "q" {
        return m, tea.Quit
    }
    return m, nil
}
```

FluffyUI custom update:

```go
app, err := fluffy.NewApp(fluffy.WithUpdate(func(app *runtime.App, msg runtime.Message) bool {
    switch m := msg.(type) {
    case runtime.KeyMsg:
        if m.Rune == 'q' {
            app.ExecuteCommand(runtime.Quit{})
            return true
        }
    }
    return runtime.DefaultUpdate(app, msg)
}))
```

Most apps can keep `runtime.DefaultUpdate` and let focused widgets handle key
messages directly.

## 4) `tea.WindowSizeMsg` handling

Bubble Tea:

```go
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
```

FluffyUI:

```go
app, err := fluffy.NewApp(
    fluffy.WithOnResize(func(app *runtime.App, width, height int) {
        // update state or metrics
    }),
)
```

Resize messages are also available via custom `Update` as `runtime.ResizeMsg`.

## 5) `tea.Cmd` -> commands/effects

Bubble Tea:

```go
func fetchCmd() tea.Cmd {
    return func() tea.Msg {
        return loadedMsg{Items: loadItems()}
    }
}
```

FluffyUI:

```go
app.Spawn(runtime.Effect{
    Run: func(ctx context.Context, post runtime.PostFunc) {
        items := loadItems()
        _ = post(runtime.CustomMsg{Value: loadedMsg{Items: items}})
    },
})
```

Widgets can also return commands from `HandleMessage` using
`runtime.WithCommand(...)`.

## 6) Model fields -> signals/computed state

Bubble Tea:

```go
type model struct {
    count int
}
```

FluffyUI:

```go
count := state.NewSignal(0)
double := state.Computed(func() int { return count.Get() * 2 })
```

Observe signal changes in `Bind` and call `services.Invalidate()` to trigger
render.

## 7) Sub-models -> child widgets

Bubble Tea often composes nested models manually and forwards messages.

FluffyUI composes child widgets structurally:

```go
root := fluffy.VStack(
    widgets.NewSearchWidget(),
    widgets.NewTable(columns...),
    fluffy.Label("Ready"),
)
```

Input routing follows focus and widget hierarchy. Containers expose children via
`runtime.ChildProvider`.

## 8) Multi-model communication

Bubble Tea:

- Parent updates child models via explicit message forwarding.
- Child emits messages that parent reinterprets.

FluffyUI:

- Share signals between widgets.
- Post `runtime.CustomMsg` for app-level events.
- Use command handlers for app-scoped actions.

```go
selectedID := state.NewSignal("")

list := widgets.NewList(items)
details := widgets.NewRichText("")

// Both widgets read/write selectedID through shared state.
```

## 9) Styling (`lipgloss`) -> style/theme/stylesheet

Bubble Tea:

```go
titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
```

FluffyUI:

```go
title := fluffy.Label(
    "Title",
    widgets.WithLabelStyle(
        backend.DefaultStyle().Bold(true).Foreground(backend.ColorMagenta),
    ),
)
```

For app-wide theming, use stylesheet rules (`docs/theming.md`) instead of
repeating inline styles.

## 10) Testing

Bubble Tea tests commonly assert against `View()` output.

FluffyUI provides deterministic rendering via simulation backend:

```go
output := fluffytest.RenderToString(fluffy.Label("Hello"), 20, 1)
if !strings.Contains(output, "Hello") {
    t.Fatal("expected label output")
}
```

Use `backend/sim` and `testing` helpers for key/mouse injection and golden
render assertions.

## Migration checklist

1. Replace top-level `tea.NewProgram(...)` with `fluffy.Run(...)` or `fluffy.NewApp(...)`.
2. Convert `View()` string assembly into widget composition.
3. Move mutable model fields into signals where reactive updates are useful.
4. Keep business logic in services/effects; keep `Render` pure and fast.
5. Add simulation-backed tests for focused interaction paths.
