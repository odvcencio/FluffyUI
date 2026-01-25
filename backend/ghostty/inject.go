//go:build linux || darwin || windows

package ghostty

import (
	"runtime"

	"github.com/odvcencio/fluffy-ui/terminal"
)

func (b *Backend) InjectKey(ev terminal.KeyEvent) bool {
	if b == nil {
		return false
	}
	key, textBuf := ghosttyInputKeyFromTerminal(ev)
	if key.Keycode == 0 && key.Text == nil {
		return false
	}

	b.mu.Lock()
	if b.surface == nil || b.lib == nil || b.lib.surfaceKey == nil {
		b.mu.Unlock()
		_ = b.PostEvent(ev)
		return false
	}
	handled := b.lib.surfaceKey(b.surface, key)
	b.mu.Unlock()
	runtime.KeepAlive(textBuf)
	return handled
}

func (b *Backend) InjectMouse(ev terminal.MouseEvent) bool {
	if b == nil {
		return false
	}

	switch ev.Action {
	case terminal.MouseMove, terminal.MousePress, terminal.MouseRelease:
	default:
		return false
	}

	needsScroll := ev.Button == terminal.MouseWheelUp || ev.Button == terminal.MouseWheelDown
	b.mu.Lock()
	if b.surface == nil || b.lib == nil {
		b.mu.Unlock()
		_ = b.PostEvent(ev)
		return false
	}
	if b.lib.surfaceMousePos == nil {
		b.mu.Unlock()
		_ = b.PostEvent(ev)
		return false
	}
	if needsScroll && b.lib.surfaceMouseScroll == nil {
		b.mu.Unlock()
		_ = b.PostEvent(ev)
		return false
	}
	if !needsScroll && ev.Action != terminal.MouseMove && b.lib.surfaceMouseButton == nil {
		b.mu.Unlock()
		_ = b.PostEvent(ev)
		return false
	}

	mods := ghosttyModsFromTerminal(ev.Alt, ev.Ctrl, ev.Shift)
	b.ensureMousePosLocked(ev.X, ev.Y, mods)

	switch ev.Action {
	case terminal.MouseMove:
		b.mu.Unlock()
		return true
	case terminal.MousePress, terminal.MouseRelease:
		if needsScroll {
			delta := 1.0
			if ev.Button == terminal.MouseWheelDown {
				delta = -1.0
			}
			b.lib.surfaceMouseScroll(b.surface, 0, delta, 0)
			b.mu.Unlock()
			return true
		}
		button, ok := ghosttyMouseButtonFromTerminal(ev.Button)
		if !ok {
			b.mu.Unlock()
			return false
		}
		action := int32(ghosttyMousePress)
		if ev.Action == terminal.MouseRelease {
			action = int32(ghosttyMouseRelease)
		}
		handled := b.lib.surfaceMouseButton(b.surface, action, button, mods)
		b.mu.Unlock()
		return handled
	default:
		b.mu.Unlock()
		return false
	}
}

func ghosttyInputKeyFromTerminal(ev terminal.KeyEvent) (ghosttyInputKey, []byte) {
	mods := ghosttyModsFromTerminal(ev.Alt, ev.Ctrl, ev.Shift)
	keycode, ok := ghosttyKeycodeFromTerminalKey(ev.Key)
	if !ok && ev.Rune != 0 {
		if code, ok := ghosttyKeycodeFromRune(ev.Rune); ok {
			keycode = code
		}
	}
	key := ghosttyInputKey{
		Action:             ghosttyActionPress,
		Mods:               mods,
		ConsumedMods:       0,
		Keycode:            keycode,
		Text:               nil,
		UnshiftedCodepoint: 0,
		Composing:          false,
	}

	var textBuf []byte
	if ev.Rune != 0 {
		textBuf = []byte(string(ev.Rune))
		textBuf = append(textBuf, 0)
		key.Text = &textBuf[0]
		key.UnshiftedCodepoint = uint32(ev.Rune)
	}

	return key, textBuf
}

func ghosttyModsFromTerminal(alt, ctrl, shift bool) int32 {
	var mods int32
	if shift {
		mods |= ghosttyModShift
	}
	if ctrl {
		mods |= ghosttyModCtrl
	}
	if alt {
		mods |= ghosttyModAlt
	}
	return mods
}

func ghosttyMouseButtonFromTerminal(button terminal.MouseButton) (int32, bool) {
	switch button {
	case terminal.MouseLeft:
		return ghosttyMouseLeft, true
	case terminal.MouseRight:
		return ghosttyMouseRight, true
	case terminal.MouseMiddle:
		return ghosttyMouseMiddle, true
	default:
		return ghosttyMouseUnknown, false
	}
}

func (b *Backend) ensureMousePosLocked(x, y int, mods int32) {
	if b.surface == nil || b.lib == nil || b.lib.surfaceMousePos == nil {
		return
	}
	if b.mousePosValid && b.mouseX == x && b.mouseY == y {
		return
	}
	px, py := b.cellToPixelLocked(x, y)
	b.lib.surfaceMousePos(b.surface, px, py, mods)
	b.mouseX = x
	b.mouseY = y
	b.mousePosValid = true
}

func (b *Backend) cellToPixelLocked(x, y int) (float64, float64) {
	if b.surface == nil || b.lib == nil || b.lib.surfaceSize == nil {
		return float64(x), float64(y)
	}
	size := b.lib.surfaceSize(b.surface)
	if size.CellWidthPx == 0 || size.CellHeightPx == 0 {
		return float64(x), float64(y)
	}
	return float64(x) * float64(size.CellWidthPx), float64(y) * float64(size.CellHeightPx)
}
