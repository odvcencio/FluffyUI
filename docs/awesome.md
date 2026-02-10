# Awesome FluffyUI

A curated list of resources, tools, examples, and projects for FluffyUI -- the Go TUI framework with 80+ widgets, GPU rendering, MCP agent integration, and full accessibility support.

## Contents

- [Official Resources](#official-resources)
- [CLI Tools](#cli-tools)
- [Widget Categories](#widget-categories)
- [Examples](#examples)
- [Tutorials](#tutorials)
- [Migration Guides](#migration-guides)
- [Integrations](#integrations)
- [Performance](#performance)
- [Accessibility](#accessibility)

## Official Resources

- [Documentation](./README.md) - Comprehensive guides and API reference
- [Getting Started](./getting-started.md) - Installation and quick start
- [Architecture](./architecture.md) - Core concepts and design
- [Changelog](../CHANGELOG.md) - Release history
- [Contributing](../CONTRIBUTING.md) - How to contribute
- [Examples Guide](./EXAMPLES.md) - Full example listing with run instructions

## CLI Tools

- `fluffy create` - Project scaffolding with 6 templates (minimal, full, game, dashboard, form, data-viewer)
- `fluffy add` - Generate widget or page boilerplate
- `fluffy dev` - Hot-reload development server
- `fluffy theme` - Theme management (init, check, export, list, install)
- `fluffy test` - Test runner with visual mode
- `fluffy doctor` - Environment diagnostics (terminal capabilities, true color, Unicode, accessibility)
- `fluffy record` - Terminal session recording and export

## Widget Categories

### Input (18 widgets)
Button, Checkbox, ColorPicker, Combobox, DatePicker, DateRangePicker, Input, MaskedInput, MultiSelect, NumberInput, Radio, Rating, Select, Slider, TagInput, TextArea, TimePicker, Toggle

### Display (22 widgets)
Alert, AsyncImage, Avatar, Badge, BarChart, Breadcrumb, Chip, Gauge, AnimatedGauge, HeatMap, Image, LineChart, List, Log, MarkdownViewer, Pagination, ProgressBar, RichText, Skeleton, Sparkline, Text, VirtualList

### Navigation (12 widgets)
Accordion, CommandPalette, Disclosure, Inspector, Link, Menu, Palette, Sidebar, SkipNav, Stepper, Tabs

### Layout (15 widgets)
AspectRatio, Card, Dialog, Drawer, Flex, FocusTrap, Grid, Panel, Popover, ScrollView, Section, Splitter, Stack, Tooltip

### Specialized (21 widgets)
AnimatedWidget, CanvasWidget, Component, DataGrid, DebugOverlay, DirectoryTree, FilePicker, Form, GPUCanvasWidget, NotificationCenter, PerformanceDashboard, Search, SegmentedControl, SignalLabel, Spacer, Spinner, StatusBar, Table, Toast/ToastStack, Tree, VideoPlayer

## Examples

### Beginner

| Example | Description | Run Command |
|---------|-------------|-------------|
| Hello World | Simplest FluffyUI app -- displays "Hello from FluffyUI!" with animated effects | `go run ./examples/quickstart` |
| Counter | Reactive state management with signals, increment/decrement, and history tracking | `go run ./examples/counter` |
| Inline Mode | Bounded inline rendering without alternate-screen takeover | `go run ./examples/inline` |

### Intermediate

| Example | Description | Run Command |
|---------|-------------|-------------|
| Widget Gallery | Complete catalog of all 35+ widgets organized by category (layout, input, data, navigation, feedback) | `go run ./examples/widgets/gallery` |
| Showcase Tabs | Curated widget showcase with multiple tabs (overview, inputs, data) | `go run ./examples/showcase` |
| Layout Widgets | Demonstrations of Grid, Stack, Panel, and other layout widgets | `go run ./examples/widgets/layout` |
| Input Widgets | Input form demonstrations with Button, Input, Select, and more | `go run ./examples/widgets/input` |
| Data Widgets | Data display demonstrations with Table, List, and Tree | `go run ./examples/widgets/data` |
| Navigation Widgets | Navigation demonstrations with Tabs, Menu, and Palette | `go run ./examples/widgets/navigation` |
| Feedback Widgets | Feedback widget demonstrations with Dialog, Toast, and Progress | `go run ./examples/widgets/feedback` |
| Canvas Graphics | Sub-cell graphics: shapes, Bezier curves, transforms, multiple blitters (Braille, Sextant, Quadrant) | `go run ./examples/graphics-demo` |
| Animation | Tweening with easing functions, spring physics, animation chaining | `go run ./examples/animation-demo` |
| Water Simulation | Interactive water ripple simulation using the canvas API | `go run ./examples/water-demo` |
| FUR Demo | Rich text and formatting with FUR markup language, styled text, inline images | `go run ./examples/fur-demo` |
| Custom Loop | Manual event handling, custom update logic, and frame rate control | `go run ./examples/custom-loop` |

### Advanced

| Example | Description | Run Command |
|---------|-------------|-------------|
| Fireworks 3D | 3D particle effects with perspective projection, physics simulation, color gradients | `go run ./examples/fireworks-demo` |
| GPU Canvas | Hardware-accelerated canvas rendering with Software, OpenGL, and Metal backends | `go run ./examples/gpu-canvas-demo` |
| Command Palette | Keybinding system: command registry, fuzzy finder, keymap stacking, custom commands | `go run ./examples/command-palette` |
| Virtual Scrolling | Performance demo: 10,000+ item lists with efficient rendering and smooth scrolling | `go run ./examples/virtual-scrolling` |
| Accessibility Demo | Screen reader support, focus management, ARIA-like roles, keyboard navigation | `go run ./examples/accessibility-demo` |
| AI Agent Integration | Out-of-process agent interaction: JSONL socket protocol, agent-driven input, snapshot observation | `go run ./examples/ai-agent-demo` |
| Recording Demo | Session recording and export: asciicast capture, video export, playback | `go run ./examples/recording` |
| Performance Dashboard | Live render stats: FPS, render/flush timing, dirty cell ratios, layer count sampling | `go run ./examples/perf-dashboard` |

### Production

| Example | Description | Run Command |
|---------|-------------|-------------|
| Candy Wars | Complete trading game (~3000 LOC) demonstrating all FluffyUI features: multi-screen tab navigation, tables with sorting, dialogs, sparkline charts, form validation, reactive state, keybindings, toast notifications, save/load | `go run ./examples/candy-wars` |
| Todo App | Full CRUD application with task management, filtering, persistence, and keyboard shortcuts | `go run ./examples/todo-app` |
| File Browser | File system browser with tree navigation, file operations, preview pane, and keyboard shortcuts | `go run ./examples/file-browser` |
| Dashboard | Data visualization dashboard with multiple chart types, real-time updates, sparklines, and layout management | `go run ./examples/dashboard` |
| Settings Form | Complex form with multiple field types, validation, error messages, and submit handling | `go run ./examples/settings-form` |

### Demo Generation

| Tool | Description | Run Command |
|------|-------------|-------------|
| Generate All Demos | Headless demo generation using simulation backend for CI/CD | `go run ./examples/generate-demos --out docs/demos --duration 6` |
| Regenerate All Demos | Convenience script that generates cast files and converts to GIFs | `./scripts/regenerate-demos.sh [duration_seconds]` |

## Tutorials

### Learning Path

1. **Start here**: Run `quickstart` and `counter` to understand app lifecycle and reactive signals
2. **Explore widgets**: Browse `widgets/gallery` and the category-specific examples for feature depth
3. **Study graphics**: Try `graphics-demo` and `animation-demo` for the sub-cell canvas system
4. **Build an app**: Read through `candy-wars` for a complete application architecture reference

### Backend Selection

```bash
# Real terminal (default)
FLUFFYUI_BACKEND=tcell go run ./examples/quickstart

# Simulation backend (headless / CI)
FLUFFYUI_BACKEND=sim go run ./examples/quickstart
```

### Recording Sessions

```bash
# Record to asciicast
FLUFFYUI_RECORD=session.cast go run ./examples/quickstart

# Record and export to video
FLUFFYUI_RECORD=session.cast FLUFFYUI_RECORD_EXPORT=output.mp4 go run ./examples/quickstart
```

### Audio

```bash
# Enable audio (requires assets)
FLUFFYUI_AUDIO_ASSETS=./examples/quickstart/assets/audio go run ./examples/quickstart

# Disable audio
FLUFFYUI_AUDIO_ASSETS=off go run ./examples/quickstart
```

## Migration Guides

- [v0.3.x to v0.4.x](./migration.md) - Migration guide between major versions
- tcell v2 to v3 backend migration shipped in v0.3.2

## Integrations

### MCP Agent Layer
- **145+ tools** (52 read, 56+ write) for AI-driven UI automation
- 3 transports: stdio, SSE, Unix socket
- Mutation tools: `set_widget_value`, `focus_widget`, `send_key`, `batch_execute`
- Agent cookbook available as `fluffy://instructions` resource (~600 lines)
- Real-time event streaming and wait-for helpers
- WebSocket transport support with fluent config builder
- See: `agent/`, `agent/mcp/`

### Web Backend
- Browser-based TUI sessions via xterm.js (~1,350 LOC)
- WebSocket terminal with auth, TLS, and multi-session support
- Configurable paths and connection options
- See: `backend/web/`

### GPU Rendering
- 4 backends: Metal, OpenGL, WebGL, Software
- Widget-level hardware-accelerated canvas via `GPUCanvasWidget`
- High-performance graphics for large canvases
- See: `gpu/`

### Clipboard
- Auto-detection chain: OSC-52 -> platform tools (pbcopy/xclip/wl-copy/PowerShell) -> memory fallback
- `AutoClipboard` integrated automatically in `fluffy.NewApp()`
- See: `clipboard/`

### TTS / Screen Reader
- Cross-platform speech backends: espeak/ssip (Linux), say (macOS), SAPI (Windows/WSL2)
- Enable with `FLUFFYUI_TTS=1` environment variable
- `SimpleAnnouncer` with assertive interrupt and polite debounce (50ms)
- External screen reader CLI: `cmd/fluffy-speak`
- See: `accessibility/tts/`

### i18n / Localization
- Localization bundle/localizers with formatting and fallback
- Pluralization rules and RTL/LTR direction helpers
- See: `i18n/`

### Forms Validation
- 20+ validators: required, min/max length, email, URL, regex, numeric range, and more
- Async validation, dependency revalidation, fluent builder API
- See: `forms/`

## Performance

### Benchmark Suite
- 56 benchmarks across 5 packages (`state`, `style`, `compositor`, `scroll`, `widgets`)
- `make bench` and `make bench-save` targets for CI regression tracking
- Render pipeline benchmarks in `runtime` with `scripts/bench-render-pipeline.sh`

### Optimization Highlights
- **Zero-alloc fast paths**: `runtime/flex.go` buffer reuse, `state/track.go` atomic pointer for Signal.Get()
- **Style resolution**: specificity caching on selectors (3x faster: 2.2ns vs 6.3ns), rule indices (type/id/class maps)
- **Compositor**: `strings.Builder` for ANSI output, `ANSIWriter.Reset()` for frame-to-frame buffer reuse
- **Widget caches**: Input rune cache with dirty flag, TextArea line metadata cache, Dialog body lines cache
- **State system**: `deepEqual` type-switch fast paths for int/string/bool/float; comma-ok assertions for Signal[any]
- **Synchronized output**: CSI ?2026h for flicker-free rendering
- **Virtual scrolling**: efficient rendering for 10,000+ item lists

## Accessibility

### ARIA Compliance
- Full WAI-ARIA 1.2/1.3 compliance audit (v0.4.0)
- 40+ roles including 22 specialized WAI-ARIA roles
- 14 extended ARIA properties (Level, Orientation, PosInSet, SetSize, and 10 string properties)
- States: Pressed (tri-state), Hidden, Busy, Modal, Expanded, Selected, Checked, Disabled
- Relationships: Owns, FlowTo for complex widget hierarchies

### Screen Reader Support
- 67+ announcement sites across all widgets
- `FormatChange()` builds screen-reader descriptions from widget metadata
- Assertive interrupt and polite debounce for announcement priority
- `SimpleAnnouncer` with bounded history (100 entries)

### Platform Bridges
- AT-SPI (Linux) via `bridge_linux.go`
- NSAccessibility (macOS) via `bridge_darwin.go`
- UI Automation (Windows) via `bridge_windows.go`

### TTS Backends
- espeak: direct command-line synthesis (Linux)
- ssip: Speech Dispatcher socket protocol (Linux)
- say: macOS built-in speech
- SAPI: persistent PowerShell bridge with generation counter for stale goroutine prevention (Windows/WSL2)

### Testing
- `testing/a11ytest` package: 18 check functions (`HasRole`, `HasLabel`, `IsFocusable`, `IsDisabled`, `HandlesKey`, etc.)
- Tree-level assertions: `AllFocusableHaveLabels`, `NoDuplicateIDs`
- Accessibility audit helper in `testing` package

### Keyboard Navigation
- Full keyboard navigation across all interactive widgets
- Chord keybindings (e.g., `Ctrl+K Ctrl+C`)
- Tab/Shift+Tab traversal with focus trap support
- `SkipNav` widget for screen-reader skip links
- Auto-focus policies: first, last, none

## Version Highlights

### v0.4.1 (2026-02-09)
Image widget, HeatMap, Form, a11y testing framework, benchmark suite, 20+ new widgets, clipboard package, convenience API

### v0.4.0 (2026-02-08)
Full WAI-ARIA 1.2/1.3 compliance, widget announcement wiring, ValueInfo diffs in MCP pipeline

### v0.3.5 (2026-02-07)
TTS system, ARIA states/relationships, 15 new roles, widget announcements for 12 widgets

### v0.3.4 (2026-02-05)
Inline mode, `fluffy doctor`, Candy Wars showcase tabs, layout diagnostics

### v0.3.0 (2026-02-03)
80+ widget catalog, web backend, cross-platform a11y bridges, real-time agent server, forms/i18n packages
