package keybind

import (
	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/runtime"
)

// RuntimeHandler adapts a KeyRouter for runtime.App.
type RuntimeHandler struct {
	Router *KeyRouter
}

// HandleKey handles a runtime key message.
func (h *RuntimeHandler) HandleKey(app *runtime.App, msg runtime.KeyMsg, focused runtime.Widget) bool {
	if h == nil || h.Router == nil {
		return false
	}
	var accessible accessibility.Accessible
	if a, ok := focused.(accessibility.Accessible); ok {
		accessible = a
	}
	ctx := Context{
		FocusedWidget: accessible,
		Focused:       focused,
		App:           app,
	}
	if handler, ok := focused.(Handler); ok {
		ctx.Keymap = handler.Keymap()
	}
	return h.Router.HandleKey(msg, ctx)
}
