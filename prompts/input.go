package prompts

import (
	"context"

	"github.com/odvcencio/fluffyui/fluffy"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
	"github.com/odvcencio/fluffyui/widgets"
)

// Input displays a text prompt and returns the user's input.
// Press Enter to submit. Press Escape or Ctrl+C to cancel (returns empty string).
//
// Example:
//
//	name, err := prompts.Input("What is your name?")
func Input(question string, opts ...fluffy.AppOption) (string, error) {
	return InputContext(context.Background(), question, opts...)
}

// InputContext is like Input but accepts a context for cancellation.
func InputContext(ctx context.Context, question string, opts ...fluffy.AppOption) (string, error) {
	ctx = promptContext(ctx)
	var result string

	label := widgets.NewLabel(question + " ")
	inp := widgets.NewInput()
	inp.SetPlaceholder("Type here...")
	row := fluffy.HStack(label, inp)

	allOpts := make([]fluffy.AppOption, 0, len(opts)+1)
	allOpts = append(allOpts, opts...)
	allOpts = append(allOpts, fluffy.WithUpdate(func(app *runtime.App, msg runtime.Message) bool {
		if key, ok := msg.(runtime.KeyMsg); ok {
			switch {
			case key.Key == terminal.KeyEnter:
				result = inp.Text()
				app.ExecuteCommand(runtime.Quit{})
				return true
			case key.Key == terminal.KeyEscape || keyIsCtrlC(key):
				result = ""
				app.ExecuteCommand(runtime.Quit{})
				return true
			}
		}
		return runtime.DefaultUpdate(app, msg)
	}))

	err := fluffy.RunInlineContext(ctx, row, 1, allOpts...)
	return result, err
}
