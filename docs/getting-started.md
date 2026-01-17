# Getting Started

FluffyUI is a batteries-included TUI framework for Go. It ships with a runtime
loop, reactive state, layout, and a growing widget catalog.

## Install

```
go get github.com/odvcencio/fluffy-ui@latest
```

## Quickstart

```go
package main

import (
    "context"
    "fmt"
    "os"
    "time"

    "github.com/odvcencio/fluffy-ui/accessibility"
    "github.com/odvcencio/fluffy-ui/backend"
    backendtcell "github.com/odvcencio/fluffy-ui/backend/tcell"
    "github.com/odvcencio/fluffy-ui/clipboard"
    "github.com/odvcencio/fluffy-ui/keybind"
    "github.com/odvcencio/fluffy-ui/runtime"
    "github.com/odvcencio/fluffy-ui/widgets"
)

func main() {
    be, err := backendtcell.New()
    if err != nil {
        fmt.Fprintf(os.Stderr, "backend init failed: %v\n", err)
        os.Exit(1)
    }

    registry := keybind.NewRegistry()
    keybind.RegisterStandardCommands(registry)
    keybind.RegisterScrollCommands(registry)
    keybind.RegisterClipboardCommands(registry)

    keymap := keybind.DefaultKeymap()
    stack := &keybind.KeymapStack{}
    stack.Push(keymap)
    router := keybind.NewKeyRouter(registry, nil, stack)
    keyHandler := &keybind.RuntimeHandler{Router: router}

    app := runtime.NewApp(runtime.AppConfig{
        Backend:    be,
        TickRate:   time.Second / 30,
        KeyHandler: keyHandler,
        Announcer:  &accessibility.SimpleAnnouncer{},
        Clipboard:  &clipboard.MemoryClipboard{},
        FocusStyle: &accessibility.FocusStyle{
            Indicator: "> ",
            Style:     backend.DefaultStyle().Bold(true),
        },
    })

    app.SetRoot(widgets.NewLabel("Hello from FluffyUI"))

    if err := app.Run(context.Background()); err != nil && err != context.Canceled {
        fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
        os.Exit(1)
    }
}
```

## Examples

Run the examples directly from the repo:

```
go run ./examples/quickstart
```

Additional demos are available in `examples/` and include:

- `examples/counter`
- `examples/todo-app`
- `examples/command-palette`
- `examples/file-browser`
- `examples/dashboard`
- `examples/settings-form`
- `examples/accessibility-demo`
- `examples/widgets/gallery`
- `examples/widgets/layout`
- `examples/widgets/data`
- `examples/widgets/input`
- `examples/widgets/navigation`
- `examples/widgets/feedback`
- `examples/recording`

## Recording output

Set `FLUFFYUI_RECORD` to capture an asciicast file:

```
FLUFFYUI_RECORD=out.cast go run ./examples/quickstart
```

You can also export to video by setting `FLUFFYUI_RECORD_EXPORT`:

```
FLUFFYUI_RECORD=out.cast FLUFFYUI_RECORD_EXPORT=out.mp4 go run ./examples/quickstart
```

See `docs/recording.md` for details.
