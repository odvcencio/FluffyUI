package main

import (
	"context"
	"fmt"
	"os"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/examples/internal/demo"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/widgets"
)

func main() {
	view := NewDataView()
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

type DataView struct {
	widgets.Component
	header   *widgets.Label
	list     *widgets.List[string]
	table    *widgets.Table
	grid     *widgets.DataGrid
	richText *widgets.RichText
	tree     *widgets.Tree
	splitter *widgets.Splitter
}

func NewDataView() *DataView {
	view := &DataView{}
	view.header = widgets.NewLabel("Data Widgets").WithStyle(backend.DefaultStyle().Bold(true))

	items := []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon", "Zeta", "Eta"}
	adapter := widgets.NewSliceAdapter(items, func(item string, index int, selected bool, ctx runtime.RenderContext) {
		style := backend.DefaultStyle()
		if selected {
			style = style.Reverse(true)
		}
		line := truncateAndPad(item, ctx.Bounds.Width)
		ctx.Buffer.SetString(ctx.Bounds.X, ctx.Bounds.Y, line, style)
	})
	view.list = widgets.NewList(adapter)
	leftPanel := widgets.NewPanel(view.list).WithBorder(backend.DefaultStyle())
	leftPanel.SetTitle("List")

	view.table = widgets.NewTable(
		widgets.TableColumn{Title: "Name"},
		widgets.TableColumn{Title: "Value"},
	)
	view.table.SetRows([][]string{{"One", "1"}, {"Two", "2"}, {"Three", "3"}, {"Four", "4"}})

	view.grid = widgets.NewDataGrid(
		widgets.TableColumn{Title: "Key"},
		widgets.TableColumn{Title: "Value"},
	)
	view.grid.SetRows([][]string{{"Alpha", "10"}, {"Beta", "20"}, {"Gamma", "30"}})

	view.richText = widgets.NewRichText("## Notes\n- Fast grid editing\n- Styled markdown\n")

	root := &widgets.TreeNode{
		Label:    "Root",
		Expanded: true,
		Children: []*widgets.TreeNode{
			{Label: "Branch A", Expanded: true, Children: []*widgets.TreeNode{{Label: "Leaf A1"}, {Label: "Leaf A2"}}},
			{Label: "Branch B", Children: []*widgets.TreeNode{{Label: "Leaf B1"}}},
		},
	}
	view.tree = widgets.NewTree(root)

	tablePanel := widgets.NewPanel(view.table).WithBorder(backend.DefaultStyle())
	tablePanel.SetTitle("Table")
	gridPanel := widgets.NewPanel(view.grid).WithBorder(backend.DefaultStyle())
	gridPanel.SetTitle("DataGrid")
	treePanel := widgets.NewPanel(view.tree).WithBorder(backend.DefaultStyle())
	treePanel.SetTitle("Tree")
	richPanel := widgets.NewPanel(view.richText).WithBorder(backend.DefaultStyle())
	richPanel.SetTitle("RichText")
	rightColumn := demo.NewVBox(tablePanel, gridPanel, treePanel, richPanel)
	rightColumn.Gap = 1

	view.splitter = widgets.NewSplitter(leftPanel, rightColumn)
	view.splitter.Ratio = 0.4
	return view
}

func (d *DataView) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (d *DataView) Layout(bounds runtime.Rect) {
	d.Component.Layout(bounds)
	y := bounds.Y
	if d.header != nil {
		d.header.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: 1})
		y++
	}
	if d.splitter != nil {
		d.splitter.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: bounds.Height - (y - bounds.Y)})
	}
}

func (d *DataView) Render(ctx runtime.RenderContext) {
	if d.header != nil {
		d.header.Render(ctx)
	}
	if d.splitter != nil {
		d.splitter.Render(ctx)
	}
}

func (d *DataView) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if d.splitter != nil {
		return d.splitter.HandleMessage(msg)
	}
	return runtime.Unhandled()
}

func (d *DataView) ChildWidgets() []runtime.Widget {
	children := []runtime.Widget{}
	if d.header != nil {
		children = append(children, d.header)
	}
	if d.splitter != nil {
		children = append(children, d.splitter)
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
