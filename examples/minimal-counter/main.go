// Minimal counter in ~25 lines using only the fluffy convenience API.
// Compare with examples/counter/ which uses the full widget system.
package main

import (
	"fmt"
	"log"

	"m31labs.dev/fluffyui/fluffy"
)

func main() {
	count := fluffy.Signal(0)

	if err := fluffy.Run(
		fluffy.VStack(
			fluffy.ReactiveText(func() string {
				return fmt.Sprintf("  Count: %d  \n\n  Press the buttons or +/-/r keys", count.Get())
			}, count),
			fluffy.HStack(
				fluffy.Button("+1", func() { count.Set(count.Get() + 1) }),
				fluffy.Button("-1", func() { count.Set(count.Get() - 1) }),
				fluffy.Button("Reset", func() { count.Set(0) }),
			),
		),
	); err != nil {
		log.Fatal(err)
	}
}
