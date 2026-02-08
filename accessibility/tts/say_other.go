//go:build !darwin

package tts

import (
	"fmt"

	"github.com/odvcencio/fluffyui/accessibility"
)

func newSaySpeaker(_ float64) (accessibility.Speaker, error) {
	return nil, fmt.Errorf("macOS say is only available on macOS")
}
