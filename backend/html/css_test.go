package html

import (
	"strings"
	"testing"

	"m31labs.dev/fluffyui/compositor"
	"m31labs.dev/fluffyui/style"
)

func TestSelectorToCSS_Type(t *testing.T) {
	sel := style.Selector{Type: "Chip"}
	got := selectorToCSS(sel)
	if got != ".fluffy-Chip" {
		t.Errorf("selectorToCSS(Chip) = %q, want .fluffy-Chip", got)
	}
}

func TestSelectorToCSS_Universal(t *testing.T) {
	sel := style.Selector{Type: "*"}
	got := selectorToCSS(sel)
	if got != "*" {
		t.Errorf("selectorToCSS(*) = %q, want *", got)
	}
}

func TestSelectorToCSS_Pseudo(t *testing.T) {
	sel := style.Selector{Type: "Chip", Pseudo: []style.PseudoClass{"hover"}}
	got := selectorToCSS(sel)
	if got != ".fluffy-Chip:hover" {
		t.Errorf("got %q, want .fluffy-Chip:hover", got)
	}
}

func TestSelectorToCSS_Descendant(t *testing.T) {
	sel := style.Selector{
		Type:   "Label",
		Parent: &style.Selector{Type: "Splitter"},
	}
	got := selectorToCSS(sel)
	if got != ".fluffy-Splitter .fluffy-Label" {
		t.Errorf("got %q", got)
	}
}

func TestStyleToCSS_Colors(t *testing.T) {
	s := style.Style{
		Foreground: compositor.RGB(240, 238, 232),
		Background: compositor.RGB(12, 12, 16),
	}
	got := styleToCSS(s)
	if !strings.Contains(got, "color: #f0eee8") {
		t.Errorf("missing fg color: %s", got)
	}
	if !strings.Contains(got, "background-color: #0c0c10") {
		t.Errorf("missing bg color: %s", got)
	}
}

func TestStyleToCSS_Padding(t *testing.T) {
	s := style.Style{
		Padding: style.PadXY(1, 0),
	}
	got := styleToCSS(s)
	if !strings.Contains(got, "padding: 0ch 1ch") {
		t.Errorf("missing padding: %s", got)
	}
}

func TestStylesheetToCSS(t *testing.T) {
	sheet := style.NewStylesheet()
	sheet.Add(style.Select("Label"), style.Style{
		Foreground: compositor.RGB(240, 238, 232),
	})
	got := StylesheetToCSS(sheet)
	if !strings.Contains(got, ".fluffy-Label") {
		t.Errorf("missing selector: %s", got)
	}
	if !strings.Contains(got, "color: #f0eee8") {
		t.Errorf("missing color: %s", got)
	}
}
