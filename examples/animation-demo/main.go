package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/odvcencio/fluffyui/animation"
	"github.com/odvcencio/fluffyui/backend"
	backendtcell "github.com/odvcencio/fluffyui/backend/tcell"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/widgets"
)

func main() {
	be, err := backendtcell.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "backend init failed: %v\n", err)
		os.Exit(1)
	}

	app := runtime.NewApp(runtime.AppConfig{
		Backend:  be,
		TickRate: time.Second / 60,
		Animator: animation.NewAnimator(),
	})

	root := NewAnimatedLabel("FluffyUI Animation")
	app.SetRoot(root)

	if err := app.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
		os.Exit(1)
	}
}

type AnimatedLabel struct {
	widgets.AnimatedWidget
	text string
}

func NewAnimatedLabel(text string) *AnimatedLabel {
	return &AnimatedLabel{
		AnimatedWidget: widgets.NewAnimatedWidget(),
		text:           text,
	}
}

func (l *AnimatedLabel) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (l *AnimatedLabel) Layout(bounds runtime.Rect) {
	l.AnimatedWidget.Layout(bounds)
}

func (l *AnimatedLabel) Mount() {
	l.FadeIn(500 * time.Millisecond)
	l.SlideIn(widgets.DirectionDown, 3, 500*time.Millisecond)
}

func (l *AnimatedLabel) Render(ctx runtime.RenderContext) {
	bounds := l.Bounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	opacity := float64(l.Opacity)
	if opacity <= 0.01 {
		return
	}
	style := backend.DefaultStyle()
	if opacity < 0.5 {
		style = style.Dim(true)
	}
	x := bounds.X + int(l.OffsetX)
	y := bounds.Y + bounds.Height/2 + int(l.OffsetY)
	ctx.Buffer.SetString(x, y, l.text, style)
}

func (l *AnimatedLabel) HandleMessage(msg runtime.Message) runtime.HandleResult {
	switch m := msg.(type) {
	case runtime.KeyMsg:
		switch m.Rune {
		case 'q':
			return runtime.WithCommand(runtime.Quit{})
		case 'f':
			l.FadeOut(300*time.Millisecond, func() {
				l.FadeIn(300 * time.Millisecond)
			})
			return runtime.Handled()
		case 's':
			l.SlideIn(widgets.DirectionUp, 3, 400*time.Millisecond)
			return runtime.Handled()
		}
	}
	return runtime.Unhandled()
}
