# Widget API Standards

This document defines the preferred API patterns for FluffyUI widgets. New or
refactored widgets should follow these conventions for consistency.

## Constructors

- Use `New*` constructors that take required arguments plus variadic options:
  - `NewButton(label string, opts ...ButtonOption)`
- Avoid chain-style configuration where possible.
- If chain methods exist, mark them as deprecated and route to option setters.

## Options Pattern

- Define a `FooOption` type and `With*` helpers:
  - `type ButtonOption func(*Button)`
  - `func WithVariant(v Variant) ButtonOption { ... }`
- Options should be nil-safe and idempotent.

## Event Handlers

- Prefer `WithOn*` option helpers for event callbacks.
- For mutable callbacks, prefer `SetOn*` methods (e.g., `SetOnClick`).
- Avoid exposing unexported `onX` fields except as internal storage.

## State / Binding

- Widgets that depend on runtime services or subscribe to signals must implement
  `Bind(services runtime.Services)` and `Unbind()`.
- Use `widgets.Component` when you need subscriptions and invalidation helpers.

## Accessibility

- Implement `AccessibleRole`, `AccessibleName`, and `AccessibleState` where
  appropriate, or ensure the base fields are kept in sync via `syncA11y`.
- Keep accessibility labels stable and meaningful.

## Nil Safety

- Public methods should guard against nil receivers when practical, especially
  in layout/render paths.

## Style Integration

- Implement `StyleType()` and `StyleClasses()` when the widget is stylable.
- Use `resolveBaseStyle` / `mergeBackendStyles` for consistent theme behavior.
