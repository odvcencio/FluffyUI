package terminal

import (
	"fmt"
	"strconv"
	"strings"
)

// Kitty keyboard protocol flag constants. These are combined as a bitmask
// when enabling the protocol via KittyKeyboardEnable.
const (
	KittyDisambiguate    = 1 // Report disambiguated key codes
	KittyReportEvents    = 2 // Report key press/release/repeat events
	KittyReportAlternates = 4 // Report alternate key representations
	KittyReportAllKeys   = 8 // Report all keys as escape codes
)

// Kitty functional key code offsets. The Kitty keyboard protocol assigns
// function keys to a dedicated Unicode private use area starting at 57344.
const (
	kittyFuncKeyBase = 57344

	kittyKeyEscape    = 27
	kittyKeyEnter     = 13
	kittyKeyTab       = 9
	kittyKeyBackspace = 127
	kittyKeyInsert    = 57348
	kittyKeyDelete    = 57349
	kittyKeyLeft      = 57350
	kittyKeyRight     = 57351
	kittyKeyUp        = 57352
	kittyKeyDown      = 57353
	kittyKeyPageUp    = 57354
	kittyKeyPageDown  = 57355
	kittyKeyHome      = 57356
	kittyKeyEnd       = 57357
	kittyKeyF1        = 57344
	kittyKeyF2        = 57345
	kittyKeyF3        = 57346
	kittyKeyF4        = 57347
	kittyKeyF5        = 57348 // Note: overlaps Insert in some docs; see below
	kittyKeyF6        = 57349
	kittyKeyF7        = 57350
	kittyKeyF8        = 57351
	kittyKeyF9        = 57352
	kittyKeyF10       = 57353
	kittyKeyF11       = 57354
	kittyKeyF12       = 57355
)

// The Kitty protocol actually uses legacy CSI sequences for arrow keys,
// home, end, insert, delete, page up/down, and F-keys in many modes.
// The functional key range 57344+ is used only with higher flag levels.
// We handle both: legacy keycodes (13, 27, 9, 127) and the functional
// key range for F1-F12 (57344-57355).

// kittyKeycodeMap maps Kitty keyboard protocol keycodes to terminal.Key.
// This covers both legacy keycodes (Enter, Escape, Tab, Backspace) and
// the Kitty functional key range (57344+).
var kittyKeycodeMap = map[int]Key{
	// Legacy keycodes
	kittyKeyEnter:     KeyEnter,
	kittyKeyEscape:    KeyEscape,
	kittyKeyTab:       KeyTab,
	kittyKeyBackspace: KeyBackspace,

	// Kitty functional key range (57344+)
	57344: KeyF1,
	57345: KeyF2,
	57346: KeyF3,
	57347: KeyF4,
	57348: KeyInsert,
	57349: KeyDelete,
	57350: KeyLeft,
	57351: KeyRight,
	57352: KeyUp,
	57353: KeyDown,
	57354: KeyPageUp,
	57355: KeyPageDown,
	57356: KeyHome,
	57357: KeyEnd,
	57358: KeyF1,
	57359: KeyF2,
	57360: KeyF3,
	57361: KeyF4,
	57362: KeyF5,
	57363: KeyF6,
	57364: KeyF7,
	57365: KeyF8,
	57366: KeyF9,
	57367: KeyF10,
	57368: KeyF11,
	57369: KeyF12,
}

// ParseKittyKey parses a CSI u (Kitty keyboard protocol) escape sequence.
//
// Format: CSI unicode-key-code:shifted-key ; modifiers:event-type u
//   - CSI is \x1b[
//   - unicode-key-code is the main key's Unicode codepoint (or functional key code)
//   - shifted-key (optional, colon-separated) is the codepoint when shift is held
//   - modifiers (optional) encode shift/alt/ctrl as a bitmask+1
//   - event-type (optional, colon-separated) is 1=press, 2=repeat, 3=release
//
// Returns a KeyEvent and true if parsed, or zero value and false if not
// a valid Kitty keyboard sequence.
func ParseKittyKey(seq []byte) (KeyEvent, bool) {
	// Minimum valid: \x1b [ <digit> u  (4 bytes minimum)
	if len(seq) < 4 {
		return KeyEvent{}, false
	}

	// Must start with CSI (\x1b[) and end with 'u'
	if seq[0] != 0x1b || seq[1] != '[' || seq[len(seq)-1] != 'u' {
		return KeyEvent{}, false
	}

	// Extract the payload between '[' and 'u'
	payload := string(seq[2 : len(seq)-1])

	// Split on ';' to separate keycode part from modifiers part
	parts := strings.SplitN(payload, ";", 2)

	// Parse keycode part (may contain colon-separated shifted key)
	keycodePart := parts[0]
	keycodeParts := strings.SplitN(keycodePart, ":", 2)

	keycode, err := strconv.Atoi(keycodeParts[0])
	if err != nil || keycode < 0 {
		return KeyEvent{}, false
	}

	// Parse shifted keycode if present
	shiftedKeycode := 0
	if len(keycodeParts) > 1 && keycodeParts[1] != "" {
		shiftedKeycode, err = strconv.Atoi(keycodeParts[1])
		if err != nil {
			return KeyEvent{}, false
		}
	}

	// Parse modifiers and event type
	var modifiers int
	if len(parts) > 1 {
		modPart := parts[1]
		modParts := strings.SplitN(modPart, ":", 2)

		modifiers, err = strconv.Atoi(modParts[0])
		if err != nil {
			return KeyEvent{}, false
		}
		// event type in modParts[1] is parsed but not used for KeyEvent
		// (we treat everything as a press for now)
	}

	// Decode modifier bitmask. The protocol sends modifiers+1,
	// so subtract 1 to get the actual bitmask.
	modBits := 0
	if modifiers > 0 {
		modBits = modifiers - 1
	}
	shift := modBits&1 != 0
	alt := modBits&2 != 0
	ctrl := modBits&4 != 0

	ev := KeyEvent{
		Shift: shift,
		Alt:   alt,
		Ctrl:  ctrl,
	}

	// Check for known special keycodes
	if key, ok := kittyKeycodeMap[keycode]; ok {
		ev.Key = key
		return ev, true
	}

	// If there is a shifted keycode and shift is held, use the shifted codepoint
	if shiftedKeycode > 0 && shift {
		ev.Key = KeyRune
		ev.Rune = rune(shiftedKeycode)
		return ev, true
	}

	// Printable character: keycode is the Unicode codepoint
	if keycode >= 32 && keycode < kittyFuncKeyBase {
		ev.Key = KeyRune
		ev.Rune = rune(keycode)
		return ev, true
	}

	// Ctrl+letter keycodes (1-26 map to Ctrl+A through Ctrl+Z)
	if keycode >= 1 && keycode <= 26 {
		ev.Ctrl = true
		ev.Key = KeyRune
		ev.Rune = rune('a' + keycode - 1)
		return ev, true
	}

	// Unknown keycode in functional range but not in our map
	if keycode >= kittyFuncKeyBase {
		ev.Key = KeyNone
		return ev, true
	}

	return KeyEvent{}, false
}

// KittyKeyboardEnable returns the escape sequence to enable the Kitty
// keyboard protocol with the given flags.
//
// Flags are a bitmask of:
//   - 1: Disambiguate escape codes
//   - 2: Report event types (press/repeat/release)
//   - 4: Report alternate keys
//   - 8: Report all keys as escape codes
//
// The resulting sequence pushes a new keyboard mode onto the terminal's
// stack, allowing nested enable/disable.
func KittyKeyboardEnable(flags int) string {
	return fmt.Sprintf("\x1b[>%du", flags)
}

// KittyKeyboardDisable returns the escape sequence to disable the Kitty
// keyboard protocol. This pops the most recent keyboard mode from the
// terminal's stack.
func KittyKeyboardDisable() string {
	return "\x1b[<u"
}
