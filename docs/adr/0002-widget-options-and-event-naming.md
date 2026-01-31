# ADR 0002: Standardize widget options and event naming

- Status: accepted
- Date: 2026-01-31

## Context

Widgets previously mixed constructor patterns (options vs. chaining) and event
handler naming (`OnX`, `SetOnX`, custom method names). This created a fragmented
API surface.

## Decision

- Standardize constructors on the options pattern (`NewWidget(..., opts...)`).
- Provide `With*` option helpers for construction-time configuration.
- Provide `SetOn*` mutators for event handlers; keep deprecated aliases for
  compatibility.

## Consequences

- New widgets must follow the options pattern and `SetOn*` naming.
- Existing chaining APIs are deprecated but retained for compatibility.
