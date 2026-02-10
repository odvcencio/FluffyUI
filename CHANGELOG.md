# Changelog

## v0.5.0 - 2026-02-09

### Highlights
- **Bidirectional text**: full RTL support for Input and TextArea widgets with cursor position mapping for Arabic, Hebrew, and other RTL scripts.
- **CSS Grid track sizing**: `Fr()`, `Px()`, `AutoTrack()` units with proportional space distribution, replacing the fixed uniform grid.
- **Standalone prompts library**: `prompts.Confirm()`, `prompts.Input()`, `prompts.Choose()` for shell-friendly interactive prompts using inline mode.
- **SSH server**: `fluffy ssh` command serves any FluffyUI app over SSH with Ed25519 key generation and configurable auth.
- **Documentation site**: mdBook configuration with GitHub Pages auto-deployment for all 51 documentation files.
- **Fuzz testing**: 15M+ executions across FSS parser, FUR markup, and Kitty keyboard protocol with zero panics.

### New Packages
- **`i18n/`**: Bidirectional text support with `DetectDirection()`, `BidiReorder()`, `MirrorBrackets()`, and `CursorMap` for logical-to-visual cursor position mapping.
- **`prompts/`**: Standalone interactive prompts (`Confirm`, `Input`, `Choose`) backed by FluffyUI inline mode with full keyboard handling and simulation backend testing.
- **`ssh/`**: SSH server package with per-session command execution, Ed25519 host key generation, and password authentication.

### Widget Enhancements
- **Base**: added `hovered` and `active` state tracking with `SetHovered()`, `IsHovered()`, `SetActive()`, `IsActive()` methods; `StyleState()` now reports all four states (focused, disabled, hovered, active).
- **Grid**: CSS Grid-like track sizing with `Fr()` (fractional), `Px()` (fixed), `AutoTrack()` units; `resolveTracks()` algorithm distributes space proportionally; `TrackGridBuilder` fluent API; `NewTrackGrid()` constructor; backward-compatible with legacy uniform grid.
- **Input**: BiDi text rendering with RTL detection, `BidiReorder()` for display, `CursorMap` for cursor positioning, and right-alignment for RTL content.
- **TextArea**: per-line BiDi reordering with RTL right-alignment.

### Styling
- `:hover` and `:active` pseudo-classes now fully functional end-to-end: FSS parsing, selector matching, and widget state tracking all wired together.
- **ThemeToggle**: convenience API for cycling through stylesheets with `NewThemeToggle()`, `Toggle()`, `SetIndex()`, and `Bind()` for automatic app updates.

### Agent & MCP
- `list_themes` tool: returns available built-in theme names (dark, light, monokai, nord).
- `set_theme` tool: switches the active theme by name at runtime.

### CLI Tools
- `fluffy snapshot`: captures a single frame from a FluffyUI app with configurable format, dimensions, and delay.
- `fluffy ssh`: serves FluffyUI apps over SSH with `--addr`, `--host-key`, `--password`, `--no-auth` flags.
- `fluffy dev`: new `--inspect`, `--debug-log`, and `--profile` flags for injecting debug environment variables into the subprocess.

### Testing
- Fuzz tests for FSS parser (`style/parser_fuzz_test.go`), FUR markup (`fur/markup_fuzz_test.go`), and Kitty keyboard protocol (`terminal/kitty_fuzz_test.go`).
- 24 new Grid track sizing tests covering `resolveTracks()`, `trackOffsets()`, builder API, layout, and backward compatibility.
- 15 prompts package tests covering Confirm, Choose, and Input with simulation backend.
- Hover/active state tests for Base widget and stylesheet selector matching.

### Documentation
- mdBook site: `book.toml`, `SUMMARY.md` mapping all docs, and `docs/README.md` landing page.
- GitHub Actions workflow (`.github/workflows/docs.yml`) for automated Pages deployment on docs changes.
- `docs/awesome.md`: curated resource list covering all 88 widgets, 44 examples, integrations, and version highlights.
- README rewritten: feature comparison table, badges, quickstart code, 68% size reduction (696 to 219 lines).

### Dependencies
- Added `golang.org/x/crypto` v0.48.0 for SSH server support.

## v0.4.1 - 2026-02-09

### Highlights
- New widgets: Image (multi-protocol terminal rendering), HeatMap (2D data visualization), Form (validated input collection).
- A11y testing framework (`testing/a11ytest`) for asserting widget accessibility compliance in unit tests.
- Benchmark suite covering all hot paths: signals, style resolution, compositor, virtual scroll, and widget lifecycles.
- 20+ new widgets and features since v0.4.0 including Combobox, SegmentedControl, Disclosure, StatusBar, NotificationCenter, Inspector, and more.

### New Widgets
- **Image**: terminal image display with Kitty, Sixel, iTerm2, and half-block fallback protocols; fit modes (Contain, Cover, Fill, ScaleDown); `NewImageFromFile` helper.
- **HeatMap**: 2D data grid visualization with configurable color scales (GreenRed, BlueRed, Grayscale); row/column labels; optional numeric value overlay; ARIA grid role with live region announcements.
- **Form**: labeled input fields with built-in validation, Tab/Shift+Tab navigation, Enter submit, Escape cancel; duck-typed `validatable`/`textProvider` interfaces; error announcements; `fluffy.Form()`/`fluffy.FormField()` convenience API.
- **Combobox**: searchable dropdown with keyboard navigation.
- **SegmentedControl**: mutually exclusive option bar.
- **Disclosure**: collapsible content sections.
- **StatusBar**: application status display with sections.
- **NotificationCenter**: stacked notification management.
- **Inspector**: widget tree debugger.
- **Badge**, **Card**, **Link**, **Pagination**, **Rating**, **Skeleton**, **TagInput**, **MarkdownViewer**: additional UI components following standard widget patterns.

### Performance
- Benchmark suite: 56 benchmarks across 5 packages (`state`, `style`, `compositor`, `scroll`, `widgets`).
- `Makefile` with `bench` and `bench-save` targets for CI regression tracking.
- Zero-alloc optimizations: `runtime/flex.go` buffer reuse, `state/track.go` atomic fast path, `style/selector.go` specificity caching (3x faster), `compositor/ansi.go` string builder pooling.
- Widget-level caches: Input rune cache, TextArea line metadata, Dialog body lines.

### Accessibility
- `testing/a11ytest` package: 18 check functions (`HasRole`, `HasLabel`, `IsFocusable`, `IsDisabled`, `HandlesKey`, etc.) plus tree-level assertions (`AllFocusableHaveLabels`, `NoDuplicateIDs`).
- `accessibility.RoleForm` added for Form widget.
- `RoleImage` wired for Image widget with configurable alt text.
- HeatMap: ARIA grid role with live region announcements for data changes.
- 67+ announcement sites across all widgets.

### Styling
- Hot reload: `runtime.WatchStylesheetFile()` for live FSS changes.
- Easing functions for style transitions.
- Themes: `style.Theme` with built-in light/dark, `fluffy theme` CLI.
- Color scheme media queries for automatic dark/light switching.
- Layout properties: min/max width/height, content-align, dock, offset-x/y, z-index.
- Style transitions: color, opacity, and spacing interpolation via `TransitionManager`.

### Runtime & Backend
- Synchronized output (CSI ?2026h) for flicker-free rendering.
- Responsive layouts with breakpoint system.
- Drag-and-drop support.
- Chord keybindings (e.g., `Ctrl+K Ctrl+C`).
- Snapshot testing framework with golden file management.
- `terminal/caps.go`: added `ITermInlineImages` detection for iTerm2 terminals.

### Clipboard
- `clipboard/` package: `OSC52Clipboard`, `PlatformClipboard`, `AutoClipboard` with auto-detection chain (OSC-52 → platform tools → memory fallback).

### Convenience API
- `fluffy.Form()`, `fluffy.FormField()` for quick form creation.
- `fluffy.SelectFromStrings()`, `fluffy.ReactiveText()`, `fluffy.Checkbox()`.
- Layout helpers: `VStack`, `HStack`, `Spacer`, `Divider`, `Center`, `Padding`.
- Text helpers: `Bold`, `Italic`, `DangerText`, `SuccessText`, `WarningText`.

### Agent & MCP
- 145+ tools (52 read, 56+ write) across 3 transports (stdio, SSE, Unix socket).
- Mutation tools: `set_widget_value`, `focus_widget`, `send_key`, `batch_execute`.
- Agent cookbook as `fluffy://instructions` resource (~600 lines).

### Forms
- `forms/` package: 20+ validators (required, min/max length, email, URL, regex, numeric range, etc.).
- Async validation, dependency revalidation, fluent builder API.

## v0.4.0 - 2026-02-08

### Highlights
- Full WAI-ARIA 1.2/1.3 compliance audit: fixed 9 critical/high issues, wired 6 interactive widgets with announcements, and added missing roles.
- Every interactive widget now announces state changes to screen readers via `Bind`/`Unbind` service wiring.
- MCP pipeline tracks `ValueInfo` diffs for rich slider/gauge change detection.

### Accessibility
- New roles: `RoleTooltip`, `RoleMeter`.
- `FormatChange()`: now includes `ErrorMessage` and substitutes `RoleDescription` for raw role when set.
- `Slider`/`RangeSlider`: `ValueInfo` now populates `Min`/`Max`/`Current` (required by WAI-ARIA `aria-valuemin`/`aria-valuemax`/`aria-valuenow`).
- `RangeSlider`: added missing `Orientation` property.
- `Radio`: uses `State.Checked` instead of `State.Selected` per ARIA spec; reports `PosInSet`/`SetSize` for group context.
- `Dialog`: `syncA11y()` now always sets `State.Modal = true`.
- `Log`: role corrected from `RoleList` to `RoleLog`.
- `selectDropdown`: role corrected from `RoleList` to `RoleListbox`.
- `AutoComplete`: fallback role fixed from `RoleTextbox` to `RoleCombobox`; sets `State.Expanded` based on suggestions; announces suggestion count changes.
- `Tooltip`: role changed from `RoleGroup` to `RoleTooltip`.
- `AnimatedGauge`: wired `RoleMeter` with label and `ValueInfo`.

### Widget Announcements
- `Select`: announces selection changes via `AnnounceChange`; sets `Expanded` state, `PosInSet`/`SetSize`.
- `MultiSelect`: added `Bind`/`Unbind`; announces "Option checked"/"Option unchecked" on toggle and focused option on navigation.
- `Slider`/`RangeSlider`: announce value changes (debounced via `lastDesc` pattern to avoid redundant speech).
- `Radio`: announces selection change via `AnnounceChange`; added `Bindable`/`Unbindable` interface assertions.
- `AutoComplete`: added `Bind`/`Unbind`; announces "N suggestions available" on count change.

### Agent & MCP
- `widgetChanges()`: added `ValueInfo` diff via `valueInfoEqual` helper for `ValueInfoMCP` comparison.

### Tools
- `cmd/fluffy-speak`: `formatWidget()` now includes `ErrorMessage`; `widgetChanges()` tracks `ValueInfo` diffs.

## v0.3.5 - 2026-02-07

### Highlights
- Full WAI-ARIA 1.2 compliance: added missing states (`Pressed`, `Hidden`, `Busy`, `Modal`), relationship attributes (`Owns`, `FlowTo`), and 15 new roles.
- TTS system: cross-platform speech backends (espeak/ssip on Linux, say on macOS, SAPI on Windows/WSL2) with `FLUFFYUI_TTS=1` one-liner.
- Widget announcement wiring: Alert, Dialog, Tabs, Menu, ToastStack, Log, Search, Section, Progress, ScrollView, Sparkline, and BarChart now announce state changes to screen readers.

### Accessibility
- `StateSet`: added `Pressed *bool` (tri-state for toggle buttons), `Hidden bool`, `Busy bool`, `Modal bool` with `Strings()` output.
- `Accessible` interface: added `AccessibleOwns()` and `AccessibleFlowTo()` relationship methods with setters on `Base`.
- New ARIA roles: `combobox`, `switch`, `spinbutton`, `heading`, `link`, `separator`, `log`, `timer`, `feed`, `toolbar`, `searchbox`, `none`, `img`, `note`, `scrollbar`.
- `Speaker` interface and `accessibility/tts` package with espeak, ssip (Linux), say (macOS), and SAPI (Windows/WSL2) backends.
- `SimpleAnnouncer`: speech dispatch with assertive interrupt and polite debounce (50ms), assertive-priority protection.
- `FormatChange`: builds screen-reader descriptions from widget metadata.

### Widgets
- Dialog: sets `State.Modal = true` by default.
- Sparkline and BarChart: added `Bind`/`Unbind` for service wiring and change-based announcements.
- Alert, Tabs, Menu, ToastStack, Log, Search, Section, Progress, ScrollView: wired `Bind`/`Unbind` with live region announcements on state changes.

### Agent & MCP
- `WidgetInfo`: added `Owns` and `FlowTo` fields, extracted in `extractWidgetInfo`.
- MCP `StateSet`: changed `Pressed` from `bool` to `*bool`, added `Modal`.
- MCP `WidgetInfo`/`WidgetNode`: added `Owns` and `FlowTo` fields.
- `stateFromAgent`: maps `Pressed`, `Hidden`, `Busy`, `Modal`.
- `roleToMCP`: handles all 15 new roles plus `chart`, `slider`, `table`, `row`, `cell`, `tablist`, `window`, `application`.
- Diff detection: `widgetChanges` tracks `Owns`/`FlowTo`; `stateDiff` tracks `Hidden`/`Pressed`/`Modal`.

### Tools & Examples
- `cmd/fluffy-speak`: external screen reader CLI with snapshot polling, ARIA-driven announcements, and hidden widget filtering.
- `examples/aria-demo`: expanded with relationship, form validation, toggle button, and busy state demos.
- `fluffy.NewApp()`: auto-wires TTS when `FLUFFYUI_TTS=1` is set.

## v0.3.4 - 2026-02-05

### Highlights
- Inline mode (experimental): bounded viewport rendering without alternate-screen takeover via `fluffy.RunInline`, `fluffy.WithInlineMode`, and `fluffy.WithInlineHeight`.
- New `fluffy doctor` command (`--json` supported) for terminal capability diagnostics (TERM, true color, mouse, Unicode width, Kitty/Sixel, accessibility).
- Candy Wars now includes Event Log and Showcase tabs, with a kitchen-sink sampler of 20+ interactive widgets.

### Runtime & Backend
- Runtime app lifecycle hooks added (`OnReady`, `OnResize`, `OnQuit`) with thread-safe state access improvements in app/snapshot paths.
- tcell backend: inline viewport sizing + event remapping and alternate-screen escape suppression for inline mode.
- Web backend: fixed ANSI background style tracking bug, tightened terminal-state concurrency, and corrected SGR mouse CSI parameter parsing.
- Buffer: `SetString` now accounts for rune display width to avoid overlap/clipping with wide characters.
- Layout diagnostics: added `FLUFFYUI_LAYOUT_DEBUG` and `runtime.SetLayoutDebug(true)` warnings for unsatisfiable constraints and zero-size measurements.

### Widgets, Forms & API
- ScrollView: added `ContentSize()` accessor.
- Forms: `NewFieldBase` now returns `*FieldBase`; `SimpleField` received stronger nil-safety guards.
- Fluffy convenience API: added `Label(...)` and `Text(...)` helpers plus optional toast overlay wiring via `fluffy.WithToastLayer`.

### Examples & Docs
- Added `examples/inline` for bounded inline rendering.
- Candy Wars: custom meta path support via `FLUFFYUI_META_PATH`, focus/mount handling improvements for new-game flow, and expanded showcase coverage.
- Documentation expanded across README, getting started, debugging, performance, testing, architecture, migration, and new `docs/inline-mode.md`.

### Testing & Tooling
- Added PTY integration coverage for inline mode escape handling.
- Added render-pipeline benchmarks in `runtime` plus `scripts/bench-render-pipeline.sh`.
- Expanded golden-test coverage guard for core widget catalog with new fixtures.
- Added doctor command tests, web handler parsing tests, and broader candy-wars simulation/showcase tests.

## v0.3.3 - 2026-02-04

### Fixes
- TextArea: return a minimal preferred size (1x1) so flex layouts with unbounded constraints don't shrink it to 0.

## v0.3.2 - 2026-02-04

### Dependencies
- Backend: migrate tcell from v2 to v3.

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
