package keybind

import "github.com/odvcencio/fluffy-ui/clipboard"

// DefaultKeymap returns a baseline global keymap.
func DefaultKeymap() *Keymap {
	return &Keymap{
		Name: "default",
		Bindings: []Binding{
			{Key: MustParseKeySequence("ctrl+c"), Command: clipboard.CommandCopy, When: WhenFocusedClipboardTarget()},
			{Key: MustParseKeySequence("ctrl+x"), Command: clipboard.CommandCut, When: WhenFocusedClipboardTarget()},
			{Key: MustParseKeySequence("ctrl+v"), Command: clipboard.CommandPaste, When: WhenFocusedClipboardTarget()},
			{Key: MustParseKeySequence("ctrl+c"), Command: "app.quit", When: WhenFocusedNotClipboardTarget()},
			{Key: MustParseKeySequence("tab"), Command: "focus.next"},
			{Key: MustParseKeySequence("shift+tab"), Command: "focus.prev"},
			{Key: MustParseKeySequence("pgup"), Command: "scroll.pageUp"},
			{Key: MustParseKeySequence("pgdn"), Command: "scroll.pageDown"},
			{Key: MustParseKeySequence("home"), Command: "scroll.home"},
			{Key: MustParseKeySequence("end"), Command: "scroll.end"},
		},
	}
}
