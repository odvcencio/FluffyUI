# Creating Custom Widgets

This guide walks through building a custom widget with FluffyUI's layout and
rendering pipeline.

## 1) Pick a base

Choose one of the base types:

- `widgets.Base` for non-focusable widgets
- `widgets.FocusableBase` for keyboard focus
- `widgets.Component` if you need reactive subscriptions

## 2) Implement the Widget interface

A minimal widget implements `Measure`, `Layout`, `Render`, and `HandleMessage`:

```go
package widgets

type MyWidget struct {
    widgets.Base
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

## 3) Add reactive state (optional)

If your widget uses signals, embed `widgets.Component` and observe changes in
`Bind`:

```go
type Counter struct {
    widgets.Component
    count *state.Signal[int]
}

func (c *Counter) Bind(services runtime.Services) {
    c.Component.Bind(services)
    c.Observe(c.count, func() {
        c.Services.Invalidate()
    })
}
```

## 4) Accessibility

Expose a role and label so screen readers can announce the widget:

```go
func (w *MyWidget) AccessibleRole() accessibility.Role {
    return accessibility.RoleButton
}

func (w *MyWidget) AccessibleLabel() string {
    return "My Widget"
}
```

## 5) Add tests

Use the simulation backend or helpers in `testing/`:

```go
func TestMyWidget(t *testing.T) {
    output := testing.RenderToString(NewMyWidget(), 20, 1)
    if !strings.Contains(output, "Hello") {
        t.Fatal("expected text not found")
    }
}
```

## 6) Compile-time interface checks

Add compile-time assertions near the bottom of your file:

```go
var _ runtime.Widget = (*MyWidget)(nil)
```

## Tips

- Use `services.Invalidate()` instead of rendering directly.
- Avoid allocations in hot render paths.
- Implement `runtime.Keyed` (or set `Base.ID`) if you need persistence.
