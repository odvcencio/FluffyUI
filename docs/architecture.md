# Architecture

FluffyUI is built around a small runtime loop, a widget tree, and reactive
state. The core goal is predictable rendering with low overhead for terminal
UIs.

## Core concepts

- Runtime app: owns the backend, message loop, and render pipeline.
- Widget tree: a hierarchy of widgets that Measure, Layout, and Render.
- State signals: `state.Signal` and `state.Computed` drive reactive updates.
- Commands: widgets emit commands to request app-level actions.

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

## Widgets

Widgets implement:

- `Measure(constraints)`
- `Layout(bounds)`
- `Render(ctx)`
- `HandleMessage(msg)`

Containers implement `ChildWidgets()` to expose their children for traversal.

## Accessibility

The screen uses an announcer and focus styles from the app configuration.
Accessible widgets expose role, label, and state via `accessibility.Accessible`.

## Performance

- Buffered rendering with dirty cell tracking.
- Minimal allocations in render hot paths.
- Simulation backend for deterministic tests.

See `docs/performance.md` for tuning tips.
