# Changelog

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
