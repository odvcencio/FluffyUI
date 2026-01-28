# Testing

FluffyUI provides comprehensive testing support through simulation backends
and test helpers. This enables deterministic, CI-friendly testing without
a real terminal.

## Simulation Backend

The `backend/sim` package provides a testable backend:

```go
import (
    "testing"
    "github.com/odvcencio/fluffyui/backend/sim"
    "github.com/odvcencio/fluffyui/runtime"
)

func TestMyWidget(t *testing.T) {
    be := sim.New(80, 24)
    if err := be.Init(); err != nil {
        t.Fatalf("init: %v", err)
    }
    defer be.Fini()

    // Use with full runtime
    app := runtime.NewApp(runtime.AppConfig{Backend: be})
    app.SetRoot(myWidget)

    // Run briefly then check output
    ctx, cancel := context.WithCancel(context.Background())
    cancel()
    _ = app.Run(ctx)

    if !be.ContainsText("expected") {
        t.Error("expected text not found")
    }
}
```

## Test Helpers

The `testing` package provides utilities for common test patterns.

### RenderToString

Render a widget to a string without a backend:

```go
import "github.com/odvcencio/fluffyui/testing"

func TestWidgetRender(t *testing.T) {
    widget := widgets.NewLabel("Hello")
    output := fluffytest.RenderToString(widget, 40, 1)

    if !strings.Contains(output, "Hello") {
        t.Error("expected 'Hello' in output")
    }
}
```

### Simulation Assertions

Convenient assertion helpers:

```go
func TestWithAssertions(t *testing.T) {
    be := sim.New(80, 24)
    be.Init()
    defer be.Fini()

    // ... render widget ...

    fluffytest.AssertContains(t, be, "expected text")
    fluffytest.AssertNotContains(t, be, "error")
    fluffytest.AssertTextAt(t, be, 0, 0, "Header")
}
```

### Accessibility Assertions

Capture announcements from `accessibility.SimpleAnnouncer`:

```go
announcer := fluffytest.NewAnnouncer()
app := runtime.NewApp(runtime.AppConfig{
    Announcer: announcer,
    // ...
})

// ... run app ...

fluffytest.AssertAnnounced(t, announcer, "Line 3 of 10")
```

### Input Injection

Test keyboard interaction:

```go
func TestKeyboardInput(t *testing.T) {
    be := sim.New(80, 24)
    be.Init()
    defer be.Fini()

    input := widgets.NewInput()
    // ... setup app ...

    // Type text
    be.InjectKeyString("hello")

    // Press special keys
    be.InjectKey(terminal.KeyEnter, 0)
    be.InjectKey(terminal.KeyTab, 0)

    // Wait for processing
    fluffytest.WaitForRender(app, 100*time.Millisecond)
}
```

## Testing Patterns

### Pattern 1: Snapshot Testing

Compare rendered output against golden files:

```go
func TestSnapshot(t *testing.T) {
    widget := NewMyWidget()
    output := fluffytest.RenderToString(widget, 80, 24)

    goldenPath := "testdata/mywidget.golden"

    if *updateSnapshots {
        os.WriteFile(goldenPath, []byte(output), 0644)
        return
    }

    expected, _ := os.ReadFile(goldenPath)
    if output != string(expected) {
        t.Errorf("snapshot mismatch:\nExpected:\n%s\nGot:\n%s", expected, output)
    }
}
```

### Pattern 2: Widget-Only Testing

Test widgets without the full runtime:

```go
func TestWidgetMeasure(t *testing.T) {
    w := widgets.NewLabel("Test")

    size := w.Measure(runtime.Constraints{MaxWidth: 100, MaxHeight: 10})

    if size.Width != 4 {
        t.Errorf("expected width 4, got %d", size.Width)
    }
}

func TestWidgetLayout(t *testing.T) {
    w := widgets.NewLabel("Test")
    w.Layout(runtime.Rect{X: 5, Y: 10, Width: 20, Height: 1})

    bounds := w.Bounds()
    if bounds.X != 5 || bounds.Y != 10 {
        t.Error("incorrect bounds after layout")
    }
}
```

### Pattern 3: Direct Render Testing

Render directly to a buffer:

```go
func TestWidgetRender(t *testing.T) {
    w := widgets.NewLabel("Hello")
    buf := runtime.NewBuffer(20, 1)

    w.Measure(runtime.Constraints{MaxWidth: 20, MaxHeight: 1})
    w.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 1})
    w.Render(runtime.RenderContext{Buffer: buf})

    // Check specific cells
    cell := buf.Get(0, 0)
    if cell.Rune != 'H' {
        t.Errorf("expected 'H' at (0,0), got %c", cell.Rune)
    }
}
```

### Pattern 4: Interactive Testing

Test user interaction flows:

```go
func TestFormSubmission(t *testing.T) {
    be := sim.New(80, 24)
    be.Init()
    defer be.Fini()

    submitted := false
    form := NewForm(func(data FormData) {
        submitted = true
    })

    app := runtime.NewApp(runtime.AppConfig{Backend: be, Root: form})

    // Run in background
    ctx, cancel := context.WithCancel(context.Background())
    go app.Run(ctx)
    defer cancel()

    // Wait for initial render
    time.Sleep(50 * time.Millisecond)

    // Fill form
    be.InjectKeyString("John Doe")
    be.InjectKey(terminal.KeyTab, 0)
    be.InjectKeyString("john@example.com")
    be.InjectKey(terminal.KeyEnter, 0)

    // Wait for processing
    time.Sleep(50 * time.Millisecond)

    if !submitted {
        t.Error("expected form to be submitted")
    }
}
```

### Pattern 5: Style Verification

Test styling is applied correctly:

```go
func TestWidgetStyling(t *testing.T) {
    be := sim.New(20, 1)
    be.Init()
    defer be.Fini()

    // Render styled widget
    style := backend.DefaultStyle().Bold(true).Foreground(backend.ColorRed)
    w := widgets.NewLabel("Bold Red")
    w.SetStyle(style)

    fluffytest.RenderTo(be, w, 20, 1)

    // Verify style at position
    _, _, capturedStyle := be.CaptureCell(0, 0)
    if capturedStyle.Attributes()&backend.AttrBold == 0 {
        t.Error("expected bold attribute")
    }
}
```

## Simulation Backend API

### Creating

```go
be := sim.New(width, height int) *Backend
```

### Lifecycle

```go
be.Init() error      // Initialize backend
be.Fini()            // Cleanup
be.Resize(w, h int)  // Change size
```

### Rendering

```go
be.SetContent(x, y int, r rune, comb []rune, style Style)
be.Show()
```

### Capture

```go
be.Capture() string                    // Full screen as string
be.CaptureRegion(x, y, w, h) string   // Region as string
be.CaptureCell(x, y) (rune, []rune, Style)  // Single cell
```

### Search

```go
be.ContainsText(text string) bool      // Check if text exists
be.FindText(text string) (x, y int)    // Find text position
```

### Input Injection

```go
be.InjectKey(key Key, rune rune)       // Inject key event
be.InjectKeyRune(r rune)               // Inject character
be.InjectKeyString(s string)           // Inject string as keys
be.InjectResize(w, h int)              // Inject resize event
```

## Best Practices

1. **Use `t.TempDir()`** for any file-based tests
2. **Clean up with `defer be.Fini()`** to avoid resource leaks
3. **Use short timeouts** in CI to catch hangs quickly
4. **Prefer snapshot tests** for complex widget output
5. **Test edge cases**: empty content, overflow, resize
6. **Test focus states**: focused vs unfocused rendering

## Continuous Integration

Tests run without a TTY:

```yaml
# GitHub Actions example
test:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version: '1.22'
    - run: go test ./...
```

The simulation backend works in headless environments.

## Debugging Failed Tests

### Capture and Print

```go
func TestDebug(t *testing.T) {
    be := sim.New(80, 24)
    // ... test code ...

    t.Logf("Screen capture:\n%s", be.Capture())
}
```

### Visual Diff

For snapshot mismatches, use diff tools:

```bash
diff testdata/expected.golden testdata/actual.txt
```

### Step-by-Step Debugging

```go
func TestStepByStep(t *testing.T) {
    be := sim.New(80, 24)
    be.Init()
    defer be.Fini()

    // After each operation, capture state
    be.InjectKeyRune('a')
    t.Logf("After 'a':\n%s", be.Capture())

    be.InjectKeyRune('b')
    t.Logf("After 'b':\n%s", be.Capture())
}
```
