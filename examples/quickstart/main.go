package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/odvcencio/fluffyui/audio"
	"github.com/odvcencio/fluffyui/audio/execdriver"
	"github.com/odvcencio/fluffyui/fluffy"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	"github.com/odvcencio/fluffyui/widgets"
)

func main() {
	recordPath := os.Getenv("FLUFFYUI_RECORD")
	exportPath := os.Getenv("FLUFFYUI_RECORD_EXPORT")
	audioService, audioStatus := setupAudio()
	app, err := fluffy.NewApp(fluffy.WithAudio(audioService))
	if err != nil {
		fmt.Fprintf(os.Stderr, "app init failed: %v\n", err)
		os.Exit(1)
	}

	count := state.NewSignal(0)
	count.SetEqualFunc(state.EqualComparable[int])

	mode := state.NewSignal("manual")
	mode.SetEqualFunc(state.EqualComparable[string])

	root := NewCounterView(count, mode, recordPath, exportPath, audioStatus)
	app.SetRoot(root)

	app.Every(400*time.Millisecond, func(now time.Time) runtime.Message {
		if mode != nil && mode.Get() == "auto" {
			count.Update(func(v int) int { return v + 1 })
		}
		return nil
	})

	if err := app.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
		os.Exit(1)
	}
}

type CounterView struct {
	widgets.Component
	count        *state.Signal[int]
	mode         *state.Signal[string]
	countValue   int
	modeValue    string
	recordPath   string
	exportPath   string
	spinnerIndex int
	spinner      []rune
	audio        audio.Service
	audioStatus  string
	audioMuted   bool
}

func NewCounterView(count *state.Signal[int], mode *state.Signal[string], recordPath string, exportPath string, audioStatus string) *CounterView {
	view := &CounterView{
		count:       count,
		mode:        mode,
		recordPath:  recordPath,
		exportPath:  exportPath,
		spinner:     []rune{'|', '/', '-', '\\'},
		audioStatus: audioStatus,
	}
	view.refresh()
	return view
}

func (c *CounterView) Bind(services runtime.Services) {
	c.Component.Bind(services)
	c.audio = services.Audio()
	if c.audio != nil {
		c.audioMuted = c.audio.Muted()
	}
}

func (c *CounterView) Mount() {
	c.Observe(c.count, c.refresh)
	c.Observe(c.mode, c.refresh)
	c.refresh()
}

func (c *CounterView) Unmount() {
	c.Subs.Clear()
}

func (c *CounterView) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (c *CounterView) Layout(bounds runtime.Rect) {
	c.Component.Layout(bounds)
}

func (c *CounterView) Render(ctx runtime.RenderContext) {
	if ctx.Buffer == nil {
		return
	}
	bounds := c.Bounds()
	if bounds.Width == 0 || bounds.Height == 0 {
		return
	}
	ctx.Clear(fluffy.DefaultStyle())

	frame := c.spinner[c.spinnerIndex%len(c.spinner)]
	lines := []string{
		"[" + string(frame) + "] FluffyUI Quickstart",
		"",
		"Count: " + strconv.Itoa(c.countValue),
		"Mode: " + c.modeValue,
		"",
		c.audioLine(),
		"Keys: +/- to change, m to toggle auto, s to mute, q or Ctrl+C to quit",
	}
	if c.recordPath != "" {
		lines = append(lines, "Recording: "+c.recordPath)
	}
	if c.exportPath != "" {
		lines = append(lines, "Export: "+c.exportPath)
	}

	for i, line := range lines {
		if i >= bounds.Height {
			break
		}
		if len(line) > bounds.Width {
			line = line[:bounds.Width]
		}
		ctx.Buffer.SetString(bounds.X, bounds.Y+i, line, fluffy.DefaultStyle())
	}
}

func (c *CounterView) HandleMessage(msg runtime.Message) runtime.HandleResult {
	switch m := msg.(type) {
	case runtime.KeyMsg:
		switch m.Rune {
		case 'q':
			return runtime.WithCommand(runtime.Quit{})
		case '+', '=':
			c.updateCount(1)
			return runtime.Handled()
		case '-':
			c.updateCount(-1)
			return runtime.Handled()
		case 'm':
			c.toggleMode()
			return runtime.Handled()
		case 's':
			c.toggleMute()
			return runtime.Handled()
		}
	case runtime.TickMsg:
		if len(c.spinner) > 0 {
			c.spinnerIndex = (c.spinnerIndex + 1) % len(c.spinner)
			c.Invalidate()
			return runtime.Handled()
		}
	}
	return runtime.Unhandled()
}

func (c *CounterView) refresh() {
	if c.count != nil {
		c.countValue = c.count.Get()
	}
	if c.mode != nil {
		c.modeValue = c.mode.Get()
	}
}

func (c *CounterView) updateCount(delta int) {
	if c.count == nil {
		return
	}
	c.count.Update(func(v int) int { return v + delta })
	if delta > 0 {
		c.playSFX("ui.up")
	} else if delta < 0 {
		c.playSFX("ui.down")
	}
}

func (c *CounterView) toggleMode() {
	if c.mode == nil {
		return
	}
	if c.mode.Get() == "auto" {
		c.mode.Set("manual")
		c.playSFX("ui.toggle")
		c.stopMusic()
		return
	}
	c.mode.Set("auto")
	c.playSFX("ui.toggle")
	c.playMusic("music.loop")
}

func (c *CounterView) toggleMute() {
	if c.audio == nil {
		return
	}
	muted := c.audio.Muted()
	c.audio.SetMuted(!muted)
	c.audioMuted = !muted
	c.Invalidate()
}

func (c *CounterView) playSFX(id string) {
	if c.audio == nil {
		return
	}
	c.audio.PlaySFX(id)
}

func (c *CounterView) playMusic(id string) {
	if c.audio == nil {
		return
	}
	c.audio.PlayMusic(id)
}

func (c *CounterView) stopMusic() {
	if c.audio == nil {
		return
	}
	c.audio.StopMusic()
}

func (c *CounterView) audioLine() string {
	line := "Audio: " + c.audioStatus
	if c.audio == nil || !strings.HasPrefix(c.audioStatus, "enabled") {
		return line
	}
	if c.audioMuted {
		return line + " [muted]"
	}
	return line + " [on]"
}

func setupAudio() (audio.Service, string) {
	assetsEnv := strings.TrimSpace(os.Getenv("FLUFFYUI_AUDIO_ASSETS"))
	if audioDisabled(assetsEnv) {
		return audio.Disabled{}, "disabled (env off)"
	}
	assetsDir := assetsEnv
	sourceLabel := "custom"
	if assetsDir == "" {
		assetsDir = defaultAudioAssetsDir()
		sourceLabel = "sample"
	}
	if assetsDir == "" {
		return audio.Disabled{}, "disabled (set FLUFFYUI_AUDIO_ASSETS)"
	}
	command, ok := execdriver.DetectCommand()
	if !ok {
		return audio.Disabled{}, "disabled (no audio command found)"
	}
	addSource := func(sources map[string]execdriver.Source, cues *[]audio.Cue, id string, filename string, cue audio.Cue) {
		path := filepath.Join(assetsDir, filename)
		if !fileExists(path) {
			return
		}
		sources[id] = execdriver.Source{Path: path}
		*cues = append(*cues, cue)
	}
	sources := make(map[string]execdriver.Source)
	cues := make([]audio.Cue, 0, 4)
	addSource(sources, &cues, "ui.up", "count-up.wav", audio.Cue{
		ID:       "ui.up",
		Kind:     audio.KindSFX,
		Volume:   80,
		Cooldown: 60 * time.Millisecond,
	})
	addSource(sources, &cues, "ui.down", "count-down.wav", audio.Cue{
		ID:       "ui.down",
		Kind:     audio.KindSFX,
		Volume:   80,
		Cooldown: 60 * time.Millisecond,
	})
	addSource(sources, &cues, "ui.toggle", "toggle.wav", audio.Cue{
		ID:       "ui.toggle",
		Kind:     audio.KindSFX,
		Volume:   70,
		Cooldown: 120 * time.Millisecond,
	})
	addSource(sources, &cues, "music.loop", "music.wav", audio.Cue{
		ID:     "music.loop",
		Kind:   audio.KindMusic,
		Volume: 50,
		Loop:   true,
	})
	if len(cues) == 0 {
		return audio.Disabled{}, "disabled (no assets found)"
	}
	driver := execdriver.NewDriver(execdriver.Config{
		Command: command,
		Sources: sources,
	})
	return audio.NewManager(driver, cues...), fmt.Sprintf("enabled (%s, %s)", command.Path, sourceLabel)
}

func fileExists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func defaultAudioAssetsDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	candidates := []string{
		filepath.Join(cwd, "examples", "quickstart", "assets", "audio"),
		filepath.Join(cwd, "assets", "audio"),
	}
	for _, candidate := range candidates {
		if dirExists(candidate) {
			return candidate
		}
	}
	return ""
}

func audioDisabled(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "0", "off", "false", "no", "disable", "disabled":
		return true
	default:
		return false
	}
}
