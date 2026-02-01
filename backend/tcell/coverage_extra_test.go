package tcell

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/terminal"
)

type fakeTty struct {
	started  bool
	stopped  bool
	drained  bool
	closed   bool
	resized  bool
	wrote    [][]byte
	readData []byte
	cb       func()
}

func (f *fakeTty) Start() error { f.started = true; return nil }
func (f *fakeTty) Stop() error  { f.stopped = true; return nil }
func (f *fakeTty) Drain() error { f.drained = true; return nil }
func (f *fakeTty) NotifyResize(cb func()) {
	f.resized = true
	f.cb = cb
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

func TestBackendSimulationBasics(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	be := NewWithScreen(screen)

	if err := be.Init(); err != nil {
		t.Fatalf("Init error: %v", err)
	}
	defer be.Fini()

	screen.SetSize(4, 2)
	w, h := be.Size()
	if w != 4 || h != 2 {
		t.Fatalf("Size = %d,%d", w, h)
	}

	style := backend.DefaultStyle().Foreground(backend.ColorRed).Bold(true)
	be.SetContent(1, 0, 'A', nil, style)
	str, _, _ := screen.Get(1, 0)
	if str == "" || []rune(str)[0] != 'A' {
		t.Fatalf("expected SetContent to set rune")
	}

	be.SetRow(1, 0, []backend.Cell{{Rune: 'B', Style: backend.DefaultStyle()}, {Rune: 'C', Style: backend.DefaultStyle()}})
	str, _, _ = screen.Get(0, 1)
	if []rune(str)[0] != 'B' {
		t.Fatalf("expected SetRow to set first cell")
	}

	be.SetRect(2, 0, 2, 1, []backend.Cell{{Rune: 'X', Style: backend.DefaultStyle()}, {Rune: 'Y', Style: backend.DefaultStyle()}})
	str, _, _ = screen.Get(2, 0)
	if []rune(str)[0] != 'X' {
		t.Fatalf("expected SetRect to set cell")
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
	screen := tcell.NewSimulationScreen("")
	screen.SetSize(2, 1)
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

func TestBackendPasteEvent(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	screen.SetSize(2, 1)
	be := NewWithScreen(screen)

	if err := be.Init(); err != nil {
		t.Fatalf("Init error: %v", err)
	}
	defer be.Fini()

	screen.PostEvent(tcell.NewEventPaste(true))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	screen.PostEvent(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	screen.PostEvent(tcell.NewEventPaste(false))

	ev := be.PollEvent()
	paste, ok := ev.(terminal.PasteEvent)
	if !ok {
		t.Fatalf("expected paste event")
	}
	if paste.Text != "a\n" {
		t.Fatalf("unexpected paste text: %q", paste.Text)
	}
}

func TestRawTTYWrapperAndImageWrites(t *testing.T) {
	fake := &fakeTty{readData: []byte("z")}
	raw := &rawTty{inner: fake}

	_ = raw.Start()
	_ = raw.Drain()
	raw.NotifyResize(func() {})
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

	screen := tcell.NewSimulationScreen("")
	screen.SetSize(2, 1)
	be := NewWithScreen(screen)
	be.raw = raw

	img := backend.Image{
		Width:      1,
		Height:     1,
		CellWidth:  1,
		CellHeight: 1,
		Format:     backend.ImageFormatRGBA,
		Protocol:   backend.ImageProtocolKitty,
		Pixels:     []byte{0xff, 0, 0, 0xff},
	}
	be.DrawImage(0, 0, img)

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

	be.WriteFrame([]byte("frame"))
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
