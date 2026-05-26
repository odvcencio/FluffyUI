//go:build !js

// Package sim provides a simulation backend for testing.
package sim

import (
	"strings"
	"sync"

	"github.com/gdamore/tcell/v3"
	tcellcolor "github.com/gdamore/tcell/v3/color"
	"m31labs.dev/fluffyui/backend"
	backendtcell "m31labs.dev/fluffyui/backend/tcell"
	"m31labs.dev/fluffyui/terminal"
)

// simCell represents a single cell in the simulation screen.
type simCell struct {
	str   string
	style tcell.Style
	width int
}

// simScreen implements tcell.Screen for testing purposes.
type simScreen struct {
	mu       sync.Mutex
	width    int
	height   int
	cells    [][]simCell
	eventQ   chan tcell.Event
	style    tcell.Style
	cursorX  int
	cursorY  int
	cursorOn bool
}

// newSimScreen creates a new simulation screen with the given dimensions.
func newSimScreen(width, height int) *simScreen {
	s := &simScreen{
		width:   width,
		height:  height,
		eventQ:  make(chan tcell.Event, 128),
		style:   tcell.StyleDefault,
		cursorX: -1,
		cursorY: -1,
	}
	s.cells = make([][]simCell, height)
	for y := 0; y < height; y++ {
		s.cells[y] = make([]simCell, width)
		for x := 0; x < width; x++ {
			s.cells[y][x] = simCell{str: " ", style: tcell.StyleDefault, width: 1}
		}
	}
	return s
}

func (s *simScreen) Init() error                              { return nil }
func (s *simScreen) Fini()                                    { close(s.eventQ) }
func (s *simScreen) Clear()                                   { s.Fill(' ', s.style) }
func (s *simScreen) Show()                                    {}
func (s *simScreen) Sync()                                    {}
func (s *simScreen) CharacterSet() string                     { return "UTF-8" }
func (s *simScreen) RegisterRuneFallback(r rune, subst string) {}
func (s *simScreen) UnregisterRuneFallback(r rune)            {}
func (s *simScreen) Resize(int, int, int, int)                {}
func (s *simScreen) Suspend() error                           { return nil }
func (s *simScreen) Resume() error                            { return nil }
func (s *simScreen) Beep() error                              { return nil }
func (s *simScreen) LockRegion(x, y, w, h int, lock bool)     {}
func (s *simScreen) Tty() (tcell.Tty, bool)                   { return nil, false }
func (s *simScreen) SetTitle(string)                          {}
func (s *simScreen) SetClipboard([]byte)                      {}
func (s *simScreen) GetClipboard()                            {}
func (s *simScreen) HasClipboard() bool                       { return false }
func (s *simScreen) ShowNotification(title, body string)      {}
func (s *simScreen) Terminal() (string, string)               { return "sim", "1.0" }
func (s *simScreen) EnableMouse(...tcell.MouseFlags)          {}
func (s *simScreen) DisableMouse()                            {}
func (s *simScreen) EnablePaste()                             {}
func (s *simScreen) DisablePaste()                            {}
func (s *simScreen) EnableFocus()                             {}
func (s *simScreen) DisableFocus()                            {}
func (s *simScreen) Colors() int                              { return 256 }

func (s *simScreen) Fill(r rune, style tcell.Style) {
	s.mu.Lock()
	defer s.mu.Unlock()
	str := string(r)
	for y := 0; y < s.height; y++ {
		for x := 0; x < s.width; x++ {
			s.cells[y][x] = simCell{str: str, style: style, width: 1}
		}
	}
}

func (s *simScreen) SetStyle(style tcell.Style) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.style = style
}

func (s *simScreen) Size() (int, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.width, s.height
}

func (s *simScreen) SetSize(w, h int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if w == s.width && h == s.height {
		return
	}
	newCells := make([][]simCell, h)
	for y := 0; y < h; y++ {
		newCells[y] = make([]simCell, w)
		for x := 0; x < w; x++ {
			if y < s.height && x < s.width {
				newCells[y][x] = s.cells[y][x]
			} else {
				newCells[y][x] = simCell{str: " ", style: tcell.StyleDefault, width: 1}
			}
		}
	}
	s.width = w
	s.height = h
	s.cells = newCells
}

func (s *simScreen) EventQ() chan tcell.Event {
	return s.eventQ
}

func (s *simScreen) SetContent(x, y int, primary rune, combining []rune, style tcell.Style) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if x < 0 || x >= s.width || y < 0 || y >= s.height {
		return
	}
	str := string(primary)
	for _, c := range combining {
		str += string(c)
	}
	s.cells[y][x] = simCell{str: str, style: style, width: 1}
}

func (s *simScreen) Get(x, y int) (string, tcell.Style, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if x < 0 || x >= s.width || y < 0 || y >= s.height {
		return "", tcell.StyleDefault, 0
	}
	cell := s.cells[y][x]
	return cell.str, cell.style, cell.width
}

func (s *simScreen) Put(x, y int, str string, style tcell.Style) (string, int) {
	if len(str) == 0 {
		return "", 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if x < 0 || x >= s.width || y < 0 || y >= s.height {
		return str, 0
	}
	// Store first grapheme
	runes := []rune(str)
	s.cells[y][x] = simCell{str: string(runes[0]), style: style, width: 1}
	if len(runes) > 1 {
		return string(runes[1:]), 1
	}
	return "", 1
}

func (s *simScreen) PutStr(x, y int, str string) {
	s.PutStrStyled(x, y, str, s.style)
}

func (s *simScreen) PutStrStyled(x, y int, str string, style tcell.Style) {
	for _, r := range str {
		if x >= s.width {
			break
		}
		s.SetContent(x, y, r, nil, style)
		x++
	}
}

func (s *simScreen) ShowCursor(x, y int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cursorX = x
	s.cursorY = y
	s.cursorOn = x >= 0 && y >= 0
}

func (s *simScreen) HideCursor() {
	s.ShowCursor(-1, -1)
}

func (s *simScreen) SetCursorStyle(cs tcell.CursorStyle, colors ...tcellcolor.Color) {}

// Backend is a testable backend using a simulation screen.
type Backend struct {
	*backendtcell.Backend
	screen *simScreen
	mu     sync.Mutex
}

// New creates a new simulation backend with the given dimensions.
func New(width, height int) *Backend {
	screen := newSimScreen(width, height)
	return &Backend{
		Backend: backendtcell.NewWithScreen(screen),
		screen:  screen,
	}
}

// Resize changes the simulation screen size.
func (s *Backend) Resize(width, height int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.screen.SetSize(width, height)
}

// InjectKey injects a key event into the simulation.
func (s *Backend) InjectKey(key terminal.Key, r rune) {
	s.PostEvent(terminal.KeyEvent{Key: key, Rune: r})
}

// InjectKeyRune injects a regular character keypress.
func (s *Backend) InjectKeyRune(r rune) {
	s.InjectKey(terminal.KeyRune, r)
}

// InjectKeyString injects a string as a sequence of key events.
func (s *Backend) InjectKeyString(str string) {
	for _, r := range str {
		s.InjectKeyRune(r)
	}
}

// InjectMouse injects a mouse press event into the simulation.
func (s *Backend) InjectMouse(x, y int, button terminal.MouseButton) {
	s.InjectMouseAction(x, y, button, terminal.MousePress)
}

// InjectMouseAction injects a mouse event with a specific action.
func (s *Backend) InjectMouseAction(x, y int, button terminal.MouseButton, action terminal.MouseAction) {
	s.PostEvent(terminal.MouseEvent{X: x, Y: y, Button: button, Action: action})
}

// InjectResize injects a resize event.
func (s *Backend) InjectResize(width, height int) {
	s.mu.Lock()
	s.screen.SetSize(width, height)
	s.mu.Unlock()
	s.PostEvent(terminal.ResizeEvent{Width: width, Height: height})
}

// Capture captures the current screen content as a string.
func (s *Backend) Capture() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	w, h := s.screen.Size()
	var lines []string

	for y := 0; y < h; y++ {
		var line strings.Builder
		for x := 0; x < w; x++ {
			str, _, _ := s.screen.Get(x, y)
			if str == "" {
				str = " "
			}
			line.WriteString(str)
		}
		lines = append(lines, line.String())
	}

	return strings.Join(lines, "\n")
}

// CaptureCell returns the content and style of a single cell.
func (s *Backend) CaptureCell(x, y int) (mainc rune, comb []rune, style backend.Style) {
	s.mu.Lock()
	defer s.mu.Unlock()

	str, tcStyle, _ := s.screen.Get(x, y)
	runes := []rune(str)
	if len(runes) == 0 {
		return ' ', nil, convertTcellStyle(tcStyle)
	}
	return runes[0], runes[1:], convertTcellStyle(tcStyle)
}

// CaptureRegion captures a rectangular region of the screen.
func (s *Backend) CaptureRegion(x, y, w, h int) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	var lines []string
	for row := y; row < y+h; row++ {
		var line strings.Builder
		for col := x; col < x+w; col++ {
			str, _, _ := s.screen.Get(col, row)
			if str == "" {
				str = " "
			}
			line.WriteString(str)
		}
		lines = append(lines, line.String())
	}
	return strings.Join(lines, "\n")
}

// FindText searches for text on the screen and returns its position.
func (s *Backend) FindText(text string) (x, y int) {
	capture := s.Capture()
	lines := strings.Split(capture, "\n")

	for row, line := range lines {
		if col := strings.Index(line, text); col >= 0 {
			return col, row
		}
	}
	return -1, -1
}

// ContainsText returns true if the text appears anywhere on screen.
func (s *Backend) ContainsText(text string) bool {
	x, y := s.FindText(text)
	return x >= 0 && y >= 0
}

// convertTcellStyle converts tcell.Style to backend.Style.
func convertTcellStyle(ts tcell.Style) backend.Style {
	fg := ts.GetForeground()
	bg := ts.GetBackground()
	style := backend.DefaultStyle().
		Foreground(convertTcellColor(fg)).
		Background(convertTcellColor(bg))

	if ts.HasBold() {
		style = style.Bold(true)
	}
	if ts.HasItalic() {
		style = style.Italic(true)
	}
	if ts.HasUnderline() {
		style = style.Underline(true)
	}
	if ts.HasDim() {
		style = style.Dim(true)
	}
	if ts.HasBlink() {
		style = style.Blink(true)
	}
	if ts.HasReverse() {
		style = style.Reverse(true)
	}
	if ts.HasStrikeThrough() {
		style = style.StrikeThrough(true)
	}

	return style
}

// convertTcellColor converts tcellcolor.Color to backend.Color.
func convertTcellColor(tc tcellcolor.Color) backend.Color {
	if tc == tcellcolor.Default {
		return backend.ColorDefault
	}
	if tc.IsRGB() {
		r, g, b := tc.RGB()
		return backend.ColorRGB(uint8(r), uint8(g), uint8(b))
	}
	return backend.Color(tc & 0xFF)
}

// Clear clears the screen to spaces with default style.
func (s *Backend) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.screen.Clear()
}

// ShowText is a convenience method to display the screen contents for debugging.
// Returns the captured screen content as a string.
func (s *Backend) ShowText() string {
	s.Show()
	return s.Capture()
}

// InjectKeyWithMod injects a key event with modifiers.
func (s *Backend) InjectKeyWithMod(key terminal.Key, r rune, ctrl, alt, shift bool) {
	s.PostEvent(terminal.KeyEvent{Key: key, Rune: r, Ctrl: ctrl, Alt: alt, Shift: shift})
}

// InjectCtrlKey injects a ctrl+key combination.
func (s *Backend) InjectCtrlKey(r rune) {
	s.InjectKeyWithMod(terminal.KeyRune, r, true, false, false)
}

// InjectAltKey injects an alt+key combination.
func (s *Backend) InjectAltKey(r rune) {
	s.InjectKeyWithMod(terminal.KeyRune, r, false, true, false)
}

// Ensure Backend implements backend.Backend
var _ backend.Backend = (*Backend)(nil)
