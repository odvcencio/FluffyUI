// Package keybind provides key binding and command routing helpers.
package keybind

import (
	"fmt"
	"strings"

	"github.com/odvcencio/fluffyui/terminal"
)

// Condition gates a binding based on runtime context.
type Condition func(ctx Context) bool

// KeyPress represents a normalized key press.
type KeyPress struct {
	Key   terminal.Key
	Rune  rune
	Alt   bool
	Ctrl  bool
	Shift bool
}

// Equal reports whether two key presses match.
func (k KeyPress) Equal(other KeyPress) bool {
	if k.Key != other.Key || k.Alt != other.Alt || k.Ctrl != other.Ctrl || k.Shift != other.Shift {
		return false
	}
	if k.Key == terminal.KeyRune {
		return k.Rune == other.Rune
	}
	return true
}

// Key describes a key sequence (single key or chord).
type Key struct {
	Sequence []KeyPress
}

// Matches reports whether the sequence matches exactly.
func (k Key) Matches(seq []KeyPress) bool {
	if len(seq) != len(k.Sequence) {
		return false
	}
	for i := range k.Sequence {
		if !k.Sequence[i].Equal(seq[i]) {
			return false
		}
	}
	return true
}

// HasPrefix reports whether seq is a prefix of this key sequence.
func (k Key) HasPrefix(seq []KeyPress) bool {
	if len(seq) > len(k.Sequence) {
		return false
	}
	for i := range seq {
		if !k.Sequence[i].Equal(seq[i]) {
			return false
		}
	}
	return true
}

// Binding maps a key sequence to a command.
type Binding struct {
	Key     Key
	Command string
	When    Condition
}

// Keymap groups related key bindings with an optional parent.
type Keymap struct {
	Name     string
	Bindings []Binding
	Parent   *Keymap
}

type keyMatch struct {
	Binding *Binding
	Prefix  bool
}

// Match returns a binding match or prefix state for a sequence.
func (k *Keymap) Match(seq []KeyPress, ctx Context) keyMatch {
	if k == nil || len(seq) == 0 {
		return keyMatch{}
	}
	for i := range k.Bindings {
		binding := &k.Bindings[i]
		if !binding.Key.HasPrefix(seq) {
			continue
		}
		if len(seq) < len(binding.Key.Sequence) {
			return keyMatch{Prefix: true}
		}
		if binding.When != nil && !binding.When(ctx) {
			continue
		}
		return keyMatch{Binding: binding}
	}
	if k.Parent != nil {
		return k.Parent.Match(seq, ctx)
	}
	return keyMatch{}
}

// ParseKeySequence parses a key sequence like "ctrl+s" or "g g".
func ParseKeySequence(input string) (Key, error) {
	fields := strings.Fields(strings.TrimSpace(input))
	if len(fields) == 0 {
		return Key{}, fmt.Errorf("empty key sequence")
	}
	sequence := make([]KeyPress, 0, len(fields))
	for _, token := range fields {
		press, err := parseKeyToken(token)
		if err != nil {
			return Key{}, err
		}
		sequence = append(sequence, press)
	}
	return Key{Sequence: sequence}, nil
}

// MustParseKeySequence parses a key sequence and returns an empty Key on error.
func MustParseKeySequence(input string) Key {
	key, err := ParseKeySequence(input)
	if err != nil {
		return Key{}
	}
	return key
}

// parseKeyToken parses a single key token with modifiers.
func parseKeyToken(token string) (KeyPress, error) {
	var press KeyPress
	parts := strings.FieldsFunc(strings.ToLower(token), func(r rune) bool {
		return r == '+' || r == '-'
	})
	if len(parts) == 0 {
		return press, fmt.Errorf("invalid key token: %q", token)
	}
	keyName := parts[len(parts)-1]
	for _, mod := range parts[:len(parts)-1] {
		switch mod {
		case "ctrl", "control":
			press.Ctrl = true
		case "alt", "option":
			press.Alt = true
		case "shift":
			press.Shift = true
		case "meta", "cmd", "command":
			press.Alt = true
		default:
			return press, fmt.Errorf("unknown modifier %q", mod)
		}
	}

	if keyName == "" {
		return press, fmt.Errorf("missing key name in %q", token)
	}
	if r, ok := keyRuneMap[keyName]; ok {
		press.Key = terminal.KeyRune
		press.Rune = r
		return press, nil
	}
	if k, ok := keyNameMap[keyName]; ok {
		press.Key = k
		return press, nil
	}
	if len(keyName) == 1 {
		r := rune(keyName[0])
		if press.Ctrl {
			if ctrlKey, ok := ctrlKeyMap[r]; ok {
				press.Key = ctrlKey
				return press, nil
			}
		}
		press.Key = terminal.KeyRune
		press.Rune = r
		return press, nil
	}
	if strings.HasPrefix(keyName, "f") && len(keyName) <= 3 {
		switch keyName {
		case "f1":
			press.Key = terminal.KeyF1
		case "f2":
			press.Key = terminal.KeyF2
		case "f3":
			press.Key = terminal.KeyF3
		case "f4":
			press.Key = terminal.KeyF4
		case "f5":
			press.Key = terminal.KeyF5
		case "f6":
			press.Key = terminal.KeyF6
		case "f7":
			press.Key = terminal.KeyF7
		case "f8":
			press.Key = terminal.KeyF8
		case "f9":
			press.Key = terminal.KeyF9
		case "f10":
			press.Key = terminal.KeyF10
		case "f11":
			press.Key = terminal.KeyF11
		case "f12":
			press.Key = terminal.KeyF12
		default:
			return press, fmt.Errorf("unknown function key %q", keyName)
		}
		return press, nil
	}
	return press, fmt.Errorf("unknown key %q", keyName)
}

var keyRuneMap = map[string]rune{
	"space": ' ',
	"spc":   ' ',
}

var keyNameMap = map[string]terminal.Key{
	"enter":     terminal.KeyEnter,
	"return":    terminal.KeyEnter,
	"tab":       terminal.KeyTab,
	"escape":    terminal.KeyEscape,
	"esc":       terminal.KeyEscape,
	"backspace": terminal.KeyBackspace,
	"bs":        terminal.KeyBackspace,
	"delete":    terminal.KeyDelete,
	"del":       terminal.KeyDelete,
	"insert":    terminal.KeyInsert,
	"home":      terminal.KeyHome,
	"end":       terminal.KeyEnd,
	"pageup":    terminal.KeyPageUp,
	"pagedown":  terminal.KeyPageDown,
	"pgup":      terminal.KeyPageUp,
	"pgdn":      terminal.KeyPageDown,
	"up":        terminal.KeyUp,
	"down":      terminal.KeyDown,
	"left":      terminal.KeyLeft,
	"right":     terminal.KeyRight,
}

var ctrlKeyMap = map[rune]terminal.Key{
	'b': terminal.KeyCtrlB,
	'c': terminal.KeyCtrlC,
	'd': terminal.KeyCtrlD,
	'f': terminal.KeyCtrlF,
	'p': terminal.KeyCtrlP,
	'v': terminal.KeyCtrlV,
	'x': terminal.KeyCtrlX,
	'z': terminal.KeyCtrlZ,
}
