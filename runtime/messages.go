package runtime

import (
	"time"

	"github.com/odvcencio/fluffyui/terminal"
)

// Message represents an event flowing into the UI.
// Messages come from terminal input, timers, or background goroutines.
type Message interface {
	isMessage()
}

// KeyMsg represents a keyboard input event.
type KeyMsg struct {
	Key   terminal.Key
	Rune  rune
	Alt   bool
	Ctrl  bool
	Shift bool
}

func (KeyMsg) isMessage() {}

// ResizeMsg indicates the terminal size changed.
type ResizeMsg struct {
	Width  int
	Height int
}

func (ResizeMsg) isMessage() {}

// MouseMsg represents a mouse input event.
type MouseMsg struct {
	X, Y   int
	Button MouseButton
	Action MouseAction
	Alt    bool
	Ctrl   bool
	Shift  bool
}

func (MouseMsg) isMessage() {}

// PasteMsg represents pasted text from bracketed paste mode.
type PasteMsg struct {
	Text string
}

func (PasteMsg) isMessage() {}

// MouseButton identifies which mouse button was involved.
type MouseButton int

const (
	MouseNone MouseButton = iota
	MouseLeft
	MouseMiddle
	MouseRight
	MouseWheelUp
	MouseWheelDown
)

// MouseAction identifies what happened with the mouse.
type MouseAction int

const (
	MousePress MouseAction = iota
	MouseRelease
	MouseMove
)

// TickMsg is sent on each frame tick for animations.
type TickMsg struct {
	Time time.Time
}

func (TickMsg) isMessage() {}

// FocusChangedMsg reports a focus transition.
type FocusChangedMsg struct {
	Prev Focusable
	Next Focusable
}

func (FocusChangedMsg) isMessage() {}

// QueueFlushMsg triggers a state queue flush in the update loop.
type QueueFlushMsg struct{}

func (QueueFlushMsg) isMessage() {}

// InvalidateMsg requests a render pass without forcing a full redraw.
type InvalidateMsg struct{}

func (InvalidateMsg) isMessage() {}

// CustomMsg allows applications to define their own message types.
// The Value field can hold any application-specific data.
//
// Example usage:
//
//	type MyStreamChunk struct {
//	    SessionID string
//	    Text      string
//	}
//
//	app.Post(runtime.CustomMsg{Value: MyStreamChunk{...}})
//
//	// In update function:
//	case runtime.CustomMsg:
//	    switch v := m.Value.(type) {
//	    case MyStreamChunk:
//	        // handle stream chunk
//	    }
type CustomMsg struct {
	Value any
}

func (CustomMsg) isMessage() {}
