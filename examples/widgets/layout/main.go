package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/examples/internal/demo"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/widgets"
)

func main() {
	view := NewLayoutView()
	bundle, err := demo.NewApp(view, demo.Options{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "app init failed: %v\n", err)
		os.Exit(1)
	}

	if err := bundle.App.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
		os.Exit(1)
	}
}

type LayoutView struct {
	widgets.Component
	header   *widgets.Label
	splitter *widgets.Splitter
	grid     *widgets.Grid
}

func NewLayoutView() *LayoutView {
	view := &LayoutView{}
	view.header = widgets.NewLabel("Layout Widgets").WithStyle(backend.DefaultStyle().Bold(true))

	longText := strings.Repeat("ScrollView keeps large content responsive. ", 8)
	text := widgets.NewText(longText)
	scroll := widgets.NewScrollView(text)
	leftPanel := widgets.NewPanel(scroll).WithBorder(backend.DefaultStyle())
	leftPanel.SetTitle("ScrollView")

	stack := widgets.NewStack(
		widgets.NewBox(nil).WithStyle(backend.DefaultStyle().Background(backend.ColorBlue)),
		widgets.NewLabel("Stack overlay (top-left)"),
	)
	rightPanel := widgets.NewPanel(stack).WithBorder(backend.DefaultStyle())
	rightPanel.SetTitle("Stack")

	view.splitter = widgets.NewSplitter(leftPanel, rightPanel)
	view.splitter.Ratio = 0.6

	view.grid = widgets.NewGrid(2, 2)
	view.grid.Gap = 1
	view.grid.Add(widgets.NewPanel(widgets.NewLabel("Grid cell A")).WithBorder(backend.DefaultStyle()), 0, 0, 1, 1)
	view.grid.Add(widgets.NewPanel(widgets.NewLabel("Grid cell B")).WithBorder(backend.DefaultStyle()), 0, 1, 1, 1)
	view.grid.Add(widgets.NewPanel(widgets.NewLabel("Grid cell C")).WithBorder(backend.DefaultStyle()), 1, 0, 1, 1)
	view.grid.Add(widgets.NewPanel(widgets.NewLabel("Grid cell D")).WithBorder(backend.DefaultStyle()), 1, 1, 1, 1)

	return view
}

func (l *LayoutView) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (l *LayoutView) Layout(bounds runtime.Rect) {
	l.Component.Layout(bounds)
	y := bounds.Y
	if l.header != nil {
		l.header.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: 1})
		y++
	}

	gridHeight := 6
	splitHeight := bounds.Height - (y - bounds.Y) - gridHeight
	if splitHeight < 0 {
		splitHeight = 0
	}
	if l.splitter != nil {
		l.splitter.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: splitHeight})
		y += splitHeight
	}
	if l.grid != nil {
		l.grid.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: gridHeight})
	}
}

func (l *LayoutView) Render(ctx runtime.RenderContext) {
	if l.header != nil {
		l.header.Render(ctx)
	}
	if l.splitter != nil {
		l.splitter.Render(ctx)
	}
	if l.grid != nil {
		l.grid.Render(ctx)
	}
}

func (l *LayoutView) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if l.splitter != nil {
		if result := l.splitter.HandleMessage(msg); result.Handled {
			return result
		}
	}
	if l.grid != nil {
		if result := l.grid.HandleMessage(msg); result.Handled {
			return result
		}
	}
	return runtime.Unhandled()
}

func (l *LayoutView) ChildWidgets() []runtime.Widget {
	children := []runtime.Widget{}
	if l.header != nil {
		children = append(children, l.header)
	}
	if l.splitter != nil {
		children = append(children, l.splitter)
	}
	if l.grid != nil {
		children = append(children, l.grid)
	}
	return children
}
