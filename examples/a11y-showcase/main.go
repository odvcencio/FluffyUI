// A11y Showcase: a mini-dashboard demonstrating that FluffyUI gives you
// full accessibility -- ARIA roles, focus announcements, and TTS -- for free.
// Run with FLUFFYUI_TTS=1 to hear every widget announce itself.
package main

import (
	"log"

	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/fluffy"
	"m31labs.dev/fluffyui/state"
	"m31labs.dev/fluffyui/widgets"
)

func main() {
	if err := fluffy.Run(buildUI()); err != nil {
		log.Fatal(err)
	}
}

// ---------------------------------------------------------------------------
// UI assembly -- everything is composed from pre-built widgets.
// ---------------------------------------------------------------------------

func buildUI() *fluffy.Flex {
	return fluffy.VFlex(
		fluffy.Expanded(buildTabs()),
		fluffy.Fixed(statusBar()),
	)
}

func buildTabs() *widgets.Tabs {
	tabs := widgets.NewTabs(
		widgets.Tab{Title: "Dashboard", Content: dashboardTab()},
		widgets.Tab{Title: "Settings", Content: settingsTab()},
		widgets.Tab{Title: "About", Content: aboutTab()},
	)
	tabs.SetLabel("Main Navigation")
	return tabs
}

// ---------------------------------------------------------------------------
// Tab 1: Dashboard
// ---------------------------------------------------------------------------

func dashboardTab() *fluffy.Flex {
	header := fluffy.Label("FluffyUI Accessibility Showcase",
		widgets.WithLabelStyle(backend.DefaultStyle().Bold(true)))

	cards := fluffy.HStack(
		statCard("Users", "1,234"),
		statCard("Revenue", "$45.6K"),
		statCard("Uptime", "99.9%"),
		statCard("Tasks", "42"),
	)

	table := buildTable()

	return fluffy.VFlex(
		fluffy.Fixed(header),
		fluffy.FixedSpace(1),
		fluffy.Fixed(cards),
		fluffy.FixedSpace(1),
		fluffy.Expanded(table),
	)
}

func statCard(title, value string) *widgets.Panel {
	label := fluffy.Label(title+": "+value,
		widgets.WithLabelA11yLabel(title),
		widgets.WithLabelStyle(backend.DefaultStyle().Bold(true)))
	return widgets.NewPanel(label,
		widgets.WithPanelBorder(backend.DefaultStyle()),
		widgets.WithPanelTitle(title))
}

func buildTable() *widgets.Table {
	table := widgets.NewTable(
		widgets.TableColumn{Title: "Name", Width: 20},
		widgets.TableColumn{Title: "Status", Width: 12},
		widgets.TableColumn{Title: "Priority", Width: 10},
	)
	table.SetLabel("Team Tasks")
	table.SetRows([][]string{
		{"Alice Johnson", "Active", "High"},
		{"Bob Smith", "Pending", "Medium"},
		{"Carol Lee", "Active", "Low"},
		{"Dave Patel", "Review", "High"},
		{"Eve Zhang", "Done", "Medium"},
	})
	return table
}

// ---------------------------------------------------------------------------
// Tab 2: Settings
// ---------------------------------------------------------------------------

func settingsTab() *fluffy.Flex {
	header := fluffy.Label("Settings",
		widgets.WithLabelStyle(backend.DefaultStyle().Bold(true)))

	nameInput := fluffy.Input("Enter your name")
	nameInput.SetLabel("Name")

	darkMode := state.NewSignal(false)
	toggle := widgets.NewToggle(darkMode,
		widgets.WithToggleLabel(state.NewSignal("Dark Mode")))

	themeSelect := fluffy.SelectFromStrings(
		[]string{"Default", "High Contrast", "Dark"},
		nil,
	)
	themeSelect.SetLabel("Theme")

	ttsCheckbox := fluffy.Checkbox("Enable TTS", nil)

	submitBtn := fluffy.PrimaryButton("Save Settings", func() {})

	return fluffy.VFlex(
		fluffy.Fixed(header),
		fluffy.FixedSpace(1),
		fluffy.Fixed(fluffy.Label("Name:")),
		fluffy.Fixed(nameInput),
		fluffy.FixedSpace(1),
		fluffy.Fixed(toggle),
		fluffy.FixedSpace(1),
		fluffy.Fixed(fluffy.Label("Theme:")),
		fluffy.Fixed(themeSelect),
		fluffy.FixedSpace(1),
		fluffy.Fixed(ttsCheckbox),
		fluffy.FixedSpace(1),
		fluffy.Fixed(submitBtn),
	)
}

// ---------------------------------------------------------------------------
// Tab 3: About
// ---------------------------------------------------------------------------

func aboutTab() *fluffy.Flex {
	header := fluffy.Label("About FluffyUI",
		widgets.WithLabelStyle(backend.DefaultStyle().Bold(true)))

	body := widgets.NewText(
		"This app demonstrates FluffyUI's built-in accessibility.\n" +
			"Every widget announces its role and state via TTS --\n" +
			"you get screen reader support for free, just by using\n" +
			"the standard widget library.\n" +
			"\n" +
			"Try navigating with:\n" +
			"  Tab       Move focus between widgets\n" +
			"  Arrows    Navigate within a widget\n" +
			"  Enter     Activate buttons and toggles\n" +
			"  Space     Toggle checkboxes\n" +
			"\n" +
			"Set FLUFFYUI_TTS=1 to hear every interaction spoken aloud.")

	return fluffy.VFlex(
		fluffy.Fixed(header),
		fluffy.FixedSpace(1),
		fluffy.Expanded(body),
	)
}

// ---------------------------------------------------------------------------
// Status bar
// ---------------------------------------------------------------------------

func statusBar() *widgets.Label {
	return fluffy.Label(
		"Tab to navigate | Enter to activate | FLUFFYUI_TTS=1 for speech",
		widgets.WithLabelStyle(backend.DefaultStyle().Reverse(true)))
}
