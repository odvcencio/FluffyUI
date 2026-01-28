package keybind

import (
	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

// Context provides data for key binding conditions and handlers.
type Context struct {
	FocusedWidget accessibility.Accessible
	Focused       runtime.Widget
	Keymap        *Keymap
	App           *runtime.App
}

// Handler exposes a keymap for widgets.
type Handler interface {
	Keymap() *Keymap
}

// HandlerBase provides a simple keymap implementation.
type HandlerBase struct {
	Map *Keymap
}

// Keymap returns the bound keymap.
func (h *HandlerBase) Keymap() *Keymap {
	if h == nil {
		return nil
	}
	return h.Map
}

// KeyRouter routes key presses to commands.
type KeyRouter struct {
	registry *CommandRegistry
	modes    *ModeManager
	keymaps  *KeymapStack
	sequence []KeyPress
}

// NewKeyRouter constructs a router.
func NewKeyRouter(registry *CommandRegistry, modes *ModeManager, keymaps *KeymapStack) *KeyRouter {
	return &KeyRouter{
		registry: registry,
		modes:    modes,
		keymaps:  keymaps,
	}
}

// Reset clears any pending key sequence.
func (r *KeyRouter) Reset() {
	if r == nil {
		return
	}
	r.sequence = nil
}

// HandleKey handles a single key press.
func (r *KeyRouter) HandleKey(msg runtime.KeyMsg, ctx Context) bool {
	if r == nil {
		return false
	}
	press := KeyPressFromKeyMsg(msg)
	r.sequence = append(r.sequence, press)

	match, ok := r.match(r.sequence, ctx)
	if ok {
		return true
	}
	if match.Binding != nil {
		return r.execute(match.Binding, ctx)
	}

	// No match; retry with just the current key.
	r.sequence = []KeyPress{press}
	match, ok = r.match(r.sequence, ctx)
	if ok {
		return true
	}
	if match.Binding != nil {
		return r.execute(match.Binding, ctx)
	}
	r.sequence = nil
	return false
}

func (r *KeyRouter) match(seq []KeyPress, ctx Context) (keyMatch, bool) {
	match := r.matchKeymaps(seq, ctx)
	if match.Prefix {
		return match, true
	}
	if match.Binding != nil {
		r.sequence = nil
		return match, false
	}
	return match, false
}

func (r *KeyRouter) matchKeymaps(seq []KeyPress, ctx Context) keyMatch {
	if r == nil {
		return keyMatch{}
	}
	if ctx.Keymap != nil {
		if match := ctx.Keymap.Match(seq, ctx); match.Binding != nil || match.Prefix {
			return match
		}
	}
	if r.keymaps != nil {
		if match := r.keymaps.Match(seq, ctx); match.Binding != nil || match.Prefix {
			return match
		}
	}
	if r.modes != nil {
		if current := r.modes.Current(); current != nil {
			return current.Match(seq, ctx)
		}
	}
	return keyMatch{}
}

func (r *KeyRouter) execute(binding *Binding, ctx Context) bool {
	if r == nil || binding == nil || binding.Command == "" {
		return false
	}
	if r.registry == nil {
		return false
	}
	return r.registry.Execute(binding.Command, ctx)
}

// KeyPressFromKeyMsg normalizes a runtime key message.
func KeyPressFromKeyMsg(msg runtime.KeyMsg) KeyPress {
	press := KeyPress{
		Key:   msg.Key,
		Rune:  msg.Rune,
		Alt:   msg.Alt,
		Ctrl:  msg.Ctrl,
		Shift: msg.Shift,
	}
	if isCtrlKey(msg.Key) {
		press.Ctrl = true
	}
	return press
}

// KeyPressFromTerminal normalizes a terminal key event.
func KeyPressFromTerminal(event terminal.KeyEvent) KeyPress {
	return KeyPress{
		Key:   event.Key,
		Rune:  event.Rune,
		Alt:   event.Alt,
		Ctrl:  event.Ctrl,
		Shift: event.Shift,
	}
}

func isCtrlKey(key terminal.Key) bool {
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
