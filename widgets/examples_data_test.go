package widgets_test

import (
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/widgets"
)

func ExampleList() {
	items := []string{"alpha", "beta", "gamma"}
	adapter := widgets.NewSliceAdapter(items, func(item string, index int, selected bool, ctx runtime.RenderContext) {
		prefix := "  "
		if selected {
			prefix = "> "
		}
		ctx.Buffer.SetString(ctx.Bounds.X, ctx.Bounds.Y, prefix+item, backend.DefaultStyle())
	})
	list := widgets.NewList(adapter)
	list.SetSelected(1)
	_ = list
}

func ExampleTable() {
	table := widgets.NewTable(
		widgets.TableColumn{Title: "Name", Width: 12},
		widgets.TableColumn{Title: "Status", Width: 8},
	)
	table.SetRows([][]string{
		{"alpha", "ok"},
		{"beta", "warn"},
	})
	_ = table
}

func ExampleTree() {
	root := &widgets.TreeNode{
		Label:    "root",
		Expanded: true,
		Children: []*widgets.TreeNode{
			{Label: "configs"},
			{
				Label:    "data",
				Expanded: true,
				Children: []*widgets.TreeNode{
					{Label: "2024"},
				},
			},
		},
	}
	tree := widgets.NewTree(root)
	_ = tree
}

func ExampleSearchWidget() {
	search := widgets.NewSearchWidget()
	search.SetOnSearch(func(query string) {})
	search.SetMatchInfo(1, 4)
	_ = search
}
