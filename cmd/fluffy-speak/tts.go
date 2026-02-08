package main

import (
	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/accessibility/tts"
)

// Speaker is an alias for accessibility.Speaker.
type Speaker = accessibility.Speaker

// detectSpeaker returns the best available TTS backend for the platform.
func detectSpeaker(name string, rate float64) (Speaker, error) {
	var opts []tts.Option
	if name != "" && name != "auto" {
		// Map legacy "spd" name to "ssip"
		if name == "spd" || name == "spd-say" {
			name = "ssip"
		}
		opts = append(opts, tts.WithBackend(name))
	}
	if rate != 0 && rate != 1.0 {
		opts = append(opts, tts.WithRate(rate))
	}
	s, err := tts.Auto(opts...)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return tts.NewMock(), nil
	}
	return s, nil
}
