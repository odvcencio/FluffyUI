# API Stability Policy

FluffyUI follows semantic versioning (semver) and targets predictable upgrade
paths for application teams adopting it in production.

## Versioning

- **MAJOR**: breaking API changes.
- **MINOR**: backward-compatible features.
- **PATCH**: backward-compatible bug fixes.

## Stability Tiers

Public APIs are grouped by stability tier:

- **Stable**: backward-compatible within a major version.
- **Provisional**: mostly stable, but may still evolve before v1.0.
- **Experimental**: subject to change without deprecation guarantees.

## v1.0 Interface Freeze (Planned)

The following runtime interfaces are treated as the v1.0 stable core and are
planned to remain source-compatible across v1.x:

- `runtime.Widget`
- `runtime.Focusable`
- `runtime.ChildProvider`
- `runtime.Invalidatable`
- `runtime.Bindable` / `runtime.Unbindable`
- Core geometry types: `runtime.Constraints`, `runtime.Size`, `runtime.Rect`

The `runtime.App` event loop contract (`Run`, `Post`, `SetRoot`, scheduler
integration) is also in the stable set.

## Provisional Surfaces

These APIs are production-ready but may receive additive or shape changes
before v1.0:

- `fluffy` convenience constructors/options (ergonomic layer).
- Theme/stylesheet composition helpers.
- Higher-level docs-site and scaffolding helpers in `cmd/fluffy`.

## Experimental Surfaces

These areas move faster and may change without deprecation windows:

- Agent protocol integration (`agent/`, MCP transport wiring).
- GPU canvas drivers (`gpu/`).
- Recording/export internals (`recording/`, `video/`).
- Web backend transport internals (`backend/web`).

## Deprecation Policy

When a stable/provisional API is slated for removal:

1. It is marked `Deprecated` in GoDoc.
2. A replacement path is documented.
3. The deprecated API remains available for **at least two minor releases**.

## Compatibility Scope

- Public Go APIs are covered by this policy.
- Internal packages and test-only helpers are not covered.
- Experimental features are explicitly excluded from compatibility guarantees.

## Migration Support

Migration guides are published in `docs/migration/` for major releases and
significant refactors.
