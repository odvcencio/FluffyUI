# Inline Mode

Inline mode runs a FluffyUI app in a bounded viewport on the primary terminal
screen instead of taking over the alternate screen.

This is useful for CLI tools that want richer UI interactions without a full
screen takeover.

## Quick Usage

One-liner:

```go
if err := fluffy.RunInline(root, 8); err != nil {
    log.Fatal(err)
}
```

Explicit app wiring:

```go
app, err := fluffy.NewApp(
    fluffy.WithRoot(root),
    fluffy.WithInlineMode(true),
    fluffy.WithInlineHeight(8),
)
```

`WithInlineHeight(n)` controls viewport rows. Values `<= 0` use backend
defaults.

## Capability Matrix

The table below reflects FluffyUI's expected inline behavior by terminal family.
Use `fluffy doctor` in your environment to confirm capabilities before shipping.

| Environment | Inline Viewport | Mouse Mapping | Color/Graphics Notes | Notes |
|-------------|-----------------|---------------|----------------------|-------|
| Kitty | Supported | Supported | True color expected, Kitty graphics available | Best feature coverage |
| WezTerm | Supported | Supported | True color expected | Strong default compatibility |
| Alacritty | Supported | Supported | True color expected | No native Kitty/Sixel protocol |
| iTerm2 | Supported | Supported | True color expected | Graphics protocol support varies |
| GNOME Terminal / other VTE terminals | Supported | Supported | True color usually available | Common Linux baseline |
| tmux | Supported (with caveats) | Supported | Depends on tmux/terminal config | Verify `TERM` and color passthrough |
| GNU screen | Supported (with caveats) | Supported | Depends on screen config | Legacy defaults may reduce color |
| Windows Terminal | Supported | Supported | True color expected | Validate in your shell profile |
| CI / `TERM=dumb` | Not supported | Not supported | Limited terminal capabilities | `fluffy doctor` reports this as fail |

## Validation Workflow

1. Run diagnostics:

```bash
go run ./cmd/fluffy doctor
```

2. Verify your inline app:

```bash
go run ./examples/inline
```

3. If behavior differs across environments, capture:
- `TERM` / `COLORTERM`
- shell + multiplexer (`tmux`, `screen`, etc.)
- `fluffy doctor --json` output

## Current Test Coverage

Inline mode is covered by:
- backend viewport and event mapping tests (`backend/tcell/coverage_extra_test.go`)
- PTY integration test for alternate-screen escape suppression (`backend/tcell/raw_tty_pty_test.go`)
- runtime option propagation tests (`runtime/app_test.go`, `fluffy/app_test.go`)
