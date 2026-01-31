package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/examples/internal/demo"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	"github.com/odvcencio/fluffyui/widgets"
)

func main() {
	view := NewTodoView()
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

type Task struct {
	Title string
	Done  bool
}

type TodoView struct {
	widgets.Component
	tasks       *state.Signal[[]Task]
	title       *widgets.Label
	input       *widgets.Input
	list        *widgets.List[Task]
	status      *widgets.Label
	listStyle   backend.Style
	selectStyle backend.Style
}

func NewTodoView() *TodoView {
	view := &TodoView{
		tasks:       state.NewSignal([]Task{}),
		listStyle:   backend.DefaultStyle(),
		selectStyle: backend.DefaultStyle().Reverse(true),
	}
	view.title = widgets.NewLabel("Todo App", widgets.WithLabelStyle(backend.DefaultStyle().Bold(true)))
	view.input = widgets.NewInput()
	view.input.SetPlaceholder("Add a task and press Enter")
	view.input.SetOnSubmit(func(text string) {
		view.addTask(text)
	})

	adapter := widgets.NewSignalAdapter(view.tasks, func(item Task, index int, selected bool, ctx runtime.RenderContext) {
		style := view.listStyle
		if item.Done {
			style = style.Dim(true)
		}
		if selected {
			style = view.selectStyle
		}
		marker := "[ ]"
		if item.Done {
			marker = "[x]"
		}
		line := marker + " " + item.Title
		line = truncateAndPad(line, ctx.Bounds.Width)
		ctx.Buffer.SetString(ctx.Bounds.X, ctx.Bounds.Y, line, style)
	})
	view.list = widgets.NewList(adapter)
	view.list.SetOnSelect(func(index int, item Task) {
		view.toggleTask(index)
	})
	view.status = widgets.NewLabel("0 items")
	view.syncStatus()
	return view
}

func (t *TodoView) Mount() {
	t.Observe(t.tasks, func() {
		t.syncStatus()
		t.Invalidate()
	})
	if t.input != nil {
		t.input.SetOnChange(func(text string) {
			t.Invalidate()
		})
	}
}

func (t *TodoView) Unmount() {
	t.Subs.Clear()
}

func (t *TodoView) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (t *TodoView) Layout(bounds runtime.Rect) {
	t.Component.Layout(bounds)
	y := bounds.Y

	if t.title != nil {
		t.title.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: 1})
		y++
	}
	if t.input != nil {
		t.input.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: 1})
		y++
	}

	statusHeight := 1
	listHeight := bounds.Height - (y - bounds.Y) - statusHeight
	if listHeight < 0 {
		listHeight = 0
	}
	if t.list != nil {
		t.list.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: listHeight})
	}
	if t.status != nil {
		t.status.Layout(runtime.Rect{X: bounds.X, Y: y + listHeight, Width: bounds.Width, Height: statusHeight})
	}
}

func (t *TodoView) Render(ctx runtime.RenderContext) {
	if t.title != nil {
		t.title.Render(ctx)
	}
	if t.input != nil {
		t.input.Render(ctx)
	}
	if t.list != nil {
		t.list.Render(ctx)
	}
	if t.status != nil {
		t.status.Render(ctx)
	}
}

func (t *TodoView) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if t.input != nil {
		if result := t.input.HandleMessage(msg); result.Handled {
			return result
		}
	}
	if t.list != nil {
		if result := t.list.HandleMessage(msg); result.Handled {
			return result
		}
	}
	return runtime.Unhandled()
}

func (t *TodoView) ChildWidgets() []runtime.Widget {
	children := []runtime.Widget{}
	if t.title != nil {
		children = append(children, t.title)
	}
	if t.input != nil {
		children = append(children, t.input)
	}
	if t.list != nil {
		children = append(children, t.list)
	}
	if t.status != nil {
		children = append(children, t.status)
	}
	return children
}

func (t *TodoView) addTask(text string) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return
	}
	t.tasks.Update(func(items []Task) []Task {
		return append(items, Task{Title: trimmed})
	})
	if t.input != nil {
		t.input.Clear()
	}
	t.Invalidate()
}

func (t *TodoView) toggleTask(index int) {
	if index < 0 {
		return
	}
	t.tasks.Update(func(items []Task) []Task {
		if index >= len(items) {
			return items
		}
		items[index].Done = !items[index].Done
		return items
	})
	t.Invalidate()
}

func (t *TodoView) syncStatus() {
	if t.status == nil || t.tasks == nil {
		return
	}
	count := len(t.tasks.Get())
	t.status.SetText(fmt.Sprintf("%d items", count))
}

func truncateAndPad(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if len(text) > width {
		if width <= 3 {
			return text[:width]
		}
		text = text[:width-3] + "..."
	}
	if len(text) < width {
		pad := make([]byte, width-len(text))
		for i := range pad {
			pad[i] = ' '
		}
		text += string(pad)
	}
	return text
}
