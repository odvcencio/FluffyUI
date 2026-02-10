# FluffyUI

**The most comprehensive TUI framework for Go.**

88 widgets. CSS-like styling. Built-in accessibility. AI-ready.

[![Go Reference](https://pkg.go.dev/badge/github.com/odvcencio/fluffyui.svg)](https://pkg.go.dev/github.com/odvcencio/fluffyui)
[![CI](https://github.com/odvcencio/fluffyui/actions/workflows/ci.yml/badge.svg)](https://github.com/odvcencio/fluffyui/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/odvcencio/fluffyui)](https://goreportcard.com/report/github.com/odvcencio/fluffyui)

<p align="center">
  <img src="docs/demos/hero.gif" alt="FluffyUI Demo" width="640">
</p>

## How does FluffyUI compare?

| Feature | FluffyUI | Bubble Tea | Textual | Ratatui |
|---------|----------|------------|---------|---------|
| Language | Go | Go | Python | Rust |
| Widgets | 88 | ~15 | ~40 | ~15 |
| CSS-like Styling | Yes (FSS) | No | Yes | No |
| Layout Engine | Flexbox + Grid | Manual | CSS Grid | Manual |
| Accessibility | ARIA + TTS | No | No | No |
| AI Agent (MCP) | 145+ tools | No | No | No |
| GPU Rendering | 4 backends | No | No | No |
| Reactive State | Signal[T] | Elm (TEA) | Reactive | Manual |
| Web Backend | WebSocket | No | textual-web | No |
| Hot Reload | Yes | No | Yes | No |

## Quick Start

```go
package main

import "github.com/odvcencio/fluffyui/fluffy"

func main() {
    app, _ := fluffy.NewApp(
        fluffy.WithRoot(fluffy.VStack(
            fluffy.Label("Hello, FluffyUI!"),
            fluffy.PrimaryButton("Click me", func() {
                // Your logic here
            }),
        )),
    )
    app.Run()
}
```

## Installation

```bash
go get github.com/odvcencio/fluffyui@latest
```

**Requirements:** Go 1.24 or later.

Scaffold a new project:

```bash
go install github.com/odvcencio/fluffyui/cmd/fluffy@latest
fluffy create my-app --template full
cd my-app && go run .
```

## Widget Catalog

88 widgets organized into six categories:

| Category | Count | Highlights |
|----------|-------|------------|
| **Input** | 18 | Button, Input, TextArea, Select, Slider, DatePicker, TimePicker, ColorPicker, TagInput, Rating |
| **Data** | 14 | Table, DataGrid, List, Tree, DirectoryTree, Log, RichText, MarkdownViewer, Heatmap, Image |
| **Layout** | 12 | Grid, Flex, Stack, Splitter, ScrollView, Panel, AspectRatio, Drawer, FocusTrap, Sidebar |
| **Navigation** | 12 | Tabs, Menu, Breadcrumb, Stepper, CommandPalette, Accordion, Pagination, Link, SkipNav |
| **Feedback** | 16 | Dialog, Alert, ToastStack, Spinner, Progress, Sparkline, BarChart, LineChart, Badge, Skeleton |
| **Utility** | 16 | Inspector, DebugOverlay, PerformanceDashboard, VideoPlayer, Canvas, GPUCanvas, Tooltip, Popover |

Full catalog: [`docs/widgets/overview.md`](docs/widgets/overview.md)

## Styling with FSS

FluffyUI's stylesheet language (FSS) brings CSS-like styling to the terminal with hot reload:

```css
Button {
    background: #3b82f6;
    color: white;
    padding: 0 2;
    border: rounded;
}

Button:hover {
    background: #2563eb;
}

Button.danger {
    background: #ef4444;
}
```

```go
app, _ := fluffy.NewApp(
    fluffy.WithStylesheet("styles.fss"),
)
```

Supports selectors, pseudo-classes, transitions, min/max sizing, flexbox properties, and live reload during development with `fluffy dev`.

## Reactive State

Type-safe signals with automatic UI updates:

```go
count := state.NewSignal(0)

// Derived values recompute automatically
doubled := state.NewComputed(func() int {
    return count.Get() * 2
})

// Widgets observe signals and re-render on change
count.Set(42) // UI updates automatically
```

## Accessibility

ARIA roles, screen reader announcements, and TTS built into every widget:

```go
// Every widget implements accessibility.Accessible
widget.AccessibleRole()   // "button", "textbox", "grid", ...
widget.AccessibleName()   // Human-readable label
widget.AccessibleState()  // Focused, selected, expanded, ...
```

- 40+ ARIA roles (WAI-ARIA 1.2/1.3)
- 14 extended ARIA properties
- TTS backends: espeak, SSIP (Linux), say (macOS), SAPI (Windows)
- Focus management with visible indicators
- Screen reader announcements across 67+ widget interaction points

## AI Agent Integration (MCP)

FluffyUI exposes 145+ MCP tools so AI agents can read, navigate, and control your TUI:

```go
import "github.com/odvcencio/fluffyui/agent"

srv, _ := agent.NewServer(agent.ServerOptions{
    Addr: "unix:/tmp/fluffyui.sock",
    App:  app,
})
go srv.Serve(ctx)
```

Three transports (stdio, SSE, Unix socket), 52 read tools, 56+ write tools, batch execution, and a built-in agent cookbook. See the [AI Agent Tutorial](docs/tutorials/04-ai-agent-integration.md).

## Graphics and Animation

<p align="center">
  <img src="docs/demos/graphics.gif" alt="Canvas Graphics" width="640">
</p>

- **Canvas API** -- sub-cell drawing with Braille, Sextant, Quadrant, and ASCII blitters
- **GPU rendering** -- Metal, OpenGL, WebGL, and Software backends
- **Animation** -- tweens (10 easing functions), physics springs, particle systems
- **Transforms** -- translate, rotate, scale with save/restore

## More Features

- **Dev Loop** -- hot-reload with `fluffy dev -- go run .`
- **Web Backend** -- serve your TUI over WebSocket with auth and TLS
- **Keybindings** -- registry-based commands with modes and stacking
- **Recording** -- capture sessions as asciicast, export to video
- **Testing** -- simulation backend for deterministic, headless tests
- **Clipboard** -- OSC-52, platform tools, memory fallback (auto-detected)
- **i18n** -- localization bundles and runtime wiring
- **Plugins** -- lightweight registry for third-party widgets

## Examples

```bash
go run ./examples/candy-wars       # Showcase trading game
go run ./examples/todo-app         # Full CRUD app
go run ./examples/dashboard        # Data visualization
go run ./examples/fireworks-demo   # 3D particle effects
go run ./examples/widget-gallery   # Complete widget sampler
```

<p align="center">
  <img src="docs/demos/candy-wars-playback.gif" alt="Candy Wars Demo" width="640">
</p>

38 example apps in `examples/` covering every feature area. See [`docs/EXAMPLES.md`](docs/EXAMPLES.md) for the full list.

## Documentation

| Guide | Description |
|-------|-------------|
| [Getting Started](docs/getting-started.md) | Installation and first app |
| [Architecture](docs/architecture.md) | Core concepts and render pipeline |
| [Widget Catalog](docs/widgets/overview.md) | All 88 widgets with usage |
| [Styling (FSS)](docs/theming.md) | Stylesheets, themes, transitions |
| [Accessibility](docs/accessibility.md) | Roles, announcements, TTS |
| [Agent Integration](docs/tutorials/04-ai-agent-integration.md) | MCP tools and transports |
| [Testing](docs/testing.md) | Simulation backend and helpers |
| [Performance](docs/performance.md) | Profiling and optimization |
| [Bubble Tea Migration](docs/migration/bubbletea.md) | Moving from Bubble Tea |
| [API Reference](docs/api/index.md) | Package-level documentation |

## Contributing

Contributions are welcome! See [`CONTRIBUTING.md`](CONTRIBUTING.md) and [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md).

## License

See the repository for license information.
