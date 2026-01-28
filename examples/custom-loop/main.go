package main

import (
	"fmt"
	"os"
	"time"

	"github.com/odvcencio/fluffyui/backend"
	backendtcell "github.com/odvcencio/fluffyui/backend/tcell"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
	"github.com/odvcencio/fluffyui/widgets"
)

func main() {
	be, err := backendtcell.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "backend error:", err)
		os.Exit(1)
	}
	if err := be.Init(); err != nil {
		fmt.Fprintln(os.Stderr, "init error:", err)
		os.Exit(1)
	}
	defer be.Fini()

	be.HideCursor()
	w, h := be.Size()
	screen := runtime.NewScreen(w, h)
	screen.SetAutoRegisterFocus(true)

	root := NewCounterView()
	screen.SetRoot(root)

	messages := make(chan runtime.Message, 128)
	quit := make(chan struct{})
	go pollEvents(be, messages, quit)

	ticker := time.NewTicker(time.Second / 30)
	defer ticker.Stop()

	running := true
	for running {
		select {
		case msg := <-messages:
			if msg == nil {
				continue
			}
			handleRuntimeMessage(screen, msg)
			result := screen.HandleMessage(msg)
			running = !handleCommands(result.Commands, screen, messages)
		case now := <-ticker.C:
			result := screen.HandleMessage(runtime.TickMsg{Time: now})
			running = !handleCommands(result.Commands, screen, messages)
		}
		render(screen, be)
	}

	close(quit)
}

type CounterView struct {
	widgets.Base
	count    int
	lastTick time.Time
}

func NewCounterView() *CounterView {
	view := &CounterView{}
	view.Base.Label = "Custom Loop"
	return view
}

func (c *CounterView) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (c *CounterView) Layout(bounds runtime.Rect) {
	c.Base.Layout(bounds)
}

func (c *CounterView) Render(ctx runtime.RenderContext) {
	if ctx.Buffer == nil {
		return
	}
	bounds := c.Bounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	style := backend.DefaultStyle()
	ctx.Buffer.Fill(bounds, ' ', style)

	tickLine := "Last tick: --"
	if !c.lastTick.IsZero() {
		tickLine = "Last tick: " + c.lastTick.Format("15:04:05.000")
	}
	lines := []string{
		"Custom Loop Example",
		"",
		fmt.Sprintf("Count: %d", c.count),
		tickLine,
		"",
		"Keys: +/- to change, q or Ctrl+C to quit",
	}
	for i, line := range lines {
		if i >= bounds.Height {
			break
		}
		if len(line) > bounds.Width {
			line = line[:bounds.Width]
		}
		ctx.Buffer.SetString(bounds.X, bounds.Y+i, line, style)
	}
}

func (c *CounterView) HandleMessage(msg runtime.Message) runtime.HandleResult {
	switch m := msg.(type) {
	case runtime.KeyMsg:
		switch {
		case m.Key == terminal.KeyCtrlC || m.Rune == 'q':
			return runtime.WithCommand(runtime.Quit{})
		case m.Rune == '+' || m.Rune == '=':
			c.count++
			c.Invalidate()
			return runtime.Handled()
		case m.Rune == '-':
			c.count--
			c.Invalidate()
			return runtime.Handled()
		}
	case runtime.TickMsg:
		c.lastTick = m.Time
		c.Invalidate()
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func pollEvents(be backend.Backend, out chan<- runtime.Message, quit <-chan struct{}) {
	for {
		select {
		case <-quit:
			return
		default:
		}

		ev := be.PollEvent()
		if ev == nil {
			continue
		}
		switch e := ev.(type) {
		case terminal.KeyEvent:
			out <- runtime.KeyMsg{Key: e.Key, Rune: e.Rune, Alt: e.Alt, Ctrl: e.Ctrl, Shift: e.Shift}
		case terminal.ResizeEvent:
			out <- runtime.ResizeMsg{Width: e.Width, Height: e.Height}
		case terminal.MouseEvent:
			out <- runtime.MouseMsg{
				X:      e.X,
				Y:      e.Y,
				Button: runtime.MouseButton(e.Button),
				Action: runtime.MouseAction(e.Action),
				Alt:    e.Alt,
				Ctrl:   e.Ctrl,
				Shift:  e.Shift,
			}
		case terminal.PasteEvent:
			out <- runtime.PasteMsg{Text: e.Text}
		}
	}
}

func handleRuntimeMessage(screen *runtime.Screen, msg runtime.Message) {
	switch m := msg.(type) {
	case runtime.ResizeMsg:
		screen.Resize(m.Width, m.Height)
	}
}

func handleCommands(cmds []runtime.Command, screen *runtime.Screen, out chan<- runtime.Message) bool {
	for _, cmd := range cmds {
		switch c := cmd.(type) {
		case runtime.Quit:
			return true
		case runtime.Refresh:
			if screen != nil {
				screen.Buffer().MarkAllDirty()
			}
		case runtime.SendMsg:
			if c.Message != nil {
				out <- c.Message
			}
		}
	}
	return false
}

func render(screen *runtime.Screen, be backend.Backend) {
	screen.Render()
	buf := screen.Buffer()
	if buf == nil {
		return
	}
	w, h := buf.Size()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			cell := buf.Get(x, y)
			r := cell.Rune
			if r == 0 {
				r = ' '
			}
			be.SetContent(x, y, r, nil, cell.Style)
		}
	}
	be.Show()
}

var _ runtime.Widget = (*CounterView)(nil)
