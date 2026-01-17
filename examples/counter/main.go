package main

import (
	"context"
	"fmt"
	"os"

	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/examples/internal/demo"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/state"
	"github.com/odvcencio/fluffy-ui/widgets"
)

func main() {
	count := state.NewSignal(0)
	count.SetEqualFunc(state.EqualComparable[int])

	view := NewCounterView(count)
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

type CounterView struct {
	widgets.Component
	count      *state.Signal[int]
	title      *widgets.Label
	countLabel *widgets.Label
	grid       *widgets.Grid
	incBtn     *widgets.Button
	decBtn     *widgets.Button
	resetBtn   *widgets.Button
}

func NewCounterView(count *state.Signal[int]) *CounterView {
	view := &CounterView{count: count}
	view.title = widgets.NewLabel("FluffyUI Counter").WithStyle(backend.DefaultStyle().Bold(true))
	view.countLabel = widgets.NewLabel("Count: 0")
	view.incBtn = widgets.NewButton("Increment", widgets.WithVariant(widgets.VariantPrimary), widgets.WithOnClick(func() {
		view.updateCount(1)
	}))
	view.decBtn = widgets.NewButton("Decrement", widgets.WithVariant(widgets.VariantSecondary), widgets.WithOnClick(func() {
		view.updateCount(-1)
	}))
	view.resetBtn = widgets.NewButton("Reset", widgets.WithVariant(widgets.VariantDanger), widgets.WithOnClick(func() {
		view.reset()
	}))

	grid := widgets.NewGrid(4, 2)
	grid.Gap = 1
	grid.Add(view.title, 0, 0, 1, 2)
	grid.Add(view.countLabel, 1, 0, 1, 2)
	grid.Add(view.decBtn, 2, 0, 1, 1)
	grid.Add(view.incBtn, 2, 1, 1, 1)
	grid.Add(view.resetBtn, 3, 0, 1, 2)
	view.grid = grid

	view.refresh()
	return view
}

func (c *CounterView) Mount() {
	c.Observe(c.count, c.refresh)
	c.refresh()
}

func (c *CounterView) Unmount() {
	c.Subs.Clear()
}

func (c *CounterView) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (c *CounterView) Layout(bounds runtime.Rect) {
	c.Component.Layout(bounds)
	if c.grid != nil {
		c.grid.Layout(bounds)
	}
}

func (c *CounterView) Render(ctx runtime.RenderContext) {
	if c.grid != nil {
		c.grid.Render(ctx)
	}
}

func (c *CounterView) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if c.grid != nil {
		if result := c.grid.HandleMessage(msg); result.Handled {
			return result
		}
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	switch key.Rune {
	case '+', '=':
		c.updateCount(1)
		return runtime.Handled()
	case '-', '_':
		c.updateCount(-1)
		return runtime.Handled()
	case 'r', 'R':
		c.reset()
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func (c *CounterView) ChildWidgets() []runtime.Widget {
	if c.grid == nil {
		return nil
	}
	return []runtime.Widget{c.grid}
}

func (c *CounterView) refresh() {
	if c.countLabel == nil || c.count == nil {
		return
	}
	c.countLabel.SetText(fmt.Sprintf("Count: %d", c.count.Get()))
}

func (c *CounterView) updateCount(delta int) {
	if c.count == nil {
		return
	}
	c.count.Update(func(v int) int { return v + delta })
	c.refresh()
	c.Invalidate()
}

func (c *CounterView) reset() {
	if c.count == nil {
		return
	}
	c.count.Set(0)
	c.refresh()
	c.Invalidate()
}
