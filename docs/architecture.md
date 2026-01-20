# Architecture

FluffyUI is built around a small runtime loop, a widget tree, and reactive
state. The core goal is predictable rendering with low overhead for terminal
UIs.

## Core concepts

- Runtime app: owns the backend, message loop, and render pipeline.
- Widget tree: a hierarchy of widgets that Measure, Layout, and Render.
- State signals: `state.Signal` and `state.Computed` drive reactive updates.
- Commands: widgets emit commands to request app-level actions.
- App services: shared helpers like announcer, clipboard, scheduler, and audio.

## Render pipeline

Each frame follows a simple flow:

1. Measure: widgets report their preferred size.
2. Layout: parents assign bounds to children.
3. Render: widgets draw into a buffer.
4. Diff: only dirty cells are flushed to the backend.

Dirty tracking happens at the cell level, so large buffers do not need full
repaints when only a small area changes.

## Messages and commands

The runtime loop processes messages:

- Input (keys, mouse, resize) from the backend.
- Tick messages at a configurable rate.
- Custom messages posted by widgets or effects.

Widgets can return commands like `runtime.Quit`, `runtime.FocusNext`, or
`runtime.PushOverlay`. Commands bubble to the app and screen for handling.

## Widget Interface Hierarchy

FluffyUI uses small, composable interfaces. Widgets implement some or all of
these depending on their capabilities.

### Core Interface

```go
// Widget is the fundamental interface all UI components implement.
type Widget interface {
    Measure(constraints Constraints) Size  // Report preferred size
    Layout(bounds Rect)                     // Accept assigned bounds
    Render(ctx RenderContext)               // Draw to buffer
    HandleMessage(msg Message) HandleResult // Process input
}
```

### Optional Interfaces

```
┌─────────────────────────────────────────────────────────────────┐
│                        Widget (required)                        │
│  Measure, Layout, Render, HandleMessage                        │
└─────────────────────────────────────────────────────────────────┘
        │
        ├── Focusable (keyboard input)
        │     CanFocus, Focus, Blur, IsFocused
        │
        ├── BoundsProvider (position query)
        │     Bounds() Rect
        │
        ├── ChildProvider (containers)
        │     ChildWidgets() []Widget
        │
        ├── Invalidatable (render scheduling)
        │     Invalidate, NeedsRender, ClearInvalidation
        │
        ├── Bindable (app services)
        │     Bind(services Services)
        │
        ├── Unbindable (cleanup)
        │     Unbind()
        │
        ├── Lifecycle (mount/unmount)
        │     Mount, Unmount
        │
        └── Accessible (screen readers)
              AccessibleRole, AccessibleLabel, AccessibleState, etc.
```

### Base Types

The `widgets` package provides base types to embed:

| Type | Use Case | Implements |
|------|----------|------------|
| `Base` | Non-focusable widgets | BoundsProvider, Invalidatable, focus stubs |
| `FocusableBase` | Focusable widgets | Base + CanFocus() → true |
| `Component` | Reactive widgets | Base + Services + Subscriptions |

### Interface Summary

| Interface | Purpose | Methods |
|-----------|---------|---------|
| `Widget` | Core rendering | Measure, Layout, Render, HandleMessage |
| `Focusable` | Keyboard focus | CanFocus, Focus, Blur, IsFocused |
| `BoundsProvider` | Position queries | Bounds |
| `ChildProvider` | Tree traversal | ChildWidgets |
| `Invalidatable` | Render scheduling | Invalidate, NeedsRender, ClearInvalidation |
| `Bindable` | Service injection | Bind |
| `Unbindable` | Service cleanup | Unbind |
| `Lifecycle` | Mount/unmount hooks | Mount, Unmount |
| `Accessible` | Screen readers | Role, Label, Description, State, Value |

### Implementation Guidelines

1. **Always embed Base or FocusableBase** - get default implementations
2. **Implement only what you need** - interfaces are optional
3. **Use Component for reactive state** - handles subscriptions automatically
4. **Implement ChildWidgets for containers** - enables tree traversal
5. **Use Bindable for service access** - clipboard, scheduler, announcer

## Go philosophy alignment

FluffyUI keeps its patterns Go-friendly:

- Explicit configuration via structs or functional options; defaults are documented and overrideable.
- Small, behavior-based interfaces; composition over inheritance.
- Errors returned instead of exceptions; panics reserved for programmer errors.
- No hidden global state; app services are passed via `runtime.App` and `Services`.
- Concurrency is explicit and contained in the app loop or effects.

## Accessibility

The screen uses an announcer and focus styles from the app configuration.
Accessible widgets expose role, label, and state via `accessibility.Accessible`.

## Performance

- Buffered rendering with dirty cell tracking.
- Minimal allocations in render hot paths.
- Simulation backend for deterministic tests.

See `docs/performance.md` for tuning tips.
