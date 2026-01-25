//go:build linux

package ghostty

import "github.com/odvcencio/fluffy-ui/terminal"

var xkbKeycodeByRune = map[rune]uint32{
	'a':  38,
	'b':  56,
	'c':  54,
	'd':  40,
	'e':  26,
	'f':  41,
	'g':  42,
	'h':  43,
	'i':  31,
	'j':  44,
	'k':  45,
	'l':  46,
	'm':  58,
	'n':  57,
	'o':  32,
	'p':  33,
	'q':  24,
	'r':  27,
	's':  39,
	't':  28,
	'u':  30,
	'v':  55,
	'w':  25,
	'x':  53,
	'y':  29,
	'z':  52,
	'1':  10,
	'2':  11,
	'3':  12,
	'4':  13,
	'5':  14,
	'6':  15,
	'7':  16,
	'8':  17,
	'9':  18,
	'0':  19,
	' ':  65,
	'-':  20,
	'_':  20,
	'=':  21,
	'+':  21,
	'[':  34,
	'{':  34,
	']':  35,
	'}':  35,
	'\\': 51,
	'|':  51,
	';':  47,
	':':  47,
	'\'': 48,
	'"':  48,
	'`':  49,
	'~':  49,
	',':  59,
	'<':  59,
	'.':  60,
	'>':  60,
	'/':  61,
	'?':  61,
	'\t': 23,
	'\n': 36,
	'\r': 36,
	'\b': 22,
}

func ghosttyKeycodeFromTerminalKey(key terminal.Key) (uint32, bool) {
	switch key {
	case terminal.KeyEnter:
		return 36, true
	case terminal.KeyBackspace:
		return 22, true
	case terminal.KeyTab:
		return 23, true
	case terminal.KeyEscape:
		return 9, true
	case terminal.KeyDelete:
		return 119, true
	case terminal.KeyInsert:
		return 118, true
	case terminal.KeyHome:
		return 110, true
	case terminal.KeyEnd:
		return 115, true
	case terminal.KeyPageUp:
		return 112, true
	case terminal.KeyPageDown:
		return 117, true
	case terminal.KeyUp:
		return 111, true
	case terminal.KeyDown:
		return 116, true
	case terminal.KeyLeft:
		return 113, true
	case terminal.KeyRight:
		return 114, true
	case terminal.KeyF1:
		return 67, true
	case terminal.KeyF2:
		return 68, true
	case terminal.KeyF3:
		return 69, true
	case terminal.KeyF4:
		return 70, true
	case terminal.KeyF5:
		return 71, true
	case terminal.KeyF6:
		return 72, true
	case terminal.KeyF7:
		return 73, true
	case terminal.KeyF8:
		return 74, true
	case terminal.KeyF9:
		return 75, true
	case terminal.KeyF10:
		return 76, true
	case terminal.KeyF11:
		return 95, true
	case terminal.KeyF12:
		return 96, true
	case terminal.KeyCtrlB:
		return ghosttyKeycodeFromRune('b')
	case terminal.KeyCtrlC:
		return ghosttyKeycodeFromRune('c')
	case terminal.KeyCtrlD:
		return ghosttyKeycodeFromRune('d')
	case terminal.KeyCtrlF:
		return ghosttyKeycodeFromRune('f')
	case terminal.KeyCtrlP:
		return ghosttyKeycodeFromRune('p')
	case terminal.KeyCtrlV:
		return ghosttyKeycodeFromRune('v')
	case terminal.KeyCtrlX:
		return ghosttyKeycodeFromRune('x')
	case terminal.KeyCtrlZ:
		return ghosttyKeycodeFromRune('z')
	default:
		return 0, false
	}
}

func ghosttyKeycodeFromRune(r rune) (uint32, bool) {
	if r >= 'A' && r <= 'Z' {
		r = r + ('a' - 'A')
	}
	code, ok := xkbKeycodeByRune[r]
	return code, ok
}
