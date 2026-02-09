// Minimal tab navigation with three content panels.
package main

import (
	"log"

	"github.com/odvcencio/fluffyui/fluffy"
	"github.com/odvcencio/fluffyui/widgets"
)

func main() {
	tabs := widgets.NewTabs(
		widgets.Tab{Title: "Home", Content: fluffy.Label("Welcome to FluffyUI!")},
		widgets.Tab{Title: "About", Content: fluffy.Label("A Go TUI framework.")},
		widgets.Tab{Title: "Help", Content: fluffy.Label("Press Tab to switch panels.")},
	)

	if err := fluffy.Run(
		fluffy.VStack(
			fluffy.Label("Minimal Tabs"),
			tabs,
		),
	); err != nil {
		log.Fatal(err)
	}
}
