package runtime

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/animation"
	"github.com/odvcencio/fluffyui/audio"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/clipboard"
	"github.com/odvcencio/fluffyui/i18n"
	"github.com/odvcencio/fluffyui/state"
	"github.com/odvcencio/fluffyui/style"
	"github.com/odvcencio/fluffyui/theme"
)

type stubAudio struct{}

func (stubAudio) Play(id string) bool         { return true }
func (stubAudio) PlaySFX(id string) bool      { return true }
func (stubAudio) PlayMusic(id string) bool    { return true }
func (stubAudio) StopMusic() bool             { return true }
func (stubAudio) SetMuted(muted bool)         {}
func (stubAudio) Muted() bool                 { return false }
func (stubAudio) SetMasterVolume(percent int) {}
func (stubAudio) SetSFXVolume(percent int)    {}
func (stubAudio) SetMusicVolume(percent int)  {}

func TestAppAccessorsAndCommands(t *testing.T) {
	bundle := i18n.NewBundle("en")
	bundle.AddMessages("en", map[string]string{"title": "Fluffy"})
	localizer := bundle.Localizer("en")
	anim := animation.NewAnimator()
	queue := state.NewQueue()
	th := theme.DefaultTheme()

	app := NewApp(AppConfig{
		Theme:      th,
		Localizer:  localizer,
		Animator:   anim,
		StateQueue: queue,
	})
	if app.StateQueue() == nil {
		t.Fatalf("expected state queue")
	}
	if app.Stylesheet() == nil {
		t.Fatalf("expected stylesheet from theme")
	}
	if app.Theme() == nil {
		t.Fatalf("expected theme")
	}
	if app.Localizer() == nil {
		t.Fatalf("expected localizer")
	}
	if app.Animator() == nil {
		t.Fatalf("expected animator")
	}

	app.screen = NewScreen(10, 5)
	sheet := style.NewStylesheet()
	app.SetLocalizer(localizer)
	app.SetStylesheet(sheet)
	app.SetTheme(nil)
	app.SetTheme(th)
	app.Relayout()
	_ = app.StateScheduler()
	_ = app.InvalidateScheduler()

	root := &appTestWidget{}
	app.SetRoot(root)
	app.PostQueueFlush()
	select {
	case <-app.messages:
	default:
		t.Fatalf("expected queue flush message")
	}

	if !app.TryPost(InvalidateMsg{}) {
		t.Fatalf("expected TryPost to succeed")
	}

	small := NewApp(AppConfig{MessageBuffer: 1})
	if !small.TryPost(InvalidateMsg{}) {
		t.Fatalf("expected first TryPost to succeed")
	}
	if small.TryPost(InvalidateMsg{}) {
		t.Fatalf("expected TryPost to fail when buffer full")
	}

	app.running.Store(true)
	app.ExecuteCommand(Refresh{})
	if app.screen == nil || !app.screen.Buffer().IsDirty() {
		t.Fatalf("expected Refresh to mark buffer dirty")
	}
	app.ExecuteCommand(SendMsg{Message: InvalidateMsg{}})

	done := make(chan struct{}, 1)
	app.ExecuteCommand(Effect{Run: func(ctx context.Context, post PostFunc) {
		done <- struct{}{}
	}})
	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("expected effect to run")
	}

	app.ExecuteCommand(Quit{})
	app.After(0, InvalidateMsg{})
	app.Every(time.Millisecond, func(time.Time) Message { return nil })
}

func TestServicesAccessors(t *testing.T) {
	announcer := &accessibility.SimpleAnnouncer{}
	clip := &clipboard.MemoryClipboard{}
	focusStyle := &accessibility.FocusStyle{Indicator: ">", Style: backend.DefaultStyle()}

	app := NewApp(AppConfig{
		Announcer:  announcer,
		Clipboard:  clip,
		Audio:      stubAudio{},
		FocusStyle: focusStyle,
	})

	services := app.Services()
	if services.Announcer() == nil {
		t.Fatalf("expected announcer")
	}
	if services.Clipboard() == nil {
		t.Fatalf("expected clipboard")
	}
	if services.Audio() == nil {
		t.Fatalf("expected audio service")
	}
	if services.FocusStyle() == nil {
		t.Fatalf("expected focus style")
	}

	services.Invalidate()
	services.Relayout()
	if !services.Post(InvalidateMsg{}) {
		t.Fatalf("expected post to succeed")
	}
	services.After(0, InvalidateMsg{})
	services.Every(time.Millisecond, func(time.Time) Message { return nil })
	services.Spawn(Effect{Run: func(ctx context.Context, post PostFunc) {}})
}

func TestAppMCPEnable(t *testing.T) {
	RegisterMCPEnabler(func(app *App, opts MCPOptions) (io.Closer, error) {
		return io.NopCloser(strings.NewReader("")), nil
	})
	defer RegisterMCPEnabler(nil)

	app := NewApp(AppConfig{})
	closer, err := app.EnableMCP(MCPOptions{Transport: "stdio"})
	if err != nil {
		t.Fatalf("EnableMCP error: %v", err)
	}
	if closer == nil {
		t.Fatalf("expected closer")
	}
}

var _ audio.Service = stubAudio{}
