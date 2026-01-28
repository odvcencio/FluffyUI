// Package execdriver provides a command-based audio driver.
package execdriver

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/odvcencio/fluffyui/audio"
)

// Command describes the command used to play an audio cue.
// Use {{path}} and {{volume}} placeholders inside args.
type Command struct {
	Path string
	Args []string
}

// Source maps a cue ID to a file path and optional command override.
type Source struct {
	Path    string
	Command Command
}

// Config configures a command driver.
type Config struct {
	Command Command
	Sources map[string]Source
}

// Driver executes external commands to play audio.
type Driver struct {
	mu          sync.Mutex
	command     Command
	sources     map[string]Source
	musicCmd    *exec.Cmd
	musicStopCh chan struct{}
}

// NewDriver creates a command-based audio driver.
func NewDriver(cfg Config) *Driver {
	driver := &Driver{
		command: cfg.Command,
		sources: make(map[string]Source),
	}
	for id, src := range cfg.Sources {
		if id == "" {
			continue
		}
		driver.sources[id] = src
	}
	return driver
}

// Register adds or replaces a cue source.
func (d *Driver) Register(id string, src Source) {
	if d == nil || id == "" {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.sources == nil {
		d.sources = make(map[string]Source)
	}
	d.sources[id] = src
}

// Play starts a command for the cue.
func (d *Driver) Play(cue audio.Cue) error {
	if d == nil {
		return nil
	}
	if cue.Kind == audio.KindMusic {
		_ = d.Stop(audio.KindMusic)
		return d.startMusic(cue)
	}
	cmd, err := d.buildCommand(cue)
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go cmd.Wait()
	return nil
}

// Stop stops music playback.
func (d *Driver) Stop(kind audio.Kind) error {
	if d == nil || kind != audio.KindMusic {
		return nil
	}
	d.mu.Lock()
	cmd := d.musicCmd
	stopCh := d.musicStopCh
	d.musicCmd = nil
	d.musicStopCh = nil
	d.mu.Unlock()

	if stopCh != nil {
		close(stopCh)
	}
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
	return nil
}

func (d *Driver) startMusic(cue audio.Cue) error {
	cmd, err := d.buildCommand(cue)
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	stopCh := make(chan struct{})
	d.mu.Lock()
	d.musicCmd = cmd
	d.musicStopCh = stopCh
	d.mu.Unlock()
	go d.loopMusic(cue, stopCh, cmd)
	return nil
}

func (d *Driver) loopMusic(cue audio.Cue, stopCh <-chan struct{}, cmd *exec.Cmd) {
	for {
		_ = cmd.Wait()
		select {
		case <-stopCh:
			return
		default:
		}
		if !cue.Loop {
			d.clearMusic(cmd)
			return
		}
		next, err := d.buildCommand(cue)
		if err != nil {
			d.clearMusic(cmd)
			return
		}
		if err := next.Start(); err != nil {
			d.clearMusic(cmd)
			return
		}
		d.mu.Lock()
		d.musicCmd = next
		d.mu.Unlock()
		cmd = next
	}
}

func (d *Driver) clearMusic(cmd *exec.Cmd) {
	d.mu.Lock()
	if d.musicCmd == cmd {
		d.musicCmd = nil
		d.musicStopCh = nil
	}
	d.mu.Unlock()
}

func (d *Driver) buildCommand(cue audio.Cue) (*exec.Cmd, error) {
	src, ok := d.sourceForCue(cue.ID)
	if !ok {
		return nil, fmt.Errorf("audio cue %q is not registered", cue.ID)
	}
	command := src.Command
	if command.Path == "" {
		command = d.command
	}
	if command.Path == "" {
		return nil, errors.New("audio command is required")
	}
	if strings.TrimSpace(src.Path) == "" {
		return nil, fmt.Errorf("audio cue %q has no path", cue.ID)
	}
	args := expandArgs(command.Args, cue, src.Path)
	return exec.Command(command.Path, args...), nil
}

func (d *Driver) sourceForCue(id string) (Source, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.sources == nil {
		return Source{}, false
	}
	src, ok := d.sources[id]
	return src, ok
}

func expandArgs(args []string, cue audio.Cue, path string) []string {
	if len(args) == 0 {
		return nil
	}
	out := make([]string, len(args))
	volume := fmt.Sprintf("%d", cue.Volume)
	for i, arg := range args {
		arg = strings.ReplaceAll(arg, "{{path}}", path)
		arg = strings.ReplaceAll(arg, "{{volume}}", volume)
		out[i] = arg
	}
	return out
}
