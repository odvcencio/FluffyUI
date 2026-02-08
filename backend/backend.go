// Package backend defines the terminal backend interface for the TUI.
// This abstraction allows swapping between tcell (real terminals) and
// simulation backends (testing), enabling golden-frame tests.
package backend

import "github.com/odvcencio/fluffyui/terminal"

// Backend is the terminal abstraction layer.
// Implementations handle terminal I/O, input events, and screen rendering.
type Backend interface {
	// Init initializes the backend (enters alt screen, raw mode, etc).
	Init() error

	// Fini cleans up the backend (restores terminal state).
	Fini()

	// Size returns the current terminal dimensions.
	Size() (width, height int)

	// SetContent sets a cell at position (x, y) with the given rune and style.
	// The comb parameter contains combining characters (can be nil).
	SetContent(x, y int, mainc rune, comb []rune, style Style)

	// Show synchronizes the internal buffer to the terminal.
	// This is where actual output happens.
	Show()

	// Clear clears the screen.
	Clear()

	// HideCursor hides the terminal cursor.
	HideCursor()

	// ShowCursor shows the terminal cursor.
	ShowCursor()

	// SetCursorPos sets the cursor position.
	SetCursorPos(x, y int)

	// PollEvent blocks until an event is available and returns it.
	// Returns nil if the backend is shutting down.
	PollEvent() terminal.Event

	// PostEvent injects an event into the event queue.
	// Useful for testing and for posting timer/internal events.
	PostEvent(ev terminal.Event) error

	// Beep emits an audible bell.
	Beep()

	// Sync forces a full redraw on next Show().
	Sync()
}

// InlineModeSetter allows backends to opt into inline rendering mode.
// Inline mode should avoid switching to the alternate screen when possible.
type InlineModeSetter interface {
	SetInlineMode(enabled bool)
}

// InlineHeightSetter allows backends to constrain inline rendering to a fixed
// number of terminal rows.
type InlineHeightSetter interface {
	SetInlineHeight(lines int)
}

// CursorShape controls the terminal cursor appearance.
type CursorShape int

const (
	CursorDefault   CursorShape = iota // Terminal default
	CursorBlock                        // Block cursor
	CursorUnderline                    // Underline cursor
	CursorBeam                         // Beam/bar cursor
)

// CursorShapeSetter allows backends to control cursor appearance.
// This is optional — backends that don't support it will be skipped.
type CursorShapeSetter interface {
	SetCursorShape(shape CursorShape)
}

// RenderTarget is a subset of Backend for rendering operations only.
// Widgets render to this interface, not the full Backend.
type RenderTarget interface {
	Size() (width, height int)
	SetContent(x, y int, mainc rune, comb []rune, style Style)
}

// SubTarget wraps a RenderTarget with an offset for sub-region rendering.
type SubTarget struct {
	parent  RenderTarget
	offsetX int
	offsetY int
	width   int
	height  int
}

// NewSubTarget creates a sub-region of a RenderTarget.
func NewSubTarget(parent RenderTarget, x, y, w, h int) *SubTarget {
	return &SubTarget{
		parent:  parent,
		offsetX: x,
		offsetY: y,
		width:   w,
		height:  h,
	}
}

// Size returns the sub-target dimensions.
func (s *SubTarget) Size() (width, height int) {
	return s.width, s.height
}

// SetContent sets content with coordinates relative to the sub-target.
func (s *SubTarget) SetContent(x, y int, mainc rune, comb []rune, style Style) {
	// Clip to bounds
	if x < 0 || x >= s.width || y < 0 || y >= s.height {
		return
	}
	s.parent.SetContent(s.offsetX+x, s.offsetY+y, mainc, comb, style)
}
