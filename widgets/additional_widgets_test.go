package widgets

import (
	"strings"
	"testing"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
	flufftest "github.com/odvcencio/fluffyui/testing"
)

func TestListSelectionRender(t *testing.T) {
	items := []string{"Alpha", "Beta", "Gamma"}
	adapter := NewSliceAdapter(items, func(item string, index int, selected bool, ctx runtime.RenderContext) {
		prefix := "  "
		if selected {
			prefix = "> "
		}
		ctx.Buffer.SetString(ctx.Bounds.X, ctx.Bounds.Y, prefix+item, backend.DefaultStyle())
	})
	list := NewList(adapter)
	list.SetSelected(1)
	out := flufftest.RenderToString(list, 12, 3)
	if !strings.Contains(out, "> Beta") {
		t.Fatalf("expected selected row to render with prefix, got:\n%s", out)
	}
}

func TestTableSetCell(t *testing.T) {
	table := NewTable(
		TableColumn{Title: "Name"},
		TableColumn{Title: "Value"},
	)
	table.SetRows([][]string{{"One", "1"}, {"Two", "2"}})
	table.SetCell(1, 1, "42")
	out := flufftest.RenderToString(table, 20, 4)
	if !strings.Contains(out, "Name") || !strings.Contains(out, "Two") {
		t.Fatalf("expected table to render header/rows, got:\n%s", out)
	}
	if !strings.Contains(out, "42") {
		t.Fatalf("expected updated cell to render, got:\n%s", out)
	}
}

func TestTreeToggle(t *testing.T) {
	root := &TreeNode{
		Label:    "Root",
		Expanded: true,
		Children: []*TreeNode{{Label: "Child"}},
	}
	tree := NewTree(root)
	out := flufftest.RenderToString(tree, 20, 3)
	if !strings.Contains(out, "- Root") {
		t.Fatalf("expected expanded prefix, got:\n%s", out)
	}
	tree.Focus()
	tree.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	out = flufftest.RenderToString(tree, 20, 3)
	if !strings.Contains(out, "+ Root") {
		t.Fatalf("expected collapsed prefix, got:\n%s", out)
	}
}

func TestMenuToggle(t *testing.T) {
	menu := NewMenu(
		&MenuItem{Title: "File", Expanded: true, Children: []*MenuItem{{Title: "Open"}}},
		&MenuItem{Title: "Help"},
	)
	out := flufftest.RenderToString(menu, 20, 3)
	if !strings.Contains(out, "- File") {
		t.Fatalf("expected expanded menu prefix, got:\n%s", out)
	}
	menu.Focus()
	menu.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	out = flufftest.RenderToString(menu, 20, 3)
	if !strings.Contains(out, "+ File") {
		t.Fatalf("expected collapsed menu prefix, got:\n%s", out)
	}
}

func TestPanelTitleRender(t *testing.T) {
	panel := NewPanel(NewLabel("Content"), WithPanelBorder(backend.DefaultStyle()), WithPanelTitle("Stats"))
	out := flufftest.RenderToString(panel, 20, 5)
	if !strings.Contains(out, "Stats") {
		t.Fatalf("expected panel title to render, got:\n%s", out)
	}
}

func TestScrollViewScrolls(t *testing.T) {
	text := NewText("Line1\nLine2\nLine3")
	view := NewScrollView(text)
	view.ScrollBy(0, 1)
	out := flufftest.RenderToString(view, 10, 2)
	if !strings.Contains(out, "Line2") {
		t.Fatalf("expected scrolled content to include Line2, got:\n%s", out)
	}
}
