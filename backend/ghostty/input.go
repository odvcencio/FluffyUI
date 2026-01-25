//go:build linux || darwin || windows

package ghostty

import (
	"unicode"

	"github.com/odvcencio/fluffy-ui/terminal"
)

const (
	ghosttyActionRelease = 0
	ghosttyActionPress   = 1
	ghosttyActionRepeat  = 2
)

const (
	ghosttyModShift = 1 << 0
	ghosttyModCtrl  = 1 << 1
	ghosttyModAlt   = 1 << 2
	ghosttyModSuper = 1 << 3
)

const (
	ghosttyMouseUnknown = 0
	ghosttyMouseLeft    = 1
	ghosttyMouseRight   = 2
	ghosttyMouseMiddle  = 3
)

const (
	ghosttyMouseRelease = 0
	ghosttyMousePress   = 1
)

// Keep in sync with ghostty_input_key_e in include/ghostty.h.
const (
	ghosttyKeyB               = 21
	ghosttyKeyC               = 22
	ghosttyKeyD               = 23
	ghosttyKeyF               = 25
	ghosttyKeyP               = 35
	ghosttyKeyV               = 41
	ghosttyKeyX               = 43
	ghosttyKeyZ               = 45
	ghosttyKeyBackspace       = 53
	ghosttyKeyEnter           = 58
	ghosttyKeyTab             = 64
	ghosttyKeyDelete          = 68
	ghosttyKeyEnd             = 69
	ghosttyKeyHome            = 71
	ghosttyKeyInsert          = 72
	ghosttyKeyPageDown        = 73
	ghosttyKeyPageUp          = 74
	ghosttyKeyArrowDown       = 75
	ghosttyKeyArrowLeft       = 76
	ghosttyKeyArrowRight      = 77
	ghosttyKeyArrowUp         = 78
	ghosttyKeyNumpadBackspace = 91
	ghosttyKeyNumpadEnter     = 97
	ghosttyKeyNumpadUp        = 109
	ghosttyKeyNumpadDown      = 110
	ghosttyKeyNumpadRight     = 111
	ghosttyKeyNumpadLeft      = 112
	ghosttyKeyNumpadHome      = 114
	ghosttyKeyNumpadEnd       = 115
	ghosttyKeyNumpadInsert    = 116
	ghosttyKeyNumpadDelete    = 117
	ghosttyKeyNumpadPageUp    = 118
	ghosttyKeyNumpadPageDown  = 119
	ghosttyKeyEscape          = 120
	ghosttyKeyF1              = 121
	ghosttyKeyF2              = 122
	ghosttyKeyF3              = 123
	ghosttyKeyF4              = 124
	ghosttyKeyF5              = 125
	ghosttyKeyF6              = 126
	ghosttyKeyF7              = 127
	ghosttyKeyF8              = 128
	ghosttyKeyF9              = 129
	ghosttyKeyF10             = 130
	ghosttyKeyF11             = 131
	ghosttyKeyF12             = 132
)

func ghosttyKeyEventToTerminal(ev ghosttyEventKeyData) (terminal.KeyEvent, bool) {
	if ev.Action == ghosttyActionRelease {
		return terminal.KeyEvent{}, false
	}
	alt, ctrl, shift := ghosttyModsToTerminal(ev.Mods)
	if ctrl {
		if key := ghosttyKeyToTerminal(ev.Key, true); isCtrlTerminalKey(key) {
			return terminal.KeyEvent{Key: key, Rune: 0, Alt: alt, Ctrl: true, Shift: shift}, true
		}
	}
	if ev.Rune != 0 {
		r := rune(ev.Rune)
		key := terminal.KeyRune
		if ctrl {
			if ctrlKey, ok := ctrlKeyFromRune(r); ok {
				key = ctrlKey
				r = 0
			}
		}
		return terminal.KeyEvent{Key: key, Rune: r, Alt: alt, Ctrl: ctrl, Shift: shift}, true
	}
	key := ghosttyKeyToTerminal(ev.Key, ctrl)
	if key == terminal.KeyNone {
		return terminal.KeyEvent{}, false
	}
	if isCtrlTerminalKey(key) {
		ctrl = true
	}
	return terminal.KeyEvent{Key: key, Rune: 0, Alt: alt, Ctrl: ctrl, Shift: shift}, true
}

func ghosttyMouseEventToTerminal(tag ghosttyEventType, ev ghosttyEventMouseData) (terminal.MouseEvent, bool) {
	alt, ctrl, shift := ghosttyModsToTerminal(ev.Mods)
	x := int(ev.X)
	y := int(ev.Y)
	switch tag {
	case ghosttyEventMouseMove:
		return terminal.MouseEvent{
			X:      x,
			Y:      y,
			Button: terminal.MouseNone,
			Action: terminal.MouseMove,
			Alt:    alt,
			Ctrl:   ctrl,
			Shift:  shift,
		}, true
	case ghosttyEventMouseButton:
		button := ghosttyMouseButtonToTerminal(ev.Button)
		if button == terminal.MouseNone && ev.Button != ghosttyMouseUnknown {
			return terminal.MouseEvent{}, false
		}
		action := terminal.MousePress
		if ev.State == ghosttyMouseRelease {
			action = terminal.MouseRelease
		}
		return terminal.MouseEvent{
			X:      x,
			Y:      y,
			Button: button,
			Action: action,
			Alt:    alt,
			Ctrl:   ctrl,
			Shift:  shift,
		}, true
	case ghosttyEventMouseScroll:
		if ev.ScrollY == 0 {
			return terminal.MouseEvent{}, false
		}
		button := terminal.MouseWheelUp
		if ev.ScrollY < 0 {
			button = terminal.MouseWheelDown
		}
		return terminal.MouseEvent{
			X:      x,
			Y:      y,
			Button: button,
			Action: terminal.MousePress,
			Alt:    alt,
			Ctrl:   ctrl,
			Shift:  shift,
		}, true
	default:
		return terminal.MouseEvent{}, false
	}
}

func ghosttyModsToTerminal(mods int32) (alt, ctrl, shift bool) {
	return mods&ghosttyModAlt != 0, mods&ghosttyModCtrl != 0, mods&ghosttyModShift != 0
}

func ghosttyKeyToTerminal(key int32, ctrl bool) terminal.Key {
	switch key {
	case ghosttyKeyEnter, ghosttyKeyNumpadEnter:
		return terminal.KeyEnter
	case ghosttyKeyBackspace, ghosttyKeyNumpadBackspace:
		return terminal.KeyBackspace
	case ghosttyKeyTab:
		return terminal.KeyTab
	case ghosttyKeyEscape:
		return terminal.KeyEscape
	case ghosttyKeyDelete, ghosttyKeyNumpadDelete:
		return terminal.KeyDelete
	case ghosttyKeyInsert, ghosttyKeyNumpadInsert:
		return terminal.KeyInsert
	case ghosttyKeyHome, ghosttyKeyNumpadHome:
		return terminal.KeyHome
	case ghosttyKeyEnd, ghosttyKeyNumpadEnd:
		return terminal.KeyEnd
	case ghosttyKeyPageUp, ghosttyKeyNumpadPageUp:
		return terminal.KeyPageUp
	case ghosttyKeyPageDown, ghosttyKeyNumpadPageDown:
		return terminal.KeyPageDown
	case ghosttyKeyArrowUp, ghosttyKeyNumpadUp:
		return terminal.KeyUp
	case ghosttyKeyArrowDown, ghosttyKeyNumpadDown:
		return terminal.KeyDown
	case ghosttyKeyArrowLeft, ghosttyKeyNumpadLeft:
		return terminal.KeyLeft
	case ghosttyKeyArrowRight, ghosttyKeyNumpadRight:
		return terminal.KeyRight
	case ghosttyKeyF1:
		return terminal.KeyF1
	case ghosttyKeyF2:
		return terminal.KeyF2
	case ghosttyKeyF3:
		return terminal.KeyF3
	case ghosttyKeyF4:
		return terminal.KeyF4
	case ghosttyKeyF5:
		return terminal.KeyF5
	case ghosttyKeyF6:
		return terminal.KeyF6
	case ghosttyKeyF7:
		return terminal.KeyF7
	case ghosttyKeyF8:
		return terminal.KeyF8
	case ghosttyKeyF9:
		return terminal.KeyF9
	case ghosttyKeyF10:
		return terminal.KeyF10
	case ghosttyKeyF11:
		return terminal.KeyF11
	case ghosttyKeyF12:
		return terminal.KeyF12
	}
	if ctrl {
		switch key {
		case ghosttyKeyB:
			return terminal.KeyCtrlB
		case ghosttyKeyC:
			return terminal.KeyCtrlC
		case ghosttyKeyD:
			return terminal.KeyCtrlD
		case ghosttyKeyF:
			return terminal.KeyCtrlF
		case ghosttyKeyP:
			return terminal.KeyCtrlP
		case ghosttyKeyV:
			return terminal.KeyCtrlV
		case ghosttyKeyX:
			return terminal.KeyCtrlX
		case ghosttyKeyZ:
			return terminal.KeyCtrlZ
		}
	}
	return terminal.KeyNone
}

func ghosttyMouseButtonToTerminal(button int32) terminal.MouseButton {
	switch button {
	case ghosttyMouseLeft:
		return terminal.MouseLeft
	case ghosttyMouseRight:
		return terminal.MouseRight
	case ghosttyMouseMiddle:
		return terminal.MouseMiddle
	default:
		return terminal.MouseNone
	}
}

func ctrlKeyFromRune(r rune) (terminal.Key, bool) {
	switch unicode.ToLower(r) {
	case 'b':
		return terminal.KeyCtrlB, true
	case 'c':
		return terminal.KeyCtrlC, true
	case 'd':
		return terminal.KeyCtrlD, true
	case 'f':
		return terminal.KeyCtrlF, true
	case 'p':
		return terminal.KeyCtrlP, true
	case 'v':
		return terminal.KeyCtrlV, true
	case 'x':
		return terminal.KeyCtrlX, true
	case 'z':
		return terminal.KeyCtrlZ, true
	default:
		return terminal.KeyNone, false
	}
}

func isCtrlTerminalKey(key terminal.Key) bool {
	switch key {
	case terminal.KeyCtrlB,
		terminal.KeyCtrlC,
		terminal.KeyCtrlD,
		terminal.KeyCtrlF,
		terminal.KeyCtrlP,
		terminal.KeyCtrlV,
		terminal.KeyCtrlX,
		terminal.KeyCtrlZ:
		return true
	default:
		return false
	}
}
