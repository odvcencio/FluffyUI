package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/examples/internal/demo"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/widgets"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: video-player <path-to-video>")
		os.Exit(1)
	}
	path := os.Args[1]
	view, err := NewVideoPlayerView(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "video init failed: %v\n", err)
		os.Exit(1)
	}

	bundle, err := demo.NewApp(view, demo.Options{TickRate: time.Second / 30})
	if err != nil {
		fmt.Fprintf(os.Stderr, "app init failed: %v\n", err)
		os.Exit(1)
	}

	if err := bundle.App.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
		os.Exit(1)
	}
}

type VideoPlayerView struct {
	widgets.Component
	player *widgets.VideoPlayer
	title  *widgets.Label
	panel  *widgets.Panel
	help   *widgets.Label
}

func NewVideoPlayerView(path string) (*VideoPlayerView, error) {
	player, err := widgets.NewVideoPlayer(path)
	if err != nil {
		return nil, err
	}
	player.Play()

	view := &VideoPlayerView{player: player}
	view.title = widgets.NewLabel("Video Player").WithStyle(backend.DefaultStyle().Bold(true))
	view.panel = widgets.NewPanel(player).WithBorder(backend.DefaultStyle())
	view.panel.SetTitle(filepath.Base(path))
	view.help = widgets.NewLabel("Space: play/pause  Ctrl+C: quit")
	return view, nil
}

func (v *VideoPlayerView) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (v *VideoPlayerView) Layout(bounds runtime.Rect) {
	v.Component.Layout(bounds)
	y := bounds.Y
	if v.title != nil {
		v.title.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: 1})
		y += 2
	}
	helpHeight := 1
	remaining := bounds.Height - (y - bounds.Y) - helpHeight
	if remaining < 1 {
		remaining = 1
	}
	if v.panel != nil {
		v.panel.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: remaining})
		y += remaining
	}
	if v.help != nil && y < bounds.Y+bounds.Height {
		v.help.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: helpHeight})
	}
}

func (v *VideoPlayerView) Render(ctx runtime.RenderContext) {
	if v.title != nil {
		v.title.Render(ctx)
	}
	if v.panel != nil {
		v.panel.Render(ctx)
	}
	if v.help != nil {
		v.help.Render(ctx)
	}
}

func (v *VideoPlayerView) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if v.panel != nil {
		if result := v.panel.HandleMessage(msg); result.Handled {
			return result
		}
	}
	return runtime.Unhandled()
}

func (v *VideoPlayerView) ChildWidgets() []runtime.Widget {
	children := []runtime.Widget{}
	if v.title != nil {
		children = append(children, v.title)
	}
	if v.panel != nil {
		children = append(children, v.panel)
	}
	if v.help != nil {
		children = append(children, v.help)
	}
	return children
}
