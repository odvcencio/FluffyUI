// Minimal temperature converter in ~35 lines using only the fluffy convenience API.
// Demonstrates bidirectional reactive state and computed values.
package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/odvcencio/fluffyui/fluffy"
)

func main() {
	celsius := fluffy.Signal(0.0)

	cInput := fluffy.Input("Celsius")
	cInput.SetText("0")
	cInput.SetOnChange(func(text string) {
		text = strings.TrimSpace(text)
		if v, err := strconv.ParseFloat(text, 64); err == nil {
			celsius.Set(v)
		}
	})

	fInput := fluffy.Input("Fahrenheit")
	fInput.SetText("32")
	fInput.SetOnChange(func(text string) {
		text = strings.TrimSpace(text)
		if v, err := strconv.ParseFloat(text, 64); err == nil {
			celsius.Set((v - 32) * 5 / 9)
		}
	})

	if err := fluffy.Run(
		fluffy.VStack(
			fluffy.Label("Temperature Converter"),
			fluffy.HStack(cInput, fluffy.Label("C")),
			fluffy.HStack(fInput, fluffy.Label("F")),
			fluffy.ReactiveText(func() string {
				c := celsius.Get()
				f := c*9/5 + 32
				return fmt.Sprintf("\n  %.1f C = %.1f F", c, f)
			}, celsius),
		),
	); err != nil {
		log.Fatal(err)
	}
}
