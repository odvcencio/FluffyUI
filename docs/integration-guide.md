# Integration Guide

This guide explains how to build applications on top of FluffyUI, from simple
widget-based UIs to complex applications with custom event loops.

## Integration Patterns

FluffyUI supports multiple integration levels:

| Level | Description | Use Case |
|-------|-------------|----------|
| **Full Runtime** | Use `runtime.App` with standard loop | Simple to medium apps |
| **Screen Only** | Use `runtime.Screen` with custom loop | Apps needing custom event handling |
| **Widgets Only** | Use widgets with custom rendering | Embedding in existing TUIs |

## Pattern 1: Full Runtime (Recommended)

The simplest approach uses `runtime.App` which handles the event loop,
rendering, and focus management automatically.

```go
package main

import (
    "context"
    "github.com/odvcencio/fluffyui/fluffy"
)

func main() {
    app := fluffy.NewApp()

    // Build your widget tree
    root := fluffy.VStack(
        fluffy.Label("Header"),
        fluffy.Expanded(fluffy.Label("Content")),
        fluffy.Label("Footer"),
    )

    app.SetRoot(root)
    app.Run(context.Background())
}
```

### Handling Custom Messages

Use `AppConfig.Update` to handle custom messages:

```go
type CounterMsg struct{ Delta int }

app := fluffy.NewApp(fluffy.WithUpdate(func(app *runtime.App, msg runtime.Message) bool {
        switch m := msg.(type) {
        case CounterMsg:
            counter.Increment(m.Delta)
            return true // request render
        }
        return false
}))

// Post messages from anywhere
app.Post(CounterMsg{Delta: 1})
```

### Handling Commands

Widgets emit commands for app-level actions:

```go
app := fluffy.NewApp(fluffy.WithCommandHandler(func(cmd runtime.Command) bool {
        switch c := cmd.(type) {
        case MyCustomCommand:
            handleCustomCommand(c)
            return true
        }
        return false
    },
}))
```

## Pattern 2: Screen Only (Advanced)

For applications needing custom event loops (streaming APIs, complex
concurrency, domain-specific timing), use `runtime.Screen` directly.

This is the pattern used by Buckley.

```go
package main

import (
    "github.com/odvcencio/fluffyui/backend/tcell"
    "github.com/odvcencio/fluffyui/runtime"
    "github.com/odvcencio/fluffyui/terminal"
)

type App struct {
    backend  *tcell.Backend
    screen   *runtime.Screen
    running  bool
    messages chan Message
}

func NewApp() (*App, error) {
    be, err := tcell.New()
    if err != nil {
        return nil, err
    }
    if err := be.Init(); err != nil {
        return nil, err
    }

    w, h := be.Size()
    screen := runtime.NewScreen(w, h)

    return &App{
        backend:  be,
        screen:   screen,
        messages: make(chan Message, 100),
    }, nil
}

func (a *App) Run() error {
    // Start event polling goroutine
    go a.pollEvents()

    // Create your widget tree
    root := buildWidgetTree()
    a.screen.SetRoot(root)

    // Main event loop
    ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
    defer ticker.Stop()

    a.running = true
    for a.running {
        select {
        case msg := <-a.messages:
            a.handleMessage(msg)
        case <-ticker.C:
            a.render()
        }
    }

    a.backend.Fini()
    return nil
}

func (a *App) pollEvents() {
    for a.running {
        ev := a.backend.PollEvent()
        switch e := ev.(type) {
        case terminal.KeyEvent:
            a.messages <- KeyMsg{Key: e.Key, Rune: e.Rune}
        case terminal.ResizeEvent:
            a.messages <- ResizeMsg{Width: e.Width, Height: e.Height}
        }
    }
}

func (a *App) handleMessage(msg Message) {
    // Convert to runtime message and dispatch to screen
    switch m := msg.(type) {
    case KeyMsg:
        runtimeMsg := runtime.KeyMsg{Key: m.Key, Rune: m.Rune}
        result := a.screen.HandleMessage(runtimeMsg)
        for _, cmd := range result.Commands {
            a.handleCommand(cmd)
        }
    case ResizeMsg:
        a.screen.SetSize(m.Width, m.Height)
    }
}

func (a *App) render() {
    a.screen.Render()
    buf := a.screen.Buffer()

    // Copy buffer to backend
    w, h := a.backend.Size()
    for y := 0; y < h; y++ {
        for x := 0; x < w; x++ {
            cell := buf.Get(x, y)
            a.backend.SetContent(x, y, cell.Rune, nil, cell.Style)
        }
    }
    a.backend.Show()
}
```

### Key Differences from Full Runtime

| Aspect | Full Runtime | Screen Only |
|--------|--------------|-------------|
| Event loop | Managed by App | You implement it |
| Tick timing | Configurable via TickRate | You control timing |
| Message passing | App.Post() | Your own channel |
| Focus management | Automatic | Manual or via Screen |
| Effects | App.Spawn() | Your own goroutines |

### When to Use Screen Only

- **Streaming APIs**: Need to process chunks as they arrive
- **Custom concurrency**: Domain-specific threading model
- **Integration**: Embedding in existing event loop
- **Performance**: Fine-grained control over render timing

## Pattern 3: Widgets Only

For embedding FluffyUI widgets in other TUI frameworks or testing:

```go
func renderWidget(w runtime.Widget, width, height int) string {
    buf := runtime.NewBuffer(width, height)

    constraints := runtime.Constraints{MaxWidth: width, MaxHeight: height}
    w.Measure(constraints)
    w.Layout(runtime.Rect{X: 0, Y: 0, Width: width, Height: height})

    ctx := runtime.RenderContext{Buffer: buf}
    w.Render(ctx)

    // Convert buffer to string
    var sb strings.Builder
    for y := 0; y < height; y++ {
        for x := 0; x < width; x++ {
            cell := buf.Get(x, y)
            if cell.Rune == 0 {
                sb.WriteRune(' ')
            } else {
                sb.WriteRune(cell.Rune)
            }
        }
        sb.WriteRune('\n')
    }
    return sb.String()
}
```

## Building Custom Widgets

All custom widgets should embed `widgets.Base` or `widgets.FocusableBase`:

```go
type MyWidget struct {
    widgets.Base  // or widgets.FocusableBase for focusable widgets

    label string
    count int
}

func NewMyWidget(label string) *MyWidget {
    return &MyWidget{label: label}
}

func (w *MyWidget) Measure(c runtime.Constraints) runtime.Size {
    // Report desired size
    return runtime.Size{Width: len(w.label) + 10, Height: 1}
}

func (w *MyWidget) Render(ctx runtime.RenderContext) {
    text := fmt.Sprintf("%s: %d", w.label, w.count)
    for i, r := range text {
        if i < ctx.Bounds.Width {
            ctx.Buffer.Set(ctx.Bounds.X+i, ctx.Bounds.Y, r, backend.DefaultStyle())
        }
    }
}

func (w *MyWidget) HandleMessage(msg runtime.Message) runtime.HandleResult {
    if key, ok := msg.(runtime.KeyMsg); ok {
        if key.Rune == '+' {
            w.count++
            w.Invalidate() // Request re-render
            return runtime.Handled()
        }
    }
    return runtime.Unhandled()
}
```

## Using Reactive State

FluffyUI provides reactive primitives for automatic UI updates:

```go
import "github.com/odvcencio/fluffyui/state"

// Create signals
counter := state.NewSignal(0)
doubled := state.NewComputed(func() int {
    return counter.Get() * 2
})

// Explicit dependencies are still supported
explicit := state.NewComputed(func() int {
    return counter.Get() * 2
}, counter)

// Subscribe to changes
counter.Subscribe(func() {
    fmt.Println("Counter changed:", counter.Get())
})

// Update triggers subscribers
counter.Set(5) // prints "Counter changed: 5"
```

Auto-tracking is designed for the single UI goroutine. If signals are read
across multiple goroutines, prefer explicit dependencies.

### Signals in Widgets

Use `widgets.Component` for automatic signal subscription management:

```go
type CounterWidget struct {
    widgets.Component
    count *state.Signal[int]
}

func (w *CounterWidget) Bind(services runtime.Services) {
    w.Component.Bind(services)

    // Observe automatically invalidates on change
    w.Observe(w.count)
}
```

## Audio (Music + SFX)

FluffyUI exposes an opinionated audio service for music and sound effects. You
wire a driver once at app setup and request cues by ID from widgets.

```go
type GameHUD struct {
    widgets.Component
    audio audio.Service
}

func (g *GameHUD) Bind(services runtime.Services) {
    g.Component.Bind(services)
    g.audio = services.Audio()
}

func (g *GameHUD) OnClick() {
    if g.audio != nil {
        g.audio.PlaySFX("ui.click")
    }
}
```

See `docs/audio.md` for cue registration and driver setup.

## Layer System

FluffyUI supports modal overlays via the layer system:

```go
// Base layer (main content)
screen.SetRoot(mainContent)

// Push modal dialog
screen.PushLayer(dialog, true)  // true = modal, blocks input to layers below

// Pop when done
screen.PopLayer()
```

## Focus Management

The screen manages focus scopes per layer:

```go
// Get the current focus scope
scope := screen.FocusScope()

// Navigate focus
scope.FocusNext()
scope.FocusPrev()

// Set focus directly
scope.SetFocus(myWidget)
```

## Example: Building a Chat Application

Here's a complete example showing these patterns together:

```go
package main

import (
    "context"
    "github.com/odvcencio/fluffyui/backend/tcell"
    "github.com/odvcencio/fluffyui/runtime"
    "github.com/odvcencio/fluffyui/state"
    "github.com/odvcencio/fluffyui/widgets"
)

func main() {
    be, _ := tcell.New()

    // State
    messages := state.NewSignal([]string{})
    input := state.NewSignal("")

    // Widgets
    messageList := widgets.NewList(
        widgets.NewSignalAdapter(messages, func(msg string, i int, sel bool, ctx runtime.RenderContext) {
            widgets.RenderText(ctx, msg, backend.DefaultStyle())
        }),
    )

    inputBox := widgets.NewInput()
    inputBox.BindText(input)

    // Layout
    root := runtime.VBox(
        runtime.Expanded(messageList),
        runtime.Fixed(inputBox),
    )

    // App
    app := runtime.NewApp(runtime.AppConfig{
        Backend: be,
        CommandHandler: func(cmd runtime.Command) bool {
            if _, ok := cmd.(runtime.Submit); ok {
                text := input.Get()
                if text != "" {
                    messages.Update(func(msgs []string) []string {
                        return append(msgs, text)
                    })
                    input.Set("")
                }
                return true
            }
            return false
        },
    })

    app.SetRoot(root)
    app.Run(context.Background())
}
```

## Best Practices

1. **Start with Full Runtime**: Use `runtime.App` unless you have specific needs
2. **Use Signals for State**: Reactive updates are cleaner than manual invalidation
3. **Embed Base Types**: Always embed `widgets.Base` or `widgets.FocusableBase`
4. **Layer for Modals**: Use `PushLayer` for dialogs, don't manage z-order manually
5. **Test with Simulation**: Use `backend/sim` for deterministic testing

## Further Reading

- [Architecture](./architecture.md) - Core concepts and design
- [Audio](./audio.md) - Music and sound effects service
- [Testing](./testing.md) - Simulation backend and test patterns
- [Widgets Overview](./widgets/overview.md) - Available widgets
- [Keybindings](./keybindings.md) - Keyboard routing system
