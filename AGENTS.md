# AGENTS.md - Contribution Guide for FluffyUI

This document provides essential context for AI agents and contributors working on FluffyUI, a Go-based terminal UI framework with sub-cell graphics, reactive state, and accessibility support.

## Project Overview

FluffyUI is a comprehensive terminal UI framework featuring:
- **28+ packages** organized by responsibility
- **35+ widgets** with consistent patterns
- **Sub-cell graphics** using Braille/Sextant/Quadrant rendering
- **Reactive state** via signals and computed values
- **Three backends**: tcell (terminal), ghostty (GPU), sim (testing)
- **Built-in accessibility** with screen reader support
- **Agent integration** via MCP protocol

**Go Version:** 1.25.1+

---

## Directory Structure

```
fluffy-ui/
├── runtime/          # Core app loop, render pipeline, message handling
├── widgets/          # 35+ reusable UI components
├── state/            # Reactive signals and computed values
├── animation/        # Tweens, springs, particle systems
├── graphics/         # Sub-cell canvas with shapes and curves
├── backend/          # Backend abstractions
│   ├── tcell/        # Real terminal (tcell v2)
│   ├── ghostty/      # GPU surface (libghostty)
│   └── sim/          # Simulation for testing
├── keybind/          # Keyboard routing and command registry
├── forms/            # Form validation and coordination
├── accessibility/    # Screen reader support, focus management
├── agent/            # Out-of-process agent server with MCP
├── recording/        # Asciicast and video export
├── audio/            # Music and SFX service
├── compositor/       # Screen buffer and ANSI rendering
├── theme/            # Theme management
├── style/            # Style system (colors, attributes)
├── toast/            # Toast notifications
├── scroll/           # Virtual scrolling utilities
├── effects/          # Visual effects (gradients, glow)
├── dragdrop/         # Drag-and-drop interfaces
├── terminal/         # Terminal abstraction (keys, mouse)
├── clipboard/        # Clipboard abstraction
├── markdown/         # Markdown parsing and highlighting
├── progress/         # Progress tracking
├── testing/          # Test helpers
├── examples/         # 19+ example applications
├── docs/             # Comprehensive documentation
└── scripts/          # Recording and agent tools
```

---

## Architecture

### Message-Loop Architecture

```
Event Loop (tick-based, 30 FPS)
    │
    ├── Backend Input (keys, mouse, resize)
    ├── Timers
    └── Custom Events (postings)
    │
    ▼
Message Loop
    │
    ▼
handleMessage (root widget)
    │ - Routes input to focused widget
    │ - Collects commands (bubbling)
    │
    ▼
Render Pipeline
    1. Measure (top-down constraints)
    2. Layout (position assignment)
    3. Render (to buffer)
    4. Diff (dirty cell tracking)
    5. Show (to backend)
```

### Key Interfaces

```go
// Core widget interface (runtime/widget.go)
type Widget interface {
    Measure(constraints Constraints) Size
    Layout(bounds Rect)
    Render(ctx RenderContext)
    HandleMessage(msg Message) HandleResult
}

// Focusable widgets
type Focusable interface {
    IsFocused() bool
    SetFocused(bool)
    CanFocus() bool
}

// Reactive binding
type Bindable interface {
    Bind(services Services)
}
type Unbindable interface {
    Unbind()
}

// Container traversal
type ChildProvider interface {
    ChildWidgets() []Widget
}
```

### Constraint-Based Layout

Three-phase rendering:
1. **Measure**: Parent provides Constraints, widget returns preferred Size
2. **Layout**: Parent assigns Rect bounds to widget
3. **Render**: Widget draws to RenderContext.Buffer

```go
type Constraints struct {
    MinWidth, MaxWidth, MinHeight, MaxHeight int
}
type Size struct { Width, Height int }
type Rect struct { X, Y, Width, Height int }
```

---

## Coding Conventions

### Naming

| Element | Convention | Example |
|---------|------------|---------|
| Packages | lowercase, single word | `widgets`, `state`, `keybind` |
| Types | PascalCase | `Button`, `Signal`, `GridLayout` |
| Option functions | `With*` prefix | `WithVariant`, `WithDisabled` |
| Constructors | `New*` prefix | `NewButton`, `NewSignal` |
| Accessors | `Get*/Set*` | `GetValue`, `SetLabel` |
| Predicates | `Is*/Can*` | `IsFocused`, `CanFocus` |
| Interface suffixes | `-able`, `-Provider` | `Focusable`, `BoundsProvider` |

### File Organization

```go
package widgets

// 1. Imports (stdlib, then third-party, then local)
import (
    "fmt"

    "github.com/external/pkg"

    "fluffy-ui/runtime"
)

// 2. Type definition
type Button struct {
    widgets.FocusableBase
    label *state.Signal[string]
}

// 3. Constructor
func NewButton(label string, opts ...ButtonOption) *Button { ... }

// 4. Option type and functions
type ButtonOption func(*Button)
func WithVariant(v Variant) ButtonOption { ... }

// 5. Interface implementations (Measure, Layout, Render, etc.)
func (b *Button) Measure(constraints runtime.Constraints) runtime.Size { ... }
func (b *Button) Render(ctx runtime.RenderContext) { ... }

// 6. Other methods
func (b *Button) SetLabel(label string) { ... }
```

### Interface Compliance

Always verify interface compliance at compile time:

```go
var _ runtime.Widget = (*MyWidget)(nil)
var _ runtime.Focusable = (*MyWidget)(nil)
```

### Nil Receiver Checks

Many methods include nil receiver guards for safety:

```go
func (b *Base) Layout(bounds runtime.Rect) {
    if b == nil {
        return
    }
    // ...
}
```

### Error Handling

- Return `error` as last value
- Use `fmt.Errorf` for formatted errors with context
- Use `errors.Is/As` for error checking
- Early returns for error cases

---

## Widget Development

### Minimal Widget

```go
type MyWidget struct {
    widgets.Base  // or FocusableBase for keyboard input
}

func NewMyWidget() *MyWidget {
    return &MyWidget{}
}

func (w *MyWidget) Measure(constraints runtime.Constraints) runtime.Size {
    return runtime.Size{Width: 10, Height: 1}
}

func (w *MyWidget) Layout(bounds runtime.Rect) {
    w.Base.Layout(bounds)
}

func (w *MyWidget) Render(ctx runtime.RenderContext) {
    bounds := w.Bounds()
    ctx.Buffer.SetString(bounds.X, bounds.Y, "Hello", style.Default)
}

func (w *MyWidget) HandleMessage(msg runtime.Message) runtime.HandleResult {
    return runtime.Unhandled()
}
```

### Reactive Widget with State

```go
type Counter struct {
    widgets.Component  // Includes Base + subscriptions
    count *state.Signal[int]
}

func NewCounter() *Counter {
    return &Counter{
        count: state.NewSignal(0),
    }
}

func (c *Counter) Bind(services runtime.Services) {
    c.Component.Bind(services)

    // Subscribe to state changes
    c.Observe(c.count, func() {
        c.services.Invalidate()  // Trigger re-render
    })
}

func (c *Counter) Render(ctx runtime.RenderContext) {
    text := fmt.Sprintf("Count: %d", c.count.Get())
    ctx.Buffer.SetString(c.Bounds().X, c.Bounds().Y, text, style.Default)
}
```

### Widget with Options Pattern

```go
type Button struct {
    widgets.FocusableBase
    label   string
    variant Variant
}

type ButtonOption func(*Button)

func WithVariant(v Variant) ButtonOption {
    return func(b *Button) {
        b.variant = v
    }
}

func WithDisabled(disabled *state.Signal[bool]) ButtonOption {
    return func(b *Button) {
        b.disabled = disabled
    }
}

func NewButton(label string, opts ...ButtonOption) *Button {
    b := &Button{label: label, variant: VariantDefault}
    for _, opt := range opts {
        opt(b)
    }
    return b
}
```

### Adding Accessibility

Every widget should implement accessibility:

```go
func (w *MyWidget) AccessibleRole() accessibility.Role {
    return accessibility.RoleButton
}

func (w *MyWidget) AccessibleName() string {
    return w.label
}

func (w *MyWidget) AccessibleState() accessibility.State {
    return accessibility.State{
        Focused:  w.IsFocused(),
        Disabled: w.disabled.Get(),
    }
}
```

### Base Types Reference

| Type | Use Case |
|------|----------|
| `widgets.Base` | Non-focusable widgets (labels, containers) |
| `widgets.FocusableBase` | Keyboard-interactive widgets (buttons, inputs) |
| `widgets.Component` | Reactive widgets with signal subscriptions |

---

## State Management

### Signals

```go
// Create
count := state.NewSignal(0)

// Read
value := count.Get()

// Update
count.Set(5)
count.Update(func(v int) int { return v + 1 })

// Subscribe (in widget Bind method)
c.Observe(count, func() {
    c.services.Invalidate()
})
```

### Computed Values

```go
firstName := state.NewSignal("John")
lastName := state.NewSignal("Doe")

fullName := state.Computed(func() string {
    return firstName.Get() + " " + lastName.Get()
})

// Automatically updates when firstName or lastName changes
```

---

## Testing

### Using Simulation Backend

```go
func TestMyWidget(t *testing.T) {
    be := sim.New(80, 24)
    be.Init()
    defer be.Fini()

    app := runtime.NewApp(runtime.AppConfig{Backend: be})
    app.SetRoot(NewMyWidget())

    ctx, cancel := context.WithCancel(context.Background())
    cancel()  // Run one frame
    app.Run(ctx)

    if !be.ContainsText("expected") {
        t.Error("expected text not found")
    }
}
```

### Using Test Helpers

```go
import "fluffy-ui/testing"

func TestLabel(t *testing.T) {
    label := widgets.NewLabel("Hello")
    output := testing.RenderToString(label, 40, 1)

    if !strings.Contains(output, "Hello") {
        t.Error("label not rendered")
    }
}
```

### Available Test Utilities

| Function | Purpose |
|----------|---------|
| `RenderToString(widget, w, h)` | Render widget to string |
| `RenderWidget(widget, w, h)` | Render to sim backend |
| `AssertContains(t, be, text)` | Assert text present |
| `AssertNotContains(t, be, text)` | Assert text absent |
| `AssertTextAt(t, be, x, y, text)` | Assert text at position |

### Input Simulation

```go
be := sim.New(80, 24)
be.InjectKey(terminal.KeyEnter, 0, 0)  // Simulate Enter key
be.InjectMouse(10, 5, terminal.ButtonLeft)  // Simulate click
```

---

## Commit Message Format

```
type(scope): message
```

### Types

| Type | Use |
|------|-----|
| `add:` | New feature |
| `update:` | Enhancement to existing feature |
| `fix:` | Bug fix |
| `refactor:` | Code restructuring |
| `chore:` | Build, CI, dependencies |
| `improve:` | Performance or quality |
| `docs:` | Documentation only |

### Examples

```
add: feat(widgets): new DatePicker component
update(animation): add bounce easing function
fix(keybind): correct modifier key detection on Linux
refactor(runtime): simplify render pipeline
chore: update tcell dependency to v2.14
```

---

## Key Patterns

### DO

- **Embed base types** (`Base`, `FocusableBase`, `Component`)
- **Use option functions** for widget configuration
- **Implement accessibility** (role, name, state)
- **Add nil receiver checks** in methods
- **Use signals** for reactive state
- **Request invalidation** via `services.Invalidate()`, not direct render
- **Write tests** using simulation backend
- **Verify interface compliance** with compile-time checks

### DON'T

- **Don't call render directly** - use Invalidate/Relayout
- **Don't store RenderContext** - it's only valid during Render call
- **Don't modify Constraints** - they're read-only input
- **Don't forget Unbind** - clean up subscriptions
- **Don't use blocking operations** in Render or HandleMessage
- **Don't allocate in hot paths** - reuse buffers where possible

---

## Important Files to Understand

### Must-Read

1. `runtime/widget.go` - Core Widget interface and types
2. `runtime/app.go` - Application loop and orchestration
3. `widgets/base.go` - Base widget implementations
4. `state/signal.go` - Reactive state primitives
5. `docs/architecture.md` - Design overview

### Reference Implementations

1. `widgets/button.go` - Simple focusable widget
2. `widgets/label.go` - Simple non-focusable widget
3. `widgets/grid.go` - Container with child layout
4. `widgets/input.go` - Complex input handling
5. `widgets/dialog.go` - Modal overlay pattern

### Examples

1. `examples/counter/` - Minimal reactive example
2. `examples/quickstart/` - Basic setup
3. `examples/candy-wars/` - Complex real application
4. `examples/graphics-demo/` - Canvas and animation

---

## Environment Variables

| Variable | Purpose |
|----------|---------|
| `FLUFFYUI_BACKEND` | Backend selection: `tcell`, `ghostty`, `sim` |
| `FLUFFYUI_RECORD` | Recording output file (`.cast`) |
| `FLUFFYUI_RECORD_EXPORT` | Video export file (`.mp4`) |
| `FLUFFYUI_AUDIO_ASSETS` | Audio files path or `off` |
| `FLUFFYUI_AGENT` | Agent socket (e.g., `unix:/tmp/sock`) |
| `FLUFFYUI_AGENT_TOKEN` | Agent authentication token |
| `FLUFFYUI_AGENT_ALLOW_TEXT` | Allow raw text in snapshots |

---

## Running and Building

```bash
# Run an example
go run ./examples/quickstart

# Run tests
go test ./...

# Run specific package tests
go test ./widgets/...

# Run with race detector
go test -race ./...

# Build all
go build ./...
```

---

## Workflow for Adding a Widget

1. **Define struct** embedding appropriate base type
2. **Implement Widget interface** (Measure, Layout, Render, HandleMessage)
3. **Add Bind/Unbind** if using reactive state
4. **Implement accessibility** (AccessibleRole, AccessibleName, AccessibleState)
5. **Create constructor** with options pattern
6. **Write tests** in `*_test.go` file
7. **Add documentation** in `docs/widgets/`
8. **Add example** demonstrating usage

---

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/gdamore/tcell/v2` | Terminal rendering |
| `github.com/alecthomas/chroma/v2` | Syntax highlighting |
| `github.com/mark3labs/mcp-go` | Model Context Protocol |
| `github.com/mattn/go-runewidth` | Unicode width |
| `github.com/yuin/goldmark` | Markdown parsing |
| `golang.org/x/term` | Terminal mode control |

---

## Getting Help

- **Documentation**: `docs/` directory
- **Examples**: `examples/` directory
- **Architecture**: `docs/architecture.md`
- **Widget Guide**: `docs/widgets/overview.md`
- **Testing Guide**: `docs/testing.md`
