package fur

import "testing"

func FuzzMarkupParse(f *testing.F) {
	f.Add("Hello [bold]world[/bold]")
	f.Add("[red on blue]colored[/red]")
	f.Add("escaped \\[bracket\\]")
	f.Add("[#ff0000]hex color[/#ff0000]")
	f.Add("[bold][italic]nested[/italic][/bold]")
	f.Add("")
	f.Add("[[[")
	f.Add("[unclosed tag")
	f.Add("normal text")
	f.Add("[bold on #abcdef italic]styled[/bold]")

	f.Fuzz(func(t *testing.T, input string) {
		p := NewMarkupParser()
		_ = p.Parse(input) // Should never panic
	})
}
