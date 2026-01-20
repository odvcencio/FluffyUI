# Audio (Music + SFX)

FluffyUI includes an opinionated audio service that keeps music and sound
effects consistent across apps. The runtime does not play audio directly; you
wire a driver that integrates your preferred Go audio library.

## Concepts

- **Cue**: A named piece of audio with a kind (music or SFX), volume, loop flag,
  and optional cooldown.
- **Manager**: Registers cues, applies volume rules, enforces cooldowns, and
  guarantees a single active music track.
- **Driver**: Your playback backend. The interface is small and keeps FluffyUI
  decoupled from audio dependencies.

## Setup

```go
import (
    "time"

    "github.com/odvcencio/fluffy-ui/audio"
    "github.com/odvcencio/fluffy-ui/runtime"
)

driver := NewMyAudioDriver() // implement audio.Driver
manager := audio.NewManager(driver,
    audio.Cue{ID: "ui.click", Kind: audio.KindSFX, Volume: 80, Cooldown: 60 * time.Millisecond},
    audio.Cue{ID: "music.menu", Kind: audio.KindMusic, Volume: 50, Loop: true},
)

app := runtime.NewApp(runtime.AppConfig{
    // ...
    Audio: manager,
})
```

## Playing Cues in Widgets

```go
type MyWidget struct {
    widgets.Component
    audio audio.Service
}

func (w *MyWidget) Bind(services runtime.Services) {
    w.Component.Bind(services)
    w.audio = services.Audio()
}

func (w *MyWidget) HandleMessage(msg runtime.Message) runtime.HandleResult {
    if key, ok := msg.(runtime.KeyMsg); ok && key.Rune == ' ' {
        if w.audio != nil {
            w.audio.PlaySFX("ui.click")
        }
    }
    return runtime.Unhandled()
}
```

## Volume and Mute

`audio.Manager` applies master + per-channel volumes when playing a cue. If you
update volumes while music is already playing, call `PlayMusic` again to apply
the new level.

```go
manager.SetMasterVolume(90)
manager.SetSFXVolume(70)
manager.SetMusicVolume(40)
manager.SetMuted(true)  // stops music and blocks new plays
```

## No-Audio Mode

Use `audio.Disabled{}` or `audio.NoopDriver{}` when you want the API without
real playback (tests, CI, or headless builds).

## Command Driver (execdriver)

`audio/execdriver` runs external commands to play cues. Configure a default
player and map cue IDs to file paths.

```go
import "github.com/odvcencio/fluffy-ui/audio/execdriver"

command, _ := execdriver.DetectCommand()
driver := execdriver.NewDriver(execdriver.Config{
    Command: command,
    Sources: map[string]execdriver.Source{
        "ui.click":   {Path: "assets/click.wav"},
        "music.menu": {Path: "assets/menu.wav"},
    },
})
```

Args support `{{path}}` and `{{volume}}` placeholders. Music cues loop by
restarting the command when `Cue.Loop` is true.

`execdriver.DefaultCommandCandidates()` provides OS-specific fallbacks if you
want to customize or skip auto-detection.
