# ADR 0003: Go 1.24 minimum

- Status: accepted
- Date: 2026-01-31

## Context

FluffyUI uses modern language features and benefits from the standard library
improvements in recent Go releases (e.g., built-in `min`/`max`). Supporting older
versions increases maintenance cost.

## Decision

The minimum supported Go version is 1.24. All tooling, templates, and CI
pipelines are aligned to Go 1.24.

## Consequences

- Contributors must use Go 1.24 or newer.
- CI and templates target `go 1.24`.
