// Minimal form with name/email inputs and a submit button showing confirmation.
package main

import (
	"fmt"
	"log"
	"strings"

	"m31labs.dev/fluffyui/fluffy"
)

func main() {
	name := fluffy.Input("Your name")
	email := fluffy.Input("you@example.com")
	status := fluffy.Signal("")

	submit := func() {
		n := strings.TrimSpace(name.Text())
		e := strings.TrimSpace(email.Text())
		if n == "" || e == "" {
			status.Set("Please fill in both fields.")
			return
		}
		status.Set(fmt.Sprintf("Thanks, %s! We'll email %s.", n, e))
	}

	if err := fluffy.Run(
		fluffy.VStack(
			fluffy.Label("Sign Up"),
			name,
			email,
			fluffy.Button("Submit", submit),
			fluffy.ReactiveText(func() string { return status.Get() }, status),
		),
	); err != nil {
		log.Fatal(err)
	}
}
