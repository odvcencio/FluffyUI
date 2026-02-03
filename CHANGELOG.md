# Changelog

## v0.3.1 - 2026-02-03

### Fixes
- Runtime: AutoFocusLast now defers focus until registration completes, ensuring the last registered widget receives focus.

## v0.3.0 - 2026-02-03

### Highlights
- Expanded widget catalog with new data, input, and navigation components.
- Cross-platform accessibility bridges for Linux/macOS/Windows.
- Real-time agent server with event streaming and wait-for helpers.
- Web backend for browser-based TUI sessions via xterm.js.

### Widgets
- AutoComplete: input + suggestion list with provider and selection callbacks.
- DataGrid: per-cell selection, inline editing, and virtualized data sources.
- DirectoryTree: lazy-loading file browser with icons and filters.
- Log: ring-buffered log viewer with auto-scroll, filtering, levels, and io.Writer.
- MaskedInput: mask patterns with placeholders and validators.
- MultiSelect: multi-choice list control.
- DateRangePicker: range calendar bound to start/end inputs.
- TimePicker: keyboard time selection with optional seconds.
- RichText: markdown rendering with scrolling support.
- PerformanceDashboard: live render stats summary.
- Plugin registry for third-party widgets.
- LineChart: completed rendering (smooth/filled series and auto-scale).
- Input: undo/redo history (Ctrl+Z/Ctrl+Y) with grouped edits.
- Input and TextArea: validator hooks aligned with forms validators.
- New widget interfaces: Searchable, Validatable, LazyLoadable, TabularDataSource.

### Runtime & Backend
- Web backend with configurable paths, auth, and TLS options.
- Auto-focus policy (first/last/none) and focus registration modes.
- Render sampler/observer APIs for profiling summaries.
- Stylesheet file watcher helper for live theme reloads.

### State & Forms
- `state.History[T]` for undo/redo with grouping and depth/size limits.
- Fluent form builder (`forms.Builder`, `FieldSpec`) for constructing forms.
- Async validation for fields and forms, dependency revalidation, async submit.

### i18n
- Localization bundle/localizers with formatting and fallback.
- Pluralization rules and RTL/LTR direction helpers.

### Accessibility
- AT-SPI (Linux), NSAccessibility (macOS), and UI Automation (Windows) bridges.
- Accessibility audit helper in `testing`.

### Agent & MCP
- Real-time agent server with UI event streaming and wait-for helpers.
- WebSocket transport support and fluent config builder.
- Expanded MCP tool surface for wait operations.

### Tooling & Testing
- `fluffy create` templates: minimal, full, game, dashboard, form, data-viewer.
- `fluffy theme` CLI: init/check/export/list/install.
- Golden widget rendering tests + update script.
- `testing.TestSync` for deterministic render waits.
- Widget API generator and VS Code extension scaffold.
- New examples: showcase, performance dashboard, agent-enhanced.
