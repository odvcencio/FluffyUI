# Migrating from Bubble Tea

This guide maps common Bubble Tea patterns to FluffyUI concepts.

## Hello World

Bubble Tea:

```go
type model struct{}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg.(type) {
    case tea.KeyMsg:
        return m, tea.Quit
    }
    return m, nil
}

func (m model) View() string {
    return "Hello"
}
```

FluffyUI:

```go
app := fluffy.NewApp()
app.SetRoot(fluffy.NewLabel("Hello"))
app.Run(context.Background())
```

## Handling messages

Bubble Tea:

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "q" {
            return m, tea.Quit
        }
    }
    return m, nil
}
```

FluffyUI (custom update):

```go
app := fluffy.NewApp(fluffy.WithUpdate(func(app *runtime.App, msg runtime.Message) bool {
    switch m := msg.(type) {
    case runtime.KeyMsg:
        if m.Rune == 'q' {
            app.Post(runtime.Quit{})
            return true
        }
    }
    return false
}))
```

## Components / widgets

Bubble Tea often builds view strings directly. In FluffyUI, you compose
widgets and let them render into a buffer:

```go
root := fluffy.VStack(
    fluffy.Label("Title"),
    fluffy.Expanded(fluffy.NewText("Content")),
)
```

## Commands

Bubble Tea commands return from `Update`. In FluffyUI, widgets emit
`runtime.Command` values via `runtime.WithCommand(...)`, and the app handles
commands via `CommandHandler` or built-in routing (e.g., focus changes).

```go
app := fluffy.NewApp(fluffy.WithCommandHandler(func(cmd runtime.Command) bool {
    switch cmd.(type) {
    case runtime.Quit:
        return false
    }
    return false
}))
```

## Focus and input

Bubble Tea uses a focused model to route key input. FluffyUI has a focus scope
and focusable widgets. Use `FocusableBase` for custom widgets that need focus
and `FocusNext` / `FocusPrev` commands for navigation.

## Testing

Bubble Tea uses snapshot testing of views. FluffyUI offers deterministic
rendering through the simulation backend and `testing.RenderToString` helpers.
