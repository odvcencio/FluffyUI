# v0.5.0 Framework Improvements Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make FluffyUI the most capable Go TUI framework by addressing architecture bottlenecks, filling widget gaps, improving terminal protocol support, and enhancing developer experience.

**Architecture:** Changes are organized into 4 phases. Phase 1 (Quick Wins) ships fast value with minimal risk. Phase 2 (Core Improvements) addresses fundamental architecture. Phase 3 (New Widgets + DX) expands the catalog. Phase 4 (Terminal Protocols + Ecosystem) adds progressive enhancement features. Each phase is independently shippable.

**Tech Stack:** Go 1.21+, tcell v3 (terminal backend), compositor package (ANSI rendering), state package (reactive signals + history)

---

## Phase 1: Quick Wins (< 1 day each)

### Task 1: Synchronized Output (CSI ?2026h)

Wrap the compositor's `Render()` and `RenderFull()` output with synchronized update markers to eliminate tearing on terminals that support it (kitty, foot, WezTerm, iTerm2).

**Files:**
- Modify: `compositor/diff.go:23-95` (Render method)
- Modify: `compositor/ansi.go:10-22` (add constants)
- Test: `compositor/compositor_test.go`

**Step 1: Add ANSI constants for synchronized output**

In `compositor/ansi.go`, add after line 22:

```go
ANSISyncStart = "\x1b[?2026h"
ANSISyncEnd   = "\x1b[?2026l"
```

**Step 2: Wrap Render() output with sync markers**

In `compositor/diff.go`, modify `Render()` at line 30 to emit sync start before `HideCursor()`:

```go
r.writer.buf.WriteString(ANSISyncStart)
r.writer.HideCursor()
```

And before the return at line 94, emit sync end:

```go
r.writer.buf.WriteString(ANSISyncEnd)
return r.writer.String()
```

**Step 3: Same for RenderFull()**

In `RenderFull()` at line 106, emit sync start before clear screen:

```go
r.writer.buf.WriteString(ANSISyncStart)
r.writer.buf.WriteString(ANSIClearScreen)
```

And before the return, emit sync end.

**Step 4: Write test verifying sync markers present**

```go
func TestRender_SynchronizedOutput(t *testing.T) {
    s := NewScreen(10, 2)
    s.Set(0, 0, Cell{Rune: 'X', Style: Style{}, Width: 1})
    r := NewRenderer(s)
    output := r.Render()
    if !strings.Contains(output, ANSISyncStart) {
        t.Error("missing sync start marker")
    }
    if !strings.Contains(output, ANSISyncEnd) {
        t.Error("missing sync end marker")
    }
}
```

**Step 5: Run tests and commit**

```bash
go test ./compositor/...
git add compositor/ansi.go compositor/diff.go compositor/compositor_test.go
git commit -m "feat(compositor): add synchronized output (CSI ?2026h) to eliminate tearing"
```

---

### Task 2: Cursor Shape Support

Add cursor shape control to the Backend interface and compositor. Widgets like Input can then set beam cursor when focused.

**Files:**
- Modify: `backend/backend.go:10-53` (add SetCursorShape method to interface)
- Modify: `compositor/ansi.go` (add cursor shape sequences)
- Modify: `backend/tcell/backend.go` (implement in tcell backend)
- Modify: `backend/sim.go` or test backend (stub implementation)
- Modify: `widgets/input.go` (set beam cursor on focus)
- Test: `backend/backend_test.go`

**Step 1: Define CursorShape type and add to Backend interface**

In `backend/backend.go`, add before the interface:

```go
// CursorShape controls the terminal cursor appearance.
type CursorShape int

const (
    CursorDefault    CursorShape = iota // Terminal default
    CursorBlock                         // Block cursor (█)
    CursorUnderline                     // Underline cursor (_)
    CursorBeam                          // Beam/bar cursor (|)
)
```

Add optional interface (don't break existing backends):

```go
// CursorShapeSetter allows backends to control cursor appearance.
type CursorShapeSetter interface {
    SetCursorShape(shape CursorShape)
}
```

**Step 2: Add ANSI sequences**

In `compositor/ansi.go`, add:

```go
// CursorShapeSeq returns the ANSI sequence for a cursor shape.
// Uses DECSCUSR (CSI Ps SP q).
func CursorShapeSeq(shape int) string {
    return fmt.Sprintf("\x1b[%d q", shape)
}
```

Mapping: Default=0, Block=2, Underline=4, Beam=6 (per DECSCUSR spec).

**Step 3: Implement in tcell backend**

In the tcell backend's `SetCursorShape` method, write the DECSCUSR sequence directly to stdout since tcell doesn't natively support cursor shapes:

```go
func (b *Backend) SetCursorShape(shape backend.CursorShape) {
    var ps int
    switch shape {
    case backend.CursorBlock:
        ps = 2
    case backend.CursorUnderline:
        ps = 4
    case backend.CursorBeam:
        ps = 6
    default:
        ps = 0
    }
    fmt.Fprintf(os.Stdout, "\x1b[%d q", ps)
}
```

**Step 4: Wire into Input widget**

In `widgets/input.go`, in `Focus()` method, after setting focus state:

```go
if i.services.Backend != nil {
    if cs, ok := i.services.Backend.(backend.CursorShapeSetter); ok {
        cs.SetCursorShape(backend.CursorBeam)
    }
}
```

In `Blur()`, reset to default.

**Step 5: Run tests and commit**

```bash
go test ./backend/... ./widgets/...
git add backend/backend.go compositor/ansi.go widgets/input.go
# + tcell backend files
git commit -m "feat(backend): add cursor shape support (DECSCUSR)"
```

---

### Task 3: OSC-8 Hyperlinks

Add hyperlink support so widgets like Breadcrumb and any Text widget can emit clickable URLs in supporting terminals.

**Files:**
- Modify: `compositor/ansi.go` (add OSC-8 helper)
- Create: `widgets/link.go` (Link widget)
- Test: `widgets/link_test.go`

**Step 1: Add OSC-8 helper to compositor**

In `compositor/ansi.go`:

```go
// Hyperlink wraps text in an OSC-8 hyperlink sequence.
// Terminals that don't support OSC-8 will ignore the sequences.
func Hyperlink(url, text string) string {
    return fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", url, text)
}
```

**Step 2: Create Link widget**

`widgets/link.go` — a simple focusable widget that renders text with OSC-8 wrapping and opens URLs on Enter:

```go
type Link struct {
    FocusableBase
    url   string
    label string
    style backend.Style
}
```

Render method uses `Hyperlink()` to wrap the label in OSC-8 sequences. HandleMessage activates on Enter key.

**Step 3: Write tests and commit**

```bash
go test ./compositor/... ./widgets/...
git add compositor/ansi.go widgets/link.go widgets/link_test.go
git commit -m "feat(widgets): add Link widget with OSC-8 hyperlink support"
```

---

### Task 4: TextArea Undo/Redo

Port the undo/redo pattern from Input to TextArea using `state.History[textAreaState]`.

**Files:**
- Modify: `widgets/textarea.go:16-32` (add history field)
- Modify: `widgets/textarea.go` (add Undo/Redo/CanUndo/CanRedo methods)
- Test: `widgets/textarea_coverage_test.go`

**Step 1: Define textAreaState type**

In `widgets/textarea.go`, add:

```go
type textAreaState struct {
    text   []rune
    cursor int
}
```

**Step 2: Add history field to TextArea struct**

```go
history *state.History[textAreaState]
```

Initialize in `NewTextArea()`:

```go
ta.history = state.NewHistory(textAreaState{}, state.WithGroupWindow(300*time.Millisecond))
```

**Step 3: Add Undo/Redo methods (mirror Input's API)**

```go
func (t *TextArea) Undo() bool {
    if t.history == nil { return false }
    s, ok := t.history.Undo()
    if !ok { return false }
    t.text = s.text
    t.cursor = s.cursor
    t.syncValue()
    return true
}
```

Same pattern for Redo, CanUndo, CanRedo.

**Step 4: Add pushHistory calls in HandleMessage**

After each edit operation (character insertion, deletion, paste, etc.), call:

```go
t.pushHistory(true) // grouped
```

Where pushHistory is:
```go
func (t *TextArea) pushHistory(grouped bool) {
    if t.history == nil { return }
    s := textAreaState{text: append([]rune(nil), t.text...), cursor: t.cursor}
    if grouped {
        t.history.PushGrouped(s)
    } else {
        t.history.Push(s)
    }
}
```

**Step 5: Wire Ctrl+Z/Ctrl+Y in HandleMessage**

In the KeyMsg handler, add:
```go
case key.Ctrl && key.Rune == 'z':
    if key.Shift { t.Redo() } else { t.Undo() }
    return runtime.Handled()
case key.Ctrl && key.Rune == 'y':
    t.Redo()
    return runtime.Handled()
```

**Step 6: Write tests**

```go
func TestTextArea_UndoRedo(t *testing.T) {
    ta := NewTextArea()
    ta.SetText("hello")
    ta.pushHistory(false)
    ta.SetText("world")
    ta.pushHistory(false)
    if !ta.CanUndo() { t.Fatal("expected CanUndo") }
    ta.Undo()
    if ta.Text() != "hello" { t.Fatalf("got %q", ta.Text()) }
    ta.Redo()
    if ta.Text() != "world" { t.Fatalf("got %q", ta.Text()) }
}
```

**Step 7: Run tests and commit**

```bash
go test ./widgets/... ./state/...
git add widgets/textarea.go widgets/textarea_coverage_test.go
git commit -m "feat(widgets): add undo/redo to TextArea (Ctrl+Z/Y)"
```

---

### Task 5: Error Boundary Wrapper

Add a recovery wrapper that catches panics in widget Render/HandleMessage and shows an error widget instead of crashing.

**Files:**
- Create: `runtime/error_boundary.go`
- Test: `runtime/error_boundary_test.go`

**Step 1: Create ErrorBoundary widget wrapper**

```go
// ErrorBoundary wraps a widget and catches panics during Render and HandleMessage.
// If the wrapped widget panics, it displays an error message instead of crashing.
type ErrorBoundary struct {
    child    Widget
    err      error
    reporter *ErrorReporter
    bounds   Rect
}

func NewErrorBoundary(child Widget) *ErrorBoundary {
    return &ErrorBoundary{child: child}
}
```

**Step 2: Implement safe Render with recover**

```go
func (e *ErrorBoundary) Render(ctx RenderContext) {
    if e.err != nil {
        // Render error display
        renderErrorDisplay(ctx, e.bounds, e.err)
        return
    }
    defer func() {
        if r := recover(); r != nil {
            e.err = fmt.Errorf("widget panic: %v", r)
            if e.reporter != nil {
                e.reporter.Report(e.err)
            }
        }
    }()
    e.child.Render(ctx)
}
```

**Step 3: Same for HandleMessage with recover**

**Step 4: Write test with a panicking widget**

```go
type panicWidget struct { Base }
func (p *panicWidget) Render(ctx runtime.RenderContext) { panic("boom") }

func TestErrorBoundary_CatchesPanic(t *testing.T) {
    eb := runtime.NewErrorBoundary(&panicWidget{})
    // Should not panic
    eb.Render(mockCtx)
    if eb.Error() == nil { t.Fatal("expected error") }
}
```

**Step 5: Run tests and commit**

```bash
go test ./runtime/...
git add runtime/error_boundary.go runtime/error_boundary_test.go
git commit -m "feat(runtime): add ErrorBoundary widget wrapper for crash resilience"
```

---

### Task 6: Stylesheet Hot Reload Flag

Add `--watch` support to the `fluffy` CLI for automatic stylesheet reloading.

**Files:**
- Modify: `cmd/fluffy/main.go` or wherever the CLI lives
- Modify: `runtime/style_watcher.go` (if exists) or create one

**Step 1: Check existing style watcher**

Read `runtime/style_resolver.go:247` and check if `StylesheetFileWatcher` exists. If yes, wire it to a CLI flag. If not, create a simple fsnotify-based watcher.

**Step 2: Add `--watch` flag**

```go
watchFlag := flag.Bool("watch", false, "Watch stylesheet files for changes and hot-reload")
```

**Step 3: Wire watcher in app initialization**

When `--watch` is set, start the watcher goroutine before `app.Run()`.

**Step 4: Run tests and commit**

```bash
go build ./cmd/fluffy/...
git add cmd/fluffy/
git commit -m "feat(cli): add --watch flag for stylesheet hot reload"
```

---

## Phase 2: Core Architecture

### Task 7: Message Priority Queue

Replace the single `chan Message` with a priority queue so keyboard events aren't delayed by bulk timer ticks.

**Files:**
- Create: `runtime/priority_queue.go`
- Modify: `runtime/app.go:77` (messages field)
- Modify: `runtime/app.go:490-540` (main event loop)
- Test: `runtime/priority_queue_test.go`

**Step 1: Define priority levels**

```go
type MessagePriority int

const (
    PriorityHigh   MessagePriority = iota // Keyboard, accessibility
    PriorityNormal                        // Mouse, custom
    PriorityLow                           // Tick, invalidate, queue flush
)
```

**Step 2: Create priority channel multiplexer**

```go
type PriorityQueue struct {
    high   chan Message
    normal chan Message
    low    chan Message
}

// Recv returns the highest-priority available message.
// Drains high before normal, normal before low.
func (q *PriorityQueue) Recv() Message {
    select {
    case msg := <-q.high:
        return msg
    default:
    }
    select {
    case msg := <-q.high:
        return msg
    case msg := <-q.normal:
        return msg
    default:
    }
    select {
    case msg := <-q.high:
        return msg
    case msg := <-q.normal:
        return msg
    case msg := <-q.low:
        return msg
    }
}
```

**Step 3: Classify messages by priority**

```go
func classifyMessage(msg Message) MessagePriority {
    switch msg.(type) {
    case KeyMsg, FocusChangedMsg:
        return PriorityHigh
    case MouseMsg, PasteMsg, CustomMsg:
        return PriorityNormal
    case TickMsg, InvalidateMsg, QueueFlushMsg:
        return PriorityLow
    default:
        return PriorityNormal
    }
}
```

**Step 4: Update App to use PriorityQueue**

Replace `messages chan Message` with `*PriorityQueue`. Update `Post()` to route through `classifyMessage()`. Update main loop to use `Recv()`.

**Step 5: Write priority drain test**

Test that posting Low, High, Normal in that order results in receiving High, Normal, Low.

**Step 6: Run tests and commit**

```bash
go test ./runtime/...
git add runtime/priority_queue.go runtime/priority_queue_test.go runtime/app.go
git commit -m "feat(runtime): add message priority queue (keyboard > mouse > ticks)"
```

---

### Task 8: Layout Caching

Add measurement caching to avoid redundant Measure calls when constraints haven't changed.

**Files:**
- Create: `runtime/layout_cache.go`
- Modify: `runtime/flex.go:78-137` (Measure method)
- Modify: `runtime/screen.go:273-305` (relayout method)
- Test: `runtime/layout_cache_test.go`

**Step 1: Create MeasureCache**

```go
type MeasureCache struct {
    entries map[Widget]measureEntry
}

type measureEntry struct {
    constraints Constraints
    result      Size
    generation  int64
}

func (c *MeasureCache) Get(w Widget, constraints Constraints, gen int64) (Size, bool) {
    e, ok := c.entries[w]
    if !ok || e.generation != gen-1 || e.constraints != constraints {
        return Size{}, false
    }
    return e.result, true
}

func (c *MeasureCache) Set(w Widget, constraints Constraints, result Size, gen int64) {
    c.entries[w] = measureEntry{constraints, result, gen}
}
```

**Step 2: Thread cache through layout**

Pass `*MeasureCache` through the layout path. Before calling `child.Widget.Measure()`, check the cache. After measuring, store the result.

**Step 3: Invalidate on state change**

When a widget calls `Invalidate()`, remove its entry from the cache. When `relayout()` runs, bump the generation counter.

**Step 4: Benchmark before/after**

Use existing `runtime/buffer_bench_test.go` as a template. Measure layout time with and without cache on a tree of 100+ widgets.

**Step 5: Run tests and commit**

```bash
go test ./runtime/... -bench=BenchmarkLayout
git add runtime/layout_cache.go runtime/layout_cache_test.go runtime/flex.go runtime/screen.go
git commit -m "feat(runtime): add layout measurement caching for faster re-renders"
```

---

### Task 9: App Struct Decomposition — Extract FocusManager

Start decomposing the monolithic App struct. First extraction: focus management.

**Files:**
- Modify: `runtime/focus.go:1-238` (extract FocusManager)
- Modify: `runtime/screen.go` (use FocusManager)
- Modify: `runtime/app.go` (delegate to FocusManager)
- Test: `runtime/focus_test.go`

**Step 1: Create FocusManager type**

Extract focus-related logic from Screen into a dedicated FocusManager:

```go
type FocusManager struct {
    scopes       []*FocusScope
    autoRegister bool
    policy       AutoFocusPolicy
    announcer    accessibility.Announcer
}
```

**Step 2: Move focus methods to FocusManager**

Move `RefreshFocusables`, `configureFocusScope`, `shouldRelayoutOnFocus`, `refreshLayerFocusables`, `announceFocus` from Screen to FocusManager.

**Step 3: Update Screen to delegate**

Screen gets a `focusManager *FocusManager` field and delegates focus operations to it.

**Step 4: Run tests — all existing focus tests must pass**

```bash
go test ./runtime/... -run TestFocus
```

**Step 5: Commit**

```bash
git add runtime/focus.go runtime/screen.go runtime/app.go
git commit -m "refactor(runtime): extract FocusManager from Screen"
```

---

### Task 10: App Struct Decomposition — Extract InputDispatcher

**Files:**
- Create: `runtime/input.go`
- Modify: `runtime/app.go:557-607` (dispatchMessage, DefaultUpdate)

**Step 1: Create InputDispatcher**

```go
type InputDispatcher struct {
    keyHandler KeyHandler
    screen     *Screen
}

func (d *InputDispatcher) Dispatch(msg Message) (handled bool, commands []Command) {
    // Move logic from DefaultUpdate and dispatchMessage here
}
```

**Step 2: Wire into App**

Replace inline dispatch logic in `DefaultUpdate` and `dispatchMessage` with calls to `InputDispatcher.Dispatch()`.

**Step 3: Run tests and commit**

```bash
go test ./runtime/...
git add runtime/input.go runtime/app.go
git commit -m "refactor(runtime): extract InputDispatcher from App"
```

---

## Phase 3: New Widgets + Developer Experience

### Task 11: Widget Bind/Unbind Coverage Pass

Wire remaining widgets that don't have Bind/Unbind. Priority targets: Table, Tree, Checkbox, Breadcrumb, Stepper, TimePicker, Tooltip, Popover, LineChart, RichText, Splitter.

**Files:**
- Modify: `widgets/table.go`
- Modify: `widgets/tree.go`
- Modify: `widgets/checkbox.go`
- Modify: `widgets/breadcrumb.go`
- Modify: `widgets/stepper.go`
- Modify: `widgets/time_picker.go`
- Modify: `widgets/tooltip.go`
- Modify: `widgets/popover.go`
- Modify: `widgets/line_chart.go`
- Modify: `widgets/rich_text.go`
- Modify: `widgets/splitter.go`
- Test: Existing test files for each widget

**Step 1: Add boilerplate to each widget**

For each widget that doesn't have Bind/Unbind, add:

```go
func (w *WidgetType) Bind(services runtime.Services) {
    w.services = services
}

func (w *WidgetType) Unbind() {
    w.services = runtime.Services{}
}
```

And add the `services runtime.Services` field to the struct.

**Step 2: Add Bindable/Unbindable assertions**

```go
var (
    _ runtime.Bindable   = (*WidgetType)(nil)
    _ runtime.Unbindable = (*WidgetType)(nil)
)
```

**Step 3: Run tests and commit**

```bash
go test ./widgets/...
git add widgets/table.go widgets/tree.go widgets/checkbox.go widgets/breadcrumb.go \
    widgets/stepper.go widgets/time_picker.go widgets/tooltip.go widgets/popover.go \
    widgets/line_chart.go widgets/rich_text.go widgets/splitter.go
git commit -m "feat(widgets): add Bind/Unbind service wiring to 11 widgets"
```

---

### Task 12: Table Sorting and Filtering

Add column sorting (click header to toggle asc/desc/none) and row filtering to the Table widget.

**Files:**
- Modify: `widgets/table.go`
- Test: `widgets/table_test.go` or new test file

**Step 1: Add sort state to Table**

```go
type SortDirection int

const (
    SortNone SortDirection = iota
    SortAsc
    SortDesc
)

type TableSortState struct {
    Column    int
    Direction SortDirection
}
```

Add to Table struct: `sortState TableSortState`, `filterFn func(row []string) bool`, `sortedIndices []int`.

**Step 2: Implement sort logic**

```go
func (t *Table) SetSortColumn(col int, dir SortDirection) {
    t.sortState = TableSortState{Column: col, Direction: dir}
    t.rebuildSortedIndices()
    t.syncA11y()
}
```

**Step 3: Implement filter**

```go
func (t *Table) SetFilter(fn func(row []string) bool) {
    t.filterFn = fn
    t.rebuildSortedIndices()
}
```

**Step 4: Update Render to use sorted/filtered indices**

**Step 5: Write tests and commit**

```bash
go test ./widgets/... -run TestTable
git add widgets/table.go widgets/table_test.go
git commit -m "feat(widgets): add column sorting and row filtering to Table"
```

---

### Task 13: New Widgets — Pagination

**Files:**
- Create: `widgets/pagination.go`
- Test: `widgets/pagination_test.go`

**Step 1: Define Pagination widget**

```go
type Pagination struct {
    FocusableBase
    total      int
    current    int
    pageSize   int
    onChange   func(page int)
    services   runtime.Services
}
```

Renders as: `< 1 2 [3] 4 5 ... 10 >` with keyboard navigation (left/right arrows).

**Step 2: Implement Measure/Layout/Render/HandleMessage**

**Step 3: Add accessibility**

Role: `RoleGroup` with child buttons. Each page number is a focusable button with `PosInSet`/`SetSize`.

**Step 4: Write tests and commit**

```bash
go test ./widgets/... -run TestPagination
git add widgets/pagination.go widgets/pagination_test.go
git commit -m "feat(widgets): add Pagination widget"
```

---

### Task 14: New Widgets — TagInput

**Files:**
- Create: `widgets/tag_input.go`
- Test: `widgets/tag_input_test.go`

**Step 1: Define TagInput widget**

Combines an Input with a list of removable tags. Typing text and pressing Enter/Tab adds a tag. Backspace on empty input removes last tag.

```go
type TagInput struct {
    FocusableBase
    tags      []string
    input     *Input
    maxTags   int
    onChange  func(tags []string)
    services  runtime.Services
}
```

**Step 2: Implement rendering (tags as pills + input)**

**Step 3: Accessibility — RoleGroup containing list of tags + input**

**Step 4: Write tests and commit**

```bash
go test ./widgets/... -run TestTagInput
git add widgets/tag_input.go widgets/tag_input_test.go
git commit -m "feat(widgets): add TagInput widget"
```

---

### Task 15: New Widgets — Rating

**Files:**
- Create: `widgets/rating.go`
- Test: `widgets/rating_test.go`

**Step 1: Define Rating widget**

```go
type Rating struct {
    FocusableBase
    value     int
    max       int
    onChange  func(value int)
    services  runtime.Services
}
```

Renders as: `★★★☆☆` or configurable characters. Left/right to change value.

**Step 2: Accessibility — RoleSlider with ValueInfo**

**Step 3: Write tests and commit**

```bash
go test ./widgets/... -run TestRating
git add widgets/rating.go widgets/rating_test.go
git commit -m "feat(widgets): add Rating widget"
```

---

### Task 16: New Widgets — Badge

**Files:**
- Create: `widgets/badge.go`
- Test: `widgets/badge_test.go`

**Step 1: Define Badge widget**

A simple non-focusable label overlay. Renders a small colored label (e.g., notification count, status dot).

```go
type Badge struct {
    Base
    text  string
    style backend.Style
}
```

**Step 2: Accessibility — RoleStatus with Live="polite"**

**Step 3: Write tests and commit**

---

### Task 17: New Widgets — Card

**Files:**
- Create: `widgets/card.go`
- Test: `widgets/card_test.go`

**Step 1: Define Card widget**

A bordered container with optional header, body, and footer sections.

```go
type Card struct {
    Base
    header  runtime.Widget
    body    runtime.Widget
    footer  runtime.Widget
    title   string
    style   CardStyle
}
```

**Step 2: Implement layout (vertical stack of header/body/footer)**

**Step 3: Accessibility — RoleGroup with label from title**

**Step 4: Write tests and commit**

---

### Task 18: New Widgets — Skeleton/Placeholder

**Files:**
- Create: `widgets/skeleton.go`
- Test: `widgets/skeleton_test.go`

**Step 1: Define Skeleton widget**

Renders animated placeholder content while data is loading. Uses `░` and `▒` characters with animation.

```go
type Skeleton struct {
    Base
    width   int
    height  int
    animate bool
    frame   int
}
```

**Step 2: Bind for animation ticks**

**Step 3: Accessibility — RoleStatus, Busy=true, Live="polite"**

**Step 4: Write tests and commit**

---

### Task 19: Interactive Widget Inspector

Build a toggleable debug overlay that shows the widget tree, focused widget properties, and accessibility state.

**Files:**
- Create: `widgets/inspector.go`
- Modify: `runtime/app.go` (add inspector toggle)
- Test: `widgets/inspector_test.go`

**Step 1: Define Inspector overlay widget**

```go
type Inspector struct {
    Base
    app       *runtime.App
    visible   bool
    selected  int
    tree      []inspectorEntry
}

type inspectorEntry struct {
    widget   runtime.Widget
    depth    int
    label    string
    role     string
    bounds   runtime.Rect
    focused  bool
}
```

**Step 2: Walk the widget tree**

Use `ChildProvider` interface to recursively build the tree.

**Step 3: Render as split panel**

Left panel: widget tree (indented). Right panel: properties of selected widget (role, label, value, state, bounds, accessibility properties).

**Step 4: Wire toggle key (Ctrl+Shift+I or F12)**

In `App`, add a key handler that toggles the inspector overlay via `PushLayer`/`PopLayer`.

**Step 5: Write tests and commit**

```bash
go test ./widgets/... -run TestInspector
git add widgets/inspector.go runtime/app.go widgets/inspector_test.go
git commit -m "feat(widgets): add interactive Inspector overlay (F12 toggle)"
```

---

## Phase 4: Terminal Protocols + Ecosystem

### Task 20: Terminal Capability Detection

Probe terminal capabilities at startup and expose them for progressive enhancement.

**Files:**
- Create: `terminal/caps.go`
- Modify: `backend/tcell/backend.go` (probe on Init)
- Test: `terminal/caps_test.go`

**Step 1: Define TerminalCaps**

```go
type TerminalCaps struct {
    TrueColor       bool
    SynchronizedOut  bool
    CursorShapes     bool
    OSC8Hyperlinks   bool
    Sixel            bool
    KittyGraphics    bool
    BracketedPaste   bool
    MouseSGR         bool
    UnicodeWidth     bool // wcwidth support
}
```

**Step 2: Probe at startup**

Use `$TERM`, `$COLORTERM`, `$TERM_PROGRAM` environment variables plus DA1/DA2 responses where safe.

**Step 3: Expose via Services**

Add `Caps() *terminal.TerminalCaps` to the Services interface so widgets can query capabilities.

**Step 4: Conditionally enable features**

- Synchronized output: only emit CSI ?2026h if `Caps.SynchronizedOut`
- Hyperlinks: only emit OSC-8 if `Caps.OSC8Hyperlinks`
- Cursor shapes: only emit DECSCUSR if `Caps.CursorShapes`

**Step 5: Write tests and commit**

```bash
go test ./terminal/... ./backend/...
git add terminal/caps.go backend/tcell/backend.go
git commit -m "feat(terminal): add capability detection for progressive enhancement"
```

---

### Task 21: DataGrid Sorting/Filtering/Column Resize

Extend DataGrid with interactive column headers.

**Files:**
- Modify: `widgets/data_grid.go`
- Test: `widgets/data_grid_test.go` or new

**Step 1: Add sort state (same pattern as Table)**

**Step 2: Add column resize via drag**

Track mouse drag on column borders. Update column widths proportionally.

**Step 3: Add filter row**

Optional filter inputs below headers. Each column gets a text input for filtering.

**Step 4: Write tests and commit**

---

### Task 22: Plugin Ecosystem — `fluffy install`

Add a simple widget package installer that fetches Go modules.

**Files:**
- Modify: `cmd/fluffy/main.go`
- Create: `cmd/fluffy/install.go`

**Step 1: Add `install` subcommand**

```bash
fluffy install github.com/user/fluffy-charts
```

Runs `go get` under the hood and registers the plugin in `plugin_registry.go`.

**Step 2: Add `list` subcommand for installed plugins**

**Step 3: Write tests and commit**

---

### Task 23: MCP Tool Namespacing

Group the 140+ MCP tools into namespaces for better LLM tool selection.

**Files:**
- Modify: `agent/mcp/tools.go`
- Modify: `agent/mcp/types.go`

**Step 1: Add namespace prefix to tool names**

```
ui.snapshot → ui/snapshot
ui.click    → ui/click
ui.type     → ui/type
ui.wait     → ui/wait
```

**Step 2: Maintain backwards compatibility**

Accept both old and new names during a transition period.

**Step 3: Write tests and commit**

---

## Execution Order

1. **Phase 1** (Tasks 1-6): Independent, can be done in any order. All quick wins.
2. **Phase 2** (Tasks 7-10): Priority queue first (7), then layout cache (8), then decompositions (9-10).
3. **Phase 3** (Tasks 11-19): Bind/Unbind coverage first (11), then new widgets in any order (12-18), inspector last (19).
4. **Phase 4** (Tasks 20-23): Capability detection first (20), then others in any order.

## Verification

After each phase:
1. `go build ./...` — must compile cleanly
2. `go test ./...` — all tests pass
3. `go vet ./...` — no warnings

After all phases:
1. Run the candy-wars example app to visually verify
2. Run `fluffy doctor` to verify capability detection
3. Run `FLUFFYUI_TTS=1` to verify announcements still work
4. Tag as v0.5.0
