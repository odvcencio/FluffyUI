package widgets_test

import (
	"github.com/odvcencio/fluffyui/keybind"
	"github.com/odvcencio/fluffyui/widgets"
)

func ExampleTabs() {
	tabs := widgets.NewTabs(
		widgets.Tab{Title: "Home", Content: widgets.NewText("Welcome")},
		widgets.Tab{Title: "Logs", Content: widgets.NewText("...")},
	)
	_ = tabs
}

func ExampleMenu() {
	menu := widgets.NewMenu(
		&widgets.MenuItem{Title: "Open", Shortcut: "Ctrl+O"},
		&widgets.MenuItem{Title: "Save", Shortcut: "Ctrl+S"},
	)
	_ = menu
}

func ExampleBreadcrumb() {
	breadcrumb := widgets.NewBreadcrumb(
		widgets.BreadcrumbItem{Label: "home"},
		widgets.BreadcrumbItem{Label: "docs"},
		widgets.BreadcrumbItem{Label: "readme.md"},
	)
	_ = breadcrumb
}

func ExampleStepper() {
	stepper := widgets.NewStepper(
		widgets.Step{Title: "Fetch", State: widgets.StepCompleted},
		widgets.Step{Title: "Build", State: widgets.StepActive},
		widgets.Step{Title: "Ship", State: widgets.StepPending},
	)
	_ = stepper
}

func ExamplePaletteWidget() {
	palette := widgets.NewPaletteWidget("Commands")
	palette.SetItems([]widgets.PaletteItem{
		{ID: "new", Category: "File", Label: "New file", Shortcut: "Ctrl+N"},
		{ID: "open", Category: "File", Label: "Open file", Shortcut: "Ctrl+O"},
	})
	_ = palette
}

func ExampleEnhancedPalette() {
	registry := keybind.NewRegistry()
	registry.Register(keybind.Command{
		ID:          "app.quit",
		Title:       "Quit",
		Description: "Exit the app",
	})
	palette := widgets.NewEnhancedPalette(registry)
	_ = palette
}
