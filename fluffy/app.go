package fluffy

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/audio"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/backend/sim"
	backendtcell "github.com/odvcencio/fluffy-ui/backend/tcell"
	"github.com/odvcencio/fluffy-ui/clipboard"
	"github.com/odvcencio/fluffy-ui/keybind"
	"github.com/odvcencio/fluffy-ui/recording"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/style"
	"github.com/odvcencio/fluffy-ui/theme"
)

// Bundle exposes default wiring plus keybinding helpers.
type Bundle struct {
	App      *runtime.App
	Registry *keybind.CommandRegistry
	Keymaps  *keybind.KeymapStack
	Router   *keybind.KeyRouter
}

// AppOption customizes the default app wiring.
type AppOption func(*appBuilder)

type appBuilder struct {
	cfg      runtime.AppConfig
	registry *keybind.CommandRegistry
	keymaps  *keybind.KeymapStack
	router   *keybind.KeyRouter
}

// NewApp creates a default app with sensible defaults and panics on setup error.
func NewApp(opts ...AppOption) *runtime.App {
	app, err := NewAppWithError(opts...)
	if err != nil {
		panic(err)
	}
	return app
}

// NewAppWithError creates a default app and returns any setup error.
func NewAppWithError(opts ...AppOption) (*runtime.App, error) {
	bundle, err := NewBundle(opts...)
	if err != nil {
		return nil, err
	}
	return bundle.App, nil
}

// NewBundle builds an app plus keybinding helpers.
func NewBundle(opts ...AppOption) (*Bundle, error) {
	builder, err := newBuilder()
	if err != nil {
		return nil, err
	}
	for _, opt := range opts {
		if opt != nil {
			opt(builder)
		}
	}
	app := runtime.NewApp(builder.cfg)
	return &Bundle{
		App:      app,
		Registry: builder.registry,
		Keymaps:  builder.keymaps,
		Router:   builder.router,
	}, nil
}

// DefaultConfig returns the default app config used by NewApp.
func DefaultConfig() (runtime.AppConfig, error) {
	builder, err := newBuilder()
	if err != nil {
		return runtime.AppConfig{}, err
	}
	return builder.cfg, nil
}

func newBuilder() (*appBuilder, error) {
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

	indicator := "> "
	focusStyle := backend.DefaultStyle().Bold(true)
	announcer := accessibility.Announcer(&accessibility.SimpleAnnouncer{})
	clip := clipboard.Clipboard(&clipboard.MemoryClipboard{})
	tick := time.Second / 30
	sheet := theme.DefaultStylesheet()

	recorder, err := buildRecorderFromEnv()
	if err != nil {
		return nil, err
	}

	cfg := runtime.AppConfig{
		Backend:           be,
		TickRate:          tick,
		KeyHandler:        keyHandler,
		Announcer:         announcer,
		Clipboard:         clip,
		Stylesheet:        sheet,
		Recorder:          recorder,
		FocusRegistration: runtime.FocusRegistrationAuto,
		FocusStyle: &accessibility.FocusStyle{
			Indicator: indicator,
			Style:     focusStyle,
		},
	}

	return &appBuilder{
		cfg:      cfg,
		registry: registry,
		keymaps:  stack,
		router:   router,
	}, nil
}

// WithBackend overrides the backend.
func WithBackend(be backend.Backend) AppOption {
	return func(b *appBuilder) {
		if b == nil || be == nil {
			return
		}
		b.cfg.Backend = be
	}
}

// WithRoot sets the root widget.
func WithRoot(root runtime.Widget) AppOption {
	return func(b *appBuilder) {
		if b == nil {
			return
		}
		b.cfg.Root = root
	}
}

// WithTickRate overrides the tick rate.
func WithTickRate(rate time.Duration) AppOption {
	return func(b *appBuilder) {
		if b == nil {
			return
		}
		b.cfg.TickRate = rate
	}
}

// WithAnnouncer overrides the accessibility announcer.
func WithAnnouncer(announcer accessibility.Announcer) AppOption {
	return func(b *appBuilder) {
		if b == nil {
			return
		}
		b.cfg.Announcer = announcer
	}
}

// WithClipboard overrides the clipboard implementation.
func WithClipboard(clip clipboard.Clipboard) AppOption {
	return func(b *appBuilder) {
		if b == nil {
			return
		}
		b.cfg.Clipboard = clip
	}
}

// WithStylesheet overrides the stylesheet.
func WithStylesheet(sheet *style.Stylesheet) AppOption {
	return func(b *appBuilder) {
		if b == nil {
			return
		}
		b.cfg.Stylesheet = sheet
	}
}

// WithFocusIndicator overrides the focus indicator string.
func WithFocusIndicator(indicator string) AppOption {
	return func(b *appBuilder) {
		if b == nil {
			return
		}
		ensureFocusStyle(&b.cfg)
		b.cfg.FocusStyle.Indicator = indicator
	}
}

// WithFocusStyle overrides the focus style.
func WithFocusStyle(style backend.Style) AppOption {
	return func(b *appBuilder) {
		if b == nil {
			return
		}
		ensureFocusStyle(&b.cfg)
		b.cfg.FocusStyle.Style = style
	}
}

// WithRecorder overrides the recorder.
func WithRecorder(rec runtime.Recorder) AppOption {
	return func(b *appBuilder) {
		if b == nil {
			return
		}
		b.cfg.Recorder = rec
	}
}

// WithAudio overrides the audio service.
func WithAudio(service audio.Service) AppOption {
	return func(b *appBuilder) {
		if b == nil {
			return
		}
		b.cfg.Audio = service
	}
}

// WithCommandHandler overrides the command handler.
func WithCommandHandler(handler runtime.CommandHandler) AppOption {
	return func(b *appBuilder) {
		if b == nil {
			return
		}
		b.cfg.CommandHandler = handler
	}
}

// WithUpdate overrides the update loop.
func WithUpdate(update runtime.UpdateFunc) AppOption {
	return func(b *appBuilder) {
		if b == nil {
			return
		}
		b.cfg.Update = update
	}
}

// WithFocusRegistration overrides focus registration behavior.
func WithFocusRegistration(mode runtime.FocusRegistrationMode) AppOption {
	return func(b *appBuilder) {
		if b == nil {
			return
		}
		b.cfg.FocusRegistration = mode
	}
}

// WithKeyHandler overrides the key handler.
func WithKeyHandler(handler runtime.KeyHandler) AppOption {
	return func(b *appBuilder) {
		if b == nil {
			return
		}
		b.cfg.KeyHandler = handler
		b.router = nil
	}
}

// WithCommandRegistry replaces the command registry and rebuilds the router.
func WithCommandRegistry(registry *keybind.CommandRegistry) AppOption {
	return func(b *appBuilder) {
		if b == nil {
			return
		}
		b.registry = registry
		b.rebuildKeyHandler()
	}
}

// WithKeymap replaces the default keymap.
func WithKeymap(keymap *keybind.Keymap) AppOption {
	return func(b *appBuilder) {
		if b == nil || keymap == nil {
			return
		}
		stack := &keybind.KeymapStack{}
		stack.Push(keymap)
		b.keymaps = stack
		b.rebuildKeyHandler()
	}
}

// WithKeymapStack replaces the keymap stack.
func WithKeymapStack(stack *keybind.KeymapStack) AppOption {
	return func(b *appBuilder) {
		if b == nil {
			return
		}
		b.keymaps = stack
		b.rebuildKeyHandler()
	}
}

// WithKeyBindings lets callers register additional commands before building the router.
func WithKeyBindings(register func(*keybind.CommandRegistry)) AppOption {
	return func(b *appBuilder) {
		if b == nil || register == nil {
			return
		}
		if b.registry == nil {
			b.registry = keybind.NewRegistry()
		}
		register(b.registry)
		b.rebuildKeyHandler()
	}
}

func (b *appBuilder) rebuildKeyHandler() {
	if b == nil {
		return
	}
	if b.registry == nil || b.keymaps == nil {
		b.cfg.KeyHandler = nil
		b.router = nil
		return
	}
	b.router = keybind.NewKeyRouter(b.registry, nil, b.keymaps)
	b.cfg.KeyHandler = &keybind.RuntimeHandler{Router: b.router}
}

func ensureFocusStyle(cfg *runtime.AppConfig) {
	if cfg == nil {
		return
	}
	if cfg.FocusStyle == nil {
		cfg.FocusStyle = &accessibility.FocusStyle{
			Indicator: "> ",
			Style:     backend.DefaultStyle().Bold(true),
		}
	}
}

func buildBackendFromEnv() (backend.Backend, error) {
	backendName := strings.ToLower(strings.TrimSpace(os.Getenv("FLUFFYUI_BACKEND")))
	switch backendName {
	case "sim", "simulation":
		width := envInt("FLUFFYUI_WIDTH", 80)
		height := envInt("FLUFFYUI_HEIGHT", 24)
		return sim.New(width, height), nil
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
