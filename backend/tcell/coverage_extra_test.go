package tcell

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/gdamore/tcell/v3"
	tcellcolor "github.com/gdamore/tcell/v3/color"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/terminal"
)

type fakeTty struct {
	started  bool
	stopped  bool
	drained  bool
	closed   bool
	resizeCh chan<- bool
	wrote    [][]byte
	readData []byte
}

func (f *fakeTty) Start() error { f.started = true; return nil }
func (f *fakeTty) Stop() error  { f.stopped = true; return nil }
func (f *fakeTty) Drain() error { f.drained = true; return nil }
func (f *fakeTty) NotifyResize(ch chan<- bool) {
	f.resizeCh = ch
}
func (f *fakeTty) WindowSize() (tcell.WindowSize, error) {
	return tcell.WindowSize{Width: 80, Height: 24}, nil
}
func (f *fakeTty) Read(p []byte) (int, error) {
	if len(f.readData) == 0 {
		return 0, nil
	}
	n := copy(p, f.readData)
	f.readData = f.readData[n:]
	return n, nil
}
func (f *fakeTty) Write(p []byte) (int, error) {
	clone := append([]byte(nil), p...)
	f.wrote = append(f.wrote, clone)
	return len(p), nil
}
func (f *fakeTty) Close() error { f.closed = true; return nil }

// testScreen is a minimal tcell.Screen implementation for testing
type testScreen struct {
	mu       sync.Mutex
	width    int
	height   int
	cells    map[int]map[int]string
	eventQ   chan tcell.Event
	cursorX  int
	cursorY  int
}

func newTestScreen(w, h int) *testScreen {
	return &testScreen{
		width:   w,
		height:  h,
		cells:   make(map[int]map[int]string),
		eventQ:  make(chan tcell.Event, 128),
		cursorX: -1,
		cursorY: -1,
	}
}

func (s *testScreen) Init() error                                               { return nil }
func (s *testScreen) Fini()                                                     { close(s.eventQ) }
func (s *testScreen) Clear()                                                    {}
func (s *testScreen) Fill(r rune, style tcell.Style)                            {}
func (s *testScreen) SetStyle(style tcell.Style)                                {}
func (s *testScreen) ShowCursor(x, y int)                                       { s.cursorX = x; s.cursorY = y }
func (s *testScreen) HideCursor()                                               { s.cursorX = -1; s.cursorY = -1 }
func (s *testScreen) SetCursorStyle(cs tcell.CursorStyle, c ...tcellcolor.Color) {}
func (s *testScreen) Show()                                                     {}
func (s *testScreen) Sync()                                                     {}
func (s *testScreen) CharacterSet() string                                      { return "UTF-8" }
func (s *testScreen) RegisterRuneFallback(r rune, subst string)                 {}
func (s *testScreen) UnregisterRuneFallback(r rune)                             {}
func (s *testScreen) Resize(int, int, int, int)                                 {}
func (s *testScreen) Suspend() error                                            { return nil }
func (s *testScreen) Resume() error                                             { return nil }
func (s *testScreen) Beep() error                                               { return nil }
func (s *testScreen) LockRegion(x, y, w, h int, lock bool)                      {}
func (s *testScreen) Tty() (tcell.Tty, bool)                                    { return nil, false }
func (s *testScreen) SetTitle(string)                                           {}
func (s *testScreen) SetClipboard([]byte)                                       {}
func (s *testScreen) GetClipboard()                                             {}
func (s *testScreen) HasClipboard() bool                                        { return false }
func (s *testScreen) ShowNotification(title, body string)                       {}
func (s *testScreen) Terminal() (string, string)                                { return "test", "1.0" }
func (s *testScreen) EnableMouse(...tcell.MouseFlags)                           {}
func (s *testScreen) DisableMouse()                                             {}
func (s *testScreen) EnablePaste()                                              {}
func (s *testScreen) DisablePaste()                                             {}
func (s *testScreen) EnableFocus()                                              {}
func (s *testScreen) DisableFocus()                                             {}
func (s *testScreen) Colors() int                                               { return 256 }

func (s *testScreen) Size() (int, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.width, s.height
}

func (s *testScreen) SetSize(w, h int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.width = w
	s.height = h
}

func (s *testScreen) EventQ() chan tcell.Event {
	return s.eventQ
}

func (s *testScreen) SetContent(x, y int, primary rune, combining []rune, style tcell.Style) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cells[y] == nil {
		s.cells[y] = make(map[int]string)
	}
	s.cells[y][x] = string(primary)
}

func (s *testScreen) Get(x, y int) (string, tcell.Style, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if row, ok := s.cells[y]; ok {
		if str, ok := row[x]; ok {
			return str, tcell.StyleDefault, 1
		}
	}
	return " ", tcell.StyleDefault, 1
}

func (s *testScreen) Put(x, y int, str string, style tcell.Style) (string, int) {
	if len(str) == 0 {
		return "", 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cells[y] == nil {
		s.cells[y] = make(map[int]string)
	}
	runes := []rune(str)
	s.cells[y][x] = string(runes[0])
	if len(runes) > 1 {
		return string(runes[1:]), 1
	}
	return "", 1
}

func (s *testScreen) PutStr(x, y int, str string) {
	s.PutStrStyled(x, y, str, tcell.StyleDefault)
}

func (s *testScreen) PutStrStyled(x, y int, str string, style tcell.Style) {
	for _, r := range str {
		s.SetContent(x, y, r, nil, style)
		x++
	}
}

func TestBackendSimulationBasics(t *testing.T) {
	screen := newTestScreen(4, 2)
	be := NewWithScreen(screen)

	if err := be.Init(); err != nil {
		t.Fatalf("Init error: %v", err)
	}
	defer be.Fini()

	w, h := be.Size()
	if w != 4 || h != 2 {
		t.Fatalf("Size = %d,%d", w, h)
	}

	style := backend.DefaultStyle().Foreground(backend.ColorRed).Bold(true)
	be.SetContent(1, 0, 'A', nil, style)
	str, _, _ := screen.Get(1, 0)
	if str != "A" {
		t.Fatalf("expected SetContent to set rune, got %q", str)
	}

	be.SetRow(1, 0, []backend.Cell{{Rune: 'B', Style: backend.DefaultStyle()}, {Rune: 'C', Style: backend.DefaultStyle()}})
	str, _, _ = screen.Get(0, 1)
	if str != "B" {
		t.Fatalf("expected SetRow to set first cell, got %q", str)
	}

	be.SetRect(2, 0, 2, 1, []backend.Cell{{Rune: 'X', Style: backend.DefaultStyle()}, {Rune: 'Y', Style: backend.DefaultStyle()}})
	str, _, _ = screen.Get(2, 0)
	if str != "X" {
		t.Fatalf("expected SetRect to set cell, got %q", str)
	}

	be.Show()
	be.Clear()
	be.HideCursor()
	be.ShowCursor()
	be.SetCursorPos(0, 0)
	be.Sync()
	be.Beep()
}

func TestBackendPostAndPollEvent(t *testing.T) {
	screen := newTestScreen(2, 1)
	be := NewWithScreen(screen)

	if err := be.Init(); err != nil {
		t.Fatalf("Init error: %v", err)
	}
	defer be.Fini()

	if err := be.PostEvent(terminal.KeyEvent{Key: terminal.KeyRune, Rune: 'x'}); err != nil {
		t.Fatalf("PostEvent error: %v", err)
	}
	ev := be.PollEvent()
	keyEvent, ok := ev.(terminal.KeyEvent)
	if !ok || keyEvent.Rune != 'x' {
		t.Fatalf("expected key event")
	}
}

func TestRawTTYWrapperAndImageWrites(t *testing.T) {
	fake := &fakeTty{readData: []byte("z")}
	raw := &rawTty{inner: fake}

	_ = raw.Start()
	_ = raw.Drain()
	ch := make(chan bool, 1)
	raw.NotifyResize(ch)
	_, _ = raw.WindowSize()
	buf := make([]byte, 2)
	_, _ = raw.Read(buf)
	_, _ = raw.Write([]byte("hi"))
	_ = raw.WriteRaw([]byte("raw"))
	_ = raw.Stop()
	_ = raw.Close()

	if !fake.started || !fake.stopped || !fake.drained || !fake.closed {
		t.Fatalf("expected raw tty forwarding")
	}
	if len(fake.wrote) == 0 {
		t.Fatalf("expected writes")
	}

	// Create a backend with our fake raw tty for image testing
	tcellBe := &Backend{raw: raw}

	img := backend.Image{
		Width:      1,
		Height:     1,
		CellWidth:  1,
		CellHeight: 1,
		Format:     backend.ImageFormatRGBA,
		Protocol:   backend.ImageProtocolKitty,
		Pixels:     []byte{0xff, 0, 0, 0xff},
	}
	tcellBe.DrawImage(0, 0, img)

	var combined bytes.Buffer
	for _, b := range fake.wrote {
		combined.Write(b)
	}
	out := combined.String()
	if !strings.Contains(out, "\x1b_G") {
		t.Fatalf("expected kitty payload")
	}
	if !strings.Contains(out, "\x1b[s") || !strings.Contains(out, "\x1b[u") {
		t.Fatalf("expected save/restore cursor")
	}

	tcellBe.WriteFrame([]byte("frame"))
	combined.Reset()
	for _, b := range fake.wrote {
		combined.Write(b)
	}
	if !strings.Contains(combined.String(), "frame") {
		t.Fatalf("expected frame write")
	}
}

func TestCursorTo(t *testing.T) {
	if cursorTo(0, 0) != "\x1b[1;1H" {
		t.Fatalf("unexpected cursor sequence")
	}
}

var _ tcell.Tty = (*fakeTty)(nil)
var _ tcell.Screen = (*testScreen)(nil)
