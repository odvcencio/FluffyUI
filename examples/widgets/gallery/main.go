package main

import (
	"context"
	"fmt"
	"os"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/examples/internal/demo"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	"github.com/odvcencio/fluffyui/widgets"
)

func main() {
	view := NewGalleryView()
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

type GalleryView struct {
	widgets.Component
	title    *widgets.Label
	splitter *widgets.Splitter
	left     *galleryLeft
	right    *galleryRight
}

func NewGalleryView() *GalleryView {
	view := &GalleryView{}
	view.title = widgets.NewLabel("Widget Gallery", widgets.WithLabelStyle(backend.DefaultStyle().Bold(true)))
	view.left = newGalleryLeft()
	view.right = newGalleryRight()
	view.splitter = widgets.NewSplitter(view.left, view.right)
	view.splitter.Ratio = 0.5
	return view
}

func (g *GalleryView) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (g *GalleryView) Layout(bounds runtime.Rect) {
	g.Component.Layout(bounds)
	y := bounds.Y
	if g.title != nil {
		g.title.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: 1})
		y++
	}
	if g.splitter != nil {
		g.splitter.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: bounds.Height - (y - bounds.Y)})
	}
}

func (g *GalleryView) Render(ctx runtime.RenderContext) {
	if g.title != nil {
		g.title.Render(ctx)
	}
	if g.splitter != nil {
		g.splitter.Render(ctx)
	}
}

func (g *GalleryView) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if g.splitter != nil {
		return g.splitter.HandleMessage(msg)
	}
	return runtime.Unhandled()
}

func (g *GalleryView) ChildWidgets() []runtime.Widget {
	children := []runtime.Widget{}
	if g.title != nil {
		children = append(children, g.title)
	}
	if g.splitter != nil {
		children = append(children, g.splitter)
	}
	return children
}

type galleryLeft struct {
	widgets.Base
	input     *widgets.Input
	selecter  *widgets.Select
	checkbox  *widgets.Checkbox
	radioA    *widgets.Radio
	radioB    *widgets.Radio
	primary   *widgets.Button
	secondary *widgets.Button
}

func newGalleryLeft() *galleryLeft {
	group := widgets.NewRadioGroup()
	left := &galleryLeft{
		input:     widgets.NewInput(),
		selecter:  widgets.NewSelect(widgets.SelectOption{Label: "Small"}, widgets.SelectOption{Label: "Medium"}, widgets.SelectOption{Label: "Large"}),
		checkbox:  widgets.NewCheckbox("Enable feature"),
		radioA:    widgets.NewRadio("Option A", group),
		radioB:    widgets.NewRadio("Option B", group),
		primary:   widgets.NewButton("Primary", widgets.WithVariant(widgets.VariantPrimary)),
		secondary: widgets.NewButton("Secondary", widgets.WithVariant(widgets.VariantSecondary)),
	}
	left.input.SetPlaceholder("Type here")
	return left
}

func (l *galleryLeft) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (l *galleryLeft) Layout(bounds runtime.Rect) {
	l.Base.Layout(bounds)
	y := bounds.Y
	line := func(w runtime.Widget) {
		if w == nil {
			return
		}
		w.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: 1})
		y++
	}
	line(l.input)
	line(l.selecter)
	line(l.checkbox)
	line(l.radioA)
	line(l.radioB)
	line(l.primary)
	line(l.secondary)
}

func (l *galleryLeft) Render(ctx runtime.RenderContext) {
	for _, child := range l.ChildWidgets() {
		if child != nil {
			child.Render(ctx)
		}
	}
}

func (l *galleryLeft) HandleMessage(msg runtime.Message) runtime.HandleResult {
	for _, child := range l.ChildWidgets() {
		if child == nil {
			continue
		}
		if result := child.HandleMessage(msg); result.Handled {
			return result
		}
	}
	return runtime.Unhandled()
}

func (l *galleryLeft) ChildWidgets() []runtime.Widget {
	children := []runtime.Widget{}
	if l.input != nil {
		children = append(children, l.input)
	}
	if l.selecter != nil {
		children = append(children, l.selecter)
	}
	if l.checkbox != nil {
		children = append(children, l.checkbox)
	}
	if l.radioA != nil {
		children = append(children, l.radioA)
	}
	if l.radioB != nil {
		children = append(children, l.radioB)
	}
	if l.primary != nil {
		children = append(children, l.primary)
	}
	if l.secondary != nil {
		children = append(children, l.secondary)
	}
	return children
}

type galleryRight struct {
	widgets.Base
	list     *widgets.List[string]
	table    *widgets.Table
	tree     *widgets.Tree
	progress *widgets.Progress
	spinner  *widgets.Spinner
	spark    *widgets.Sparkline
}

func newGalleryRight() *galleryRight {
	items := []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon"}
	adapter := widgets.NewSliceAdapter(items, func(item string, index int, selected bool, ctx runtime.RenderContext) {
		style := backend.DefaultStyle()
		if selected {
			style = style.Reverse(true)
		}
		line := truncateAndPad(item, ctx.Bounds.Width)
		ctx.Buffer.SetString(ctx.Bounds.X, ctx.Bounds.Y, line, style)
	})

	list := widgets.NewList(adapter)
	table := widgets.NewTable(
		widgets.TableColumn{Title: "Name"},
		widgets.TableColumn{Title: "Value"},
	)
	table.SetRows([][]string{{"One", "1"}, {"Two", "2"}, {"Three", "3"}})

	root := &widgets.TreeNode{
		Label:    "Root",
		Expanded: true,
		Children: []*widgets.TreeNode{{Label: "Child A"}, {Label: "Child B"}},
	}
	tree := widgets.NewTree(root)

	progress := widgets.NewProgress()
	progress.Value = 65

	spinner := widgets.NewSpinner()

	sparkData := state.NewSignal([]float64{1, 3, 2, 5, 4, 6, 5})
	spark := widgets.NewSparkline(sparkData)

	return &galleryRight{
		list:     list,
		table:    table,
		tree:     tree,
		progress: progress,
		spinner:  spinner,
		spark:    spark,
	}
}

func (r *galleryRight) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (r *galleryRight) Layout(bounds runtime.Rect) {
	r.Base.Layout(bounds)
	y := bounds.Y
	line := func(w runtime.Widget, height int) {
		if w == nil {
			return
		}
		w.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: height})
		y += height
	}
	line(r.list, 5)
	line(r.table, 4)
	line(r.tree, 4)
	line(r.progress, 1)
	line(r.spinner, 1)
	line(r.spark, 1)
}

func (r *galleryRight) Render(ctx runtime.RenderContext) {
	for _, child := range r.ChildWidgets() {
		if child != nil {
			child.Render(ctx)
		}
	}
}

func (r *galleryRight) HandleMessage(msg runtime.Message) runtime.HandleResult {
	for _, child := range r.ChildWidgets() {
		if child == nil {
			continue
		}
		if result := child.HandleMessage(msg); result.Handled {
			return result
		}
	}
	return runtime.Unhandled()
}

func (r *galleryRight) ChildWidgets() []runtime.Widget {
	children := []runtime.Widget{}
	if r.list != nil {
		children = append(children, r.list)
	}
	if r.table != nil {
		children = append(children, r.table)
	}
	if r.tree != nil {
		children = append(children, r.tree)
	}
	if r.progress != nil {
		children = append(children, r.progress)
	}
	if r.spinner != nil {
		children = append(children, r.spinner)
	}
	if r.spark != nil {
		children = append(children, r.spark)
	}
	return children
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
