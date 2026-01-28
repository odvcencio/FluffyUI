package demo

import (
	"time"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/clipboard"
	"github.com/odvcencio/fluffyui/fluffy"
	"github.com/odvcencio/fluffyui/keybind"
	"github.com/odvcencio/fluffyui/runtime"
)

// Options configures demo app setup.
type Options struct {
	TickRate       time.Duration
	FocusIndicator string
	FocusStyle     *backend.Style
	Announcer      accessibility.Announcer
	Clipboard      clipboard.Clipboard
	CommandHandler runtime.CommandHandler
}

// Bundle exposes shared demo wiring.
type Bundle struct {
	App      *runtime.App
	Registry *keybind.CommandRegistry
	Keymaps  *keybind.KeymapStack
	Router   *keybind.KeyRouter
}

// NewApp builds a demo app with keybindings and focus registration.
func NewApp(root runtime.Widget, opts Options) (*Bundle, error) {
	var options []fluffy.AppOption
	options = append(options, fluffy.WithRoot(root))

	if opts.TickRate > 0 {
		options = append(options, fluffy.WithTickRate(opts.TickRate))
	}
	if opts.CommandHandler != nil {
		options = append(options, fluffy.WithCommandHandler(opts.CommandHandler))
	}
	if opts.FocusIndicator != "" {
		options = append(options, fluffy.WithFocusIndicator(opts.FocusIndicator))
	}
	if opts.FocusStyle != nil {
		options = append(options, fluffy.WithFocusStyle(*opts.FocusStyle))
	}
	if opts.Announcer != nil {
		options = append(options, fluffy.WithAnnouncer(opts.Announcer))
	}
	if opts.Clipboard != nil {
		options = append(options, fluffy.WithClipboard(opts.Clipboard))
	}

	bundle, err := fluffy.NewBundle(options...)
	if err != nil {
		return nil, err
	}

	return &Bundle{
		App:      bundle.App,
		Registry: bundle.Registry,
		Keymaps:  bundle.Keymaps,
		Router:   bundle.Router,
	}, nil
}
