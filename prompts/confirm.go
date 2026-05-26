package prompts

import (
	"context"

	"m31labs.dev/fluffyui/fluffy"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/terminal"
)

// Confirm displays a yes/no prompt and returns the user's choice.
// The default answer is true (yes). Press y/Y or Enter for yes, n/N for no.
// Ctrl+C and Escape return false.
//
// Example:
//
//	ok, err := prompts.Confirm("Proceed with deployment?")
func Confirm(question string, opts ...fluffy.AppOption) (bool, error) {
	return ConfirmContext(context.Background(), question, opts...)
}

// ConfirmContext is like Confirm but accepts a context for cancellation.
func ConfirmContext(ctx context.Context, question string, opts ...fluffy.AppOption) (bool, error) {
	ctx = promptContext(ctx)
	result := true

	label := question + " [Y/n] "
	w := promptWidget(label)

	err := runPrompt(ctx, w, 1, func(key runtime.KeyMsg) bool {
		switch {
		case key.Key == terminal.KeyRune && (key.Rune == 'y' || key.Rune == 'Y'):
			result = true
			return true
		case key.Key == terminal.KeyRune && (key.Rune == 'n' || key.Rune == 'N'):
			result = false
			return true
		case key.Key == terminal.KeyEnter:
			// Default is yes.
			return true
		case key.Key == terminal.KeyEscape || keyIsCtrlC(key):
			result = false
			return true
		}
		return false
	}, opts...)

	return result, err
}
