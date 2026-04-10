//go:build !js

// Package tcell provides a Backend implementation using tcell.
package tcell

import (
	"fmt"
	"os"
	"strings"
	"sync/atomic"

	"github.com/gdamore/tcell/v3"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/terminal"
)

// Backend implements backend.Backend using tcell.
type Backend struct {
	screen tcell.Screen
	raw    *rawTty
	// Inline rendering mode (no alternate screen, bounded viewport).
	// atomic.Bool because SetInlineMode and the read sites are called
	// from different goroutines (test harness, runtime event loop).
	inlineMode   atomic.Bool
	inlineHeight int

	// Bracketed paste state
	inPaste     bool
	pasteBuffer strings.Builder

	styleCache    map[backend.Style]tcell.Style
	styleCacheCap int

	// syncOutput indicates whether the terminal supports DEC mode 2026.
	syncOutput bool
}

// New creates a new tcell backend.
func New() (*Backend, error) {
	tty, ttyErr := tcell.NewDevTty()
	if ttyErr == nil {
		raw := &rawTty{inner: tty}
		screen, err := tcell.NewTerminfoScreenFromTty(raw)
		if err == nil {
			return &Backend{screen: screen, raw: raw}, nil
		}
	}
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	return &Backend{screen: screen}, nil
}

// NewWithScreen creates a backend with an existing tcell screen (for testing).
func NewWithScreen(screen tcell.Screen) *Backend {
	return &Backend{screen: screen}
}

// SetInlineMode enables/disables inline rendering mode.
func (b *Backend) SetInlineMode(enabled bool) {
	if b == nil {
		return
	}
	b.inlineMode.Store(enabled)
	if b.raw != nil {
		b.raw.SetInlineMode(enabled)
	}
}

// SetInlineHeight constrains inline rendering to a fixed number of rows.
// Values <= 0 use the backend default.
func (b *Backend) SetInlineHeight(lines int) {
	if b == nil {
		return
	}
	b.inlineHeight = lines
}

// Init initializes the backend.
func (b *Backend) Init() error {
	if err := b.screen.Init(); err != nil {
		return err
	}
	b.screen.EnableMouse()
	b.screen.EnablePaste()
	return nil
}

// Fini cleans up the backend.
func (b *Backend) Fini() {
	b.screen.Fini()
}

// Size returns the terminal dimensions.
func (b *Backend) Size() (width, height int) {
	if b == nil || b.screen == nil {
		return 0, 0
	}
	width, fullHeight := b.screen.Size()
	if !b.inlineMode.Load() {
		return width, fullHeight
	}
	return width, b.inlineViewportHeight(fullHeight)
}

// SetContent sets a cell at position (x, y).
func (b *Backend) SetContent(x, y int, mainc rune, comb []rune, style backend.Style) {
	if b == nil || b.screen == nil {
		return
	}
	width, fullHeight := b.screen.Size()
	if x < 0 || x >= width || y < 0 {
		return
	}
	if b.inlineMode.Load() {
		y += b.inlineOffsetY()
	}
	if y < 0 || y >= fullHeight {
		return
	}
	b.screen.SetContent(x, y, mainc, comb, b.cachedStyle(style))
}

// SetRow updates a row using the runtime row-writer fast path.
func (b *Backend) SetRow(y int, startX int, cells []backend.Cell) {
	if b == nil || b.screen == nil || startX < 0 || y < 0 || len(cells) == 0 {
		return
	}

	width, fullHeight := b.screen.Size()
	if startX >= width {
		return
	}
	if b.inlineMode.Load() {
		y += b.inlineOffsetY()
	}
	if y < 0 || y >= fullHeight {
		return
	}

	maxCols := min(len(cells), width-startX)
	for i := 0; i < maxCols; i++ {
		cell := cells[i]
		b.screen.SetContent(startX+i, y, cell.Rune, nil, b.cachedStyle(cell.Style))
	}
}

// SetRect updates a rectangle using row-major cells.
func (b *Backend) SetRect(x, y, width, height int, cells []backend.Cell) {
	if b == nil || b.screen == nil || width <= 0 || height <= 0 {
		return
	}
	expected := width * height
	if len(cells) < expected {
		return
	}

	screenWidth, fullHeight := b.screen.Size()
	offsetY := 0
	if b.inlineMode.Load() {
		offsetY = b.inlineOffsetY()
	}

	for row := 0; row < height; row++ {
		dstY := y + row
		if dstY < 0 {
			continue
		}
		absY := dstY + offsetY
		if absY < 0 || absY >= fullHeight {
			continue
		}

		rowStart := row * width
		rowEnd := rowStart + width
		rowCells := cells[rowStart:rowEnd]

		startX := max(0, x)
		endX := min(screenWidth, x+width)
		if startX >= endX {
			continue
		}
		srcOffset := startX - x
		for col := startX; col < endX; col++ {
			cell := rowCells[srcOffset]
			b.screen.SetContent(col, absY, cell.Rune, nil, b.cachedStyle(cell.Style))
			srcOffset++
		}
	}
}

// Show synchronizes the buffer to the terminal.
func (b *Backend) Show() {
	b.screen.Show()
}

// Clear clears the screen.
func (b *Backend) Clear() {
	if b == nil || b.screen == nil {
		return
	}
	if b.inlineMode.Load() {
		width, fullHeight := b.screen.Size()
		top := b.inlineOffsetY()
		height := b.inlineViewportHeight(fullHeight)
		defaultStyle := tcell.StyleDefault
		for y := 0; y < height; y++ {
			row := top + y
			for x := 0; x < width; x++ {
				b.screen.SetContent(x, row, ' ', nil, defaultStyle)
			}
		}
		return
	}
	b.screen.Clear()
}

// HideCursor hides the cursor.
func (b *Backend) HideCursor() {
	b.screen.HideCursor()
}

// ShowCursor shows the cursor.
func (b *Backend) ShowCursor() {
	// tcell shows cursor when we set its position
}

// SetCursorPos sets the cursor position.
func (b *Backend) SetCursorPos(x, y int) {
	if b.inlineMode.Load() {
		y += b.inlineOffsetY()
	}
	b.screen.ShowCursor(x, y)
}

// PollEvent blocks until an event is available.
func (b *Backend) PollEvent() terminal.Event {
	for {
		ev, ok := <-b.screen.EventQ()
		if !ok || ev == nil {
			return nil
		}

		// Handle bracketed paste state machine
		switch e := ev.(type) {
		case *tcell.EventPaste:
			if e.Start() {
				// Begin paste mode, buffer subsequent key events
				b.inPaste = true
				b.pasteBuffer.Reset()
				continue
			}
			if e.End() {
				// End paste mode, emit PasteEvent with accumulated content
				b.inPaste = false
				text := b.pasteBuffer.String()
				b.pasteBuffer.Reset()
				if text != "" {
					return terminal.PasteEvent{Text: text}
				}
				continue
			}

		case *tcell.EventKey:
			if b.inPaste {
				// Accumulate string during paste (v3 uses Str() instead of Rune())
				if e.Key() == tcell.KeyRune {
					b.pasteBuffer.WriteString(e.Str())
				} else if e.Key() == tcell.KeyEnter {
					b.pasteBuffer.WriteRune('\n')
				} else if e.Key() == tcell.KeyTab {
					b.pasteBuffer.WriteRune('\t')
				}
				continue
			}
		}

		// Normal event handling
		mapped := b.mapInlineEvent(convertEvent(ev))
		if mapped == nil {
			continue
		}
		return mapped
	}
}

// PostEvent injects an event into the queue.
func (b *Backend) PostEvent(ev terminal.Event) error {
	if b.inlineMode.Load() {
		if mouse, ok := ev.(terminal.MouseEvent); ok {
			mouse.Y += b.inlineOffsetY()
			ev = mouse
		}
	}
	tev := reverseConvertEvent(ev)
	if tev == nil {
		return nil
	}

	// In tcell v3, we write directly to the EventQ channel
	select {
	case b.screen.EventQ() <- tev:
		return nil
	default:
		// Channel full, try blocking send
		b.screen.EventQ() <- tev
		return nil
	}
}

// Beep emits an audible bell.
func (b *Backend) Beep() {
	b.screen.Beep()
}

// Sync forces a full redraw.
func (b *Backend) Sync() {
	b.screen.Sync()
}

func (b *Backend) inlineViewportHeight(fullHeight int) int {
	if fullHeight <= 0 {
		return 0
	}
	view := b.inlineHeight
	if view <= 0 {
		view = 12
	}
	if view > fullHeight {
		view = fullHeight
	}
	return view
}

func (b *Backend) inlineOffsetY() int {
	if b == nil || b.screen == nil || !b.inlineMode.Load() {
		return 0
	}
	_, fullHeight := b.screen.Size()
	view := b.inlineViewportHeight(fullHeight)
	if view >= fullHeight {
		return 0
	}
	return fullHeight - view
}

func (b *Backend) mapInlineEvent(ev terminal.Event) terminal.Event {
	if !b.inlineMode.Load() || ev == nil {
		return ev
	}
	switch e := ev.(type) {
	case terminal.ResizeEvent:
		e.Height = b.inlineViewportHeight(e.Height)
		return e
	case terminal.MouseEvent:
		relativeY := e.Y - b.inlineOffsetY()
		if relativeY < 0 {
			return nil
		}
		width, _ := b.Size()
		_, height := b.Size()
		if e.X < 0 || e.X >= width || relativeY >= height {
			return nil
		}
		e.Y = relativeY
		return e
	default:
		return ev
	}
}

const defaultStyleCacheCap = 256

func (b *Backend) cachedStyle(s backend.Style) tcell.Style {
	if b.styleCache == nil {
		b.styleCacheCap = defaultStyleCacheCap
		b.styleCache = make(map[backend.Style]tcell.Style, b.styleCacheCap)
	}
	if cached, ok := b.styleCache[s]; ok {
		return cached
	}
	if b.styleCacheCap <= 0 {
		b.styleCacheCap = defaultStyleCacheCap
	}
	if len(b.styleCache) >= b.styleCacheCap {
		b.styleCache = make(map[backend.Style]tcell.Style, b.styleCacheCap)
	}
	style := convertStyle(s)
	b.styleCache[s] = style
	return style
}

// convertStyle converts backend.Style to tcell.Style.
func convertStyle(s backend.Style) tcell.Style {
	fg, bg, attrs := s.Decompose()
	style := tcell.StyleDefault.
		Foreground(convertColor(fg)).
		Background(convertColor(bg))

	if attrs&backend.AttrBold != 0 {
		style = style.Bold(true)
	}
	if attrs&backend.AttrItalic != 0 {
		style = style.Italic(true)
	}
	if attrs&backend.AttrUnderline != 0 {
		style = style.Underline(true)
	}
	if attrs&backend.AttrDim != 0 {
		style = style.Dim(true)
	}
	if attrs&backend.AttrBlink != 0 {
		style = style.Blink(true)
	}
	if attrs&backend.AttrReverse != 0 {
		style = style.Reverse(true)
	}
	if attrs&backend.AttrStrikeThrough != 0 {
		style = style.StrikeThrough(true)
	}

	return style
}

// convertColor converts backend.Color to tcell.Color.
func convertColor(c backend.Color) tcell.Color {
	if c == backend.ColorDefault {
		return tcell.ColorDefault
	}
	if c.IsRGB() {
		r, g, b := c.RGB()
		return tcell.NewRGBColor(int32(r), int32(g), int32(b))
	}
	return tcell.PaletteColor(int(c))
}

// convertEvent converts a tcell event to terminal.Event.
func convertEvent(ev tcell.Event) terminal.Event {
	switch e := ev.(type) {
	case *tcell.EventKey:
		// In tcell v3, Str() returns a string instead of Rune()
		// Extract first rune for compatibility
		str := e.Str()
		r := rune(0)
		if len(str) > 0 {
			r = []rune(str)[0]
		}
		return terminal.KeyEvent{
			Key:   convertKey(e.Key()),
			Rune:  r,
			Alt:   e.Modifiers()&tcell.ModAlt != 0,
			Ctrl:  e.Modifiers()&tcell.ModCtrl != 0,
			Shift: e.Modifiers()&tcell.ModShift != 0,
		}
	case *tcell.EventResize:
		w, h := e.Size()
		return terminal.ResizeEvent{Width: w, Height: h}
	case *tcell.EventMouse:
		x, y := e.Position()
		mods := e.Modifiers()
		return terminal.MouseEvent{
			X:      x,
			Y:      y,
			Button: convertMouseButton(e.Buttons()),
			Action: convertMouseAction(e.Buttons()),
			Alt:    mods&tcell.ModAlt != 0,
			Ctrl:   mods&tcell.ModCtrl != 0,
			Shift:  mods&tcell.ModShift != 0,
		}
	default:
		return nil
	}
}

// convertKey converts tcell.Key to terminal.Key.
func convertKey(k tcell.Key) terminal.Key {
	switch k {
	case tcell.KeyRune:
		return terminal.KeyRune
	case tcell.KeyUp:
		return terminal.KeyUp
	case tcell.KeyDown:
		return terminal.KeyDown
	case tcell.KeyRight:
		return terminal.KeyRight
	case tcell.KeyLeft:
		return terminal.KeyLeft
	case tcell.KeyPgUp:
		return terminal.KeyPageUp
	case tcell.KeyPgDn:
		return terminal.KeyPageDown
	case tcell.KeyHome:
		return terminal.KeyHome
	case tcell.KeyEnd:
		return terminal.KeyEnd
	case tcell.KeyInsert:
		return terminal.KeyInsert
	case tcell.KeyDelete:
		return terminal.KeyDelete
	case tcell.KeyBackspace, tcell.KeyDEL:
		return terminal.KeyBackspace
	case tcell.KeyTab:
		return terminal.KeyTab
	case tcell.KeyEnter:
		return terminal.KeyEnter
	case tcell.KeyEscape:
		return terminal.KeyEscape
	case tcell.KeyCtrlA:
		return terminal.KeyCtrlA
	case tcell.KeyCtrlB:
		return terminal.KeyCtrlB
	case tcell.KeyCtrlC:
		return terminal.KeyCtrlC
	case tcell.KeyCtrlD:
		return terminal.KeyCtrlD
	case tcell.KeyCtrlF:
		return terminal.KeyCtrlF
	case tcell.KeyCtrlP:
		return terminal.KeyCtrlP
	case tcell.KeyCtrlV:
		return terminal.KeyCtrlV
	case tcell.KeyCtrlX:
		return terminal.KeyCtrlX
	case tcell.KeyCtrlY:
		return terminal.KeyCtrlY
	case tcell.KeyCtrlZ:
		return terminal.KeyCtrlZ
	case tcell.KeyCtrlE:
		return terminal.KeyCtrlE
	case tcell.KeyCtrlK:
		return terminal.KeyCtrlK
	case tcell.KeyCtrlL:
		return terminal.KeyCtrlL
	case tcell.KeyCtrlN:
		return terminal.KeyCtrlN
	case tcell.KeyCtrlO:
		return terminal.KeyCtrlO
	case tcell.KeyCtrlQ:
		return terminal.KeyCtrlQ
	case tcell.KeyCtrlR:
		return terminal.KeyCtrlR
	case tcell.KeyCtrlS:
		return terminal.KeyCtrlS
	case tcell.KeyCtrlT:
		return terminal.KeyCtrlT
	case tcell.KeyCtrlU:
		return terminal.KeyCtrlU
	case tcell.KeyCtrlW:
		return terminal.KeyCtrlW
	case tcell.KeyF1:
		return terminal.KeyF1
	case tcell.KeyF2:
		return terminal.KeyF2
	case tcell.KeyF3:
		return terminal.KeyF3
	case tcell.KeyF4:
		return terminal.KeyF4
	case tcell.KeyF5:
		return terminal.KeyF5
	case tcell.KeyF6:
		return terminal.KeyF6
	case tcell.KeyF7:
		return terminal.KeyF7
	case tcell.KeyF8:
		return terminal.KeyF8
	case tcell.KeyF9:
		return terminal.KeyF9
	case tcell.KeyF10:
		return terminal.KeyF10
	case tcell.KeyF11:
		return terminal.KeyF11
	case tcell.KeyF12:
		return terminal.KeyF12
	default:
		return terminal.KeyNone
	}
}

// convertMouseButton converts tcell button mask to terminal.MouseButton.
func convertMouseButton(buttons tcell.ButtonMask) terminal.MouseButton {
	switch {
	case buttons&tcell.WheelUp != 0:
		return terminal.MouseWheelUp
	case buttons&tcell.WheelDown != 0:
		return terminal.MouseWheelDown
	case buttons&tcell.Button1 != 0:
		return terminal.MouseLeft
	case buttons&tcell.Button2 != 0:
		return terminal.MouseMiddle
	case buttons&tcell.Button3 != 0:
		return terminal.MouseRight
	default:
		return terminal.MouseNone
	}
}

// convertMouseAction determines the mouse action from button state.
func convertMouseAction(buttons tcell.ButtonMask) terminal.MouseAction {
	if buttons == tcell.ButtonNone {
		return terminal.MouseRelease
	}
	if buttons&(tcell.WheelUp|tcell.WheelDown) != 0 {
		return terminal.MousePress // Wheel events are instantaneous
	}
	return terminal.MousePress
}

// reverseConvertEvent converts terminal.Event to tcell.Event for PostEvent.
func reverseConvertEvent(ev terminal.Event) tcell.Event {
	switch e := ev.(type) {
	case terminal.KeyEvent:
		mod := tcell.ModNone
		if e.Alt {
			mod |= tcell.ModAlt
		}
		if e.Ctrl {
			mod |= tcell.ModCtrl
		}
		if e.Shift {
			mod |= tcell.ModShift
		}
		return tcell.NewEventKey(reverseConvertKey(e.Key), string(e.Rune), mod)
	case terminal.MouseEvent:
		mod := tcell.ModNone
		if e.Alt {
			mod |= tcell.ModAlt
		}
		if e.Ctrl {
			mod |= tcell.ModCtrl
		}
		if e.Shift {
			mod |= tcell.ModShift
		}
		buttons := reverseConvertMouseButton(e.Button)
		if e.Action == terminal.MouseRelease {
			buttons = tcell.ButtonNone
		}
		return tcell.NewEventMouse(e.X, e.Y, buttons, mod)
	case terminal.ResizeEvent:
		return tcell.NewEventResize(e.Width, e.Height)
	default:
		return nil
	}
}

// reverseConvertKey converts terminal.Key to tcell.Key.
func reverseConvertKey(k terminal.Key) tcell.Key {
	switch k {
	case terminal.KeyRune:
		return tcell.KeyRune
	case terminal.KeyUp:
		return tcell.KeyUp
	case terminal.KeyDown:
		return tcell.KeyDown
	case terminal.KeyRight:
		return tcell.KeyRight
	case terminal.KeyLeft:
		return tcell.KeyLeft
	case terminal.KeyPageUp:
		return tcell.KeyPgUp
	case terminal.KeyPageDown:
		return tcell.KeyPgDn
	case terminal.KeyHome:
		return tcell.KeyHome
	case terminal.KeyEnd:
		return tcell.KeyEnd
	case terminal.KeyInsert:
		return tcell.KeyInsert
	case terminal.KeyDelete:
		return tcell.KeyDelete
	case terminal.KeyBackspace:
		return tcell.KeyDEL
	case terminal.KeyTab:
		return tcell.KeyTab
	case terminal.KeyEnter:
		return tcell.KeyEnter
	case terminal.KeyEscape:
		return tcell.KeyEscape
	case terminal.KeyCtrlA:
		return tcell.KeyCtrlA
	case terminal.KeyCtrlB:
		return tcell.KeyCtrlB
	case terminal.KeyCtrlC:
		return tcell.KeyCtrlC
	case terminal.KeyCtrlD:
		return tcell.KeyCtrlD
	case terminal.KeyCtrlF:
		return tcell.KeyCtrlF
	case terminal.KeyCtrlP:
		return tcell.KeyCtrlP
	case terminal.KeyCtrlV:
		return tcell.KeyCtrlV
	case terminal.KeyCtrlX:
		return tcell.KeyCtrlX
	case terminal.KeyCtrlY:
		return tcell.KeyCtrlY
	case terminal.KeyCtrlZ:
		return tcell.KeyCtrlZ
	case terminal.KeyCtrlE:
		return tcell.KeyCtrlE
	case terminal.KeyCtrlK:
		return tcell.KeyCtrlK
	case terminal.KeyCtrlL:
		return tcell.KeyCtrlL
	case terminal.KeyCtrlN:
		return tcell.KeyCtrlN
	case terminal.KeyCtrlO:
		return tcell.KeyCtrlO
	case terminal.KeyCtrlQ:
		return tcell.KeyCtrlQ
	case terminal.KeyCtrlR:
		return tcell.KeyCtrlR
	case terminal.KeyCtrlS:
		return tcell.KeyCtrlS
	case terminal.KeyCtrlT:
		return tcell.KeyCtrlT
	case terminal.KeyCtrlU:
		return tcell.KeyCtrlU
	case terminal.KeyCtrlW:
		return tcell.KeyCtrlW
	case terminal.KeyF1:
		return tcell.KeyF1
	case terminal.KeyF2:
		return tcell.KeyF2
	case terminal.KeyF3:
		return tcell.KeyF3
	case terminal.KeyF4:
		return tcell.KeyF4
	case terminal.KeyF5:
		return tcell.KeyF5
	case terminal.KeyF6:
		return tcell.KeyF6
	case terminal.KeyF7:
		return tcell.KeyF7
	case terminal.KeyF8:
		return tcell.KeyF8
	case terminal.KeyF9:
		return tcell.KeyF9
	case terminal.KeyF10:
		return tcell.KeyF10
	case terminal.KeyF11:
		return tcell.KeyF11
	case terminal.KeyF12:
		return tcell.KeyF12
	default:
		return tcell.KeyRune
	}
}

// reverseConvertMouseButton converts terminal.MouseButton to tcell.ButtonMask.
func reverseConvertMouseButton(b terminal.MouseButton) tcell.ButtonMask {
	switch b {
	case terminal.MouseLeft:
		return tcell.Button1
	case terminal.MouseMiddle:
		return tcell.Button2
	case terminal.MouseRight:
		return tcell.Button3
	case terminal.MouseWheelUp:
		return tcell.WheelUp
	case terminal.MouseWheelDown:
		return tcell.WheelDown
	default:
		return tcell.ButtonNone
	}
}

// SetCursorShape changes the terminal cursor appearance using DECSCUSR.
func (b *Backend) SetCursorShape(shape backend.CursorShape) {
	var ps int
	switch shape {
	case backend.CursorBlock:
		ps = 2
	case backend.CursorUnderline:
		ps = 4
	case backend.CursorBeam:
		ps = 6
	default:
		ps = 0
	}
	seq := fmt.Sprintf("\x1b[%d q", ps)
	if b.raw != nil {
		_ = b.raw.WriteRaw([]byte(seq))
	} else {
		fmt.Fprint(os.Stdout, seq)
	}
}

// SetSyncOutput enables or disables synchronized output support.
// Call this after detecting terminal capabilities (see terminal.TerminalCaps).
func (b *Backend) SetSyncOutput(enabled bool) {
	if b == nil {
		return
	}
	b.syncOutput = enabled
}

// SyncOutputSupported reports whether synchronized output (DEC mode 2026)
// is enabled for this backend.
func (b *Backend) SyncOutputSupported() bool {
	if b == nil {
		return false
	}
	return b.syncOutput
}

// BeginSync writes the DEC mode 2026 enable sequence to the terminal.
func (b *Backend) BeginSync() {
	if b == nil || !b.syncOutput {
		return
	}
	if b.raw != nil {
		_ = b.raw.WriteRaw([]byte("\x1b[?2026h"))
	} else {
		fmt.Fprint(os.Stdout, "\x1b[?2026h")
	}
}

// EndSync writes the DEC mode 2026 disable sequence to the terminal.
func (b *Backend) EndSync() {
	if b == nil || !b.syncOutput {
		return
	}
	if b.raw != nil {
		_ = b.raw.WriteRaw([]byte("\x1b[?2026l"))
	} else {
		fmt.Fprint(os.Stdout, "\x1b[?2026l")
	}
}

// Ensure Backend implements backend.Backend and optional interfaces.
var _ backend.Backend = (*Backend)(nil)
var _ backend.CursorShapeSetter = (*Backend)(nil)
var _ backend.SyncOutputWriter = (*Backend)(nil)
