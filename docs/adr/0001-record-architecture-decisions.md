# ADR 0001: Record architecture decisions

- Status: accepted
- Date: 2026-01-31

## Context

FluffyUI has grown into a large framework with many subsystems, widgets, and
public APIs. Decisions around API consistency, performance trade-offs, and
cross-platform support need durable records for future contributors.

## Decision

We will maintain Architecture Decision Records (ADRs) in `docs/adr/` using a
simple template, and reference them from relevant documentation.

## Consequences

- Major changes require a short ADR to preserve rationale and context.
- Contributors can find historical decisions without diff archaeology.
