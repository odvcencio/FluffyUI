package runtime

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/backend/sim"
	"github.com/odvcencio/fluffyui/terminal"
)

type testCommand struct{}

func (testCommand) Command() {}

type appTestWidget struct {
	keyCommands map[rune]Command
	renderChar  rune
	boundsCh    chan Rect
}

type inlineAwareBackend struct {
	*sim.Backend
	inline       bool
	inlineHeight int
}

func (b *inlineAwareBackend) SetInlineMode(enabled bool) {
	if b == nil {
		return
	}
	b.inline = enabled
}

func (b *inlineAwareBackend) SetInlineHeight(lines int) {
	if b == nil {
		return
	}
	b.inlineHeight = lines
}

func (w *appTestWidget) Measure(c Constraints) Size {
	return c.MaxSize()
}

func (w *appTestWidget) Layout(bounds Rect) {
	if w.boundsCh == nil {
		return
	}
	select {
	case w.boundsCh <- bounds:
	default:
	}
}

func (w *appTestWidget) Render(ctx RenderContext) {
	if w.renderChar == 0 || ctx.Buffer == nil {
		return
	}
	ctx.Buffer.Set(ctx.Bounds.X, ctx.Bounds.Y, w.renderChar, backend.DefaultStyle())
}

func (w *appTestWidget) HandleMessage(msg Message) HandleResult {
	key, ok := msg.(KeyMsg)
	if !ok {
		return Unhandled()
	}
	if cmd, ok := w.keyCommands[key.Rune]; ok {
		return WithCommand(cmd)
	}
	return Unhandled()
}

func TestApp_RunQuit(t *testing.T) {
	be := sim.New(5, 3)
	w := &appTestWidget{
		keyCommands: map[rune]Command{'q': Quit{}},
		renderChar:  'X',
	}

	app := NewApp(AppConfig{
		Backend: be,
		Root:    w,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx)
	}()

	waitForScreen(t, app)

	app.Post(KeyMsg{Key: terminal.KeyRune, Rune: 'q'})

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Run did not exit after Quit command")
	}
}

func TestApp_CommandHandler(t *testing.T) {
	be := sim.New(5, 3)
	w := &appTestWidget{
		keyCommands: map[rune]Command{'c': testCommand{}, 'q': Quit{}},
		renderChar:  'X',
	}

	handled := make(chan struct{}, 1)
	app := NewApp(AppConfig{
		Backend: be,
		Root:    w,
		CommandHandler: func(cmd Command) bool {
			if _, ok := cmd.(testCommand); ok {
				handled <- struct{}{}
				return true
			}
			return false
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx)
	}()

	waitForScreen(t, app)

	app.Post(KeyMsg{Key: terminal.KeyRune, Rune: 'c'})

	select {
	case <-handled:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("CommandHandler did not receive testCommand")
	}

	app.Post(KeyMsg{Key: terminal.KeyRune, Rune: 'q'})
	if err := <-done; err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
}

func TestApp_Resize(t *testing.T) {
	be := sim.New(5, 3)
	boundsCh := make(chan Rect, 4)
	w := &appTestWidget{
		keyCommands: map[rune]Command{'q': Quit{}},
		renderChar:  'X',
		boundsCh:    boundsCh,
	}

	app := NewApp(AppConfig{
		Backend: be,
		Root:    w,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx)
	}()

	waitForScreen(t, app)
	drainBounds(boundsCh)

	app.Post(ResizeMsg{Width: 12, Height: 7})
	waitForBounds(t, boundsCh, 12, 7)

	app.Post(KeyMsg{Key: terminal.KeyRune, Rune: 'q'})
	if err := <-done; err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
}

func TestApp_Call(t *testing.T) {
	be := sim.New(5, 3)
	w := &appTestWidget{
		keyCommands: map[rune]Command{'q': Quit{}},
		renderChar:  'X',
	}

	app := NewApp(AppConfig{
		Backend: be,
		Root:    w,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx)
	}()

	waitForScreen(t, app)

	expectedW, expectedH := be.Size()
	callCtx, callCancel := context.WithTimeout(context.Background(), time.Second)
	defer callCancel()

	if err := app.Call(callCtx, func(app *App) error {
		screen := app.Screen()
		if screen == nil {
			return fmt.Errorf("screen not initialized")
		}
		w, h := screen.Size()
		if w != expectedW || h != expectedH {
			return fmt.Errorf("screen size = %dx%d, want %dx%d", w, h, expectedW, expectedH)
		}
		return nil
	}); err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	app.Post(KeyMsg{Key: terminal.KeyRune, Rune: 'q'})
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Run did not exit after Quit")
	}
}

func TestApp_OnReady(t *testing.T) {
	be := sim.New(5, 3)
	w := &appTestWidget{
		keyCommands: map[rune]Command{'q': Quit{}},
	}

	readyCh := make(chan *Screen, 1)
	app := NewApp(AppConfig{
		Backend: be,
		Root:    w,
		OnReady: func(a *App) {
			readyCh <- a.Screen()
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx)
	}()

	select {
	case screen := <-readyCh:
		if screen == nil {
			t.Fatal("OnReady called with nil Screen")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("OnReady was not called")
	}

	app.Post(KeyMsg{Key: terminal.KeyRune, Rune: 'q'})
	if err := <-done; err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
}

func TestApp_OnResize(t *testing.T) {
	be := sim.New(5, 3)
	w := &appTestWidget{
		keyCommands: map[rune]Command{'q': Quit{}},
	}

	resizeCh := make(chan [2]int, 4)
	app := NewApp(AppConfig{
		Backend: be,
		Root:    w,
		OnResize: func(app *App, width, height int) {
			resizeCh <- [2]int{width, height}
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx)
	}()

	select {
	case size := <-resizeCh:
		if size[0] != 5 || size[1] != 3 {
			t.Fatalf("initial OnResize size = %dx%d, want 5x3", size[0], size[1])
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("initial OnResize was not called")
	}

	app.Post(ResizeMsg{Width: 12, Height: 7})
	deadline := time.After(500 * time.Millisecond)
	for {
		select {
		case size := <-resizeCh:
			if size[0] == 12 && size[1] == 7 {
				app.Post(KeyMsg{Key: terminal.KeyRune, Rune: 'q'})
				if err := <-done; err != nil {
					t.Fatalf("Run returned error: %v", err)
				}
				return
			}
		case <-deadline:
			t.Fatal("OnResize was not called with resized dimensions")
		}
	}
}

func TestApp_OnQuit(t *testing.T) {
	be := sim.New(5, 3)
	w := &appTestWidget{
		keyCommands: map[rune]Command{'q': Quit{}},
	}

	quitCh := make(chan struct{}, 1)
	app := NewApp(AppConfig{
		Backend: be,
		Root:    w,
		OnQuit: func(app *App) {
			quitCh <- struct{}{}
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx)
	}()

	waitForScreen(t, app)
	app.Post(KeyMsg{Key: terminal.KeyRune, Rune: 'q'})

	select {
	case <-quitCh:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("OnQuit was not called")
	}

	if err := <-done; err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
}

func TestApp_InlineModeSetter(t *testing.T) {
	w := &appTestWidget{
		keyCommands: map[rune]Command{'q': Quit{}},
	}
	be := &inlineAwareBackend{Backend: sim.New(5, 3)}

	app := NewApp(AppConfig{
		Backend:      be,
		Root:         w,
		InlineMode:   true,
		InlineHeight: 7,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx)
	}()

	waitForScreen(t, app)
	app.Post(KeyMsg{Key: terminal.KeyRune, Rune: 'q'})
	if err := <-done; err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !be.inline {
		t.Fatal("backend inline mode was not enabled")
	}
	if be.inlineHeight != 7 {
		t.Fatalf("backend inline height = %d, want 7", be.inlineHeight)
	}
}

func waitForScreen(t *testing.T, app *App) {
	t.Helper()

	deadline := time.After(500 * time.Millisecond)
	for {
		if app.Screen() != nil {
			return
		}
		select {
		case <-deadline:
			t.Fatal("screen did not initialize in time")
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}

func drainBounds(ch <-chan Rect) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}

func waitForBounds(t *testing.T, ch <-chan Rect, width, height int) {
	t.Helper()

	deadline := time.After(500 * time.Millisecond)
	for {
		select {
		case bounds := <-ch:
			if bounds.Width == width && bounds.Height == height {
				return
			}
		case <-deadline:
			t.Fatalf("layout with %dx%d not observed", width, height)
		}
	}
}
