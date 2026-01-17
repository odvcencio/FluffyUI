package widgets_test

import (
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/widgets"
)

func ExampleGrid() {
	grid := widgets.NewGrid(2, 2)
	grid.Gap = 1
	grid.Add(widgets.NewLabel("Top"), 0, 0, 1, 2)
	grid.Add(widgets.NewLabel("Left"), 1, 0, 1, 1)
	grid.Add(widgets.NewLabel("Right"), 1, 1, 1, 1)
	_ = grid
}

func ExampleSplitter() {
	left := widgets.NewText("Logs")
	right := widgets.NewText("Details")
	split := widgets.NewSplitter(left, right)
	split.Orientation = widgets.SplitHorizontal
	split.Ratio = 0.65
	_ = split
}

func ExampleStack() {
	base := widgets.NewBox(widgets.NewLabel("Base"))
	overlay := widgets.NewPanel(widgets.NewLabel("Overlay")).WithBorder(backend.DefaultStyle())
	stack := widgets.NewStack(base, overlay)
	_ = stack
}

func ExampleScrollView() {
	content := widgets.NewText("Line 1\nLine 2\nLine 3")
	scroll := widgets.NewScrollView(content)
	scroll.ScrollBy(0, 1)
	_ = scroll
}

func ExamplePanel() {
	panel := widgets.NewPanel(widgets.NewLabel("Details")).WithBorder(backend.DefaultStyle())
	panel.SetTitle("Summary")
	_ = panel
}

func ExampleBox() {
	box := widgets.NewBox(widgets.NewLabel("Status"))
	box.SetStyle(backend.DefaultStyle().Reverse(true))
	_ = box
}
