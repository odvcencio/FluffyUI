# ADR 0004: Render sampling and performance dashboard

- Status: accepted
- Date: 2026-01-31

## Context

We need a low-friction way to inspect runtime performance (FPS, render time,
flush time, dirty ratios) without external profilers for day-to-day tuning.

## Decision

- Use `runtime.RenderSampler` as the canonical render timing aggregator.
- Provide a `widgets.PerformanceDashboard` widget that renders summary stats.
- Include a demo (`examples/perf-dashboard`) and documentation in
  `docs/performance.md`.

## Consequences

- Apps can surface live performance metrics with minimal setup.
- Render sampling remains optional and lightweight.
