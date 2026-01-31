# Advanced Patterns

This guide collects higher-level patterns for composing FluffyUI applications.

## Custom Widgets

- Embed `widgets.Base` or `widgets.FocusableBase`.
- Implement `Measure`, `Layout`, `Render`, and `HandleMessage`.
- If you subscribe to signals, embed `widgets.Component` and implement `Bind/Unbind`.

## Reactive State

Use `state.Signal` for reactive values and `state.Computed` for derived state. When a signal changes, call `services.Invalidate()` in your observer to trigger a re-render.

## Large Lists

Use `widgets.VirtualList` for large datasets. Provide a height function for variable row heights and use `SetLazyLoad` to fetch data as the viewport approaches the edges.

## Accessibility

- Set `Base.Role`, `Base.Label`, and `Base.Value`.
- Keep labels stable and concise.
- Update value text as user selection changes.

## Performance

- Avoid allocations in `Render` and `HandleMessage`.
- Cache measurements when possible.
- Use `runtime.RenderChild` for early out on non-visible widgets.
