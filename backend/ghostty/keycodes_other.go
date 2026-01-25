//go:build !linux

package ghostty

import "github.com/odvcencio/fluffy-ui/terminal"

func ghosttyKeycodeFromTerminalKey(_ terminal.Key) (uint32, bool) {
	return 0, false
}

func ghosttyKeycodeFromRune(_ rune) (uint32, bool) {
	return 0, false
}
