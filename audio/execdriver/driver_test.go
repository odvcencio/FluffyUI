package execdriver

import (
	"reflect"
	"testing"

	"github.com/odvcencio/fluffy-ui/audio"
)

func TestExpandArgs(t *testing.T) {
	cue := audio.Cue{ID: "click", Volume: 42}
	args := []string{"--file", "{{path}}", "--volume={{volume}}"}
	got := expandArgs(args, cue, "/tmp/sound.wav")
	want := []string{"--file", "/tmp/sound.wav", "--volume=42"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected args: %#v", got)
	}
}

func TestBuildCommandUsesDefaults(t *testing.T) {
	driver := NewDriver(Config{
		Command: Command{Path: "player", Args: []string{"{{path}}"}},
		Sources: map[string]Source{
			"ui.click": {Path: "click.wav"},
		},
	})
	cmd, err := driver.buildCommand(audio.Cue{ID: "ui.click", Volume: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.Path != "player" {
		t.Fatalf("expected command path player, got %q", cmd.Path)
	}
	if len(cmd.Args) < 2 || cmd.Args[1] != "click.wav" {
		t.Fatalf("expected args to include path, got %#v", cmd.Args)
	}
}

func TestCommandsForOS(t *testing.T) {
	cases := map[string]string{
		"darwin":  "afplay",
		"linux":   "paplay",
		"windows": "powershell",
	}
	for goos, want := range cases {
		cmds := commandsForOS(goos)
		if len(cmds) == 0 {
			t.Fatalf("expected commands for %s", goos)
		}
		if cmds[0].Path != want {
			t.Fatalf("expected %s default %s, got %s", goos, want, cmds[0].Path)
		}
	}
}
