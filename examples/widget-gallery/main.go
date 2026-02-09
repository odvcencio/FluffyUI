// Widget Gallery — interactive showcase of FluffyUI's widget library.
// Run: go run ./examples/widget-gallery
package main

import (
	"fmt"
	"log"

	"github.com/odvcencio/fluffyui/fluffy"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	"github.com/odvcencio/fluffyui/widgets"
)

func main() {
	if err := fluffy.Run(buildGallery()); err != nil {
		log.Fatal(err)
	}
}

func buildGallery() runtime.Widget {
	return widgets.NewTabs(
		widgets.Tab{Title: "Input", Content: inputTab()},
		widgets.Tab{Title: "Display", Content: displayTab()},
		widgets.Tab{Title: "Data", Content: dataTab()},
		widgets.Tab{Title: "Navigation", Content: navigationTab()},
		widgets.Tab{Title: "Feedback", Content: feedbackTab()},
		widgets.Tab{Title: "Pickers", Content: pickersTab()},
	)
}

// ─── Input Widgets ──────────────────────────────────────────────────────────

func inputTab() runtime.Widget {
	status := state.NewSignal("")

	nameInput := fluffy.Input("Enter your name...")
	emailInput := fluffy.Input("you@example.com")

	checked := state.NewSignal(false)
	cb := fluffy.Checkbox("Enable notifications", func(v *bool) {
		if v != nil {
			checked.Set(*v)
		}
	})

	toggleState := state.NewSignal(false)
	toggle := widgets.NewToggle(toggleState)
	toggle.SetLabel("Dark mode")

	sliderValue := state.NewSignal(42.0)
	slider := widgets.NewSlider(sliderValue,
		widgets.WithSliderRange(0, 100, 1),
	)

	numValue := state.NewSignal(42.0)
	numInput := widgets.NewNumberInput(numValue,
		widgets.WithNumberRange(0, 999),
		widgets.WithNumberStep(1),
	)

	tagInput := widgets.NewTagInput()
	tagInput.SetPlaceholder("Add tags...")

	submit := fluffy.Button("Submit", func() {
		name := nameInput.Text()
		email := emailInput.Text()
		status.Set(fmt.Sprintf("Submitted: %s <%s>", name, email))
	})

	reset := fluffy.Button("Reset", func() {
		nameInput.SetText("")
		emailInput.SetText("")
		status.Set("Form reset")
	})

	return widgets.NewScrollView(
		fluffy.VStack(
			section("Buttons",
				fluffy.HStack(
					fluffy.PrimaryButton("Primary", func() { status.Set("Primary clicked") }),
					fluffy.SecondaryButton("Secondary", func() { status.Set("Secondary clicked") }),
					fluffy.DangerButton("Danger", func() { status.Set("Danger clicked") }),
					fluffy.Button("Default", func() { status.Set("Default clicked") }),
				),
			),
			section("Text Input",
				fluffy.VStack(
					fluffy.Label("Name:"),
					nameInput,
					fluffy.Label("Email:"),
					emailInput,
					fluffy.HStack(submit, reset),
				),
			),
			section("Checkbox & Toggle",
				fluffy.VStack(
					cb,
					fluffy.ReactiveText(func() string {
						if checked.Get() {
							return "  Notifications: ON"
						}
						return "  Notifications: OFF"
					}, checked),
					toggle,
				),
			),
			section("Slider & Number Input",
				fluffy.VStack(
					fluffy.Label("Slider (0-100):"),
					slider,
					fluffy.Label("Number Input:"),
					numInput,
				),
			),
			section("Tag Input",
				tagInput,
			),
			fluffy.ReactiveText(func() string {
				s := status.Get()
				if s == "" {
					return ""
				}
				return "  Status: " + s
			}, status),
		),
	)
}

// ─── Display Widgets ────────────────────────────────────────────────────────

func displayTab() runtime.Widget {
	rating := widgets.NewRating(5)
	rating.SetValue(3)

	return widgets.NewScrollView(
		fluffy.VStack(
			section("Text & Labels",
				fluffy.VStack(
					fluffy.Text("Plain text widget"),
					fluffy.Label("Label widget"),
					fluffy.Labelf("Formatted label: %d items", 42),
				),
			),
			section("Badge",
				fluffy.HStack(
					widgets.NewBadge("New"),
					widgets.NewBadge("3"),
					widgets.NewBadge("Warning"),
					widgets.NewBadge("Error"),
				),
			),
			section("Avatar",
				fluffy.HStack(
					widgets.NewAvatar("Alice"),
					widgets.NewAvatar("Bob"),
					widgets.NewAvatar("Charlie"),
				),
			),
			section("Spinner",
				fluffy.HStack(
					widgets.NewSpinner(),
					fluffy.Label("  Loading..."),
				),
			),
			section("Skeleton (loading placeholder)",
				widgets.NewSkeleton(40, 3),
			),
			section("Rating (use arrow keys)",
				rating,
			),
			section("Link",
				widgets.NewLink("FluffyUI on GitHub", "https://github.com/odvcencio/FluffyUI"),
			),
			section("Card",
				widgets.NewCard("Example Card",
					fluffy.VStack(
						fluffy.Text("Cards group related content with a visual border."),
						fluffy.Text("They have a title and body area."),
					),
				),
			),
		),
	)
}

// ─── Data Widgets ───────────────────────────────────────────────────────────

func dataTab() runtime.Widget {
	table := widgets.NewTable(
		widgets.TableColumn{Title: "Name", Width: 15},
		widgets.TableColumn{Title: "Language", Width: 10},
		widgets.TableColumn{Title: "Stars", Width: 10},
	)
	table.SetRows([][]string{
		{"FluffyUI", "Go", "500+"},
		{"Bubble Tea", "Go", "39k"},
		{"Textual", "Python", "34k"},
		{"Ratatui", "Rust", "18k"},
		{"Ink", "JS/React", "27k"},
	})

	root := &widgets.TreeNode{Label: "Project", Expanded: true, Children: []*widgets.TreeNode{
		{Label: "src/", Children: []*widgets.TreeNode{
			{Label: "main.go"},
			{Label: "app.go"},
			{Label: "widgets/"},
		}},
		{Label: "docs/", Children: []*widgets.TreeNode{
			{Label: "README.md"},
			{Label: "CHANGELOG.md"},
		}},
		{Label: "go.mod"},
	}}
	tree := widgets.NewTree(root)

	return widgets.NewScrollView(
		fluffy.VStack(
			section("Table (sortable)", table),
			section("Tree", tree),
		),
	)
}

// ─── Navigation Widgets ─────────────────────────────────────────────────────

func navigationTab() runtime.Widget {
	breadcrumb := widgets.NewBreadcrumb(
		widgets.BreadcrumbItem{Label: "Home"},
		widgets.BreadcrumbItem{Label: "Products"},
		widgets.BreadcrumbItem{Label: "Widget Gallery"},
	)

	pagination := widgets.NewPagination(100, 10)

	stepper := widgets.NewStepper(
		widgets.Step{Title: "Account"},
		widgets.Step{Title: "Profile"},
		widgets.Step{Title: "Confirm"},
	)

	innerTabs := widgets.NewTabs(
		widgets.Tab{Title: "Tab A", Content: fluffy.Text("  Content for Tab A")},
		widgets.Tab{Title: "Tab B", Content: fluffy.Text("  Content for Tab B")},
		widgets.Tab{Title: "Tab C", Content: fluffy.Text("  Content for Tab C")},
	)

	return widgets.NewScrollView(
		fluffy.VStack(
			section("Breadcrumb", breadcrumb),
			section("Pagination (100 items, 10/page)", pagination),
			section("Stepper", stepper),
			section("Nested Tabs", innerTabs),
		),
	)
}

// ─── Feedback Widgets ───────────────────────────────────────────────────────

func feedbackTab() runtime.Widget {
	alertInfo := widgets.NewAlert("Info: FluffyUI is the most accessible TUI framework.", widgets.AlertInfo)
	alertWarn := widgets.NewAlert("Warning: You are about to do something awesome.", widgets.AlertWarning)
	alertErr := widgets.NewAlert("Error: No errors found. Everything is great!", widgets.AlertError)

	tooltipBtn := widgets.NewButton("Hover me for tooltip")
	tooltip := widgets.NewTooltip(tooltipBtn, fluffy.Text("This is a tooltip!"))

	progress := widgets.NewProgress()
	progress.Value = 65
	progress.Max = 100

	return widgets.NewScrollView(
		fluffy.VStack(
			section("Alerts",
				fluffy.VStack(alertInfo, alertWarn, alertErr),
			),
			section("Tooltip (focus the button)", tooltip),
			section("Progress Bar (65%)", progress),
		),
	)
}

// ─── Picker Widgets ─────────────────────────────────────────────────────────

func pickersTab() runtime.Widget {
	sel := fluffy.SelectFromStrings(
		[]string{"Option A", "Option B", "Option C", "Option D"},
		nil,
	)

	datePicker := widgets.NewDatePicker()
	timePicker := widgets.NewTimePicker()
	colorPicker := widgets.NewColorPicker()

	return widgets.NewScrollView(
		fluffy.VStack(
			section("Select", sel),
			section("Date Picker", datePicker),
			section("Time Picker", timePicker),
			section("Color Picker", colorPicker),
		),
	)
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func section(title string, content ...runtime.Widget) runtime.Widget {
	children := make([]runtime.Widget, 0, len(content)+1)
	children = append(children, fluffy.Label(fmt.Sprintf("── %s ──", title)))
	children = append(children, content...)
	return widgets.NewPanel(fluffy.VStack(children...))
}
