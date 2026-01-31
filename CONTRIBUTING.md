# Contributing to FluffyUI

Thanks for taking the time to contribute! This guide summarizes the preferred workflows and standards for the repo.

## Prerequisites

- Go 1.24+
- Recommended: `staticcheck` (`go install honnef.co/go/tools/cmd/staticcheck@latest`)

## Local Setup

```bash
# Run the quickstart example
go run ./examples/quickstart

# Run all tests
go test ./...
```

## Code Style

- Prefer the options pattern for construction (`NewWidget(..., WithWidgetFoo(...))`).
- Prefer `SetOn*` for mutable callbacks; leave `On*` only for backward-compatible aliases.
- Add nil receiver guards for public methods when practical.
- Keep hot paths allocation-light (reuse buffers, avoid temp slices).

## Tests

- Use the simulation backend for widget tests.
- Prefer snapshot helpers when output is visual.
- Add at least one happy-path test plus a critical edge case.

## Commit Messages

```
type(scope): message
```

Examples:
- `add(widgets): new DataGrid widget`
- `fix(runtime): prevent nil deref in focus manager`

## Documentation

Update docs when you add or change:
- public APIs
- widgets or examples
- behavior that affects users

## Questions

If you are unsure about an approach, open an issue or start a discussion before large refactors.
