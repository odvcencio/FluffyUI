//go:build !linux

package tts

import (
	"fmt"

	"m31labs.dev/fluffyui/accessibility"
)

func newSSIPSpeaker(_ float64) (accessibility.Speaker, error) {
	return nil, fmt.Errorf("SSIP speech-dispatcher is only available on Linux")
}
