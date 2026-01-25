package demo

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/backend/ghostty"
	"github.com/odvcencio/fluffy-ui/backend/sim"
	backendtcell "github.com/odvcencio/fluffy-ui/backend/tcell"
	"github.com/odvcencio/fluffy-ui/clipboard"
	"github.com/odvcencio/fluffy-ui/keybind"
	"github.com/odvcencio/fluffy-ui/recording"
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
	be, err := buildBackendFromEnv()
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

	recorder, err := buildRecorderFromEnv()
	if err != nil {
		return nil, err
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
		Recorder: recorder,
	})

	return &Bundle{
		App:      app,
		Registry: registry,
		Keymaps:  stack,
		Router:   router,
	}, nil
}

func buildBackendFromEnv() (backend.Backend, error) {
	backendName := strings.ToLower(strings.TrimSpace(os.Getenv("FLUFFYUI_BACKEND")))
	switch backendName {
	case "sim", "simulation":
		width := envInt("FLUFFYUI_WIDTH", 80)
		height := envInt("FLUFFYUI_HEIGHT", 24)
		return sim.New(width, height), nil
	case "ghostty":
		return ghostty.New()
	}
	return backendtcell.New()
}

func buildRecorderFromEnv() (runtime.Recorder, error) {
	recordPath := strings.TrimSpace(os.Getenv("FLUFFYUI_RECORD"))
	exportPath := strings.TrimSpace(os.Getenv("FLUFFYUI_RECORD_EXPORT"))
	if recordPath == "" && exportPath == "" {
		return nil, nil
	}

	title := strings.TrimSpace(os.Getenv("FLUFFYUI_RECORD_TITLE"))
	if title == "" {
		title = "FluffyUI Demo"
	}

	opts := recording.AsciicastOptions{Title: title}
	if exportPath != "" {
		return recording.NewVideoRecorder(exportPath, recording.VideoRecorderOptions{
			Cast:     opts,
			CastPath: recordPath,
			KeepCast: recordPath != "",
			Video: recording.VideoOptions{
				Agg: recording.AggOptions{
					Theme:    "monokai",
					FontSize: 16,
					FPS:      30,
				},
				FFmpeg: recording.FFmpegOptions{
					VideoCodec: "libx264",
					Preset:     "medium",
					CRF:        22,
				},
			},
		})
	}

	if recordPath == "" {
		return nil, fmt.Errorf("FLUFFYUI_RECORD is required when FLUFFYUI_RECORD_EXPORT is unset")
	}
	return recording.NewAsciicastRecorder(recordPath, opts)
}

func envInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
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
