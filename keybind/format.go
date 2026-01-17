package keybind

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/odvcencio/fluffy-ui/terminal"
)

// String returns a formatted key press.
func (k KeyPress) String() string {
	return FormatKeyPress(k)
}

// String returns a formatted key sequence.
func (k Key) String() string {
	return FormatKeySequence(k)
}

// FormatKeySequence formats a key sequence for display.
func FormatKeySequence(key Key) string {
	if len(key.Sequence) == 0 {
		return ""
	}
	parts := make([]string, 0, len(key.Sequence))
	for _, press := range key.Sequence {
		parts = append(parts, FormatKeyPress(press))
	}
	return strings.Join(parts, " ")
}

// FormatKeySequences formats multiple key sequences for display.
func FormatKeySequences(keys []Key) string {
	if len(keys) == 0 {
		return ""
	}
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		if seq := FormatKeySequence(key); seq != "" {
			parts = append(parts, seq)
		}
	}
	return strings.Join(parts, ", ")
}

// FormatKeyPress formats a single key press for display.
func FormatKeyPress(press KeyPress) string {
	ctrl := press.Ctrl || isCtrlKey(press.Key)
	mods := make([]string, 0, 3)
	if ctrl {
		mods = append(mods, "Ctrl")
	}
	if press.Alt {
		mods = append(mods, "Alt")
	}
	if press.Shift {
		mods = append(mods, "Shift")
	}
	base := keyDisplayName(press)
	if base == "" {
		base = "?"
	}
	mods = append(mods, base)
	return strings.Join(mods, "+")
}

func keyDisplayName(press KeyPress) string {
	if press.Key == terminal.KeyRune {
		return formatRune(press.Rune)
	}
	if name, ok := keyDisplayNameMap[press.Key]; ok {
		return name
	}
	if r, ok := ctrlKeyName[press.Key]; ok {
		return formatRune(r)
	}
	return fmt.Sprintf("Key%d", press.Key)
}

func formatRune(r rune) string {
	if r == 0 {
		return ""
	}
	switch r {
	case ' ':
		return "Space"
	case '\t':
		return "Tab"
	case '\n':
		return "Enter"
	}
	if unicode.IsLetter(r) {
		return strings.ToUpper(string(r))
	}
	return string(r)
}

var keyDisplayNameMap = map[terminal.Key]string{
	terminal.KeyEnter:     "Enter",
	terminal.KeyBackspace: "Backspace",
	terminal.KeyTab:       "Tab",
	terminal.KeyEscape:    "Esc",
	terminal.KeyUp:        "Up",
	terminal.KeyDown:      "Down",
	terminal.KeyLeft:      "Left",
	terminal.KeyRight:     "Right",
	terminal.KeyHome:      "Home",
	terminal.KeyEnd:       "End",
	terminal.KeyPageUp:    "PgUp",
	terminal.KeyPageDown:  "PgDn",
	terminal.KeyDelete:    "Del",
	terminal.KeyInsert:    "Ins",
	terminal.KeyF1:        "F1",
	terminal.KeyF2:        "F2",
	terminal.KeyF3:        "F3",
	terminal.KeyF4:        "F4",
	terminal.KeyF5:        "F5",
	terminal.KeyF6:        "F6",
	terminal.KeyF7:        "F7",
	terminal.KeyF8:        "F8",
	terminal.KeyF9:        "F9",
	terminal.KeyF10:       "F10",
	terminal.KeyF11:       "F11",
	terminal.KeyF12:       "F12",
}

var ctrlKeyName = map[terminal.Key]rune{
	terminal.KeyCtrlB: 'b',
	terminal.KeyCtrlC: 'c',
	terminal.KeyCtrlD: 'd',
	terminal.KeyCtrlF: 'f',
	terminal.KeyCtrlP: 'p',
	terminal.KeyCtrlV: 'v',
	terminal.KeyCtrlX: 'x',
	terminal.KeyCtrlZ: 'z',
}
