# Keybindings

The `keybind` package decouples key sequences from commands. It provides a
registry for commands, a keymap system, optional modes, and a router.

## Registry

```go
registry := keybind.NewRegistry()
registry.Register(keybind.Command{
    ID:          "app.quit",
    Title:       "Quit",
    Description: "Exit the application",
    Handler: func(ctx keybind.Context) {
        if ctx.App != nil {
            ctx.App.ExecuteCommand(runtime.Quit{})
        }
    },
})
```

## Keymaps

```go
keymap := &keybind.Keymap{
    Name: "main",
    Bindings: []keybind.Binding{
        {Key: keybind.MustParseKeySequence("ctrl+p"), Command: "palette.open"},
        {Key: keybind.MustParseKeySequence("tab"), Command: "focus.next"},
    },
}
```

Keymaps can be stacked, and you can define multiple modes using a
`keybind.ModeManager`.

If you need explicit error handling, use `ParseKeySequence` and handle the
returned error before constructing bindings.

## Conditions

Use conditions to gate bindings based on context:

```go
{Key: keybind.MustParseKeySequence("ctrl+c"), Command: "app.quit",
 When: keybind.WhenFocusedNotClipboardTarget()},
```

## Router integration

```go
stack := &keybind.KeymapStack{}
stack.Push(keymap)
router := keybind.NewKeyRouter(registry, nil, stack)
handler := &keybind.RuntimeHandler{Router: router}

app := runtime.NewApp(runtime.AppConfig{KeyHandler: handler})
```

## Focus registration

The focus scope needs a list of focusable widgets. You can register them once
after the screen is initialized:

```go
if screen := app.Screen(); screen != nil {
    runtime.RegisterFocusables(screen.FocusScope(), root)
}
```

To enable automatic focus registration when roots and overlays change:

```go
app := runtime.NewApp(runtime.AppConfig{
    FocusRegistration: runtime.FocusRegistrationAuto,
})
```

If your widget tree changes dynamically, call `screen.RefreshFocusables()` to
rescan.

## Command palette

`widgets.EnhancedPalette` builds a palette from the registry and can show
shortcuts when keymaps are provided. See `examples/command-palette`.
