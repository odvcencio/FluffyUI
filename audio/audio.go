// Package audio provides opinionated music and sound effect hooks for apps.
package audio

import "time"

// Kind describes the playback channel for a cue.
type Kind int

const (
	KindSFX Kind = iota
	KindMusic
)

const DefaultVolume = 100

// Cue defines a registered piece of audio.
//
// Volume is 0-100. A zero value uses DefaultVolume.
// Cooldown prevents rapid replays of the same cue.
type Cue struct {
	ID       string
	Kind     Kind
	Volume   int
	Loop     bool
	Cooldown time.Duration
}

// Driver executes playback requests.
type Driver interface {
	Play(cue Cue) error
	Stop(kind Kind) error
}

// Service exposes the opinionated audio API used by widgets.
type Service interface {
	Play(id string) bool
	PlaySFX(id string) bool
	PlayMusic(id string) bool
	StopMusic() bool
	SetMuted(muted bool)
	Muted() bool
	SetMasterVolume(percent int)
	SetSFXVolume(percent int)
	SetMusicVolume(percent int)
}

// Manager routes cue playback through a driver.
// Use NewManager to initialize defaults.
type Manager struct {
	driver       Driver
	cues         map[string]Cue
	lastPlayed   map[string]time.Time
	masterVolume int
	sfxVolume    int
	musicVolume  int
	muted        bool
	currentMusic string
	clockNow     func() time.Time
}

// NewManager creates a manager with optional pre-registered cues.
func NewManager(driver Driver, cues ...Cue) *Manager {
	manager := &Manager{
		driver:       driver,
		cues:         make(map[string]Cue),
		lastPlayed:   make(map[string]time.Time),
		masterVolume: DefaultVolume,
		sfxVolume:    DefaultVolume,
		musicVolume:  DefaultVolume,
		clockNow:     time.Now,
	}
	manager.RegisterAll(cues...)
	return manager
}

// Register adds or replaces a cue definition.
func (m *Manager) Register(cue Cue) {
	if m == nil || cue.ID == "" {
		return
	}
	if m.cues == nil {
		m.cues = make(map[string]Cue)
	}
	m.cues[cue.ID] = normalizeCue(cue)
}

// RegisterAll registers multiple cues.
func (m *Manager) RegisterAll(cues ...Cue) {
	if m == nil {
		return
	}
	for _, cue := range cues {
		m.Register(cue)
	}
}

// Play plays a cue by ID, regardless of kind.
func (m *Manager) Play(id string) bool {
	if m == nil || m.driver == nil || m.muted {
		return false
	}
	if m.cues == nil {
		return false
	}
	cue, ok := m.cues[id]
	if !ok {
		return false
	}
	now := time.Now()
	if m.clockNow != nil {
		now = m.clockNow()
	}
	if cue.Cooldown > 0 {
		if last, ok := m.lastPlayed[id]; ok && now.Sub(last) < cue.Cooldown {
			return false
		}
	}
	if cue.Kind == KindMusic {
		if m.currentMusic == cue.ID {
			return false
		}
		if m.currentMusic != "" {
			_ = m.driver.Stop(KindMusic)
		}
	}
	cue = m.applyVolumes(cue)
	if cue.Volume <= 0 {
		return false
	}
	if err := m.driver.Play(cue); err != nil {
		return false
	}
	if m.lastPlayed == nil {
		m.lastPlayed = make(map[string]time.Time)
	}
	m.lastPlayed[id] = now
	if cue.Kind == KindMusic {
		m.currentMusic = cue.ID
	}
	return true
}

// PlaySFX plays a sound effect cue.
func (m *Manager) PlaySFX(id string) bool {
	if m == nil {
		return false
	}
	cue, ok := m.cues[id]
	if !ok || cue.Kind != KindSFX {
		return false
	}
	return m.Play(id)
}

// PlayMusic plays a music cue.
func (m *Manager) PlayMusic(id string) bool {
	if m == nil {
		return false
	}
	cue, ok := m.cues[id]
	if !ok || cue.Kind != KindMusic {
		return false
	}
	return m.Play(id)
}

// StopMusic stops the current music track, if any.
func (m *Manager) StopMusic() bool {
	if m == nil {
		return false
	}
	hadMusic := m.currentMusic != ""
	m.currentMusic = ""
	if m.driver == nil {
		return hadMusic
	}
	_ = m.driver.Stop(KindMusic)
	return hadMusic
}

// SetMuted toggles whether new cues are played.
func (m *Manager) SetMuted(muted bool) {
	if m == nil {
		return
	}
	m.muted = muted
	if muted {
		m.StopMusic()
	}
}

// Muted reports the current mute state.
func (m *Manager) Muted() bool {
	if m == nil {
		return true
	}
	return m.muted
}

// SetMasterVolume configures the global volume percentage.
func (m *Manager) SetMasterVolume(percent int) {
	if m == nil {
		return
	}
	m.masterVolume = clampPercent(percent)
}

// SetSFXVolume configures the sound effects volume percentage.
func (m *Manager) SetSFXVolume(percent int) {
	if m == nil {
		return
	}
	m.sfxVolume = clampPercent(percent)
}

// SetMusicVolume configures the music volume percentage.
func (m *Manager) SetMusicVolume(percent int) {
	if m == nil {
		return
	}
	m.musicVolume = clampPercent(percent)
}

// Disabled is a no-op audio service.
type Disabled struct{}

func (Disabled) Play(id string) bool         { return false }
func (Disabled) PlaySFX(id string) bool      { return false }
func (Disabled) PlayMusic(id string) bool    { return false }
func (Disabled) StopMusic() bool             { return false }
func (Disabled) SetMuted(muted bool)         {}
func (Disabled) Muted() bool                 { return true }
func (Disabled) SetMasterVolume(percent int) {}
func (Disabled) SetSFXVolume(percent int)    {}
func (Disabled) SetMusicVolume(percent int)  {}

// NoopDriver is a driver that accepts requests without playing audio.
type NoopDriver struct{}

func (NoopDriver) Play(cue Cue) error   { return nil }
func (NoopDriver) Stop(kind Kind) error { return nil }

func (m *Manager) applyVolumes(cue Cue) Cue {
	volume := cue.Volume
	if volume == 0 {
		volume = DefaultVolume
	}
	volume = clampPercent(volume)
	volume = applyPercent(volume, m.masterVolume)
	if cue.Kind == KindMusic {
		volume = applyPercent(volume, m.musicVolume)
	} else {
		volume = applyPercent(volume, m.sfxVolume)
	}
	cue.Volume = volume
	return cue
}

func normalizeCue(cue Cue) Cue {
	if cue.Volume == 0 {
		cue.Volume = DefaultVolume
	}
	cue.Volume = clampPercent(cue.Volume)
	if cue.Cooldown < 0 {
		cue.Cooldown = 0
	}
	return cue
}

func clampPercent(value int) int {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

func applyPercent(value, percent int) int {
	if value <= 0 || percent <= 0 {
		return 0
	}
	return (value*percent + 50) / 100
}
