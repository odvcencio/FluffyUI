# Getting Started

FluffyUI is a batteries-included TUI framework for Go. It ships with a runtime
loop, reactive state, layout, and a growing widget catalog.

## Install

```
go get github.com/odvcencio/fluffyui@latest
```

## Quickstart

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/odvcencio/fluffyui/fluffy"
)

func main() {
    app, err := fluffy.NewApp()
    if err != nil {
        fmt.Fprintf(os.Stderr, "app init failed: %v\n", err)
        os.Exit(1)
    }
    app.SetRoot(fluffy.NewLabel("Hello from FluffyUI"))

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
