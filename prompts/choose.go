package prompts

import (
	"context"
	"fmt"
	"strings"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/fluffy"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
	"github.com/odvcencio/fluffyui/widgets"
)

// Choose displays a list of options and returns the index of the selected one.
// Use arrow keys (or j/k) to navigate, Enter to confirm.
// Returns -1 on cancel (Escape or Ctrl+C).
//
// Example:
//
//	idx, err := prompts.Choose("Pick a color:", []string{"Red", "Green", "Blue"})
func Choose(question string, options []string, opts ...fluffy.AppOption) (int, error) {
	return ChooseContext(context.Background(), question, options, opts...)
}

// ChooseContext is like Choose but accepts a context for cancellation.
func ChooseContext(ctx context.Context, question string, options []string, opts ...fluffy.AppOption) (int, error) {
	ctx = promptContext(ctx)
	if len(options) == 0 {
		return -1, fmt.Errorf("prompts.Choose: no options provided")
	}

	selected := 0
	result := -1
	lines := len(options) + 1 // question + options

	w := widgets.NewSimpleWidget()
	w.MeasureFunc = func(c runtime.Constraints) runtime.Size {
		// Measure: question line + one line per option.
		width := len([]rune(question))
		for _, opt := range options {
			optWidth := len([]rune(opt)) + 4 // "  > " prefix
			if optWidth > width {
				width = optWidth
			}
		}
		if width > c.MaxWidth {
			width = c.MaxWidth
		}
		return runtime.Size{Width: width, Height: lines}
	}
	w.RenderFunc = func(ctx runtime.RenderContext) {
		b := w.ContentBounds()
		if b.Width <= 0 || b.Height <= 0 {
			return
		}
		// Render the question on the first line.
		ctx.Buffer.SetString(b.X, b.Y, question, backend.DefaultStyle().Bold(true))

		// Render options below.
		normal := backend.DefaultStyle()
		highlight := backend.DefaultStyle().Reverse(true)
		for i, opt := range options {
			y := b.Y + 1 + i
			if y >= b.Y+b.Height {
				break
			}
			var line string
			var s backend.Style
			if i == selected {
				line = fmt.Sprintf("  > %s", opt)
				s = highlight
			} else {
				line = fmt.Sprintf("    %s", opt)
				s = normal
			}
			// Pad to full width to clear previous content.
			if pad := b.Width - len([]rune(line)); pad > 0 {
				line += strings.Repeat(" ", pad)
			}
			ctx.Buffer.SetString(b.X, y, line, s)
		}
	}

	err := runPrompt(ctx, w, lines, func(key runtime.KeyMsg) bool {
		switch {
		case key.Key == terminal.KeyUp || (key.Key == terminal.KeyRune && key.Rune == 'k'):
			if selected > 0 {
				selected--
			}
			return false
		case key.Key == terminal.KeyDown || (key.Key == terminal.KeyRune && key.Rune == 'j'):
			if selected < len(options)-1 {
				selected++
			}
			return false
		case key.Key == terminal.KeyEnter:
			result = selected
			return true
		case key.Key == terminal.KeyEscape || keyIsCtrlC(key):
			result = -1
			return true
		}
		return false
	}, opts...)

	return result, err
}
