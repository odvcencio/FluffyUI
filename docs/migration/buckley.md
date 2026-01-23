# Using FluffyUI Without runtime.App

This guide documents patterns for applications that need custom event loops
or want to use FluffyUI widgets without the standard `runtime.App`.

This is the approach used by Buckley and is appropriate for applications with:
- Streaming APIs (LLM responses, real-time data)
- Complex concurrency requirements
- Custom timing or frame rate control
- Integration with existing event loops

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Your Application                         │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ Event Loop  │  │ Message     │  │ Domain Logic        │ │
│  │ (Custom)    │  │ Channel     │  │ (API, Business)     │ │
│  └──────┬──────┘  └──────┬──────┘  └──────────┬──────────┘ │
│         │                │                     │            │
│         ▼                ▼                     ▼            │
│  ┌─────────────────────────────────────────────────────────┐│
│  │              FluffyUI Integration Layer                 ││
│  │  - runtime.Screen for widget tree & focus              ││
│  │  - Custom widgets extending widgets.Base               ││
│  │  - backend.Backend for terminal I/O                    ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Backend (Terminal I/O)

```go
import "github.com/odvcencio/fluffy-ui/backend/tcell"

be, err := tcell.New()
if err != nil {
    return err
}
if err := be.Init(); err != nil {
    return err
}
defer be.Fini()

w, h := be.Size()
```

### 2. Screen (Widget Tree Management)

```go
import "github.com/odvcencio/fluffy-ui/runtime"

screen := runtime.NewScreen(w, h)
screen.SetAutoRegisterFocus(true)  // Auto-register focusable widgets

// Set root widget
screen.SetRoot(mainWidget)

// Push modal overlays
screen.PushLayer(dialog, true)  // true = modal

// Pop overlays
screen.PopLayer()
```

### 3. Custom Event Loop

```go
type App struct {
    backend   *tcell.Backend
    screen    *runtime.Screen
    messages  chan Message
    running   bool
    dirty     bool
}

func (a *App) Run() error {
    // Start event polling
    go a.pollEvents()

    // Main loop at ~60 FPS
    ticker := time.NewTicker(16 * time.Millisecond)
    defer ticker.Stop()

    a.running = true
    for a.running {
        select {
        case msg := <-a.messages:
            if a.update(msg) {
                a.dirty = true
            }
        case <-ticker.C:
            if a.dirty {
                a.render()
                a.dirty = false
            }
        }
    }

    return nil
}
```

### 4. Event Polling

```go
import "github.com/odvcencio/fluffy-ui/terminal"

func (a *App) pollEvents() {
    for a.running {
        ev := a.backend.PollEvent()
        switch e := ev.(type) {
        case terminal.KeyEvent:
            a.messages <- KeyMsg{Key: e.Key, Rune: e.Rune, Mod: e.Mod}
        case terminal.MouseEvent:
            a.messages <- MouseMsg{X: e.X, Y: e.Y, Button: e.Button}
        case terminal.ResizeEvent:
            a.messages <- ResizeMsg{Width: e.Width, Height: e.Height}
        case terminal.PasteEvent:
            a.messages <- PasteMsg{Text: e.Text}
        }
    }
}
```

### 5. Message Handling

```go
func (a *App) update(msg Message) bool {
    switch m := msg.(type) {
    case KeyMsg:
        return a.handleKey(m)
    case ResizeMsg:
        a.screen.SetSize(m.Width, m.Height)
        return true
    case StreamChunkMsg:
        // Handle streaming data
        a.appendContent(m.Content)
        return true
    }
    return false
}

func (a *App) handleKey(msg KeyMsg) bool {
    // Convert to runtime message
    runtimeMsg := runtime.KeyMsg{
        Key:  msg.Key,
        Rune: msg.Rune,
        Mod:  msg.Mod,
    }

    // Dispatch to screen (routes to focused widget)
    result := a.screen.HandleMessage(runtimeMsg)

    // Handle any commands
    for _, cmd := range result.Commands {
        a.handleCommand(cmd)
    }

    return result.Handled
}
```

### 6. Rendering

```go
func (a *App) render() {
    // Have screen render widget tree
    a.screen.Render()
    buf := a.screen.Buffer()

    // Copy to backend
    w, h := a.backend.Size()
    for y := 0; y < h; y++ {
        for x := 0; x < w; x++ {
            cell := buf.Get(x, y)
            r := cell.Rune
            if r == 0 {
                r = ' '
            }
            a.backend.SetContent(x, y, r, nil, cell.Style)
        }
    }

    a.backend.Show()
}
```

## Building Custom Widgets

### Extend Base Types

Always embed `widgets.Base` or `widgets.FocusableBase`:

```go
import "github.com/odvcencio/fluffy-ui/widgets"

type MyWidget struct {
    widgets.Base  // Non-focusable
    // or
    widgets.FocusableBase  // Focusable
}
```

### Implement Widget Interface

```go
func (w *MyWidget) Measure(c runtime.Constraints) runtime.Size {
    // Return desired size within constraints
    return runtime.Size{
        Width:  min(w.desiredWidth, c.MaxWidth),
        Height: min(w.desiredHeight, c.MaxHeight),
    }
}

func (w *MyWidget) Layout(bounds runtime.Rect) {
    w.Base.Layout(bounds)  // Store bounds
    // Layout children if any
}

func (w *MyWidget) Render(ctx runtime.RenderContext) {
    // Draw to ctx.Buffer within ctx.Bounds
    for i, r := range w.text {
        x := ctx.Bounds.X + i
        if x < ctx.Bounds.X+ctx.Bounds.Width {
            ctx.Buffer.Set(x, ctx.Bounds.Y, r, w.style)
        }
    }
}

func (w *MyWidget) HandleMessage(msg runtime.Message) runtime.HandleResult {
    switch m := msg.(type) {
    case runtime.KeyMsg:
        if m.Rune == 'x' {
            w.doSomething()
            return runtime.Handled()
        }
    }
    return runtime.Unhandled()
}
```

### Container Widgets

Implement `ChildWidgets()` for containers:

```go
type Container struct {
    widgets.Base
    children []runtime.Widget
}

func (c *Container) ChildWidgets() []runtime.Widget {
    return c.children
}

func (c *Container) Layout(bounds runtime.Rect) {
    c.Base.Layout(bounds)
    // Layout children
    y := bounds.Y
    for _, child := range c.children {
        size := child.Measure(runtime.Constraints{
            MaxWidth:  bounds.Width,
            MaxHeight: bounds.Height - (y - bounds.Y),
        })
        child.Layout(runtime.Rect{
            X: bounds.X, Y: y,
            Width: bounds.Width, Height: size.Height,
        })
        y += size.Height
    }
}

func (c *Container) Render(ctx runtime.RenderContext) {
    for _, child := range c.children {
        child.Render(ctx)
    }
}
```

## Using FluffyUI Layouts

Leverage built-in flex containers:

```go
import "github.com/odvcencio/fluffy-ui/runtime"

// Vertical stack
root := runtime.VBox(
    runtime.Fixed(header),      // Fixed height
    runtime.Expanded(content),  // Fill remaining
    runtime.Fixed(footer),      // Fixed height
)

// Horizontal stack
row := runtime.HBox(
    runtime.Flexible(sidebar, 1),  // 25% of space
    runtime.Flexible(main, 3),     // 75% of space
)

// Sized widget
sized := runtime.Sized(widget, 40)  // Fixed 40 width
```

## Focus Management

### Automatic Registration

```go
screen.SetAutoRegisterFocus(true)
```

### Manual Registration

```go
scope := screen.FocusScope()
scope.Register(myFocusableWidget)
scope.SetFocus(myFocusableWidget)
```

### Focus Navigation

```go
scope.FocusNext()
scope.FocusPrev()
scope.FocusFirst()
```

## Handling Streaming Data

For LLM responses or real-time data:

```go
// Start streaming in background
go func() {
    stream := api.StreamCompletion(ctx, req)
    for chunk := range stream {
        a.Post(StreamChunkMsg{Content: chunk.Content})
    }
    a.Post(StreamCompleteMsg{})
}()

// Handle in update loop
func (a *App) update(msg Message) bool {
    switch m := msg.(type) {
    case StreamChunkMsg:
        a.chatView.AppendContent(m.Content)
        return true
    case StreamCompleteMsg:
        a.chatView.FinishMessage()
        return true
    }
    return false
}
```

## Style Caching

For performance, cache styles:

```go
type StyleCache struct {
    cache map[theme.Color]backend.Style
}

func (c *StyleCache) Get(color theme.Color) backend.Style {
    if style, ok := c.cache[color]; ok {
        return style
    }
    style := backend.DefaultStyle().Foreground(backend.Color(color))
    c.cache[color] = style
    return style
}
```

## Testing Custom Widgets

Use the simulation backend:

```go
func TestMyWidget(t *testing.T) {
    be := sim.New(80, 24)
    be.Init()
    defer be.Fini()

    w := NewMyWidget()
    renderWidget(t, be, w, 80, 24)

    if !be.ContainsText("expected") {
        t.Error("expected text not found")
    }
}

func renderWidget(t *testing.T, be *sim.Backend, w runtime.Widget, width, height int) {
    buf := runtime.NewBuffer(width, height)

    w.Measure(runtime.Constraints{MaxWidth: width, MaxHeight: height})
    w.Layout(runtime.Rect{X: 0, Y: 0, Width: width, Height: height})
    w.Render(runtime.RenderContext{Buffer: buf})

    for y := 0; y < height; y++ {
        for x := 0; x < width; x++ {
            cell := buf.Get(x, y)
            r := cell.Rune
            if r == 0 {
                r = ' '
            }
            be.SetContent(x, y, r, nil, cell.Style)
        }
    }
    be.Show()
}
```

## Complete Example

See `examples/custom-loop` for a working example of this pattern.

## When to Use This Pattern

**Use custom event loop when:**
- Processing streaming API responses
- Integrating with existing event systems
- Need precise timing control
- Building complex multi-threaded applications

**Use runtime.App when:**
- Building standard interactive TUIs
- Don't need custom concurrency
- Want simpler code with less boilerplate
- Starting a new project

## Migration from runtime.App

If you're migrating from `runtime.App`:

1. Replace `app.Post()` with your message channel
2. Replace `AppConfig.Update` with your update function
3. Replace `AppConfig.CommandHandler` with your command handler
4. Implement your own render loop with dirty tracking
5. Use `screen.HandleMessage()` instead of relying on automatic dispatch

The widget code remains the same - only the orchestration layer changes.
