// Package tts provides platform TTS backends implementing accessibility.Speaker.
package tts

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"m31labs.dev/fluffyui/accessibility"
)

// Option configures TTS auto-detection.
type Option func(*options)

type options struct {
	backend string
	rate    float64
}

// WithBackend forces a specific backend by name.
func WithBackend(name string) Option {
	return func(o *options) { o.backend = name }
}

// WithRate sets the speech rate multiplier (1.0 = normal).
func WithRate(rate float64) Option {
	return func(o *options) { o.rate = rate }
}

// Auto returns the best available TTS speaker for the current platform.
// Returns (nil, nil) if no TTS engine is available — the app runs silently.
func Auto(opts ...Option) (accessibility.Speaker, error) {
	o := &options{rate: 1.0}
	for _, opt := range opts {
		opt(o)
	}

	name := strings.ToLower(strings.TrimSpace(o.backend))
	switch name {
	case "mock":
		return NewMock(), nil
	case "espeak":
		return newEspeakSpeaker(o.rate)
	case "ssip":
		return newSSIPSpeaker(o.rate)
	case "say":
		return newSaySpeaker(o.rate)
	case "sapi":
		return newSAPISpeaker(o.rate)
	case "", "auto":
		return autoDetect(o.rate)
	default:
		return nil, fmt.Errorf("unknown TTS backend: %s", name)
	}
}

func autoDetect(rate float64) (accessibility.Speaker, error) {
	switch runtime.GOOS {
	case "darwin":
		if hasBinary("say") {
			return newSaySpeaker(rate)
		}
	case "linux":
		// SSIP (speech-dispatcher socket) — native Linux
		if s, err := newSSIPSpeaker(rate); err == nil {
			return s, nil
		}
		// SAPI via powershell.exe — WSL2 with Windows audio
		if s, err := newSAPISpeaker(rate); err == nil {
			return s, nil
		}
	case "windows":
		return newSAPISpeaker(rate)
	}
	if hasBinary("espeak-ng") || hasBinary("espeak") {
		return newEspeakSpeaker(rate)
	}
	return nil, nil
}

func hasBinary(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
