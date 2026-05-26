package demo

import (
	"time"

	"m31labs.dev/fluffyui/accessibility"
	_ "m31labs.dev/fluffyui/agent/mcp" // register MCP enabler for FLUFFY_MCP env
	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/clipboard"
	"m31labs.dev/fluffyui/fluffy"
	"m31labs.dev/fluffyui/keybind"
	"m31labs.dev/fluffyui/runtime"
)

// Options configures demo app setup.
type Options struct {
	TickRate       time.Duration
	FocusIndicator string
	FocusStyle     *backend.Style
	Announcer      accessibility.Announcer
	Clipboard      clipboard.Clipboard
	CommandHandler runtime.CommandHandler
	// MCP enables the MCP agent server on the given unix socket path,
	// allowing fluffy-speak and other tools to connect. Empty string disables.
	MCP string
	// TTS enables text-to-speech using the best available platform engine.
	TTS bool
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
	if opts.MCP != "" {
		options = append(options, fluffy.WithMCP(opts.MCP))
	}
	if opts.TTS {
		options = append(options, fluffy.WithTTS())
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
