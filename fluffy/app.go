package fluffy

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/accessibility/tts"
	"github.com/odvcencio/fluffyui/audio"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/backend/sim"
	backendtcell "github.com/odvcencio/fluffyui/backend/tcell"
	"github.com/odvcencio/fluffyui/backend/web"
	"github.com/odvcencio/fluffyui/clipboard"
	"github.com/odvcencio/fluffyui/i18n"
	"github.com/odvcencio/fluffyui/keybind"
	"github.com/odvcencio/fluffyui/recording"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/style"
	"github.com/odvcencio/fluffyui/theme"
	"github.com/odvcencio/fluffyui/toast"
	"github.com/odvcencio/fluffyui/widgets"
)

// Bundle exposes default wiring plus keybinding helpers.
type Bundle struct {
	App        *runtime.App
	Registry   *keybind.CommandRegistry
	Keymaps    *keybind.KeymapStack
	Router     *keybind.KeyRouter
	ToastStack *widgets.ToastStack
}

// AppOption customizes the default app wiring.
type AppOption func(*appBuilder)

type appBuilder struct {
	cfg          runtime.AppConfig
	registry     *keybind.CommandRegistry
	keymaps      *keybind.KeymapStack
	router       *keybind.KeyRouter
	toastManager *toast.ToastManager
}

// NewApp creates a default app with sensible defaults.
// Returns an error if setup fails.
func NewApp(opts ...AppOption) (*runtime.App, error) {
	return NewAppWithError(opts...)
}

// NewAppWithError creates a default app and returns any setup error.
func NewAppWithError(opts ...AppOption) (*runtime.App, error) {
	bundle, err := NewBundle(opts...)
	if err != nil {
		return nil, err
	}
	return bundle.App, nil
}

// Run builds and runs an app with the provided root widget.
// It uses context.Background(); prefer RunContext for cancellation.
func Run(root runtime.Widget, opts ...AppOption) error {
	return RunContext(context.Background(), root, opts...)
}

// RunContext builds and runs an app with the provided root widget and context.
func RunContext(ctx context.Context, root runtime.Widget, opts ...AppOption) error {
	merged := make([]AppOption, 0, len(opts)+1)
	merged = append(merged, opts...)
	merged = append(merged, WithRoot(root))

	app, err := NewApp(merged...)
	if err != nil {
		return err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return app.Run(ctx)
}

// RunInline builds and runs an app in inline mode with a bounded viewport.
// The app stays in the primary terminal screen (no alternate-screen takeover).
// It uses context.Background(); prefer RunInlineContext for cancellation.
func RunInline(root runtime.Widget, lines int, opts ...AppOption) error {
	return RunInlineContext(context.Background(), root, lines, opts...)
}

// RunInlineContext builds and runs an app in inline mode with context support.
func RunInlineContext(ctx context.Context, root runtime.Widget, lines int, opts ...AppOption) error {
	merged := make([]AppOption, 0, len(opts)+2)
	merged = append(merged, opts...)
	merged = append(merged, WithInlineMode(true), WithInlineHeight(lines))
	return RunContext(ctx, root, merged...)
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

	var toastStack *widgets.ToastStack
	if builder.toastManager != nil {
		toastStack = widgets.NewToastStack()
		manager := builder.toastManager
		originalOnReady := builder.cfg.OnReady
		builder.cfg.OnReady = func(a *runtime.App) {
			if screen := a.Screen(); screen != nil {
				screen.PushLayer(toastStack, false)
			}
			manager.SetOnChange(func(toasts []*toast.Toast) {
				toastStack.SetToasts(toasts)
				a.Invalidate()
			})
			if originalOnReady != nil {
				originalOnReady(a)
			}
		}
	}

	app := runtime.NewApp(builder.cfg)
	return &Bundle{
		App:        app,
		Registry:   builder.registry,
		Keymaps:    builder.keymaps,
		Router:     builder.router,
		ToastStack: toastStack,
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
	clip := clipboard.Clipboard(clipboard.NewAutoClipboard(os.Stdout))
	tick := time.Second / 30
	sheet := theme.DefaultStylesheet()

	recorder, err := buildRecorderFromEnv()
	if err != nil {
		return nil, err
	}

	speaker := buildSpeakerFromEnv()

	cfg := runtime.AppConfig{
		Backend:           be,
		TickRate:          tick,
		KeyHandler:        keyHandler,
		Announcer:         announcer,
		Clipboard:         clip,
		Stylesheet:        sheet,
		Recorder:          recorder,
		Speaker:           speaker,
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

// WithInlineMode enables or disables inline rendering mode for backends that
// support it. In inline mode, the app avoids alternate-screen takeover.
func WithInlineMode(enabled bool) AppOption {
	return func(b *appBuilder) {
		if b == nil {
			return
		}
		b.cfg.InlineMode = enabled
	}
}

// WithInlineHeight sets the number of terminal rows reserved for inline mode.
// Values <= 0 use the backend default.
func WithInlineHeight(lines int) AppOption {
	return func(b *appBuilder) {
		if b == nil {
			return
		}
		b.cfg.InlineHeight = lines
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

// WithFrameBudget sets a render frame budget for animation updates.
func WithFrameBudget(budget time.Duration) AppOption {
	return func(b *appBuilder) {
		if b == nil {
			return
		}
		b.cfg.FrameBudget = budget
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

// WithLocalizer overrides the app localizer.
func WithLocalizer(localizer i18n.Localizer) AppOption {
	return func(b *appBuilder) {
		if b == nil {
			return
		}
		b.cfg.Localizer = localizer
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

// WithAutoFocusPolicy sets the auto-focus policy for focus scopes.
// AutoFocusFirst (default) focuses the first focusable widget.
// AutoFocusLast focuses the last focusable widget.
// AutoFocusNone disables auto-focus; apps must call SetFocus explicitly.
func WithAutoFocusPolicy(policy runtime.AutoFocusPolicy) AppOption {
	return func(b *appBuilder) {
		if b == nil {
			return
		}
		b.cfg.AutoFocusPolicy = policy
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

// WithWebServer configures the app to use the web backend.
// The app will be accessible via browser at the specified address.
// Use ":0" to let the system assign a random port.
//
// Example:
//
//	app, err := fluffy.NewApp(fluffy.WithWebServer(":8080"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Open", app.Backend().(*web.Backend).URL())
func WithWebServer(addr string) AppOption {
	return func(b *appBuilder) {
		if b == nil {
			return
		}
		b.cfg.Backend = web.NewWithAddr(addr)
	}
}

// WithWebConfig configures the app to use the web backend with full configuration.
func WithWebConfig(config web.Config) AppOption {
	return func(b *appBuilder) {
		if b == nil {
			return
		}
		b.cfg.Backend = web.New(config)
	}
}

// WithOnReady registers a callback invoked once during Run, after the Screen
// is initialized but before the event loop starts.
func WithOnReady(fn func(app *runtime.App)) AppOption {
	return func(b *appBuilder) {
		if b == nil || fn == nil {
			return
		}
		b.cfg.OnReady = fn
	}
}

// WithOnResize registers a callback invoked when the app size is known and
// whenever the terminal is resized.
func WithOnResize(fn func(app *runtime.App, width, height int)) AppOption {
	return func(b *appBuilder) {
		if b == nil || fn == nil {
			return
		}
		b.cfg.OnResize = fn
	}
}

// WithOnQuit registers a callback invoked once when the app event loop exits.
func WithOnQuit(fn func(app *runtime.App)) AppOption {
	return func(b *appBuilder) {
		if b == nil || fn == nil {
			return
		}
		b.cfg.OnQuit = fn
	}
}

// WithMCP enables the MCP agent server on the given unix socket path,
// allowing tools like fluffy-speak to connect and read ARIA semantics.
// An empty addr defaults to /tmp/fluffyui.sock.
func WithMCP(addr string) AppOption {
	return func(b *appBuilder) {
		if b == nil {
			return
		}
		if addr == "" {
			addr = "/tmp/fluffyui.sock"
		}
		b.cfg.MCPOptions = &runtime.MCPOptions{
			Transport: "unix",
			Addr:      addr,
		}
	}
}

// WithTTS enables text-to-speech using the best available platform engine.
// If no TTS engine is found, the app runs silently with no error.
func WithTTS() AppOption {
	return func(b *appBuilder) {
		if b == nil {
			return
		}
		speaker, _ := tts.Auto()
		if speaker != nil {
			b.cfg.Speaker = speaker
		}
	}
}

// WithTTSBackend forces a specific TTS backend by name.
// Valid names: "espeak", "ssip", "say", "sapi", "mock".
func WithTTSBackend(name string) AppOption {
	return func(b *appBuilder) {
		if b == nil {
			return
		}
		speaker, err := tts.Auto(tts.WithBackend(name))
		if err != nil {
			return
		}
		b.cfg.Speaker = speaker
	}
}

// WithSpeaker sets a custom Speaker implementation for TTS.
func WithSpeaker(s accessibility.Speaker) AppOption {
	return func(b *appBuilder) {
		if b == nil || s == nil {
			return
		}
		b.cfg.Speaker = s
	}
}

// WithToastLayer configures the bundle to push a ToastStack as a non-modal
// overlay layer on startup, wired to the given ToastManager. The resulting
// ToastStack is available via Bundle.ToastStack for styling.
func WithToastLayer(manager *toast.ToastManager) AppOption {
	return func(b *appBuilder) {
		if b == nil || manager == nil {
			return
		}
		b.toastManager = manager
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
	case "web":
		addr := os.Getenv("FLUFFYUI_WEB_ADDR")
		if addr == "" {
			addr = ":8080"
		}
		return web.NewWithAddr(addr), nil
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

// buildSpeakerFromEnv checks FLUFFYUI_TTS for TTS configuration.
// Values: "1"/"true" = auto-detect, or a backend name like "sapi", "espeak", "ssip", "say".
// Returns nil if unset or no engine is available (app runs silently).
func buildSpeakerFromEnv() accessibility.Speaker {
	val := strings.TrimSpace(os.Getenv("FLUFFYUI_TTS"))
	if val == "" {
		return nil
	}
	var opts []tts.Option
	switch strings.ToLower(val) {
	case "1", "true", "yes", "on", "auto":
		// auto-detect
	default:
		opts = append(opts, tts.WithBackend(val))
	}
	speaker, _ := tts.Auto(opts...)
	return speaker
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
