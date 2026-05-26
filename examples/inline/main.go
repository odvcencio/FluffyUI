package main

import (
	"log"

	"m31labs.dev/fluffyui/fluffy"
	"m31labs.dev/fluffyui/runtime"
)

func main() {
	root := fluffy.VStack(
		fluffy.Label("Inline Mode Demo"),
		fluffy.Text("Press q to exit."),
		fluffy.Text("This runs without alternate-screen takeover."),
	)

	err := fluffy.RunInline(
		root,
		8,
		fluffy.WithUpdate(func(app *runtime.App, msg runtime.Message) bool {
			if key, ok := msg.(runtime.KeyMsg); ok && key.Rune == 'q' {
				app.ExecuteCommand(runtime.Quit{})
				return true
			}
			return runtime.DefaultUpdate(app, msg)
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
}
