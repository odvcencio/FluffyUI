package runtime

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/backend/sim"
	"github.com/odvcencio/fluffy-ui/terminal"
)

type testCommand struct{}

func (testCommand) Command() {}

type appTestWidget struct {
	keyCommands map[rune]Command
	renderChar  rune
	boundsCh    chan Rect
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
