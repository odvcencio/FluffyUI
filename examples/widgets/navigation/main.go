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
	view := NewNavigationView()
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

type NavigationView struct {
	widgets.Component
	header *widgets.Label
	tabs   *widgets.Tabs
}

func NewNavigationView() *NavigationView {
	view := &NavigationView{}
	view.header = widgets.NewLabel("Navigation Widgets").WithStyle(backend.DefaultStyle().Bold(true))

	menuTab := demo.NewVBox(newMenuSection(), newBreadcrumbSection())
	menuTab.Gap = 1

	steps := widgets.NewStepper(
		widgets.Step{Title: "Plan", State: widgets.StepCompleted},
		widgets.Step{Title: "Build", State: widgets.StepActive},
		widgets.Step{Title: "Ship", State: widgets.StepPending},
	)
	stepsPanel := widgets.NewPanel(steps).WithBorder(backend.DefaultStyle())
	stepsPanel.SetTitle("Stepper")

	palette := widgets.NewPaletteWidget("Quick Actions")
	palette.SetItems([]widgets.PaletteItem{
		{ID: "new", Label: "New file", Description: "Create a new file"},
		{ID: "open", Label: "Open", Description: "Open an existing file"},
		{ID: "save", Label: "Save", Description: "Save current changes"},
	})

	view.tabs = widgets.NewTabs(
		widgets.Tab{Title: "Menu", Content: menuTab},
		widgets.Tab{Title: "Stepper", Content: stepsPanel},
		widgets.Tab{Title: "Palette", Content: palette},
	)

	return view
}

func newMenuSection() runtime.Widget {
	status := widgets.NewLabel("Select a menu item")
	items := []*widgets.MenuItem{
		{Title: "Dashboard", OnSelect: func() { status.SetText("Dashboard selected") }},
		{Title: "Settings", OnSelect: func() { status.SetText("Settings selected") }},
		{
			Title: "More",
			Children: []*widgets.MenuItem{
				{Title: "Help", OnSelect: func() { status.SetText("Help selected") }},
				{Title: "About", OnSelect: func() { status.SetText("About selected") }},
			},
		},
	}
	menu := widgets.NewMenu(items...)
	panel := widgets.NewPanel(menu).WithBorder(backend.DefaultStyle())
	panel.SetTitle("Menu")

	section := demo.NewVBox(panel, status)
	section.Gap = 1
	return section
}

func newBreadcrumbSection() runtime.Widget {
	crumbs := widgets.NewBreadcrumb(
		widgets.BreadcrumbItem{Label: "Home"},
		widgets.BreadcrumbItem{Label: "Projects"},
		widgets.BreadcrumbItem{Label: "FluffyUI"},
	)
	panel := widgets.NewPanel(crumbs).WithBorder(backend.DefaultStyle())
	panel.SetTitle("Breadcrumb")
	return panel
}

func (n *NavigationView) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (n *NavigationView) Layout(bounds runtime.Rect) {
	n.Component.Layout(bounds)
	y := bounds.Y
	if n.header != nil {
		n.header.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: 1})
		y++
	}
	if n.tabs != nil {
		n.tabs.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: bounds.Height - (y - bounds.Y)})
	}
}

func (n *NavigationView) Render(ctx runtime.RenderContext) {
	if n.header != nil {
		n.header.Render(ctx)
	}
	if n.tabs != nil {
		n.tabs.Render(ctx)
	}
}

func (n *NavigationView) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if n.tabs != nil {
		if result := n.tabs.HandleMessage(msg); result.Handled {
			return result
		}
	}
	return runtime.Unhandled()
}

func (n *NavigationView) ChildWidgets() []runtime.Widget {
	children := []runtime.Widget{}
	if n.header != nil {
		children = append(children, n.header)
	}
	if n.tabs != nil {
		children = append(children, n.tabs)
	}
	return children
}
