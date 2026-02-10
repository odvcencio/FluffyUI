package style

import "testing"

func FuzzParse(f *testing.F) {
	// Seed with valid FSS inputs
	f.Add(`Button { padding: 1; }`)
	f.Add(`Button.primary:focus { foreground: red; bold: true; }`)
	f.Add(`@media (min-width: 80) { Panel { background: #333; } }`)
	f.Add(`Button:hover:active { background: blue; }`)
	f.Add(`#my-id .class1.class2 { margin: 1 2 3 4; }`)
	f.Add(`/* comment */ Button { color: rgb(255, 0, 0); }`)
	f.Add(`Button { transition: background 200ms ease-in-out; }`)
	f.Add(``)
	f.Add(`{{{`)
	f.Add(`Button { unknown-property: value; }`)
	f.Add(`@media { }`)

	f.Fuzz(func(t *testing.T, input string) {
		// Parse should never panic regardless of input
		_, _ = Parse(input)
	})
}
