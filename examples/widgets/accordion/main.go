package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/examples/internal/demo"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/widgets"
)

func main() {
	view := NewAccordionView()
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

type AccordionView struct {
	widgets.Component
	header    *widgets.Label
	accordion *widgets.Accordion
	footer    *widgets.Label
}

func NewAccordionView() *AccordionView {
	view := &AccordionView{}
	view.header = widgets.NewLabel("Accordion Widget").WithStyle(backend.DefaultStyle().Bold(true))

	overview := strings.Repeat("Accordion sections keep layouts tidy. ", 2)
	section1 := widgets.NewAccordionSection("Overview", widgets.NewText(overview), widgets.WithSectionExpanded(true))
	section2 := widgets.NewAccordionSection("Details", widgets.NewLabel("Use arrows + Enter to toggle."))
	section3 := widgets.NewAccordionSection("Disabled", widgets.NewLabel("This section is disabled."), widgets.WithSectionDisabled(true))

	view.accordion = widgets.NewAccordion(section1, section2, section3)
	view.footer = widgets.NewLabel("Tip: Arrow keys move between sections.")

	return view
}

func (a *AccordionView) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (a *AccordionView) Layout(bounds runtime.Rect) {
	a.Component.Layout(bounds)
	y := bounds.Y
	if a.header != nil {
		a.header.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: 1})
		y++
	}
	footerHeight := 1
	accordionHeight := bounds.Height - (y - bounds.Y) - footerHeight
	if accordionHeight < 0 {
		accordionHeight = 0
	}
	if a.accordion != nil {
		a.accordion.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: accordionHeight})
		y += accordionHeight
	}
	if a.footer != nil {
		a.footer.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: footerHeight})
	}
}

func (a *AccordionView) Render(ctx runtime.RenderContext) {
	for _, child := range a.ChildWidgets() {
		if child != nil {
			child.Render(ctx)
		}
	}
}

func (a *AccordionView) HandleMessage(msg runtime.Message) runtime.HandleResult {
	for _, child := range a.ChildWidgets() {
		if child == nil {
			continue
		}
		if result := child.HandleMessage(msg); result.Handled {
			return result
		}
	}
	return runtime.Unhandled()
}

func (a *AccordionView) ChildWidgets() []runtime.Widget {
	children := []runtime.Widget{}
	if a.header != nil {
		children = append(children, a.header)
	}
	if a.accordion != nil {
		children = append(children, a.accordion)
	}
	if a.footer != nil {
		children = append(children, a.footer)
	}
	return children
}
