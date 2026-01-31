# API Stability Policy

FluffyUI follows semantic versioning (semver) and aims to provide predictable upgrade paths.

## Versioning

- **MAJOR**: breaking API changes
- **MINOR**: backward-compatible features
- **PATCH**: backward-compatible fixes

## Deprecations

When an API is slated for removal:

1. It is marked as **Deprecated** in GoDoc.
2. A recommended replacement is provided.
3. The deprecated API remains available for **at least two minor releases**.

## Compatibility Scope

- Public Go APIs are covered by this policy.
- Internal packages and `internal/` (if present) are not covered.
- Experimental features may change without notice and will be labeled as such.

## Migration Support

Migration guides are published in `docs/migration/` for major releases and significant refactors.
