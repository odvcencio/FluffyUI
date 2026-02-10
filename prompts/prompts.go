// Package prompts provides standalone interactive prompts for terminal applications.
// Each function creates a minimal FluffyUI inline-mode app, displays a prompt,
// waits for user input, and returns the result.
//
// These are designed for simple, one-shot interactions — like Charm's "huh" library.
// Each prompt takes over the terminal briefly, collects input, and exits.
//
// Example:
//
//	ok, err := prompts.Confirm("Delete all files?")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if ok {
//	    deleteAllFiles()
//	}
package prompts

import (
	"context"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/fluffy"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
	"github.com/odvcencio/fluffyui/widgets"
)

// promptWidget creates a SimpleWidget that displays a single line of text.
func promptWidget(text string) *widgets.SimpleWidget {
	w := widgets.NewSimpleWidget()
	w.MeasureFunc = func(c runtime.Constraints) runtime.Size {
		width := len([]rune(text))
		if width > c.MaxWidth {
			width = c.MaxWidth
		}
		return runtime.Size{Width: width, Height: 1}
	}
	w.RenderFunc = func(ctx runtime.RenderContext) {
		b := w.ContentBounds()
		if b.Width <= 0 || b.Height <= 0 {
			return
		}
		ctx.Buffer.SetString(b.X, b.Y, text, backend.DefaultStyle())
	}
	return w
}

// runPrompt creates an inline FluffyUI app with the given widget and
// an update function that intercepts key events. The keyHandler is called
// for each KeyMsg; it should return true when the prompt is complete, which
// triggers app quit. Additional fluffy.AppOption values can be passed for
// customization (e.g., injecting a test backend).
func runPrompt(ctx context.Context, w runtime.Widget, lines int, keyHandler func(runtime.KeyMsg) bool, opts ...fluffy.AppOption) error {
	allOpts := make([]fluffy.AppOption, 0, len(opts)+1)
	allOpts = append(allOpts, opts...)
	allOpts = append(allOpts, fluffy.WithUpdate(func(app *runtime.App, msg runtime.Message) bool {
		if key, ok := msg.(runtime.KeyMsg); ok {
			if keyHandler(key) {
				app.ExecuteCommand(runtime.Quit{})
				return true
			}
		}
		return runtime.DefaultUpdate(app, msg)
	}))
	return fluffy.RunInlineContext(ctx, w, lines, allOpts...)
}

// promptContext returns the provided context, or context.Background() if nil.
func promptContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

// keyIsCtrlC reports whether a KeyMsg represents Ctrl+C.
func keyIsCtrlC(key runtime.KeyMsg) bool {
	return key.Key == terminal.KeyCtrlC
}
