// Minimal todo list in ~45 lines using only the fluffy convenience API.
// Demonstrates reactive state with slices, input handling, and dynamic text.
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/odvcencio/fluffyui/fluffy"
)

func main() {
	items := fluffy.Signal([]string{})
	// Slices are compared by reflect.DeepEqual, which works for append patterns.

	input := fluffy.Input("What needs to be done?")

	addItem := func() {
		text := strings.TrimSpace(input.Text())
		if text == "" {
			return
		}
		current := items.Get()
		items.Set(append(append([]string{}, current...), text))
		input.SetText("")
	}

	input.SetOnSubmit(func(_ string) { addItem() })

	if err := fluffy.Run(
		fluffy.VStack(
			fluffy.Label("Minimal Todo"),
			fluffy.HStack(
				input,
				fluffy.Button("Add", addItem),
			),
			fluffy.ReactiveText(func() string {
				list := items.Get()
				if len(list) == 0 {
					return "  No items yet. Type something and press Enter."
				}
				var b strings.Builder
				for i, item := range list {
					fmt.Fprintf(&b, "  %d. %s\n", i+1, item)
				}
				fmt.Fprintf(&b, "\n  %d item(s)", len(list))
				return b.String()
			}, items),
		),
	); err != nil {
		log.Fatal(err)
	}
}
