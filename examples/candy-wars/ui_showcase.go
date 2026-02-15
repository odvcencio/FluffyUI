package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/examples/internal/demo"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	"github.com/odvcencio/fluffyui/terminal"
	"github.com/odvcencio/fluffyui/widgets"
)

// ShowcaseTabContent renders a kitchen sink of widgets in a scrollable layout.
type ShowcaseTabContent struct {
	widgets.Component
	scroll     *widgets.ScrollView
	content    runtime.Widget
	focusables []runtime.Focusable
	focusIdx   int
	status     *widgets.Label
}

func NewShowcaseTabContent() *ShowcaseTabContent {
	s := &ShowcaseTabContent{focusIdx: -1}
	s.status = widgets.NewLabel("Tab to cycle focus • Type to interact • Scroll to explore")
	statusPanel := showcasePanel("Kitchen Sink", s.status)

	// Inputs.
	input := widgets.NewInput()
	input.SetPlaceholder("Type here")
	input.SetOnSubmit(func(text string) {
		s.setStatus("Input: " + truncateStatus(text, 24))
	})
	masked := widgets.NewMaskedInput("####-##", widgets.WithMaskedLabel("Promo code"))
	masked.SetOnChange(func(value string) {
		if value == "" {
			s.setStatus("Masked input cleared")
			return
		}
		s.setStatus("Masked: " + value)
	})
	auto := widgets.NewAutoComplete()
	auto.SetOptions([]string{"Alpha", "Beta", "Gamma", "Delta"})
	auto.SetOnSelect(func(value string) {
		s.setStatus("AutoComplete: " + value)
	})
	textarea := widgets.NewTextArea()
	textarea.SetLabel("Notes")
	textarea.SetText("Multi-line input\nwith scrolling")
	textarea.SetOnChange(func(text string) {
		s.setStatus(fmt.Sprintf("TextArea: %d chars", len(text)))
	})

	inputsBox := demo.NewVBox(input, masked, auto, textarea)
	inputsBox.Gap = 1
	inputsPanel := showcasePanel("Inputs", inputsBox)

	// Selection controls.
	checkbox := widgets.NewCheckbox("Enable feature")
	checkbox.SetOnChange(func(value *bool) {
		state := "off"
		if value != nil && *value {
			state = "on"
		}
		s.setStatus("Checkbox: " + state)
	})
	selecter := widgets.NewSelect(
		widgets.SelectOption{Label: "Low"},
		widgets.SelectOption{Label: "Medium"},
		widgets.SelectOption{Label: "High"},
	)
	selecter.SetOnChange(func(option widgets.SelectOption) {
		s.setStatus("Select: " + option.Label)
	})
	multiSel := widgets.NewMultiSelect(
		widgets.MultiSelectOption{Label: "One"},
		widgets.MultiSelectOption{Label: "Two"},
		widgets.MultiSelectOption{Label: "Three"},
	)
	multiSel.SetOnChange(func(selected []widgets.MultiSelectOption) {
		s.setStatus(fmt.Sprintf("MultiSelect: %d selected", len(selected)))
	})
	group := widgets.NewRadioGroup()
	radioA := widgets.NewRadio("Option A", group)
	radioB := widgets.NewRadio("Option B", group)
	group.SetOnChange(func(index int) {
		label := "A"
		if index == 1 {
			label = "B"
		}
		s.setStatus("Radio: " + label)
	})

	selectionBox := demo.NewVBox(checkbox, selecter, multiSel, radioA, radioB)
	selectionBox.Gap = 1
	selectionPanel := showcasePanel("Selections", selectionBox)

	// Sliders.
	value := state.NewSignal(35.0)
	slider := widgets.NewSlider(value, widgets.WithSliderRange(0, 100, 5), widgets.WithSliderShowValue(true))
	rangeMin := state.NewSignal(20.0)
	rangeMax := state.NewSignal(80.0)
	rangeSlider := widgets.NewRangeSlider(rangeMin, rangeMax, widgets.WithRangeSliderRange(0, 100, 5), widgets.WithRangeSliderShowValue(true))
	slidersBox := demo.NewVBox(slider, rangeSlider)
	slidersBox.Gap = 1
	slidersPanel := showcasePanel("Sliders", slidersBox)

	// Pickers.
	datePicker := widgets.NewDatePicker()
	if cal := datePicker.Calendar(); cal != nil {
		cal.SetOnSelect(func(date time.Time) {
			s.setStatus("Date: " + date.Format("Jan 2"))
		})
	}
	dateRange := widgets.NewDateRangePicker()
	dateRange.SetOnRangeSelect(func(start, end time.Time) {
		s.setStatus(fmt.Sprintf("Range: %s - %s", start.Format("Jan 2"), end.Format("Jan 2")))
	})
	timePick := widgets.NewTimePicker()
	timePick.SetShowSeconds(true)
	timePick.SetOnChange(func(value time.Time) {
		s.setStatus("Time: " + value.Format("15:04:05"))
	})

	pickersBox := demo.NewVBox(datePicker, dateRange, timePick)
	pickersBox.Gap = 1
	pickersPanel := showcasePanel("Pickers", pickersBox)

	// Navigation + commands.
	crumbs := widgets.NewBreadcrumb(
		widgets.BreadcrumbItem{Label: "Home"},
		widgets.BreadcrumbItem{Label: "Examples"},
		widgets.BreadcrumbItem{Label: "Candy Wars"},
	)
	crumbs.SetOnNavigate(func(index int) {
		if index >= 0 && index < len(crumbs.Items) {
			s.setStatus("Breadcrumb: " + crumbs.Items[index].Label)
		}
	})

	steps := widgets.NewStepper(
		widgets.Step{Title: "Plan", State: widgets.StepCompleted},
		widgets.Step{Title: "Build", State: widgets.StepActive},
		widgets.Step{Title: "Ship", State: widgets.StepPending},
	)

	menu := widgets.NewMenu(
		&widgets.MenuItem{ID: "open", Title: "Open", Shortcut: "Ctrl+O", OnSelect: func() { s.setStatus("Menu: Open") }},
		&widgets.MenuItem{ID: "save", Title: "Save", Shortcut: "Ctrl+S", OnSelect: func() { s.setStatus("Menu: Save") }},
		&widgets.MenuItem{ID: "export", Title: "Export", Shortcut: "Ctrl+E", OnSelect: func() { s.setStatus("Menu: Export") }},
	)

	palette := widgets.NewPaletteWidget("Quick Actions")
	palette.SetItems([]widgets.PaletteItem{
		{ID: "new", Label: "New Run", Description: "Start fresh"},
		{ID: "save", Label: "Save Game", Description: "Write save slot"},
		{ID: "stats", Label: "Open Stats", Description: "View performance"},
	})
	palette.SetOnSelect(func(item widgets.PaletteItem) {
		s.setStatus("Palette: " + item.Label)
	})

	search := widgets.NewSearchWidget()
	search.SetMatchInfo(1, 4)
	search.SetOnSearch(func(query string) {
		if query == "" {
			s.setStatus("Search cleared")
			return
		}
		s.setStatus("Search: " + truncateStatus(query, 20))
	})

	navBox := demo.NewVBox(crumbs, steps, menu)
	navBox.Gap = 1
	navPanel := showcasePanel("Navigation", navBox)

	commandBox := demo.NewVBox(palette, search)
	commandBox.Gap = 1
	commandPanel := showcasePanel("Commands", commandBox)

	// Data widgets.
	items := []string{"Alpha", "Beta", "Gamma", "Delta"}
	adapter := widgets.NewSliceAdapter(items, func(item string, index int, selected bool, ctx runtime.RenderContext) {
		style := backend.DefaultStyle()
		if selected {
			style = style.Reverse(true)
		}
		line := truncateStatus(item, ctx.Bounds.Width)
		ctx.Buffer.SetString(ctx.Bounds.X, ctx.Bounds.Y, line, style)
	})
	list := widgets.NewList(adapter)
	list.SetOnSelect(func(index int, item string) {
		s.setStatus("List: " + item)
	})

	table := widgets.NewTable(
		widgets.TableColumn{Title: "Name"},
		widgets.TableColumn{Title: "Value"},
	)
	table.SetRows([][]string{{"One", "1"}, {"Two", "2"}, {"Three", "3"}})

	grid := widgets.NewDataGrid([]widgets.DataGridColumn{
		{Header: "Key", Width: 10},
		{Header: "Value", Width: 10},
	})
	grid.SetRows([][]string{{"Alpha", "10"}, {"Beta", "20"}, {"Gamma", "30"}})
	grid.SetOnCellEdit(func(row, col int, oldValue, newValue string) {
		s.setStatus(fmt.Sprintf("DataGrid: [%d,%d] %s", row, col, truncateStatus(newValue, 12)))
	})

	dataBox := demo.NewVBox(list, table, grid)
	dataBox.Gap = 1
	dataPanel := showcasePanel("Data", dataBox)

	// Trees.
	root := &widgets.TreeNode{
		Label:    "Root",
		Expanded: true,
		Children: []*widgets.TreeNode{
			{Label: "Branch A", Expanded: true, Children: []*widgets.TreeNode{{Label: "Leaf A1"}, {Label: "Leaf A2"}}},
			{Label: "Branch B", Children: []*widgets.TreeNode{{Label: "Leaf B1"}}},
		},
	}
	tree := widgets.NewTree(root)

	rootDir := "."
	if cwd, err := os.Getwd(); err == nil {
		rootDir = cwd
	}
	dirTree := widgets.NewDirectoryTree(rootDir, widgets.WithShowHidden(false), widgets.WithLazyLoad(true), widgets.WithDirectoryFilter(func(entry os.DirEntry) bool {
		if entry.IsDir() {
			return true
		}
		name := strings.ToLower(entry.Name())
		return strings.HasSuffix(name, ".go")
	}))
	dirTree.SetOnSelect(func(path string) {
		s.setStatus("Dir: " + truncateStatus(path, 24))
	})

	treesBox := demo.NewVBox(tree, dirTree)
	treesBox.Gap = 1
	treesPanel := showcasePanel("Trees", treesBox)

	// Feedback.
	alert := widgets.NewAlert("All systems nominal", widgets.AlertSuccess)
	progress := widgets.NewProgress()
	progress.Label = "Capacity"
	progress.Value = 72
	spinner := widgets.NewSpinner()
	log := widgets.NewLog(widgets.WithShowTime(false), widgets.WithMaxLines(50))
	log.Info("Subsystem online")
	log.Warn("Latency spike detected")
	log.Error("Recovered from error")

	feedbackBox := demo.NewVBox(alert, progress, spinner, log)
	feedbackBox.Gap = 1
	feedbackPanel := showcasePanel("Feedback", feedbackBox)

	// Structure + markdown.
	accordion := widgets.NewAccordion(
		widgets.NewAccordionSection("Overview", widgets.NewText("Use arrows + Enter to toggle."), widgets.WithSectionExpanded(true)),
		widgets.NewAccordionSection("Details", widgets.NewText("Accordion supports multiple sections.")),
		widgets.NewAccordionSection("Disabled", widgets.NewText("This section is disabled."), widgets.WithSectionDisabled(true)),
	)

	section := widgets.NewSection("Build Steps")
	section.SetItems([]widgets.SectionItem{
		{Icon: '✓', Text: "Initialize project"},
		{Icon: '→', Text: "Render UI", Active: true, SubText: "running"},
		{Icon: '○', Text: "Ship"},
	})

	structureBox := demo.NewVBox(accordion, section)
	structureBox.Gap = 1
	structurePanel := showcasePanel("Structure", structureBox)

	rich := widgets.NewRichText("## Markdown\n- **Bold** support\n- Lists and `inline code`\n- _Italic_ emphasis\n")
	richPanel := showcasePanel("Markdown", rich)

	leftCol := widgets.VBox(
		widgets.FlexFixed(statusPanel),
		widgets.FlexFixed(inputsPanel),
		widgets.FlexFixed(selectionPanel),
		widgets.FlexFixed(slidersPanel),
		widgets.FlexFixed(pickersPanel),
		widgets.FlexFixed(navPanel),
		widgets.FlexFixed(commandPanel),
	)
	leftCol.Gap = 1

	rightCol := widgets.VBox(
		widgets.FlexFixed(dataPanel),
		widgets.FlexFixed(treesPanel),
		widgets.FlexFixed(feedbackPanel),
		widgets.FlexFixed(structurePanel),
		widgets.FlexFixed(richPanel),
	)
	rightCol.Gap = 1

	columns := newShowcaseColumns(leftCol, rightCol, 2)
	s.content = columns
	s.scroll = widgets.NewScrollView(columns)
	s.scroll.SetLabel("Kitchen Sink Showcase")

	// Focusable ordering (top-to-bottom).
	s.addFocusable(input)
	s.addFocusable(masked)
	s.addFocusable(auto)
	s.addFocusable(textarea)
	s.addFocusable(checkbox)
	s.addFocusable(selecter)
	s.addFocusable(multiSel)
	s.addFocusable(radioA)
	s.addFocusable(slider)
	s.addFocusable(rangeSlider)
	s.addFocusable(datePicker)
	s.addFocusable(dateRange)
	s.addFocusable(timePick)
	s.addFocusable(crumbs)
	s.addFocusable(menu)
	s.addFocusable(palette)
	s.addFocusable(search)
	s.addFocusable(list)
	s.addFocusable(grid)
	s.addFocusable(tree)
	s.addFocusable(dirTree)
	s.addFocusable(log)
	s.addFocusable(accordion)
	s.addFocusable(section)

	return s
}

func (s *ShowcaseTabContent) Measure(constraints runtime.Constraints) runtime.Size {
	if s.scroll != nil {
		return s.scroll.Measure(constraints)
	}
	return constraints.MinSize()
}

func (s *ShowcaseTabContent) Layout(bounds runtime.Rect) {
	s.Component.Layout(bounds)
	if s.scroll != nil {
		s.scroll.Layout(bounds)
	}
}

func (s *ShowcaseTabContent) Render(ctx runtime.RenderContext) {
	if s.scroll != nil {
		s.scroll.Render(ctx)
	}
}

func (s *ShowcaseTabContent) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if s == nil {
		return runtime.Unhandled()
	}
	if key, ok := msg.(runtime.KeyMsg); ok && key.Key == terminal.KeyTab {
		if key.Shift {
			s.setFocus(s.focusIdx - 1)
		} else {
			s.setFocus(s.focusIdx + 1)
		}
		return runtime.Handled()
	}
	s.ensureFocus()
	if s.focusIdx >= 0 && s.focusIdx < len(s.focusables) {
		if result := s.focusables[s.focusIdx].HandleMessage(msg); result.Handled {
			return result
		}
	}
	if s.scroll != nil {
		return s.scroll.HandleMessage(msg)
	}
	return runtime.Unhandled()
}

func (s *ShowcaseTabContent) ChildWidgets() []runtime.Widget {
	if s.scroll == nil {
		return nil
	}
	return []runtime.Widget{s.scroll}
}

func (s *ShowcaseTabContent) setStatus(text string) {
	if s == nil || s.status == nil {
		return
	}
	s.status.SetText(text)
	s.Invalidate()
}

func (s *ShowcaseTabContent) addFocusable(w runtime.Widget) {
	if w == nil {
		return
	}
	if f, ok := w.(runtime.Focusable); ok && f.CanFocus() {
		s.focusables = append(s.focusables, f)
	}
}

func (s *ShowcaseTabContent) ensureFocus() {
	if len(s.focusables) == 0 {
		return
	}
	if s.focusIdx >= 0 && s.focusIdx < len(s.focusables) {
		return
	}
	s.setFocus(0)
}

func (s *ShowcaseTabContent) setFocus(index int) {
	if len(s.focusables) == 0 {
		return
	}
	if index < 0 {
		index = len(s.focusables) - 1
	}
	if index >= len(s.focusables) {
		index = 0
	}
	if s.focusIdx >= 0 && s.focusIdx < len(s.focusables) {
		s.focusables[s.focusIdx].Blur()
	}
	s.focusIdx = index
	s.focusables[s.focusIdx].Focus()
	s.Invalidate()
}

type showcaseColumns struct {
	widgets.Base
	left  runtime.Widget
	right runtime.Widget
	gap   int
}

func newShowcaseColumns(left, right runtime.Widget, gap int) *showcaseColumns {
	return &showcaseColumns{left: left, right: right, gap: gap}
}

func (c *showcaseColumns) Measure(constraints runtime.Constraints) runtime.Size {
	if c == nil {
		return constraints.MinSize()
	}
	width := constraints.MaxWidth
	if width <= 0 {
		width = constraints.MinWidth
	}
	if width <= 0 {
		width = 1
	}
	available := width - c.gap
	if available < 1 {
		available = width
	}
	colWidth := available / 2
	if colWidth < 1 {
		colWidth = 1
	}
	maxInt := int(^uint(0) >> 1)
	childConstraints := runtime.Constraints{MaxWidth: colWidth, MaxHeight: maxInt}
	leftSize := runtime.Size{}
	rightSize := runtime.Size{}
	if c.left != nil {
		leftSize = c.left.Measure(childConstraints)
	}
	if c.right != nil {
		rightSize = c.right.Measure(childConstraints)
	}
	height := leftSize.Height
	if rightSize.Height > height {
		height = rightSize.Height
	}
	if height < 1 {
		height = 1
	}
	return constraints.Constrain(runtime.Size{Width: width, Height: height})
}

func (c *showcaseColumns) Layout(bounds runtime.Rect) {
	c.Base.Layout(bounds)
	if c == nil || bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	available := bounds.Width - c.gap
	if available < 1 {
		available = bounds.Width
	}
	colWidth := available / 2
	if colWidth < 1 {
		colWidth = 1
	}
	rightWidth := bounds.Width - colWidth - c.gap
	if rightWidth < 1 {
		rightWidth = 1
	}
	leftBounds := runtime.Rect{X: bounds.X, Y: bounds.Y, Width: colWidth, Height: bounds.Height}
	rightBounds := runtime.Rect{X: bounds.X + colWidth + c.gap, Y: bounds.Y, Width: rightWidth, Height: bounds.Height}
	if c.left != nil {
		c.left.Layout(leftBounds)
	}
	if c.right != nil {
		c.right.Layout(rightBounds)
	}
}

func (c *showcaseColumns) Render(ctx runtime.RenderContext) {
	runtime.RenderChild(ctx, c.left)
	runtime.RenderChild(ctx, c.right)
}

func (c *showcaseColumns) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if c.left != nil {
		if result := c.left.HandleMessage(msg); result.Handled {
			return result
		}
	}
	if c.right != nil {
		if result := c.right.HandleMessage(msg); result.Handled {
			return result
		}
	}
	return runtime.Unhandled()
}

func (c *showcaseColumns) ChildWidgets() []runtime.Widget {
	if c == nil {
		return nil
	}
	children := make([]runtime.Widget, 0, 2)
	if c.left != nil {
		children = append(children, c.left)
	}
	if c.right != nil {
		children = append(children, c.right)
	}
	return children
}

func showcasePanel(title string, child runtime.Widget) *widgets.Panel {
	panel := widgets.NewPanel(child, widgets.WithPanelBorder(backend.DefaultStyle()))
	panel.SetTitle(title)
	return panel
}

func truncateStatus(text string, max int) string {
	if len(text) <= max {
		return text
	}
	if max <= 3 {
		return text[:max]
	}
	return text[:max-3] + "..."
}
