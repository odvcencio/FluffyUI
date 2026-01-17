package widgets_test

import (
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/state"
	"github.com/odvcencio/fluffy-ui/widgets"
)

func ExampleText() {
	text := widgets.NewText("Hello\nWorld")
	text.SetStyle(backend.DefaultStyle().Bold(true))
	_ = text
}

func ExampleLabel() {
	label := widgets.NewLabel("Title")
	label.SetAlignment(widgets.AlignCenter)
	_ = label
}

func ExampleSignalLabel() {
	signal := state.NewSignal("Ready")
	label := widgets.NewSignalLabel(signal, state.DirectScheduler)
	_ = label
}

func ExampleSection() {
	section := widgets.NewSection("Pipeline")
	section.SetItems([]widgets.SectionItem{
		{Icon: '>', Text: "Build", Active: true},
		{Icon: 'o', Text: "Test"},
		{Icon: 'x', Text: "Deploy"},
	})
	section.SetMaxItems(3)
	_ = section
}
