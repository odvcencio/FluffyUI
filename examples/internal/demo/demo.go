package demo

import (
	"time"

	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/backend"
	backendtcell "github.com/odvcencio/fluffy-ui/backend/tcell"
	"github.com/odvcencio/fluffy-ui/clipboard"
	"github.com/odvcencio/fluffy-ui/keybind"
	"github.com/odvcencio/fluffy-ui/runtime"
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
	be, err := backendtcell.New()
	if err != nil {
		return nil, err
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

	indicator := opts.FocusIndicator
	if indicator == "" {
		indicator = "> "
	}
	focusStyle := backend.DefaultStyle().Bold(true)
	if opts.FocusStyle != nil {
		focusStyle = *opts.FocusStyle
	}

	announcer := opts.Announcer
	if announcer == nil {
		announcer = &accessibility.SimpleAnnouncer{}
	}
	clip := opts.Clipboard
	if clip == nil {
		clip = &clipboard.MemoryClipboard{}
	}

	tick := opts.TickRate
	if tick <= 0 {
		tick = time.Second / 30
	}

	update := runtime.DefaultUpdate
	if root != nil {
		update = withFocusRegistration(root, update)
	}

	app := runtime.NewApp(runtime.AppConfig{
		Backend:        be,
		Root:           root,
		Update:         update,
		CommandHandler: opts.CommandHandler,
		TickRate:       tick,
		KeyHandler:     keyHandler,
		Announcer:      announcer,
		Clipboard:      clip,
		FocusStyle: &accessibility.FocusStyle{
			Indicator: indicator,
			Style:     focusStyle,
		},
	})

	return &Bundle{
		App:      app,
		Registry: registry,
		Keymaps:  stack,
		Router:   router,
	}, nil
}

func withFocusRegistration(root runtime.Widget, next runtime.UpdateFunc) runtime.UpdateFunc {
	registered := false
	return func(app *runtime.App, msg runtime.Message) bool {
		if !registered && app != nil {
			if screen := app.Screen(); screen != nil {
				runtime.RegisterFocusables(screen.FocusScope(), root)
				registered = true
			}
		}
		if next == nil {
			return runtime.DefaultUpdate(app, msg)
		}
		return next(app, msg)
	}
}
